package repository

import (
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
)

// AuditLogRepository 审计日志数据访问接口。
type AuditLogRepository interface {
	Create(log *model.AuditLog) error
	List(page, pageSize int, username string) ([]model.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository 构造审计日志仓储。
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *model.AuditLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return fmt.Errorf("create audit log %s: %w", log.Action, err)
	}
	return nil
}

func (r *auditLogRepository) List(page, pageSize int, username string) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	q := r.db.Model(&model.AuditLog{})
	if username != "" {
		q = q.Where("username = ?", username)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	if err := q.Scopes(paginate(page, pageSize)).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, total, nil
}
