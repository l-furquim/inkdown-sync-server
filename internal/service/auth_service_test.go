package service

import (
	"testing"
	"time"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/pkg/hash"
	. "inkdown-sync-server/pkg/jwt"
)

type mockUserRepository struct {
	users map[string]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, &userNotFoundError{}
}

func (m *mockUserRepository) FindByID(id string) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, &userNotFoundError{}
}

func (m *mockUserRepository) FindByUsername(username string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, &userNotFoundError{}
}

func (m *mockUserRepository) Update(user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) EmailExists(email string) (bool, error) {
	_, err := m.FindByEmail(email)
	return err == nil, nil
}

func (m *mockUserRepository) UsernameExists(username string) (bool, error) {
	_, err := m.FindByUsername(username)
	return err == nil, nil
}

type userNotFoundError struct{}

func (e *userNotFoundError) Error() string {
	return "user not found"
}

func TestAuthService_Register(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name    string
		req     *domain.RegisterRequest
		wantErr bool
		setup   func()
	}{
		{
			name: "successful registration",
			req: &domain.RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "Password123!",
			},
			wantErr: false,
			setup:   func() {},
		},
		{
			name: "duplicate email",
			req: &domain.RegisterRequest{
				Username: "anotheruser",
				Email:    "existing@example.com",
				Password: "Password123!",
			},
			wantErr: true,
			setup: func() {
				hashedPw, _ := hash.Hash("ExistingPass123!")
				repo.Create(&domain.User{
					ID:       "existing-id",
					Username: "existinguser",
					Email:    "existing@example.com",
					Password: hashedPw,
				})
			},
		},
		{
			name: "duplicate username",
			req: &domain.RegisterRequest{
				Username: "duplicateuser",
				Email:    "unique@example.com",
				Password: "Password123!",
			},
			wantErr: true,
			setup: func() {
				hashedPw, _ := hash.Hash("Pass123!")
				repo.Create(&domain.User{
					ID:       "dup-id",
					Username: "duplicateuser",
					Email:    "other@example.com",
					Password: hashedPw,
				})
			},
		},
		{
			name: "weak password",
			req: &domain.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
			wantErr: true,
			setup:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.users = make(map[string]*domain.User)
			tt.setup()

			err := service.Register(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Register() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Register() unexpected error = %v", err)
				}

				exists, _ := repo.EmailExists(tt.req.Email)
				if !exists {
					t.Error("Register() user not created in repository")
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	repo := newMockUserRepository()
	service := NewAuthService(repo, "test-secret-key", 15*time.Minute, 7*24*time.Hour)

	password := "UserPassword123!"
	hashedPassword, _ := hash.Hash(password)

	repo.Create(&domain.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	})

	tests := []struct {
		name    string
		req     *domain.LoginRequest
		wantErr bool
	}{
		{
			name: "successful login",
			req: &domain.LoginRequest{
				Email:    "test@example.com",
				Password: password,
			},
			wantErr: false,
		},
		{
			name: "wrong password",
			req: &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPassword",
			},
			wantErr: true,
		},
		{
			name: "non-existent email",
			req: &domain.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: password,
			},
			wantErr: true,
		},
		{
			name: "empty password",
			req: &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Login(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Login() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Login() unexpected error = %v", err)
				return
			}

			if resp.AccessToken == "" {
				t.Error("Login() returned empty access token")
			}

			if resp.RefreshToken == "" {
				t.Error("Login() returned empty refresh token")
			}

			if resp.User == nil {
				t.Error("Login() returned nil user")
			}

			if resp.User.Password != "" {
				t.Error("Login() returned user with password (security issue)")
			}

			if resp.ExpiresIn != int64(15*time.Minute.Seconds()) {
				t.Errorf("Login() expiresIn = %v, want %v", resp.ExpiresIn, 15*60)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	repo := newMockUserRepository()
	secret := "refresh-test-secret-key"
	service := NewAuthService(repo, secret, 15*time.Minute, 7*24*time.Hour)

	repo.Create(&domain.User{
		ID:       "refresh-user-id",
		Username: "refreshuser",
		Email:    "refresh@example.com",
		Password: "hashed",
	})

	validToken, _ := GenerateRefreshToken("refresh-user-id", 7*24*time.Hour, secret)
	expiredToken, _ := GenerateRefreshToken("refresh-user-id", -1*time.Hour, secret)

	tests := []struct {
		name    string
		req     *domain.RefreshTokenRequest
		wantErr bool
	}{
		{
			name: "valid refresh token",
			req: &domain.RefreshTokenRequest{
				RefreshToken: validToken,
			},
			wantErr: false,
		},
		{
			name: "expired refresh token",
			req: &domain.RefreshTokenRequest{
				RefreshToken: expiredToken,
			},
			wantErr: true,
		},
		{
			name: "invalid refresh token",
			req: &domain.RefreshTokenRequest{
				RefreshToken: "invalid.token.here",
			},
			wantErr: true,
		},
		{
			name: "empty refresh token",
			req: &domain.RefreshTokenRequest{
				RefreshToken: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.RefreshToken(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("RefreshToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("RefreshToken() unexpected error = %v", err)
				return
			}

			if resp.AccessToken == "" {
				t.Error("RefreshToken() returned empty access token")
			}

			if resp.ExpiresIn != int64(15*time.Minute.Seconds()) {
				t.Errorf("RefreshToken() expiresIn = %v, want %v", resp.ExpiresIn, 15*60)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	repo := newMockUserRepository()
	secret := "validation-test-secret"
	service := NewAuthService(repo, secret, 15*time.Minute, 7*24*time.Hour)

	validToken, _ := GenerateToken("user-id", 1*time.Hour, secret)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error = %v", err)
				return
			}

			if claims == nil {
				t.Error("ValidateToken() returned nil claims")
			}
		})
	}
}
