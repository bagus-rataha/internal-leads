# ARCHITECTURE.md — Lead Sales App (Backend)

## 1. Tujuan Aplikasi

Aplikasi internal tim sales ISP korporat. Menggantikan Google Form + rekap Excel.

Alur inti: sales input lead (calon pelanggan korporat) → mencatat follow-up berulang → lead menembus pipeline Survey → SC → Registrasi → Instalasi → Trial → Invoice Bulanan, atau ditandai lost.

Skala: puluhan user internal, ribuan lead. Tidak ada beban konkuren berarti.

## 2. Stack

| | |
|---|---|
| Backend | Go + Fiber (dari boilerplate `da-backend`) |
| ORM | GORM |
| DB | PostgreSQL |
| Migration | `golang-migrate` |
| Frontend | React + Vite + TypeScript (repo sama, folder `frontend/`) |
| Serving | Frontend di-embed ke binary Go (`//go:embed`) |
| Deploy | Docker multi-stage → satu image, via Portainer Git stack |

## 3. Struktur Repo

Backend memakai **layered architecture**, mengikuti boilerplate `da-backend` yang sudah ada.

**IKUTI STRUKTUR FOLDERING YANG SUDAH ADA DI BOILERPLATE.** Jangan restrukturisasi, jangan pindah ke modular/domain-based, jangan perkenalkan pola direktori baru. Tambahkan file baru mengikuti konvensi penamaan yang sudah berjalan di sana.

Bentuknya kurang lebih:

```
lead-sales/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go
│   │   └── seed/main.go          # seeder wilayah, dijalankan terpisah
│   ├── internal/
│   │   ├── config/
│   │   ├── container/            # DI, sudah ada
│   │   ├── database/
│   │   ├── migrations/           # golang-migrate, .up.sql / .down.sql
│   │   ├── middleware/           # jwt.go (ada), rbac.go (BARU)
│   │   ├── models/               # user.go (ada), team, lead, follow_up, wilayah (BARU)
│   │   ├── dto/
│   │   ├── handlers/
│   │   ├── services/
│   │   ├── repositories/
│   │   ├── routes/
│   │   ├── static/
│   │   │   ├── embed.go          # //go:embed all:dist
│   │   │   └── dist/             # hasil build frontend (gitignore)
│   │   └── utils/
│   ├── go.mod
│   └── go.sum
├── frontend/
├── data/                          # CSV export Odoo untuk seeder
├── Dockerfile
├── docker-compose.yml
└── ARCHITECTURE.md
```

Domain yang perlu ditambahkan (masing-masing tersebar di lapisan `handlers` / `services` / `repositories` / `dto` / `routes` sesuai pola boilerplate): `team`, `lead`, `followup`, `reference` (lead_source, service_type, wilayah), `dashboard`, `export`.

**Pengecualian:** `dashboard` dan `export` **tidak punya repository**. Keduanya hanya membaca, memakai query agregat langsung di service. Jangan buatkan `dashboard_repository.go` demi konsistensi — itu abstraksi kosong.

## 4. Lapisan & Alur Request

```
Route
  → middleware.JWTProtected      (access token dari Authorization: Bearer)
  → middleware.RequireRole       (baca c.Locals("role") dari claim)
  → Handler                      (parse + ParseAndValidate DTO)
  → Service                      (logika bisnis + DATA SCOPING)
  → Repository                   (GORM)
```

**Data scoping ada di Service.** Bukan di handler, bukan di frontend. Ini titik keamanan utama aplikasi.

## 5. Auth

Mengikuti boilerplate, tanpa perubahan mekanisme:

- **Access token**: `Authorization: Bearer <token>`, umur pendek (maks 15 menit)
- **Refresh token**: httpOnly cookie, `SameSite: Strict`, disimpan di DB
- Same-origin di dev (Vite proxy) maupun produksi (embed) — tidak ada CORS

**`POST /auth/register` DIHAPUS.** Tidak ada registrasi mandiri. Akun dibuat Admin lewat Settings → User.

