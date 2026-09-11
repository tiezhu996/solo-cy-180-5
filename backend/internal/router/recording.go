package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
)

// RegisterRecordingRoutes 注册录音路由。
func RegisterRecordingRoutes(g *gin.RouterGroup, h *handler.RecordingHandler, cfg *config.Config, logger *slog.Logger) {
	group := g.Group("/recordings", middleware.Auth(cfg.JWTSecret, logger))
	{
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Get)
		group.PUT("/:id", h.Update)
		group.PUT("/:id/summary", h.UpdateSummary)
		group.POST("/:id/audio", h.UploadAudio)
		group.GET("/:id/audio", h.PlayAudio)
		group.DELETE("/:id", h.Delete)
	}
}
