package domain

import "time"

type CLIToken struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Token       string     `json:"token"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      []string   `json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP  string     `json:"last_used_ip,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	IsRevoked   bool       `json:"is_revoked"`
}

type CLITokenPublic struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      []string   `json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	IsRevoked   bool       `json:"is_revoked"`
}

type CreateCLITokenRequest struct {
	Name   string   `json:"name" validate:"required,min=1,max=100"`
	Scopes []string `json:"scopes"`
}

type CreateCLITokenResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Token       string    `json:"token"`
	TokenPrefix string    `json:"token_prefix"`
	Scopes      []string  `json:"scopes"`
	CreatedAt   time.Time `json:"created_at"`
	Message     string    `json:"message"`
}

type CLILoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
}

type ValidateCLITokenRequest struct {
	Token string `json:"token" validate:"required"`
}

const (
	ScopeThemePublish  = "theme:publish"
	ScopeThemeDelete   = "theme:delete"
	ScopeThemeUpdate   = "theme:update"
	ScopePluginPublish = "plugin:publish"
	ScopePluginDelete  = "plugin:delete"
	ScopePluginUpdate  = "plugin:update"
)

func DefaultCLIScopes() []string {
	return []string{
		ScopeThemePublish,
		ScopeThemeDelete,
		ScopeThemeUpdate,
		ScopePluginPublish,
		ScopePluginDelete,
		ScopePluginUpdate,
	}
}

func (t *CLIToken) ToPublic() *CLITokenPublic {
	return &CLITokenPublic{
		ID:          t.ID,
		Name:        t.Name,
		TokenPrefix: t.TokenPrefix,
		Scopes:      t.Scopes,
		LastUsedAt:  t.LastUsedAt,
		CreatedAt:   t.CreatedAt,
		IsRevoked:   t.IsRevoked,
	}
}
