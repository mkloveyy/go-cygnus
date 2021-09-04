package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"

	"go-cygnus/utils/logging"
)

func AccessLogger() gin.HandlerFunc {
	logger := logging.GetLogger("access")

	return func(c *gin.Context) {
		start := time.Now()
		url := c.Request.URL

		c.Next()

		logger.WithFields(map[string]interface{}{
			"method":     c.Request.Method,
			"cost":       time.Since(start).Seconds(),
			"client-ip":  c.ClientIP(),
			"user-agent": c.Request.UserAgent(),
		}).Info(url)
	}
}
