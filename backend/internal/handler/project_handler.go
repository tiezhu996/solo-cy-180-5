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

// ProjectHandler 采访项目接口处理器。
type ProjectHandler struct {
	projectSvc service.ProjectService
	auditSvc   service.AuditService
	logger     *slog.Logger
}

// NewProjectHandler 构造项目处理器。
func NewProjectHandler(projectSvc service.ProjectService, auditSvc service.AuditService, logger *slog.Logger) *ProjectHandler {
	return &ProjectHandler{projectSvc: projectSvc, auditSvc: auditSvc, logger: logger}
}

// Create 创建采访项目。
func (h *ProjectHandler) Create(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	var req dto.CreateProjectRequest
	if !bindJSON(c, &req) {
		return
	}
	project, err := h.projectSvc.Create(actor, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "project.create", "project", project.ID,
		"创建采访项目 "+project.Title, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgProjectCreated, project)
}

// Get 查询项目详情。
func (h *ProjectHandler) Get(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	project, err := h.projectSvc.Get(id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, project)
}

// List 项目列表（可按状态筛选）。
func (h *ProjectHandler) List(c *gin.Context) {
	var p dto.PageParams
	if !bindQuery(c, &p) {
		return
	}
	p.Normalize()
	projects, total, err := h.projectSvc.List(p.Page, p.PageSize, c.Query("status"))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"list": projects, "total": total, "page": p.Page, "page_size": p.PageSize})
}

// ListMine 我的项目列表。
func (h *ProjectHandler) ListMine(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	var p dto.PageParams
	if !bindQuery(c, &p) {
		return
	}
	p.Normalize()
	projects, total, err := h.projectSvc.ListMine(actor.ID, p.Page, p.PageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"list": projects, "total": total, "page": p.Page, "page_size": p.PageSize})
}

// Update 更新项目基本信息。
func (h *ProjectHandler) Update(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateProjectRequest
	if !bindJSON(c, &req) {
		return
	}
	project, err := h.projectSvc.Update(actor, id, &req)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "project.update", "project", project.ID,
		"更新采访项目 "+project.Title, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgProjectUpdated, project)
}

// TransitionStatus 项目状态流转。
func (h *ProjectHandler) TransitionStatus(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateProjectStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	project, err := h.projectSvc.TransitionStatus(actor, id, req.Status)
	if err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "project.status", "project", project.ID,
		"项目状态流转为 "+req.Status, c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgProjectStatusOK, project)
}

// Delete 删除项目。
func (h *ProjectHandler) Delete(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.projectSvc.Delete(actor, id); err != nil {
		c.Error(err)
		return
	}
	h.auditSvc.Record(actor.ID, actor.Username, actor.Role, "project.delete", "project", id,
		"删除采访项目", c.ClientIP(), middleware.RequestID(c))
	util.OKMessage(c, constants.MsgProjectDeleted, nil)
}

// Stats 项目统计。
func (h *ProjectHandler) Stats(c *gin.Context) {
	stats, err := h.projectSvc.Stats()
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, stats)
}
