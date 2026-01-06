package domain

import "time"

type NoteType string

const (
	NoteTypeFile      NoteType = "file"
	NoteTypeDirectory NoteType = "directory"
)

type Note struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	WorkspaceID string   `json:"workspace_id"`
	ParentID    *string  `json:"parent_id"`
	Type        NoteType `json:"type"`

	EncryptedTitle   string `json:"encrypted_title"`
	EncryptedContent string `json:"encrypted_content,omitempty"`
	EncryptionAlgo   string `json:"encryption_algo"`
	Nonce            string `json:"nonce"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	IsDeleted      bool      `json:"is_deleted"`
	Version        int64     `json:"version"`
	ContentHash    string    `json:"content_hash"`
	LastEditDevice string    `json:"last_edit_device"`
}

type CreateNoteRequest struct {
	WorkspaceID      string   `json:"workspace_id" validate:"required"`
	ParentID         *string  `json:"parent_id"`
	Type             NoteType `json:"type" validate:"required,oneof=file directory"`
	EncryptedTitle   string   `json:"encrypted_title" validate:"required"`
	EncryptedContent string   `json:"encrypted_content"`
	EncryptionAlgo   string   `json:"encryption_algo" validate:"required"`
	Nonce            string   `json:"nonce" validate:"required"`
	ContentHash      string   `json:"content_hash"`
	DeviceID         string   `json:"device_id" validate:"required"`
}

type UpdateNoteRequest struct {
	EncryptedTitle   *string `json:"encrypted_title"`
	EncryptedContent *string `json:"encrypted_content"`
	EncryptionAlgo   *string `json:"encryption_algo"`
	Nonce            *string `json:"nonce"`
	ParentID         *string `json:"parent_id"`
	IsDeleted        *bool   `json:"is_deleted"`
	ExpectedVersion  *int64  `json:"expected_version"`
	ContentHash      *string `json:"content_hash"`
	DeviceID         string  `json:"device_id"`
}

type NoteResponse struct {
	ID               string    `json:"id"`
	WorkspaceID      string    `json:"workspace_id"`
	ParentID         *string   `json:"parent_id"`
	Type             NoteType  `json:"type"`
	EncryptedTitle   string    `json:"encrypted_title"`
	EncryptedContent string    `json:"encrypted_content,omitempty"`
	EncryptionAlgo   string    `json:"encryption_algo"`
	Nonce            string    `json:"nonce"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	IsDeleted        bool      `json:"is_deleted"`
	Version          int64     `json:"version"`
	ContentHash      string    `json:"content_hash"`
	LastEditDevice   string    `json:"last_edit_device"`
}
