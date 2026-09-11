// Package middleware 提供认证、RBAC、审计、限流等横切中间件。
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/util"
)

// ContextKey 上下文键。
const (
	ContextKeyUserID    = "auth_user_id"
	ContextKeyUsername  = "auth_username"
	ContextKeyRole      = "auth_role"
	ContextKeyClaims    = "auth_claims"
	ContextKeyRequestID = "request_id"
	ContextKeyAudited   = "audited"
)

// CurrentUser 从上下文还原当前用户。
func CurrentUser(c *gin.Context) (*model.User, error) {
	claimsVal, ok := c.Get(ContextKeyClaims)
	if !ok {
		return nil, util.NewAppError(40100, "未认证用户", nil)
	}
	claims, ok := claimsVal.(*util.Claims)
	if !ok {
		return nil, util.NewAppError(40100, "认证信息解析失败", nil)
	}
	return &model.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

// RequestID 从上下文取请求 ID。
func RequestID(c *gin.Context) string {
	return c.GetString(ContextKeyRequestID)
}
