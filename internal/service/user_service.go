package service

import (
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetByID(id string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateUsername(userID, newUsername string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	usernameExists, err := s.userRepo.UsernameExists(newUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if usernameExists && user.Username != newUsername {
		return nil, fmt.Errorf("username already taken")
	}

	user.Username = newUsername
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.Password = ""
	return user, nil
}
