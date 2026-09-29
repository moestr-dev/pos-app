package auth

import "time"

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	ID          string
	Name        string
	Description string
}

type Permission struct {
	ID          string
	Code        string
	Description string
}

type AuditLog struct {
	ID        string
	UserID    *string
	Action    string
	Detail    string
	CreatedAt time.Time
}
