package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/middleware"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/service"
	"github.com/oralhistory/oralhistory/internal/util"
)

// stubRecordingService 仅覆写删除编排，其余方法沿用接口零值（删除接口不会触达）。
type stubRecordingService struct {
	service.RecordingService

	mu       sync.Mutex
	calls    int
	deleted  bool
	failN    int // 前 failN 次删除返回对象清理失败，之后成功
	observer func(ctx context.Context, id uint)
}

func (s *stubRecordingService) Delete(ctx context.Context, _ *model.User, id uint) error {
	if s.observer != nil {
		s.observer(ctx, id)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.deleted {
		// 记录已被并发请求先删走，后到者拿到 not found。
		return util.NewAppError(constants.CodeNotFound, "录音已不存在", nil)
	}
	if s.failN > 0 {
		s.failN--
		// 与真实编排一致：存储清理失败包成 5xx 业务错误，且记录仍保留以便重试。
		return util.NewAppError(constants.CodeInternal, "录音对象清理失败", errors.New("minio remove failed"))
	}
	s.deleted = true
	return nil
}

// stubAuditService 记录成功删除后的审计埋点次数。
type stubAuditService struct {
	service.AuditService
	mu      sync.Mutex
	records int
}

func (s *stubAuditService) Record(_ uint, _, _, _, _ string, _ uint, _, _, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records++
}

func newDeleteEngine(svc service.RecordingService) (*gin.Engine, *stubAuditService) {
	gin.SetMode(gin.TestMode)
	audit := &stubAuditService{}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyClaims, &util.Claims{UserID: 7, Username: "interviewer", Role: constants.RoleInterviewer})
		c.Next()
	}, middleware.ErrorHandler(slog.Default()))
	h := NewRecordingHandler(svc, audit, slog.Default())
	r.DELETE("/api/v1/recordings/:id", h.Delete)
	return r, audit
}

func doDelete(r http.Handler, ctx context.Context, id string) *httptest.ResponseRecorder {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/recordings/"+id, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

type deleteResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func TestRecordingHandlerDeleteRetryAfterCleanupFailure(t *testing.T) {
	svc := &stubRecordingService{failN: 1}
	r, audit := newDeleteEngine(svc)
	ctx := context.Background()

	// 第一次：对象清理失败，接口返回统一 5xx，记录仍在，且不写成功审计。
	w1 := doDelete(r, ctx, "5")
	var body1 deleteResp
	if err := json.Unmarshal(w1.Body.Bytes(), &body1); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if w1.Code != http.StatusInternalServerError || body1.Code != constants.CodeInternal {
		t.Fatalf("first delete = (%d, code=%d msg=%s), want (500, %d)", w1.Code, body1.Code, body1.Message, constants.CodeInternal)
	}
	if audit.records != 0 {
		t.Fatalf("failed delete must not be audited as success, got %d records", audit.records)
	}

	// 第二次重试：清理成功、记录删除，返回统一成功结构并补写审计。
	w2 := doDelete(r, ctx, "5")
	var body2 deleteResp
	if err := json.Unmarshal(w2.Body.Bytes(), &body2); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if w2.Code != http.StatusOK || body2.Code != constants.CodeOK || body2.Message != constants.MsgRecordingDeleted {
		t.Fatalf("retry delete = (%d, code=%d msg=%q), want (200, %d, %q)",
			w2.Code, body2.Code, body2.Message, constants.CodeOK, constants.MsgRecordingDeleted)
	}
	if svc.calls != 2 {
		t.Fatalf("service calls = %d, want 2", svc.calls)
	}
	if audit.records != 1 {
		t.Fatalf("successful delete audited %d times, want 1", audit.records)
	}
}

func TestRecordingHandlerDeleteConcurrent(t *testing.T) {
	svc := &stubRecordingService{}
	r, audit := newDeleteEngine(svc)

	const n = 8
	var wg sync.WaitGroup
	statuses := make([]int, n)
	codes := make([]int, n)
	start := make(chan struct{})
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start // 尽量让请求同时发出，竞争结果由 stub 内的锁决定，不依赖调度时序
			w := doDelete(r, context.Background(), "9")
			statuses[i] = w.Code
			var body deleteResp
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			codes[i] = body.Code
		}(i)
	}
	close(start)
	wg.Wait()

	var ok, notFound int
	for i := 0; i < n; i++ {
		switch codes[i] {
		case constants.CodeOK:
			if statuses[i] != http.StatusOK {
				t.Fatalf("winner http status = %d, want 200", statuses[i])
			}
			ok++
		case constants.CodeNotFound:
			if statuses[i] != http.StatusNotFound {
				t.Fatalf("loser http status = %d, want 404", statuses[i])
			}
			notFound++
		default:
			t.Fatalf("unexpected response code %d (http %d)", codes[i], statuses[i])
		}
	}
	if ok != 1 {
		t.Fatalf("exactly one delete should succeed, got %d", ok)
	}
	if notFound != n-1 {
		t.Fatalf("remaining %d requests should be not found, got %d", n-1, notFound)
	}
	if audit.records != 1 {
		t.Fatalf("exactly one success audit expected, got %d", audit.records)
	}
}

func TestRecordingHandlerDeletePropagatesCancellation(t *testing.T) {
	var sawCanceled bool
	svc := &stubRecordingService{observer: func(ctx context.Context, _ uint) {
		sawCanceled = errors.Is(ctx.Err(), context.Canceled)
	}}
	// 请求取消时，业务层与真实编排一致地把底层取消包成 5xx 统一错误。
	svc.failN = 1 << 30
	r, audit := newDeleteEngine(svc)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := doDelete(r, ctx, "5")

	if !sawCanceled {
		t.Fatalf("handler must pass the request context to the service; ctx was not canceled")
	}
	var body deleteResp
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if w.Code != http.StatusInternalServerError || body.Code != constants.CodeInternal {
		t.Fatalf("canceled delete = (%d, code=%d), want (500, %d)", w.Code, body.Code, constants.CodeInternal)
	}
	if audit.records != 0 {
		t.Fatalf("canceled delete must not be audited, got %d records", audit.records)
	}
}
