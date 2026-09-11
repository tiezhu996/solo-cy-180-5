package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
)

// RegisterProjectRoutes 注册采访项目路由。
func RegisterProjectRoutes(g *gin.RouterGroup, h *handler.ProjectHandler, cfg *config.Config, logger *slog.Logger) {
	group := g.Group("/projects", middleware.Auth(cfg.JWTSecret, logger))
	{
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/mine", h.ListMine)
		group.GET("/stats", h.Stats)
		group.GET("/:id", h.Get)
		group.PUT("/:id", h.Update)
		group.PUT("/:id/status", h.TransitionStatus)
		group.DELETE("/:id", h.Delete)
	}
}
