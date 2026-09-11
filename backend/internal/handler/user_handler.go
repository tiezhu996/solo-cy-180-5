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

// UserHandler 用户接口处理器。
type UserHandler struct {
	userSvc service.UserService
	logger  *slog.Logger
}

// NewUserHandler 构造用户处理器。
func NewUserHandler(userSvc service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{userSvc: userSvc, logger: logger}
}

// Register 用户注册。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	user, err := h.userSvc.Register(&req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OKMessage(c, constants.MsgRegisterSuccess, user)
}

// Login 用户登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.userSvc.Login(&req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OKMessage(c, constants.MsgLoginSuccess, resp)
}

// Me 当前用户信息。
func (h *UserHandler) Me(c *gin.Context) {
	actor, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}
	user, err := h.userSvc.Me(actor.ID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

// List 用户列表（管理员）。
func (h *UserHandler) List(c *gin.Context) {
	var p dto.PageParams
	if !bindQuery(c, &p) {
		return
	}
	p.Normalize()
	users, total, err := h.userSvc.List(p.Page, p.PageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"list": users, "total": total, "page": p.Page, "page_size": p.PageSize})
}

// UpdateRole 更新用户角色（管理员）。
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Role string `json:"role" binding:"required,oneof=admin interviewer archivist"`
	}
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.UpdateRole(id, req.Role); err != nil {
		c.Error(err)
		return
	}
	util.OKMessage(c, constants.MsgOK, nil)
}

// Delete 删除用户（管理员）。
func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.userSvc.Delete(id); err != nil {
		c.Error(err)
		return
	}
	util.OKMessage(c, constants.MsgOK, nil)
}
