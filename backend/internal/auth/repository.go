package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbsqlc "pos-app/db/sqlc"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	WriteAuditLog(ctx context.Context, userID *uuid.UUID, action, detail string) error
}

type Repository struct {
	q *dbsqlc.Queries
}

func NewRepository(db dbsqlc.DBTX) *Repository {
	return &Repository{q: dbsqlc.New(db)}
}

func toUser(row dbsqlc.AuthUser) User {
	return User{
		ID:           row.ID,
		Username:     row.Username,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toPgUUID(id *uuid.UUID) (pgtype.UUID, error) {
	if id == nil {
		return pgtype.UUID{Valid: false}, nil
	}
	var out pgtype.UUID
	if err := out.Scan(id.String()); err != nil {
		return pgtype.UUID{}, err
	}
	return out, nil
}

func (r *Repository) CreateUser(ctx context.Context, username, email, passwordHash string) (*User, error) {
	row, err := r.q.CreateUser(ctx, dbsqlc.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}
	u := toUser(row)
	return &u, nil
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	row, err := r.q.FindUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u := toUser(row)
	return &u, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toUser(row))
	}
	return users, nil
}

func (r *Repository) UserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.q.UserPermissions(ctx, userID)
}

func (r *Repository) WriteAuditLog(ctx context.Context, userID *uuid.UUID, action, detail string) error {
	if detail == "" {
		detail = "{}"
	}
	pgID, err := toPgUUID(userID)
	if err != nil {
		return err
	}
	return r.q.WriteAuditLog(ctx, dbsqlc.WriteAuditLogParams{
		UserID: pgID,
		Action: action,
		Detail: []byte(detail),
	})
}
