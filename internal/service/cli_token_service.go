package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/repository"
	"inkdown-sync-server/pkg/hash"

	"github.com/google/uuid"
)

type CLITokenService struct {
	tokenRepo repository.CLITokenRepository
	userRepo  repository.UserRepository
}

func NewCLITokenService(tokenRepo repository.CLITokenRepository, userRepo repository.UserRepository) *CLITokenService {
	return &CLITokenService{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
	}
}

// generateSecureToken creates a cryptographically secure random token
// Format: ink_<random 40 chars> (total 44 chars)
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return "ink_" + hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// LoginAndCreateToken authenticates a user and creates a new CLI token
func (s *CLITokenService) LoginAndCreateToken(req *domain.CLILoginRequest) (*domain.CreateCLITokenResponse, error) {
	// Authenticate user
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := hash.Compare(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create token for authenticated user
	createReq := &domain.CreateCLITokenRequest{
		Name:   req.Name,
		Scopes: domain.DefaultCLIScopes(),
	}

	return s.CreateToken(user.ID, createReq)
}

// CreateToken creates a new CLI token for a user (requires authentication)
func (s *CLITokenService) CreateToken(userID string, req *domain.CreateCLITokenRequest) (*domain.CreateCLITokenResponse, error) {
	// Verify user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate secure token
	plainToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash token for storage
	hashedToken := hashToken(plainToken)

	// Set default scopes if none provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = domain.DefaultCLIScopes()
	}

	// Create token record
	token := &domain.CLIToken{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Token:       hashedToken,
		TokenPrefix: plainToken[:12], // "ink_" + first 8 chars of random
		Scopes:      scopes,
		CreatedAt:   time.Now(),
		IsRevoked:   false,
	}

	if err := s.tokenRepo.Create(token); err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &domain.CreateCLITokenResponse{
		ID:          token.ID,
		Name:        token.Name,
		Token:       plainToken, // Return the plain token ONLY ONCE
		TokenPrefix: token.TokenPrefix,
		Scopes:      token.Scopes,
		CreatedAt:   token.CreatedAt,
		Message:     "Token created successfully. Store it safely - it won't be shown again!",
	}, nil
}

// ValidateToken validates a CLI token and returns the associated user
func (s *CLITokenService) ValidateToken(plainToken string) (*domain.User, *domain.CLIToken, error) {
	// Hash the provided token
	hashedToken := hashToken(plainToken)

	// Find token in database
	token, err := s.tokenRepo.FindByToken(hashedToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid or revoked token")
	}

	// Check if token is revoked
	if token.IsRevoked {
		return nil, nil, fmt.Errorf("token has been revoked")
	}

	// Get user
	user, err := s.userRepo.FindByID(token.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	// Clear sensitive data
	user.Password = ""

	return user, token, nil
}

// ValidateTokenWithScope validates a token and checks for a specific scope
func (s *CLITokenService) ValidateTokenWithScope(plainToken string, requiredScope string) (*domain.User, *domain.CLIToken, error) {
	user, token, err := s.ValidateToken(plainToken)
	if err != nil {
		return nil, nil, err
	}

	// Check if token has the required scope
	hasScope := false
	for _, scope := range token.Scopes {
		if scope == requiredScope {
			hasScope = true
			break
		}
	}

	if !hasScope {
		return nil, nil, fmt.Errorf("token does not have required scope: %s", requiredScope)
	}

	return user, token, nil
}

// UpdateLastUsed updates the last used timestamp and IP
func (s *CLITokenService) UpdateLastUsed(tokenID string, ip string) error {
	return s.tokenRepo.UpdateLastUsed(tokenID, ip)
}

// ListTokens returns all CLI tokens for a user (without the actual token values)
func (s *CLITokenService) ListTokens(userID string) ([]*domain.CLITokenPublic, error) {
	tokens, err := s.tokenRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}

	publicTokens := make([]*domain.CLITokenPublic, len(tokens))
	for i, token := range tokens {
		publicTokens[i] = token.ToPublic()
	}

	return publicTokens, nil
}

// RevokeToken revokes a CLI token
func (s *CLITokenService) RevokeToken(userID string, tokenID string) error {
	// Verify token belongs to user
	token, err := s.tokenRepo.FindByID(tokenID)
	if err != nil {
		return fmt.Errorf("token not found")
	}

	if token.UserID != userID {
		return fmt.Errorf("token does not belong to user")
	}

	return s.tokenRepo.Revoke(tokenID)
}

// DeleteToken permanently deletes a CLI token
func (s *CLITokenService) DeleteToken(userID string, tokenID string) error {
	// Verify token belongs to user
	token, err := s.tokenRepo.FindByID(tokenID)
	if err != nil {
		return fmt.Errorf("token not found")
	}

	if token.UserID != userID {
		return fmt.Errorf("token does not belong to user")
	}

	return s.tokenRepo.Delete(tokenID)
}

// GetToken returns a specific token (public info only)
func (s *CLITokenService) GetToken(userID string, tokenID string) (*domain.CLITokenPublic, error) {
	token, err := s.tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, fmt.Errorf("token not found")
	}

	if token.UserID != userID {
		return nil, fmt.Errorf("token does not belong to user")
	}

	return token.ToPublic(), nil
}
