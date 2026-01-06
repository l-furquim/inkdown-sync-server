package service

import (
	"errors"
	"testing"
	"time"

	"inkdown-sync-server/internal/domain"
)

type mockDeviceRepo struct {
	devices map[string]*domain.Device
}

func newMockDeviceRepo() *mockDeviceRepo {
	return &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
	}
}

func (m *mockDeviceRepo) Create(device *domain.Device) error {
	if _, exists := m.devices[device.ID]; exists {
		return errors.New("device already exists")
	}
	m.devices[device.ID] = device
	return nil
}

func (m *mockDeviceRepo) List(userID string) ([]*domain.Device, error) {
	var devices []*domain.Device
	for _, d := range m.devices {
		if d.UserID == userID {
			devices = append(devices, d)
		}
	}
	return devices, nil
}

func (m *mockDeviceRepo) FindByID(deviceID string) (*domain.Device, error) {
	if d, exists := m.devices[deviceID]; exists {
		return d, nil
	}
	return nil, errors.New("device not found")
}

func (m *mockDeviceRepo) Revoke(deviceID string) error {
	if d, exists := m.devices[deviceID]; exists {
		d.IsRevoked = true
		return nil
	}
	return errors.New("device not found")
}

func (m *mockDeviceRepo) UpdateLastActive(deviceID string) error {
	if d, exists := m.devices[deviceID]; exists {
		d.LastActive = time.Now()
		return nil
	}
	return errors.New("device not found")
}

func TestDeviceService_Register(t *testing.T) {
	repo := newMockDeviceRepo()
	service := NewDeviceService(repo)

	req := &domain.RegisterDeviceRequest{
		Name:       "Test Device",
		Type:       "desktop",
		OS:         "linux",
		AppVersion: "1.0.0",
	}

	resp, err := service.Register("user1", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, resp.Name)
	}
	if resp.ID == "" {
		t.Error("expected device ID to be generated")
	}
}

func TestDeviceService_List(t *testing.T) {
	repo := newMockDeviceRepo()
	service := NewDeviceService(repo)

	repo.Create(&domain.Device{ID: "d1", UserID: "user1", Name: "D1"})
	repo.Create(&domain.Device{ID: "d2", UserID: "user1", Name: "D2"})
	repo.Create(&domain.Device{ID: "d3", UserID: "user2", Name: "D3"})

	list, err := service.List("user1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 devices, got %d", len(list))
	}
}

func TestDeviceService_Revoke(t *testing.T) {
	repo := newMockDeviceRepo()
	service := NewDeviceService(repo)

	repo.Create(&domain.Device{ID: "d1", UserID: "user1", Name: "D1"})

	err := service.Revoke("user1", "d1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	d, _ := repo.FindByID("d1")
	if !d.IsRevoked {
		t.Error("expected device to be revoked")
	}

	err = service.Revoke("user2", "d1")
	if err == nil {
		t.Error("expected unauthorized error")
	}
}
