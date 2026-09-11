package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/util"
)

// fakeRecordingRepo 录音仓储的内存实现，仅用于业务层单测。
type fakeRecordingRepo struct {
	recordings map[uint]*model.Recording
	deletedIDs []uint
	forUpdate  bool
	updateErr  error
}

func (f *fakeRecordingRepo) Create(recording *model.Recording) error {
	f.recordings[recording.ID] = recording
	return nil
}
func (f *fakeRecordingRepo) FindByID(id uint) (*model.Recording, error) {
	if r, ok := f.recordings[id]; ok {
		return r, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeRecordingRepo) ListByProject(projectID uint) ([]model.Recording, error) {
	return nil, nil
}
func (f *fakeRecordingRepo) ListByQuestion(questionID uint) ([]model.Recording, error) {
	return nil, nil
}
func (f *fakeRecordingRepo) FindByIDForUpdate(id uint) (*model.Recording, error) {
	f.forUpdate = true
	return f.FindByID(id)
}
func (f *fakeRecordingRepo) Update(recording *model.Recording) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.recordings[recording.ID] = recording
	return nil
}
func (f *fakeRecordingRepo) UpdateStatus(recording *model.Recording) error {
	return f.Update(recording)
}
func (f *fakeRecordingRepo) Delete(id uint) error {
	f.deletedIDs = append(f.deletedIDs, id)
	delete(f.recordings, id)
	return nil
}
func (f *fakeRecordingRepo) CountByProject(projectID uint) (int64, error) {
	return 0, nil
}

// fakeStorage 对象存储的内存实现，记录调用与对象内容。
type fakeStorage struct {
	objects     map[string][]byte
	removed     []string
	uploadErr   error
	getErr      error
	removeErr   error
	removeFailN int // Remove 前 N 次返回 removeErr，之后恢复成功（模拟存储短暂故障后恢复）
	removeCalls int
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: map[string][]byte{}}
}
func (f *fakeStorage) Upload(_ context.Context, objectKey string, reader io.Reader, _ int64, _ string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	f.objects[objectKey] = data
	return nil
}
func (f *fakeStorage) Get(_ context.Context, objectKey string) (io.ReadCloser, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	data, ok := f.objects[objectKey]
	if !ok {
		return nil, errors.New("object not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// Remove 模拟 S3/MinIO 删除：对象不存在也视为删除成功（删除幂等）。
func (f *fakeStorage) Remove(_ context.Context, objectKey string) error {
	f.removeCalls++
	f.removed = append(f.removed, objectKey)
	if f.removeFailN > 0 {
		f.removeFailN--
		if f.removeErr != nil {
			return f.removeErr
		}
		return errors.New("remove failed")
	}
	delete(f.objects, objectKey)
	return nil
}

func newTestRecordingService(repo repository.RecordingRepository, storage StorageService) RecordingService {
	return NewRecordingService(repo, nil, nil, storage, slog.Default())
}

func TestRecordingServiceUploadAudio(t *testing.T) {
	actor := &model.User{ID: 7, Username: "interviewer", Role: constants.RoleInterviewer}

	t.Run("success uploads object and attaches recording", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			3: {ID: 3, ProjectID: 1, QuestionID: 2, Status: constants.RecordingStatusRecording},
		}}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		payload := []byte("audio-bytes")
		recording, err := svc.UploadAudio(context.Background(), actor, 3, "clip.MP3",
			bytes.NewReader(payload), int64(len(payload)), "audio/mpeg", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		key := recording.AudioKey
		if !strings.HasPrefix(key, "recordings/7/3_") || !strings.HasSuffix(key, ".mp3") {
			t.Fatalf("object key = %s, want recordings/7/3_<rand>.mp3", key)
		}
		if got := storage.objects[key]; string(got) != string(payload) {
			t.Fatalf("stored object = %q, want %q", got, payload)
		}
		if recording.Status != constants.RecordingStatusReady {
			t.Fatalf("status = %s, want ready", recording.Status)
		}
		if recording.DurationSeconds != 42 {
			t.Fatalf("duration = %d, want 42", recording.DurationSeconds)
		}
		if !repo.forUpdate {
			t.Fatalf("expected SELECT FOR UPDATE path used")
		}
		if len(storage.removed) != 0 {
			t.Fatalf("expected no object removal, got %v", storage.removed)
		}
	})

	t.Run("empty filename defaults to webm extension", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			3: {ID: 3, Status: constants.RecordingStatusRecording},
		}}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		recording, err := svc.UploadAudio(context.Background(), actor, 3, "",
			strings.NewReader("x"), 1, "audio/webm", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasSuffix(recording.AudioKey, ".webm") {
			t.Fatalf("object key = %s, want .webm suffix", recording.AudioKey)
		}
	})

	t.Run("storage failure returns internal error without attach", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			3: {ID: 3, Status: constants.RecordingStatusRecording},
		}}
		storage := newFakeStorage()
		storage.uploadErr = errors.New("minio down")
		svc := newTestRecordingService(repo, storage)

		_, err := svc.UploadAudio(context.Background(), actor, 3, "a.webm",
			strings.NewReader("x"), 1, "audio/webm", 0)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeInternal {
			t.Fatalf("expected internal app error, got %v", err)
		}
		if repo.recordings[3].AudioKey != "" {
			t.Fatalf("recording should not be attached on upload failure")
		}
	})

	t.Run("attach failure rolls back uploaded object", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			3: {ID: 3, Status: constants.RecordingStatusRecording},
		}, updateErr: errors.New("db locked")}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		_, err := svc.UploadAudio(context.Background(), actor, 3, "a.webm",
			strings.NewReader("x"), 1, "audio/webm", 0)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeInternal {
			t.Fatalf("expected internal app error, got %v", err)
		}
		if len(storage.removed) != 1 {
			t.Fatalf("expected rollback removal of 1 object, got %v", storage.removed)
		}
		if len(storage.objects) != 0 {
			t.Fatalf("expected orphan object removed, remaining %v", storage.objects)
		}
	})

	t.Run("recording missing returns not found and cleans object", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{}}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		_, err := svc.UploadAudio(context.Background(), actor, 99, "a.webm",
			strings.NewReader("x"), 1, "audio/webm", 0)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeNotFound {
			t.Fatalf("expected not found app error, got %v", err)
		}
		if len(storage.removed) != 1 {
			t.Fatalf("expected uploaded object rolled back, got removed=%v", storage.removed)
		}
	})
}

