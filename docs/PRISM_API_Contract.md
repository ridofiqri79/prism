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

Baris yang sudah ada akan di-skip. Detail baris preview dikembalikan di `sheets[].rows` dengan `status`: `create`, `skip`, atau `failed`, sehingga frontend dapat memberi tab/filter sebelum eksekusi. Baris yang gagal validasi juga dikembalikan di `sheets[].errors`. Frontend wajib meminta preview terlebih dahulu sebelum user menekan eksekusi import.

### Country

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/countries` | read: `country` |
| `POST` | `/master/countries` | create: `country` |
| `PUT` | `/master/negara/:id` | update: `country` |
| `DELETE` | `/master/negara/:id` | delete: `country` |

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
| `type` | Filter: `Bilateral`, `Multilateral`, `KSA` |

---

### Institution

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/institutions` | read: `institution` |
| `GET` | `/master/institutions/:id` | read: `institution` |
| `POST` | `/master/institutions` | create: `institution` |
| `PUT` | `/master/institutions/:id` | update: `institution` |
| `DELETE` | `/master/institutions/:id` | delete: `institution` |

**`POST /master/institutions` Request:**
```json
{
  "name": "Kementerian PUPR",
  "level": "Kementerian",
  "parent_id": null           // null untuk Kementerian, uuid untuk Eselon I
}
```

**`GET /master/institutions` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `level` | Filter: `Kementerian`, `Eselon I` |
| `parent_id` | Filter by parent |

---

### Region

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/regions` | read: `region` |
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
| `type` | Filter: `COUNTRY`, `PROVINCE`, `CITY` |
| `parent_code` | Filter by parent code — untuk load CITY per PROVINCE |

---

### Program Title

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/program-titles` | read: `program_title` |
| `POST` | `/master/program-titles` | create: `program_title` |
| `PUT` | `/master/program-titles/:id` | update: `program_title` |
| `DELETE` | `/master/program-titles/:id` | delete: `program_title` |

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
| `POST` | `/master/bappenas-partners` | create: `bappenas_partner` |
| `PUT` | `/master/bappenas-partners/:id` | update: `bappenas_partner` |
| `DELETE` | `/master/bappenas-partners/:id` | delete: `bappenas_partner` |

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

**`POST /blue-books` Request:**
```json
{
  "period_id": "uuid",
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
`GET /blue-books/:bb_id/import-projects/template` mengunduh workbook `.xlsx` dengan sheet `Panduan` yang deskriptif, `Master Data`, `Input Data`, dan semua sheet relasi. Kolom relasi memiliki dropdown Excel dari master data dan BB Code pada sheet relasi memiliki dropdown dari `Input Data`. Sheet `Master Data` berisi snapshot master data saat template dibuat; national priority difilter ke period milik Blue Book target. Workbook juga memiliki sheet `_Dropdowns` tersembunyi sebagai sumber pilihan dropdown.

**Kolom utama workbook:**

| Sheet | Kolom |
|-------|-------|
| `Input Data` | `Program Title (*)`, `Bappenas Partner`, `BB Code (*)`, `Project Name (*)`, `Duration`, `Objective`, `Scope of Work`, `Outputs`, `Outcomes` |
| `Relasi - EA` | `BB Code (*)`, `Executing Agency Name (*)` |
| `Relasi - IA` | `BB Code (*)`, `Implementing Agency Name (*)` |
| `Relasi - Locations` | `BB Code (*)`, `Location Name (*)` |
| `Relasi - National Priority` | `BB Code (*)`, `National Priority Name (*)` |
| `Relasi - Project Cost` | `BB Code (*)`, `Funding Type (*)`, `Funding Category (*)`, `Amount USD` |
| `Relasi - Lender Indication` | `BB Code (*)`, `Lender Name (*)`, `Keterangan` |

**Preview:**
`POST /blue-books/:bb_id/import-projects/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /blue-books/:bb_id/import-projects/execute` menyimpan data jika hasil pemrosesan tidak memiliki baris gagal.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`. Frontend wajib menampilkan preview dan meminta konfirmasi user sebelum eksekusi.

Baris dengan `BB Code` yang sudah ada di database akan di-skip. Relasi valid akan dibuat bersama proyek baru; relasi untuk proyek yang di-skip ikut di-skip. National Priority divalidasi terhadap period milik Blue Book target.

---

