package handler

import (
	"encoding/json"
	"net/http"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/middleware"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type NoteHandler struct {
	service  *service.NoteService
	validate *validator.Validate
}

func NewNoteHandler(service *service.NoteService) *NoteHandler {
	return &NoteHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(r)

	note, err := h.service.Create(userID, &req)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create note"})
		return
	}

	response.JSON(w, http.StatusCreated, note)
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	notes, err := h.service.List(userID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list notes"})
		return
	}

	response.JSON(w, http.StatusOK, notes)
}

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]
	if noteID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Note ID is required"})
		return
	}

	userID := middleware.GetUserID(r)

	note, err := h.service.GetByID(userID, noteID)
	if err != nil {
		if err.Error() == "unauthorized: note does not belong to user" {
			response.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		// Assuming not found error from repo would be generic, but ideally check for "not found"
		response.JSON(w, http.StatusNotFound, map[string]string{"error": "Note not found"})
		return
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]
	if noteID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Note ID is required"})
		return
	}

	var req domain.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	userID := middleware.GetUserID(r)

	note, err := h.service.Update(userID, noteID, &req)
	if err != nil {
		if err.Error() == "unauthorized: note does not belong to user" {
			response.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		// Check if it's a conflict error
		if conflictErr, ok := err.(*service.ConflictError); ok {
			response.JSON(w, http.StatusConflict, map[string]interface{}{
				"error":    "version_conflict",
				"conflict": conflictErr.Conflict,
			})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update note"})
		return
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]
	if noteID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Note ID is required"})
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.Delete(userID, noteID); err != nil {
		if err.Error() == "unauthorized: note does not belong to user" {
			response.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete note"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Note deleted successfully"})
}
