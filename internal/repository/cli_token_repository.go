package repository

import (
	"context"
	"fmt"
	"time"

	"inkdown-sync-server/internal/domain"

	"github.com/go-kivik/kivik/v4"
)

type CLITokenRepository interface {
	Create(token *domain.CLIToken) error
	FindByID(id string) (*domain.CLIToken, error)
	FindByToken(hashedToken string) (*domain.CLIToken, error)
	FindByUserID(userID string) ([]*domain.CLIToken, error)
	UpdateLastUsed(id string, ip string) error
	Revoke(id string) error
	Delete(id string) error
}

type cliTokenRepository struct {
	client *kivik.Client
	dbName string
}

func NewCLITokenRepository(client *kivik.Client, dbName string) CLITokenRepository {
	return &cliTokenRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *cliTokenRepository) Create(token *domain.CLIToken) error {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("cli_token:%s", token.ID)
	_, err := db.Put(context.Background(), docID, token)
	if err != nil {
		return fmt.Errorf("failed to create CLI token: %w", err)
	}

	return nil
}

func (r *cliTokenRepository) FindByID(id string) (*domain.CLIToken, error) {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("cli_token:%s", id)
	row := db.Get(context.Background(), docID)

	var token domain.CLIToken
	if err := row.ScanDoc(&token); err != nil {
		return nil, fmt.Errorf("CLI token not found: %w", err)
	}

	return &token, nil
}

func (r *cliTokenRepository) FindByToken(hashedToken string) (*domain.CLIToken, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"token":      hashedToken,
			"is_revoked": false,
		},
		"limit": 1,
	}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query CLI token: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("CLI token not found or revoked")
	}

	var token domain.CLIToken
	if err := rows.ScanDoc(&token); err != nil {
		return nil, fmt.Errorf("failed to scan CLI token: %w", err)
	}

	return &token, nil
}

func (r *cliTokenRepository) FindByUserID(userID string) ([]*domain.CLIToken, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"user_id": userID,
		},
		"sort": []map[string]string{
			{"created_at": "desc"},
		},
	}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query CLI tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.CLIToken
	for rows.Next() {
		var token domain.CLIToken
		if err := rows.ScanDoc(&token); err != nil {
			return nil, fmt.Errorf("failed to scan CLI token: %w", err)
		}
		tokens = append(tokens, &token)
	}

	return tokens, nil
}

func (r *cliTokenRepository) UpdateLastUsed(id string, ip string) error {
	token, err := r.FindByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	token.LastUsedAt = &now
	token.LastUsedIP = ip

	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("cli_token:%s", id)
	_, err = db.Put(context.Background(), docID, token)
	if err != nil {
		return fmt.Errorf("failed to update CLI token: %w", err)
	}

	return nil
}

func (r *cliTokenRepository) Revoke(id string) error {
	token, err := r.FindByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	token.IsRevoked = true
	token.RevokedAt = &now

	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("cli_token:%s", id)
	_, err = db.Put(context.Background(), docID, token)
	if err != nil {
		return fmt.Errorf("failed to revoke CLI token: %w", err)
	}

	return nil
}

func (r *cliTokenRepository) Delete(id string) error {
	db := r.client.DB(r.dbName)
	docID := fmt.Sprintf("cli_token:%s", id)

	row := db.Get(context.Background(), docID)
	var doc map[string]interface{}
	if err := row.ScanDoc(&doc); err != nil {
		return fmt.Errorf("CLI token not found: %w", err)
	}

	rev, ok := doc["_rev"].(string)
	if !ok {
		return fmt.Errorf("failed to get document revision")
	}

	_, err := db.Delete(context.Background(), docID, rev)
	if err != nil {
		return fmt.Errorf("failed to delete CLI token: %w", err)
	}

	return nil
}
