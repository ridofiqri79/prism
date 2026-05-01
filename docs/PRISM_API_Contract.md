# PRISM — API Contract & Endpoint Design

> Base URL: `/api/v1`
> Auth: Bearer JWT di header `Authorization`
> Format: JSON (`Content-Type: application/json`)

---

## Konvensi Umum

### Request

Semua request yang membutuhkan body menggunakan JSON. Field `id` selalu UUID v4.

### Response Sukses

```json
{
  "data": { ... },
  "meta": { ... }   // hanya untuk list — berisi pagination
}
```

### Response Error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "bb_code sudah digunakan",
    "details": [
      { "field": "bb_code", "message": "sudah digunakan" }
    ]
  }
}
```

### Error Codes

| HTTP Status | Code | Keterangan |
|-------------|------|-----------|
| 400 | `VALIDATION_ERROR` | Input tidak valid |
| 401 | `UNAUTHORIZED` | Token tidak ada / expired |
| 403 | `FORBIDDEN` | Tidak punya permission |
| 404 | `NOT_FOUND` | Resource tidak ditemukan |
| 409 | `CONFLICT` | Duplikat data (misal: bb_code sudah ada) |
| 500 | `INTERNAL_ERROR` | Server error |

### Pagination

Query params untuk semua endpoint list:

| Param | Default | Keterangan |
|-------|---------|-----------|
| `page` | `1` | Halaman |
| `limit` | `20` | Jumlah item per halaman |
| `sort` | `created_at` | Field untuk sorting |
| `order` | `desc` | `asc` atau `desc` |
| `search` | kosong | Kata kunci pencarian; hanya tersedia pada endpoint list yang mencantumkannya di bagian Query Params tambahan |

Filter multi-value dapat dikirim sebagai query berulang, comma-separated, atau array suffix, misalnya `?type=Bilateral&type=KSA`, `?type=Bilateral,KSA`, atau `?type[]=Bilateral&type[]=KSA`.

Response meta:

```json
{
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Auth

### `POST /auth/login`

**Permission:** Public

**Request:**
```json
{
  "username": "staff01",
  "password": "secret"
}
```

**Response `200`:**
```json
{
  "data": {
    "access_token": "eyJ...",
    "expires_in": 86400,
    "user": {
      "id": "uuid",
      "username": "staff01",
      "email": "staff01@bappenas.go.id",
      "role": "STAFF"
    }
  }
}
```

---

### `POST /auth/logout`

**Permission:** Authenticated

**Response `204`:** No content

---

### `GET /auth/me`

**Permission:** Authenticated

**Response `200`:**
```json
{
  "data": {
    "id": "uuid",
    "username": "staff01",
    "email": "staff01@bappenas.go.id",
    "role": "STAFF",
    "permissions": [
      { "module": "bb_project", "can_create": true, "can_read": true, "can_update": false, "can_delete": false }
    ]
  }
}
```

---

## Master Data

### Import Data

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/import-data/template` | ADMIN only |
| `POST` | `/master/import-data/preview` | ADMIN only |
| `POST` | `/master/import-data/execute` | ADMIN only |

**Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Program Titles`, `Bappenas Partners`, `Institutions`, `Regions`, `Periods`, `National Priorities`, dan `Lenders` |

**Template:**
`GET /master/import-data/template` mengunduh workbook `.xlsx` dengan sheet `Panduan` yang deskriptif, header sheet import, dropdown Excel untuk kolom yang punya pilihan master data, dan sheet `Master Data Snapshot` berisi data master yang ada di database saat template dibuat. Workbook juga memiliki sheet `_Dropdowns` tersembunyi sebagai sumber pilihan dropdown. Response memakai `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` dan `Content-Disposition: attachment`.

**Preview:**
`POST /master/import-data/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /master/import-data/execute` menyimpan data jika hasil pemrosesan tidak memiliki baris gagal. Endpoint lama `/master/import-data` tetap tersedia sebagai alias eksekusi.

**Response `200`:**
```json
{
  "data": {
    "file_name": "master_data_import_template_data_awal.xlsx",
    "total_inserted": 120,
    "total_skipped": 10,
    "total_failed": 0,
    "sheets": [
      {
        "sheet": "Program Titles",
        "inserted": 12,
        "skipped": 0,
        "failed": 0,
        "rows": [
          {
            "row": 2,
            "status": "create",
            "label": "Infrastruktur Transportasi"
          },
          {
            "row": 3,
            "status": "skip",
            "label": "Energi",
            "message": "Data sudah ada, dilewati"
          },
          {
            "row": 4,
            "status": "failed",
            "label": "Baris 4",
            "message": "Title wajib diisi"
          }
        ]
      }
    ]
  }
}
```

Baris yang sudah ada akan di-skip. Untuk sheet `Institutions`, duplikat dicek sesuai scope: top-level berdasarkan nama, child berdasarkan kombinasi parent dan nama. `Parent Name` dapat diisi dengan nama jika unik, UUID institution, atau path `Nama Child; Nama Parent; Nama Root;`. Jika `Parent Name` hanya berisi nama polos dan mengarah ke lebih dari satu institution karena nama child duplikat lintas parent, baris dianggap `failed` agar import tidak memilih parent yang salah. Sheet `Panduan` pada template menjelaskan fallback referensi Institution: path dropdown sebagai prioritas utama, UUID dari sheet Master Data sebagai fallback paling spesifik, dan nama polos hanya jika unik. Detail baris preview dikembalikan di `sheets[].rows` dengan `status`: `create`, `skip`, atau `failed`, sehingga frontend dapat memberi tab/filter sebelum eksekusi. Baris yang gagal validasi juga dikembalikan di `sheets[].errors`. Frontend wajib meminta preview terlebih dahulu sebelum user menekan eksekusi import.

### Country

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/countries` | read: `country` |
| `POST` | `/master/countries` | create: `country` |
| `PUT` | `/master/negara/:id` | update: `country` |
| `DELETE` | `/master/negara/:id` | delete: `country` |

**`GET /master/countries` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `search` | Cari berdasarkan `name` atau `code` |
| `sort` | `name`, `code` |

**`GET /master/countries` Response `200`:**
```json
{
  "data": [
    { "id": "uuid", "name": "Japan", "code": "JPN" }
  ],
  "meta": { "page": 1, "limit": 20, "total": 195, "total_pages": 10 }
}
```

**`POST /master/countries` Request:**
```json
{
  "name": "Japan",
  "code": "JPN"
}
```

---

### Currency

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/currencies` | read: `currency` |
| `GET` | `/master/currencies/:id` | read: `currency` |
| `POST` | `/master/currencies` | create: `currency` |
| `PUT` | `/master/currencies/:id` | update: `currency` |
| `DELETE` | `/master/currencies/:id` | delete: `currency` |

**`GET /master/currencies` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `active` | Filter `true` atau `false`; kosong mengembalikan semua currency |
| `search` | Cari berdasarkan `code` atau `name` |
| `sort` | `code`, `name`, `sort_order`, `is_active` |

**`GET /master/currencies` Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "code": "JPY",
      "name": "Japanese Yen",
      "symbol": "JPY",
      "is_active": true,
      "sort_order": 30
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 12, "total_pages": 1 }
}
```

**`POST /master/currencies` Request:**
```json
{
  "code": "JPY",
  "name": "Japanese Yen",
  "symbol": "JPY",
  "is_active": true,
  "sort_order": 30
}
```

Currency pada Green Book, DK, dan LA harus memakai kode ISO 4217 yang terdaftar aktif di Master Currency. Seed awal mengikuti mata uang negara donor/lender dan mata uang yang umum digunakan lembaga multilateral.

---

### Lender

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/lenders` | read: `lender` |
| `GET` | `/master/lenders/:id` | read: `lender` |
| `POST` | `/master/lenders` | create: `lender` |
| `PUT` | `/master/lenders/:id` | update: `lender` |
| `DELETE` | `/master/lenders/:id` | delete: `lender` |

**`POST /master/lenders` Request:**
```json
{
  "name": "JICA",
  "type": "Bilateral",
  "country_id": "uuid"        // wajib jika type Bilateral atau KSA
}
```

