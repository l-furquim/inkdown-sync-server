package jwt

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		expiration time.Duration
		secret     string
		wantErr    bool
	}{
		{
			name:       "valid token generation",
			userID:     "user-123",
			expiration: 15 * time.Minute,
			secret:     "test-secret-key-32-characters!",
			wantErr:    false,
		},
		{
			name:       "short expiration",
			userID:     "user-456",
			expiration: 1 * time.Second,
			secret:     "test-secret",
			wantErr:    false,
		},
		{
			name:       "long expiration",
			userID:     "user-789",
			expiration: 24 * time.Hour,
			secret:     "test-secret",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.expiration, tt.secret)

			if tt.wantErr {
				if err == nil {
					t.Error("GenerateToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateToken() error = %v", err)
				return
			}

			if token == "" {
				t.Error("GenerateToken() returned empty token")
			}

			if len(token) < 100 {
				t.Errorf("GenerateToken() token too short, len = %d", len(token))
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := "user-refresh-test"
	expiration := 7 * 24 * time.Hour
	secret := "refresh-secret-key"

	token, err := GenerateRefreshToken(userID, expiration, secret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}
}

func TestValidateToken(t *testing.T) {
	userID := "test-user-id"
	secret := "validation-secret-key-32-chars"

	validToken, _ := GenerateToken(userID, 1*time.Hour, secret)
	expiredToken, _ := GenerateToken(userID, -1*time.Hour, secret)

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
		checkID bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			secret:  secret,
			wantErr: false,
			checkID: true,
		},
		{
			name:    "expired token",
			token:   expiredToken,
			secret:  secret,
			wantErr: true,
			checkID: false,
		},
		{
			name:    "wrong secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: true,
			checkID: false,
		},
		{
			name:    "invalid token format",
			token:   "invalid.token.format",
			secret:  secret,
			wantErr: true,
			checkID: false,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
			checkID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() error = %v", err)
				return
			}

			if claims == nil {
				t.Error("ValidateToken() returned nil claims")
				return
			}

			if tt.checkID && claims.UserID != userID {
				t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, userID)
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	userID := "expiration-test-user"
	secret := "expiration-test-secret"

	token, err := GenerateToken(userID, 1*time.Second, secret)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() immediate validation error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, userID)
	}

	time.Sleep(2 * time.Second)

	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Error("ValidateToken() expected error for expired token")
	}
}

func TestClaimsTimestamps(t *testing.T) {
	userID := "timestamp-test-user"
	secret := "timestamp-test-secret"
	expiration := 1 * time.Hour

	before := time.Now().Add(-1 * time.Second)
	token, err := GenerateToken(userID, expiration, secret)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	after := time.Now().Add(1 * time.Second)

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	issuedAt := claims.IssuedAt.Time
	if issuedAt.Before(before) || issuedAt.After(after) {
		t.Errorf("IssuedAt timestamp out of expected range: got %v, range [%v, %v]",
			issuedAt, before, after)
	}

	notBefore := claims.NotBefore.Time
	if notBefore.Before(before) || notBefore.After(after) {
		t.Errorf("NotBefore timestamp out of expected range: got %v, range [%v, %v]",
			notBefore, before, after)
	}

	expiresAt := claims.ExpiresAt.Time
	expectedExpiry := before.Add(expiration)
	upperBound := after.Add(expiration)
	if expiresAt.Before(expectedExpiry) || expiresAt.After(upperBound) {
		t.Errorf("ExpiresAt timestamp out of expected range: got %v, range [%v, %v]",
			expiresAt, expectedExpiry, upperBound)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	userID := "benchmark-user"
	expiration := 15 * time.Minute
	secret := "benchmark-secret-key"

	for i := 0; i < b.N; i++ {
		_, err := GenerateToken(userID, expiration, secret)
		if err != nil {
			b.Fatalf("GenerateToken() error = %v", err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	userID := "benchmark-user"
	expiration := 15 * time.Minute
	secret := "benchmark-secret-key"

	token, _ := GenerateToken(userID, expiration, secret)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ValidateToken(token, secret)
		if err != nil {
			b.Fatalf("ValidateToken() error = %v", err)
		}
	}
}
