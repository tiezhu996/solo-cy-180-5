package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
)

// RegisterTimelineMarkerRoutes 注册时间轴节点路由。
func RegisterTimelineMarkerRoutes(g *gin.RouterGroup, h *handler.TimelineMarkerHandler, cfg *config.Config, logger *slog.Logger) {
	group := g.Group("/timeline-markers", middleware.Auth(cfg.JWTSecret, logger))
	{
		group.GET("", h.List)
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}
