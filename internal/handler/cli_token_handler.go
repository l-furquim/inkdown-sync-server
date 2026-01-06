package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type CLITokenHandler struct {
	cliTokenService *service.CLITokenService
	validator       *validator.Validate
}

func NewCLITokenHandler(cliTokenService *service.CLITokenService) *CLITokenHandler {
	return &CLITokenHandler{
		cliTokenService: cliTokenService,
		validator:       validator.New(),
	}
}

func (h *CLITokenHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.CLILoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tokenResp, err := h.cliTokenService.LoginAndCreateToken(&req)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	clientIP := getClientIP(r)
	h.cliTokenService.UpdateLastUsed(tokenResp.ID, clientIP)

	response.Success(w, tokenResp)
}

func (h *CLITokenHandler) Validate(w http.ResponseWriter, r *http.Request) {
	token := extractCLIToken(r)
	if token == "" {
		response.Unauthorized(w, "CLI token required")
		return
	}

	user, cliToken, err := h.cliTokenService.ValidateToken(token)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	// Update last used
	clientIP := getClientIP(r)
	h.cliTokenService.UpdateLastUsed(cliToken.ID, clientIP)

	response.Success(w, map[string]interface{}{
		"valid":  true,
		"user":   user,
		"scopes": cliToken.Scopes,
	})
}

func (h *CLITokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req domain.CreateCLITokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tokenResp, err := h.cliTokenService.CreateToken(userID, &req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, tokenResp)
}

func (h *CLITokenHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	tokens, err := h.cliTokenService.ListTokens(userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	})
}

func (h *CLITokenHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	tokenID := mux.Vars(r)["id"]

	token, err := h.cliTokenService.GetToken(userID, tokenID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}

	response.Success(w, token)
}

func (h *CLITokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	tokenID := mux.Vars(r)["id"]

	if err := h.cliTokenService.RevokeToken(userID, tokenID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{
		"message": "Token revoked successfully",
	})
}

func (h *CLITokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	tokenID := mux.Vars(r)["id"]

	if err := h.cliTokenService.DeleteToken(userID, tokenID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{
		"message": "Token deleted successfully",
	})
}

func extractCLIToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	token := parts[1]
	if !strings.HasPrefix(token, "ink_") {
		return ""
	}

	return token
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
