package domain

import "time"

type Workspace struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDefault bool      `json:"is_default"`
}

type CreateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateWorkspaceRequest struct {
	Name string `json:"name,omitempty"`
}

type WorkspaceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDefault bool      `json:"is_default"`
	NoteCount int       `json:"note_count,omitempty"`
}
