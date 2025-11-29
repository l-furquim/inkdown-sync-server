package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"inkdown-sync-server/internal/domain"
)

type ConflictRepository interface {
	Create(conflict *domain.Conflict) error
	Get(conflictID string) (*domain.Conflict, error)
	ListByUser(userID string) ([]*domain.Conflict, error)
	ListByNote(noteID string) ([]*domain.Conflict, error)
	MarkResolved(conflictID string, choice domain.ResolutionStrategy) error
	Delete(conflictID string) error
}

type conflictRepo struct {
	baseURL string
	client  *http.Client
}

func NewConflictRepository(baseURL string) ConflictRepository {
	return &conflictRepo{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *conflictRepo) Create(conflict *domain.Conflict) error {
	doc := map[string]interface{}{
		"_id":               fmt.Sprintf("conflict:%s", conflict.ID),
		"note_id":           conflict.NoteID,
		"user_id":           conflict.UserID,
		"type":              conflict.Type,
		"base_version":      conflict.BaseVersion,
		"server_version":    conflict.ServerVersion,
		"client_version":    conflict.ClientVersion,
		"server_note":       conflict.ServerNote,
		"client_data":       conflict.ClientData,
		"device_id":         conflict.DeviceID,
		"detected_at":       conflict.DetectedAt,
		"resolved_at":       conflict.ResolvedAt,
		"resolution_choice": conflict.ResolutionChoice,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	resp, err := r.client.Post(r.baseURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create conflict: status %d", resp.StatusCode)
	}

	return nil
}

func (r *conflictRepo) Get(conflictID string) (*domain.Conflict, error) {
	url := fmt.Sprintf("%s/conflict:%s", r.baseURL, conflictID)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("conflict not found")
	}

	var conflict domain.Conflict
	if err := json.NewDecoder(resp.Body).Decode(&conflict); err != nil {
		return nil, err
	}

	return &conflict, nil
}

func (r *conflictRepo) ListByUser(userID string) ([]*domain.Conflict, error) {
	viewURL := fmt.Sprintf("%s/_design/conflicts/_view/by_user?key=\"%s\"", r.baseURL, userID)

	resp, err := r.client.Get(viewURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Rows []struct {
			Value domain.Conflict `json:"value"`
		} `json:"rows"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	conflicts := make([]*domain.Conflict, len(result.Rows))
	for i, row := range result.Rows {
		c := row.Value
		conflicts[i] = &c
	}

	return conflicts, nil
}

func (r *conflictRepo) ListByNote(noteID string) ([]*domain.Conflict, error) {
	viewURL := fmt.Sprintf("%s/_design/conflicts/_view/by_note?key=\"%s\"", r.baseURL, noteID)

	resp, err := r.client.Get(viewURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Rows []struct {
			Value domain.Conflict `json:"value"`
		} `json:"rows"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	conflicts := make([]*domain.Conflict, len(result.Rows))
	for i, row := range result.Rows {
		c := row.Value
		conflicts[i] = &c
	}

	return conflicts, nil
}

func (r *conflictRepo) MarkResolved(conflictID string, choice domain.ResolutionStrategy) error {
	conflict, err := r.Get(conflictID)
	if err != nil {
		return err
	}

	now := time.Now()
	conflict.ResolvedAt = &now
	conflict.ResolutionChoice = choice

	url := fmt.Sprintf("%s/conflict:%s", r.baseURL, conflictID)
	respGet, err := r.client.Get(url)
	if err != nil {
		return err
	}

	var existingDoc map[string]interface{}
	json.NewDecoder(respGet.Body).Decode(&existingDoc)
	respGet.Body.Close()

	doc := map[string]interface{}{
		"_id":               fmt.Sprintf("conflict:%s", conflictID),
		"_rev":              existingDoc["_rev"],
		"note_id":           conflict.NoteID,
		"user_id":           conflict.UserID,
		"type":              conflict.Type,
		"base_version":      conflict.BaseVersion,
		"server_version":    conflict.ServerVersion,
		"client_version":    conflict.ClientVersion,
		"server_note":       conflict.ServerNote,
		"client_data":       conflict.ClientData,
		"device_id":         conflict.DeviceID,
		"detected_at":       conflict.DetectedAt,
		"resolved_at":       conflict.ResolvedAt,
		"resolution_choice": conflict.ResolutionChoice,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("failed to mark conflict as resolved: status %d", resp.StatusCode)
	}

	return nil
}

func (r *conflictRepo) Delete(conflictID string) error {
	url := fmt.Sprintf("%s/conflict:%s", r.baseURL, conflictID)

	resp, err := r.client.Get(url)
	if err != nil {
		return err
	}

	var doc map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&doc)
	resp.Body.Close()

	if rev, ok := doc["_rev"].(string); ok {
		deleteURL := fmt.Sprintf("%s?rev=%s", url, rev)
		req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
		if err != nil {
			return err
		}

		resp, err := r.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to delete conflict: status %d", resp.StatusCode)
		}
	}

	return nil
}