### BB Project

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/blue-books/:bb_id/projects` | read: `bb_project` |
| `GET` | `/blue-books/:bb_id/projects/:id` | read: `bb_project` |
| `POST` | `/blue-books/:bb_id/projects` | create: `bb_project` |
| `PUT` | `/blue-books/:bb_id/projects/:id` | update: `bb_project` |
| `DELETE` | `/blue-books/:bb_id/projects/:id` | delete: `bb_project` |

**`POST /blue-books/:bb_id/projects` Request:**
```json
{
  "program_title_id": "uuid",
  "bappenas_partner_id": "uuid",
  "bb_code": "BB-2025-001",
  "project_name": "Pembangunan Jalan Tol Trans Sumatera",
  "duration": "2025-2030",
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
    "bb_code": "BB-2025-001",
    "project_name": "Pembangunan Jalan Tol Trans Sumatera",
    "program_title": { "id": "uuid", "title": "Infrastruktur Transportasi" },
    "bappenas_partner": {
      "id": "uuid",
      "name": "Direktorat Transportasi",
      "level": "Eselon II",
      "parent": { "id": "uuid", "name": "Deputi Bidang Sarana dan Prasarana", "level": "Eselon I" }
    },
    "executing_agencies": [
      { "id": "uuid", "name": "Kementerian PUPR", "level": "Kementerian" }
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
    "created_at": "2025-01-15T08:00:00Z",
    "updated_at": "2025-01-15T08:00:00Z"
  }
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

**`POST /green-books` Request:**
```json
{
  "publish_year": 2025,
  "revision_number": 0
}
```

---

### GB Project

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/green-books/:gb_id/projects` | read: `gb_project` |
| `GET` | `/green-books/:gb_id/projects/:id` | read: `gb_project` |
| `POST` | `/green-books/:gb_id/projects` | create: `gb_project` |
| `PUT` | `/green-books/:gb_id/projects/:id` | update: `gb_project` |
| `DELETE` | `/green-books/:gb_id/projects/:id` | delete: `gb_project` |

**`POST /green-books/:gb_id/projects` Request:**
```json
{
  "program_title_id": "uuid",
  "gb_code": "GB-2025-001",
  "project_name": "Trans Sumatra Toll Road Section 1",
  "duration": "2025-2030",
  "objective": "Meningkatkan konektivitas...",
  "scope_of_project": "Pembangunan 200km...",
  "bb_project_ids": ["uuid-bb-project"],
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

---

## Daftar Kegiatan

### Daftar Kegiatan (Header Surat)

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/daftar-kegiatan` | read: `daftar_kegiatan` |
| `GET` | `/daftar-kegiatan/:id` | read: `daftar_kegiatan` |
| `POST` | `/daftar-kegiatan` | create: `daftar_kegiatan` |
| `PUT` | `/daftar-kegiatan/:id` | update: `daftar_kegiatan` |
| `DELETE` | `/daftar-kegiatan/:id` | delete: `daftar_kegiatan` |

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

**`POST /daftar-kegiatan/:dk_id/projects` Request:**
```json
{
  "program_title_id": "uuid",
  "institution_id": "uuid-executing-agency",
  "duration": "2025-2030",
  "objectives": "Meningkatkan konektivitas...",
  "gb_project_ids": ["uuid-gb-project-1"],
  "location_ids": ["uuid-sumut"],
  "financing_details": [
    {
      "lender_id": "uuid-jica",
      "amount_usd": 300000000,
      "grant_usd": 0,
      "counterpart_usd": 50000000,
      "remarks": "Termasuk biaya supervisi"
    }
  ],
  "loan_allocations": [
    {
      "institution_id": "uuid-ditjen-bina-marga",
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

### `GET /projects/:bb_project_id/journey`

**Permission:** Authenticated

Menampilkan seluruh alur proyek dari BB → GB → DK → LA → Monitoring dalam satu response.

**Response `200`:**
```json
{
  "data": {
    "bb_project": {
      "id": "uuid",
      "bb_code": "BB-2025-001",
      "project_name": "Trans Sumatra Toll Road"
    },
    "loi": [
      { "id": "uuid", "lender": { "name": "JICA" }, "tanggal": "2025-03-10" }
    ],
    "gb_projects": [
      {
        "id": "uuid",
        "gb_code": "GB-2025-001",
        "project_name": "Trans Sumatra Section 1",
        "dk_projects": [
          {
            "id": "uuid",
            "daftar_kegiatan": { "subject": "DK TA 2025", "tanggal": "2025-02-01" },
            "loan_agreement": {
              "id": "uuid",
              "loan_code": "IP-603",
              "effective_date": "2025-06-01",
              "is_extended": false,
              "monitoring": [
                {
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
