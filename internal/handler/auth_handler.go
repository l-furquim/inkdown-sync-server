package handler

import (
	"encoding/json"
	"net/http"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := h.authService.Register(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, map[string]string{
		"message": "User registered successfully. Please login.",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	loginResp, err := h.authService.Login(&req)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.Success(w, loginResp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tokenResp, err := h.authService.RefreshToken(&req)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.Success(w, tokenResp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"message": "Logged out successfully",
	})
}
