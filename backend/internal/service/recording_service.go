package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/util"
)

// RecordingService 录音片段业务接口。
type RecordingService interface {
	Create(actor *model.User, req *dto.CreateRecordingRequest) (*model.Recording, error)
	Get(id uint) (*model.Recording, error)
	// List 同时服务「按项目」与「按问题」两个接口，复用同一 service 方法。
	List(projectID, questionID uint) ([]model.Recording, error)
	Update(actor *model.User, id uint, req *dto.UpdateRecordingRequest) (*model.Recording, error)
	UpdateSummary(actor *model.User, id uint, summary string) (*model.Recording, error)
	AttachAudio(actor *model.User, id uint, audioKey string, duration int) (*model.Recording, error)
	Delete(actor *model.User, id uint) error
	CountByProject(projectID uint) (int64, error)
}

type recordingService struct {
	recordingRepo repository.RecordingRepository
	projectRepo   repository.ProjectRepository
	questionRepo  repository.QuestionRepository
	logger        *slog.Logger
}

// NewRecordingService 构造录音服务。
func NewRecordingService(recordingRepo repository.RecordingRepository, projectRepo repository.ProjectRepository, questionRepo repository.QuestionRepository, logger *slog.Logger) RecordingService {
	return &recordingService{recordingRepo: recordingRepo, projectRepo: projectRepo, questionRepo: questionRepo, logger: logger}
}

func (s *recordingService) Create(actor *model.User, req *dto.CreateRecordingRequest) (*model.Recording, error) {
	if _, err := s.projectRepo.FindByID(req.ProjectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("项目 %d 不存在", req.ProjectID), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询项目 %d 失败", req.ProjectID), err)
	}
	if _, err := s.questionRepo.FindByID(req.QuestionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("问题 %d 不存在", req.QuestionID), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询问题 %d 失败", req.QuestionID), err)
	}
	recording := &model.Recording{
		ProjectID:       req.ProjectID,
		QuestionID:      req.QuestionID,
		DurationSeconds: req.DurationSeconds,
		Summary:         req.Summary,
		Status:          constants.RecordingStatusRecording,
		CreatedBy:       actor.ID,
	}
	if err := s.recordingRepo.Create(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("创建问题 %d 的录音失败", req.QuestionID), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingUpload, actor.Username, recording.ProjectID, recording.QuestionID, recording.DurationSeconds, recording.Status))
	return recording, nil
}

func (s *recordingService) Get(id uint) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	return recording, nil
}

func (s *recordingService) List(projectID, questionID uint) ([]model.Recording, error) {
	var (
		recordings []model.Recording
		err        error
	)
	if projectID > 0 {
		recordings, err = s.recordingRepo.ListByProject(projectID)
	} else {
		recordings, err = s.recordingRepo.ListByQuestion(questionID)
	}
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternal, "录音列表查询失败", err)
	}
	return recordings, nil
}

func (s *recordingService) Update(actor *model.User, id uint, req *dto.UpdateRecordingRequest) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	if req.DurationSeconds > 0 {
		recording.DurationSeconds = req.DurationSeconds
	}
	if req.Summary != "" {
		recording.Summary = req.Summary
	}
	if req.Status != "" {
		if !constants.ValidRecordingStatus(req.Status) {
			return nil, util.NewAppError(constants.CodeValidation, fmt.Sprintf("录音状态 %s 不合法", req.Status), nil)
		}
		if !constants.CanTransitionRecording(recording.Status, req.Status) {
			return nil, util.NewAppError(constants.CodeRecordingStatus,
				fmt.Sprintf("录音 %d 状态不允许从 %s 流转到 %s", id, recording.Status, req.Status), nil)
		}
		recording.Status = req.Status
	}
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("更新录音 %d 失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingStatus, actor.Username, recording.ID, recording.Status, recording.Status))
	return recording, nil
}

func (s *recordingService) UpdateSummary(actor *model.User, id uint, summary string) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	recording.Summary = summary
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("更新录音 %d 摘要失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingSummary, actor.Username, recording.ID, summary))
	return recording, nil
}

func (s *recordingService) AttachAudio(actor *model.User, id uint, audioKey string, duration int) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByIDForUpdate(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	recording.AudioKey = audioKey
	if duration > 0 {
		recording.DurationSeconds = duration
	}
	if recording.Status == constants.RecordingStatusRecording || recording.Status == constants.RecordingStatusProcessing {
		recording.Status = constants.RecordingStatusReady
	}
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("录音 %d 音频关联失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingUpload, actor.Username, recording.ProjectID, recording.QuestionID, recording.DurationSeconds, recording.Status))
	return recording, nil
}

func (s *recordingService) Delete(actor *model.User, id uint) error {
	if _, err := s.recordingRepo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	if err := s.recordingRepo.Delete(id); err != nil {
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("删除录音 %d 失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingDelete, actor.Username, id))
	return nil
}

func (s *recordingService) CountByProject(projectID uint) (int64, error) {
	return s.recordingRepo.CountByProject(projectID)
}
