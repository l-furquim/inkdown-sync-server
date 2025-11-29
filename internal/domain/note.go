package domain

import "time"

// NoteType defines if the node is a file or directory
type NoteType string

const (
	NoteTypeFile      NoteType = "file"
	NoteTypeDirectory NoteType = "directory"
)

// Note represents a markdown note or directory in the system.
// Supports End-to-End Encryption (E2EE) by storing encrypted blobs for sensitive data.
type Note struct {
	ID       string   `json:"id"`
	UserID   string   `json:"user_id"`
	ParentID *string  `json:"parent_id"` // Nullable for root items
	Type     NoteType `json:"type"`      // "file" or "directory"

	// E2EE Fields
	// The server treats these as opaque strings. The client handles encryption/decryption.
	EncryptedTitle   string `json:"encrypted_title"`
	EncryptedContent string `json:"encrypted_content,omitempty"` // Empty for directories
	EncryptionAlgo   string `json:"encryption_algo"`             // Ex: "AES-256-GCM"
	Nonce            string `json:"nonce"`                       // IV/Nonce used for encryption

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"` // Soft delete for sync
	Version   int64     `json:"version"`    // Monotonic counter for sync conflict resolution
}

// CreateNoteRequest represents the payload to create a new note/directory
type CreateNoteRequest struct {
	ParentID         *string  `json:"parent_id"`
	Type             NoteType `json:"type" validate:"required,oneof=file directory"`
	EncryptedTitle   string   `json:"encrypted_title" validate:"required"`
	EncryptedContent string   `json:"encrypted_content"` // Optional for directories
	EncryptionAlgo   string   `json:"encryption_algo" validate:"required"`
	Nonce            string   `json:"nonce" validate:"required"`
}

// UpdateNoteRequest represents the payload to update a note
type UpdateNoteRequest struct {
	EncryptedTitle   *string `json:"encrypted_title"`
	EncryptedContent *string `json:"encrypted_content"`
	EncryptionAlgo   *string `json:"encryption_algo"`
	Nonce            *string `json:"nonce"`
	ParentID         *string `json:"parent_id"` // For moving files
	IsDeleted        *bool   `json:"is_deleted"`
}

// NoteResponse represents the note data returned to the client
type NoteResponse struct {
	ID               string    `json:"id"`
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
}
