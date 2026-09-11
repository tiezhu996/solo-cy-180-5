package service

import (
	"log/slog"

	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/util"
)

// AuditService 审计日志业务接口。
type AuditService interface {
	Record(userID uint, username, role, action, entityType string, entityID uint, detail, ip, requestID string)
	List(page, pageSize int, username string) ([]model.AuditLog, int64, error)
}

type auditService struct {
	auditRepo repository.AuditLogRepository
	logger    *slog.Logger
}

// NewAuditService 构造审计服务。
func NewAuditService(auditRepo repository.AuditLogRepository, logger *slog.Logger) AuditService {
	return &auditService{auditRepo: auditRepo, logger: logger}
}

func (s *auditService) Record(userID uint, username, role, action, entityType string, entityID uint, detail, ip, requestID string) {
	log := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Role:       role,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Detail:     detail,
		IP:         ip,
		RequestID:  requestID,
	}
	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Error("audit log record failed", "action", action, "error", err)
		return
	}
	s.logger.Info("audit recorded", "username", username, "action", action, "entity_type", entityType)
}

func (s *auditService) List(page, pageSize int, username string) ([]model.AuditLog, int64, error) {
	logs, total, err := s.auditRepo.List(page, pageSize, username)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternal, "审计日志查询失败", err)
	}
	return logs, total, nil
}
