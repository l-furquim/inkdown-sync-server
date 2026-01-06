package service

import (
	"errors"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"

	"github.com/google/uuid"
)

type DeviceService struct {
	repo repository.DeviceRepository
}

func NewDeviceService(repo repository.DeviceRepository) *DeviceService {
	return &DeviceService{
		repo: repo,
	}
}

func (s *DeviceService) Register(userID string, req *domain.RegisterDeviceRequest) (*domain.DeviceResponse, error) {
	// TODO: Check if device limit is reached (optional future feature)

	deviceID := uuid.New().String()
	now := time.Now()

	device := &domain.Device{
		ID:         deviceID,
		UserID:     userID,
		Name:       req.Name,
		Type:       req.Type,
		OS:         req.OS,
		AppVersion: req.AppVersion,
		LastActive: now,
		CreatedAt:  now,
		IsRevoked:  false,
	}

	if err := s.repo.Create(device); err != nil {
		return nil, err
	}

	return &domain.DeviceResponse{
		ID:         device.ID,
		Name:       device.Name,
		Type:       device.Type,
		OS:         device.OS,
		LastActive: device.LastActive,
		IsRevoked:  device.IsRevoked,
	}, nil
}

func (s *DeviceService) List(userID string) ([]*domain.DeviceResponse, error) {
	devices, err := s.repo.List(userID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.DeviceResponse
	for _, d := range devices {
		responses = append(responses, &domain.DeviceResponse{
			ID:         d.ID,
			Name:       d.Name,
			Type:       d.Type,
			OS:         d.OS,
			LastActive: d.LastActive,
			IsRevoked:  d.IsRevoked,
		})
	}

	return responses, nil
}

func (s *DeviceService) Revoke(userID, deviceID string) error {
	// Verify device belongs to user
	device, err := s.repo.FindByID(deviceID)
	if err != nil {
		return err
	}

	if device.UserID != userID {
		return errors.New("unauthorized: device does not belong to user")
	}

	return s.repo.Revoke(deviceID)
}

func (s *DeviceService) UpdateLastActive(deviceID string) error {
	return s.repo.UpdateLastActive(deviceID)
}
