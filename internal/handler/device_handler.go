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

type DeviceHandler struct {
	service  *service.DeviceService
	validate *validator.Validate
}

func NewDeviceHandler(service *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(r)

	device, err := h.service.Register(userID, &req)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to register device"})
		return
	}

	response.JSON(w, http.StatusCreated, device)
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	devices, err := h.service.List(userID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list devices"})
		return
	}

	response.JSON(w, http.StatusOK, devices)
}

func (h *DeviceHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["id"]
	if deviceID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "Device ID is required"})
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.Revoke(userID, deviceID); err != nil {
		if err.Error() == "unauthorized: device does not belong to user" {
			response.JSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to revoke device"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Device revoked successfully"})
}
