package domain

import "time"

type ConflictType string

const (
	ConflictTypeUpdate ConflictType = "update"
	ConflictTypeDelete ConflictType = "delete"
)

type ResolutionStrategy string

const (
	ResolutionLWW    ResolutionStrategy = "lww"
	ResolutionServer ResolutionStrategy = "server"
	ResolutionClient ResolutionStrategy = "client"
	ResolutionManual ResolutionStrategy = "manual"
)

type Conflict struct {
	ID               string             `json:"id"`
	NoteID           string             `json:"note_id"`
	UserID           string             `json:"user_id"`
	Type             ConflictType       `json:"type"`
	BaseVersion      int64              `json:"base_version"`
	ServerVersion    int64              `json:"server_version"`
	ClientVersion    int64              `json:"client_version"`
	ServerNote       *Note              `json:"server_note"`
	ClientData       *UpdateNoteRequest `json:"client_data"`
	DeviceID         string             `json:"device_id"`
	DetectedAt       time.Time          `json:"detected_at"`
	ResolvedAt       *time.Time         `json:"resolved_at,omitempty"`
	ResolutionChoice ResolutionStrategy `json:"resolution_choice,omitempty"`
}

type ConflictResolutionRequest struct {
	Strategy ResolutionStrategy `json:"strategy" validate:"required,oneof=lww server client manual"`
	NoteData *UpdateNoteRequest `json:"note_data,omitempty"`
}
