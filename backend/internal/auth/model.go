package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type Permission struct {
	ID          uuid.UUID
	Code        string
	Description string
}

type AuditLog struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Action    string
	Detail    string
	CreatedAt time.Time
}
