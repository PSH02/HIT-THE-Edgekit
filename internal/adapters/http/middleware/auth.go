package middleware

import (
	"net/http"
	"strings"

	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/adapters/http/response"
	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/gin-gonic/gin"
)

func Auth(tokenSvc *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.AbortWithError(c, apperror.New(apperror.CodeUnauthorized, "missing authorization header"))
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			response.AbortWithError(c, apperror.New(apperror.CodeUnauthorized, "invalid authorization format"))
			return
		}

		ac, err := tokenSvc.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Success: false,
				Error: &response.ErrorBody{
					Code:    apperror.CodeUnauthorized,
					Message: "invalid or expired token",
				},
			})
			return
		}

		ctx := auth.WithAuth(c.Request.Context(), ac)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
