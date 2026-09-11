package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/handler"
	"github.com/oralhistory/oralhistory/internal/middleware"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/service"
	"github.com/oralhistory/oralhistory/internal/util"
)

// ---- 跨层替身：录音仓储 ----

type flowRecordingRepo struct {
	mu         sync.Mutex
	recordings map[uint]*model.Recording
	deleteCnt  int
}

func (r *flowRecordingRepo) Create(rec *model.Recording) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recordings[rec.ID] = rec
	return nil
}
func (r *flowRecordingRepo) FindByID(id uint) (*model.Recording, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rec, ok := r.recordings[id]; ok {
		cp := *rec
		return &cp, nil
	}
	return nil, repository.ErrNotFound
}
func (r *flowRecordingRepo) ListByProject(uint) ([]model.Recording, error) {
	return nil, nil
}
func (r *flowRecordingRepo) ListByQuestion(uint) ([]model.Recording, error) {
	return nil, nil
}
func (r *flowRecordingRepo) FindByIDForUpdate(id uint) (*model.Recording, error) {
	return r.FindByID(id)
}
func (r *flowRecordingRepo) Update(rec *model.Recording) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recordings[rec.ID] = rec
	return nil
}
func (r *flowRecordingRepo) UpdateStatus(rec *model.Recording) error { return r.Update(rec) }
func (r *flowRecordingRepo) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleteCnt++
	delete(r.recordings, id)
	return nil
}
func (r *flowRecordingRepo) CountByProject(uint) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.recordings)), nil
}
func (r *flowRecordingRepo) exists(id uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.recordings[id]
	return ok
}

// ---- 跨层替身：对象存储（尊重 ctx 取消，可编排前 N 次失败）----

type flowStorage struct {
	mu         sync.Mutex
	objects    map[string]struct{}
	removeCnt  int
	removeFail int // 前 removeFail 次 Remove 失败（在取消判断之后），之后恢复
}

func (s *flowStorage) Upload(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	if _, err := io.Copy(io.Discard, r); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = struct{}{}
	return nil
}
func (s *flowStorage) Get(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("flow storage Get not used in delete tests")
}
func (s *flowStorage) Remove(ctx context.Context, key string) error {
	// 与真实 MinIO 客户端一致：请求上下文取消时删除中断并返回错误。
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeCnt++
	if s.removeFail > 0 {
		s.removeFail--
		return errors.New("minio remove failed")
	}
	delete(s.objects, key)
	return nil
}
func (s *flowStorage) has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.objects[key]
	return ok
}

// ---- 跨层替身：审计仓储（配合真实 auditService）----

type flowAuditRepo struct {
	mu   sync.Mutex
	logs []model.AuditLog
}

func (a *flowAuditRepo) Create(l *model.AuditLog) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logs = append(a.logs, *l)
	return nil
}
func (a *flowAuditRepo) List(int, int, string) ([]model.AuditLog, int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]model.AuditLog(nil), a.logs...), int64(len(a.logs)), nil
}
func (a *flowAuditRepo) deleteAuditsFor(id uint) []model.AuditLog {
	a.mu.Lock()
	defer a.mu.Unlock()
	var got []model.AuditLog
	for _, l := range a.logs {
		if l.Action == "recording.delete" && l.EntityID == id {
			got = append(got, l)
		}
	}
	return got
}

// ---- 装配：真实 HTTP 中间件 + 真实业务编排 + 真实审计，仅仓储/对象存储为替身 ----

func newCleanupFlow(t *testing.T, recRepo repository.RecordingRepository, storage service.StorageService, auditRepo repository.AuditLogRepository) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recSvc := service.NewRecordingService(recRepo, nil, nil, storage, slog.Default())
	auditSvc := service.NewAuditService(auditRepo, slog.Default())

	r := gin.New()
	r.Use(
		middleware.RequestIDMiddleware(),
		func(c *gin.Context) {
			c.Set(middleware.ContextKeyClaims, &util.Claims{UserID: 7, Username: "interviewer", Role: constants.RoleInterviewer})
			c.Next()
		},
		middleware.ErrorHandler(slog.Default()),
	)
	h := handler.NewRecordingHandler(recSvc, auditSvc, slog.Default())
	r.DELETE("/api/v1/recordings/:id", h.Delete)
	return r
}

func flowDelete(engine http.Handler, ctx context.Context, id string) (int, cleanupBody) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/recordings/"+id, nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	var body cleanupBody
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

type cleanupBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

const flowAudioKey = "recordings/7/5/a.webm"

func newFlowDeps() (*flowRecordingRepo, *flowStorage, *flowAuditRepo) {
	recRepo := &flowRecordingRepo{recordings: map[uint]*model.Recording{
		5: {ID: 5, ProjectID: 1, QuestionID: 2, AudioKey: flowAudioKey, Status: constants.RecordingStatusReady},
	}}
	storage := &flowStorage{objects: map[string]struct{}{flowAudioKey: {}}}
	auditRepo := &flowAuditRepo{}
	return recRepo, storage, auditRepo
}