**`GET /master/lenders` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `type` | Filter multi-value: `Bilateral`, `Multilateral`, `KSA` |
| `search` | Cari berdasarkan `name` atau `short_name` |
| `sort` | `name`, `short_name`, `type`, `country` |

---

### Institution

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/institutions` | read: `institution` |
| `GET` | `/master/institutions/lookup` | read: `institution` |
| `GET` | `/master/institutions/:id` | read: `institution` |
| `POST` | `/master/institutions` | create: `institution` |
| `PUT` | `/master/institutions/:id` | update: `institution` |
| `DELETE` | `/master/institutions/:id` | delete: `institution` |

**`POST /master/institutions` Request:**
```json
{
  "name": "Kementerian PUPR",
  "level": "Kementerian/Badan/Lembaga",
  "parent_id": null           // null untuk Kementerian/Badan/Lembaga, uuid untuk child
}
```

**`GET /master/institutions` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `level` | Filter multi-value: `Kementerian/Badan/Lembaga`, `Eselon I`, `Eselon II`, `BUMN`, `Pemerintah Daerah Tk. I`, `Pemerintah Daerah Tk. II`, `BUMD`, `Lainya` |
| `parent_id` | Filter by parent |
| `search` | Cari berdasarkan `name` atau `short_name` |
| `sort` | `name`, `short_name`, `level` |

Response `/master/institutions` digunakan untuk TreeTable:
- Tanpa `parent_id`: paginasi dihitung dari root/top-level yang match diri sendiri atau descendant.
- Dengan `parent_id`: mengembalikan direct child dari parent tersebut untuk lazy expand.
- Item menyertakan `has_children` jika masih memiliki child.

Response `/master/institutions/lookup` adalah list flat untuk selector/dropdown parent. Endpoint ini tetap mendukung `level`, `parent_id`, `search`, `sort`, `page`, dan `limit`.

Validasi nama institution:
- `parent_id = null` (top-level) tidak boleh memiliki nama yang sama dengan top-level lain.
- Child tidak boleh memiliki nama yang sama dalam parent yang sama. Nama child yang sama boleh dipakai di parent berbeda.

---

### Region

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/regions` | read: `region` |
| `GET` | `/master/regions/lookup` | read: `region` |
| `GET` | `/master/wilayah/:id` | read: `region` |
| `POST` | `/master/regions` | create: `region` |
| `PUT` | `/master/wilayah/:id` | update: `region` |
| `DELETE` | `/master/wilayah/:id` | delete: `region` |

**`POST /master/regions` Request:**
```json
{
  "name": "Jawa Barat",
  "type": "PROVINCE",
  "parent_code": "ID"
}
```

**`GET /master/regions` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `type` | Filter multi-value: `COUNTRY`, `PROVINCE`, `CITY` |
| `parent_code` | Filter by parent code — untuk load CITY per PROVINCE |
| `search` | Cari berdasarkan `name` atau `code` |
| `sort` | `code`, `name`, `type` |

Response `/master/regions` digunakan untuk TreeTable:
- Tanpa `parent_code`: paginasi dihitung dari root `COUNTRY` yang match diri sendiri atau descendant.
- Dengan `parent_code`: mengembalikan direct child dari code parent tersebut untuk lazy expand.
- Item menyertakan `has_children` jika masih memiliki child.

Response `/master/regions/lookup` adalah list flat untuk selector/dropdown parent. Endpoint ini tetap mendukung `type`, `parent_code`, `search`, `sort`, `page`, dan `limit`.

---

### Program Title

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/program-titles` | read: `program_title` |
| `GET` | `/master/program-titles/lookup` | read: `program_title` |
| `POST` | `/master/program-titles` | create: `program_title` |
| `PUT` | `/master/program-titles/:id` | update: `program_title` |
| `DELETE` | `/master/program-titles/:id` | delete: `program_title` |

**`GET /master/program-titles` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `search` | Cari berdasarkan `title` |
| `parent_id` | Untuk lazy load child di TreeTable |
| `sort` | `title` |

Response `/master/program-titles` digunakan untuk TreeTable:
- Tanpa `parent_id`: paginasi dihitung dari root title yang match diri sendiri atau descendant.
- Dengan `parent_id`: mengembalikan direct child dari parent tersebut untuk lazy expand.
- Item menyertakan `has_children` jika masih memiliki child.

Response `/master/program-titles/lookup` adalah list flat untuk selector/dropdown parent. Endpoint ini tetap mendukung `search`, `sort`, `page`, dan `limit`.

**`POST /master/program-titles` Request:**
```json
{
  "title": "Infrastruktur Transportasi",
  "parent_id": null           // null untuk Parent, uuid untuk Child
}
```

---

### Bappenas Partner

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/bappenas-partners` | read: `bappenas_partner` |
| `GET` | `/master/bappenas-partners/lookup` | read: `bappenas_partner` |
| `POST` | `/master/bappenas-partners` | create: `bappenas_partner` |
| `PUT` | `/master/bappenas-partners/:id` | update: `bappenas_partner` |
| `DELETE` | `/master/bappenas-partners/:id` | delete: `bappenas_partner` |

**`GET /master/bappenas-partners` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `level` | Filter multi-value: `Eselon I`, `Eselon II` |
| `parent_id` | Untuk lazy load child di TreeTable |
| `search` | Cari berdasarkan `name` |
| `sort` | `name`, `level` |

Response `/master/bappenas-partners` digunakan untuk TreeTable:
- Tanpa `parent_id`: paginasi dihitung dari root `Eselon I` yang match diri sendiri atau descendant.
- Dengan `parent_id`: mengembalikan direct child dari parent tersebut untuk lazy expand.
- Item menyertakan `has_children` jika masih memiliki child.

Response `/master/bappenas-partners/lookup` adalah list flat untuk selector/dropdown parent. Endpoint ini tetap mendukung `level`, `search`, `sort`, `page`, dan `limit`.

**`POST /master/bappenas-partners` Request:**
```json
{
  "name": "Direktorat Transportasi",
  "level": "Eselon II",
  "parent_id": "uuid-eselon-i"
}
```

---

### Period

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/periods` | read: `period` |
| `POST` | `/master/periods` | create: `period` |
| `PUT` | `/master/periods/:id` | update: `period` |
| `DELETE` | `/master/periods/:id` | delete: `period` |

**`GET /master/periods` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `sort` | `name`, `year_start`, `year_end` |

**`POST /master/periods` Request:**
```json
{
  "name": "2025-2029",
  "year_start": 2025,
  "year_end": 2029
}
```

---

### National Priority

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/national-priorities` | read: `national_priority` |
| `POST` | `/master/national-priorities` | create: `national_priority` |
| `PUT` | `/master/national-priorities/:id` | update: `national_priority` |
| `DELETE` | `/master/national-priorities/:id` | delete: `national_priority` |

**`GET /master/national-priorities` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `period_id` | Filter multi-value berdasarkan periode |
| `search` | Cari berdasarkan `title` |
| `sort` | `title`, `period` |

**`POST /master/national-priorities` Request:**
```json
{
  "period_id": "uuid",
  "title": "Ketahanan Pangan"
}
```

---

## Blue Book

### Blue Book (Header)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/blue-books` | read: `blue_book` |
| `GET` | `/blue-books/:id` | read: `blue_book` |
| `POST` | `/blue-books` | create: `blue_book` |
| `PUT` | `/blue-books/:id` | update: `blue_book` |
| `DELETE` | `/blue-books/:id` | delete: `blue_book` |

**`GET /blue-books` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan nama periode, tanggal terbit, tahun revisi, atau status |
| `period_id` | multi-value UUID | Filter periode Blue Book |
| `status` | multi-value enum | `active`, `superseded` |

**`POST /blue-books` Request:**
```json
{
  "period_id": "uuid",
  "replaces_blue_book_id": "uuid-blue-book-sebelumnya",
  "publish_date": "2025-01-15",
  "revision_number": 0,
  "revision_year": null
}
```

