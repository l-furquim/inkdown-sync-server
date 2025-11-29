package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"inkdown-sync-server/internal/domain"
)

type NoteVersionRepository interface {
	SaveVersion(note *domain.Note) error
	GetVersions(noteID string, limit int) ([]*domain.NoteVersion, error)
	GetVersion(noteID string, version int64) (*domain.NoteVersion, error)
	DeleteOldVersions(noteID string, keepLast int) error
}

type noteVersionRepo struct {
	baseURL string
	client  *http.Client
}

func NewNoteVersionRepository(baseURL string) NoteVersionRepository {
	return &noteVersionRepo{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *noteVersionRepo) SaveVersion(note *domain.Note) error {
	version := &domain.NoteVersion{
		ID:               fmt.Sprintf("version:%s:%d", note.ID, note.Version),
		NoteID:           note.ID,
		Version:          note.Version,
		EncryptedContent: note.EncryptedContent,
		EncryptedTitle:   note.EncryptedTitle,
		ContentHash:      note.ContentHash,
		DeviceID:         note.LastEditDevice,
		CreatedAt:        time.Now(),
	}

	data, err := json.Marshal(version)
	if err != nil {
		return err
	}

	resp, err := r.client.Post(r.baseURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to save version: status %d", resp.StatusCode)
	}

	return nil
}

func (r *noteVersionRepo) GetVersions(noteID string, limit int) ([]*domain.NoteVersion, error) {
	viewURL := fmt.Sprintf("%s/_design/versions/_view/by_note?key=\"%s\"&limit=%d&descending=true",
		r.baseURL, noteID, limit)

	resp, err := r.client.Get(viewURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Rows []struct {
			Value domain.NoteVersion `json:"value"`
		} `json:"rows"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	versions := make([]*domain.NoteVersion, len(result.Rows))
	for i, row := range result.Rows {
		v := row.Value
		versions[i] = &v
	}

	return versions, nil
}

func (r *noteVersionRepo) GetVersion(noteID string, version int64) (*domain.NoteVersion, error) {
	docID := fmt.Sprintf("version:%s:%d", noteID, version)
	url := fmt.Sprintf("%s/%s", r.baseURL, docID)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("version not found")
	}

	var noteVersion domain.NoteVersion
	if err := json.NewDecoder(resp.Body).Decode(&noteVersion); err != nil {
		return nil, err
	}

	return &noteVersion, nil
}

func (r *noteVersionRepo) DeleteOldVersions(noteID string, keepLast int) error {
	versions, err := r.GetVersions(noteID, 100)
	if err != nil {
		return err
	}

	if len(versions) <= keepLast {
		return nil
	}

	toDelete := versions[keepLast:]
	for _, v := range toDelete {
		url := fmt.Sprintf("%s/%s", r.baseURL, v.ID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			continue
		}

		r.client.Do(req)
	}

	return nil
}
