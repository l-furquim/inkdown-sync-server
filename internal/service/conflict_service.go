package service

import (
	"errors"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"

	"github.com/google/uuid"
)

type ConflictService struct {
	conflictRepo repository.ConflictRepository
	versionRepo  repository.NoteVersionRepository
	noteRepo     repository.NoteRepository
}

func NewConflictService(
	conflictRepo repository.ConflictRepository,
	versionRepo repository.NoteVersionRepository,
	noteRepo repository.NoteRepository,
) *ConflictService {
	return &ConflictService{
		conflictRepo: conflictRepo,
		versionRepo:  versionRepo,
		noteRepo:     noteRepo,
	}
}

func (s *ConflictService) DetectConflict(noteID, userID, deviceID string, expectedVersion int64, updateReq *domain.UpdateNoteRequest) (*domain.Conflict, error) {
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, err
	}

	if note.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if note.Version == expectedVersion {
		return nil, nil
	}

	conflict := &domain.Conflict{
		ID:            uuid.New().String(),
		NoteID:        noteID,
		UserID:        userID,
		Type:          domain.ConflictTypeUpdate,
		BaseVersion:   expectedVersion,
		ServerVersion: note.Version,
		ClientVersion: expectedVersion + 1,
		ServerNote:    note,
		ClientData:    updateReq,
		DeviceID:      deviceID,
		DetectedAt:    time.Now(),
	}

	if err := s.conflictRepo.Create(conflict); err != nil {
		return nil, err
	}

	return conflict, nil
}

func (s *ConflictService) ResolveWithLWW(conflict *domain.Conflict) (*domain.Note, error) {
	serverNote := conflict.ServerNote

	if conflict.ClientData == nil {
		return serverNote, nil
	}

	clientUpdatedAt := time.Now()
	if conflict.ClientData.ExpectedVersion != nil {
		versions, err := s.versionRepo.GetVersions(conflict.NoteID, 10)
		if err == nil && len(versions) > 0 {
			for _, v := range versions {
				if v.Version == *conflict.ClientData.ExpectedVersion {
					clientUpdatedAt = v.CreatedAt
					break
				}
			}
		}
	}

	if serverNote.UpdatedAt.After(clientUpdatedAt) {
		if err := s.conflictRepo.MarkResolved(conflict.ID, domain.ResolutionLWW); err != nil {
			return nil, err
		}
		return serverNote, nil
	}

	if conflict.ClientData.EncryptedContent != nil {
		serverNote.EncryptedContent = *conflict.ClientData.EncryptedContent
	}
	if conflict.ClientData.EncryptedTitle != nil {
		serverNote.EncryptedTitle = *conflict.ClientData.EncryptedTitle
	}
	if conflict.ClientData.Nonce != nil {
		serverNote.Nonce = *conflict.ClientData.Nonce
	}
	if conflict.ClientData.ContentHash != nil {
		serverNote.ContentHash = *conflict.ClientData.ContentHash
	}

	serverNote.UpdatedAt = time.Now()
	serverNote.Version++
	serverNote.LastEditDevice = conflict.DeviceID

	if err := s.noteRepo.Update(serverNote); err != nil {
		return nil, err
	}

	if err := s.conflictRepo.MarkResolved(conflict.ID, domain.ResolutionLWW); err != nil {
		return nil, err
	}

	return serverNote, nil
}

func (s *ConflictService) ApplyResolution(conflictID string, strategy domain.ResolutionStrategy, noteData *domain.UpdateNoteRequest) (*domain.Note, error) {
	conflict, err := s.conflictRepo.Get(conflictID)
	if err != nil {
		return nil, err
	}

	switch strategy {
	case domain.ResolutionLWW:
		return s.ResolveWithLWW(conflict)

	case domain.ResolutionServer:
		if err := s.conflictRepo.MarkResolved(conflictID, domain.ResolutionServer); err != nil {
			return nil, err
		}
		return conflict.ServerNote, nil

	case domain.ResolutionClient:
		if conflict.ClientData == nil {
			return nil, errors.New("no client data available")
		}

		note := conflict.ServerNote
		if conflict.ClientData.EncryptedContent != nil {
			note.EncryptedContent = *conflict.ClientData.EncryptedContent
		}
		if conflict.ClientData.EncryptedTitle != nil {
			note.EncryptedTitle = *conflict.ClientData.EncryptedTitle
		}
		if conflict.ClientData.Nonce != nil {
			note.Nonce = *conflict.ClientData.Nonce
		}
		if conflict.ClientData.ContentHash != nil {
			note.ContentHash = *conflict.ClientData.ContentHash
		}

		note.UpdatedAt = time.Now()
		note.Version++
		note.LastEditDevice = conflict.DeviceID

		if err := s.noteRepo.Update(note); err != nil {
			return nil, err
		}

		if err := s.conflictRepo.MarkResolved(conflictID, domain.ResolutionClient); err != nil {
			return nil, err
		}

		return note, nil

	case domain.ResolutionManual:
		if noteData == nil {
			return nil, errors.New("manual resolution requires note data")
		}

		note := conflict.ServerNote
		if noteData.EncryptedContent != nil {
			note.EncryptedContent = *noteData.EncryptedContent
		}
		if noteData.EncryptedTitle != nil {
			note.EncryptedTitle = *noteData.EncryptedTitle
		}
		if noteData.Nonce != nil {
			note.Nonce = *noteData.Nonce
		}
		if noteData.ContentHash != nil {
			note.ContentHash = *noteData.ContentHash
		}

		note.UpdatedAt = time.Now()
		note.Version++
		note.LastEditDevice = noteData.DeviceID

		if err := s.noteRepo.Update(note); err != nil {
			return nil, err
		}

		if err := s.conflictRepo.MarkResolved(conflictID, domain.ResolutionManual); err != nil {
			return nil, err
		}

		return note, nil

	default:
		return nil, fmt.Errorf("unknown resolution strategy: %s", strategy)
	}
}

func (s *ConflictService) Get(conflictID string) (*domain.Conflict, error) {
	return s.conflictRepo.Get(conflictID)
}

func (s *ConflictService) ListByUser(userID string) ([]*domain.Conflict, error) {
	return s.conflictRepo.ListByUser(userID)
}

func (s *ConflictService) ListByNote(noteID string) ([]*domain.Conflict, error) {
	return s.conflictRepo.ListByNote(noteID)
}
