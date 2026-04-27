-- ===== MONITORING DISBURSEMENT =====

-- name: ListMonitoringByLA :many
SELECT *
FROM monitoring_disbursement
WHERE loan_agreement_id = $1
ORDER BY budget_year ASC, quarter ASC
LIMIT $2 OFFSET $3;

-- name: CountMonitoringByLA :one
SELECT COUNT(*)
FROM monitoring_disbursement
WHERE loan_agreement_id = $1;

-- name: GetMonitoring :one
SELECT *
FROM monitoring_disbursement
WHERE id = $1;

-- name: GetMonitoringByLA :one
SELECT *
FROM monitoring_disbursement
WHERE id = $1
  AND loan_agreement_id = $2;

-- name: GetMonitoringByLAAndPeriod :one
SELECT *
FROM monitoring_disbursement
WHERE loan_agreement_id = $1
  AND budget_year = $2
  AND quarter = $3;

-- name: CreateMonitoring :one
INSERT INTO monitoring_disbursement (
    loan_agreement_id,
    budget_year,
    quarter,
    exchange_rate_usd_idr,
    exchange_rate_la_idr,
    planned_la,
    planned_usd,
    planned_idr,
    realized_la,
    realized_usd,
    realized_idr
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateMonitoring :one
UPDATE monitoring_disbursement
SET exchange_rate_usd_idr = $3,
    exchange_rate_la_idr = $4,
    planned_la = $5,
    planned_usd = $6,
    planned_idr = $7,
    realized_la = $8,
    realized_usd = $9,
    realized_idr = $10,
    updated_at = NOW()
WHERE id = $1
  AND loan_agreement_id = $2
RETURNING *;

-- name: DeleteMonitoring :exec
DELETE FROM monitoring_disbursement
WHERE id = $1
  AND loan_agreement_id = $2;

-- name: GetKomponenByMonitoring :many
SELECT *
FROM monitoring_komponen
WHERE monitoring_disbursement_id = $1
ORDER BY component_name ASC;

-- name: CreateKomponen :one
INSERT INTO monitoring_komponen (
    monitoring_disbursement_id,
    component_name,
    planned_la,
    planned_usd,
    planned_idr,
    realized_la,
    realized_usd,
    realized_idr
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: DeleteKomponenByMonitoring :exec
DELETE FROM monitoring_komponen
WHERE monitoring_disbursement_id = $1;

-- ===== DASHBOARD =====

-- name: GetDashboardSummary :one
SELECT
    (SELECT COUNT(*) FROM bb_project WHERE status = 'active')::bigint AS total_bb_projects,
    (SELECT COUNT(*) FROM gb_project WHERE status = 'active')::bigint AS total_gb_projects,
    (SELECT COUNT(*) FROM loan_agreement)::bigint AS total_loan_agreements,
    COALESCE((SELECT SUM(amount_usd) FROM loan_agreement), 0)::numeric AS total_amount_usd,
    COALESCE((SELECT SUM(realized_usd) FROM monitoring_disbursement), 0)::numeric AS total_realized_usd,
    (SELECT COUNT(*) FROM monitoring_disbursement)::bigint AS active_monitoring;

-- name: GetMonitoringSummary :many
SELECT
    md.budget_year,
    md.quarter,
    l.id AS lender_id,
    l.name AS lender_name,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS total_planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS total_realized_usd
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (sqlc.narg('lender_id')::uuid IS NULL OR l.id = sqlc.narg('lender_id')::uuid)
GROUP BY md.budget_year, md.quarter, l.id, l.name
ORDER BY md.budget_year ASC, md.quarter ASC, l.name ASC;

-- ===== JOURNEY =====

-- name: GetJourneyBBProject :one
SELECT id, bb_code, project_name
FROM bb_project
WHERE id = $1;

-- name: ListJourneyLoIsByBBProject :many
SELECT
    loi.id,
    loi.bb_project_id,
    loi.lender_id,
    loi.subject,
    loi.date,
    loi.letter_number,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loi
JOIN lender l ON l.id = loi.lender_id
WHERE loi.bb_project_id = $1
ORDER BY loi.date DESC;

-- name: ListJourneyGBProjectsByBBProject :many
SELECT DISTINCT
    gp.id,
    gp.gb_code,
    gp.project_name,
    gp.status
FROM gb_project_bb_project gbp
JOIN gb_project gp ON gp.id = gbp.gb_project_id
WHERE gbp.bb_project_id = $1
ORDER BY gp.gb_code ASC;

-- name: ListJourneyDKProjectsByGBProject :many
SELECT DISTINCT
    dp.id,
    dp.objectives
FROM dk_project_gb_project dpg
JOIN dk_project dp ON dp.id = dpg.dk_project_id
WHERE dpg.gb_project_id = $1
ORDER BY dp.id ASC;

-- name: GetJourneyLoanAgreementByDKProject :one
SELECT
    la.id,
    la.dk_project_id,
    la.lender_id,
    la.loan_code,
    la.effective_date,
    la.original_closing_date,
    la.closing_date,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE la.dk_project_id = $1;

-- name: ListJourneyMonitoringByLA :many
SELECT
    id,
    budget_year,
    quarter,
    planned_usd,
    realized_usd
FROM monitoring_disbursement
WHERE loan_agreement_id = $1
ORDER BY budget_year ASC, quarter ASC;
