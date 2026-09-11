package database

import (
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
)

// Migrate 执行表结构迁移。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Question{},
		&model.Recording{},
		&model.TimelineMarker{},
		&model.AuditLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
