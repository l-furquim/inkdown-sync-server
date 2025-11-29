package handler

import (
	"encoding/json"
	"net/http"

	"inkdown-sync-server/internal/middleware"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.Success(w, user)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Username == "" {
		response.BadRequest(w, "Username is required")
		return
	}

	user, err := h.userService.UpdateUsername(userID, req.Username)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, user)
}
