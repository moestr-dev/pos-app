# PRD: POS Distributor/Grosir Multi-Cabang (Go, Modular Monolith)

Versi 0.4 | 29 September 2026 | Status: fondasi auth dikerjakan


---


## 1. Tujuan


**Tujuan utama:** membangun satu proyek dengan pemrograman Go yang cukup dalam dapat didemokan bahkan kedepan memungkinkan untuk diimplementasikan.


**Tujuan pendukung:** proyek ini jadi fondasi produk POS untuk segmen distributor/grosir/multi-cabang, memakai domain procurement dengan approval, sales dengan diskon, return, laporan, akuntansi, dan lain sebagainya.


**Kenapa proyek ini dibuat:** untuk konsistensi transaksi, concurrency (dua orang mengubah stok barang yang sama bersamaan), dan keputusan arsitektur yang bisa dijalankan.

---


## 2. Keputusan Teknis


| Area | Keputusan | Status |
|---|---|---|
| Bahasa | Go | Diputuskan |
| Arsitektur | Modular monolith (satu aplikasi, satu deploy) | Diputuskan |
| Database | PostgreSQL 16 | Diputuskan |
| Pemisahan data | Satu database, satu schema per modul | Diputuskan |
| Primary key | UUID (`gen_random_uuid()` via `pgcrypto`) | Diputuskan v0.4 |
| HTTP | Gin | Diputuskan |
| Akses DB | pgx + sqlc | Diputuskan |
| Validasi dan config | go-playground/validator, env-based config | Diputuskan |
| Auth | JWT, disimpan di cookie `httpOnly` | Diputuskan |
| Password hashing | bcrypt | Diputuskan v0.4 |
| Frontend | React + TypeScript + Vite | Diputuskan |
| UI frontend | Tailwind + shadcn/ui | Diputuskan |
| Data fetching dan routing | TanStack Query + React Router | Diputuskan |
| Tipe API di frontend | OpenAPI codegen | Opsional |
| Kontainer | Docker + docker-compose | Diputuskan |
| CI | GitHub Actions | Diputuskan |
| Migrasi | goose, direktori `backend/migrations` | Diputuskan v0.4 |
| Nama proyek/repo | pos-app | Diputuskan v0.4 |


**Kenapa bukan microservice:** transaksi lintas modul harus atomik dalam satu database. Memecah jadi banyak service menuntut logika pembatalan (saga) dan menghilangkan `JOIN` untuk laporan.


**Kenapa PostgreSQL:** punya schema sungguhan, tipe `NUMERIC` untuk uang, dan cocok dengan pgx/sqlc.


**Kenapa UUID:** ID tidak bisa ditebak, aman untuk merge data antar cabang, dan tidak bocor informasi jumlah record seperti auto-increment.


**Kenapa goose:** simpel (file SQL murni), mudah di-review, cocok untuk portofolio. Migrasi disimpan di `backend/migrations` sejajar dengan `go.mod`.


**Kenapa frontend dipisah (SPA):** backend Go murni berupa API, frontend berdiri sendiri. Frontend dijaga tipis karena fokus portofolio adalah backend.


---


## 3. Ruang Lingkup


### Fase 1 (yang dikerjakan sekarang)
1. **Auth dan user management:** login, user, peran, izin granular per aksi, log audit. Dikerjakan lebih dulu karena akan dipakai approval di fase berikutnya.
2. **Inventory (dibuat dalam):** master barang, satuan, cabang, stok per cabang, kartu stok, pengurangan stok yang aman dari minus, transfer stok antar cabang, stok opname.
3. **Frontend:** halaman login, manajemen user/peran, master barang dan cabang, daftar stok, kartu stok, transfer, dan opname.


Kedua modul memakai schema terpisah: `auth` dan `inventory`.


### Fase berikutnya (belum dikerjakan)
- Purchasing dengan approval (memakai izin dari modul auth)
- Sales: diskon, return
- Accounting: jurnal otomatis dan laporan


### Di luar cakupan
- Aplikasi mobile, integrasi marketplace/pembayaran digital, printer struk
- Multi-tenant SaaS dan penagihan langganan
- Pemisahan menjadi microservice
- Valuasi/costing stok (fase 1 hanya mencatat kuantitas)


---


## 4. Arsitektur


### Struktur folder (aktual)


```
pos-app/
├── docker-compose.yml      // PostgreSQL 16
├── backend/
│   ├── cmd/api/main.go     // titik masuk, merakit semua modul
│   ├── internal/auth/      // handler, service, repository
│   ├── internal/inventory/ // (fase berikutnya)
│   ├── internal/platform/  // database, config, logger, middleware
│   ├── migrations/         // goose (00001_auth_schema, 00002_seed_auth, ...)
│   ├── docs/               // (dipindah ke root: docs/)
│   └── go.mod              // module pos-app
└── frontend/               // React + TypeScript + Vite
└── docs/                   // PRD dan dokumentasi (file ini)
```


Modul purchasing, sales, dan accounting ditambahkan di `backend/internal/` pada fase berikutnya dengan pola yang sama.


