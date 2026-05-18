package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// --- Mocks for QueueSyncHandler ---

type mockSyncJobRepoForHandler struct {
	jobs    map[uuid.UUID]*entity.SyncJob
	created []*entity.SyncJob
}

func newMockSyncJobRepoForHandler() *mockSyncJobRepoForHandler {
	return &mockSyncJobRepoForHandler{jobs: make(map[uuid.UUID]*entity.SyncJob)}
}

func (m *mockSyncJobRepoForHandler) Create(_ context.Context, job *entity.SyncJob) error {
	m.jobs[job.ID] = job
	m.created = append(m.created, job)
	return nil
}
func (m *mockSyncJobRepoForHandler) FindByID(_ context.Context, id uuid.UUID) (*entity.SyncJob, error) {
	if j, ok := m.jobs[id]; ok {
		return j, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockSyncJobRepoForHandler) FindByStatus(_ context.Context, status entity.SyncJobStatus) ([]*entity.SyncJob, error) {
	var result []*entity.SyncJob
	for _, j := range m.jobs {
		if j.Status == status {
			result = append(result, j)
		}
	}
	return result, nil
}
func (m *mockSyncJobRepoForHandler) FindActiveByAppIDAndType(_ context.Context, appID uuid.UUID, jobType string) (*entity.SyncJob, error) {
	for _, j := range m.jobs {
		if j.AppID == appID && j.JobType == jobType && !j.IsTerminal() {
			return j, nil
		}
	}
	return nil, nil
}
func (m *mockSyncJobRepoForHandler) FindByParentJobID(_ context.Context, _ uuid.UUID) ([]*entity.SyncJob, error) {
	return nil, nil
}
func (m *mockSyncJobRepoForHandler) ListByAppID(_ context.Context, appID uuid.UUID, _ string, _ string, limit, offset int) ([]*entity.SyncJob, int, error) {
	var result []*entity.SyncJob
	for _, j := range m.jobs {
		if j.AppID == appID {
			result = append(result, j)
		}
	}
	total := len(result)
	if offset >= len(result) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}
func (m *mockSyncJobRepoForHandler) UpdateStatus(_ context.Context, id uuid.UUID, status entity.SyncJobStatus) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = status
	}
	return nil
}
func (m *mockSyncJobRepoForHandler) UpdateProgress(_ context.Context, id uuid.UUID, total, completed int) error {
	if j, ok := m.jobs[id]; ok {
		j.TotalItems = total
		j.CompletedItems = completed
	}
	return nil
}
func (m *mockSyncJobRepoForHandler) MarkStarted(_ context.Context, id uuid.UUID, workerID string) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusProcessing
		j.WorkerID = workerID
		now := time.Now().UTC()
		j.StartedAt = &now
	}
	return nil
}
func (m *mockSyncJobRepoForHandler) MarkCompleted(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusCompleted
		now := time.Now().UTC()
		j.CompletedAt = &now
	}
	return nil
}
func (m *mockSyncJobRepoForHandler) MarkFailed(_ context.Context, id uuid.UUID, errMsg string) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusFailed
		j.ErrorMessage = errMsg
		now := time.Now().UTC()
		j.CompletedAt = &now
	}
	return nil
}
func (m *mockSyncJobRepoForHandler) MarkPendingIfProcessing(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		if j.Status == entity.SyncJobStatusProcessing {
			j.Status = entity.SyncJobStatusPending
			j.WorkerID = ""
			j.StartedAt = nil
		}
	}
	return nil
}

type mockAppRepoForQueueSync struct {
	apps map[uuid.UUID]*entity.App
}

func (m *mockAppRepoForQueueSync) Create(_ context.Context, _ *entity.App) error { return nil }
func (m *mockAppRepoForQueueSync) FindByID(_ context.Context, id uuid.UUID) (*entity.App, error) {
	if a, ok := m.apps[id]; ok {
		return a, nil
	}
	return nil, persistence.ErrAppNotFound
}
func (m *mockAppRepoForQueueSync) FindByPartnerAccountID(_ context.Context, _ uuid.UUID) ([]*entity.App, error) {
	return nil, nil
}
func (m *mockAppRepoForQueueSync) FindByPartnerAppID(_ context.Context, _ uuid.UUID, _ string) (*entity.App, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockAppRepoForQueueSync) FindAllByPartnerAppID(_ context.Context, _ string) ([]*entity.App, error) {
	return nil, nil
}
func (m *mockAppRepoForQueueSync) Update(_ context.Context, _ *entity.App) error { return nil }
func (m *mockAppRepoForQueueSync) Delete(_ context.Context, _ uuid.UUID) error   { return nil }

