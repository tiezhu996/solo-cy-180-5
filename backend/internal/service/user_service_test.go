package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/util"
)

type fakeUserRepo struct {
	users   []model.User
	created *model.User
	err     error
}

func (f *fakeUserRepo) Create(user *model.User) error {
	if f.err != nil {
		return f.err
	}
	for _, u := range f.users {
		if u.Username == user.Username {
			return repository.ErrDuplicateName
		}
	}
	f.created = user
	f.users = append(f.users, *user)
	return nil
}
func (f *fakeUserRepo) FindByID(id uint) (*model.User, error) {
	for i := range f.users {
		if f.users[i].ID == id {
			return &f.users[i], nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) FindByUsername(username string) (*model.User, error) {
	for i := range f.users {
		if f.users[i].Username == username {
			return &f.users[i], nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) List(page, pageSize int) ([]model.User, int64, error) {
	return f.users, int64(len(f.users)), nil
}
func (f *fakeUserRepo) Update(user *model.User) error { return nil }
func (f *fakeUserRepo) Delete(id uint) error          { return nil }

func newTestUserService(repo repository.UserRepository) *userService {
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpireH: 1}
	return &userService{userRepo: repo, cfg: cfg, logger: slog.Default()}
}

func TestUserServiceRegister(t *testing.T) {
	cases := []struct {
		name    string
		repo    repository.UserRepository
		req     *dto.RegisterRequest
		wantErr bool
		wantRole string
	}{
		{
			name: "valid interviewer default role",
			repo: &fakeUserRepo{},
			req:  &dto.RegisterRequest{Username: "alice", Password: "secret123", DisplayName: "Alice"},
			wantRole: constants.RoleInterviewer,
		},
		{
			name: "explicit archivist role",
			repo: &fakeUserRepo{},
			req:  &dto.RegisterRequest{Username: "bob", Password: "secret123", DisplayName: "Bob", Role: constants.RoleArchivist},
			wantRole: constants.RoleArchivist,
		},
		{
			name:    "invalid role rejected",
			repo:    &fakeUserRepo{},
			req:     &dto.RegisterRequest{Username: "eve", Password: "secret123", DisplayName: "Eve", Role: "root"},
			wantErr: true,
		},
		{
			name:    "duplicate username rejected",
			repo:    &fakeUserRepo{users: []model.User{{Username: "alice"}}},
			req:     &dto.RegisterRequest{Username: "alice", Password: "secret123", DisplayName: "Alice"},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestUserService(tc.repo)
			user, err := svc.Register(tc.req)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Role != tc.wantRole {
				t.Fatalf("role = %s, want %s", user.Role, tc.wantRole)
			}
			if util.CheckPassword(user.PasswordHash, tc.req.Password) == false {
				t.Fatalf("password hash mismatch")
			}
		})
	}
}

func TestUserServiceLogin(t *testing.T) {
	hash, _ := util.HashPassword("secret123")
	repo := &fakeUserRepo{users: []model.User{{ID: 1, Username: "alice", PasswordHash: hash, Role: constants.RoleInterviewer}}}
	svc := newTestUserService(repo)

	cases := []struct {
		name    string
		req     *dto.LoginRequest
		wantErr bool
	}{
		{name: "correct credentials", req: &dto.LoginRequest{Username: "alice", Password: "secret123"}},
		{name: "wrong password", req: &dto.LoginRequest{Username: "alice", Password: "wrong"}, wantErr: true},
		{name: "unknown user", req: &dto.LoginRequest{Username: "nobody", Password: "secret123"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.Login(tc.req)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Token == "" {
				t.Fatalf("token should not be empty")
			}
			claims, err := util.ParseToken(resp.Token, svc.cfg.JWTSecret)
			if err != nil || claims.UserID != 1 {
				t.Fatalf("token invalid: %v", err)
			}
		})
	}
}

func TestUserServiceMe(t *testing.T) {
	repo := &fakeUserRepo{users: []model.User{{ID: 1, Username: "alice", Role: constants.RoleAdmin}}}
	svc := newTestUserService(repo)
	if _, err := svc.Me(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Me(999); err == nil {
		t.Fatalf("expected not found error")
	} else {
		var appErr *util.AppError
		if !errors.As(err, &appErr) || appErr.Code != constants.CodeNotFound {
			t.Fatalf("expected not found app error, got %v", err)
		}
	}
}
