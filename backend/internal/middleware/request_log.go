package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
)

// RequestLog 输出包含 request_id/method/path/status/latency_ms 的结构化请求日志。
func RequestLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info(fmt.Sprintf(constants.LogRequest, RequestID(c), c.Request.Method, c.Request.URL.Path,
			c.Writer.Status(), time.Since(start).Milliseconds()))
	}
}
