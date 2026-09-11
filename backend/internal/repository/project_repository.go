package repository

import (
	"errors"
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProjectRepository 采访项目数据访问接口。
type ProjectRepository interface {
	Create(project *model.Project) error
	FindByID(id uint) (*model.Project, error)
	List(page, pageSize int, status string) ([]model.Project, int64, error)
	ListByUser(userID uint, page, pageSize int) ([]model.Project, int64, error)
	FindByIDForUpdate(id uint) (*model.Project, error)
	Update(project *model.Project) error
	UpdateStatus(project *model.Project) error
	Delete(id uint) error
	Count() (int64, error)
}

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 构造项目仓储。
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *model.Project) error {
	if err := r.db.Create(project).Error; err != nil {
		return fmt.Errorf("create project %s: %w", project.Title, err)
	}
	return nil
}

func (r *projectRepository) FindByID(id uint) (*model.Project, error) {
	var project model.Project
	if err := r.db.Preload("Questions").Preload("Recordings").Preload("TimelineMarkers").First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find project by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find project by id: %w", err)
	}
	return &project, nil
}

func (r *projectRepository) List(page, pageSize int, status string) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64
	q := r.db.Model(&model.Project{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}
	if err := q.Scopes(paginate(page, pageSize)).Order("id DESC").Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	return projects, total, nil
}

func (r *projectRepository) ListByUser(userID uint, page, pageSize int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64
	if err := r.db.Model(&model.Project{}).Where("created_by = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count projects by user %d: %w", userID, err)
	}
	if err := r.db.Model(&model.Project{}).Where("created_by = ?", userID).Scopes(paginate(page, pageSize)).
		Order("id DESC").Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("list projects by user: %w", err)
	}
	return projects, total, nil
}

func (r *projectRepository) FindByIDForUpdate(id uint) (*model.Project, error) {
	var project model.Project
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find project for update by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find project for update: %w", err)
	}
	return &project, nil
}

func (r *projectRepository) Update(project *model.Project) error {
	if err := r.db.Save(project).Error; err != nil {
		return fmt.Errorf("update project %d: %w", project.ID, err)
	}
	return nil
}

func (r *projectRepository) UpdateStatus(project *model.Project) error {
	if err := r.db.Model(project).Update("status", project.Status).Error; err != nil {
		return fmt.Errorf("update project %d status: %w", project.ID, err)
	}
	return nil
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", id).Delete(&model.TimelineMarker{}).Error; err != nil {
			return fmt.Errorf("delete markers of project %d: %w", id, err)
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.Recording{}).Error; err != nil {
			return fmt.Errorf("delete recordings of project %d: %w", id, err)
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.Question{}).Error; err != nil {
			return fmt.Errorf("delete questions of project %d: %w", id, err)
		}
		if err := tx.Delete(&model.Project{}, id).Error; err != nil {
			return fmt.Errorf("delete project %d: %w", id, err)
		}
		return nil
	})
}

func (r *projectRepository) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&model.Project{}).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count projects: %w", err)
	}
	return total, nil
}
