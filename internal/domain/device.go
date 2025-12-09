package domain

import "time"

type Device struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	OS         string    `json:"os"`
	AppVersion string    `json:"app_version"`
	LastActive time.Time `json:"last_active"`
	CreatedAt  time.Time `json:"created_at"`
	IsRevoked  bool      `json:"is_revoked"`
}

type RegisterDeviceRequest struct {
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	OS         string `json:"os" validate:"required"`
	AppVersion string `json:"app_version" validate:"required"`
}

type DeviceResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	OS         string    `json:"os"`
	LastActive time.Time `json:"last_active"`
	IsRevoked  bool      `json:"is_revoked"`
}
