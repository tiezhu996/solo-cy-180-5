package repository

import (
	"errors"
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
)

// TimelineMarkerRepository 时间轴节点数据访问接口。
type TimelineMarkerRepository interface {
	Create(marker *model.TimelineMarker) error
	FindByID(id uint) (*model.TimelineMarker, error)
	ListByProject(projectID uint) ([]model.TimelineMarker, error)
	ListByRecording(recordingID uint) ([]model.TimelineMarker, error)
	Update(marker *model.TimelineMarker) error
	Delete(id uint) error
}

type timelineMarkerRepository struct {
	db *gorm.DB
}

// NewTimelineMarkerRepository 构造时间轴节点仓储。
func NewTimelineMarkerRepository(db *gorm.DB) TimelineMarkerRepository {
	return &timelineMarkerRepository{db: db}
}

func (r *timelineMarkerRepository) Create(marker *model.TimelineMarker) error {
	if err := r.db.Create(marker).Error; err != nil {
		return fmt.Errorf("create marker %s: %w", marker.Label, err)
	}
	return nil
}

func (r *timelineMarkerRepository) FindByID(id uint) (*model.TimelineMarker, error) {
	var marker model.TimelineMarker
	if err := r.db.First(&marker, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find marker by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find marker by id: %w", err)
	}
	return &marker, nil
}

func (r *timelineMarkerRepository) ListByProject(projectID uint) ([]model.TimelineMarker, error) {
	var markers []model.TimelineMarker
	if err := r.db.Where("project_id = ?", projectID).Order("timestamp_second ASC, id ASC").Find(&markers).Error; err != nil {
		return nil, fmt.Errorf("list markers of project %d: %w", projectID, err)
	}
	return markers, nil
}

func (r *timelineMarkerRepository) ListByRecording(recordingID uint) ([]model.TimelineMarker, error) {
	var markers []model.TimelineMarker
	if err := r.db.Where("recording_id = ?", recordingID).Order("timestamp_second ASC, id ASC").Find(&markers).Error; err != nil {
		return nil, fmt.Errorf("list markers of recording %d: %w", recordingID, err)
	}
	return markers, nil
}

func (r *timelineMarkerRepository) Update(marker *model.TimelineMarker) error {
	if err := r.db.Save(marker).Error; err != nil {
		return fmt.Errorf("update marker %d: %w", marker.ID, err)
	}
	return nil
}

func (r *timelineMarkerRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.TimelineMarker{}, id).Error; err != nil {
		return fmt.Errorf("delete marker %d: %w", id, err)
	}
	return nil
}
