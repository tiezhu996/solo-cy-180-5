package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	return db, mock
}

func TestUserRepositoryCreate(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &model.User{Username: "alice", DisplayName: "Alice", Role: "interviewer"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 1 {
		t.Fatalf("id = %d, want 1", user.ID)
	}
}

func TestUserRepositoryFindByUsernameNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ?")).
		WithArgs("nobody", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByUsername("nobody"); err == nil {
		t.Fatalf("expected not found error")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepositoryFindByUsernameFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserRepository(db)
	rows := sqlmock.NewRows([]string{"id", "username", "display_name", "role"}).
		AddRow(1, "alice", "Alice", "interviewer")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ?")).
		WithArgs("alice", 1).
		WillReturnRows(rows)

	user, err := repo.FindByUsername("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "alice" || user.Role != "interviewer" {
		t.Fatalf("unexpected user: %+v", user)
	}
}
