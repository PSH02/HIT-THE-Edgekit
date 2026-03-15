package middleware

import (
	"fmt"

	"github.com/edgekit/edgekit/internal/adapters/http/response"
	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/edgekit/edgekit/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

func RateLimit(limiter ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if ac, ok := auth.FromContext(c.Request.Context()); ok {
			key = "user:" + ac.UserID
		}

		result, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", itoa(result.Remaining))

		if !result.Allowed {
			c.Header("Retry-After", itoa(result.RetryAfter))
			response.AbortWithError(c, apperror.New(apperror.CodeRateLimited, "rate limit exceeded"))
			return
		}

		c.Next()
	}
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}
