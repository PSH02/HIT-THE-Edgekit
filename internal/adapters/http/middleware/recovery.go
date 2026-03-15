package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/edgekit/edgekit/internal/adapters/http/response"
	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/edgekit/edgekit/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"error", fmt.Sprintf("%v", r),
					"stack", string(debug.Stack()),
					"request_id", c.GetString("request_id"),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error: &response.ErrorBody{
						Code:    apperror.CodeInternal,
						Message: "internal server error",
					},
				})
			}
		}()
		c.Next()
	}
}
