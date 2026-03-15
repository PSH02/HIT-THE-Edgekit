package httpadapter

import (
	"github.com/edgekit/edgekit/internal/adapters/http/handler"
	"github.com/edgekit/edgekit/internal/adapters/http/middleware"
	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/session"
	"github.com/edgekit/edgekit/internal/core/user"
	"github.com/edgekit/edgekit/pkg/logger"
	"github.com/edgekit/edgekit/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

type RouterDeps struct {
	UserService    *user.Service
	SessionService *session.Service
	TokenService   *auth.TokenService
	RateLimiter    ratelimit.Limiter
	Logger         logger.Logger
}

func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()

	r.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.CORS(),
		middleware.Logging(deps.Logger),
	)

	health := handler.NewHealthHandler()
	r.GET("/healthz", health.Healthz)
	r.GET("/readyz", health.Readyz)

	v1 := r.Group("/api/v1")

	if deps.RateLimiter != nil {
		v1.Use(middleware.RateLimit(deps.RateLimiter))
	}

	authGroup := v1.Group("/auth")
	{
		uh := handler.NewUserHandler(deps.UserService)
		authGroup.POST("/register", uh.Register)
		authGroup.POST("/login", uh.Login)
	}

	protected := v1.Group("")
	protected.Use(middleware.Auth(deps.TokenService))
	{
		uh := handler.NewUserHandler(deps.UserService)
		protected.GET("/users/me", uh.GetProfile)
		protected.PATCH("/users/me", uh.UpdateProfile)

		sh := handler.NewSessionHandler(deps.SessionService)
		protected.POST("/sessions", sh.Create)
		protected.GET("/sessions", sh.ListWaiting)
		protected.GET("/sessions/:id", sh.Get)
		protected.POST("/sessions/:id/join", sh.Join)
		protected.POST("/sessions/:id/leave", sh.Leave)
	}

	return r
}
