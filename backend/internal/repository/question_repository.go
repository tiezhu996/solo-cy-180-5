package repository

import (
	"errors"
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
)

// QuestionRepository 采访问题数据访问接口。
type QuestionRepository interface {
	Create(question *model.Question) error
	FindByID(id uint) (*model.Question, error)
	ListByProject(projectID uint) ([]model.Question, error)
	Update(question *model.Question) error
	Delete(id uint) error
	CountByProject(projectID uint) (int64, error)
}

type questionRepository struct {
	db *gorm.DB
}

// NewQuestionRepository 构造问题仓储。
func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) Create(question *model.Question) error {
	if err := r.db.Create(question).Error; err != nil {
		return fmt.Errorf("create question %s: %w", question.Content, err)
	}
	return nil
}

func (r *questionRepository) FindByID(id uint) (*model.Question, error) {
	var question model.Question
	if err := r.db.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find question by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find question by id: %w", err)
	}
	return &question, nil
}

func (r *questionRepository) ListByProject(projectID uint) ([]model.Question, error) {
	var questions []model.Question
	if err := r.db.Where("project_id = ?", projectID).Order("sort_order ASC, id ASC").Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("list questions of project %d: %w", projectID, err)
	}
	return questions, nil
}

func (r *questionRepository) Update(question *model.Question) error {
	if err := r.db.Save(question).Error; err != nil {
		return fmt.Errorf("update question %d: %w", question.ID, err)
	}
	return nil
}

func (r *questionRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", id).Delete(&model.Recording{}).Error; err != nil {
			return fmt.Errorf("delete recordings of question %d: %w", id, err)
		}
		if err := tx.Delete(&model.Question{}, id).Error; err != nil {
			return fmt.Errorf("delete question %d: %w", id, err)
		}
		return nil
	})
}

func (r *questionRepository) CountByProject(projectID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.Question{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count questions of project %d: %w", projectID, err)
	}
	return total, nil
}
