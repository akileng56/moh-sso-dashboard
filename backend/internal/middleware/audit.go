package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

func AuditMiddleware(audit *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip auditing of health checks or static routes
		if c.FullPath() == "/health" {
			c.Next()
			return
		}

		start := time.Now()

		// Process request first
		c.Next()

		duration := time.Since(start)

		userID := c.GetString("user_id")

		// Build metadata
		meta := map[string]interface{}{
			"method":     c.Request.Method,
			"status":     c.Writer.Status(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency_ms": duration.Milliseconds(),
		}

		err := audit.Log(
			c.Request.Context(),
			utils.ToNullUUID(userID),
			"API_CALL:"+c.FullPath(),
			meta,
		)

		if err != nil {
			// Log error internally, don't interrupt API
			log.Printf("[AUDIT ERROR] Failed to record audit log: %v", err)
		}
	}
}
