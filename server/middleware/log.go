package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zeu5/visualizations/log"
)

func Logger(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	// Process request
	c.Next()

	end := time.Now()
	if raw != "" {
		path = path + "?" + raw
	}
	log.With(log.LogParams{
		"timestamp":   end,
		"latency":     end.Sub(start),
		"client_ip":   c.ClientIP(),
		"method":      c.Request.Method,
		"status_code": c.Writer.Status(),
		"error":       c.Errors.ByType(gin.ErrorTypePrivate).String(),
		"body_size":   c.Writer.Size(),
		"path":        path,
	}).Info("Handled request")
}
