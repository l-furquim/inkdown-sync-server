package service

import (
	"errors"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrAccessDenied      = errors.New("access denied")
)

type WorkspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	noteRepo      repository.NoteRepository
}

func NewWorkspaceService(workspaceRepo repository.WorkspaceRepository, noteRepo repository.NoteRepository) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		noteRepo:      noteRepo,
	}
}

// Create creates a new workspace for the user
func (s *WorkspaceService) Create(ownerID string, req *domain.CreateWorkspaceRequest) (*domain.WorkspaceResponse, error) {
	workspace := &domain.Workspace{
		ID:        "workspace:" + uuid.New().String(),
		OwnerID:   ownerID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDefault: false,
	}

	if err := s.workspaceRepo.Create(workspace); err != nil {
		return nil, err
	}

	return s.workspaceToResponse(workspace), nil
}

// List returns all workspaces for a user
func (s *WorkspaceService) List(ownerID string) ([]*domain.WorkspaceResponse, error) {
	workspaces, err := s.workspaceRepo.GetByOwner(ownerID)
	if err != nil {
		return nil, err
	}

	responses := make([]*domain.WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		responses[i] = s.workspaceToResponse(ws)
	}

	return responses, nil
}

// Get returns a specific workspace if the user has access
func (s *WorkspaceService) Get(userID, workspaceID string) (*domain.WorkspaceResponse, error) {
	workspace, err := s.workspaceRepo.Get(workspaceID)
	if err != nil {
		return nil, err
	}

	if workspace.OwnerID != userID {
		return nil, ErrAccessDenied
	}

	return s.workspaceToResponse(workspace), nil
}

// Update updates a workspace
func (s *WorkspaceService) Update(userID, workspaceID string, req *domain.UpdateWorkspaceRequest) (*domain.WorkspaceResponse, error) {
	workspace, err := s.workspaceRepo.Get(workspaceID)
	if err != nil {
		return nil, err
	}

	if workspace.OwnerID != userID {
		return nil, ErrAccessDenied
	}

	if req.Name != "" {
		workspace.Name = req.Name
	}
	workspace.UpdatedAt = time.Now()

	if err := s.workspaceRepo.Update(workspace); err != nil {
		return nil, err
	}

	return s.workspaceToResponse(workspace), nil
}

// Delete deletes a workspace
func (s *WorkspaceService) Delete(userID, workspaceID string) error {
	workspace, err := s.workspaceRepo.Get(workspaceID)
	if err != nil {
		return err
	}

	if workspace.OwnerID != userID {
		return ErrAccessDenied
	}

	if workspace.IsDefault {
		return errors.New("cannot delete default workspace")
	}

	return s.workspaceRepo.Delete(workspaceID)
}

// ValidateAccess checks if a user has access to a workspace
func (s *WorkspaceService) ValidateAccess(userID, workspaceID string) error {
	workspace, err := s.workspaceRepo.Get(workspaceID)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	if workspace.OwnerID != userID {
		// Future: check workspace membership for shared workspaces
		return ErrAccessDenied
	}

	return nil
}

// CreateDefaultForUser creates a default workspace for a new user
func (s *WorkspaceService) CreateDefaultForUser(userID string) (*domain.WorkspaceResponse, error) {
	workspace := &domain.Workspace{
		ID:        "workspace:" + uuid.New().String(),
		OwnerID:   userID,
		Name:      "My Workspace",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDefault: true,
	}

	if err := s.workspaceRepo.Create(workspace); err != nil {
		return nil, err
	}

	return s.workspaceToResponse(workspace), nil
}

// GetDefaultWorkspace returns the default workspace for a user
func (s *WorkspaceService) GetDefaultWorkspace(userID string) (*domain.WorkspaceResponse, error) {
	workspace, err := s.workspaceRepo.GetDefault(userID)
	if err != nil {
		return nil, err
	}

	return s.workspaceToResponse(workspace), nil
}

func (s *WorkspaceService) workspaceToResponse(ws *domain.Workspace) *domain.WorkspaceResponse {
	noteCount := 0
	if notes, err := s.noteRepo.ListByWorkspace(ws.ID); err == nil {
		// Count only non-deleted notes
		for _, note := range notes {
			if !note.IsDeleted {
				noteCount++
			}
		}
	}

	return &domain.WorkspaceResponse{
		ID:        ws.ID,
		Name:      ws.Name,
		CreatedAt: ws.CreatedAt,
		UpdatedAt: ws.UpdatedAt,
		IsDefault: ws.IsDefault,
		NoteCount: noteCount,
	}
}