### Aturan modul (wajib dipatuhi)
1. Modul hanya dipanggil lewat fungsi/interface yang ia buka; tidak boleh mengakses tabel modul lain langsung.
2. Tiap modul memakai schema sendiri (`auth`, `inventory`); tidak ada query ke schema modul lain.
3. Tidak ada foreign key lintas schema. Referensi antar modul disimpan sebagai ID biasa (mis. `user_id` UUID di kartu stok, `branch_id` UUID di `auth.user_branches`).
4. Transaksi lintas modul memakai satu `*sql.Tx` / `pgx.Tx` yang dioper antar fungsi, sehingga semuanya berhasil atau dibatalkan bersama.
5. Query yang sengaja melintasi schema (untuk laporan nanti) adalah satu-satunya pengecualian dan harus ditulis di dokumentasi.


### Relasi user ↔ cabang (diputuskan v0.4)
- Many-to-many via `auth.user_branches (user_id, branch_id)`.
- Ada user yang terikat satu cabang (kasir), ada yang multi-cabang (supervisor/owner).
- `branch_id` di sini adalah UUID yang merujuk ke `inventory.cabang` secara logis (tanpa FK lintas schema).


### Izin granular (diputuskan v0.4)
- Endpoint dilindungi berdasarkan kode izin (mis. `inventory.transfer`), bukan nama peran.
- Seed awal: `user.read/create/update/delete`, `role.read/manage`, `inventory.read/manage/transfer/opname`.
- Role awal: `admin` (semua izin), `manager`, `staff` (`inventory.read`).


### Contoh alur: transfer stok antar cabang
1. `POST /transfers` diterima modul inventory; middleware auth memeriksa izin `inventory.transfer`
2. Buka transaksi database
3. Kurangi stok cabang asal dengan syarat `qty >= jumlah` (mencegah stok minus)
4. Tambah stok cabang tujuan
5. Tulis dua baris kartu stok beserta `user_id` pelaku
6. Commit; jika ada langkah gagal, rollback semuanya


---


## 5. Kebutuhan Fungsional


**Auth dan user management**
- FR-AUT-1: Login menghasilkan JWT yang disimpan di cookie `httpOnly`.
- FR-AUT-2: CRUD user; user bisa dinonaktifkan tanpa dihapus.
- FR-AUT-3: Peran memiliki kumpulan izin granular per aksi (mis. `inventory.adjust`, `inventory.transfer`, `po.approve`), dan endpoint dilindungi berdasarkan izin, bukan nama peran. Dengan begitu approval di fase berikutnya tinggal memakai izin yang sama.
- FR-AUT-4: Aksi penting (perubahan peran, koreksi stok, opname, transfer) tercatat di log audit: siapa, kapan, apa.
- FR-AUT-5: Password disimpan dengan hash bcrypt, tidak pernah dalam bentuk asli.
- FR-AUT-6: User dapat dipasangkan ke satu atau banyak cabang via `auth.user_branches`.


**Inventory**
- FR-INV-1: CRUD barang, satuan, dan cabang.
- FR-INV-2: Stok dicatat per barang per cabang.
- FR-INV-3: Setiap perubahan stok menulis satu baris kartu stok (siapa, kapan, alasan, sebelum/sesudah).
- FR-INV-4: Stok tidak boleh minus, termasuk saat banyak request bersamaan.
- FR-INV-5: Transfer stok antar cabang bersifat atomik.
- FR-INV-6: Stok opname mencatat selisih dan alasan, lalu menyesuaikan stok lewat kartu stok.


**Frontend**
- FR-FE-1: Halaman login dan penanganan sesi habis.
- FR-FE-2: Halaman user dan peran, hanya tampil untuk yang berizin.
- FR-FE-3: Halaman barang, cabang, stok, kartu stok, transfer, dan opname dengan tabel, pencarian, dan form validasi.
- FR-FE-4: Menu dan tombol menyesuaikan izin user yang sedang login.


---


## 6. Kebutuhan Non-Fungsional


- **Konsistensi:** transaksi atomik dibuktikan dengan test rollback.
- **Concurrency:** uji ratusan request bersamaan pada barang yang sama; stok tidak boleh minus dan jumlah akhir harus tepat.
- **Keamanan dasar:** password di-hash (bcrypt), cookie `httpOnly`, validasi input di backend, CORS dibatasi, ID bertipe UUID.
- **Waktu:** simpan sebagai UTC (`TIMESTAMPTZ`), tampilkan sesuai zona Asia/Jakarta.
- **Test:** unit test untuk service, integration test dengan database asli lewat Docker.
- **Operasional:** berjalan dengan `docker-compose up`, konfigurasi lewat environment variable, log terstruktur.
- **Dokumentasi:** README menjelaskan arsitektur dan alasan tiap keputusan; dokumentasi API (OpenAPI atau koleksi Postman); PRD di `docs/`.


---


## 7. Model Data Awal (ringkas)