### Claim JWT

Claim memuat `user_id`, `email`, `role`. **`team_id` TIDAK dimasukkan ke claim.**

Alasannya: `team_id` mutable. Data mutable di token yang tidak bisa dicabut menghasilkan bug staleness.

**Aturan pengambilan identitas:**

| Data | Sumber | Query DB? |
|---|---|---|
| `user_id` | claim | tidak |
| `role` | claim | tidak |
| `team_id` | DB | **ya, tapi HANYA saat role == LEADER** |

Sales dan Admin/SU tidak pernah menyentuh DB untuk keperluan auth. Query `team_id` hanya terjadi untuk Leader, dan hanya di endpoint yang menyentuh lead.

### Revoke refresh token (WAJIB)

Karena `role` diambil dari claim, perubahan role tidak berlaku sampai access token kedaluwarsa. Untuk menutup jendela ini: **hapus seluruh refresh token milik user** saat terjadi salah satu dari:

1. User dinonaktifkan (`is_active → false`)
2. **Role diubah**
3. **`team_id` diubah**
4. Password direset oleh Admin

Revoke dilakukan **dalam transaksi yang sama** dengan perubahan user. Kalau update berhasil tapi revoke gagal, user berjalan dengan role lama — persis kegagalan yang ingin dicegah.

Efeknya: access token yang beredar tetap valid maksimal 15 menit tapi tidak bisa diperpanjang. Setelah itu user terpaksa login ulang dan mendapat claim baru.

Boilerplate sudah punya `logout-all` — logikanya persis ini. Panggil ulang dari service user, jangan tulis mekanisme baru.

### Middleware baru: `RequireRole`

Belum ada di boilerplate. Membaca `c.Locals("role")` yang sudah diisi `JWTProtected`. Harus dipasang **setelah** `JWTProtected`.

## 6. Role & Scoping

| Role | Cakupan lead |
|---|---|
| `SALES` | Hanya lead miliknya (`owner_id = user_id`) |
| `LEADER` | Seluruh lead di timnya (`owner.team_id = team_id`). Boleh input lead & follow-up atas nama sales-nya |
| `ADMIN_SALES` | Semua lead, semua tim. Bisa reassign owner |
| `SU` | Semua, plus konfigurasi |

**Satu fungsi scoping, dipakai semua query yang menyentuh lead** — list, detail, dashboard, export, count. Tidak ada pengecualian, termasuk endpoint yang "cuma menghitung".

Filter dari query param (`owner_id`, `team_id`) diterapkan **setelah** scoping, sebagai penyempitan tambahan. Tidak pernah menggantikannya. Sales yang mengirim `?owner_id=<orang-lain>` tetap hanya melihat lead miliknya.

## 7. Data Model

Pakai GORM model + tag. Struktur di bawah menyatakan maksud, bukan DDL.

### `sales_team`
`id` (uuid), `name`, `is_active` (default true), timestamps.

Tidak bisa dihapus, hanya dinonaktifkan. Tim yang masih punya anggota tidak boleh dinonaktifkan.

### `user`
`id` (uuid), `name`, `email` (unique), `password_hash`, `role`, `team_id` (nullable, FK sales_team), `is_active`, timestamps.

- `role`: `SALES` | `LEADER` | `ADMIN_SALES` | `SU`
- `team_id` **wajib** untuk SALES dan LEADER
- `team_id` **harus null** untuk ADMIN_SALES dan SU
- Satu sales hanya berada di satu tim. Tidak ada multi-tim
- Index pada `team_id` (untuk scoping LEADER)

### `lead_source` & `service_type`
Struktur identik: `id` (uuid), `name` (unique), `is_active`.

Seed `service_type`: Dedicated, Broadband.
Seed `lead_source`: Database Eksisting, Google, AI, Sistem Internal, Referensi Pelanggan, Komunitas Business, Event/Pameran, Media Sosial, LinkedIn Sales Navigator, Upselling/Cross Selling.

