# Rencana Berikutnya (Fase 1 — Auth → Inventory)

## Auth (sisa)
- [ ] `platform/middleware/auth.go`: baca JWT dari cookie `httpOnly`, tolak jika
      expired/tidak valid, simpan `user_id` + `perms` di context; helper `RequirePermission(code)`.
- [ ] `auth/handler.go`: `POST /auth/register`, `POST /auth/login` (set cookie `httpOnly`),
      `POST /auth/logout` (hapus cookie), `GET /users`, `GET /users/me/permissions`,
      CRUD role + assign permission, `GET /audit-log` — semua (kecuali login) di belakang middleware.
- [ ] `cmd/api/main.go`: rakit config → pool → repo → service → handler → Gin;
      proteksi endpoint by permission; CORS terbatas; `r.Run(":8080")`.
- [ ] Test manual via curl (register → login → akses endpoint terproteksi → cek audit_log).
- [ ] Frontend: halaman login, manajemen user/role (guard by permission), penanganan sesi habis.

## Inventory (modul kedua)
- [ ] Migrasi schema `inventory`: `barang`, `satuan`, `cabang`, `stok`, `kartu_stok`,
      `transfer_stok(+item)`, `opname(+item)` — UUID, tanpa FK ke schema `auth`.
- [ ] Service transfer atomik + cegah stok minus (`UPDATE ... WHERE qty >= n`), kartu stok tiap mutasi.
- [ ] Frontend: barang, cabang, stok, kartu stok, transfer, opname.

## Operasional (Tahap 4 PRD)
- [ ] `docker-compose up` menjalankan backend + frontend + db; env-based config; log terstruktur.
- [ ] CI GitHub Actions (test otomatis); deploy publik; README arsitektur.
