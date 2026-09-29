# Yang Sudah Dikerjakan (Fase 1 — Auth, per 29 Sep 2026)

## 1. Infrastruktur
- `docker-compose.yml`: PostgreSQL 16, user/db `pos`/`pos_db`, port 5432.
- Go module `pos-app` di `backend/` (Gin, pgx v5, validator, bcrypt, JWT v5, google/uuid, goose).
- Frontend Vite + React + TypeScript + Tailwind v4 + shadcn/ui (terinstall, belum dipakai).

## 2. Migrasi (`backend/migrations`, goose)
- `00001_auth_schema.sql`: extension `pgcrypto`, schema `auth`, 7 tabel
  (`permissions`, `roles`, `role_permissions`, `users`, `user_roles`,
  `user_branches`, `audit_log`) — semua PK/FK UUID, plus index di `audit_log(user_id, created_at)`.
- `00002_seed_auth.sql`: 10 permission, 3 role (admin/manager/staff),
  mapping role→permission. Idempotent (`ON CONFLICT DO NOTHING`), Down selektif.

## 3. Kode backend (`backend/internal/`)
- `platform/config`: baca env (DB_HOST/PORT/USER/PASSWORD/NAME, JWT_SECRET).
- `platform/database`: koneksi `pgxpool.Pool`.
- `auth/model.go`: entity DB (User, Role, Permission, AuditLog), ID string (UUID).
- `auth/dto.go`: response DTO + mapper `ToResponse()` — `PasswordHash` tidak pernah
  keluar ke API; `detail` audit dirender sebagai JSON object dengan guard `json.Valid`.
- `auth/repository.go`: query pgx mentah; `DBTX` interface (terima Pool maupun Tx
  untuk transaksi lintas modul); `UserRepository` interface untuk mocking;
  `ErrUserNotFound` (tidak bocor `pgx.ErrNoRows`); slice selalu non-nil (`make`);
  `WriteAuditLog` default `"{}"` agar cast `::jsonb` tidak error.
- `auth/service.go`: `Register` (bcrypt + audit), `Login` (cek aktif, cek password,
  ambil permission dengan error diteruskan, JWT HS256 dengan `exp`/`iat`
  `NewNumericDate`), `List` (return DTO), `Permissions` (delegasi ke repo).

## 4. Keputusan yang terkunci
goose, UUID semua PK/FK, bcrypt, izin granular per aksi, relasi user↔cabang
many-to-many, nama proyek `pos-app`, docs di root `docs/`.
