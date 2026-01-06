package domain

import "time"

type NoteVersion struct {
	ID               string    `json:"id"`
	NoteID           string    `json:"note_id"`
	Version          int64     `json:"version"`
	EncryptedContent string    `json:"encrypted_content"`
	EncryptedTitle   string    `json:"encrypted_title"`
	ContentHash      string    `json:"content_hash"`
	DeviceID         string    `json:"device_id"`
	CreatedAt        time.Time `json:"created_at"`
}
