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
	objects   map[string][]byte
	removed   []string
	uploadErr error
	getErr    error
	removeErr error
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
func (f *fakeStorage) Remove(_ context.Context, objectKey string) error {
	f.removed = append(f.removed, objectKey)
	if f.removeErr != nil {
		return f.removeErr
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

	t.Run("deletes db row and audio object", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			5: {ID: 5, AudioKey: "recordings/7/5/a.webm"},
		}}
		storage := newFakeStorage()
		storage.objects["recordings/7/5/a.webm"] = []byte("x")
		svc := newTestRecordingService(repo, storage)

		if err := svc.Delete(actor, 5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 5 {
			t.Fatalf("deleted ids = %v, want [5]", repo.deletedIDs)
		}
		if len(storage.removed) != 1 || storage.removed[0] != "recordings/7/5/a.webm" {
			t.Fatalf("removed objects = %v", storage.removed)
		}
	})

	t.Run("recording without audio skips removal", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			6: {ID: 6, AudioKey: ""},
		}}
		storage := newFakeStorage()
		svc := newTestRecordingService(repo, storage)

		if err := svc.Delete(actor, 6); err != nil {
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

		err := svc.Delete(actor, 404)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
		if len(repo.deletedIDs) != 0 || len(storage.removed) != 0 {
			t.Fatalf("expected no side effects, deletes=%v removes=%v", repo.deletedIDs, storage.removed)
		}
	})

	t.Run("object cleanup failure surfaces internal error after db delete", func(t *testing.T) {
		repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{
			5: {ID: 5, AudioKey: "k"},
		}}
		storage := newFakeStorage()
		storage.removeErr = errors.New("minio down")
		svc := newTestRecordingService(repo, storage)

		err := svc.Delete(actor, 5)
		var appErr *util.AppError
		if !asAppError(err, &appErr) || appErr.Code != constants.CodeInternal {
			t.Fatalf("expected internal error, got %v", err)
		}
		if len(repo.deletedIDs) != 1 {
			t.Fatalf("db row should already be deleted")
		}
	})
}
