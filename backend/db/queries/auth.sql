-- name: CreateUser :one
INSERT INTO auth.users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, username, email, password_hash, is_active, created_at, updated_at;

-- name: FindUserByEmail :one
SELECT id, username, email, password_hash, is_active, created_at, updated_at
FROM auth.users WHERE email = $1;

-- name: ListUsers :many
SELECT id, username, email, password_hash, is_active, created_at, updated_at
FROM auth.users ORDER BY created_at DESC;

-- name: UserPermissions :many
SELECT DISTINCT p.code
FROM auth.permissions p
JOIN auth.role_permissions rp ON rp.permission_id = p.id
JOIN auth.user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;

-- name: WriteAuditLog :exec
INSERT INTO auth.audit_log (user_id, action, detail)
VALUES (sqlc.arg(user_id), sqlc.arg(action), sqlc.arg(detail)::jsonb);
