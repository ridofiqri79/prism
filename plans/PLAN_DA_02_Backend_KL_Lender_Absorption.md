# PLAN DA-02 - Backend Kementerian/Lembaga, Lender, and Absorption Analytics

> **Scope:** Implementasi agregasi utama Dashboard Analytics: Kementerian/Lembaga, lender, performa tahunan, penyerapan, dan proporsi lender.
> **Deliverable:** Endpoint analytics utama mengembalikan data nyata dari DB, tanpa placeholder.
> **Dependencies:** `PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`.

---

## Instruksi Untuk Agent

Baca sebelum mulai:

- `AGENTS.md`
- `docs/PRISM_Business_Rules.md`
- `docs/prism_ddl.sql`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`
- `plans/PLAN_DA_00_Dashboard_Analytics_Roadmap.md`
- `plans/PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`

Aturan:

- Query-first.
- Jangan campur source lender antar stage.
- Default latest snapshot untuk portfolio.
- Penyerapan dihitung server-side dan div-by-zero safe.
- Untuk Kementerian/Lembaga, tampilkan root institution hasil roll-up.

---

## Task 1 - Overview Analytics

Endpoint:

```text
GET /dashboard/analytics/overview
```

Response target:

```json
{
  "data": {
    "portfolio": {
      "project_count": 0,
      "assignment_count": 0,
      "total_pipeline_loan_usd": 0,
      "total_agreement_amount_usd": 0,
      "total_planned_usd": 0,
      "total_realized_usd": 0,
      "absorption_pct": 0
    },
    "pipeline_funnel": [
      { "stage": "BB", "project_count": 0, "total_loan_usd": 0 },
      { "stage": "GB", "project_count": 0, "total_loan_usd": 0 },
      { "stage": "DK", "project_count": 0, "total_loan_usd": 0 },
      { "stage": "LA", "project_count": 0, "total_loan_usd": 0 },
      { "stage": "Monitoring", "project_count": 0, "total_loan_usd": 0 }
    ],
    "top_insights": []
  }
}
```

Notes:

- `project_count` harus distinct logical project.
- `assignment_count` boleh lebih besar dari project count jika satu project punya banyak Kementerian/Lembaga.
- `total_pipeline_loan_usd` mengikuti logic Project Master funding summary.
- `total_agreement_amount_usd` dari `loan_agreement.amount_usd`.

---

## Task 2 - Institution Analytics

Endpoint:

```text
GET /dashboard/analytics/institutions
```

Response target:

```json
{
  "data": {
    "summary": {
      "institution_count": 0,
      "project_count": 0,
      "assignment_count": 0,
      "total_agreement_amount_usd": 0,
      "total_planned_usd": 0,
      "total_realized_usd": 0,
      "absorption_pct": 0
    },
    "items": [
      {
        "institution": {
          "id": "uuid",
          "name": "Kementerian PUPR",
          "short_name": "PUPR",
          "level": "Kementerian/Badan/Lembaga"
        },
        "project_count": 0,
        "assignment_count": 0,
        "loan_agreement_count": 0,
        "monitoring_count": 0,
        "agreement_amount_usd": 0,
        "planned_usd": 0,
        "realized_usd": 0,
        "absorption_pct": 0,
        "pipeline_breakdown": {
          "BB": 0,
          "GB": 0,
          "DK": 0,
          "LA": 0,
          "Monitoring": 0
        },
        "drilldown": {
          "target": "projects",
          "query": { "executing_agency_ids": ["uuid"] }
        }
      }
    ]
  }
}
```

Rules:

- Untuk monitoring performance, institution source utama adalah `dk_project.institution_id`.
- Untuk portfolio distribution, boleh memakai Project Master Executing Agency dari latest BB project.
- Jika memakai dua source berbeda, response harus membedakan `portfolio_*` dan `monitoring_*`.

---

## Task 3 - Lender Analytics

Endpoint:

```text
GET /dashboard/analytics/lenders
```

Response target:

```json
{
  "data": {
    "summary": {
      "lender_count": 0,
      "loan_agreement_count": 0,
      "total_agreement_amount_usd": 0,
      "total_planned_usd": 0,
      "total_realized_usd": 0,
      "absorption_pct": 0
    },
    "items": [
      {
        "lender": {
          "id": "uuid",
          "name": "Japan International Cooperation Agency",
          "short_name": "JICA",
          "type": "Bilateral"
        },
        "loan_agreement_count": 0,
        "project_count": 0,
        "institution_count": 0,
        "agreement_amount_usd": 0,
        "planned_usd": 0,
        "realized_usd": 0,
        "absorption_pct": 0,
        "drilldown": {
          "target": "projects",
          "query": { "fixed_lender_ids": ["uuid"] }
        }
      }
    ]
  }
}
```

Rules:

- Performa lender default memakai `loan_agreement.lender_id`.
- Jangan pakai `lender_indication` untuk performa lender legal binding.
- Funding source Green Book boleh tampil sebagai "Fixed Lender Pipeline" jika diperlukan, bukan sebagai realisasi.

---

## Task 4 - Lender per Kementerian/Lembaga Matrix

Endpoint boleh digabung di `/dashboard/analytics/lenders` atau dibuat sebagai section dalam response.

Response target:

```json
{
  "lender_institution_matrix": [
    {
      "institution": { "id": "uuid", "name": "Kementerian PUPR" },
      "lender": { "id": "uuid", "name": "JICA", "type": "Bilateral" },
      "project_count": 0,
      "loan_agreement_count": 0,
      "agreement_amount_usd": 0,
      "planned_usd": 0,
      "realized_usd": 0,
      "absorption_pct": 0
    }
  ]
}
```

SQL source:

- `loan_agreement -> dk_project -> institution`
- `loan_agreement -> lender`
- `monitoring_disbursement -> loan_agreement`

---

## Task 5 - Absorption Analytics

Endpoint:

```text
GET /dashboard/analytics/absorption
```

Response target:

```json
{
  "data": {
    "summary": {
      "planned_usd": 0,
      "realized_usd": 0,
      "absorption_pct": 0
    },
    "by_institution": [],
    "by_project": [],
    "by_lender": []
  }
}
```

Each item minimal:

```json
{
  "rank": 1,
  "name": "Label",
  "planned_usd": 0,
  "realized_usd": 0,
  "absorption_pct": 0,
  "variance_usd": 0,
  "status": "low|normal|high",
  "drilldown": {}
}
```

Status default:

- `low`: `< 50`
- `normal`: `50 - 89.99`
- `high`: `>= 90`

Jadikan threshold sebagai query param hanya jika dibutuhkan. Default cukup hardcoded di service dan terdokumentasi.

---

## Task 6 - Yearly Performance

Endpoint:

```text
GET /dashboard/analytics/yearly
```

Response target:

```json
{
  "data": {
    "items": [
      {
        "budget_year": 2025,
        "quarter": "TW1",
        "planned_usd": 0,
        "realized_usd": 0,
        "absorption_pct": 0,
        "loan_agreement_count": 0,
        "project_count": 0
      }
    ]
  }
}
```

Rules:

- Jika `quarter` filter kosong, return semua quarter dalam tahun.
- Jika `budget_year` kosong, return trend semua tahun yang ada di monitoring.
- Order by `budget_year ASC`, lalu `TW1`, `TW2`, `TW3`, `TW4`.

---

## Task 7 - Lender Type Proportion

Endpoint:

```text
GET /dashboard/analytics/lender-proportion
```

Response target:

```json
{
  "data": {
    "by_stage": [
      {
        "stage": "Loan Agreement",
        "items": [
          {
            "type": "Bilateral",
            "project_count": 0,
            "lender_count": 0,
            "amount_usd": 0,
            "share_pct": 0
          }
        ]
      }
    ]
  }
}
```

Stage yang disarankan:

- `Lender Indication`
- `Green Book Funding Source`
- `Loan Agreement`
- `Monitoring Realization`

Rules:

- Share percent dihitung terhadap total stage masing-masing.
- KSA harus tampil sebagai tipe mandiri, bukan digabung ke Bilateral.

---

## Task 8 - Backend Tests

Tambahkan test service untuk:

- absorption planned 0 menghasilkan 0,
- KSA tampil di lender type proportion,
- latest snapshot tidak double count,
- lender per Kementerian/Lembaga memakai Loan Agreement untuk performa,
- yearly order benar,
- filter `budget_year`, `quarter`, `lender_types`, `institution_ids` bekerja.

---

## Acceptance Criteria

- Semua endpoint di phase ini mengembalikan data nyata.
- Tidak ada placeholder atau hardcoded sample data.
- Hitungan default tidak double-count revisi.
- KSA muncul sebagai kategori lender type terpisah.
- `absorption_pct` div-by-zero safe.
- `go test ./...` berhasil atau blocker dicatat.

---

## Verification

```powershell
cd prism-backend
make generate
go test ./...
```

Opsional smoke dengan backend running:

```powershell
Invoke-RestMethod http://localhost:8080/api/v1/dashboard/analytics/overview -Headers @{ Authorization = "Bearer <token>" }
```

---

## Checklist

- [x] Overview analytics query/service selesai
- [x] Institution analytics query/service selesai
- [x] Lender analytics query/service selesai
- [x] Lender per Kementerian/Lembaga matrix selesai
- [x] Absorption analytics selesai
- [x] Yearly performance selesai
- [x] Lender type proportion selesai
- [x] DTO response strongly typed
- [x] Tests coverage agregasi utama
- [x] `make generate` berhasil
- [x] `go test ./...` berhasil atau blocker dicatat
