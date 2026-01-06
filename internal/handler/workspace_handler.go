package handler

import (
	"encoding/json"
	"net/http"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/middleware"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"

	"github.com/gorilla/mux"
)

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
}

func NewWorkspaceHandler(workspaceService *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
	}
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspace, err := h.workspaceService.Create(userID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workspaces, err := h.workspaceService.List(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, workspaces)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	workspaceID := vars["id"]

	workspace, err := h.workspaceService.Get(userID, workspaceID)
	if err != nil {
		if err == service.ErrAccessDenied {
			response.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if err == service.ErrWorkspaceNotFound {
			response.Error(w, http.StatusNotFound, "workspace not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	workspaceID := vars["id"]

	var req domain.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspace, err := h.workspaceService.Update(userID, workspaceID, &req)
	if err != nil {
		if err == service.ErrAccessDenied {
			response.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if err == service.ErrWorkspaceNotFound {
			response.Error(w, http.StatusNotFound, "workspace not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	workspaceID := vars["id"]

	err := h.workspaceService.Delete(userID, workspaceID)
	if err != nil {
		if err == service.ErrAccessDenied {
			response.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if err == service.ErrWorkspaceNotFound {
			response.Error(w, http.StatusNotFound, "workspace not found")
			return
		}
		if err.Error() == "cannot delete default workspace" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "workspace deleted"})
}
