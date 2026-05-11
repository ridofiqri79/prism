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
| `file` | Workbook `.xlsx` berisi sheet `Program Titles`, `Bappenas Partners`, `Institutions`, `Regions`, `Periods`, `National Priorities`, `Lenders`, dan `Kurs Tengah` |

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

Baris yang sudah ada akan di-skip. Untuk sheet `Institutions`, duplikat dicek sesuai scope: top-level berdasarkan nama, child berdasarkan kombinasi parent dan nama. `Parent Name` dapat diisi dengan nama jika unik, UUID institution, atau path `Nama Child; Nama Parent; Nama Root;`. Jika `Parent Name` hanya berisi nama polos dan mengarah ke lebih dari satu institution karena nama child duplikat lintas parent, baris dianggap `failed` agar import tidak memilih parent yang salah. Sheet `Kurs Tengah` memakai kolom `Currency`, `Kurs`, `Kurs Tengah BI`, dan `Cut Off Date`; kombinasi currency dan cut off date yang sudah ada akan di-skip. Sheet `Panduan` pada template menjelaskan fallback referensi Institution: path dropdown sebagai prioritas utama, UUID dari sheet Master Data sebagai fallback paling spesifik, dan nama polos hanya jika unik. Detail baris preview dikembalikan di `sheets[].rows` dengan `status`: `create`, `skip`, atau `failed`, sehingga frontend dapat memberi tab/filter sebelum eksekusi. Baris yang gagal validasi juga dikembalikan di `sheets[].errors`. Frontend wajib meminta preview terlebih dahulu sebelum user menekan eksekusi import.

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

### Kurs Tengah BI

Kurs Tengah BI memakai permission module `currency` karena merupakan turunan master mata uang. Operasi tulis hanya tersedia dalam bentuk bulk.

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/master/kurs-tengah` | read: `currency` |
| `POST` | `/master/kurs-tengah/bulk` | create: `currency` |
| `PUT` | `/master/kurs-tengah/bulk` | update: `currency` |
| `DELETE` | `/master/kurs-tengah/bulk` | delete: `currency` |

**`GET /master/kurs-tengah` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `currency_id` | Filter multi-value UUID Master Currency |
| `cut_off_date_from` | Filter tanggal mulai, format `YYYY-MM-DD` |
| `cut_off_date_to` | Filter tanggal akhir, format `YYYY-MM-DD` |
| `search` | Cari berdasarkan kode atau nama currency |
| `sort` | `currency`, `cut_off_date`, `kurs`, `kurs_tengah_bi` |

**`GET /master/kurs-tengah` Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "currency_id": "uuid-currency-jpy",
      "currency": {
        "id": "uuid-currency-jpy",
        "code": "JPY",
        "name": "Japanese Yen",
        "symbol": "JPY",
        "is_active": true
      },
      "kurs": 105.25,
      "kurs_tengah_bi": 105.18,
      "cut_off_date": "2026-05-07"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 1, "total_pages": 1 }
}
```

**`POST /master/kurs-tengah/bulk` Request:**
```json
{
  "items": [
    {
      "currency_id": "uuid-currency-jpy",
      "kurs": 105.25,
      "kurs_tengah_bi": 105.18,
      "cut_off_date": "2026-05-07"
    }
  ]
}
```

**`PUT /master/kurs-tengah/bulk` Request:**
```json
{
  "items": [
    {
      "id": "uuid-kurs-tengah",
      "currency_id": "uuid-currency-jpy",
      "kurs": 105.25,
      "kurs_tengah_bi": 105.18,
      "cut_off_date": "2026-05-07"
    }
  ]
}
```

Response bulk create/update membungkus array data yang tersimpan dalam `data`. `DELETE /master/kurs-tengah/bulk` menerima body `{"ids":["uuid-kurs-tengah"]}` dan mengembalikan `204`.

Validasi: `currency_id` wajib merujuk Master Currency, `kurs` dan `kurs_tengah_bi` wajib lebih dari 0, `cut_off_date` wajib format `YYYY-MM-DD`, dan kombinasi `(currency_id, cut_off_date)` unik.

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
  "country_id": "uuid"        // wajib jika type Bilateral; opsional untuk KSA; null/kosong untuk Multilateral
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

`DELETE /blue-books/:id` melakukan hard delete hanya jika Blue Book belum memiliki Project Blue Book dan tidak dipakai sebagai sumber revisi Blue Book lain. Jika masih memiliki Project Blue Book, backend mengembalikan `409 CONFLICT` dengan pesan aman agar user menghapus Project Blue Book terlebih dahulu.

**`GET /blue-books` Query Params tambahan:**

| Param | Format | Keterangan |
|-------|--------|------------|
| `search` | string | Cari berdasarkan nama periode, tanggal terbit, tahun revisi, atau status |
| `period_id` | multi-value UUID | Filter periode Blue Book |
| `status` | multi-value enum | `active` = Berlaku, `superseded` = Tidak Berlaku |
| `sort` | enum | `period`, `publish_date`, `revision`, `status`, `project_count`, `created_at` |
| `order` | enum | `asc` atau `desc` |

Response list Blue Book menyertakan `project_count` pada tiap item untuk menentukan apakah tombol hapus boleh ditampilkan.

**`POST /blue-books` Request:**
```json
{
  "period_id": "uuid",
  "replaces_blue_book_id": "uuid-blue-book-sebelumnya",
  "publish_date": "2025-01-15",
  "revision_number": 0,
  "revision_year": null,
  "status": "active"
}
```

Status dikirim eksplisit saat create/update. Backend tidak otomatis mengubah Blue Book lain menjadi `superseded` ketika Blue Book baru dibuat.
Create Blue Book baru selalu dimulai kosong. Jika user ingin membawa Project Blue Book dari dokumen lain, gunakan endpoint import di detail Blue Book melalui tombol `Impor Proyek dari Blue Book Lain`.

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
    "project_count": 0,
    "created_at": "2025-01-15T08:00:00Z",
    "updated_at": "2025-01-15T08:00:00Z"
  }
}
```

---

### Import Blue Book Projects

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/blue-books/import/template` | ADMIN only |
| `POST` | `/blue-books/import/preview` | ADMIN only |
| `POST` | `/blue-books/import/execute` | ADMIN only |
| `GET` | `/blue-books/:bb_id/import-projects/template` | ADMIN only |
| `POST` | `/blue-books/:bb_id/import-projects/preview` | ADMIN only |
| `POST` | `/blue-books/:bb_id/import-projects/execute` | ADMIN only |
| `POST` | `/blue-books/:bb_id/import-projects/from-blue-book` | create: `bb_project` |

**Import dari Blue Book lain:**
```json
{
  "source_blue_book_id": "uuid-blue-book-sumber",
  "project_ids": ["uuid-bb-project-sumber"]
}
```

