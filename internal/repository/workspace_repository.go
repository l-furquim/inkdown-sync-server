package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"

	"github.com/go-kivik/kivik/v4"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrWorkspaceExists   = errors.New("workspace already exists")
)

type WorkspaceRepository interface {
	Create(workspace *domain.Workspace) error
	Get(id string) (*domain.Workspace, error)
	GetByOwner(ownerID string) ([]*domain.Workspace, error)
	GetDefault(ownerID string) (*domain.Workspace, error)
	Update(workspace *domain.Workspace) error
	Delete(id string) error
}

type CouchDBWorkspaceRepository struct {
	db *kivik.DB
}

type workspaceDoc struct {
	ID        string `json:"_id"`
	Rev       string `json:"_rev,omitempty"`
	DocType   string `json:"doc_type"`
	OwnerID   string `json:"owner_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsDefault bool   `json:"is_default"`
}

func NewWorkspaceRepository(client *kivik.Client, dbName string) *CouchDBWorkspaceRepository {
	return &CouchDBWorkspaceRepository{
		db: client.DB(dbName),
	}
}

func (r *CouchDBWorkspaceRepository) Create(workspace *domain.Workspace) error {
	doc := workspaceDoc{
		ID:        workspace.ID,
		DocType:   "workspace",
		OwnerID:   workspace.OwnerID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: workspace.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsDefault: workspace.IsDefault,
	}

	_, err := r.db.Put(context.Background(), doc.ID, doc)
	if err != nil {
		if kivik.HTTPStatus(err) == 409 {
			return ErrWorkspaceExists
		}
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

func (r *CouchDBWorkspaceRepository) Get(id string) (*domain.Workspace, error) {
	row := r.db.Get(context.Background(), id)

	var doc workspaceDoc
	if err := row.ScanDoc(&doc); err != nil {
		if kivik.HTTPStatus(err) == 404 {
			return nil, ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return docToWorkspace(&doc)
}

func (r *CouchDBWorkspaceRepository) GetByOwner(ownerID string) ([]*domain.Workspace, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"doc_type": "workspace",
			"owner_id": ownerID,
		},
	}

	rows := r.db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var doc workspaceDoc
		if err := rows.ScanDoc(&doc); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}

		ws, err := docToWorkspace(&doc)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
}

func (r *CouchDBWorkspaceRepository) GetDefault(ownerID string) (*domain.Workspace, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"doc_type":   "workspace",
			"owner_id":   ownerID,
			"is_default": true,
		},
		"limit": 1,
	}

	rows := r.db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query default workspace: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrWorkspaceNotFound
	}

	var doc workspaceDoc
	if err := rows.ScanDoc(&doc); err != nil {
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return docToWorkspace(&doc)
}

func (r *CouchDBWorkspaceRepository) Update(workspace *domain.Workspace) error {
	row := r.db.Get(context.Background(), workspace.ID)
	var existingDoc workspaceDoc
	if err := row.ScanDoc(&existingDoc); err != nil {
		if kivik.HTTPStatus(err) == 404 {
			return ErrWorkspaceNotFound
		}
		return fmt.Errorf("failed to get workspace for update: %w", err)
	}

	doc := workspaceDoc{
		ID:        workspace.ID,
		Rev:       existingDoc.Rev,
		DocType:   "workspace",
		OwnerID:   workspace.OwnerID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: workspace.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsDefault: workspace.IsDefault,
	}

	_, err := r.db.Put(context.Background(), doc.ID, doc)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

func (r *CouchDBWorkspaceRepository) Delete(id string) error {
	row := r.db.Get(context.Background(), id)
	var doc workspaceDoc
	if err := row.ScanDoc(&doc); err != nil {
		if kivik.HTTPStatus(err) == 404 {
			return ErrWorkspaceNotFound
		}
		return fmt.Errorf("failed to get workspace for delete: %w", err)
	}

	_, err := r.db.Delete(context.Background(), id, doc.Rev)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	return nil
}

func docToWorkspace(doc *workspaceDoc) (*domain.Workspace, error) {
	createdAt, err := parseTime(doc.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	updatedAt, err := parseTime(doc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &domain.Workspace{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Name:      doc.Name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		IsDefault: doc.IsDefault,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