Yang sudah dipakai lead tidak bisa dihapus — hanya dinonaktifkan (hilang dari dropdown lead baru, tapi lead lama tetap menampilkannya).

### `lead` — tabel inti

**Identitas & status**
- `id` (uuid), `code` (unique) — format `LD-YYMM-NNNN`, mis. `LD-2607-0042`
- `status`: `BARU` | `FOLLOW_UP` | `SURVEY` | `SALES_CONFIRMATION` | `REGISTRASI` | `INSTALASI` | `TRIAL` | `INVOICE_BULANAN` | `LOST`, plus `HANDOFF_ODOO` (legacy — masih valid di CHECK, tidak dipakai untuk lead baru). Simpan sebagai text + CHECK constraint, **bukan enum PostgreSQL** (enum PG menyulitkan penambahan nilai)
- `lost_reason` (nullable, wajib saat status LOST)

**Kepemilikan — dua field terpisah, jangan digabung**
- `owner_id` — sales pemilik lead. Dipakai untuk scoping dan seluruh metrik
- `created_by_id` — siapa yang benar-benar menginput

Keduanya berbeda saat Leader input lead untuk sales-nya. Tanpa pemisahan ini, metrik keaktifan sales akan bohong sejak hari pertama.

**Perusahaan**: `company_name`, `business_field`, `website`

**Alamat**: `province_id`, `city_id`, `district_id`, `village_id`, `zip_id` (nullable), `rt`, `rw`, `street`

**PIC**: `pic_name`, `pic_position`, `office_phone`, `mobile_phone`, `email`

**Layanan eksisting** — semua nullable, sering belum diketahui saat lead masuk:
`service_type_id`, `capacity_mbps`, `existing_isp`, `price`, `other_services`

**Sumber**: `lead_source_id`

**Denormalisasi**: `last_follow_up_at` (nullable), `follow_up_count` (integer, default 0) —
counter jumlah follow-up, di-increment dalam transaksi yang sama dengan insert follow-up (lihat
§10). Sama alasannya dengan `last_follow_up_at`: hindari `COUNT(follow_up.*)` di runtime pada
query list/detail yang sering diakses.

**Index**: `owner_id`, `status`, `last_follow_up_at`, `created_at`, `city_id`

**Tidak ada `team_id` di lead.** Tim diturunkan dari `owner.team_id`. Konsekuensi yang diterima: sales pindah tim → lead lamanya ikut pindah cakupan tim. Snapshot hanya ditambahkan kalau nanti terbukti mengganggu. Jangan tambahkan sekarang.

### `follow_up` — append-only
`id` (uuid), `lead_id`, `note`, `created_by_id`, `created_at`.

**Tanpa `updated_at`.** Tidak ada endpoint update maupun delete. Riwayat follow-up tidak bisa diubah.

Index: `(lead_id, created_at)`.

### Wilayah — read-only, seed dari export Odoo

Integer PK, **bukan uuid**. Pertahankan ID asli Odoo.

| Tabel | Field |
|---|---|
| `province` | `id`, `external_id` (unique), `code`, `name` |
| `city` | `id`, `external_id`, `province_id`, `name` |
| `district` | `id`, `external_id`, `city_id`, `name` |
| `zip` | `id`, `external_id`, `code`, `city_id`, `district_id` |
| `village` | `id`, `external_id`, `district_id`, `zip_id` (nullable), `name` |

**Kenapa ID dan external_id Odoo dipertahankan:** aplikasi ini akan export/import lead ke Odoo. Odoo tidak menerima string `"13110"` — ia menerima referensi record `res.city.zip` lewat external_id `rv_location_custom.kode_pos_13110`. Mempertahankan bentuk sumber membuat export jadi pemetaan langsung, bukan penerjemahan.

**Kenapa `zip` tabel tersendiri, bukan kolom di village:** alasan yang sama. Di Odoo, zip adalah entitas (`res.city.zip`), bukan atribut.

