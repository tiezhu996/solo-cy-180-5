package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/util"
)

// RBAC 校验当前用户角色是否在允许列表内。
func RBAC(logger *slog.Logger, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(ContextKeyRole)
		username := c.GetString(ContextKeyUsername)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		logger.Warn(fmt.Sprintf(constants.LogAuthDenied, username, c.FullPath(), role, fmt.Sprint(roles)))
		util.Fail(c, http.StatusForbidden, constants.CodeForbidden,
			fmt.Sprintf("%s: 用户 %s 的角色 %s 不允许访问 %s", constants.MsgForbidden, username, role, c.FullPath()))
		c.Abort()
	}
}
