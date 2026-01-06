package middleware

import (
	"context"
	"net/http"
	"strings"

	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/pkg/response"
)

func CLIAuthMiddleware(cliTokenService *service.CLITokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Authorization header required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Unauthorized(w, "Invalid authorization header format")
				return
			}

			token := parts[1]

			if !strings.HasPrefix(token, "ink_") {
				response.Unauthorized(w, "Invalid CLI token format. Expected ink_xxxxx")
				return
			}

			user, cliToken, err := cliTokenService.ValidateToken(token)
			if err != nil {
				response.Unauthorized(w, "Invalid or revoked CLI token")
				return
			}

			go func() {
				clientIP := getClientIPFromRequest(r)
				cliTokenService.UpdateLastUsed(cliToken.ID, clientIP)
			}()

			// Add user info to context
			ctx := context.WithValue(r.Context(), "user_id", user.ID)
			ctx = context.WithValue(ctx, "user", user)
			ctx = context.WithValue(ctx, "cli_token", cliToken)
			ctx = context.WithValue(ctx, "cli_token_id", cliToken.ID)
			ctx = context.WithValue(ctx, "cli_scopes", cliToken.Scopes)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CLIScopeMiddleware(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			scopes, ok := r.Context().Value("cli_scopes").([]string)
			if !ok {
				response.Forbidden(w, "CLI token scopes not found")
				return
			}

			hasScope := false
			for _, scope := range scopes {
				if scope == requiredScope {
					hasScope = true
					break
				}
			}

			if !hasScope {
				response.Forbidden(w, "CLI token does not have required scope: "+requiredScope)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIPFromRequest(r *http.Request) string {

	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}
