package service

import (
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"
	"inkdown-sync-server/internal/websocket"
)

type SyncService struct {
	noteRepo     repository.NoteRepository
	versionRepo  repository.NoteVersionRepository
	metadataRepo repository.SyncMetadataRepository
	wsManager    *websocket.Manager
}

func NewSyncService(
	noteRepo repository.NoteRepository,
	versionRepo repository.NoteVersionRepository,
	metadataRepo repository.SyncMetadataRepository,
	wsManager *websocket.Manager,
) *SyncService {
	return &SyncService{
		noteRepo:     noteRepo,
		versionRepo:  versionRepo,
		metadataRepo: metadataRepo,
		wsManager:    wsManager,
	}
}

func (s *SyncService) ProcessSyncRequest(userID, deviceID string, req *domain.SyncRequest) (*domain.SyncResponse, error) {
	notes, err := s.noteRepo.List(userID)
	if err != nil {
		return nil, err
	}

	var changes []*domain.NoteChange

	for _, note := range notes {
		clientVersion, exists := req.NoteVersions[note.ID]

		if !exists || clientVersion < note.Version {
			operation := "update"
			if note.IsDeleted {
				operation = "delete"
			}

			changes = append(changes, &domain.NoteChange{
				NoteID:    note.ID,
				Operation: operation,
				Version:   note.Version,
				Note: &domain.NoteResponse{
					ID:               note.ID,
					ParentID:         note.ParentID,
					Type:             note.Type,
					EncryptedTitle:   note.EncryptedTitle,
					EncryptedContent: note.EncryptedContent,
					EncryptionAlgo:   note.EncryptionAlgo,
					Nonce:            note.Nonce,
					CreatedAt:        note.CreatedAt,
					UpdatedAt:        note.UpdatedAt,
					IsDeleted:        note.IsDeleted,
					Version:          note.Version,
					ContentHash:      note.ContentHash,
					LastEditDevice:   note.LastEditDevice,
				},
			})
		}
	}

	syncTime := time.Now()
	if err := s.metadataRepo.UpdateLastSync(userID, deviceID, syncTime); err != nil {
		return nil, err
	}

	for noteID, version := range req.NoteVersions {
		if err := s.metadataRepo.UpdateNoteVersion(userID, deviceID, noteID, version); err != nil {
			continue
		}
	}

	return &domain.SyncResponse{
		Changes:  changes,
		SyncTime: syncTime,
		HasMore:  false,
	}, nil
}

func (s *SyncService) GetChangesSince(userID string, since time.Time) ([]*domain.NoteChange, error) {
	notes, err := s.noteRepo.List(userID)
	if err != nil {
		return nil, err
	}

	var changes []*domain.NoteChange

	for _, note := range notes {
		if note.UpdatedAt.After(since) {
			operation := "update"
			if note.IsDeleted {
				operation = "delete"
			}

			changes = append(changes, &domain.NoteChange{
				NoteID:    note.ID,
				Operation: operation,
				Version:   note.Version,
				Note: &domain.NoteResponse{
					ID:               note.ID,
					ParentID:         note.ParentID,
					Type:             note.Type,
					EncryptedTitle:   note.EncryptedTitle,
					EncryptedContent: note.EncryptedContent,
					EncryptionAlgo:   note.EncryptionAlgo,
					Nonce:            note.Nonce,
					CreatedAt:        note.CreatedAt,
					UpdatedAt:        note.UpdatedAt,
					IsDeleted:        note.IsDeleted,
					Version:          note.Version,
					ContentHash:      note.ContentHash,
					LastEditDevice:   note.LastEditDevice,
				},
			})
		}
	}

	return changes, nil
}

func (s *SyncService) BroadcastNoteUpdate(userID, deviceID string, note *domain.NoteResponse) error {
	msg, err := websocket.NewMessage(websocket.TypeNoteUpdate, &websocket.NoteUpdatePayload{
		NoteID:           note.ID,
		Version:          note.Version,
		EncryptedTitle:   note.EncryptedTitle,
		EncryptedContent: note.EncryptedContent,
		UpdatedAt:        note.UpdatedAt,
		DeviceID:         deviceID,
	})
	if err != nil {
		return err
	}

	return s.wsManager.BroadcastToUser(userID, msg, deviceID)
}

func (s *SyncService) BroadcastNoteDelete(userID, deviceID, noteID string, version int64) error {
	msg, err := websocket.NewMessage(websocket.TypeNoteDelete, &websocket.NoteDeletePayload{
		NoteID:   noteID,
		Version:  version,
		DeviceID: deviceID,
	})
	if err != nil {
		return err
	}

	return s.wsManager.BroadcastToUser(userID, msg, deviceID)
}
