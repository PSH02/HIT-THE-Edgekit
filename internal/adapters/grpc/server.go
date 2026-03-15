package grpcadapter

import (
	"github.com/edgekit/edgekit/internal/adapters/grpc/interceptor"
	"github.com/edgekit/edgekit/internal/adapters/grpc/service"
	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/session"
	"github.com/edgekit/edgekit/internal/core/user"
	"github.com/edgekit/edgekit/pkg/logger"
	"github.com/edgekit/edgekit/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerDeps struct {
	UserService    *user.Service
	SessionService *session.Service
	Limiter        ratelimit.Limiter
	TokenService   *auth.TokenService
	Logger         logger.Logger
}

func NewServer(deps ServerDeps) *grpc.Server {
	recoveryInt := interceptor.NewRecoveryInterceptor(deps.Logger)
	loggingInt := interceptor.NewLoggingInterceptor(deps.Logger)
	authInt := interceptor.NewAuthInterceptor(deps.TokenService)
	rateLimitInt := interceptor.NewRateLimitInterceptor(deps.Limiter)
	requestIDInt := interceptor.NewRequestIDInterceptor()

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInt.Unary(),
			requestIDInt.Unary(),
			loggingInt.Unary(),
			rateLimitInt.Unary(),
			authInt.Unary(),
		),
		grpc.ChainStreamInterceptor(
			recoveryInt.Stream(),
			loggingInt.Stream(),
		),
	)

	userSvc := service.NewUserServiceServer(deps.UserService)
	sessionSvc := service.NewSessionServiceServer(deps.SessionService)

	userSvc.Register(srv)
	sessionSvc.Register(srv)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	reflection.Register(srv)

	return srv
}
