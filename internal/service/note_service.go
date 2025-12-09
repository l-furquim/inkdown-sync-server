package service

import (
	"errors"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"

	"github.com/google/uuid"
)

type NoteService struct {
	repo            repository.NoteRepository
	versionRepo     repository.NoteVersionRepository
	conflictService *ConflictService
	syncService     *SyncService
}

func NewNoteService(
	repo repository.NoteRepository,
	versionRepo repository.NoteVersionRepository,
	conflictService *ConflictService,
	syncService *SyncService,
) *NoteService {
	return &NoteService{
		repo:            repo,
		versionRepo:     versionRepo,
		conflictService: conflictService,
		syncService:     syncService,
	}
}

func (s *NoteService) Create(userID string, req *domain.CreateNoteRequest) (*domain.NoteResponse, error) {
	noteID := uuid.New().String()
	now := time.Now()

	note := &domain.Note{
		ID:               noteID,
		UserID:           userID,
		ParentID:         req.ParentID,
		Type:             req.Type,
		EncryptedTitle:   req.EncryptedTitle,
		EncryptedContent: req.EncryptedContent,
		EncryptionAlgo:   req.EncryptionAlgo,
		Nonce:            req.Nonce,
		CreatedAt:        now,
		UpdatedAt:        now,
		IsDeleted:        false,
		Version:          1,
		ContentHash:      req.ContentHash,
		LastEditDevice:   req.DeviceID,
		WorkspaceID:      req.WorkspaceID,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}

	response := &domain.NoteResponse{
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
		WorkspaceID:      note.WorkspaceID,
	}

	if s.syncService != nil {
		s.syncService.BroadcastNoteUpdate(userID, req.DeviceID, response)
	}

	return response, nil
}

func (s *NoteService) List(userID string) ([]*domain.NoteResponse, error) {
	notes, err := s.repo.List(userID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.NoteResponse
	for _, n := range notes {
		responses = append(responses, &domain.NoteResponse{
			ID:               n.ID,
			ParentID:         n.ParentID,
			Type:             n.Type,
			EncryptedTitle:   n.EncryptedTitle,
			EncryptedContent: n.EncryptedContent,
			EncryptionAlgo:   n.EncryptionAlgo,
			Nonce:            n.Nonce,
			CreatedAt:        n.CreatedAt,
			UpdatedAt:        n.UpdatedAt,
			IsDeleted:        n.IsDeleted,
			Version:          n.Version,
			ContentHash:      n.ContentHash,
			LastEditDevice:   n.LastEditDevice,
			WorkspaceID:      n.WorkspaceID,
		})
	}

	return responses, nil
}

func (s *NoteService) GetByID(userID, noteID string) (*domain.NoteResponse, error) {
	note, err := s.repo.FindByID(noteID)
	if err != nil {
		return nil, err
	}

	if note.UserID != userID {
		return nil, errors.New("unauthorized: note does not belong to user")
	}

	return &domain.NoteResponse{
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
		WorkspaceID:      note.WorkspaceID,
	}, nil
}

func (s *NoteService) Update(userID, noteID string, req *domain.UpdateNoteRequest) (*domain.NoteResponse, error) {
	note, err := s.repo.FindByID(noteID)
	if err != nil {
		return nil, err
	}

	if note.UserID != userID {
		return nil, errors.New("unauthorized: note does not belong to user")
	}

	if req.ExpectedVersion != nil && *req.ExpectedVersion != note.Version {
		conflict, err := s.conflictService.DetectConflict(noteID, userID, req.DeviceID, *req.ExpectedVersion, req)
		if err != nil {
			return nil, err
		}
		return nil, &ConflictError{Conflict: conflict}
	}

	if s.versionRepo != nil {
		s.versionRepo.SaveVersion(note)
	}

	if req.EncryptedTitle != nil {
		note.EncryptedTitle = *req.EncryptedTitle
	}
	if req.EncryptedContent != nil {
		note.EncryptedContent = *req.EncryptedContent
	}
	if req.EncryptionAlgo != nil {
		note.EncryptionAlgo = *req.EncryptionAlgo
	}
	if req.Nonce != nil {
		note.Nonce = *req.Nonce
	}
	if req.ParentID != nil {
		note.ParentID = req.ParentID
	}
	if req.IsDeleted != nil {
		note.IsDeleted = *req.IsDeleted
	}
	if req.ContentHash != nil {
		note.ContentHash = *req.ContentHash
	}

	note.UpdatedAt = time.Now()
	note.Version++
	note.LastEditDevice = req.DeviceID

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	response := &domain.NoteResponse{
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
		WorkspaceID:      note.WorkspaceID,
	}

	if s.syncService != nil {
		s.syncService.BroadcastNoteUpdate(userID, req.DeviceID, response)
	}

	return response, nil
}

func (s *NoteService) Delete(userID, noteID string) error {
	note, err := s.repo.FindByID(noteID)
	if err != nil {
		return err
	}

	if note.UserID != userID {
		return errors.New("unauthorized: note does not belong to user")
	}

	return s.repo.Delete(noteID)
}
