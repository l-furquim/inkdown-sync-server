package websocket

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	TypeSyncRequest  MessageType = "sync_request"
	TypeSyncResponse MessageType = "sync_response"
	TypeNoteUpdate   MessageType = "note_update"
	TypeNoteDelete   MessageType = "note_delete"
	TypeConflict     MessageType = "conflict"
	TypeAck          MessageType = "ack"
	TypePing         MessageType = "ping"
	TypePong         MessageType = "pong"
)

type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type SyncRequestPayload struct {
	DeviceID     string           `json:"device_id"`
	LastSyncTime time.Time        `json:"last_sync_time"`
	NoteVersions map[string]int64 `json:"note_versions"`
}

type SyncResponsePayload struct {
	Changes  []NoteChange `json:"changes"`
	HasMore  bool         `json:"has_more"`
	SyncTime time.Time    `json:"sync_time"`
}

type NoteChange struct {
	NoteID    string          `json:"note_id"`
	Operation string          `json:"operation"`
	Version   int64           `json:"version"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type NoteUpdatePayload struct {
	NoteID           string    `json:"note_id"`
	Version          int64     `json:"version"`
	EncryptedTitle   string    `json:"encrypted_title"`
	EncryptedContent string    `json:"encrypted_content"`
	UpdatedAt        time.Time `json:"updated_at"`
	DeviceID         string    `json:"device_id"`
}

type NoteDeletePayload struct {
	NoteID   string `json:"note_id"`
	Version  int64  `json:"version"`
	DeviceID string `json:"device_id"`
}

type ConflictPayload struct {
	ConflictID    string          `json:"conflict_id"`
	NoteID        string          `json:"note_id"`
	ServerVersion int64           `json:"server_version"`
	ClientVersion int64           `json:"client_version"`
	ServerData    json.RawMessage `json:"server_data"`
}

type AckPayload struct {
	MessageID string `json:"message_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	var payloadBytes json.RawMessage
	if payload != nil {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		payloadBytes = bytes
	}

	return &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}, nil
}

func (m *Message) UnmarshalPayload(v interface{}) error {
	if m.Payload == nil {
		return nil
	}
	return json.Unmarshal(m.Payload, v)
}