**`GET /blue-books/:id` Response `200`:**
```json
{
  "data": {
    "id": "uuid",
    "period": { "id": "uuid", "name": "2025-2029", "year_start": 2025, "year_end": 2029 },
    "replaces_blue_book_id": null,
    "publish_date": "2025-01-15",
    "revision_number": 0,
    "revision_year": null,
    "status": "active",
    "created_at": "2025-01-15T08:00:00Z",
    "updated_at": "2025-01-15T08:00:00Z"
  }
}
```

---

### Import Blue Book Projects

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/blue-books/:bb_id/import-projects/template` | ADMIN only |
| `POST` | `/blue-books/:bb_id/import-projects/preview` | ADMIN only |
| `POST` | `/blue-books/:bb_id/import-projects/execute` | ADMIN only |

**Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Input Data`, `Relasi - EA`, `Relasi - IA`, `Relasi - Locations`, `Relasi - National Priority`, `Relasi - Project Cost`, dan `Relasi - Lender Indication` |

Workbook diimport ke Blue Book target dari `:bb_id`. Sheet relasi memakai `BB Code (*)` sebagai kunci penghubung ke sheet `Input Data`.

**Template:**
`GET /blue-books/:bb_id/import-projects/template` mengunduh workbook `.xlsx` dengan sheet `Panduan` yang deskriptif, `Master Data`, `Input Data`, dan semua sheet relasi. Kolom relasi memiliki dropdown Excel dari master data dan BB Code pada sheet relasi memiliki dropdown dari `Input Data`. Sheet `Master Data` berisi snapshot master data saat template dibuat; national priority menampilkan seluruh master data dan tidak dibatasi period Blue Book target. Workbook juga memiliki sheet `_Dropdowns` tersembunyi sebagai sumber pilihan dropdown.

**Kolom utama workbook:**

| Sheet | Kolom |
|-------|-------|
| `Input Data` | `Program Title (*)`, `Bappenas Partners`, `BB Code (*)`, `Project Name (*)`, `Duration`, `Objective`, `Scope of Work`, `Outputs`, `Outcomes` |
| `Relasi - EA` | `BB Code (*)`, `Executing Agency Name (*)` |
| `Relasi - IA` | `BB Code (*)`, `Implementing Agency Name (*)` |
| `Relasi - Locations` | `BB Code (*)`, `Location Name (*)` |
| `Relasi - National Priority` | `BB Code (*)`, `National Priority Name (*)` |
| `Relasi - Project Cost` | `BB Code (*)`, `Funding Type (*)`, `Funding Category (*)`, `Amount USD` |
| `Relasi - Lender Indication` | `BB Code (*)`, `Lender Name (*)`, `Keterangan` |

Kolom `Duration` pada workbook diisi sebagai angka jumlah bulan. Kolom `Bappenas Partners` opsional; isi lebih dari satu mitra dengan pemisah koma atau titik koma. Kolom institution pada `Relasi - EA` dan `Relasi - IA` dapat diisi dengan nama jika unik, UUID dari sheet `Master Data`, atau path `Nama Child; Nama Parent; Nama Root;`. Template dropdown memakai path agar nama child yang sama di parent berbeda tetap bisa dipilih tanpa ambigu. Kolom `Relasi - Lender Indication.Lender Name` di-resolve ke master Lender berdasarkan `name`; jika tidak cocok, import mencoba fallback ke `short_name` yang unik. Sheet `Panduan` menjelaskan fallback ini agar operator tidak perlu menebak ketika nama Institution sama atau workbook memakai singkatan lender seperti `ADB`, `IFAD`, `EIB`, atau `UKEF`.

**Preview:**
`POST /blue-books/:bb_id/import-projects/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /blue-books/:bb_id/import-projects/execute` menyimpan data jika hasil pemrosesan tidak memiliki baris gagal.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`. Frontend wajib menampilkan preview dan meminta konfirmasi user sebelum eksekusi.

Baris dengan `BB Code` yang sudah ada dalam Blue Book target akan di-skip. `BB Code` yang hanya ada pada revisi lama tidak di-skip; jika cocok dengan revisi sumber, snapshot baru memakai `project_identity_id` yang sama. Relasi valid akan dibuat bersama proyek baru; relasi untuk proyek yang di-skip ikut di-skip. National Priority divalidasi terhadap master data tanpa pembatasan period Blue Book target.

---

### BB Project

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/blue-books/:bb_id/projects` | read: `bb_project` |
| `GET` | `/blue-books/:bb_id/projects/:id` | read: `bb_project` |
| `POST` | `/blue-books/:bb_id/projects` | create: `bb_project` |
| `PUT` | `/blue-books/:bb_id/projects/:id` | update: `bb_project` |
| `DELETE` | `/blue-books/:bb_id/projects/:id` | delete: `bb_project` |

`DELETE /blue-books/:bb_id/projects/:id` melakukan hard delete. Backend menolak penghapusan jika BB Project masih menjadi referensi Green Book Project, Daftar Kegiatan, Loan Agreement, atau Monitoring. Untuk record yang sudah memiliki relasi turunan, STAFF menerima `403 FORBIDDEN` dan ADMIN menerima `409 CONFLICT`; keduanya berisi `error.details[]` dengan daftar relasi yang harus dibersihkan terlebih dahulu.

Contoh response ketika masih dipakai downstream:

```json
{
  "error": {
    "code": "CONFLICT",
    "message": "BB Project tidak bisa dihapus permanen karena masih memiliki relasi turunan. Hapus relasi turunan terlebih dahulu.",
    "details": [
      {
        "field": "green_book_project",
        "message": "GB-2025-001 - Trans Sumatra Section 1 | Green Book 2025 Revisi 0 | id=uuid-gb-project"
      },
      {
        "field": "monitoring_disbursement",
        "message": "2026 TW1 | Green Book Project GB-2025-001 -> Daftar Kegiatan DK-001 -> Loan Agreement IP-603 -> Monitoring 2026 TW1 | id=uuid-monitoring"
      }
    ]
  }
}
```

**`GET /blue-books/:bb_id/projects` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan `project_name` atau nama/nama singkat Executing Agency |
| `executing_agency_ids` | multi-value UUID | Filter institution role `Executing Agency` |
| `location_ids` | multi-value UUID | Filter region lokasi proyek |

**`POST /blue-books/:bb_id/projects` Request:**
```json
{
  "project_identity_id": "uuid-logical-project-opsional",
  "program_title_id": "uuid",
  "bappenas_partner_ids": ["uuid-mitra-bappenas-1", "uuid-mitra-bappenas-2"],
  "bb_code": "BB-2025-001",
  "project_name": "Pembangunan Jalan Tol Trans Sumatera",
  "duration": 60,
  "objective": "Meningkatkan konektivitas...",
  "scope_of_work": "Pembangunan 500km...",
  "outputs": "500km jalan tol terbangun",
  "outcomes": "Waktu tempuh berkurang 40%",
  "executing_agency_ids": ["uuid-kemen-pupr"],
  "implementing_agency_ids": ["uuid-eselon-i"],
  "location_ids": ["uuid-sumatra-utara", "uuid-riau"],
  "national_priority_ids": ["uuid-np-1", "uuid-np-2"],
  "project_costs": [
    { "funding_type": "Foreign", "funding_category": "Loan", "amount_usd": 500000000 },
    { "funding_type": "Counterpart", "funding_category": "Central Government", "amount_usd": 100000000 }
  ],
  "lender_indications": [
    { "lender_id": "uuid-jica", "remarks": "Minat untuk membiayai seksi 1-3" }
  ]
}
```