- `auth`: `users`, `roles`, `permissions`, `role_permissions`, `user_roles`, `user_branches`, `audit_log` — semua PK/FK bertipe UUID
- `inventory`: `barang`, `satuan`, `cabang`, `stok` (barang_id, cabang_id, qty), `kartu_stok`, `transfer_stok`, `transfer_stok_item`, `opname`, `opname_item` — semua PK/FK bertipe UUID


Detail kolom dirancang saat mengerjakan modul masing-masing.


---


## 8. Tahapan Pengerjaan


Kerjakan berurutan dan jangan lompat ke tahap berikutnya sebelum kriteria selesai tahap ini terpenuhi. Kecepatan datang dari konsistensi mengerjakan satu tahap sampai tuntas, bukan dari target waktu.


| Tahap | Fokus | Selesai jika |
|---|---|---|
| 1 | Dasar Go: struct, interface, error handling, goroutine | CLI kecil (mis. parser CSV stok) berjalan dan bisa dijelaskan |
| 2 | Auth dan user management: API, migrasi, JWT, RBAC granular, log audit, user_branches, halaman login dan user | Login dan pengaturan peran berjalan end-to-end |
| 3 | Inventory: barang, stok, kartu stok, pengurangan aman, transfer, opname, halaman terkait | Modul inventory lengkap dengan frontend |
| 4 | Uji concurrency dan rollback, Docker, CI, deploy | Aplikasi publik yang bisa didemokan |


Tahap berikutnya adalah Purchasing, sales, dan accounting dikerjakan setelahnya sebagai lanjutan.


---


## 9. Cara Kerja dengan AI


- **Konsep baru:** usaha manual dulu (baca, coba, tanya kalau macet).
- **Implementasi proyek:** AI dipakai penuh sebagai pair-programmer, termasuk untuk sebagian besar kode frontend.
- **Aturan salin-lalu-ubah:** kode yang disalin dari contoh harus diubah minimal satu hal dan penulisnya harus bisa menjelaskan tiap barisnya.
- **Baca kode idiomatik:** 20-30 menit sehari.
- **Ritme kerja:** bertahap, satu file/langkah per giliran; jangan lanjut sebelum langkah saat ini dikonfirmasi jalan.


---


## 10. Kriteria Selesai (fase 1)


- [ ] Login, manajemen user, dan peran/izin berjalan dari frontend sampai database
- [ ] Endpoint terlindungi berdasarkan izin granular, dan aksi penting tercatat di log audit
- [ ] Relasi user multi-cabang berjalan
- [ ] Stok tidak pernah minus di bawah uji concurrency
- [ ] Transfer stok terbukti atomik lewat test rollback
- [ ] `docker-compose up` menjalankan backend, frontend, dan database
- [ ] CI menjalankan test otomatis
- [ ] Sudah deploy dan bisa diakses publik
- [ ] README menjelaskan arsitektur dan alasan memilih modular monolith
- [ ] Pemilik proyek bisa menjelaskan setiap modul tanpa membuka kode


---


## 11. Risiko dan Pertanyaan Terbuka


**Risiko**
- Frontend menghabiskan waktu. Mitigasi: gunakan shadcn/ui, jaga halaman tetap sedikit, andalkan AI untuk kode frontend.
- Menyalin tanpa memahami. Mitigasi: aturan salin-lalu-ubah dan latihan tanpa AI.
- Pengalaman produksi. Mitigasi: tahap pengerjaaan sampai dengan deploy.
- Sistem izin terlalu rumit di awal. Mitigasi: mulai dengan izin seed saat ini dan tambah seiring kebutuhan.


**Terjawab di v0.4**
1. ~~golang-migrate atau goose?~~ → goose, di `backend/migrations`.
2. ~~Apakah user terikat ke satu cabang atau bisa banyak cabang?~~ → many-to-many via `auth.user_branches`.
3. ~~Apakah model izin granular sudah sesuai, atau cukup peran tetap?~~ → granular per aksi.
4. ~~Nama proyek dan repositori.~~ → pos-app.


---


## Riwayat Perubahan


**v0.4 (29 September 2026)**
- Diputuskan: goose untuk migrasi (`backend/migrations`), UUID untuk semua PK/FK, bcrypt untuk password hash.
- Diputuskan: relasi user ↔ cabang many-to-many, izin granular per aksi, nama proyek pos-app.
- Struktur folder diperbarui sesuai kondisi aktual (migrasi + docs di dalam `backend/`).
- Pertanyaan terbuka v0.3 semuanya terjawab.


**v0.3**
- Semua perkiraan waktu dihapus; roadmap diganti tahapan berurutan dengan kriteria selesai.


**v0.2**
- Semua saran teknis (PostgreSQL, schema per modul, Gin, pgx + sqlc, frontend React) berubah dari "Disarankan" menjadi "Diputuskan".
- Scope fase 1 diperkecil menjadi auth/user management dan inventory, dengan schema terpisah dan frontend.
- Purchasing, sales, dan accounting dipindah ke fase berikutnya.
- Auth dikerjakan lebih dulu karena akan dipakai approval; contoh alur diganti dari penjualan menjadi transfer stok.
