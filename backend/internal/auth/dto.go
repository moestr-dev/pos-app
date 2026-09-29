package auth

import (
	"encoding/json"
	"time"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type AuditLogResponse struct {
	ID        string          `json:"id"`
	UserID    *string         `json:"user_id,omitempty"`
	Action    string          `json:"action"`
	Detail    json.RawMessage `json:"detail,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (r *Role) ToResponse() RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}

func (p *Permission) ToResponse() PermissionResponse {
	return PermissionResponse{
		ID:          p.ID,
		Code:        p.Code,
		Description: p.Description,
	}
}

func (a *AuditLog) ToResponse() AuditLogResponse {
	var detailJSON json.RawMessage
	if a.Detail != "" && json.Valid([]byte(a.Detail)) {
		detailJSON = json.RawMessage(a.Detail)
	}

	return AuditLogResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Action:    a.Action,
		Detail:    detailJSON,
		CreatedAt: a.CreatedAt,
	}
}
