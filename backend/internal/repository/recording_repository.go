package repository

import (
	"errors"
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecordingRepository 录音片段数据访问接口。
type RecordingRepository interface {
	Create(recording *model.Recording) error
	FindByID(id uint) (*model.Recording, error)
	ListByProject(projectID uint) ([]model.Recording, error)
	ListByQuestion(questionID uint) ([]model.Recording, error)
	FindByIDForUpdate(id uint) (*model.Recording, error)
	Update(recording *model.Recording) error
	UpdateStatus(recording *model.Recording) error
	Delete(id uint) error
	CountByProject(projectID uint) (int64, error)
}

type recordingRepository struct {
	db *gorm.DB
}

// NewRecordingRepository 构造录音仓储。
func NewRecordingRepository(db *gorm.DB) RecordingRepository {
	return &recordingRepository{db: db}
}

func (r *recordingRepository) Create(recording *model.Recording) error {
	if err := r.db.Create(recording).Error; err != nil {
		return fmt.Errorf("create recording of question %d: %w", recording.QuestionID, err)
	}
	return nil
}

func (r *recordingRepository) FindByID(id uint) (*model.Recording, error) {
	var recording model.Recording
	if err := r.db.First(&recording, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find recording by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find recording by id: %w", err)
	}
	return &recording, nil
}

func (r *recordingRepository) ListByProject(projectID uint) ([]model.Recording, error) {
	var recordings []model.Recording
	if err := r.db.Where("project_id = ?", projectID).Order("id ASC").Find(&recordings).Error; err != nil {
		return nil, fmt.Errorf("list recordings of project %d: %w", projectID, err)
	}
	return recordings, nil
}

func (r *recordingRepository) ListByQuestion(questionID uint) ([]model.Recording, error) {
	var recordings []model.Recording
	if err := r.db.Where("question_id = ?", questionID).Order("id ASC").Find(&recordings).Error; err != nil {
		return nil, fmt.Errorf("list recordings of question %d: %w", questionID, err)
	}
	return recordings, nil
}

func (r *recordingRepository) FindByIDForUpdate(id uint) (*model.Recording, error) {
	var recording model.Recording
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&recording, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find recording for update by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find recording for update: %w", err)
	}
	return &recording, nil
}

func (r *recordingRepository) Update(recording *model.Recording) error {
	if err := r.db.Save(recording).Error; err != nil {
		return fmt.Errorf("update recording %d: %w", recording.ID, err)
	}
	return nil
}

func (r *recordingRepository) UpdateStatus(recording *model.Recording) error {
	if err := r.db.Model(recording).Update("status", recording.Status).Error; err != nil {
		return fmt.Errorf("update recording %d status: %w", recording.ID, err)
	}
	return nil
}

func (r *recordingRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("recording_id = ?", id).Delete(&model.TimelineMarker{}).Error; err != nil {
			return fmt.Errorf("delete markers of recording %d: %w", id, err)
		}
		if err := tx.Delete(&model.Recording{}, id).Error; err != nil {
			return fmt.Errorf("delete recording %d: %w", id, err)
		}
		return nil
	})
}

func (r *recordingRepository) CountByProject(projectID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.Recording{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count recordings of project %d: %w", projectID, err)
	}
	return total, nil
}
