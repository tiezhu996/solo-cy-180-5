package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/service"
)

// Audit 为写操作生成审计日志（service 层同时埋点，此处兜底记录请求级信息）。
func Audit(auditSvc service.AuditService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		userID := c.GetUint(ContextKeyUserID)
		username := c.GetString(ContextKeyUsername)
		role := c.GetString(ContextKeyRole)
		if userID == 0 {
			return
		}
		action := c.Request.Method + " " + c.FullPath()
		detail := ""
		if c.Writer.Status() >= 400 {
			detail = "status=" + itoa(c.Writer.Status())
		}
		auditSvc.Record(userID, username, role, action, "request", 0, detail, c.ClientIP(), RequestID(c))
		logger.Debug("audit middleware done", "path", c.FullPath(), "latency_ms", time.Since(start).Milliseconds())
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
