package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

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
	// UploadAudio 负责录音文件上传的全部编排：生成对象 key、写入对象存储、关联录音记录。
	UploadAudio(ctx context.Context, actor *model.User, id uint, filename string, file io.Reader, size int64, contentType string, duration int) (*model.Recording, error)
	// OpenAudio 返回录音元数据与其音频内容流，调用方负责关闭流。
	OpenAudio(ctx context.Context, id uint) (*model.Recording, io.ReadCloser, error)
	// Delete 先清理对象存储中的音频文件，再删除录音记录；任一步失败后重试都能收敛到最终一致。
	Delete(ctx context.Context, actor *model.User, id uint) error
	CountByProject(projectID uint) (int64, error)
}

type recordingService struct {
	recordingRepo repository.RecordingRepository
	projectRepo   repository.ProjectRepository
	questionRepo  repository.QuestionRepository
	storageSvc    StorageService
	logger        *slog.Logger
}

// NewRecordingService 构造录音服务。
func NewRecordingService(recordingRepo repository.RecordingRepository, projectRepo repository.ProjectRepository, questionRepo repository.QuestionRepository, storageSvc StorageService, logger *slog.Logger) RecordingService {
	return &recordingService{recordingRepo: recordingRepo, projectRepo: projectRepo, questionRepo: questionRepo, storageSvc: storageSvc, logger: logger}
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

// UploadAudio 编排录音文件上传：对象 key 规则、对象存储写入、录音记录关联都收口在业务层。
func (s *recordingService) UploadAudio(ctx context.Context, actor *model.User, id uint, filename string, file io.Reader, size int64, contentType string, duration int) (*model.Recording, error) {
	objectKey := buildAudioObjectKey(actor.ID, id, filename)
	if err := s.storageSvc.Upload(ctx, objectKey, file, size, contentType); err != nil {
		s.logger.Error("upload audio failed", "recording_id", id, "object_key", objectKey, "error", err)
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("录音 %d 文件上传失败", id), err)
	}
	recording, err := s.attachAudio(actor, id, objectKey, duration)
	if err != nil {
		// 对象已上传但记录关联失败，回滚新对象，避免留下孤儿文件。
		if rmErr := s.storageSvc.Remove(ctx, objectKey); rmErr != nil {
			s.logger.Error("rollback uploaded audio failed", "recording_id", id, "object_key", objectKey, "error", rmErr)
		}
		return nil, err
	}
	return recording, nil
}

// OpenAudio 校验录音及其音频是否存在，并从对象存储取出内容流。
func (s *recordingService) OpenAudio(ctx context.Context, id uint) (*model.Recording, io.ReadCloser, error) {
	recording, err := s.Get(id)
	if err != nil {
		return nil, nil, err
	}
	if recording.AudioKey == "" {
		return nil, nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 尚无音频文件", id), nil)
	}
	obj, err := s.storageSvc.Get(ctx, recording.AudioKey)
	if err != nil {
		s.logger.Error("get audio failed", "recording_id", id, "object_key", recording.AudioKey, "error", err)
		return nil, nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("录音 %d 音频读取失败", id), err)
	}
	return recording, obj, nil
}

// buildAudioObjectKey 按固定规则生成对象存储 key，并从原始文件名推断扩展名。
func buildAudioObjectKey(actorID, recordingID uint, filename string) string {
	ext := "webm"
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		if suffix := strings.ToLower(filename[idx+1:]); suffix != "" {
			ext = suffix
		}
	}
	return fmt.Sprintf("recordings/%d/%d_%s.%s", actorID, recordingID, util.RandHex(8), ext)
}

// attachAudio 将已上传的对象 key 关联到录音记录，并驱动状态机流转。
func (s *recordingService) attachAudio(actor *model.User, id uint, audioKey string, duration int) (*model.Recording, error) {
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

func (s *recordingService) Delete(ctx context.Context, actor *model.User, id uint) error {
	recording, err := s.recordingRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	// 先清对象、后删记录。只要记录还在，audio_key 就在；即使上一次清理在删对象后、删记录前中断，
	// 重试仍能凭记录中的 key 再清一次（对象存储删除幂等），最终删掉记录，保证两边一致且可重试。
	if recording.AudioKey != "" {
		if err := s.storageSvc.Remove(ctx, recording.AudioKey); err != nil {
			s.logger.Error("remove audio object failed", "recording_id", id, "object_key", recording.AudioKey, "error", err)
			return util.NewAppError(constants.CodeInternal, fmt.Sprintf("录音 %d 音频文件清理失败", id), err)
		}
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
