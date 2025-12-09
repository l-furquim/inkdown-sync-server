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

// Login authenticates via email/password and creates a new CLI token
// This is the main endpoint for CLI authentication
// POST /api/v1/cli/login
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

	// Update last used with client IP
	clientIP := getClientIP(r)
	h.cliTokenService.UpdateLastUsed(tokenResp.ID, clientIP)

	response.Success(w, tokenResp)
}

// Validate checks if a CLI token is valid
// POST /api/v1/cli/validate
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

// Create creates a new CLI token (requires JWT authentication)
// POST /api/v1/cli/tokens
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

// List returns all CLI tokens for the authenticated user
// GET /api/v1/cli/tokens
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

// Get returns a specific CLI token
// GET /api/v1/cli/tokens/{id}
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

// Revoke revokes a CLI token (it remains in the database but is no longer valid)
// POST /api/v1/cli/tokens/{id}/revoke
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

// Delete permanently deletes a CLI token
// DELETE /api/v1/cli/tokens/{id}
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

// extractCLIToken extracts the CLI token from the Authorization header
// Expects: Authorization: Bearer ink_xxxxx
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
	// CLI tokens start with "ink_"
	if !strings.HasPrefix(token, "ink_") {
		return ""
	}

	return token
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
