package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"inkdown-sync-server/internal/domain"
)

type SyncMetadataRepository interface {
	Get(userID, deviceID string) (*domain.SyncMetadata, error)
	Upsert(metadata *domain.SyncMetadata) error
	UpdateLastSync(userID, deviceID string, timestamp time.Time) error
	UpdateNoteVersion(userID, deviceID, noteID string, version int64) error
}

type syncMetadataRepo struct {
	baseURL string
	client  *http.Client
}

func NewSyncMetadataRepository(baseURL string) SyncMetadataRepository {
	return &syncMetadataRepo{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *syncMetadataRepo) Get(userID, deviceID string) (*domain.SyncMetadata, error) {
	docID := fmt.Sprintf("sync:%s:%s", userID, deviceID)
	url := fmt.Sprintf("%s/%s", r.baseURL, docID)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &domain.SyncMetadata{
			UserID:           userID,
			DeviceID:         deviceID,
			LastSyncTime:     time.Time{},
			NoteVersions:     make(map[string]int64),
			PendingConflicts: []string{},
			UpdatedAt:        time.Now(),
		}, nil
	}

	var metadata domain.SyncMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (r *syncMetadataRepo) Upsert(metadata *domain.SyncMetadata) error {
	docID := fmt.Sprintf("sync:%s:%s", metadata.UserID, metadata.DeviceID)

	existing, _ := r.Get(metadata.UserID, metadata.DeviceID)

	doc := map[string]interface{}{
		"_id":               docID,
		"user_id":           metadata.UserID,
		"device_id":         metadata.DeviceID,
		"last_sync_time":    metadata.LastSyncTime,
		"note_versions":     metadata.NoteVersions,
		"pending_conflicts": metadata.PendingConflicts,
		"updated_at":        time.Now(),
	}

	if existing != nil {
		url := fmt.Sprintf("%s/%s", r.baseURL, docID)
		resp, _ := r.client.Get(url)
		if resp != nil && resp.StatusCode == http.StatusOK {
			var existingDoc map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&existingDoc)
			if rev, ok := existingDoc["_rev"].(string); ok {
				doc["_rev"] = rev
			}
			resp.Body.Close()
		}
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", r.baseURL, docID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upsert sync metadata: status %d", resp.StatusCode)
	}

	return nil
}

func (r *syncMetadataRepo) UpdateLastSync(userID, deviceID string, timestamp time.Time) error {
	metadata, err := r.Get(userID, deviceID)
	if err != nil {
		return err
	}

	metadata.LastSyncTime = timestamp
	metadata.UpdatedAt = time.Now()

	return r.Upsert(metadata)
}

func (r *syncMetadataRepo) UpdateNoteVersion(userID, deviceID, noteID string, version int64) error {
	metadata, err := r.Get(userID, deviceID)
	if err != nil {
		return err
	}

	if metadata.NoteVersions == nil {
		metadata.NoteVersions = make(map[string]int64)
	}

	metadata.NoteVersions[noteID] = version
	metadata.UpdatedAt = time.Now()

	return r.Upsert(metadata)
}
