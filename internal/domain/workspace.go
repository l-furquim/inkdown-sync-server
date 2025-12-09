package domain

import "time"

// Workspace represents a remote workspace that can be synced with a local directory
type Workspace struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDefault bool      `json:"is_default"`
}

// CreateWorkspaceRequest represents the payload to create a new workspace
type CreateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateWorkspaceRequest represents the payload to update a workspace
type UpdateWorkspaceRequest struct {
	Name string `json:"name,omitempty"`
}

// WorkspaceResponse represents the workspace data returned to the client
type WorkspaceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDefault bool      `json:"is_default"`
	NoteCount int       `json:"note_count,omitempty"`
}