**`GET /blue-books/:bb_id/projects/:id` Response `200`:**
```json
{
  "data": {
    "id": "uuid",
    "project_identity_id": "uuid-logical-project",
    "blue_book_id": "uuid",
    "bb_code": "BB-2025-001",
    "project_name": "Pembangunan Jalan Tol Trans Sumatera",
    "program_title": { "id": "uuid", "title": "Infrastruktur Transportasi" },
    "bappenas_partners": [
      { "id": "uuid", "name": "Direktorat Transportasi", "level": "Eselon II", "parent_id": "uuid-eselon-i" }
    ],
    "executing_agencies": [
      { "id": "uuid", "name": "Kementerian PUPR", "level": "Kementerian/Badan/Lembaga" }
    ],
    "implementing_agencies": [
      { "id": "uuid", "name": "Ditjen Bina Marga", "level": "Eselon I" }
    ],
    "locations": [
      { "id": "uuid", "name": "Sumatera Utara", "level": "Provinsi" }
    ],
    "national_priorities": [
      { "id": "uuid", "title": "Ketahanan Pangan" }
    ],
    "project_costs": [
      { "id": "uuid", "funding_type": "Foreign", "funding_category": "Loan", "amount_usd": 500000000 }
    ],
    "lender_indications": [
      { "id": "uuid", "lender": { "id": "uuid", "name": "JICA", "type": "Bilateral" }, "remarks": "Minat seksi 1-3" }
    ],
    "status": "active",
    "is_latest": true,
    "has_newer_revision": false,
    "created_at": "2025-01-15T08:00:00Z",
    "updated_at": "2025-01-15T08:00:00Z"
  }
}
```

### BB Project History

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/bb-projects/:id/history` | read: `bb_project` |

History selalu mengembalikan daftar snapshot revisi. Untuk user ADMIN, response juga menyertakan audit rail ringkas per snapshot (`last_changed_*` dan `audit_entries`) agar terlihat siapa mengubah section/field apa. Untuk STAFF, field audit tidak dikirim karena `audit_log` adalah resource ADMIN only.

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid-bb-project-snapshot",
      "project_identity_id": "uuid-logical-project",
      "blue_book_id": "uuid",
      "bb_code": "BB-2025-001",
      "project_name": "Pembangunan Jalan Tol Trans Sumatera",
      "book_label": "BB 2025-2029 Revisi ke-1",
      "revision_number": 1,
      "revision_year": 2026,
      "book_status": "active",
      "is_latest": true,
      "used_by_downstream": false,
      "last_changed_by": "admin",
      "last_changed_at": "2026-05-01T08:30:00Z",
      "last_change_summary": "Mengubah Informasi proyek: Nama proyek, Durasi",
      "audit_entries": [
        {
          "id": "uuid-audit-log",
          "section": "Informasi proyek",
          "action": "UPDATE",
          "action_label": "Mengubah",
          "changed_fields": ["project_name", "duration"],
          "changed_field_labels": ["Nama proyek", "Durasi"],
          "changed_by_id": "uuid-user",
          "changed_by_username": "admin",
          "changed_at": "2026-05-01T08:30:00Z",
          "summary": "Mengubah Informasi proyek: Nama proyek, Durasi"
        }
      ]
    }
  ]
}
```

---

### LoI (Letter of Intent)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/bb-projects/:bb_project_id/loi` | read: `bb_project` |
| `POST` | `/bb-projects/:bb_project_id/loi` | update: `bb_project` |
| `PUT` | `/bb-projects/:bb_project_id/loi/:id` | update: `bb_project` |
| `DELETE` | `/bb-projects/:bb_project_id/loi/:id` | update: `bb_project` |

**`POST /bb-projects/:bb_project_id/loi` Request:**
```json
{
  "lender_id": "uuid",
  "subject": "Letter of Intent for Trans Sumatra Toll Road",
  "tanggal": "2025-03-10",
  "letter_number": "JICA/LOI/2025/001"   // opsional
}
```

---

## Green Book

### Green Book (Header)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/green-books` | read: `green_book` |
| `GET` | `/green-books/:id` | read: `green_book` |
| `POST` | `/green-books` | create: `green_book` |
| `PUT` | `/green-books/:id` | update: `green_book` |
| `DELETE` | `/green-books/:id` | delete: `green_book` |

**`GET /green-books` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan tahun terbit, nomor revisi, atau status |
| `publish_year` | multi-value number | Filter tahun terbit Green Book |
| `status` | multi-value enum | `active`, `superseded` |

**`POST /green-books` Request:**
```json
{
  "publish_year": 2025,
  "replaces_green_book_id": "uuid-green-book-sebelumnya",
  "revision_number": 0
}
```

Validasi: kombinasi `publish_year` + `revision_number` harus unik. Jika sudah ada, backend mengembalikan `409 CONFLICT`.

---

### Import Green Book Projects

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/green-books/:gb_id/import-projects/template` | ADMIN only |
| `POST` | `/green-books/:gb_id/import-projects/preview` | ADMIN only |
| `POST` | `/green-books/:gb_id/import-projects/execute` | ADMIN only |

**Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Input Data`, `Relasi - BB Project`, `Relasi - EA`, `Relasi - IA`, `Relasi - Locations`, `Relasi - Activities`, `Relasi - Funding Source`, `Relasi - Disbursement Plan`, dan `Relasi - Funding Allocation` |

Workbook diimport ke Green Book target dari `:gb_id`. Sheet relasi memakai `GB Code (*)` sebagai kunci penghubung ke sheet `Input Data`. Import ini tidak membuat header Green Book baru.

**Template:**
`GET /green-books/:gb_id/import-projects/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Input Data`, semua sheet relasi, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot master data dan BB Project aktif saat template dibuat.

**Kolom utama workbook:**

| Sheet | Kolom |
|-------|-------|
| `Input Data` | `Program Title (*)`, `GB Code (*)`, `Project Name (*)`, `Duration`, `Objective`, `Scope of Project` |
| `Relasi - BB Project` | `GB Code (*)`, `BB Code (*)` |
| `Relasi - EA` | `GB Code (*)`, `Executing Agency Name (*)` |
| `Relasi - IA` | `GB Code (*)`, `Implementing Agency Name (*)` |
| `Relasi - Locations` | `GB Code (*)`, `Location Name (*)` |
| `Relasi - Activities` | `GB Code (*)`, `Activity No (*)`, `Activity Name (*)`, `Implementation Location`, `PIU`, `Sort Order` |
| `Relasi - Funding Source` | `GB Code (*)`, `Lender Name (*)`, `Institution Name`, `Currency`, `Loan Original`, `Grant Original`, `Local Original`, `Loan USD`, `Grant USD`, `Local USD` |
| `Relasi - Disbursement Plan` | `GB Code (*)`, `Year (*)`, `Amount USD` |
| `Relasi - Funding Allocation` | `GB Code (*)`, `Activity No (*)`, `Services`, `Constructions`, `Goods`, `Trainings`, `Other` |

Kolom `Duration` pada workbook diisi sebagai angka jumlah bulan. Kolom institution pada `Relasi - EA`, `Relasi - IA`, dan `Relasi - Funding Source` dapat diisi dengan nama jika unik, UUID dari sheet `Master Data`, atau path `Nama Child; Nama Parent; Nama Root;`. Template dropdown memakai path agar nama child yang sama di parent berbeda tetap bisa dipilih tanpa ambigu. Kolom `Relasi - Funding Source.Lender Name` di-resolve ke master Lender berdasarkan `name`; jika tidak cocok, import mencoba fallback ke `short_name` yang unik. Sheet `Panduan` menjelaskan fallback ini agar operator tidak perlu menebak ketika nama Institution sama atau workbook memakai singkatan lender seperti `ADB`, `IFAD`, `EIB`, atau `UKEF`.

**Preview:**
`POST /green-books/:gb_id/import-projects/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /green-books/:gb_id/import-projects/execute` menyimpan data jika hasil pemrosesan tidak memiliki baris gagal.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`. Frontend wajib menampilkan preview dan meminta konfirmasi user sebelum eksekusi.

Baris dengan `GB Code` yang sudah ada dalam Green Book target akan di-skip. `GB Code` yang hanya ada pada revisi lama tidak di-skip; jika cocok dengan revisi sumber, snapshot baru memakai `gb_project_identity_id` yang sama. Relasi BB Project di-resolve ke latest BB Project snapshot saat import dieksekusi. Proyek baru wajib memiliki minimal satu BB Project, EA, IA, dan lokasi. `Currency` kosong dianggap `USD`; jika `USD`, nilai USD disamakan dengan nilai original sehingga user tidak perlu mengisi dua kali. `Year` pada Disbursement Plan harus unik per `GB Code`. Funding Allocation mengacu ke `Activity No`; activity tanpa Funding Allocation eksplisit tetap dibuat dengan allocation bernilai 0.

---

### GB Project

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/green-books/:gb_id/projects` | read: `gb_project` |
| `GET` | `/green-books/:gb_id/projects/:id` | read: `gb_project` |
| `POST` | `/green-books/:gb_id/projects` | create: `gb_project` |
| `PUT` | `/green-books/:gb_id/projects/:id` | update: `gb_project` |
| `DELETE` | `/green-books/:gb_id/projects/:id` | delete: `gb_project` |

