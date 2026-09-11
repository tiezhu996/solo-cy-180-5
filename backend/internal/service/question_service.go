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

// QuestionService 采访问题业务接口。
type QuestionService interface {
	Create(actor *model.User, projectID uint, req *dto.CreateQuestionRequest) (*model.Question, error)
	ListByProject(projectID uint) ([]model.Question, error)
	Update(actor *model.User, id uint, req *dto.UpdateQuestionRequest) (*model.Question, error)
	Delete(actor *model.User, id uint) error
	CountByProject(projectID uint) (int64, error)
}

type questionService struct {
	questionRepo repository.QuestionRepository
	projectRepo  repository.ProjectRepository
	logger       *slog.Logger
}

// NewQuestionService 构造问题服务。
func NewQuestionService(questionRepo repository.QuestionRepository, projectRepo repository.ProjectRepository, logger *slog.Logger) QuestionService {
	return &questionService{questionRepo: questionRepo, projectRepo: projectRepo, logger: logger}
}

func (s *questionService) Create(actor *model.User, projectID uint, req *dto.CreateQuestionRequest) (*model.Question, error) {
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("项目 %d 不存在", projectID), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询项目 %d 失败", projectID), err)
	}
	question := &model.Question{
		ProjectID: projectID,
		Content:   req.Content,
		SortOrder: req.SortOrder,
	}
	if err := s.questionRepo.Create(question); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("向项目 %d 添加问题失败", projectID), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogQuestionCreate, actor.Username, projectID, question.Content))
	return question, nil
}

func (s *questionService) ListByProject(projectID uint) ([]model.Question, error) {
	questions, err := s.questionRepo.ListByProject(projectID)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询项目 %d 问题列表失败", projectID), err)
	}
	return questions, nil
}

func (s *questionService) Update(actor *model.User, id uint, req *dto.UpdateQuestionRequest) (*model.Question, error) {
	question, err := s.questionRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("问题 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询问题 %d 失败", id), err)
	}
	if req.Content != "" {
		question.Content = req.Content
	}
	if req.SortOrder != 0 {
		question.SortOrder = req.SortOrder
	}
	if err := s.questionRepo.Update(question); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("更新问题 %d 失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogQuestionUpdate, actor.Username, question.ID, question.Content))
	return question, nil
}

func (s *questionService) Delete(actor *model.User, id uint) error {
	if _, err := s.questionRepo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("问题 %d 不存在", id), err)
		}
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询问题 %d 失败", id), err)
	}
	if err := s.questionRepo.Delete(id); err != nil {
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("删除问题 %d 失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogQuestionDelete, actor.Username, id))
	return nil
}

func (s *questionService) CountByProject(projectID uint) (int64, error) {
	return s.questionRepo.CountByProject(projectID)
}