type mockPartnerRepoForQueueSync struct {
	accounts map[uuid.UUID]*entity.PartnerAccount // keyed by UserID
}

func (m *mockPartnerRepoForQueueSync) Create(_ context.Context, _ *entity.PartnerAccount) error {
	return nil
}
func (m *mockPartnerRepoForQueueSync) FindByID(_ context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, persistence.ErrPartnerAccountNotFound
}
func (m *mockPartnerRepoForQueueSync) FindByUserID(_ context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	if a, ok := m.accounts[userID]; ok {
		return a, nil
	}
	return nil, persistence.ErrPartnerAccountNotFound
}
func (m *mockPartnerRepoForQueueSync) FindByOrgID(_ context.Context, _ uuid.UUID) (*entity.PartnerAccount, error) {
	return nil, persistence.ErrPartnerAccountNotFound
}
func (m *mockPartnerRepoForQueueSync) FindByPartnerID(_ context.Context, _ string) (*entity.PartnerAccount, error) {
	return nil, persistence.ErrPartnerAccountNotFound
}
func (m *mockPartnerRepoForQueueSync) Update(_ context.Context, _ *entity.PartnerAccount) error {
	return nil
}
func (m *mockPartnerRepoForQueueSync) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockPartnerRepoForQueueSync) GetAllIDs(_ context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

// --- Test helpers ---

func setupHandlerTest(t *testing.T) (
	*QueueSyncHandler,
	*mockSyncJobRepoForHandler,
	*mockAppRepoForQueueSync,
	*mockPartnerRepoForQueueSync,
	*entity.User,
	*entity.PartnerAccount,
	*entity.App,
) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	syncJobRepo := newMockSyncJobRepoForHandler()
	appRepo := &mockAppRepoForQueueSync{apps: make(map[uuid.UUID]*entity.App)}
	partnerRepo := &mockPartnerRepoForQueueSync{accounts: make(map[uuid.UUID]*entity.PartnerAccount)}

	userID := uuid.New()
	partnerAccountID := uuid.New()
	appID := uuid.New()

	user := &entity.User{ID: userID, Email: "test@test.com", Role: valueobject.RoleOwner}
	partner := &entity.PartnerAccount{
		ID:     partnerAccountID,
		UserID: userID,
	}
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
		Name:             "Test App",
	}

	appRepo.apps[appID] = app
	partnerRepo.accounts[userID] = partner

	lockManager := queue.NewLockManager(client)
	progressTracker := queue.NewProgressTracker(client, syncJobRepo, 0, 0)

	queueSyncService := service.NewQueueSyncService(
		syncJobRepo, appRepo, partnerRepo, client, lockManager, progressTracker,
	)

	handler := NewQueueSyncHandler(queueSyncService, partnerRepo, appRepo)

	return handler, syncJobRepo, appRepo, partnerRepo, user, partner, app
}

