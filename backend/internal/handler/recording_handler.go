package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/middleware"
	"github.com/oralhistory/oralhistory/internal/service"
	"github.com/oralhistory/oralhistory/internal/util"
)

// RecordingHandler 录音片段接口处理器，只负责参数解析与响应，业务编排由 RecordingService 完成。
type RecordingHandler struct {
	recordingSvc service.RecordingService
	auditSvc     service.AuditService
	logger       *slog.Logger
}

// NewRecordingHandler 构造录音处理器。
func NewRecordingHandler(recordingSvc service.RecordingService, auditSvc service.AuditService, logger *slog.Logger) *RecordingHandler {
	return &RecordingHandler{recordingSvc: recordingSvc, auditSvc: auditSvc, logger: logger}
}

// Create 创建录音记录。
func (h *RecordingHandler) Create(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	var req dto.CreateRecordingRequest
	if !bindJSON(c, &req) {
		return
	}
	recording, err := h.recordingSvc.Create(actor, &req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OKMessage(c, constants.MsgOK, recording)
}

// Get 查询录音详情。
func (h *RecordingHandler) Get(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	recording, err := h.recordingSvc.Get(id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, recording)
}

// List 录音列表（project_id 或 question_id 二选一，复用同一 service 方法）。
func (h *RecordingHandler) List(c *gin.Context) {
	var projectID, questionID uint
	if raw := c.Query("project_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			projectID = uint(v)
		}
	}
	if raw := c.Query("question_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			questionID = uint(v)
		}
	}
	if projectID == 0 && questionID == 0 {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "录音列表查询必须提供 project_id 或 question_id")
		return
	}
	recordings, err := h.recordingSvc.List(projectID, questionID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"list": recordings})
}

// Update 更新录音摘要/时长/状态。
func (h *RecordingHandler) Update(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateRecordingRequest
	if !bindJSON(c, &req) {
		return
	}
	recording, err := h.recordingSvc.Update(actor, id, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "recording.update", "recording", recording.ID,
		"更新录音信息", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgRecordingUpdated, recording)
}

// UpdateSummary 更新录音一句话摘要。
func (h *RecordingHandler) UpdateSummary(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Summary string `json:"summary" binding:"required,max=512"`
	}
	if !bindJSON(c, &req) {
		return
	}
	recording, err := h.recordingSvc.UpdateSummary(actor, id, req.Summary)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "recording.summary", "recording", recording.ID,
		"更新录音摘要 "+req.Summary, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgRecordingUpdated, recording)
}

// UploadAudio 上传录音文件并关联到录音记录；上传编排在业务层完成。
func (h *RecordingHandler) UploadAudio(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "录音文件上传失败: 缺少 file 字段")
		return
	}
	defer file.Close()

	duration, _ := strconv.Atoi(c.PostForm("duration_seconds"))
	recording, err := h.recordingSvc.UploadAudio(c.Request.Context(), actor, id, header.Filename,
		file, header.Size, header.Header.Get("Content-Type"), duration)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "recording.upload", "recording", recording.ID,
		"上传录音文件 "+recording.AudioKey, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgRecordingUploaded, recording)
}

// PlayAudio 流式返回录音音频；取流与校验在业务层完成，这里只负责拷贝到响应体。
func (h *RecordingHandler) PlayAudio(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	recording, obj, err := h.recordingSvc.OpenAudio(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	defer obj.Close()
	c.Header("Content-Type", "audio/webm")
	c.Header("Content-Disposition", "inline; filename=recording_"+strconv.FormatUint(uint64(recording.ID), 10)+".webm")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, obj)
}

// Delete 删除录音并清理其音频文件（清理编排在业务层）。
func (h *RecordingHandler) Delete(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.recordingSvc.Delete(actor, id); err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "recording.delete", "recording", id,
		"删除录音", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgRecordingDeleted, nil)
}
