package handler

import (
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

// TimelineMarkerHandler 时间轴节点接口处理器。
type TimelineMarkerHandler struct {
	markerSvc service.TimelineMarkerService
	auditSvc  service.AuditService
	logger    *slog.Logger
}

// NewTimelineMarkerHandler 构造时间轴节点处理器。
func NewTimelineMarkerHandler(markerSvc service.TimelineMarkerService, auditSvc service.AuditService, logger *slog.Logger) *TimelineMarkerHandler {
	return &TimelineMarkerHandler{markerSvc: markerSvc, auditSvc: auditSvc, logger: logger}
}

// Create 标注时间轴关键节点。
func (h *TimelineMarkerHandler) Create(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	var req dto.CreateTimelineMarkerRequest
	if !bindJSON(c, &req) {
		return
	}
	marker, err := h.markerSvc.Create(actor, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "marker.create", "timeline_marker", marker.ID,
		"标注时间轴节点 "+marker.Label, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgMarkerCreated, marker)
}

// List 时间轴节点列表（project_id 或 recording_id 二选一，复用同一 service 方法）。
func (h *TimelineMarkerHandler) List(c *gin.Context) {
	var projectID, recordingID uint
	if raw := c.Query("project_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			projectID = uint(v)
		}
	}
	if raw := c.Query("recording_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			recordingID = uint(v)
		}
	}
	if projectID == 0 && recordingID == 0 {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "时间轴节点查询必须提供 project_id 或 recording_id")
		return
	}
	markers, err := h.markerSvc.List(projectID, recordingID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"list": markers})
}

// Update 更新时间轴节点。
func (h *TimelineMarkerHandler) Update(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateTimelineMarkerRequest
	if !bindJSON(c, &req) {
		return
	}
	marker, err := h.markerSvc.Update(actor, id, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "marker.update", "timeline_marker", marker.ID,
		"更新时间轴节点", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgMarkerUpdated, marker)
}

// Delete 删除时间轴节点。
func (h *TimelineMarkerHandler) Delete(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.markerSvc.Delete(actor, id); err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "marker.delete", "timeline_marker", id,
		"删除时间轴节点", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgMarkerDeleted, nil)
}