Index: `city(province_id)`, `district(city_id)`, `village(district_id)`, `zip(district_id)`, dan index pencarian nama pada `village(name)`.

## 8. Migration & Seed

**Migration**: `golang-migrate`, file `.up.sql` / `.down.sql`. Bukan `AutoMigrate`.

**Sequence kode lead** harus ditulis sebagai raw SQL di migration — GORM tidak punya abstraksi untuk ini:
```sql
CREATE SEQUENCE lead_code_seq;
```

**Seeder wilayah**: perintah terpisah `go run ./cmd/seed`, membaca CSV export Odoo dari `data/`. **Bukan** dijalankan saat aplikasi start — 83.000 baris tidak boleh ada di jalur startup, dan tidak boleh berjalan ulang tanpa sengaja di produksi.

Urutan seed wajib (karena FK): `province` → `city` → `district` → `zip` → `village`.

Catatan format CSV Odoo:
- File `res.city` berisi kota banyak negara — **filter hanya `Country/ID = 100`** (Indonesia)
- Kecamatan & kelurahan di export menyimpan `State/ID` dan `City/ID` yang redundan — **abaikan**, cukup simpan parent langsung
- Ada kelurahan tanpa zip — karena itu `village.zip_id` nullable

## 9. API Contract

Base path mengikuti konvensi boilerplate. Response format mengikuti boilerplate (`Data` / `Errors` terpisah).

### Auth
| Method | Path | Akses |
|---|---|---|
| POST | `/auth/login` | publik |
| POST | `/auth/refresh` | cookie refresh |
| POST | `/auth/logout` | authenticated |
| POST | `/auth/logout-all` | authenticated |

`GET /users/me` (sudah ada di boilerplate) — response perlu ditambah `role` dan `team_id`.

### Lead
| Method | Path | Akses |
|---|---|---|
| GET | `/leads` | semua role, di-scope |
| POST | `/leads` | semua role |
| GET | `/leads/:code` | di-scope |
| PATCH | `/leads/:code` | pemilik, leader se-tim, admin |
| PATCH | `/leads/:code/status` | idem |
| GET | `/leads/export` | di-scope, mengembalikan xlsx |

`GET /leads` query: `q`, `status`, `source_id`, `team_id`, `owner_id`, `province_id`, `city_id`, `date_from`, `date_to`, `follow_up_from`, `follow_up_to`, `stale`, `sort`, `page`, `limit`.

- `q` — `ILIKE` (OR) terhadap `code`, `company_name`, `pic_name`. Satu kotak pencarian mencakup baik kode yang setengah diingat maupun nama perusahaan/PIC. `GET /leads/:code` tetap terpisah untuk direct-lookup exact-match (mis. navigasi setelah klik dari list), bukan pengganti pencarian.
- `date_from`/`date_to` — filter `created_at`.
- `follow_up_from`/`follow_up_to` — filter `last_follow_up_at`.
- `sort` — whitelist `code`, `company_name` (`?sort=field` asc, `?sort=-field` desc). `created_at` sengaja tidak disertakan terpisah: `code` digenerate dari sequence global di transaksi yang sama dengan insert, jadi urutan text `code` identik dengan urutan `created_at` — menyertakan keduanya redundan. Sort by nama kota (butuh JOIN ke `cities`) belum diimplementasikan — catatan untuk nanti kalau dibutuhkan.
- `page`/`limit` — default `page=1`, `limit=20`, `limit` di-cap 100 (nilai berlebih di-clamp, bukan error).

Response list: `{ items: [...], total, page, limit }` — bukan array polos, supaya frontend bisa render jumlah halaman tanpa request count terpisah.

Default sort: `last_follow_up_at ASC NULLS FIRST` — lead yang belum pernah di-follow-up (`last_follow_up_at` masih kosong) dianggap paling mendesak, konsisten dengan definisi lead terlantar di §10 yang fallback ke `created_at` untuk kasus yang sama.

