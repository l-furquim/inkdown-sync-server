package service

import "inkdown-sync-server/internal/domain"

type ConflictError struct {
	Conflict *domain.Conflict
}

func (e *ConflictError) Error() string {
	return "conflict detected"
}
