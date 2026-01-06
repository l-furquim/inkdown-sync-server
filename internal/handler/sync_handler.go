package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/middleware"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"

	"github.com/gorilla/mux"
)

type SyncHandler struct {
	syncService     *service.SyncService
	conflictService *service.ConflictService
}

func NewSyncHandler(syncService *service.SyncService, conflictService *service.ConflictService) *SyncHandler {
	return &SyncHandler{
		syncService:     syncService,
		conflictService: conflictService,
	}
}

func (h *SyncHandler) ProcessSync(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.syncService.ProcessSyncRequest(userID, req.DeviceID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *SyncHandler) GetChanges(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sinceParam := r.URL.Query().Get("since")
	var since time.Time
	if sinceParam != "" {
		var err error
		since, err = time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid since parameter")
			return
		}
	}

	changes, err := h.syncService.GetChangesSince(userID, since)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"changes":   changes,
		"sync_time": time.Now(),
	})
}

func (h *SyncHandler) ListConflicts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conflicts, err := h.conflictService.ListByUser(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, conflicts)
}

func (h *SyncHandler) ResolveConflict(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	vars := mux.Vars(r)
	conflictID := vars["id"]

	var req domain.ConflictResolutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	conflict, err := h.conflictService.Get(conflictID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "conflict not found")
		return
	}

	if conflict.UserID != userID {
		response.Error(w, http.StatusForbidden, "unauthorized")
		return
	}

	note, err := h.conflictService.ApplyResolution(conflictID, req.Strategy, req.NoteData)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "conflict resolved",
		"note":    note,
	})
}

func (h *SyncHandler) GetManifest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workspaceID := r.URL.Query().Get("workspace_id")

	manifest, err := h.syncService.GetManifest(userID, workspaceID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, manifest)
}

func (h *SyncHandler) BatchDiff(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.BatchDiffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	diff, err := h.syncService.ProcessBatchDiff(userID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, diff)
}
