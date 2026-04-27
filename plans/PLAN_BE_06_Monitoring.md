# PLAN BE-06 — Monitoring Disbursement

> **Scope:** CRUD monitoring per triwulan (level LA + breakdown komponen opsional).
> **Deliverable:** Monitoring tersimpan dengan guard effective_date. Absorption pct dihitung di server.
> **Referensi:** docs/PRISM_API_Contract.md (Monitoring), docs/PRISM_Business_Rules.md (bagian 7)

---

## Task 1 — sql/queries/monitoring.sql

```sql
-- name: ListMonitoringByLA :many
SELECT * FROM monitoring_disbursement
WHERE loan_agreement_id = $1
ORDER BY budget_year ASC, quarter ASC
LIMIT $2 OFFSET $3;

-- name: GetMonitoring :one
SELECT * FROM monitoring_disbursement WHERE id = $1;

-- name: GetMonitoringByLAAndPeriod :one
SELECT * FROM monitoring_disbursement
WHERE loan_agreement_id = $1 AND budget_year = $2 AND quarter = $3;

-- name: CreateMonitoring :one
INSERT INTO monitoring_disbursement (
    loan_agreement_id, budget_year, quarter,
    exchange_rate_usd_idr, exchange_rate_la_idr,
    planned_la, planned_usd, planned_idr,
    realized_la, realized_usd, realized_idr
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateMonitoring :one
UPDATE monitoring_disbursement
SET exchange_rate_usd_idr=$2, exchange_rate_la_idr=$3,
    planned_la=$4, planned_usd=$5, planned_idr=$6,
    realized_la=$7, realized_usd=$8, realized_idr=$9, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: DeleteMonitoring :exec
DELETE FROM monitoring_disbursement WHERE id=$1;

-- name: GetKomponenByMonitoring :many
SELECT * FROM monitoring_komponen
WHERE monitoring_disbursement_id = $1;

-- name: CreateKomponen :one
INSERT INTO monitoring_komponen (
    monitoring_disbursement_id, component_name,
    planned_la, planned_usd, planned_idr,
    realized_la, realized_usd, realized_idr
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: DeleteKomponenByMonitoring :exec
DELETE FROM monitoring_komponen WHERE monitoring_disbursement_id = $1;

-- Dashboard queries:
-- name: GetDashboardSummary :one
SELECT
    COUNT(DISTINCT bp.id) FILTER (WHERE bp.status = 'active') as total_bb_projects,
    COUNT(DISTINCT gp.id) FILTER (WHERE gp.status = 'active') as total_gb_projects,
    COUNT(DISTINCT la.id) as total_loan_agreements,
    COALESCE(SUM(la.amount_usd), 0) as total_amount_usd,
    COALESCE(SUM(md.realized_usd), 0) as total_realisasi_usd,
    COUNT(DISTINCT md.id) as active_monitoring
FROM bb_project bp
FULL JOIN gb_project gp ON true
FULL JOIN loan_agreement la ON true
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id;

-- name: GetMonitoringSummary :many
SELECT
    md.budget_year, md.quarter,
    l.id as lender_id, l.name as lender_name,
    SUM(md.planned_usd) as total_planned_usd,
    SUM(md.realized_usd) as total_realized_usd
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
WHERE ($1::int IS NULL OR md.budget_year = $1)
  AND ($2::varchar IS NULL OR md.quarter = $2)
  AND ($3::uuid IS NULL OR l.id = $3)
GROUP BY md.budget_year, md.quarter, l.id, l.name
ORDER BY md.budget_year, md.quarter;
```

Jalankan `make generate`.

---

## Task 2 — internal/service/monitoring_service.go

