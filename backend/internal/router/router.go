// Package router 统一注册路由。
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/service"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Setup 装配依赖并构建路由引擎。
func Setup(cfg *config.Config, db *gorm.DB, rdb *redis.Client, logger *slog.Logger) (*gin.Engine, error) {
	if cfg.RunMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// repository
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	recordingRepo := repository.NewRecordingRepository(db)
	markerRepo := repository.NewTimelineMarkerRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)

	// service
	userSvc := service.NewUserService(userRepo, cfg, logger)
	projectSvc := service.NewProjectService(projectRepo, logger)
	questionSvc := service.NewQuestionService(questionRepo, projectRepo, logger)
	recordingSvc := service.NewRecordingService(recordingRepo, projectRepo, questionRepo, logger)
	markerSvc := service.NewTimelineMarkerService(markerRepo, projectRepo, recordingRepo, logger)
	auditSvc := service.NewAuditService(auditRepo, logger)
	storageSvc, err := service.NewStorageService(cfg, logger)
	if err != nil {
		return nil, err
	}

	// handler
	userHandler := handler.NewUserHandler(userSvc, logger)
	projectHandler := handler.NewProjectHandler(projectSvc, auditSvc, logger)
	questionHandler := handler.NewQuestionHandler(questionSvc, auditSvc, logger)
	recordingHandler := handler.NewRecordingHandler(recordingSvc, storageSvc, auditSvc, logger)
	markerHandler := handler.NewTimelineMarkerHandler(markerSvc, auditSvc, logger)
	auditHandler := handler.NewAuditHandler(auditSvc, logger)

	engine := gin.New()
	engine.Use(middleware.RequestIDMiddleware())
	engine.Use(middleware.CORS())
	engine.Use(middleware.RequestLog(logger))
	engine.Use(middleware.ErrorHandler(logger))
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.RateLimit(rdb, 300, time.Minute, logger))

	engine.GET("/healthz", func(c *gin.Context) {
		sqlDB, err := db.DB()
		status := http.StatusOK
		msg := "ok"
		if err != nil || sqlDB.Ping() != nil {
			status = http.StatusServiceUnavailable
			msg = "db unavailable"
		}
		c.JSON(status, gin.H{"code": 0, "message": msg, "data": gin.H{"service": "oralhistory", "db": status == http.StatusOK}})
	})

	v1 := engine.Group("/api/v1")
	RegisterUserRoutes(v1, userHandler, cfg, logger)
	RegisterProjectRoutes(v1, projectHandler, cfg, logger)
	RegisterQuestionRoutes(v1, questionHandler, cfg, logger)
	RegisterRecordingRoutes(v1, recordingHandler, cfg, logger)
	RegisterTimelineMarkerRoutes(v1, markerHandler, cfg, logger)
	RegisterAuditRoutes(v1, auditHandler, cfg, logger)

	return engine, nil
}
