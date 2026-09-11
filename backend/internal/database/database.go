// Package database 负责 GORM 初始化与迁移。
package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// New 建立 MySQL 连接并自动迁移表结构。
func New(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Question{},
		&model.Recording{},
		&model.TimelineMarker{},
		&model.AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	logger.Info("database migrated", "host", cfg.DBHost, "db", cfg.DBName)
	return db, nil
}

// SeedAdmin 初始化默认管理员账号。
func SeedAdmin(db *gorm.DB, logger *slog.Logger) error {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return fmt.Errorf("count admin: %w", err)
	}
	if count > 0 {
		return nil
	}
	hash, err := util.HashPassword("admin123456")
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	admin := model.User{
		Username:     "admin",
		PasswordHash: hash,
		DisplayName:  "系统管理员",
		Email:        "admin@oralhistory.local",
		Role:         "admin",
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	logger.Info("admin seeded", "username", admin.Username)
	return nil
}