```go
func (s *MonitoringService) CreateMonitoring(ctx context.Context, laID pgtype.UUID, req model.CreateMonitoringRequest) (*model.MonitoringResponse, error) {
    // Guard: cek LA sudah efektif
    la, err := s.queries.GetLoanAgreement(ctx, laID)
    if err != nil {
        return nil, errors.NotFound("Loan Agreement tidak ditemukan")
    }
    if la.EffectiveDate.Time.After(time.Now()) {
        return nil, errors.BusinessRule("Monitoring hanya bisa dibuat setelah Loan Agreement efektif")
    }

    // Cek duplikat (budget_year + quarter)
    existing, _ := s.queries.GetMonitoringByLAAndPeriod(ctx, laID, req.BudgetYear, req.Quarter)
    if existing != nil {
        return nil, errors.Conflict(fmt.Sprintf("Monitoring %s %d sudah ada", req.Quarter, req.BudgetYear))
    }

    tx, err := s.db.Begin(ctx)
    defer tx.Rollback(ctx)
    qtx := s.queries.WithTx(tx)

    monitoring, err := qtx.CreateMonitoring(ctx, ...)

    // Insert komponen (opsional)
    for _, k := range req.Komponen {
        qtx.CreateKomponen(ctx, queries.CreateKomponenParams{
            MonitoringDisbursementID: monitoring.ID,
            ComponentName: k.ComponentName,
            // ...
        })
    }

    tx.Commit(ctx)
    s.notification.Publish("monitoring.created", ...)

    return s.buildResponse(monitoring), nil
}

func (s *MonitoringService) buildResponse(m *queries.MonitoringDisbursement, komponen []queries.MonitoringKomponen) *model.MonitoringResponse {
    // Hitung absorption_pct
    absorptionPct := 0.0
    if m.PlannedUsd.InexactFloat64() > 0 {
        absorptionPct = math.Round(m.RealizedUsd.InexactFloat64() / m.PlannedUsd.InexactFloat64() * 1000) / 10
    }
    return &model.MonitoringResponse{
        // ...
        AbsorptionPct: absorptionPct,
        Komponen: toKomponenResponse(komponen),
    }
}
```

---

## Task 3 — Dashboard Service

```go
func (s *DashboardService) GetSummary(ctx context.Context) (*model.DashboardSummary, error) {
    row, _ := s.queries.GetDashboardSummary(ctx)
    totalAmount := row.TotalAmountUsd.InexactFloat64()
    totalRealisasi := row.TotalRealisasiUsd.InexactFloat64()
    absorptionPct := 0.0
    if totalAmount > 0 {
        absorptionPct = math.Round(totalRealisasi/totalAmount*1000) / 10
    }
    return &model.DashboardSummary{
        TotalBBProjects: int(row.TotalBbProjects),
        TotalGBProjects: int(row.TotalGbProjects),
        // ...
        OverallAbsorptionPct: absorptionPct,
    }, nil
}
```

---

## Task 4 — Journey Endpoint

```sql
-- sql/queries/journey.sql
-- name: GetProjectJourney :one
-- Query kompleks yang fetch seluruh alur: BB → GB → DK → LA → Monitoring
-- Gunakan multiple queries dan assemble di service layer (lebih mudah di-maintain)
```

Journey service: fetch BB project, lalu fetch semua GB terkait, lalu DK, lalu LA, lalu monitoring — assemble menjadi `JourneyResponse`.

---

## Task 5 — Handler & Routes

```go
// Monitoring
monGroup := api.Group("/loan-agreements/:laId/monitoring")
monGroup.GET("", monHandler.List, permission.Require("monitoring_disbursement", "read"))
monGroup.POST("", monHandler.Create, permission.Require("monitoring_disbursement", "create"))
monGroup.GET("/:id", monHandler.Get, permission.Require("monitoring_disbursement", "read"))
monGroup.PUT("/:id", monHandler.Update, permission.Require("monitoring_disbursement", "update"))
monGroup.DELETE("/:id", monHandler.Delete, permission.Require("monitoring_disbursement", "delete"))

// Dashboard
dash := api.Group("/dashboard")
dash.GET("/summary", dashHandler.Summary)
dash.GET("/monitoring-summary", dashHandler.MonitoringSummary)

// Journey
api.GET("/projects/:bbProjectId/journey", journeyHandler.GetJourney, permission.Require("bb_project", "read"))
```

---

## Checklist

- [x] `sql/queries/monitoring.sql` — monitoring + komponen + dashboard queries
- [x] `make generate`
- [x] `internal/model/monitoring.go` + `internal/model/dashboard.go`
- [x] `internal/service/monitoring_service.go` — guard effective_date + absorption_pct computed
- [x] `internal/service/dashboard_service.go`
- [x] `internal/service/journey_service.go` — assemble multi-level response
- [x] Handler monitoring, dashboard, journey
- [x] Routes terdaftar
- [x] `POST /monitoring` sebelum LA efektif → 422
- [x] `POST /monitoring` duplikat quarter → 409
- [x] `GET /dashboard/summary` → angka agregat benar
- [x] `GET /projects/:id/journey` → full tree response
