package service

import (
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"
)

type SecurityService struct {
	repo repository.KeyStoreRepository
}

func NewSecurityService(repo repository.KeyStoreRepository) *SecurityService {
	return &SecurityService{
		repo: repo,
	}
}

func (s *SecurityService) UploadKey(userID string, req *domain.UploadKeyRequest) error {
	now := time.Now()

	key := &domain.EncryptedMasterKey{
		UserID:         userID,
		EncryptedKey:   req.EncryptedKey,
		KeySalt:        req.KeySalt,
		KDFParams:      req.KDFParams,
		EncryptionAlgo: req.EncryptionAlgo,
		CreatedAt:      now, // Repository handles update vs create logic for timestamps usually, but here we set it.
		UpdatedAt:      now,
	}

	return s.repo.Save(key)
}

func (s *SecurityService) GetKey(userID string) (*domain.KeyResponse, error) {
	key, err := s.repo.Get(userID)
	if err != nil {
		return nil, err
	}

	return &domain.KeyResponse{
		EncryptedKey:   key.EncryptedKey,
		KeySalt:        key.KeySalt,
		KDFParams:      key.KDFParams,
		EncryptionAlgo: key.EncryptionAlgo,
		UpdatedAt:      key.UpdatedAt,
	}, nil
}