func TestRecordingServiceOpenAudio(t *testing.T) {
	cases := []struct {
		name        string
		recording   *model.Recording
		storageGet  error
		wantCode    int
		wantContent string
	}{
		{name: "success streams object", recording: &model.Recording{ID: 1, AudioKey: "k1"}, wantContent: "voice"},
		{name: "missing recording", recording: nil, wantCode: constants.CodeNotFound},
		{name: "recording without audio", recording: &model.Recording{ID: 1, AudioKey: ""}, wantCode: constants.CodeNotFound},
		{name: "storage failure", recording: &model.Recording{ID: 1, AudioKey: "k1"}, storageGet: errors.New("minio down"), wantCode: constants.CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{}}
			storage := newFakeStorage()
			if tc.recording != nil {
				repo.recordings[tc.recording.ID] = tc.recording
			}
			storage.objects["k1"] = []byte("voice")
			storage.getErr = tc.storageGet
			svc := newTestRecordingService(repo, storage)

			recording, stream, err := svc.OpenAudio(context.Background(), 1)
			if tc.wantCode != 0 {
				var appErr *util.AppError
				if !asAppError(err, &appErr) || appErr.Code != tc.wantCode {
					t.Fatalf("expected app error code %d, got %v", tc.wantCode, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer stream.Close()
			if recording.AudioKey != "k1" {
				t.Fatalf("audio key = %s, want k1", recording.AudioKey)
			}
			data, _ := io.ReadAll(stream)
			if string(data) != tc.wantContent {
				t.Fatalf("stream = %q, want %q", data, tc.wantContent)
			}
		})
	}
}

func TestRecordingServiceDeleteCleansObject(t *testing.T) {
	actor := &model.User{ID: 7, Username: "interviewer", Role: constants.RoleInterviewer}

	t.Run("deletes object before db row", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			5: {ID: 5, AudioKey: "recordings/7/5/a.webm"},
		}}
		storage := newFakeStorage()
		storage.objects["recordings/7/5/a.webm"] = []byte("x")
		svc := newTestRecordingService(repo, storage)

		if err := svc.Delete(context.Background(), actor, 5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 5 {
			t.Fatalf("deleted ids = %v, want [5]", repo.deletedIDs)
		}
		if len(storage.removed) != 1 || storage.removed[0] != "recordings/7/5/a.webm" {
			t.Fatalf("removed objects = %v", storage.removed)
		}
		if _, ok := storage.objects["recordings/7/5/a.webm"]; ok {
			t.Fatalf("audio object should be removed from storage")
		}
	})

	t.Run("recording without audio skips removal", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			6: {ID: 6, AudioKey: ""},
		}}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		if err := svc.Delete(context.Background(), actor, 6); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(storage.removed) != 0 {
			t.Fatalf("expected no removal, got %v", storage.removed)
		}
	})

	t.Run("missing recording returns not found", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{}}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		err := svc.Delete(context.Background(), actor, 404)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
		if len(repo.deletedIDs) != 0 || len(storage.removed) != 0 {
			t.Fatalf("expected no side effects, deletes=%v removes=%v", repo.deletedIDs, storage.removed)
		}
	})

	t.Run("cleanup failure keeps row so retry converges", func(t *testing.T) {
		key := "recordings/7/5/a.webm"
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			5: {ID: 5, AudioKey: key},
		}}
		storage := newFakeStorage()
		storage.objects[key] = []byte("x")
		// 第一次删除对象失败，模拟存储抖动；之后恢复。
		storage.removeErr = errors.New("minio down")
		storage.removeFailN = 1
		svc := newTestRecordingService(repo, storage)

		// 第一次：返回失败，但记录必须保留、对象仍在。
		err := svc.Delete(context.Background(), actor, 5)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeInternal {
			t.Fatalf("first delete: expected internal error, got %v", err)
		}
		if len(repo.deletedIDs) != 0 {
			t.Fatalf("db row must be kept on object removal failure, deletes=%v", repo.deletedIDs)
		}
		if _, ok := repo.recordings[5]; !ok {
			t.Fatalf("recording row must remain for retry")
		}
		if _, ok := storage.objects[key]; !ok {
			t.Fatalf("audio object must remain after failed removal")
		}

		// 第二次重试：存储已恢复，对象删除（幂等）后删除记录，最终一致。
		if err := svc.Delete(context.Background(), actor, 5); err != nil {
			t.Fatalf("retry delete: unexpected error: %v", err)
		}
		if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 5 {
			t.Fatalf("retry: deleted ids = %v, want [5]", repo.deletedIDs)
		}
		if _, ok := repo.recordings[5]; ok {
			t.Fatalf("recording row should be gone after retry")
		}
		if _, ok := storage.objects[key]; ok {
			t.Fatalf("audio object should be gone after retry")
		}
		if storage.removeCalls != 2 {
			t.Fatalf("expected 2 remove calls (fail then success), got %d", storage.removeCalls)
		}
	})

	t.Run("retry succeeds even when object already vanished", func(t *testing.T) {
		// 上次可能已删掉对象、但删记录前崩溃；记录仍在。对象删除幂等，重试应能继续删掉记录。
		key := "recordings/7/8/a.webm"
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			8: {ID: 8, AudioKey: key},
		}}
		storage := newFakeStorage() // 对象本就不存在
		svc := newTestRecordingService(repo, storage)

		if err := svc.Delete(context.Background(), actor, 8); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 8 {
			t.Fatalf("deleted ids = %v, want [8]", repo.deletedIDs)
		}
		if len(storage.removed) != 1 || storage.removed[0] != key {
			t.Fatalf("removed = %v, want [%s]", storage.removed, key)
		}
	})
}
