// Package service 承载业务逻辑，依赖 repository 层。
package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/util"
)

// UserService 用户业务接口。
type UserService interface {
	Register(req *dto.RegisterRequest) (*model.User, error)
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	Me(userID uint) (*model.User, error)
	List(page, pageSize int) ([]model.User, int64, error)
	UpdateRole(userID uint, role string) error
	Delete(userID uint) error
}

type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
	logger   *slog.Logger
}

// NewUserService 构造用户服务。
func NewUserService(userRepo repository.UserRepository, cfg *config.Config, logger *slog.Logger) UserService {
	return &userService{userRepo: userRepo, cfg: cfg, logger: logger}
}

func (s *userService) Register(req *dto.RegisterRequest) (*model.User, error) {
	role := req.Role
	if role == "" {
		role = constants.RoleInterviewer
	}
	if !constants.ValidRoles(role) {
		return nil, util.NewAppError(constants.CodeValidation, fmt.Sprintf("用户角色 %s 不合法", role), nil)
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternal, "用户密码加密失败", err)
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		Role:         role,
	}
	if err := s.userRepo.Create(user); err != nil {
		if errors.Is(err, repository.ErrDuplicateName) {
			return nil, util.NewAppError(constants.CodeDuplicateName, fmt.Sprintf("用户名 %s 已存在", req.Username), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, "用户注册失败", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserRegister, user.Username, user.Role))
	return user, nil
}

func (s *userService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		s.logger.Info(fmt.Sprintf(constants.LogUserLogin, req.Username, false))
		return nil, util.NewAppError(constants.CodeLoginFailed, fmt.Sprintf("用户名 %s 或密码错误", req.Username), nil)
	}
	if !util.CheckPassword(user.PasswordHash, req.Password) {
		s.logger.Info(fmt.Sprintf(constants.LogUserLogin, req.Username, false))
		return nil, util.NewAppError(constants.CodeLoginFailed, fmt.Sprintf("用户名 %s 或密码错误", req.Username), nil)
	}
	token, err := util.GenerateToken(user.ID, user.Username, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpireH)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternal, "用户登录令牌生成失败", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserLogin, user.Username, true))
	return &dto.LoginResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *userService) Me(userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("用户 %d 不存在", userID), err)
	}
	return user, nil
}

func (s *userService) List(page, pageSize int) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternal, "用户列表查询失败", err)
	}
	return users, total, nil
}

func (s *userService) UpdateRole(userID uint, role string) error {
	if !constants.ValidRoles(role) {
		return util.NewAppError(constants.CodeValidation, fmt.Sprintf("用户角色 %s 不合法", role), nil)
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("用户 %d 不存在", userID), err)
	}
	user.Role = role
	if err := s.userRepo.Update(user); err != nil {
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("更新用户 %d 角色失败", userID), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserUpdate, "admin", user.Username))
	return nil
}

func (s *userService) Delete(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("用户 %d 不存在", userID), err)
	}
	if user.Role == constants.RoleAdmin {
		return util.NewAppError(constants.CodeForbidden, "管理员用户不可删除", nil)
	}
	if err := s.userRepo.Delete(userID); err != nil {
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("删除用户 %d 失败", userID), err)
	}
	return nil
}

func toUserResponse(u *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Role:        u.Role,
		CreatedAt:   util.FormatDateTime(u.CreatedAt),
	}
}