func makeRequest(t *testing.T, handler http.HandlerFunc, method, path string, user *entity.User, chiParams map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)

	// Set user context
	if user != nil {
		ctx := middleware.SetUserContext(req.Context(), user)
		req = req.WithContext(ctx)
	}

	// Set chi URL params
	if len(chiParams) > 0 {
		rctx := chi.NewRouteContext()
		for k, v := range chiParams {
			rctx.URLParams.Add(k, v)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// --- Tests ---

func TestQueueSync_EnqueueSync_Success(t *testing.T) {
	h, syncJobRepo, _, _, user, _, app := setupHandlerTest(t)

	rr := makeRequest(t, h.EnqueueSync, "POST",
		"/api/v1/sync/enqueue/"+app.ID.String()+"?type=full",
		user, map[string]string{"appID": app.ID.String()})

	if rr.Code != http.StatusAccepted {
		t.Fatalf("Expected 202, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp["job_type"] != entity.SyncJobTypeFullSync {
		t.Errorf("Expected job_type %q, got %q", entity.SyncJobTypeFullSync, resp["job_type"])
	}
	if resp["status"] != string(entity.SyncJobStatusPending) {
		t.Errorf("Expected status pending, got %q", resp["status"])
	}

	// Verify job was created in repo
	if len(syncJobRepo.created) != 1 {
		t.Fatalf("Expected 1 created job, got %d", len(syncJobRepo.created))
	}
}

func TestQueueSync_EnqueueSync_Duplicate(t *testing.T) {
	h, syncJobRepo, _, _, user, partner, app := setupHandlerTest(t)

	// Pre-create an active job
	existingJob := entity.NewSyncJob(app.ID, user.ID, partner.ID, entity.SyncJobTypeFullSync, 0)
	syncJobRepo.jobs[existingJob.ID] = existingJob

	rr := makeRequest(t, h.EnqueueSync, "POST",
		"/api/v1/sync/enqueue/"+app.ID.String()+"?type=full",
		user, map[string]string{"appID": app.ID.String()})

	if rr.Code != http.StatusConflict {
		t.Fatalf("Expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestQueueSync_EnqueueSync_Unauthorized(t *testing.T) {
	h, _, _, _, _, _, app := setupHandlerTest(t)

	rr := makeRequest(t, h.EnqueueSync, "POST",
		"/api/v1/sync/enqueue/"+app.ID.String(),
		nil, map[string]string{"appID": app.ID.String()})

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d", rr.Code)
	}
}

func TestQueueSync_EnqueueSync_ForbiddenApp(t *testing.T) {
	h, _, appRepo, _, _, _, _ := setupHandlerTest(t)

	// Create a different user
	otherUser := &entity.User{ID: uuid.New(), Email: "other@test.com", Role: valueobject.RoleOwner}

	// App belongs to a different partner
	otherApp := &entity.App{
		ID:               uuid.New(),
		PartnerAccountID: uuid.New(), // Different partner
		Name:             "Other App",
	}
	appRepo.apps[otherApp.ID] = otherApp

	rr := makeRequest(t, h.EnqueueSync, "POST",
		"/api/v1/sync/enqueue/"+otherApp.ID.String(),
		otherUser, map[string]string{"appID": otherApp.ID.String()})

	// Should fail because otherUser has no partner account
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 (no partner), got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestQueueSync_GetJobStatus_Success(t *testing.T) {
	h, syncJobRepo, _, _, user, partner, app := setupHandlerTest(t)

	job := entity.NewSyncJob(app.ID, user.ID, partner.ID, entity.SyncJobTypeTransactionSync, 0)
	syncJobRepo.jobs[job.ID] = job

	rr := makeRequest(t, h.GetJobStatus, "GET",
		"/api/v1/sync/jobs/"+job.ID.String(),
		user, map[string]string{"jobID": job.ID.String()})

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp["id"] != job.ID.String() {
		t.Errorf("Expected job ID %s, got %s", job.ID, resp["id"])
	}
}

func TestQueueSync_GetJobStatus_TenantIsolation(t *testing.T) {
	h, syncJobRepo, _, _, _, partner, app := setupHandlerTest(t)

	// Job belongs to a different user
	otherUserID := uuid.New()
	job := entity.NewSyncJob(app.ID, otherUserID, partner.ID, entity.SyncJobTypeTransactionSync, 0)
	syncJobRepo.jobs[job.ID] = job

	// Try to access with the test user (different from job's user)
	testUser := &entity.User{ID: uuid.New(), Email: "test@test.com", Role: valueobject.RoleOwner}
	rr := makeRequest(t, h.GetJobStatus, "GET",
		"/api/v1/sync/jobs/"+job.ID.String(),
		testUser, map[string]string{"jobID": job.ID.String()})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("Expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestQueueSync_GetJobProgress_Success(t *testing.T) {
	h, syncJobRepo, _, _, user, partner, app := setupHandlerTest(t)

	job := entity.NewSyncJob(app.ID, user.ID, partner.ID, entity.SyncJobTypeTransactionSync, 0)
	job.TotalItems = 100
	job.CompletedItems = 50
	syncJobRepo.jobs[job.ID] = job

	rr := makeRequest(t, h.GetJobProgress, "GET",
		"/api/v1/sync/jobs/"+job.ID.String()+"/progress",
		user, map[string]string{"jobID": job.ID.String()})

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)

	if int(resp["total"].(float64)) != 100 {
		t.Errorf("Expected total 100, got %v", resp["total"])
	}
	if int(resp["completed"].(float64)) != 50 {
		t.Errorf("Expected completed 50, got %v", resp["completed"])
	}
}

func TestQueueSync_ListJobs_Success(t *testing.T) {
	h, syncJobRepo, _, _, user, partner, app := setupHandlerTest(t)

	// Create a few jobs
	for i := 0; i < 3; i++ {
		j := entity.NewSyncJob(app.ID, user.ID, partner.ID, entity.SyncJobTypeTransactionSync, 0)
		j.Status = entity.SyncJobStatusCompleted
		syncJobRepo.jobs[j.ID] = j
	}

	rr := makeRequest(t, h.ListJobs, "GET",
		"/api/v1/sync/jobs?app_id="+app.ID.String()+"&limit=10",
		user, nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)

	jobs := resp["jobs"].([]interface{})
	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(jobs))
	}
	if int(resp["total"].(float64)) != 3 {
		t.Errorf("Expected total 3, got %v", resp["total"])
	}
}

func TestQueueSync_ListJobs_MissingAppID(t *testing.T) {
	h, _, _, _, user, _, _ := setupHandlerTest(t)

	rr := makeRequest(t, h.ListJobs, "GET",
		"/api/v1/sync/jobs",
		user, nil)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d", rr.Code)
	}
}

func TestQueueSync_CancelJob_Success(t *testing.T) {
	h, syncJobRepo, _, _, user, partner, app := setupHandlerTest(t)

	job := entity.NewSyncJob(app.ID, user.ID, partner.ID, entity.SyncJobTypeTransactionSync, 0)
	job.Status = entity.SyncJobStatusProcessing
	syncJobRepo.jobs[job.ID] = job

	rr := makeRequest(t, h.CancelJob, "POST",
		"/api/v1/sync/jobs/"+job.ID.String()+"/cancel",
		user, map[string]string{"jobID": job.ID.String()})

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify job was cancelled
	if job.Status != entity.SyncJobStatusCancelled {
		t.Errorf("Expected cancelled, got %s", job.Status)
	}
}

func TestQueueSync_CancelJob_AlreadyCompleted(t *testing.T) {
	h, syncJobRepo, _, _, user, partner, app := setupHandlerTest(t)

	job := entity.NewSyncJob(app.ID, user.ID, partner.ID, entity.SyncJobTypeTransactionSync, 0)
	job.Status = entity.SyncJobStatusCompleted
	syncJobRepo.jobs[job.ID] = job

	rr := makeRequest(t, h.CancelJob, "POST",
		"/api/v1/sync/jobs/"+job.ID.String()+"/cancel",
		user, map[string]string{"jobID": job.ID.String()})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestQueueSync_EnqueueSync_TypeMapping(t *testing.T) {
	tests := []struct {
		queryType   string
		wantJobType string
	}{
		{"full", entity.SyncJobTypeFullSync},
		{"transaction", entity.SyncJobTypeTransactionSync},
		{"snapshot", entity.SyncJobTypeSnapshotSync},
		{"event", entity.SyncJobTypeEventSync},
		{"review", entity.SyncJobTypeReviewSync},
		{"", entity.SyncJobTypeFullSync}, // default
	}

	for _, tt := range tests {
		t.Run(tt.queryType, func(t *testing.T) {
			h, _, _, _, user, _, app := setupHandlerTest(t)

			path := "/api/v1/sync/enqueue/" + app.ID.String()
			if tt.queryType != "" {
				path += "?type=" + tt.queryType
			}

			rr := makeRequest(t, h.EnqueueSync, "POST", path,
				user, map[string]string{"appID": app.ID.String()})

			if rr.Code != http.StatusAccepted {
				t.Fatalf("Expected 202, got %d: %s", rr.Code, rr.Body.String())
			}

			var resp map[string]interface{}
			json.NewDecoder(rr.Body).Decode(&resp)

			if resp["job_type"] != tt.wantJobType {
				t.Errorf("Expected job_type %q, got %q", tt.wantJobType, resp["job_type"])
			}
		})
	}
}

// Suppress unused import warning
var _ repository.SyncJobRepository = (*mockSyncJobRepoForHandler)(nil)
