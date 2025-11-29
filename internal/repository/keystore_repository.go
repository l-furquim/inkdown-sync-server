package repository

import (
	"context"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"

	"github.com/go-kivik/kivik/v4"
)

type KeyStoreRepository interface {
	Save(key *domain.EncryptedMasterKey) error
	Get(userID string) (*domain.EncryptedMasterKey, error)
}

type keyStoreRepository struct {
	client *kivik.Client
	dbName string
}

func NewKeyStoreRepository(client *kivik.Client, dbName string) KeyStoreRepository {
	return &keyStoreRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *keyStoreRepository) Save(key *domain.EncryptedMasterKey) error {
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("key_store:%s", key.UserID)

	// Check if key already exists to preserve _rev for update
	var rawDoc map[string]interface{}
	row := db.Get(context.Background(), docID)

	// If it exists, we update it. If not (error), we create new.
	if err := row.ScanDoc(&rawDoc); err == nil {
		// Update existing
		rawDoc["encrypted_key"] = key.EncryptedKey
		rawDoc["key_salt"] = key.KeySalt
		rawDoc["kdf_params"] = key.KDFParams
		rawDoc["encryption_algo"] = key.EncryptionAlgo
		rawDoc["updated_at"] = time.Now()
		// CreatedAt should be preserved from rawDoc if we wanted, but domain object has it.
		// Let's just update the fields we care about.

		_, err := db.Put(context.Background(), docID, rawDoc)
		if err != nil {
			return fmt.Errorf("failed to update key store: %w", err)
		}
	} else {
		// Create new
		// We can just put the domain object directly as it's new
		_, err := db.Put(context.Background(), docID, key)
		if err != nil {
			return fmt.Errorf("failed to create key store: %w", err)
		}
	}

	return nil
}

func (r *keyStoreRepository) Get(userID string) (*domain.EncryptedMasterKey, error) {
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("key_store:%s", userID)

	row := db.Get(context.Background(), docID)

	var key domain.EncryptedMasterKey
	if err := row.ScanDoc(&key); err != nil {
		return nil, fmt.Errorf("failed to get key store: %w", err)
	}

	return &key, nil
}
