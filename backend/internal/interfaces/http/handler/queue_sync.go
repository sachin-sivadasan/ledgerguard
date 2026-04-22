package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// QueueSyncHandler handles async queue-based sync endpoints
type QueueSyncHandler struct {
	queueSyncService *service.QueueSyncService
	partnerRepo      repository.PartnerAccountRepository
	appRepo          repository.AppRepository
}

// NewQueueSyncHandler creates a new queue sync handler
func NewQueueSyncHandler(
	queueSyncService *service.QueueSyncService,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *QueueSyncHandler {
	return &QueueSyncHandler{
		queueSyncService: queueSyncService,
		partnerRepo:      partnerRepo,
		appRepo:          appRepo,
	}
}

// EnqueueSync enqueues a new sync job
// POST /api/v1/sync/enqueue/{appID}?type=full&priority=normal
func (h *QueueSyncHandler) EnqueueSync(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appIDStr := chi.URLParam(r, "appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid app ID")
		return
	}

	// Tenant isolation
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	app, err := h.appRepo.FindByID(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	if app.PartnerAccountID != partnerAccount.ID {
		writeJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse job type (default: full_sync)
	jobType := r.URL.Query().Get("type")
	if jobType == "" {
		jobType = "full"
	}
	// Map short names to full job type constants
	switch jobType {
	case "full":
		jobType = entity.SyncJobTypeFullSync
	case "transaction":
		jobType = entity.SyncJobTypeTransactionSync
	case "snapshot":
		jobType = entity.SyncJobTypeSnapshotSync
	case "event":
		jobType = entity.SyncJobTypeEventSync
	case "status":
		jobType = entity.SyncJobTypeStatusSync
	case "store":
		jobType = entity.SyncJobTypeStoreSync
	case "review":
		jobType = entity.SyncJobTypeReviewSync
	}

	priority := 0
	if r.URL.Query().Get("priority") == "high" {
		priority = 1
	}

	job, err := h.queueSyncService.EnqueueSync(r.Context(), appID, user.ID, partnerAccount.ID, jobType, priority)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateJob) {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":   job.ID.String(),
		"job_type": job.JobType,
		"status":   string(job.Status),
		"message":  "Sync job enqueued",
	})
}

// GetJobStatus returns a job's current status
// GET /api/v1/sync/jobs/{jobID}
func (h *QueueSyncHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	job, err := h.queueSyncService.GetJobStatus(r.Context(), jobID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "job not found")
		return
	}

	// Tenant isolation: verify job belongs to user
	if job.UserID != user.ID {
		writeJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobToJSON(job))
}

// GetJobProgress returns detailed progress including Redis overlay and children
// GET /api/v1/sync/jobs/{jobID}/progress
func (h *QueueSyncHandler) GetJobProgress(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	progress, err := h.queueSyncService.GetJobProgress(r.Context(), jobID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "job not found")
		return
	}

	// Tenant isolation
	if progress.Job.UserID != user.ID {
		writeJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	resp := map[string]interface{}{
		"job":       jobToJSON(progress.Job),
		"total":     progress.Total,
		"completed": progress.Completed,
		"message":   progress.Message,
	}

	if len(progress.Children) > 0 {
		children := make([]map[string]interface{}, len(progress.Children))
		for i, child := range progress.Children {
			children[i] = map[string]interface{}{
				"job":       jobToJSON(child.Job),
				"total":     child.Total,
				"completed": child.Completed,
				"message":   child.Message,
			}
		}
		resp["children"] = children
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListJobs returns paginated job history for an app
// GET /api/v1/sync/jobs?app_id={appID}&status=&job_type=&limit=&offset=
func (h *QueueSyncHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appIDStr := r.URL.Query().Get("app_id")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app_id is required")
		return
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid app_id")
		return
	}

	// Tenant isolation
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	app, err := h.appRepo.FindByID(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	if app.PartnerAccountID != partnerAccount.ID {
		writeJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	status := r.URL.Query().Get("status")
	jobType := r.URL.Query().Get("job_type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	jobs, total, err := h.queueSyncService.ListJobs(r.Context(), appID, status, jobType, limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}

	jobsJSON := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		jobsJSON[i] = jobToJSON(job)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobsJSON,
		"total": total,
	})
}

// CancelJob requests cancellation of a running job
// POST /api/v1/sync/jobs/{jobID}/cancel
func (h *QueueSyncHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	job, err := h.queueSyncService.GetJobStatus(r.Context(), jobID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "job not found")
		return
	}

	if job.UserID != user.ID {
		writeJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.queueSyncService.CancelJob(r.Context(), jobID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Cancellation requested",
		"job_id":  jobID.String(),
	})
}

func jobToJSON(job *entity.SyncJob) map[string]interface{} {
	result := map[string]interface{}{
		"id":                 job.ID.String(),
		"app_id":             job.AppID.String(),
		"user_id":            job.UserID.String(),
		"partner_account_id": job.PartnerAccountID.String(),
		"job_type":           job.JobType,
		"status":             string(job.Status),
		"priority":           job.Priority,
		"total_items":        job.TotalItems,
		"completed_items":    job.CompletedItems,
		"entity_type":        job.EntityType,
		"error_message":      job.ErrorMessage,
		"worker_id":          job.WorkerID,
		"created_at":         job.CreatedAt,
		"updated_at":         job.UpdatedAt,
	}

	if job.ParentJobID != nil {
		result["parent_job_id"] = job.ParentJobID.String()
	}
	if job.StartedAt != nil {
		result["started_at"] = *job.StartedAt
	}
	if job.CompletedAt != nil {
		result["completed_at"] = *job.CompletedAt
	}

	return result
}
