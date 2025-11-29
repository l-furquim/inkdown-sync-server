package repository

import (
	"context"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"

	"github.com/go-kivik/kivik/v4"
)

type DeviceRepository interface {
	Create(device *domain.Device) error
	List(userID string) ([]*domain.Device, error)
	FindByID(deviceID string) (*domain.Device, error)
	Revoke(deviceID string) error
	UpdateLastActive(deviceID string) error
}

type deviceRepository struct {
	client *kivik.Client
	dbName string
}

func NewDeviceRepository(client *kivik.Client, dbName string) DeviceRepository {
	return &deviceRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *deviceRepository) Create(device *domain.Device) error {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("device:%s", device.ID)
	_, err := db.Put(context.Background(), docID, device)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

func (r *deviceRepository) List(userID string) ([]*domain.Device, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"user_id": userID,
			"type":    map[string]interface{}{"$in": []string{"desktop", "mobile", "tablet"}}, // Simple filter to ensure we get devices
		},
	}

	// Note: Ideally we should have a "doc_type": "device" field in the domain model to distinguish docs,
	// but for now relying on the structure and ID prefix pattern (though selectors don't filter by ID prefix easily without a view).
	// Adding a specific check for fields present in Device struct via selector is a good practice if no doc_type.
	// Let's assume for now we filter by user_id and existence of 'os' field to be sure.
	query["selector"].(map[string]interface{})["os"] = map[string]interface{}{"$exists": true}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		var device domain.Device
		if err := rows.ScanDoc(&device); err != nil {
			continue // Skip malformed docs
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

func (r *deviceRepository) FindByID(deviceID string) (*domain.Device, error) {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("device:%s", deviceID)
	row := db.Get(context.Background(), docID)

	var device domain.Device
	if err := row.ScanDoc(&device); err != nil {
		return nil, fmt.Errorf("failed to find device: %w", err)
	}

	return &device, nil
}

func (r *deviceRepository) Revoke(deviceID string) error {
	device, err := r.FindByID(deviceID)
	if err != nil {
		return err
	}

	device.IsRevoked = true

	// We need to update the document. In CouchDB updates usually require the _rev,
	// but ScanDoc might not populate it if not in the struct.
	// However, kivik handles optimistic locking if we pass the struct back.
	// Let's ensure we are doing a proper update.
	// Actually, FindByID gets the doc. To update, we just Put it back with the same ID.
	// Kivik should handle the revision if it was unmarshaled into a struct with a _rev tag,
	// OR we might need to handle it.
	// The domain.Device struct does NOT have a _rev field.
	// This is a potential issue for updates in CouchDB (conflict).
	// For this implementation, let's fetch the current revision first or use a map to preserve it.

	// Better approach for update without _rev in domain:
	// 1. Get the raw document as map to keep _rev
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("device:%s", deviceID)

	var rawDoc map[string]interface{}
	row := db.Get(context.Background(), docID)
	if err := row.ScanDoc(&rawDoc); err != nil {
		return err
	}

	rawDoc["is_revoked"] = true

	_, err = db.Put(context.Background(), docID, rawDoc)
	if err != nil {
		return fmt.Errorf("failed to revoke device: %w", err)
	}

	return nil
}

func (r *deviceRepository) UpdateLastActive(deviceID string) error {
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("device:%s", deviceID)

	// Similar to Revoke, we need to preserve _rev
	var rawDoc map[string]interface{}
	row := db.Get(context.Background(), docID)
	if err := row.ScanDoc(&rawDoc); err != nil {
		return err
	}

	rawDoc["last_active"] = time.Now()

	_, err := db.Put(context.Background(), docID, rawDoc)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	return nil
}