**`POST /leads`:**
- `owner_id` di body **diabaikan** kalau pemanggil SALES — di-set ke dirinya sendiri
- `owner_id` untuk LEADER **wajib divalidasi** bahwa target adalah anggota timnya
- `created_by_id` **selalu** dari token, tidak pernah dari body

**`PATCH /leads/:code/status`** — transisi yang diizinkan:
```
BARU               → LOST                         (→ FOLLOW_UP otomatis: follow-up pertama)
FOLLOW_UP          → SURVEY | LOST
SURVEY             → SALES_CONFIRMATION | LOST
SALES_CONFIRMATION → REGISTRASI | LOST
REGISTRASI         → INSTALASI | LOST
INSTALASI          → TRIAL | LOST
TRIAL              → INVOICE_BULANAN | LOST
INVOICE_BULANAN    → (hanya admin/SU yang bisa menariknya mundur)
LOST               → (tidak ada transisi keluar)
```
Maju: tepat satu tahap, semua role. Mundur: ke tahap lebih awal mana pun, ADMIN_SALES / SU saja. `survey_at` di-stempel sekali saat pertama masuk SURVEY. `HANDOFF_ODOO` legacy diperlakukan seindeks SURVEY.

`lost_reason` wajib saat LOST. Transisi divalidasi di service — bukan sekadar tombolnya disembunyikan di UI.

`BARU → FOLLOW_UP` **tidak** lewat endpoint ini; terjadi otomatis saat follow-up pertama dibuat.

### Follow-up
| Method | Path | Akses |
|---|---|---|
| GET | `/leads/:code/followups` | mewarisi scope lead induk |
| POST | `/leads/:code/followups` | mewarisi scope lead induk |

Akses follow-up **tidak punya kolom sendiri** — mengikuti scope lead induknya. Kalau lead di luar cakupan caller, 404 (bukan 403), supaya keberadaan lead tidak bocor.

**Tidak ada PATCH dan DELETE.** Body POST hanya `{ note }` — `created_by_id` dari token, `created_at` dari server.

### Reference (read-only, authenticated)
| Method | Path |
|---|---|
| GET | `/refs/lead-sources` |
| GET | `/refs/service-types` |
| GET | `/refs/provinces` |
| GET | `/refs/cities?province_id=` |
| GET | `/refs/districts?city_id=` |
| GET | `/refs/villages?district_id=&q=` |

**`/refs/villages` wajib membatasi hasil**: filter `district_id`, `q` opsional, limit 50. Tanpa batas, dropdown akan berat.

Response village menyertakan objek zip (id + code), agar form mengisi kode pos tanpa request tambahan.

### Dashboard
| Method | Path | Akses |
|---|---|---|
| GET | `/dashboard/summary` | semua, di-scope |
| GET | `/dashboard/activity` | semua, di-scope |
| GET | `/dashboard/stale-leads` | semua, di-scope |
| GET | `/dashboard/sales-activity` | **LEADER, ADMIN_SALES, SU saja → 403 untuk SALES** |
| GET | `/dashboard/segments` | semua, di-scope |

Query: `date_from`, `date_to`, `team_id`, `owner_id`. Default rentang: 30 hari terakhir; pembanding: 30 hari sebelumnya.

### User & Team (ADMIN_SALES, SU saja)
| Method | Path |
|---|---|
| GET/POST | `/users` |
| GET/PATCH | `/users/:id` |
| POST | `/users/:id/reset-password` |
| POST | `/users/:id/deactivate` |
| GET/POST | `/teams` |
| GET/PATCH | `/teams/:id` |

**`POST /users/:id/deactivate`**, body `{ reassign_to_user_id? }`:
- Kalau user punya lead aktif dan `reassign_to_user_id` kosong → **422**, payload berisi jumlah lead aktif. Frontend memakai ini untuk menampilkan dialog reassign
- Reassign lead + nonaktifkan user + revoke refresh token = **satu transaksi**

### Master reference (ADMIN_SALES, SU)
`GET/POST/PATCH` untuk `/lead-sources` dan `/service-types`.

