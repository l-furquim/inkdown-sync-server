package handler

import (
	"encoding/json"
	"net/http"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/middleware"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"

	"github.com/go-playground/validator/v10"
)

type SecurityHandler struct {
	service  *service.SecurityService
	validate *validator.Validate
}

func NewSecurityHandler(service *service.SecurityService) *SecurityHandler {
	return &SecurityHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *SecurityHandler) UploadKey(w http.ResponseWriter, r *http.Request) {
	var req domain.UploadKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.UploadKey(userID, &req); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to upload key"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Key uploaded successfully"})
}

func (h *SecurityHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	key, err := h.service.GetKey(userID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{"error": "Key not found"})
		return
	}

	response.JSON(w, http.StatusOK, key)
}
