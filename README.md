# PRISM

PRISM (Project Loan Integrated Monitoring System) adalah aplikasi untuk mencatat dan memonitor alur pinjaman luar negeri Bappenas dari Blue Book, Green Book, Daftar Kegiatan, Loan Agreement, sampai Monitoring Disbursement.

## Stack

| Layer | Teknologi |
| --- | --- |
| Backend | Go, Echo, sqlc, pgx/v5, golang-migrate |
| Frontend | Vue 3, Vite, Pinia, PrimeVue 4, Tailwind CSS v4 |
| Database | PostgreSQL 16 |
| Realtime | Server-Sent Events |

## Struktur Repo

```text
prism/
+-- docs/             # Kontrak, business rules, DDL, dan rencana versioning
+-- plans/            # Checklist implementasi per fase
+-- prism-backend/    # Backend Go
+-- prism-frontend/   # Frontend Vue
```

## Dokumen Utama

Baca [AGENTS.md](AGENTS.md) sebelum mengubah kode. Dokumen yang paling sering menjadi sumber aturan:

- [docs/PRISM_Business_Rules.md](docs/PRISM_Business_Rules.md)
- [docs/PRISM_API_Contract.md](docs/PRISM_API_Contract.md)
- [docs/prism_ddl.sql](docs/prism_ddl.sql)
- [docs/PRISM_BB_GB_Revision_Versioning_Plan.md](docs/PRISM_BB_GB_Revision_Versioning_Plan.md)

Checklist implementasi ada di [plans/](plans/).

## Setup Development

Salin env contoh:

```powershell
Copy-Item .env.example .env
```

Jalankan stack development:

```powershell
docker compose -f docker-compose.dev.yml up -d --build
```

Service development:

- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Health check: http://localhost:8080/health
- PostgreSQL: `localhost:5432`

Admin seed awal:

- Username: `admin`
- Password: `admin123`

## Fresh DB

Untuk reset database development:

```powershell
docker compose -f docker-compose.dev.yml down -v --remove-orphans
docker compose -f docker-compose.dev.yml up -d --build
```

Container PostgreSQL development membuat schema awal dari [docs/prism_ddl.sql](docs/prism_ddl.sql). File migration di [prism-backend/migrations/](prism-backend/migrations/) tetap disimpan untuk database yang sudah hidup dari schema lama dan untuk histori perubahan, tetapi tidak perlu dijalankan manual setelah fresh DB dari DDL terbaru.

## Init Migration State

Setelah fresh DB dari `docs/prism_ddl.sql`, gunakan skrip ini untuk mengisi seed admin:

```powershell
# Jalankan dari root repo setelah container PostgreSQL dev hidup.
$env:DATABASE_URL = "postgres://prism:prism_secret@localhost:5432/prism_dev?sslmode=disable"

Set-Location .\prism-backend
migrate -path migrations -database $env:DATABASE_URL up
Set-Location ..
```

## Backend

Masuk ke folder backend:

```powershell
Set-Location prism-backend
```

Command umum:

```powershell
go test ./...
go vet ./...
sqlc generate
```

Jika mengubah SQL source di `prism-backend/sql/queries/*.sql`, jalankan generate sebelum menulis kode Go yang bergantung pada query tersebut:

```powershell
make generate
```

Jangan edit file generated di `prism-backend/internal/database/queries/`.

## Frontend

Masuk ke folder frontend:

```powershell
Set-Location prism-frontend
```

Command umum di Windows:

```powershell
npm.cmd install
npm.cmd run dev
npm.cmd run build
npm.cmd run lint
```

Gunakan `npm.cmd`, bukan `npm`, untuk menghindari masalah eksekusi PowerShell di Windows.

## Revision Versioning

Fase BB/GB Revision Versioning mengikuti urutan:

```text
BE-07 -> BE-08 -> BE-09 -> BE-10 -> BE-11 -> FE-10
```

Plan terkait:

- [plans/PLAN_BE_07_Revision_Versioning_Schema.md](plans/PLAN_BE_07_Revision_Versioning_Schema.md)
- [plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md](plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md)
- [plans/PLAN_BE_09_Green_Book_Revision_Versioning.md](plans/PLAN_BE_09_Green_Book_Revision_Versioning.md)
- [plans/PLAN_BE_10_DK_LA_Frozen_Snapshot.md](plans/PLAN_BE_10_DK_LA_Frozen_Snapshot.md)
- [plans/PLAN_BE_11_Journey_Import_Project_List_Versioning.md](plans/PLAN_BE_11_Journey_Import_Project_List_Versioning.md)
- [plans/PLAN_10_BB_GB_Revision_UI.md](plans/PLAN_10_BB_GB_Revision_UI.md)

Aturan ringkas revision versioning:

- BB/GB Project adalah snapshot per dokumen/revisi.
- Logical identity menghubungkan snapshot lintas revisi.
- GB baru memakai versi BB terbaru saat dibuat.
- DK baru memakai versi GB terbaru saat dibuat.
- DK/LA yang sudah dibuat tetap menunjuk snapshot konkret dan tidak auto-pindah saat ada revisi baru.

## Verifikasi Cepat

Backend:

```powershell
Set-Location prism-backend
go test ./...
```

Frontend:

```powershell
Set-Location prism-frontend
npm.cmd run build
npm.cmd run lint
```
