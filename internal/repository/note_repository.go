package repository

import (
	"context"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"

	"github.com/go-kivik/kivik/v4"
)

type NoteRepository interface {
	Create(note *domain.Note) error
	FindByID(id string) (*domain.Note, error)
	List(userID string) ([]*domain.Note, error)
	ListByWorkspace(workspaceID string) ([]*domain.Note, error)
	Update(note *domain.Note) error
	Delete(id string) error
}

type noteRepository struct {
	client *kivik.Client
	dbName string
}

func NewNoteRepository(client *kivik.Client, dbName string) NoteRepository {
	return &noteRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *noteRepository) Create(note *domain.Note) error {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("note:%s", note.ID)
	_, err := db.Put(context.Background(), docID, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

func (r *noteRepository) FindByID(id string) (*domain.Note, error) {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("note:%s", id)
	row := db.Get(context.Background(), docID)

	var note domain.Note
	if err := row.ScanDoc(&note); err != nil {
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	return &note, nil
}

func (r *noteRepository) List(userID string) ([]*domain.Note, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"user_id":         userID,
			"encrypted_title": map[string]interface{}{"$exists": true},
		},
	}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		var note domain.Note
		if err := rows.ScanDoc(&note); err != nil {
			continue
		}
		notes = append(notes, &note)
	}

	return notes, nil
}

func (r *noteRepository) ListByWorkspace(workspaceID string) ([]*domain.Note, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"workspace_id":    workspaceID,
			"encrypted_title": map[string]interface{}{"$exists": true},
		},
	}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to list notes by workspace: %w", err)
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		var note domain.Note
		if err := rows.ScanDoc(&note); err != nil {
			continue
		}
		notes = append(notes, &note)
	}

	return notes, nil
}

func (r *noteRepository) Update(note *domain.Note) error {
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("note:%s", note.ID)

	var existingDoc map[string]interface{}
	row := db.Get(context.Background(), docID)
	if err := row.ScanDoc(&existingDoc); err != nil {
		return fmt.Errorf("failed to fetch existing note for update: %w", err)
	}

	existingDoc["encrypted_title"] = note.EncryptedTitle
	existingDoc["encrypted_content"] = note.EncryptedContent
	existingDoc["encryption_algo"] = note.EncryptionAlgo
	existingDoc["nonce"] = note.Nonce
	existingDoc["content_hash"] = note.ContentHash
	existingDoc["last_edit_device"] = note.LastEditDevice
	existingDoc["updated_at"] = time.Now()
	existingDoc["version"] = note.Version // Service should increment this
	existingDoc["is_deleted"] = note.IsDeleted

	if note.ParentID != nil {
		existingDoc["parent_id"] = *note.ParentID
	} else {
		existingDoc["parent_id"] = nil
	}

	_, err := db.Put(context.Background(), docID, existingDoc)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

func (r *noteRepository) Delete(id string) error {
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("note:%s", id)

	var existingDoc map[string]interface{}
	row := db.Get(context.Background(), docID)
	if err := row.ScanDoc(&existingDoc); err != nil {
		return err
	}

	existingDoc["is_deleted"] = true
	existingDoc["updated_at"] = time.Now()

	if v, ok := existingDoc["version"].(float64); ok {
		existingDoc["version"] = int64(v) + 1
	}

	_, err := db.Put(context.Background(), docID, existingDoc)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
