package domain

import "time"

type SyncMetadata struct {
	UserID           string           `json:"user_id"`
	DeviceID         string           `json:"device_id"`
	LastSyncTime     time.Time        `json:"last_sync_time"`
	NoteVersions     map[string]int64 `json:"note_versions"`
	PendingConflicts []string         `json:"pending_conflicts"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type SyncRequest struct {
	DeviceID     string           `json:"device_id" validate:"required"`
	LastSyncTime time.Time        `json:"last_sync_time"`
	NoteVersions map[string]int64 `json:"note_versions"`
}

type SyncResponse struct {
	Changes  []*NoteChange `json:"changes"`
	SyncTime time.Time     `json:"sync_time"`
	HasMore  bool          `json:"has_more"`
}

type NoteChange struct {
	NoteID    string        `json:"note_id"`
	Operation string        `json:"operation"`
	Version   int64         `json:"version"`
	Note      *NoteResponse `json:"note,omitempty"`
}
