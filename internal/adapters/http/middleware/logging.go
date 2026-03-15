package middleware

import (
	"time"

	"github.com/edgekit/edgekit/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Logging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		rid, _ := c.Get("request_id")

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("http request",
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"request_id", rid,
			"client_ip", c.ClientIP(),
		)
	}
}
