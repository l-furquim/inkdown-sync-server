package domain

import "time"

type EncryptedMasterKey struct {
	UserID         string    `json:"user_id"`
	EncryptedKey   string    `json:"encrypted_key"`
	KeySalt        string    `json:"key_salt"`
	KDFParams      string    `json:"kdf_params"`
	EncryptionAlgo string    `json:"encryption_algo"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UploadKeyRequest struct {
	EncryptedKey   string `json:"encrypted_key" validate:"required"`
	KeySalt        string `json:"key_salt" validate:"required"`
	KDFParams      string `json:"kdf_params" validate:"required"`
	EncryptionAlgo string `json:"encryption_algo" validate:"required"`
}

type KeyResponse struct {
	EncryptedKey   string    `json:"encrypted_key"`
	KeySalt        string    `json:"key_salt"`
	KDFParams      string    `json:"kdf_params"`
	EncryptionAlgo string    `json:"encryption_algo"`
	UpdatedAt      time.Time `json:"updated_at"`
}