`DELETE /green-books/:gb_id/projects/:id` melakukan hard delete. Backend menolak penghapusan jika GB Project masih menjadi referensi Daftar Kegiatan, Loan Agreement, atau Monitoring. Untuk record yang sudah memiliki relasi turunan, STAFF menerima `403 FORBIDDEN` dan ADMIN menerima `409 CONFLICT`; keduanya berisi `error.details[]` dengan daftar relasi yang harus dibersihkan terlebih dahulu.

**`GET /green-books/:gb_id/projects` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan kode/nama proyek Green Book, kode/nama proyek Blue Book terkait, Executing Agency, lokasi, atau lender funding source |
| `bb_project_ids` | multi-value UUID | Filter relasi proyek Blue Book |
| `executing_agency_ids` | multi-value UUID | Filter institution role `Executing Agency` |
| `location_ids` | multi-value UUID | Filter region lokasi proyek |
| `status` | multi-value enum | `active` saja; Project Green Book yang dihapus tidak tersedia karena hard delete |

**`POST /green-books/:gb_id/projects` Request:**
```json
{
  "gb_project_identity_id": "uuid-logical-gb-project-opsional",
  "program_title_id": "uuid",
  "gb_code": "GB-2025-001",
  "project_name": "Trans Sumatra Toll Road Section 1",
  "duration": 60,
  "objective": "Meningkatkan konektivitas...",
  "scope_of_project": "Pembangunan 200km...",
  "bb_project_ids": ["uuid-bb-project"],
  "bappenas_partner_ids": ["uuid-mitra-bappenas-1", "uuid-mitra-bappenas-2"],
  "executing_agency_ids": ["uuid"],
  "implementing_agency_ids": ["uuid"],
  "location_ids": ["uuid-sumut"],
  "activities": [
    {
      "activity_name": "Land Acquisition",
      "implementation_location": "Kabupaten Deli Serdang",
      "piu": "Balai Besar Pelaksanaan Jalan Nasional",
      "sort_order": 1
    }
  ],
  "funding_sources": [
    {
      "lender_id": "uuid-jica",
      "institution_id": "uuid-ditjen-bina-marga",
      "currency": "JPY",
      "loan_original": 45000000000,
      "grant_original": 0,
      "local_original": 7500000000,
      "loan_usd": 300000000,
      "grant_usd": 0,
      "local_usd": 50000000
    }
  ],
  "disbursement_plan": [
    { "year": 2025, "amount_usd": 50000000 },
    { "year": 2026, "amount_usd": 100000000 },
    { "year": 2027, "amount_usd": 150000000 }
  ],
  "funding_allocations": [
    {
      "activity_index": 0,
      "services": 10000000,
      "constructions": 200000000,
      "goods": 5000000,
      "trainings": 1000000,
      "other": 500000
    }
  ]
}
```

> **Catatan:** `funding_allocations[].activity_index` merujuk ke index array `activities` dalam request yang sama. Setelah disimpan, relasi menggunakan `gb_activity_id`.
> **Versioning:** `bb_project_ids` boleh berisi snapshot lama, tetapi backend selalu menyimpan concrete latest BB Project snapshot untuk logical project tersebut pada saat GB Project dibuat/diupdate.
> **Relasi BB:** semua `bb_project_ids` pada satu GB Project harus resolve ke header Blue Book yang sama. Satu BB Project boleh dipakai oleh lebih dari satu GB Project.
> **Mitra Kerja Bappenas:** `bappenas_partner_ids` opsional dan boleh kosong pada BB Project, GB Project, dan DK Project.
> **Currency:** Funding Source GB adalah titik awal pencatatan currency downstream. Jika `funding_sources[].currency` adalah `USD`, backend menyimpan nilai USD sama dengan nilai original.

Frontend dapat membuka form GB Project dari action BB Project "Tambah Green Book" dengan query `source_bb_project_id` dan `source_mode`. Dialog memakai checkbox "Gunakan data di Blue Book sebagai data Green Book": tidak dicentang mengirim `source_mode=new` dan hanya membawa BB Code serta relasi BB Project; dicentang mengirim `source_mode=existing` untuk mengisi field yang sama dari BB Project sumber, tetapi tetap editable sebelum disimpan.

**`GET /green-books/:gb_id/projects/:id` Response `200` menambahkan field versioning:**
```json
{
  "data": {
    "id": "uuid-gb-project-snapshot",
    "gb_project_identity_id": "uuid-logical-gb-project",
    "green_book_id": "uuid",
    "gb_code": "GB-2025-001",
    "is_latest": true,
    "has_newer_revision": false,
    "bappenas_partners": [
      { "id": "uuid", "name": "Direktorat Transportasi", "level": "Eselon II", "parent_id": "uuid-eselon-i" }
    ],
    "bb_projects": [
      {
        "id": "uuid-bb-project-snapshot",
        "project_identity_id": "uuid-logical-project",
        "bb_code": "BB-2025-001",
        "project_name": "Pembangunan Jalan Tol Trans Sumatera",
        "is_latest": true,
        "has_newer_revision": false
      }
    ]
  }
}
```

### GB Project History

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/gb-projects/:id/history` | read: `gb_project` |

History selalu mengembalikan daftar snapshot revisi. Untuk user ADMIN, response juga menyertakan audit rail ringkas per snapshot (`last_changed_*` dan `audit_entries`) agar terlihat siapa mengubah section/field apa. Untuk STAFF, field audit tidak dikirim karena `audit_log` adalah resource ADMIN only.

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid-gb-project-snapshot",
      "gb_project_identity_id": "uuid-logical-gb-project",
      "green_book_id": "uuid",
      "gb_code": "GB-2025-001",
      "project_name": "Trans Sumatra Section 1",
      "book_label": "GB 2025 Revisi ke-1",
      "publish_year": 2025,
      "revision_number": 1,
      "book_status": "active",
      "is_latest": true,
      "used_by_downstream": false,
      "bb_projects": [],
      "last_changed_by": "admin",
      "last_changed_at": "2026-05-01T08:35:00Z",
      "last_change_summary": "Mengubah Funding Source: Lender, Pinjaman USD",
      "audit_entries": [
        {
          "id": "uuid-audit-log",
          "section": "Funding Source",
          "action": "UPDATE",
          "action_label": "Mengubah",
          "changed_fields": ["lender_id", "loan_usd"],
          "changed_field_labels": ["Lender", "Pinjaman USD"],
          "changed_by_id": "uuid-user",
          "changed_by_username": "admin",
          "changed_at": "2026-05-01T08:35:00Z",
          "summary": "Mengubah Funding Source: Lender, Pinjaman USD"
        }
      ]
    }
  ]
}
```

---

## Daftar Kegiatan

### Import Daftar Kegiatan

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/daftar-kegiatan/import/template` | ADMIN only |
| `POST` | `/daftar-kegiatan/import/preview` | ADMIN only |
| `POST` | `/daftar-kegiatan/import/execute` | ADMIN only |

**Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Daftar Kegiatan`, `Input Data`, `Relasi - GB Project`, `Relasi - Locations`, `Relasi - Financing Detail`, `Relasi - Loan Allocation`, dan `Relasi - Activity Detail` |

Import ini membuat header Daftar Kegiatan baru beserta DK Project dan seluruh relasinya. Workbook mendukung multi header DK. `DK Key (*)` adalah kunci sementara workbook untuk header; `Project Key (*)` wajib unik per `DK Key` dan hanya dipakai untuk menghubungkan sheet relasi. `Letter Number (*)` wajib untuk import dan menjadi idempotency key.