Endpoint `POST /blue-books/:bb_id/import-projects/from-blue-book` meng-clone Project Blue Book terpilih dari Blue Book sumber ke Blue Book target. Blue Book sumber harus berbeda dari target dan berada pada periode yang sama. Backend menolak import jika salah satu project tidak berasal dari Blue Book sumber, `BB Code` sudah ada di target, atau logical project yang sama sudah ada di target. Relasi anak ikut di-clone sebagai snapshot baru: institution, Mitra Kerja Bappenas, lokasi, national priority, project cost, lender indication, dan LoI.

**Import workbook Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Blue Book`, `Input Data`, `Relasi - EA`, `Relasi - IA`, `Relasi - Locations`, `Relasi - National Priority`, `Relasi - Project Cost`, dan `Relasi - Lender Indication` |

Endpoint collection-level `/blue-books/import/*` mengimpor beberapa header Blue Book dalam satu workbook. Sheet `Blue Book` menjadi daftar target dan memakai `Blue Book Key (*)` sebagai kunci kerja antar sheet. Key ini tidak disimpan di database. Sheet proyek dan relasi memakai kombinasi `Blue Book Key (*)` + `BB Code (*)` sebagai kunci penghubung.

Endpoint legacy `/blue-books/:bb_id/import-projects/*` tetap tersedia untuk import proyek ke satu Blue Book target dari detail Blue Book. Format workbook legacy tidak memakai sheet `Blue Book` dan tidak membuat header Blue Book baru.

**Template:**
`GET /blue-books/import/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Blue Book`, `Input Data`, semua sheet relasi, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot master data saat template dibuat; national priority menampilkan seluruh master data dan tidak dibatasi period Blue Book target.

`GET /blue-books/:bb_id/import-projects/template` mengunduh template legacy untuk satu target Blue Book.

**Kolom utama workbook:**

| Sheet | Kolom |
|-------|-------|
| `Blue Book` | `Blue Book Key (*)`, `Period Name (*)`, `Publish Date (*)`, `Revision Number`, `Revision Year`, `Status (*)`, `Replaces Blue Book Ref` |
| `Input Data` | `Blue Book Key (*)`, `Program Title (*)`, `Bappenas Partners`, `BB Code (*)`, `Project Name (*)`, `Duration`, `Objective`, `Scope of Work`, `Outputs`, `Outcomes` |
| `Relasi - EA` | `Blue Book Key (*)`, `BB Code (*)`, `Executing Agency Name (*)` |
| `Relasi - IA` | `Blue Book Key (*)`, `BB Code (*)`, `Implementing Agency Name (*)` |
| `Relasi - Locations` | `Blue Book Key (*)`, `BB Code (*)`, `Location Name (*)` |
| `Relasi - National Priority` | `Blue Book Key (*)`, `BB Code (*)`, `National Priority Name (*)` |
| `Relasi - Project Cost` | `Blue Book Key (*)`, `BB Code (*)`, `Funding Type (*)`, `Funding Category (*)`, `Amount USD` |
| `Relasi - Lender Indication` | `Blue Book Key (*)`, `BB Code (*)`, `Lender Name (*)`, `Keterangan` |

`Status` menerima `Berlaku` atau `Tidak Berlaku` dan disimpan sebagai `active` atau `superseded`. `Revision Number` kosong dianggap `0`. `Replaces Blue Book Ref` opsional dan dapat berisi `Blue Book Key` dari workbook yang sama atau UUID Blue Book existing; sumber revisi harus Period yang sama. Jika kombinasi `Period Name + Revision Number + Revision Year` sudah ada di database, baris header masuk `skip` hanya jika metadata workbook cocok dengan data existing; jika tidak cocok, baris `failed`.

Kolom `Duration` pada workbook diisi sebagai angka jumlah bulan. Kolom `Bappenas Partners` opsional; isi lebih dari satu mitra dengan pemisah koma atau titik koma. Kolom institution pada `Relasi - EA` dan `Relasi - IA` dapat diisi dengan nama jika unik, UUID dari sheet `Master Data`, atau path `Nama Child; Nama Parent; Nama Root;`. Template dropdown memakai path agar nama child yang sama di parent berbeda tetap bisa dipilih tanpa ambigu. Kolom `Relasi - Lender Indication.Lender Name` di-resolve ke master Lender berdasarkan `name`; jika tidak cocok, import mencoba fallback ke `short_name` yang unik. Sheet `Panduan` menjelaskan fallback ini agar operator tidak perlu menebak ketika nama Institution sama atau workbook memakai singkatan lender seperti `ADB`, `IFAD`, `EIB`, atau `UKEF`.

**Preview:**
`POST /blue-books/import/preview` dan `POST /blue-books/:bb_id/import-projects/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /blue-books/import/execute` dan `POST /blue-books/:bb_id/import-projects/execute` menyimpan data jika hasil pemrosesan tidak memiliki baris gagal.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`. Frontend wajib menampilkan preview dan meminta konfirmasi user sebelum eksekusi.

Baris dengan `BB Code` yang sudah ada dalam Blue Book target terkait akan di-skip. `BB Code` boleh sama pada `Blue Book Key` berbeda, tetapi tidak boleh duplikat dalam `Blue Book Key` yang sama. `BB Code` yang hanya ada pada revisi lama tidak di-skip; jika cocok dengan revisi sumber, snapshot baru memakai `project_identity_id` yang sama. Relasi valid akan dibuat bersama proyek baru; relasi untuk proyek yang di-skip ikut di-skip. National Priority divalidasi terhadap master data tanpa pembatasan period Blue Book target.

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
| `sort` | enum | `bb_code`, `project_name`, `executing_agency`, `location`, `created_at` |
| `order` | enum | `asc` atau `desc` |

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
| `search` | string | Cari berdasarkan tahun terbit, nomor revisi, status teknis, atau label status |
| `publish_year` | multi-value number | Filter tahun terbit Green Book |
| `status` | multi-value enum | `active` = Berlaku, `superseded` = Tidak Berlaku |
| `sort` | enum | `publish_year`, `revision`, `status`, `project_count`, `created_at` |
| `order` | enum | `asc` atau `desc` |

Response list dan detail Green Book menyertakan `project_count` pada tiap item untuk menentukan apakah tombol hapus boleh ditampilkan.

**`POST /green-books` Request:**
```json
{
  "publish_year": 2025,
  "replaces_green_book_id": "uuid-green-book-sebelumnya",
  "revision_number": 0,
  "status": "active"
}
```

Validasi: kombinasi `publish_year` + `revision_number` harus unik. Jika sudah ada, backend mengembalikan `409 CONFLICT`.
Status dikirim eksplisit saat create/update. Backend tidak otomatis mengubah Green Book lain menjadi `superseded` ketika Green Book baru dibuat. Green Book baru selalu dimulai kosong; Project Green Book dari dokumen/revisi lain hanya dibuat jika user membuatnya manual, menjalankan import workbook, atau memakai tombol `Tambahkan Proyek dari Green Book Lain`.

`DELETE /green-books/:id` melakukan hard delete. Backend menolak penghapusan jika Green Book masih memiliki Project Green Book atau masih dipakai sebagai sumber revisi Green Book lain.

---

### Import Green Book Projects

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/green-books/import/template` | ADMIN only |
| `POST` | `/green-books/import/preview` | ADMIN only |
| `POST` | `/green-books/import/execute` | ADMIN only |
| `GET` | `/green-books/:gb_id/import-projects/template` | ADMIN only |
| `POST` | `/green-books/:gb_id/import-projects/preview` | ADMIN only |
| `POST` | `/green-books/:gb_id/import-projects/execute` | ADMIN only |
| `POST` | `/green-books/:gb_id/import-projects/from-green-book` | create: `gb_project` |

**Tambahkan dari Green Book lain:**
```json
{
  "source_green_book_id": "uuid-green-book-sumber",
  "project_ids": ["uuid-gb-project-sumber"]
}
```

Endpoint `POST /green-books/:gb_id/import-projects/from-green-book` meng-clone Project Green Book terpilih dari Green Book sumber ke Green Book target. Green Book sumber harus berbeda dari target, tetapi boleh berasal dari `publish_year` yang berbeda. Backend menolak import jika salah satu project tidak berasal dari Green Book sumber, `GB Code` sudah ada di target, atau logical Green Book project yang sama sudah ada di target. Relasi anak ikut di-clone sebagai snapshot baru: relasi BB Project di-resolve ke latest BB Project snapshot, Mitra Kerja Bappenas, institution, lokasi, activities, funding source, disbursement plan, dan funding allocation.

**Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Green Book`, `Input Data`, `Relasi - BB Project`, `Relasi - EA`, `Relasi - IA`, `Relasi - Locations`, `Relasi - Activities`, `Relasi - Funding Source`, `Relasi - Disbursement Plan`, dan `Relasi - Funding Allocation` |

Endpoint collection-level `/green-books/import/*` mengimpor beberapa header Green Book dalam satu workbook. Sheet `Green Book` menjadi daftar target dan memakai `Green Book Key (*)` sebagai kunci kerja antar sheet. Key ini tidak disimpan di database. Sheet proyek dan relasi memakai kombinasi `Green Book Key (*)` + `GB Code (*)` sebagai kunci penghubung.

Endpoint legacy `/green-books/:gb_id/import-projects/*` tetap tersedia untuk import proyek ke satu Green Book target dari detail Green Book. Format workbook legacy tidak memakai sheet `Green Book` dan tidak membuat header Green Book baru.

**Template:**
`GET /green-books/import/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Green Book`, `Input Data`, semua sheet relasi, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot master data dan BB Project aktif saat template dibuat.

`GET /green-books/:gb_id/import-projects/template` mengunduh template legacy untuk satu target Green Book.

**Kolom utama workbook:**

| Sheet | Kolom |
|-------|-------|
| `Green Book` | `Green Book Key (*)`, `Publish Year (*)`, `Revision Number`, `Status (*)`, `Replaces Green Book Ref` |
| `Input Data` | `Green Book Key (*)`, `Program Title (*)`, `GB Code (*)`, `Project Name (*)`, `Duration`, `Objective`, `Scope of Project` |
| `Relasi - BB Project` | `Green Book Key (*)`, `GB Code (*)`, `BB Code (*)` |
| `Relasi - EA` | `Green Book Key (*)`, `GB Code (*)`, `Executing Agency Name (*)` |
| `Relasi - IA` | `Green Book Key (*)`, `GB Code (*)`, `Implementing Agency Name (*)` |
| `Relasi - Locations` | `Green Book Key (*)`, `GB Code (*)`, `Location Name (*)` |
| `Relasi - Activities` | `Green Book Key (*)`, `GB Code (*)`, `Activity No (*)`, `Activity Name (*)`, `Implementation Location`, `PIU`, `Sort Order` |
| `Relasi - Funding Source` | `Green Book Key (*)`, `GB Code (*)`, `Lender Name (*)`, `Institution Name`, `Currency`, `Loan Original`, `Grant Original`, `Local Original`, `Loan USD`, `Grant USD`, `Local USD` |
| `Relasi - Disbursement Plan` | `Green Book Key (*)`, `GB Code (*)`, `Year (*)`, `Amount USD` |
| `Relasi - Funding Allocation` | `Green Book Key (*)`, `GB Code (*)`, `Activity No (*)`, `Services`, `Constructions`, `Goods`, `Trainings`, `Other` |

`Status` menerima `Berlaku` atau `Tidak Berlaku` dan disimpan sebagai `active` atau `superseded`. `Revision Number` kosong dianggap `0`. `Replaces Green Book Ref` opsional dan dapat berisi `Green Book Key` dari workbook yang sama atau UUID Green Book existing; sumber revisi harus `Publish Year` yang sama. Jika kombinasi `Publish Year + Revision Number` sudah ada di database, baris header masuk `skip` hanya jika metadata workbook cocok dengan data existing; jika tidak cocok, baris `failed`.

Kolom `Duration` pada workbook diisi sebagai angka jumlah bulan. Kolom institution pada `Relasi - EA`, `Relasi - IA`, dan `Relasi - Funding Source` dapat diisi dengan nama jika unik, UUID dari sheet `Master Data`, atau path `Nama Child; Nama Parent; Nama Root;`. Template dropdown memakai path agar nama child yang sama di parent berbeda tetap bisa dipilih tanpa ambigu. Kolom `Relasi - Funding Source.Lender Name` di-resolve ke master Lender berdasarkan `name`; jika tidak cocok, import mencoba fallback ke `short_name` yang unik. Sheet `Panduan` menjelaskan fallback ini agar operator tidak perlu menebak ketika nama Institution sama atau workbook memakai singkatan lender seperti `ADB`, `IFAD`, `EIB`, atau `UKEF`.

**Preview:**
`POST /green-books/import/preview` dan `POST /green-books/:gb_id/import-projects/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /green-books/import/execute` dan `POST /green-books/:gb_id/import-projects/execute` menyimpan data jika hasil pemrosesan tidak memiliki baris gagal.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`. Frontend wajib menampilkan preview dan meminta konfirmasi user sebelum eksekusi.

Baris dengan `GB Code` yang sudah ada dalam Green Book target terkait akan di-skip. `GB Code` boleh sama pada `Green Book Key` berbeda, tetapi tidak boleh duplikat dalam `Green Book Key` yang sama. `GB Code` yang hanya ada pada revisi lama tidak di-skip; jika cocok dengan revisi sumber, snapshot baru memakai `gb_project_identity_id` yang sama. Relasi BB Project di-resolve ke latest BB Project snapshot saat import dieksekusi. Proyek baru wajib memiliki minimal satu BB Project, EA, IA, dan lokasi. `Currency` kosong dianggap `USD`; jika `USD`, nilai USD disamakan dengan nilai original sehingga user tidak perlu mengisi dua kali. `Year` pada Disbursement Plan harus unik per kombinasi `Green Book Key + GB Code`. Funding Allocation mengacu ke `Activity No`; activity tanpa Funding Allocation eksplisit tetap dibuat dengan allocation bernilai 0.

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
| `sort` | enum | `gb_code`, `project_name`, `bb_projects`, `status`, `created_at` |
| `order` | enum | `asc` atau `desc` |

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
    "program_title_id": "uuid-program-title",
    "program_title": { "id": "uuid-program-title", "title": "Infrastruktur Transportasi" },
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
`GET /daftar-kegiatan/import/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Daftar Kegiatan`, `Input Data`, semua sheet relasi, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot master data, Master Lender, dan GB Project aktif saat template dibuat.

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

Jika `Letter Number` sudah ada di DB, header dan semua project/relasi di bawahnya berstatus `skip`. Duplikat `Letter Number` dalam workbook berstatus `failed`. Project baru wajib punya Project Name, Executing Agency, minimal 1 GB Project aktif, Location, Financing Detail, Loan Allocation, dan Activity Detail. `Project Name` adalah nama snapshot di Daftar Kegiatan dan boleh berbeda dari nama Green Book. `Program Title` opsional, tetapi jika diisi harus ada di master data. Kolom institution pada `Input Data.Executing Agency Name` dan `Relasi - Loan Allocation.Institution Name` dapat diisi dengan nama jika unik, UUID dari sheet `Master Data`, atau path `Nama Child; Nama Parent; Nama Root;`. Sheet `Panduan` menjelaskan fallback ini dan Preview tetap gagal untuk nama polos yang ambigu. Lender Financing Detail harus ada di Master Lender dan boleh berbeda dari funding source Green Book terkait. `Currency` kosong dianggap `USD`; jika diisi harus kode ISO 4217 yang aktif di Master Currency. Amount kosong dianggap `0` dan tidak boleh negatif. `Activity No` duplikat per project berstatus `failed`.
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
| `sort` | enum | `subject`, `date`, `letter_number`, `project_count`, `created_at` |
| `order` | enum | `asc` atau `desc` |

Response list dan detail Daftar Kegiatan menyertakan `project_count` untuk menentukan apakah tombol hapus boleh ditampilkan.

`DELETE /daftar-kegiatan/:id` melakukan hard delete hanya jika Daftar Kegiatan belum memiliki Project di Daftar Kegiatan. Jika masih memiliki proyek, backend mengembalikan `409 CONFLICT` dengan pesan aman agar user menghapus Project Daftar Kegiatan terlebih dahulu.

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
| `sort` | enum | `project_name`, `executing_agency`, `duration`, `created_at` |
| `order` | enum | `asc` atau `desc` |

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

Response DK Project pada `GET /daftar-kegiatan/:dk_id/projects` dan `GET /daftar-kegiatan/:dk_id/projects/:id` juga menyertakan `loan_agreements` sebagai array Loan Agreement yang sudah dibuat untuk proyek tersebut. Jika belum ada, array kosong dikirim.

```json
{
  "id": "uuid-dk-project",
  "project_name": "Trans Sumatra Section 1 - DK",
  "loan_agreements": [
    {
      "id": "uuid-loan-agreement",
      "loan_code": "IP-603"
    },
    {
      "id": "uuid-loan-agreement-2",
      "loan_code": "IP-604"
    }
  ]
}
```

Frontend dapat membuka form `Buat Loan Agreement` dari setiap proyek pada detail Header Daftar Kegiatan dengan query `dk_id` dan `dk_project_id`. Form Loan Agreement menggunakan query tersebut untuk preselect satu Proyek Daftar Kegiatan awal, lalu user tetap dapat menambahkan project lain dan mengatur alokasi. Tombol dibuat aktif hanya jika user memiliki permission `create: loan_agreement` dan proyek DK memiliki lender pada `financing_details`. Loan Agreement yang sudah ada ditampilkan sebagai daftar aksi `Buka Loan Agreement` dengan nilai alokasi project terkait.

---

## Loan Agreement

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/loan-agreements` | read: `loan_agreement` |
| `GET` | `/loan-agreements/:id` | read: `loan_agreement` |
| `POST` | `/loan-agreements` | create: `loan_agreement` |
| `PUT` | `/loan-agreements/:id` | update: `loan_agreement` |
| `DELETE` | `/loan-agreements/:id` | delete: `loan_agreement` |
| `GET` | `/loan-agreements/import/template` | ADMIN only |
| `POST` | `/loan-agreements/import/preview` | ADMIN only |
| `POST` | `/loan-agreements/import/execute` | ADMIN only |

### Import Loan Agreement

**Content-Type:** `multipart/form-data`

**Form field:**

| Field | Keterangan |
|-------|------------|
| `file` | Workbook `.xlsx` berisi sheet `Loan Agreement` dan `Relasi - DK Project` |

Import Loan Agreement bersifat **create-only**. Endpoint ini hanya membuat Loan Agreement baru. Satu `Loan Code` dapat memiliki beberapa baris relasi DK Project pada sheet `Relasi - DK Project`, termasuk relasi lintas header Daftar Kegiatan. `Loan Code` yang sudah dipakai oleh record lain masuk status `failed`.

**Template:**
`GET /loan-agreements/import/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Loan Agreement`, `Relasi - DK Project`, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot DK Project, allowed lender dari `dk_financing_detail`, lender, currency aktif, dan Kurs Tengah BI saat template dibuat.

**Kolom workbook:**

| Sheet | Kolom |
|-------|-------|
| `Loan Agreement` | `Lender Name (*)`, `Loan Code (*)`, `Agreement Date (*)`, `Effective Date (*)`, `Original Closing Date`, `Closing Date (*)`, `Currency (*)`, `Amount Original (*)`, `Cumulative Disbursement` |
| `Relasi - DK Project` | `Loan Code (*)`, `DK Project Ref (*)`, `Allocation Original (*)` |

**Preview:**
`POST /loan-agreements/import/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

**Execute:**
`POST /loan-agreements/import/execute` menjalankan import hanya jika hasil validasi memiliki `total_failed = 0`. Jika masih ada failed, response error validasi dan data tidak disimpan.

**Response `200`:**
Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`.

`DK Project Ref` dapat diisi dari dropdown template atau UUID DK Project. `Lender Name` di-resolve dari master Lender berdasarkan `name`, lalu fallback ke `short_name` unik. Lender wajib berasal dari Financing Detail semua DK Project terkait. `Currency` wajib kode ISO 4217 aktif di Master Currency. `Original Closing Date` opsional dan diisi hanya jika pinjaman diperpanjang. Jika diisi, `Closing Date` tidak boleh lebih awal dari `Original Closing Date`. `Amount Original` wajib lebih dari `0`. Total `Allocation Original` pada sheet relasi untuk satu `Loan Code` wajib sama dengan `Amount Original`. Backend menghitung `amount_usd` dari Kurs Tengah BI terbaru untuk currency non-USD; jika Kurs Tengah BI untuk currency tersebut belum tersedia, preview/import berstatus `failed`. Jika `Currency` adalah `USD`, `amount_usd` disamakan dengan `Amount Original`. `allocation_usd` per project dihitung dari alokasi original dan disesuaikan rounding-nya agar total sama dengan `amount_usd`. `Cumulative Disbursement` opsional, tidak boleh negatif, dan memakai currency Loan Agreement yang dipilih. Workbook lama yang masih mengirim `DK Project Ref` di sheet `Loan Agreement` tetap diterima sebagai fallback single-project.

**`POST /loan-agreements` Request:**
`original_closing_date` boleh dikosongkan/diomit untuk pinjaman yang belum diperpanjang. `is_extended=false` dan `extension_days=0` saat field ini kosong.
```json
{
  "dk_project_allocations": [
    { "dk_project_id": "uuid-dk-project-1", "allocation_original": 30000000000 },
    { "dk_project_id": "uuid-dk-project-2", "allocation_original": 15000000000 }
  ],
  "lender_id": "uuid-jica",
  "loan_code": "IP-603",
  "agreement_date": "2025-03-15",
  "effective_date": "2025-06-01",
  "original_closing_date": "2030-12-31",
  "closing_date": "2030-12-31",
  "currency": "JPY",
  "amount_original": 45000000000,
  "cumulative_disbursement": 12500000000
}
```

`dk_project_allocations` wajib untuk kontrak baru. Field lama `dk_project_id` masih diterima sebagai fallback single-project dan backend mengisi alokasi sebesar `amount_original`. Backend menolak project kosong, project duplikat, project yang tidak ditemukan, lender yang tidak ada di Financing Detail salah satu project, dan total alokasi yang tidak sama dengan `amount_original`. `amount_usd` pada request diabaikan untuk perhitungan baru dan dipertahankan hanya untuk kompatibilitas client lama; backend selalu menghitung nilai USD dari `amount_original` dan Kurs Tengah BI terbaru.

**`GET /loan-agreements/:id` Response `200`:**
```json
{
  "data": {
    "id": "uuid",
    "loan_code": "IP-603",
    "dk_projects": [
      {
        "id": "uuid-dk-project-1",
        "dk_id": "uuid-dk-header-1",
        "project_name": "Trans Sumatra Section 1 - DK",
        "objectives": "Meningkatkan konektivitas",
        "gb_codes": "GB-2025-001",
        "daftar_kegiatan": {
          "id": "uuid-dk-header-1",
          "subject": "DK TA 2025",
          "date": "2025-02-01",
          "letter_number": "B-001/2025"
        },
        "allocation_original": 30000000000,
        "allocation_usd": 200000000
      },
      {
        "id": "uuid-dk-project-2",
        "dk_id": "uuid-dk-header-2",
        "project_name": "Trans Sumatra Section 2 - DK",
        "objectives": "Meningkatkan konektivitas",
        "gb_codes": "GB-2025-002",
        "daftar_kegiatan": {
          "id": "uuid-dk-header-2",
          "subject": "DK TA 2026",
          "date": "2026-02-01",
          "letter_number": "B-002/2026"
        },
        "allocation_original": 15000000000,
        "allocation_usd": 100000000
      }
    ],
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
    "cumulative_disbursement": 12500000000,
    "cumulative_disbursement_usd": 83333333.33,
    "disbursement_ratio": 27.78,
    "estimated_time_ratio": 41.67,
    "performance_value": 0.67,
    "performance_status": "Behind Schedule",
    "kurs_tengah_bi": 150,
    "kurs_cut_off_date": "2026-05-07",
    "created_at": "2025-03-15T08:00:00Z",
    "updated_at": "2025-03-15T08:00:00Z"
  }
}
```

`amount_usd` dan `cumulative_disbursement_usd` dihitung dari Kurs Tengah BI terbaru untuk currency non-USD. `amount_original`, `amount_usd`, `cumulative_disbursement`, dan metrik kinerja adalah nilai global Loan Agreement. `dk_projects[].allocation_original` dan `dk_projects[].allocation_usd` adalah nilai komitmen per DK Project. `cumulative_disbursement` tetap nilai input manual dalam `currency` Loan Agreement yang dipilih dan tidak dialokasikan per project. `disbursement_ratio`, `estimated_time_ratio`, `performance_value`, dan `performance_status` adalah field tampilan dan tidak disimpan sebagai kolom DB.

**`GET /loan-agreements` Query Params tambahan:**

| Param | Keterangan |
|-------|-----------|
| `search` | Cari berdasarkan `loan_code`, nama lender, atau short name lender |
| `lender_id` | Filter by lender |
| `is_extended` | Filter: `true` / `false` |
| `closing_date_before` | Filter LA yang akan berakhir sebelum tanggal ini |
| `sort` | `loan_code`, `lender`, `effective_date`, `closing_date`, `currency`, `amount_usd`, `cumulative_disbursement_usd`, `disbursement_ratio`, `estimated_time_ratio`, `performance_value`, `performance_status`, `status`, `created_at` |
| `order` | `asc` atau `desc` |

---

## Monitoring Disbursement

> Status aktif: dinonaktifkan. Menu, route frontend, import flow, dan endpoint
> CRUD Monitoring Disbursement (`/monitoring/*` dan
> `/loan-agreements/:la_id/monitoring*`) tidak lagi diregistrasikan di aplikasi.
> Bagian di bawah dipertahankan sebagai catatan kontrak historis sampai dokumen
> baseline dipangkas menyeluruh.

### Monitoring Workspace & Import

| Method | Endpoint | Permission |
|--------|----------|-----------|
| `GET` | `/monitoring/loan-agreements` | read: `monitoring_disbursement` |
| `GET` | `/monitoring/import/template` | ADMIN only |
| `POST` | `/monitoring/import/preview` | ADMIN only |
| `POST` | `/monitoring/import/execute` | ADMIN only |

`GET /monitoring/loan-agreements` dipakai halaman Monitoring Disbursement sebagai daftar Loan Agreement yang dapat dibuka ke monitoring triwulan. Query params mengikuti pola list umum: `page`, `limit`, `search`, dan `is_effective`. Search mencakup `loan_code`, lender, nomor surat Daftar Kegiatan, dan nama proyek Daftar Kegiatan.

Import Monitoring Disbursement bersifat **create-only**. Endpoint ini hanya membuat monitoring baru untuk kombinasi `Loan Agreement + Budget Year + Quarter` yang belum ada. Kombinasi periode yang sudah ada masuk status `skip`; duplikat periode baru dalam workbook masuk status `failed`.

`GET /monitoring/import/template` mengunduh workbook `.xlsx` dengan sheet `Panduan`, `Master Data`, `Monitoring Disbursement`, `Relasi - Komponen`, dan sheet `_Dropdowns` tersembunyi. Sheet `Master Data` berisi snapshot Loan Agreement, lender, currency, status efektif, dan periode monitoring yang sudah ada saat template dibuat.

Workbook:
- `Monitoring Disbursement`: `Loan Agreement Ref (*)`, `Budget Year (*)`, `Quarter (*)`, `Exchange Rate USD/IDR (*)`, `Exchange Rate Loan Agreement/IDR (*)`, `Planned Loan Agreement`, `Planned USD`, `Planned IDR`, `Realized Loan Agreement`, `Realized USD`, `Realized IDR`.
- `Relasi - Komponen`: `Loan Agreement Ref (*)`, `Budget Year (*)`, `Quarter (*)`, `Component Name (*)`, `Planned Loan Agreement`, `Planned USD`, `Planned IDR`, `Realized Loan Agreement`, `Realized USD`, `Realized IDR`.

Validasi import:
- Loan Agreement harus ada di snapshot template dan `effective_date <= CURRENT_DATE`.
- `Quarter` wajib `TW1`, `TW2`, `TW3`, atau `TW4`.
- Kurs USD/IDR dan kurs Loan Agreement/IDR wajib lebih dari 0.
- Nilai rencana/realisasi boleh kosong dan dianggap 0; jika diisi tidak boleh negatif.
- Baris komponen hanya boleh mengacu ke monitoring yang dibuat dari sheet `Monitoring Disbursement` pada workbook yang sama. Import tidak menambah/mengubah komponen untuk monitoring yang sudah ada.

`POST /monitoring/import/preview` membaca workbook dan menjalankan validasi dalam transaksi yang di-rollback. Tidak ada data tersimpan.

`POST /monitoring/import/execute` menjalankan import hanya jika hasil validasi memiliki `total_failed = 0`. Jika masih ada failed, response error validasi dan data tidak disimpan.

Format response sama dengan Import Data Master: `data.file_name`, `total_inserted`, `total_skipped`, `total_failed`, dan `sheets[].rows[]` dengan status `create`, `skip`, atau `failed`.

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

## Spatial Distribution

Endpoint ini dipakai menu frontend `/spatial-distribution` dengan label sidebar **Sebaran Wilayah**. Route frontend tetap berbahasa Inggris, sedangkan label UI/menu memakai Bahasa Indonesia.

### `GET /spatial-distribution/choropleth`

**Permission:** read: `bb_project`

Mengembalikan data peta choropleth provinsi atau kabupaten/kota. Cakupan lokasi mengikuti aturan wilayah PRISM:
- lokasi `COUNTRY`/Nasional dihitung mencakup seluruh provinsi pada peta level provinsi,
- lokasi `PROVINCE` dihitung pada provinsi terkait,
- lokasi `CITY` dihitung pada kabupaten/kota terkait dan provinsi induknya.
- saat `level=city`, peta hanya menghitung lokasi `CITY` eksplisit di dalam `province_code`; lokasi `COUNTRY`/`PROVINCE` tidak digandakan ke semua kabupaten/kota karena tidak ada basis alokasi ke child wilayah.

**Query Params:**

| Param | Keterangan |
|-------|-----------|
| `level` | `province` atau `city`; default `province` |
| `province_code` | Wajib jika `level=city`; kode provinsi parent |
| `loan_types` | Multi value: `Bilateral`, `Multilateral`, `KSA` |
| `project_statuses` | Multi value: `Pipeline`, `Ongoing` |
| `pipeline_statuses` | Multi value: `BB`, `GB`, `DK`, `LA`, `Monitoring` |
| `reached_stages` | Multi value tahap yang sudah dicapai. `GB` mencakup project dengan status `GB`, `DK`, `LA`, atau `Monitoring`; `DK` mencakup `DK`, `LA`, atau `Monitoring`; `LA` mencakup `LA` atau `Monitoring`; `Monitoring` hanya `Monitoring`; `BB` mencakup seluruh project aktif |
| `missing_stages` | Multi value tahap yang belum dicapai untuk analisa bottleneck. Contoh: `missing_stages=GB` hanya project yang masih di `BB`; `missing_stages=LA` mencakup project `BB`, `GB`, atau `DK` |
| `has_loi` | Boolean `true`/`false`; filter keberadaan LoI pada BB Project |
| `has_lender_indication` | Boolean `true`/`false`; filter keberadaan indikasi lender pada BB Project |
| `search` | Cari berdasarkan kode atau nama proyek |
| `include_history` | `true` untuk menghitung snapshot historis, default latest snapshot saja |

**Response `200`:**
```json
{
  "data": {
    "level": "province",
    "regions": [
      {
        "region_id": "uuid",
        "region_code": "32",
        "region_name": "Jawa Barat",
        "region_type": "PROVINCE",
        "project_count": 12,
        "total_loan_usd": 2400000000
      }
    ],
    "summary": {
      "total_regions": 38,
      "active_regions": 12,
      "total_project_count": 52,
      "total_loan_usd": 9000000000,
      "max_project_count": 12,
      "max_loan_usd": 2400000000
    }
  }
}
```

---

### `GET /spatial-distribution/projects`

**Permission:** read: `bb_project`

Mengembalikan daftar proyek untuk wilayah yang menjadi fokus pada peta. Query filter sama dengan endpoint choropleth, ditambah pagination standar.
Untuk fokus nasional/Indonesia, frontend mengirim `level=province&region_code=ID`; daftar proyek hanya berisi proyek yang lokasi tersimpannya `COUNTRY`/Nasional, bukan seluruh proyek provinsi turunannya.
Untuk `level=city`, daftar proyek mengikuti angka peta dan hanya memakai lokasi `CITY` eksplisit pada `region_code` yang dipilih.

**Query Params:**

| Param | Keterangan |
|-------|-----------|
| `level` | `province` atau `city`; `province` menerima `COUNTRY` nasional atau `PROVINCE`, `city` menerima `CITY` |
| `region_code` | Kode nasional (`ID`), provinsi, atau kabupaten/kota yang dipilih |
| `province_code` | Konteks parent saat `level=city` |
| `page`, `limit`, `sort`, `order` | Pagination dan sorting standar project master |
| `loan_types`, `project_statuses`, `pipeline_statuses`, `reached_stages`, `missing_stages`, `has_loi`, `has_lender_indication`, `search`, `include_history` | Sama dengan choropleth |

**Response `200`:**
```json
{
  "level": "province",
  "region_id": "uuid",
  "region_code": "32",
  "region_name": "Jawa Barat",
  "region_type": "PROVINCE",
  "data": [
    {
      "id": "uuid",
      "bb_code": "BB-2025-001",
      "project_name": "Pembangunan Infrastruktur",
      "pipeline_status": "GB",
      "foreign_loan_usd": 250000000
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 12, "total_pages": 2 },
  "summary": {
    "total_loan_usd": 2500000000,
    "total_grant_usd": 0,
    "total_counterpart_usd": 300000000
  }
}
```

---

### `GET /dashboard/stage-overview`

**Permission:** read: `bb_project`

Mengembalikan ringkasan tahap untuk funnel dashboard. Hitungan memakai entitas pada tahap masing-masing: Project Blue Book terbaru per `project_identity_id`, Project Green Book, Project Daftar Kegiatan, dan Loan Agreement. Sebaran `regions[]` memakai `region.region_group`: lokasi `CITY` dinaikkan ke provinsi induknya lalu dihitung satu kali per proyek per grup, lokasi `PROVINCE` masuk ke `region_group`, dan lokasi nasional tampil sebagai `Indonesia`.

**Query Params:**

| Param | Keterangan |
|-------|------------|
| `period_ids` | Multi-value UUID periode Blue Book untuk membatasi seluruh tahap berdasarkan relasi BB |
| `region_ids` | Multi-value UUID wilayah/lokasi untuk membatasi ringkasan tahap pada wilayah terpilih |

**Response `200`:**
```json
{
  "data": {
    "stages": [
      {
        "stage": "GB",
        "project_count": 36,
        "total_loan_usd": 1970815490,
        "regions": [
          {
            "label": "Jawa",
            "level": "Region Group",
            "project_count": 18,
            "foreign_loan_usd": 900000000
          }
        ]
      }
    ]
  }
}
```

---

### `GET /dashboard/blue-book-distribution`

**Permission:** read: `bb_project`

Mengembalikan agregasi distribusi Blue Book untuk panel dashboard. Endpoint hanya memakai snapshot Blue Book terbaru per `project_identity_id` agar revisi lama tidak double-count. Relasi K/L dihitung dari role `Executing Agency`, lalu setiap institution dinaikkan ke ancestor level tertinggi (`parent_id IS NULL`). Distribusi Program memakai `program_title_id` exact pada Project Blue Book agar item dashboard dapat deep-link ke Project Master dengan `program_title_ids`.

**Response `200`:**
```json
{
  "data": {
    "agency_groups": [
      {
        "label": "Kementerian/Lembaga",
        "level": "Kementerian/Lembaga",
        "project_count": 85,
        "foreign_loan_usd": 25550770000
      }
    ],
    "top_agencies": [
      {
        "id": "uuid-root-institution",
        "label": "Kemen PU",
        "level": "Kementerian/Badan/Lembaga",
        "project_count": 27,
        "foreign_loan_usd": 12055300000
      }
    ],
    "programs": [
      {
        "id": "uuid-program-title",
        "label": "Konektivitas & Transportasi",
        "level": "Program Title",
        "project_count": 38,
        "foreign_loan_usd": 12100000000
      }
    ]
  }
}
```

Catatan: satu proyek dapat memiliki lebih dari satu Executing Agency pada root berbeda. Karena itu total `project_count` pada distribusi K/L dapat lebih besar dari total proyek Blue Book.

---

### `GET /dashboard/green-book-distribution`

**Permission:** read: `bb_project`

Mengembalikan agregasi distribusi Green Book untuk panel dashboard. Endpoint hanya memakai snapshot Blue Book terbaru per `project_identity_id` yang sudah memiliki relasi Green Book. Tipe lender dan top lender dihitung dari `gb_funding_source`; top K/L dihitung dari role `Executing Agency` pada Project Blue Book lalu dinaikkan ke institution level tertinggi agar deep link memakai `executing_agency_ids` exact.

**Response `200`:**
```json
{
  "data": {
    "lender_types": [
      {
        "label": "Bilateral",
        "level": "Lender Type",
        "project_count": 16,
        "foreign_loan_usd": 8700000000
      }
    ],
    "top_lenders": [
      {
        "id": "uuid-lender",
        "label": "JICA",
        "level": "Bilateral",
        "project_count": 10,
        "foreign_loan_usd": 4100000000
      }
    ],
    "top_agencies": [
      {
        "id": "uuid-root-institution",
        "label": "Kemen PU",
        "level": "Kementerian/Badan/Lembaga",
        "project_count": 11,
        "foreign_loan_usd": 5100000000
      }
    ]
  }
}
```

Contoh deep link dari item endpoint ini:
- `/projects?reached_stages=GB&loan_types=Bilateral`
- `/projects?reached_stages=GB&fixed_lender_ids=<uuid-lender>`
- `/projects?reached_stages=GB&executing_agency_ids=<uuid-root-institution>`

Catatan: satu proyek Green Book dapat memiliki lebih dari satu lender atau Executing Agency root. Karena itu total `project_count` per distribusi dapat lebih besar dari total proyek Green Book.

---

### `GET /dashboard/daftar-kegiatan-distribution`

**Permission:** read: `bb_project`

Mengembalikan agregasi distribusi Daftar Kegiatan untuk panel dashboard. Endpoint hanya memakai snapshot Blue Book terbaru per `project_identity_id` yang sudah mencapai Daftar Kegiatan. Tipe lender dan top lender dihitung dari `dk_financing_detail`; top K/L dihitung dari `dk_project.institution_id` lalu dinaikkan ke institution level tertinggi agar deep link memakai `dk_executing_agency_ids` exact.

Response shape sama dengan `GET /dashboard/green-book-distribution`:
- `lender_types[]`
- `top_lenders[]`
- `top_agencies[]`
- `programs[]`

Contoh deep link dari item endpoint ini:
- `/projects?reached_stages=DK&loan_types=Bilateral`
- `/projects?reached_stages=DK&dk_lender_ids=<uuid-lender>`
- `/projects?reached_stages=DK&dk_executing_agency_ids=<uuid-root-institution>`
- `/projects?reached_stages=DK&program_title_ids=<uuid-program-title>`

Catatan: satu proyek Daftar Kegiatan dapat memiliki lebih dari satu lender. Karena itu total `project_count` per distribusi dapat lebih besar dari total proyek Daftar Kegiatan.

---

### `GET /dashboard/loan-agreement-distribution`

**Permission:** read: `bb_project`

Mengembalikan agregasi distribusi Loan Agreement untuk panel dashboard. Endpoint hanya memakai snapshot Blue Book terbaru per `project_identity_id` yang sudah memiliki Loan Agreement. Tipe lender dan top lender dihitung dari `loan_agreement.lender_id`; top K/L tetap mengikuti Executing Agency pada `dk_project.institution_id` lalu dinaikkan ke institution level tertinggi agar deep link memakai `dk_executing_agency_ids` exact.

Response shape sama dengan `GET /dashboard/green-book-distribution`:
- `lender_types[]`
- `top_lenders[]`
- `top_agencies[]`
- `programs[]`

Contoh deep link dari item endpoint ini:
- `/projects?reached_stages=LA&loan_types=Bilateral`
- `/projects?reached_stages=LA&loan_agreement_lender_ids=<uuid-lender>`
- `/projects?reached_stages=LA&dk_executing_agency_ids=<uuid-root-institution>`
- `/projects?reached_stages=LA&program_title_ids=<uuid-program-title>`

Catatan: satu proyek Daftar Kegiatan dapat memiliki lebih dari satu Loan Agreement, dan satu Loan Agreement dapat dialokasikan ke beberapa proyek Daftar Kegiatan. Agregasi nilai per project memakai `loan_agreement_dk_project.allocation_usd`; agregasi nilai global LA memakai `COUNT(DISTINCT loan_agreement.id)` atau nilai LA distinct agar tidak double count.

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
| `period_ids` | Multi value UUID periode Blue Book |
| `loan_types` | Multi value: `Bilateral`, `Multilateral`, `KSA` |
| `indication_lender_ids` | Multi value UUID lender dari `lender_indication` BB |
| `executing_agency_ids` | Multi value UUID institution role `Executing Agency`. Jika UUID adalah institution level tertinggi, filter juga mencakup EA turunan di bawah root tersebut |
| `fixed_lender_ids` | Multi value UUID lender dari `gb_funding_source` Green Book |
| `dk_lender_ids` | Multi value UUID lender dari `dk_financing_detail` Daftar Kegiatan |
| `loan_agreement_lender_ids` | Multi value UUID lender dari `loan_agreement.lender_id` |
| `dk_executing_agency_ids` | Multi value UUID institution dari `dk_project.institution_id`. Jika UUID adalah institution level tertinggi, filter juga mencakup EA DK turunan di bawah root tersebut |
| `project_statuses` | Multi value: `Pipeline`, `Ongoing` |
| `pipeline_statuses` | Multi value: `BB`, `GB`, `DK`, `LA`, `Monitoring` |
| `reached_stages` | Multi value tahap yang sudah dicapai. `GB` mencakup project dengan status `GB`, `DK`, `LA`, atau `Monitoring`; `DK` mencakup `DK`, `LA`, atau `Monitoring`; `LA` mencakup `LA` atau `Monitoring`; `Monitoring` hanya `Monitoring`; `BB` mencakup seluruh project aktif |
| `missing_stages` | Multi value tahap yang belum dicapai untuk analisa bottleneck. Contoh: `missing_stages=GB` hanya project yang masih di `BB`; `missing_stages=LA` mencakup project `BB`, `GB`, atau `DK` |
| `has_loi` | Boolean `true`/`false`; filter keberadaan LoI pada BB Project |
| `has_lender_indication` | Boolean `true`/`false`; filter keberadaan indikasi lender pada BB Project |
| `program_title_ids` | Multi value UUID program title |
| `region_ids` | Multi value UUID region/location |
| `foreign_loan_min`, `foreign_loan_max` | Range nilai pinjaman foreign loan dalam USD |
| `dk_date_from`, `dk_date_to` | Range tanggal DK, format `YYYY-MM-DD` |
| `search` | Search global untuk kode/nama proyek, indikasi lender, fixed lender Green Book, dan executing agency |
| `include_history` | `true` untuk menampilkan semua snapshot, default `false` |

Multi value dapat dikirim sebagai repeated query param (`loan_types=Bilateral&loan_types=KSA`), comma-separated value, atau format array query string (`loan_types[]=Bilateral`).

Contoh deep link dari dashboard:
- `/projects?reached_stages=GB`
- `/projects?reached_stages=GB&fixed_lender_ids=<uuid>`
- `/projects?reached_stages=GB&executing_agency_ids=<uuid>`
- `/projects?reached_stages=DK&dk_lender_ids=<uuid>`
- `/projects?reached_stages=DK&dk_executing_agency_ids=<uuid>`
- `/projects?reached_stages=LA&loan_agreement_lender_ids=<uuid>`
- `/projects?reached_stages=LA&dk_executing_agency_ids=<uuid>`
- `/projects?reached_stages=BB&missing_stages=GB&has_loi=true`
- `/projects?reached_stages=BB&missing_stages=GB&has_lender_indication=false`
- `/projects?reached_stages=BB&program_title_ids=<uuid>`
- `/spatial-distribution?level=province&region_code=ID&reached_stages=BB&metric=count`

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

Mengunduh workbook Excel (`.xlsx`) berisi seluruh project yang cocok dengan filter aktif. Endpoint memakai query param yang sama dengan `GET /projects`, termasuk `reached_stages`, `missing_stages`, `has_loi`, `has_lender_indication`, `dk_lender_ids`, `loan_agreement_lender_ids`, dan `dk_executing_agency_ids`; `page` dan `limit` diabaikan karena export selalu mengambil semua hasil filter. Sorting tetap mengikuti `sort` dan `order`. Sheet `Ringkasan` berisi total pendanaan serta daftar filter aktif yang dipakai saat export.

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
        "latest_loan_usd": 300000000,
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
            "loan_agreements": [
              {
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
                "allocation_original": 300000000,
                "allocation_usd": 300000000,
                "cumulative_disbursement": 125000000,
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
            ]
          }
        ]
      }
    ]
  }
}
```

`latest_loan_usd` dihitung dari total `gb_funding_source.loan_usd` pada snapshot Green Book terbaru untuk `gb_project_identity_id` yang sama. Field ini dipakai untuk nilai visual terbaru tanpa mengubah path historis `funding_sources`, DK, dan LA yang tetap menunjuk concrete snapshot.

---

## Catatan Implementasi

- Semua timestamp menggunakan **ISO 8601** dengan timezone (`2025-01-15T08:00:00Z`).
- Field `id` selalu **UUID v4** — tidak ada integer ID yang diekspos ke client.
- Endpoint yang mengembalikan data hierarki (seperti `/journey`) tidak menggunakan pagination — data di-fetch sekaligus karena jumlahnya terbatas per proyek.
- `penyerapan_pct` (persentase penyerapan) dihitung di server: `(realisasi / rencana) * 100`, tidak disimpan di database.
- Untuk endpoint list yang memiliki banyak relasi (seperti GB Project), response default hanya memuat field ringkas. Gunakan query param `?expand=true` untuk mendapatkan nested object lengkap.
- Perubahan permission user via `PUT /users/:id/permissions` bersifat **transaksional** — semua permission diupdate dalam satu transaksi, gagal sebagian berarti tidak ada yang tersimpan.
