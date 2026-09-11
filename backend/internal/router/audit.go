package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
)

// RegisterAuditRoutes 注册审计日志路由。
func RegisterAuditRoutes(g *gin.RouterGroup, h *handler.AuditHandler, cfg *config.Config, logger *slog.Logger) {
	group := g.Group("/audit-logs", middleware.Auth(cfg.JWTSecret, logger), middleware.RBAC(logger, constants.RoleAdmin))
	{
		group.GET("", h.List)
	}
}