**Template:**
`GET /daftar-kegiatan/import/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Daftar Kegiatan`, `Input Data`, semua sheet relasi, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot master data, GB Project aktif, dan referensi allowed lender per GB Project saat template dibuat.

**Kolom workbook:**

| Sheet | Kolom |
|-------|-------|
| `Daftar Kegiatan` | `DK Key (*)`, `Letter Number (*)`, `Subject (*)`, `Date (*)` |
| `Input Data` | `DK Key (*)`, `Project Key (*)`, `Project Name (*)`, `Program Title`, `Executing Agency Name (*)`, `Duration`, `Objectives` |
| `Relasi - GB Project` | `DK Key (*)`, `Project Key (*)`, `GB Code (*)` |
| `Relasi - Locations` | `DK Key (*)`, `Project Key (*)`, `Location Name (*)` |
| `Relasi - Financing Detail` | `DK Key (*)`, `Project Key (*)`, `Lender Name (*)`, `Currency`, `Amount Original`, `Grant Original`, `Counterpart Original`, `Amount USD`, `Grant USD`, `Counterpart USD`, `Remarks` |
| `Relasi - Loan Allocation` | `DK Key (*)`, `Project Key (*)`, `Institution Name (*)`, `Currency`, `Amount Original`, `Grant Original`, `Counterpart Original`, `Amount USD`, `Grant USD`, `Counterpart USD`, `Remarks` |
| `Relasi - Activity Detail` | `DK Key (*)`, `Project Key (*)`, `Activity No (*)`, `Activity Name (*)` |

**Preview:**
`POST /daftar-kegiatan/import/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /daftar-kegiatan/import/execute` menjalankan import hanya jika hasil validasi memiliki `total_failed = 0`. Jika masih ada failed, response error validasi dan data tidak disimpan.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`.

Jika `Letter Number` sudah ada di DB, header dan semua project/relasi di bawahnya berstatus `skip`. Duplikat `Letter Number` dalam workbook berstatus `failed`. Project baru wajib punya Project Name, Executing Agency, minimal 1 GB Project aktif, Location, Financing Detail, Loan Allocation, dan Activity Detail. `Project Name` adalah nama snapshot di Daftar Kegiatan dan boleh berbeda dari nama Green Book. `Program Title` opsional, tetapi jika diisi harus ada di master data. Kolom institution pada `Input Data.Executing Agency Name` dan `Relasi - Loan Allocation.Institution Name` dapat diisi dengan nama jika unik, UUID dari sheet `Master Data`, atau path `Nama Child; Nama Parent; Nama Root;`. Sheet `Panduan` menjelaskan fallback ini dan Preview tetap gagal untuk nama polos yang ambigu. Lender Financing Detail harus berasal dari allowed lender GB Project terkait. `Currency` kosong dianggap `USD`; jika diisi harus kode ISO 4217 yang aktif di Master Currency. Amount kosong dianggap `0` dan tidak boleh negatif. `Activity No` duplikat per project berstatus `failed`.
Kolom `Duration` pada workbook diisi sebagai angka jumlah bulan.
`Date` pada sheet `Daftar Kegiatan` memakai format `YYYY-MM-DD`.

`GB Project` pada DK di-resolve ke latest GB Project snapshot saat DK Project dibuat atau saat pilihan GB diganti eksplisit. Setelah tersimpan, relasi `dk_project_gb_project` tetap menunjuk concrete snapshot yang tersimpan dan tidak auto-pindah ketika ada revisi BB/GB baru.

Pada form create/edit DK Project, picker `GB Project` ditampilkan sebagai field pertama. Saat user memilih GB Project, frontend mengisi otomatis field DK yang memiliki padanan dari GB Project terpilih: nama proyek Daftar Kegiatan dari nama proyek Green Book, program title, executing agency, Mitra Kerja Bappenas, durasi bulan, tujuan/objective, lokasi, rincian pembiayaan dari funding source, alokasi pinjaman dari institution funding source atau institution proyek, dan rincian kegiatan dari activities GB. Hasil autofill tetap dapat diedit user sebelum request `POST` atau `PUT` dikirim.
Jika currency hasil autofill adalah `USD`, field USD tidak perlu diisi terpisah karena backend menyamakan nilai USD dengan nilai original.

---

### Daftar Kegiatan (Header Surat)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/daftar-kegiatan` | read: `daftar_kegiatan` |
| `GET` | `/daftar-kegiatan/:id` | read: `daftar_kegiatan` |
| `POST` | `/daftar-kegiatan` | create: `daftar_kegiatan` |
| `PUT` | `/daftar-kegiatan/:id` | update: `daftar_kegiatan` |
| `DELETE` | `/daftar-kegiatan/:id` | delete: `daftar_kegiatan` |

**`GET /daftar-kegiatan` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan `subject`, `letter_number`, atau tanggal surat |
| `date_from` | date `YYYY-MM-DD` | Batas awal tanggal surat |
| `date_to` | date `YYYY-MM-DD` | Batas akhir tanggal surat |

**`POST /daftar-kegiatan` Request:**
```json
{
  "letter_number": "B-001/D.8/PP.01.02/01/2025",   // opsional
  "subject": "Daftar Kegiatan Pinjaman Luar Negeri TA 2025",
  "tanggal": "2025-02-01"
}
```

---

### DK Project (Proyek dalam Surat)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/daftar-kegiatan/:dk_id/projects` | read: `daftar_kegiatan` |
| `GET` | `/daftar-kegiatan/:dk_id/projects/:id` | read: `daftar_kegiatan` |
| `POST` | `/daftar-kegiatan/:dk_id/projects` | create: `daftar_kegiatan` |
| `PUT` | `/daftar-kegiatan/:dk_id/projects/:id` | update: `daftar_kegiatan` |
| `DELETE` | `/daftar-kegiatan/:dk_id/projects/:id` | delete: `daftar_kegiatan` |

**`GET /daftar-kegiatan/:dk_id/projects` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan nama proyek Daftar Kegiatan, proyek Green Book terkait, objectives, lokasi, lender, atau activity detail |
| `gb_project_ids` | multi-value UUID | Filter relasi proyek Green Book |
| `executing_agency_ids` | multi-value UUID | Filter institution/executing agency DK Project |
| `location_ids` | multi-value UUID | Filter region lokasi proyek |
| `lender_ids` | multi-value UUID | Filter lender financing detail |

**`POST /daftar-kegiatan/:dk_id/projects` Request:**
```json
{
  "program_title_id": "uuid",
  "institution_id": "uuid-executing-agency",
  "project_name": "Trans Sumatra Toll Road Section 1 - DK",
  "duration": 60,
  "objectives": "Meningkatkan konektivitas...",
  "gb_project_ids": ["uuid-gb-project-1"],
  "bappenas_partner_ids": ["uuid-mitra-bappenas-1", "uuid-mitra-bappenas-2"],
  "location_ids": ["uuid-sumut"],
  "financing_details": [
    {
      "lender_id": "uuid-jica",
      "currency": "JPY",
      "amount_original": 45000000000,
      "grant_original": 0,
      "counterpart_original": 7500000000,
      "amount_usd": 300000000,
      "grant_usd": 0,
      "counterpart_usd": 50000000,
      "remarks": "Termasuk biaya supervisi"
    }
  ],
  "loan_allocations": [
    {
      "institution_id": "uuid-ditjen-bina-marga",
      "currency": "JPY",
      "amount_original": 45000000000,
      "grant_original": 0,
      "counterpart_original": 7500000000,
      "amount_usd": 300000000,
      "grant_usd": 0,
      "counterpart_usd": 50000000,
      "remarks": null
    }
  ],
  "activity_details": [
    { "activity_number": 1, "activity_name": "Pembebasan Lahan" },
    { "activity_number": 2, "activity_name": "Konstruksi Jalan" },
    { "activity_number": 3, "activity_name": "Supervisi" }
  ]
}
```

**`GET /daftar-kegiatan/:dk_id/projects/:id` Response `200`** menyertakan `bappenas_partners` sebagai array Mitra Kerja Bappenas Eselon II. Field ini boleh kosong.

---

