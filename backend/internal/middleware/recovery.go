package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/util"
)

// Recovery 捕获 panic 并返回统一错误响应。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf(constants.LogPanicRecovered, RequestID(c), r))
				util.Fail(c, http.StatusInternalServerError, constants.CodeInternal, fmt.Sprintf("%s: %v", constants.MsgInternalError, r))
				c.Abort()
			}
		}()
		c.Next()
	}
}
