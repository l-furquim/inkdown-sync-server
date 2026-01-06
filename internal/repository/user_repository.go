package repository

import (
	"context"
	"fmt"

	"inkdown-sync-server/internal/domain"

	"github.com/go-kivik/kivik/v4"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
	FindByUsername(username string) (*domain.User, error)
	Update(user *domain.User) error
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}

type userRepository struct {
	client *kivik.Client
	dbName string
}

func NewUserRepository(client *kivik.Client, dbName string) UserRepository {
	return &userRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *userRepository) Create(user *domain.User) error {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("user:%s", user.ID)
	_, err := db.Put(context.Background(), docID, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"email": email,
		},
		"limit": 1,
	}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	var user domain.User
	if err := rows.ScanDoc(&user); err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("user:%s", id)
	row := db.Get(context.Background(), docID)

	var user domain.User
	if err := row.ScanDoc(&user); err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	db := r.client.DB(r.dbName)

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"username": username,
		},
		"limit": 1,
	}

	rows := db.Find(context.Background(), query)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	var user domain.User
	if err := rows.ScanDoc(&user); err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	db := r.client.DB(r.dbName)

	docID := fmt.Sprintf("user:%s", user.ID)
	_, err := db.Put(context.Background(), docID, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepository) EmailExists(email string) (bool, error) {
	_, err := r.FindByEmail(email)
	if err != nil {
		if err.Error() == "user not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *userRepository) UsernameExists(username string) (bool, error) {
	_, err := r.FindByUsername(username)
	if err != nil {
		if err.Error() == "user not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
