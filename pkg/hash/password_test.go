package hash

import (
	"strings"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:     "minimum length password",
			password: "Pass123!",
			wantErr:  false,
		},
		{
			name:     "password too short",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := Hash(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Hash() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Hash() unexpected error = %v", err)
				return
			}

			if hash == "" {
				t.Error("Hash() returned empty hash")
			}

			if hash == tt.password {
				t.Error("Hash() returned unhashed password")
			}

			if !strings.HasPrefix(hash, "$2a$12$") {
				t.Errorf("Hash() invalid bcrypt format, got = %s", hash[:10])
			}
		})
	}
}

func TestHashDifferentOutputs(t *testing.T) {
	password := "SamePassword123!"

	hash1, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	hash2, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash() should generate different hashes for same password (salt)")
	}
}

func TestCompare(t *testing.T) {
	password := "MySecurePassword123!"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hash,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hash,
			password:       "WrongPassword",
			wantErr:        true,
		},
		{
			name:           "empty password",
			hashedPassword: hash,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "case sensitive",
			hashedPassword: hash,
			password:       strings.ToUpper(password),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Compare(tt.hashedPassword, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("Compare() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Compare() unexpected error = %v", err)
				}
			}
		})
	}
}

func BenchmarkHash(b *testing.B) {
	password := "BenchmarkPassword123!"

	for i := 0; i < b.N; i++ {
		_, err := Hash(password)
		if err != nil {
			b.Fatalf("Hash() error = %v", err)
		}
	}
}

func BenchmarkCompare(b *testing.B) {
	password := "BenchmarkPassword123!"
	hash, _ := Hash(password)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Compare(hash, password)
	}
}
