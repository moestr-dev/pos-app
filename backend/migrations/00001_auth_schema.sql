-- +goose Up
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.permissions (
    id SERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE auth.roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE auth.role_permissions (
    role_id INT NOT NULL REFERENCES auth.roles(id),
    permission_id INT NOT NULL REFERENCES auth.permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE auth.users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auth.user_roles (
    user_id INT NOT NULL REFERENCES auth.users(id),
    role_id INT NOT NULL REFERENCES auth.roles(id),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE auth.user_branches (
    user_id INT NOT NULL REFERENCES auth.users(id),
    branch_id INT NOT NULL,
    PRIMARY KEY (user_id, branch_id)
);

CREATE TABLE auth.audit_log (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action TEXT NOT NULL,
    detail JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS auth.audit_log;
DROP TABLE IF EXISTS auth.user_branches;
DROP TABLE IF EXISTS auth.user_roles;
DROP TABLE IF EXISTS auth.users;
DROP TABLE IF EXISTS auth.role_permissions;
DROP TABLE IF EXISTS auth.roles;
DROP TABLE IF EXISTS auth.permissions;
DROP SCHEMA IF EXISTS auth;