## Loan Agreement

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/loan-agreements` | read: `loan_agreement` |
| `GET` | `/loan-agreements/:id` | read: `loan_agreement` |
| `POST` | `/loan-agreements` | create: `loan_agreement` |
| `PUT` | `/loan-agreements/:id` | update: `loan_agreement` |
| `DELETE` | `/loan-agreements/:id` | delete: `loan_agreement` |

**`POST /loan-agreements` Request:**
```json
{
  "dk_project_id": "uuid",
  "lender_id": "uuid-jica",
  "loan_code": "IP-603",
  "agreement_date": "2025-03-15",
  "effective_date": "2025-06-01",
  "original_closing_date": "2030-12-31",
  "closing_date": "2030-12-31",
  "currency": "JPY",
  "amount_original": 45000000000,
  "amount_usd": 300000000
}
```

**`GET /loan-agreements/:id` Response `200`:**
```json
{
  "data": {
    "id": "uuid",
    "loan_code": "IP-603",
    "dk_project": { "id": "uuid", "objectives": "..." },
    "lender": { "id": "uuid", "name": "JICA", "type": "Bilateral" },
    "agreement_date": "2025-03-15",
    "effective_date": "2025-06-01",
    "original_closing_date": "2030-12-31",
    "closing_date": "2031-12-31",
    "is_extended": true,
    "extension_days": 365,
    "currency": "JPY",
    "amount_original": 45000000000,
    "amount_usd": 300000000,
    "created_at": "2025-03-15T08:00:00Z",
    "updated_at": "2025-03-15T08:00:00Z"
  }
}
```

**`GET /loan-agreements` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `search` | Cari berdasarkan `loan_code`, nama lender, atau short name lender |
| `lender_id` | Filter by lender |
| `is_extended` | Filter: `true` / `false` |
| `closing_date_before` | Filter LA yang akan berakhir sebelum tanggal ini |

---

## Monitoring Disbursement

### Monitoring (Level LA)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/loan-agreements/:la_id/monitoring` | read: `monitoring_disbursement` |
| `GET` | `/loan-agreements/:la_id/monitoring/:id` | read: `monitoring_disbursement` |
| `POST` | `/loan-agreements/:la_id/monitoring` | create: `monitoring_disbursement` |
| `PUT` | `/loan-agreements/:la_id/monitoring/:id` | update: `monitoring_disbursement` |
| `DELETE` | `/loan-agreements/:la_id/monitoring/:id` | delete: `monitoring_disbursement` |

**`GET /loan-agreements/:la_id/monitoring` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan tahun anggaran, triwulan, atau nama komponen |
| `budget_year` | number | Filter tahun anggaran |
| `quarter` | enum | `TW1`, `TW2`, `TW3`, `TW4` |

**`POST /loan-agreements/:la_id/monitoring` Request:**
```json
{
  "budget_year": 2025,
  "quarter": "TW1",
  "exchange_rate_usd_idr": 15750.50,
  "exchange_rate_la_idr": 105.25,
  "planned_la": 500000000,
  "planned_usd": 3333333,
  "planned_idr": 52500000000,
  "realized_la": 420000000,
  "realized_usd": 2800000,
  "realized_idr": 44100000000,
  "komponen": [                    // opsional
    {
      "component_name": "Konstruksi",
      "planned_la": 400000000,
      "planned_usd": 2666667,
      "planned_idr": 42000000000,
      "realized_la": 380000000,
      "realized_usd": 2533333,
      "realized_idr": 39900000000
    },
    {
      "component_name": "Supervisi",
      "planned_la": 100000000,
      "planned_usd": 666667,
      "planned_idr": 10500000000,
      "realized_la": 40000000,
      "realized_usd": 266667,
      "realized_idr": 4200000000
    }
  ]
}
```

**`GET /loan-agreements/:la_id/monitoring` Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "budget_year": 2025,
      "quarter": "TW1",
      "exchange_rate_usd_idr": 15750.50,
      "exchange_rate_la_idr": 105.25,
      "planned_la": 500000000,
      "planned_usd": 3333333,
      "planned_idr": 52500000000,
      "realized_la": 420000000,
      "realized_usd": 2800000,
      "realized_idr": 44100000000,
      "penyerapan_pct": 84.0,
      "komponen": [
        {
          "id": "uuid",
          "component_name": "Konstruksi",
          "planned_la": 400000000,
          "realized_la": 380000000
        }
      ]
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 8, "total_pages": 1 }
}
```

---

## User Management

> Semua endpoint ini hanya dapat diakses oleh **ADMIN**.

### Users

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/users` | ADMIN only |
| `GET` | `/users/:id` | ADMIN only |
| `POST` | `/users` | ADMIN only |
| `PUT` | `/users/:id` | ADMIN only |
| `DELETE` | `/users/:id` | ADMIN only |

**`POST /users` Request:**
```json
{
  "username": "staff02",
  "email": "staff02@bappenas.go.id",
  "password": "initialPassword123",
  "role": "STAFF"
}
```

---

### User Permissions

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/users/:id/permissions` | ADMIN only |
| `PUT` | `/users/:id/permissions` | ADMIN only |

**`PUT /users/:id/permissions` Request:**
```json
{
  "permissions": [
    {
      "module": "bb_project",
      "can_create": true,
      "can_read": true,
      "can_update": true,
      "can_delete": false
    },
    {
      "module": "gb_project",
      "can_create": false,
      "can_read": true,
      "can_update": false,
      "can_delete": false
    },
    {
      "module": "monitoring_disbursement",
      "can_create": true,
      "can_read": true,
      "can_update": true,
      "can_delete": false
    }
  ]
}
```

> **Catatan:** `PUT` ini bersifat **replace-all** — semua permission user diganti sesuai payload. Module yang tidak disertakan akan dihapus permissionnya.

---

## SSE — Realtime Events

### `GET /events`

**Permission:** Authenticated

Client subscribe ke endpoint ini untuk menerima notifikasi realtime. Koneksi bersifat long-lived (Server-Sent Events).

**Response (stream):**
```
event: bb_project.created
data: {"id":"uuid","bb_code":"BB-2025-002","project_name":"...","created_by":"staff01"}

