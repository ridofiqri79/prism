# PLAN DASH 01 - Ringkasan Eksekutif

## Objective

Membangun dashboard pimpinan untuk melihat posisi portofolio nasional: pipeline, komitmen legal, bottleneck utama, top K/L, top lender, dan proyek berisiko.

## Endpoint

```http
GET /api/v1/dashboard/executive-portfolio
```

## Query Params

| Param | Type | Keterangan |
|---|---|---|
| `period_id` | UUID optional | Filter periode Blue Book |
| `publish_year` | int optional | Filter tahun Green Book |
| `include_history` | bool default false | Hitung snapshot historis jika true |

## Layout

1. Header: `Ringkasan Eksekutif`
2. Filter bar: periode Blue Book dan tahun Green Book
3. KPI cards: Blue Book, Green Book, Daftar Kegiatan, Loan Agreement
4. Funnel chart BB -> GB -> DK -> LA
5. Top 10 K/L by exposure
6. Top 10 lenders by exposure
7. Risk table:
   - Loan Agreement closing <= 12 bulan
   - Loan Agreement berjalan lama
   - Green Book tanpa Daftar Kegiatan
   - Daftar Kegiatan tanpa Loan Agreement

## Acceptance Criteria

- Dashboard bisa dibuka di route `/dashboard/executive-portfolio`.
- Semua angka berasal dari endpoint dashboard, bukan hitung manual di komponen.
- Funnel tidak double count akibat relasi many-to-many.
- Risk table bisa diklik menuju detail proyek/journey bila route tersedia.
- Tidak ada filter quarter atau budget year.
- Frontend build/type-check lulus.
