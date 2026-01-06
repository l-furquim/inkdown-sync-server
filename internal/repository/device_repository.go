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
			"type":    map[string]interface{}{"$in": []string{"desktop", "mobile", "tablet"}},
		},
	}

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
