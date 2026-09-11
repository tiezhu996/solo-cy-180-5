package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
)

// RegisterUserRoutes 注册用户相关路由。
func RegisterUserRoutes(g *gin.RouterGroup, h *handler.UserHandler, cfg *config.Config, logger *slog.Logger) {
	authGroup := g.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.GET("/me", middleware.Auth(cfg.JWTSecret, logger), h.Me)
	}

	adminGroup := g.Group("/users", middleware.Auth(cfg.JWTSecret, logger), middleware.RBAC(logger, constants.RoleAdmin))
	{
		adminGroup.GET("", h.List)
		adminGroup.PUT("/:id/role", h.UpdateRole)
		adminGroup.DELETE("/:id", h.Delete)
	}
}
