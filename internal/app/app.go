package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	grpcadapter "github.com/edgekit/edgekit/internal/adapters/grpc"
	httpadapter "github.com/edgekit/edgekit/internal/adapters/http"
	"github.com/edgekit/edgekit/internal/adapters/repository/memory"
	"github.com/edgekit/edgekit/internal/adapters/repository/postgres"
	redisrepo "github.com/edgekit/edgekit/internal/adapters/repository/redis"
	"github.com/edgekit/edgekit/internal/app/config"
	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/session"
	"github.com/edgekit/edgekit/internal/core/user"
	"github.com/edgekit/edgekit/pkg/jwt"
	"github.com/edgekit/edgekit/pkg/logger"
	"github.com/edgekit/edgekit/pkg/ratelimit"
)

type App struct {
	cfg    *config.Config
	log    logger.Logger
	closer []func()
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
		log: logger.New(cfg.LogLevel),
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(a.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	a.addCloser(func() { pool.Close() })

	redisClient := redisrepo.Connect(a.cfg.RedisURL)
	a.addCloser(func() { redisClient.Close() })

	jwtMgr, err := jwt.NewManager(jwt.Config{
		PrivateKeyPath: a.cfg.JWT.PrivateKeyPath,
		PublicKeyPath:  a.cfg.JWT.PublicKeyPath,
		AccessTTL:      a.cfg.JWT.AccessTTL,
		RefreshTTL:     a.cfg.JWT.RefreshTTL,
	})
	if err != nil {
		return fmt.Errorf("init jwt: %w", err)
	}

	tokenSvc := auth.NewTokenService(jwtMgr)
	hasher := auth.NewBcryptHasher()

	userRepo := postgres.NewUserRepository(pool)
	userCache := redisrepo.NewUserCache(redisClient)
	userSvc := user.NewService(userRepo, userCache, hasher, tokenSvc)

	sessionRepo := postgres.NewSessionRepository(pool)
	sessionSvc := session.NewService(sessionRepo)

	var limiter ratelimit.Limiter = memory.NewRateLimiter(ratelimit.Config{
		Rate:   a.cfg.RateLimit.Rate,
		Window: a.cfg.RateLimit.Window,
	})

	httpRouter := httpadapter.NewRouter(httpadapter.RouterDeps{
		UserService:    userSvc,
		SessionService: sessionSvc,
		TokenService:   tokenSvc,
		RateLimiter:    limiter,
		Logger:         a.log,
	})

	httpServer := &http.Server{
		Addr:         a.cfg.HTTPAddr,
		Handler:      httpRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	grpcServer := grpcadapter.NewServer(grpcadapter.ServerDeps{
		UserService:    userSvc,
		SessionService: sessionSvc,
		Limiter:        limiter,
		TokenService:   tokenSvc,
		Logger:         a.log,
	})

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		a.log.Info("http server starting", "addr", a.cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		lis, err := net.Listen("tcp", a.cfg.GRPCAddr)
		if err != nil {
			return fmt.Errorf("grpc listen: %w", err)
		}
		a.log.Info("grpc server starting", "addr", a.cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		a.log.Info("shutting down servers")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		grpcServer.GracefulStop()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}

		a.close()
		return nil
	})

	return g.Wait()
}

func (a *App) addCloser(fn func()) {
	a.closer = append(a.closer, fn)
}

func (a *App) close() {
	for i := len(a.closer) - 1; i >= 0; i-- {
		a.closer[i]()
	}
}
