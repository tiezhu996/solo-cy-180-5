package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/middleware"
	"github.com/oralhistory/oralhistory/internal/service"
	"github.com/oralhistory/oralhistory/internal/util"
)

// QuestionHandler 采访问题接口处理器。
type QuestionHandler struct {
	questionSvc service.QuestionService
	auditSvc    service.AuditService
	logger      *slog.Logger
}

// NewQuestionHandler 构造问题处理器。
func NewQuestionHandler(questionSvc service.QuestionService, auditSvc service.AuditService, logger *slog.Logger) *QuestionHandler {
	return &QuestionHandler{questionSvc: questionSvc, auditSvc: auditSvc, logger: logger}
}

// Create 向项目添加问题。
func (h *QuestionHandler) Create(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	projectID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateQuestionRequest
	if !bindJSON(c, &req) {
		return
	}
	question, err := h.questionSvc.Create(actor, projectID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "question.create", "question", question.ID,
		"添加采访问题 "+question.Content, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgQuestionCreated, question)
}

// ListByProject 项目问题列表。
func (h *QuestionHandler) ListByProject(c *gin.Context) {
	projectID, ok := parseID(c, "id")
	if !ok {
		return
	}
	questions, err := h.questionSvc.ListByProject(projectID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"list": questions})
}

// Update 更新问题。
func (h *QuestionHandler) Update(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateQuestionRequest
	if !bindJSON(c, &req) {
		return
	}
	question, err := h.questionSvc.Update(actor, id, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "question.update", "question", question.ID,
		"更新采访问题", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgQuestionUpdated, question)
}

// Delete 删除问题。
func (h *QuestionHandler) Delete(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.questionSvc.Delete(actor, id); err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "question.delete", "question", id,
		"删除采访问题", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgQuestionDeleted, nil)
}
