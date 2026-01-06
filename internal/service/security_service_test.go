package service

import (
	"errors"
	"testing"
	"time"

	"inkdown-sync-server/internal/domain"
)

type mockKeyStoreRepo struct {
	keys map[string]*domain.EncryptedMasterKey
}

func newMockKeyStoreRepo() *mockKeyStoreRepo {
	return &mockKeyStoreRepo{
		keys: make(map[string]*domain.EncryptedMasterKey),
	}
}

func (m *mockKeyStoreRepo) Save(key *domain.EncryptedMasterKey) error {
	m.keys[key.UserID] = key
	return nil
}

func (m *mockKeyStoreRepo) Get(userID string) (*domain.EncryptedMasterKey, error) {
	if key, exists := m.keys[userID]; exists {
		return key, nil
	}
	return nil, errors.New("key not found")
}

func TestSecurityService_UploadKey(t *testing.T) {
	repo := newMockKeyStoreRepo()
	service := NewSecurityService(repo)

	req := &domain.UploadKeyRequest{
		EncryptedKey:   "enc-key-data",
		KeySalt:        "salt-data",
		KDFParams:      "{}",
		EncryptionAlgo: "AES-256-GCM",
	}

	err := service.UploadKey("user1", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	key, _ := repo.Get("user1")
	if key.EncryptedKey != req.EncryptedKey {
		t.Errorf("expected key %s, got %s", req.EncryptedKey, key.EncryptedKey)
	}
}

func TestSecurityService_GetKey(t *testing.T) {
	repo := newMockKeyStoreRepo()
	service := NewSecurityService(repo)

	repo.Save(&domain.EncryptedMasterKey{
		UserID:       "user1",
		EncryptedKey: "existing-key",
		UpdatedAt:    time.Now(),
	})

	resp, err := service.GetKey("user1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.EncryptedKey != "existing-key" {
		t.Errorf("expected key existing-key, got %s", resp.EncryptedKey)
	}

	_, err = service.GetKey("user2")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
