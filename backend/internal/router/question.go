package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
)

// RegisterQuestionRoutes 注册采访问题路由。
func RegisterQuestionRoutes(g *gin.RouterGroup, h *handler.QuestionHandler, cfg *config.Config, logger *slog.Logger) {
	group := g.Group("/projects/:id/questions", middleware.Auth(cfg.JWTSecret, logger))
	{
		group.GET("", h.ListByProject)
		group.POST("", h.Create)
	}
	g.Group("/questions", middleware.Auth(cfg.JWTSecret, logger)).PUT("/:id", h.Update)
	g.Group("/questions", middleware.Auth(cfg.JWTSecret, logger)).DELETE("/:id", h.Delete)
}
