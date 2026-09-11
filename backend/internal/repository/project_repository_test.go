package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
)

func TestProjectRepositoryList(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewProjectRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `projects`")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	rows := sqlmock.NewRows([]string{"id", "title", "interviewee_name", "status"}).
		AddRow(1, "老城记忆", "王奶奶", "draft").
		AddRow(2, "渡江战役亲历", "张爷爷", "in_progress")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects`")).
		WillReturnRows(rows)

	projects, total, err := repo.List(1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(projects) != 2 {
		t.Fatalf("total=%d len=%d, want 2/2", total, len(projects))
	}
	if projects[1].Status != "in_progress" {
		t.Fatalf("status = %s, want in_progress", projects[1].Status)
	}
}

func TestProjectRepositoryFindByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewProjectRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects` WHERE `projects`.`id` = ?")).
		WithArgs(99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByID(99); err == nil {
		t.Fatalf("expected not found error")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProjectRepositoryUpdateStatus(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewProjectRepository(db)
	project := &model.Project{ID: 1, Status: "completed"}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `projects` SET `status`=?,`updated_at`=? WHERE `id` = ?")).
		WithArgs("completed", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.UpdateStatus(project); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
