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

type ManifestEntry struct {
	ID          string    `json:"id"`
	ContentHash string    `json:"content_hash"`
	Version     int64     `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsDeleted   bool      `json:"is_deleted"`
}

type ManifestResponse struct {
	Notes    []ManifestEntry `json:"notes"`
	SyncTime time.Time       `json:"sync_time"`
}

type BatchDiffRequest struct {
	WorkspaceID string          `json:"workspace_id" validate:"required"`
	DeviceID    string          `json:"device_id" validate:"required"`
	LocalNotes  []LocalNoteInfo `json:"local_notes"`
}

type LocalNoteInfo struct {
	ID          string `json:"id"`
	ContentHash string `json:"content_hash"`
	Version     int64  `json:"version"`
}

type BatchDiffResponse struct {
	ToDownload []NoteResponse `json:"to_download"`
	ToUpload   []string       `json:"to_upload"`
	ToDelete   []string       `json:"to_delete"`
	Conflicts  []ConflictInfo `json:"conflicts"`
	SyncTime   time.Time      `json:"sync_time"`
}

type ConflictInfo struct {
	NoteID        string `json:"note_id"`
	LocalHash     string `json:"local_hash"`
	ServerHash    string `json:"server_hash"`
	LocalVersion  int64  `json:"local_version"`
	ServerVersion int64  `json:"server_version"`
}
