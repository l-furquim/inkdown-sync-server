package service

import (
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"
	"inkdown-sync-server/pkg/hash"
	"inkdown-sync-server/pkg/jwt"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo          repository.UserRepository
	jwtSecret         string
	jwtExpiration     time.Duration
	refreshExpiration time.Duration
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExp, refreshExp time.Duration) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		jwtSecret:         jwtSecret,
		jwtExpiration:     jwtExp,
		refreshExpiration: refreshExp,
	}
}

func (s *AuthService) Register(req *domain.RegisterRequest) error {
	emailExists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return fmt.Errorf("email already registered")
	}

	usernameExists, err := s.userRepo.UsernameExists(req.Username)
	if err != nil {
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("username already taken")
	}

	hashedPassword, err := hash.Hash(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthService) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := hash.Compare(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, err := jwt.GenerateToken(user.ID, s.jwtExpiration, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID, s.refreshExpiration, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	user.Password = ""

	return &domain.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtExpiration.Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(req *domain.RefreshTokenRequest) (*domain.TokenResponse, error) {
	claims, err := jwt.ValidateToken(req.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	accessToken, err := jwt.GenerateToken(claims.UserID, s.jwtExpiration, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &domain.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.jwtExpiration.Seconds()),
	}, nil
}

func (s *AuthService) ValidateToken(token string) (*jwt.Claims, error) {
	claims, err := jwt.ValidateToken(token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}