**Tidak ada DELETE di seluruh API.** Semua nonaktif lewat `is_active`.

## 10. Aturan Bisnis

### Kode lead
Format `LD-YYMM-NNNN` (`LD-2607-0042`), digenerate server dalam **transaksi yang sama** dengan INSERT lead, memakai PostgreSQL sequence.

Sequence **tidak reset per bulan** — reset butuh locking dan menghasilkan race condition. Nomor urut global, prefix bulan hanya penanda.

**Dilarang** memakai `COUNT(*)+1`. Dua sales menyimpan bersamaan akan menghasilkan kode kembar. `code` punya UNIQUE constraint sebagai jaring pengaman terakhir.

### Transaksi follow-up
Menulis satu follow-up menyentuh empat hal. **Satu transaksi, tanpa kecuali:**

1. Insert baris `follow_up`
2. Update `lead.last_follow_up_at = now()`
3. Update `lead.follow_up_count = follow_up_count + 1` (increment atomic di level SQL, bukan read-modify-write di Go)
4. Update `lead.status = FOLLOW_UP` **hanya jika status saat ini `BARU`**

Kondisi di langkah 4 mencegah lead yang sudah lepas dari tahap `BARU` tertarik mundur.

`last_follow_up_at` adalah denormalisasi yang disengaja — dipakai untuk sorting default dan deteksi lead terlantar di query terpanas aplikasi. **Jangan** diganti dengan `MAX(follow_up.created_at)` di runtime.

### Lead terlantar
Satu definisi, dipakai di list dan dashboard. Simpan ambang sebagai konstanta, jangan sebar literal di banyak query.

Lead dianggap terlantar bila:
- Statusnya masih `BARU` atau `FOLLOW_UP` (sejak `SURVEY`, lead ada di pipeline Odoo — tidak dihitung terlantar; `LOST` sudah selesai), **dan**
- Sudah lewat **7 hari** sejak follow-up terakhir — atau, bila belum pernah di-follow-up, sejak lead dibuat

### Dashboard
Satu endpoint = satu query agregat. **Dilarang N+1** — jangan ambil daftar sales lalu query metrik per sales dalam loop.

**Kontrak zero-state — dihitung di backend, bukan frontend:**

- Kalau periode pembanding bernilai 0, kirim `change_pct: null`. **Bukan `100`.** `100%` dari baseline nol adalah pembagian dengan nol yang disamarkan
- Conversion rate juga `null` saat penyebutnya nol, bukan `0`

Frontend menampilkan "baru" atau "—" saat null. Frontend tidak menghitung persen sendiri.

Aplikasi ini akan berjalan berbulan-bulan dengan data tipis. Dashboard harus tetap benar saat angkanya kecil atau nol.

### Export Excel
Library `excelize`. Sinkron, di-stream sebagai response.

Dua sheet:
- **`Leads`** — satu baris per lead. Kolom lead + **nama** wilayah (bukan ID) + nama sales + nama tim + jumlah follow-up + tanggal follow-up terakhir
- **`Follow-up`** — satu baris per follow-up (kode lead, tanggal, penulis, catatan), bernormalisasi untuk analisis

**Export MELEWATI fungsi scoping yang sama.** Sales yang menekan Export hanya mendapat lead miliknya. Ini titik kebocoran data paling mudah terlewat — export sering ditulis sebagai query terpisah lalu lupa di-scope.

Kalau hasil melebihi 50.000 baris, tolak dengan 422 dan minta user mempersempit filter.

## 11. Larangan

- **Tidak ada DELETE** di seluruh API. Semua nonaktif lewat `is_active`
- **Follow-up tidak bisa diedit maupun dihapus**
- **Tidak ada** Redis, message queue, background worker, event bus, soft-delete generik, audit log, atau multi-tenancy
- **YAGNI.** Jangan usulkan pola atau abstraksi baru kalau yang sudah ada di boilerplate sudah menyelesaikan masalah