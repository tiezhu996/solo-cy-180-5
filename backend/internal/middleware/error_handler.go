package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/util"
)

// ErrorHandler 统一将业务错误转换为标准 JSON 响应。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			httpStatus := statusOf(appErr.Code)
			util.Fail(c, httpStatus, appErr.Code, appErr.Message)
			logger.Warn("request error", "request_id", RequestID(c), "path", c.FullPath(), "code", appErr.Code, "message", appErr.Message)
			return
		}
		util.Fail(c, http.StatusInternalServerError, constants.CodeInternal, fmt.Sprintf("%s: %v", constants.MsgInternalError, err))
		logger.Error("unhandled error", "request_id", RequestID(c), "path", c.FullPath(), "error", err)
	}
}

func statusOf(code int) int {
	switch code {
	case constants.CodeBadRequest, constants.CodeValidation:
		return http.StatusBadRequest
	case constants.CodeUnauthorized, constants.CodeLoginFailed:
		return http.StatusUnauthorized
	case constants.CodeForbidden:
		return http.StatusForbidden
	case constants.CodeNotFound:
		return http.StatusNotFound
	case constants.CodeConflict, constants.CodeDuplicateName, constants.CodeProjectStatus, constants.CodeRecordingStatus, constants.CodeMarkerConflict:
		return http.StatusConflict
	case constants.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