// 场景一：对象清理失败后重试，响应 / 记录 / 对象 / 审计最终一致。
func TestRecordingCleanupFlow_RetryAfterFailure(t *testing.T) {
	recRepo, storage, auditRepo := newFlowDeps()
	storage.removeFail = 1 // 第一次删除对象失败，之后恢复
	engine := newCleanupFlow(t, recRepo, storage, auditRepo)
	ctx := context.Background()

	// 第一次：对象清理失败 -> 500 统一错误；记录与对象都保留；不产生删除审计。
	status, body := flowDelete(engine, ctx, "5")
	if status != http.StatusInternalServerError || body.Code != constants.CodeInternal {
		t.Fatalf("first delete = (%d, code=%d msg=%q), want (500, %d)", status, body.Code, body.Message, constants.CodeInternal)
	}
	if body.Message != "录音 5 音频文件清理失败" {
		t.Fatalf("first delete message = %q", body.Message)
	}
	if !recRepo.exists(5) {
		t.Fatalf("recording row must remain after cleanup failure")
	}
	if !storage.has(flowAudioKey) {
		t.Fatalf("audio object must remain after cleanup failure")
	}
	if len(auditRepo.deleteAuditsFor(5)) != 0 {
		t.Fatalf("failed cleanup must not write delete audit")
	}
	if storage.removeCnt != 1 {
		t.Fatalf("remove calls = %d, want 1", storage.removeCnt)
	}

	// 第二次重试：成功 -> 200；记录与对象都清掉；恰好一条删除审计。
	status, body = flowDelete(engine, ctx, "5")
	if status != http.StatusOK || body.Code != constants.CodeOK || body.Message != constants.MsgRecordingDeleted {
		t.Fatalf("retry delete = (%d, code=%d msg=%q), want (200, %d, %q)",
			status, body.Code, body.Message, constants.CodeOK, constants.MsgRecordingDeleted)
	}
	if recRepo.exists(5) {
		t.Fatalf("recording row must be gone after retry")
	}
	if storage.has(flowAudioKey) {
		t.Fatalf("audio object must be gone after retry")
	}
	audits := auditRepo.deleteAuditsFor(5)
	if len(audits) != 1 {
		t.Fatalf("delete audits = %d, want 1", len(audits))
	}
	if audits[0].Action != "recording.delete" || audits[0].EntityType != "recording" || audits[0].Username != "interviewer" {
		t.Fatalf("audit content mismatch: %+v", audits[0])
	}
	if storage.removeCnt != 2 || recRepo.deleteCnt != 1 {
		t.Fatalf("calls mismatch: remove=%d want 2, repoDelete=%d want 1", storage.removeCnt, recRepo.deleteCnt)
	}
}

// 场景二：重复请求——首次成功后再次删除，第二次得到 404 且无额外副作用/审计。
func TestRecordingCleanupFlow_DuplicateRequests(t *testing.T) {
	recRepo, storage, auditRepo := newFlowDeps()
	engine := newCleanupFlow(t, recRepo, storage, auditRepo)
	ctx := context.Background()

	status, body := flowDelete(engine, ctx, "5")
	if status != http.StatusOK || body.Code != constants.CodeOK {
		t.Fatalf("first delete = (%d, code=%d), want (200, 0)", status, body.Code)
	}

	// 重复删除：记录已不存在 -> 404；不再触达对象存储；审计仍只有一条。
	status, body = flowDelete(engine, ctx, "5")
	if status != http.StatusNotFound || body.Code != constants.CodeNotFound {
		t.Fatalf("duplicate delete = (%d, code=%d msg=%q), want (404, %d)",
			status, body.Code, body.Message, constants.CodeNotFound)
	}
	if body.Message != "录音 5 不存在" {
		t.Fatalf("duplicate delete message = %q", body.Message)
	}
	if storage.removeCnt != 1 {
		t.Fatalf("duplicate request must not call storage remove again, got %d", storage.removeCnt)
	}
	if recRepo.deleteCnt != 1 {
		t.Fatalf("repo delete count = %d, want 1", recRepo.deleteCnt)
	}
	if len(auditRepo.deleteAuditsFor(5)) != 1 {
		t.Fatalf("delete audits = %d, want 1", len(auditRepo.deleteAuditsFor(5)))
	}
}

// 场景三：请求取消——上下文已取消时删除中断，返回统一错误且记录/对象/审计均不变。
func TestRecordingCleanupFlow_RequestCancel(t *testing.T) {
	recRepo, storage, auditRepo := newFlowDeps()
	engine := newCleanupFlow(t, recRepo, storage, auditRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	status, body := flowDelete(engine, ctx, "5")
	if status != http.StatusInternalServerError || body.Code != constants.CodeInternal {
		t.Fatalf("canceled delete = (%d, code=%d msg=%q), want (500, %d)",
			status, body.Code, body.Message, constants.CodeInternal)
	}
	if !recRepo.exists(5) {
		t.Fatalf("recording row must remain after cancellation")
	}
	if !storage.has(flowAudioKey) {
		t.Fatalf("audio object must remain after cancellation")
	}
	if recRepo.deleteCnt != 0 {
		t.Fatalf("repo delete must not be reached on cancellation, got %d", recRepo.deleteCnt)
	}
	if len(auditRepo.deleteAuditsFor(5)) != 0 {
		t.Fatalf("canceled delete must not write audit")
	}
}