event: monitoring.updated
data: {"id":"uuid","loan_agreement_id":"uuid","quarter":"TW2","updated_by":"staff02"}
```

**Event Types:**

| Event | Trigger |
|-------|---------|
| `bb_project.created` | BB Project baru dibuat |
| `bb_project.updated` | BB Project diupdate |
| `gb_project.created` | GB Project baru dibuat |
| `gb_project.updated` | GB Project diupdate |
| `daftar_kegiatan.created` | Daftar Kegiatan baru dibuat |
| `loan_agreement.created` | Loan Agreement baru dibuat |
| `loan_agreement.extended` | Closing Date LA diperbarui |
| `monitoring.created` | Entri monitoring baru |
| `monitoring.updated` | Entri monitoring diupdate |

---

## Dashboard & Aggregasi

### `GET /dashboard/summary`

**Permission:** Authenticated

**Response `200`:**
```json
{
  "data": {
    "total_bb_projects": 120,
    "total_gb_projects": 85,
    "total_loan_agreements": 42,
    "total_amount_usd": 15000000000,
    "total_realized_usd": 8500000000,
    "overall_absorption_pct": 56.7,
    "active_monitoring": 38
  }
}
```

---

### `GET /dashboard/monitoring-summary`

**Permission:** Authenticated

**Query Params:**

| Param | Keterangan |
|-------|-----------|
| `budget_year` | Filter tahun anggaran |
| `quarter` | Filter: `TW1`, `TW2`, `TW3`, `TW4` |
| `lender_id` | Filter by lender |

**Response `200`:**
```json
{
  "data": {
    "budget_year": 2025,
    "quarter": "TW1",
    "total_planned_usd": 500000000,
    "total_realized_usd": 380000000,
    "absorption_pct": 76.0,
    "by_lender": [
      {
        "lender": { "id": "uuid", "name": "JICA" },
        "planned_usd": 300000000,
        "realized_usd": 240000000,
        "absorption_pct": 80.0
      }
    ]
  }
}
```

---

### `GET /projects`

**Permission:** read: `bb_project`

Menampilkan master table seluruh BB Project aktif. Tanpa query filter, endpoint default hanya mengembalikan latest snapshot per `project_identity_id` supaya revisi lama tidak double-count. Gunakan `include_history=true` untuk melihat semua snapshot historis.

Ringkasan pendanaan (`summary`) dihitung dari seluruh hasil filter, bukan hanya page pagination saat ini. Untuk project yang masih di tahap Blue Book, total memakai `bb_project_cost`: pinjaman = `Foreign/Loan`, hibah = `Foreign/Grant`, dana pendamping = seluruh `Counterpart`. Untuk project yang sudah memiliki relasi Green Book, total memakai `gb_funding_source`: `loan_usd`, `grant_usd`, dan `local_usd`.

**Query Params:**

| Param | Keterangan |
|-------|-----------|
| `page`, `limit` | Pagination standar |
| `sort`, `order` | Sorting standar. `sort`: `project_name`, `bb_code`, `loan_types`, `indication_lenders`, `executing_agencies`, `fixed_lenders`, `project_status`, `pipeline_status`, `program_title`, `locations`, `foreign_loan_usd`, `dk_dates`. `order`: `asc` atau `desc` |
| `loan_types` | Multi value: `Bilateral`, `Multilateral`, `KSA` |
| `indication_lender_ids` | Multi value UUID lender dari `lender_indication` BB |
| `executing_agency_ids` | Multi value UUID institution role `Executing Agency` |
| `fixed_lender_ids` | Multi value UUID lender dari `gb_funding_source` Green Book |
| `project_statuses` | Multi value: `Pipeline`, `Ongoing` |
| `pipeline_statuses` | Multi value: `BB`, `GB`, `DK`, `LA`, `Monitoring` |
| `program_title_ids` | Multi value UUID program title |
| `region_ids` | Multi value UUID region/location |
| `foreign_loan_min`, `foreign_loan_max` | Range nilai pinjaman foreign loan dalam USD |
| `dk_date_from`, `dk_date_to` | Range tanggal DK, format `YYYY-MM-DD` |
| `search` | Search global untuk kode/nama proyek, indikasi lender, fixed lender Green Book, dan executing agency |
| `include_history` | `true` untuk menampilkan semua snapshot, default `false` |

Multi value dapat dikirim sebagai repeated query param (`loan_types=Bilateral&loan_types=KSA`), comma-separated value, atau format array query string (`loan_types[]=Bilateral`).

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "blue_book_id": "uuid",
      "project_identity_id": "uuid-logical-project",
      "bb_code": "BB-2025-001",
      "project_name": "Pembangunan Jalan Tol Trans Sumatera",
      "loan_types": ["Bilateral"],
      "indication_lenders": ["JICA"],
      "executing_agencies": ["Kementerian PUPR"],
      "fixed_lenders": ["JICA"],
      "project_status": "Ongoing",
      "pipeline_status": "Monitoring",
      "program_title": "Infrastruktur Transportasi",
      "locations": ["Sumatera Utara"],
      "foreign_loan_usd": 250000000,
      "dk_dates": ["2025-02-01"],
      "is_latest": true,
      "has_newer_revision": false,
      "blue_book_revision_label": "BB 2025-2029"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 1, "total_pages": 1 },
  "summary": {
    "total_loan_usd": 250000000,
    "total_grant_usd": 0,
    "total_counterpart_usd": 50000000
  }
}
```

### `GET /projects/export`

**Permission:** read: `bb_project`

Mengunduh workbook Excel (`.xlsx`) berisi seluruh project yang cocok dengan filter aktif. Endpoint memakai query param yang sama dengan `GET /projects`; `page` dan `limit` diabaikan karena export selalu mengambil semua hasil filter. Sorting tetap mengikuti `sort` dan `order`. Sheet `Ringkasan` berisi total pendanaan serta daftar filter aktif yang dipakai saat export.

**Response `200`:** `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` dengan `Content-Disposition: attachment`.

---

### `GET /projects/:bb_project_id/journey`

**Permission:** read: `bb_project`

Menampilkan seluruh alur proyek dari BB → GB → DK → LA → Monitoring dalam satu response.

**Response `200`:**
```json
{
  "data": {
    "bb_project": {
      "id": "uuid",
      "blue_book_id": "uuid",
      "project_identity_id": "uuid-logical-project",
      "bb_code": "BB-2025-001",
      "project_name": "Trans Sumatra Toll Road",
      "blue_book_revision_label": "BB 2025-2029",
      "is_latest": true,
      "has_newer_revision": false,
      "latest_bb_project_id": "uuid",
      "latest_blue_book_revision_label": "BB 2025-2029",
      "lender_indications": [
        {
          "id": "uuid",
          "lender": { "id": "uuid", "name": "JICA", "short_name": "JICA", "type": "Bilateral" },
          "remarks": "Indicative"
        }
      ]
    },
    "loi": [
      {
        "id": "uuid",
        "lender": { "id": "uuid", "name": "JICA", "short_name": "JICA", "type": "Bilateral" },
        "subject": "Letter of Intent",
        "date": "2025-03-10",
        "letter_number": "LoI-001"
      }
    ],
    "gb_projects": [
      {
        "id": "uuid",
        "green_book_id": "uuid",
        "gb_project_identity_id": "uuid-logical-gb-project",
        "gb_code": "GB-2025-001",
        "project_name": "Trans Sumatra Section 1",
        "status": "active",
        "green_book_revision_label": "GB 2025",
        "is_latest": true,
        "has_newer_revision": false,
        "latest_gb_project_id": "uuid",
        "latest_green_book_revision_label": "GB 2025",
        "funding_sources": [
          {
            "id": "uuid",
            "lender": { "id": "uuid", "name": "JICA", "short_name": "JICA", "type": "Bilateral" },
            "institution": { "id": "uuid", "name": "Kementerian PUPR", "short_name": "PUPR" },
            "currency": "USD",
            "loan_original": 300000000,
            "grant_original": 0,
            "local_original": 0,
            "loan_usd": 300000000,
            "grant_usd": 0,
            "local_usd": 0
          }
        ],
        "dk_projects": [
          {
            "id": "uuid",
            "project_name": "Trans Sumatra Section 1 - DK",
            "objectives": "Meningkatkan konektivitas",
            "daftar_kegiatan": {
              "id": "uuid",
              "subject": "DK TA 2025",
              "date": "2025-02-01",
              "letter_number": "B-001/2025"
            },
            "loan_agreement": {
              "id": "uuid",
              "loan_code": "IP-603",
              "lender": { "id": "uuid", "name": "JICA", "short_name": "JICA", "type": "Bilateral" },
              "agreement_date": "2025-05-01",
              "effective_date": "2025-06-01",
              "original_closing_date": "2030-12-31",
              "closing_date": "2030-12-31",
              "is_extended": false,
              "extension_days": 0,
              "currency": "USD",
              "amount_original": 300000000,
              "amount_usd": 300000000,
              "monitoring": [
                {
                  "id": "uuid",
                  "budget_year": 2025,
                  "quarter": "TW1",
                  "planned_usd": 3333333,
                  "realized_usd": 2800000,
                  "absorption_pct": 84.0
                }
              ]
            }
          }
        ]
      }
    ]
  }
}
```

---

## Catatan Implementasi

- Semua timestamp menggunakan **ISO 8601** dengan timezone (`2025-01-15T08:00:00Z`).
- Field `id` selalu **UUID v4** — tidak ada integer ID yang diekspos ke client.
- Endpoint yang mengembalikan data hierarki (seperti `/journey`) tidak menggunakan pagination — data di-fetch sekaligus karena jumlahnya terbatas per proyek.
- `penyerapan_pct` (persentase penyerapan) dihitung di server: `(realisasi / rencana) * 100`, tidak disimpan di database.
- Untuk endpoint list yang memiliki banyak relasi (seperti GB Project), response default hanya memuat field ringkas. Gunakan query param `?expand=true` untuk mendapatkan nested object lengkap.
- Perubahan permission user via `PUT /users/:id/permissions` bersifat **transaksional** — semua permission diupdate dalam satu transaksi, gagal sebagian berarti tidak ada yang tersimpan.
