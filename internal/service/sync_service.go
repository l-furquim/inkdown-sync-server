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

// GetManifest returns a compact list of all notes for efficient sync comparison
// If workspaceID is provided, returns notes for that workspace only
func (s *SyncService) GetManifest(userID, workspaceID string) (*domain.ManifestResponse, error) {
	var notes []*domain.Note
	var err error

	if workspaceID != "" {
		notes, err = s.noteRepo.ListByWorkspace(workspaceID)
	} else {
		notes, err = s.noteRepo.List(userID)
	}
	if err != nil {
		return nil, err
	}

	entries := make([]domain.ManifestEntry, 0, len(notes))
	for _, note := range notes {
		entries = append(entries, domain.ManifestEntry{
			ID:          note.ID,
			ContentHash: note.ContentHash,
			Version:     note.Version,
			UpdatedAt:   note.UpdatedAt,
			IsDeleted:   note.IsDeleted,
		})
	}

	return &domain.ManifestResponse{
		Notes:    entries,
		SyncTime: time.Now(),
	}, nil
}

// ProcessBatchDiff compares client state with server and returns needed actions
func (s *SyncService) ProcessBatchDiff(userID string, req *domain.BatchDiffRequest) (*domain.BatchDiffResponse, error) {
	// Get server notes - use workspace if provided
	var serverNotes []*domain.Note
	var err error

	if req.WorkspaceID != "" {
		serverNotes, err = s.noteRepo.ListByWorkspace(req.WorkspaceID)
	} else {
		serverNotes, err = s.noteRepo.List(userID)
	}
	if err != nil {
		return nil, err
	}

	// Build server map for quick lookup
	serverMap := make(map[string]*domain.Note)
	for _, note := range serverNotes {
		serverMap[note.ID] = note
	}

	// Build client map for quick lookup
	clientMap := make(map[string]*domain.LocalNoteInfo)
	for i := range req.LocalNotes {
		clientMap[req.LocalNotes[i].ID] = &req.LocalNotes[i]
	}

	response := &domain.BatchDiffResponse{
		ToDownload: []domain.NoteResponse{},
		ToUpload:   []string{},
		ToDelete:   []string{},
		Conflicts:  []domain.ConflictInfo{},
		SyncTime:   time.Now(),
	}

	// Check each server note against client state
	for noteID, serverNote := range serverMap {
		clientNote, existsOnClient := clientMap[noteID]

		if serverNote.IsDeleted {
			// Server deleted this note
			if existsOnClient {
				response.ToDelete = append(response.ToDelete, noteID)
			}
			continue
		}

		// Client doesn't have this note - download
		if !existsOnClient {
			response.ToDownload = append(response.ToDownload, domain.NoteResponse{
				ID:               serverNote.ID,
				WorkspaceID:      serverNote.WorkspaceID,
				ParentID:         serverNote.ParentID,
				Type:             serverNote.Type,
				EncryptedTitle:   serverNote.EncryptedTitle,
				EncryptedContent: serverNote.EncryptedContent,
				EncryptionAlgo:   serverNote.EncryptionAlgo,
				Nonce:            serverNote.Nonce,
				CreatedAt:        serverNote.CreatedAt,
				UpdatedAt:        serverNote.UpdatedAt,
				IsDeleted:        serverNote.IsDeleted,
				Version:          serverNote.Version,
				ContentHash:      serverNote.ContentHash,
				LastEditDevice:   serverNote.LastEditDevice,
			})
			continue
		}

		// Both have the note - compare
		if serverNote.ContentHash == clientNote.ContentHash {
			// Same content, no action needed
			continue
		}

		// Different content - check versions
		if serverNote.Version > clientNote.Version {
			// Server has been updated since client's last sync
			// This is a REAL conflict - both sides changed
			response.Conflicts = append(response.Conflicts, domain.ConflictInfo{
				NoteID:        noteID,
				LocalHash:     clientNote.ContentHash,
				ServerHash:    serverNote.ContentHash,
				LocalVersion:  clientNote.Version,
				ServerVersion: serverNote.Version,
			})
		} else {
			// Server version <= client version means:
			// - Either client is newer (serverVersion < clientVersion) - shouldn't happen normally
			// - Or same version with different hash means client edited since last sync
			// In both cases, client should upload
			response.ToUpload = append(response.ToUpload, noteID)
		}
	}

	// Notes that exist only on client (not in serverMap) should be uploaded
	// But we don't know about them here since they don't have server IDs yet
	// Those are handled by the client's "unmapped files" logic

	return response, nil
}
