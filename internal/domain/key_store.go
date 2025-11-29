package domain

import "time"

// EncryptedMasterKey represents the user's master key encrypted by their password/recovery key.
// The server stores this blob but cannot decrypt it.
type EncryptedMasterKey struct {
	UserID         string    `json:"user_id"`
	EncryptedKey   string    `json:"encrypted_key"`   // A chave mestra criptografada (Base64)
	KeySalt        string    `json:"key_salt"`        // Salt usado para derivar a chave da senha (Base64)
	KDFParams      string    `json:"kdf_params"`      // Parâmetros do Argon2id/Scrypt (JSON string)
	EncryptionAlgo string    `json:"encryption_algo"` // Ex: "AES-256-GCM"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UploadKeyRequest represents the payload to upload the encrypted master key
type UploadKeyRequest struct {
	EncryptedKey   string `json:"encrypted_key" validate:"required"`
	KeySalt        string `json:"key_salt" validate:"required"`
	KDFParams      string `json:"kdf_params" validate:"required"`
	EncryptionAlgo string `json:"encryption_algo" validate:"required"`
}

// KeyResponse represents the key data returned to the client
type KeyResponse struct {
	EncryptedKey   string    `json:"encrypted_key"`
	KeySalt        string    `json:"key_salt"`
	KDFParams      string    `json:"kdf_params"`
	EncryptionAlgo string    `json:"encryption_algo"`
	UpdatedAt      time.Time `json:"updated_at"`
}
