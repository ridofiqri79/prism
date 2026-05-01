-- ===== MONITORING DISBURSEMENT =====

-- name: ListMonitoringByLA :many
SELECT *
FROM monitoring_disbursement
WHERE loan_agreement_id = sqlc.arg('loan_agreement_id')
AND (
    sqlc.narg('search')::text IS NULL
    OR budget_year::text ILIKE '%' || sqlc.narg('search')::text || '%'
    OR quarter ILIKE '%' || sqlc.narg('search')::text || '%'
    OR EXISTS (
        SELECT 1
        FROM monitoring_komponen mk
        WHERE mk.monitoring_disbursement_id = monitoring_disbursement.id
          AND mk.component_name ILIKE '%' || sqlc.narg('search')::text || '%'
    )
)
AND (sqlc.narg('budget_year')::int IS NULL OR budget_year = sqlc.narg('budget_year')::int)
AND (sqlc.narg('quarter')::varchar IS NULL OR quarter = sqlc.narg('quarter')::varchar)
ORDER BY budget_year ASC, quarter ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountMonitoringByLA :one
SELECT COUNT(*)
FROM monitoring_disbursement
WHERE loan_agreement_id = sqlc.arg('loan_agreement_id')
AND (
    sqlc.narg('search')::text IS NULL
    OR budget_year::text ILIKE '%' || sqlc.narg('search')::text || '%'
    OR quarter ILIKE '%' || sqlc.narg('search')::text || '%'
    OR EXISTS (
        SELECT 1
        FROM monitoring_komponen mk
        WHERE mk.monitoring_disbursement_id = monitoring_disbursement.id
          AND mk.component_name ILIKE '%' || sqlc.narg('search')::text || '%'
    )
)
AND (sqlc.narg('budget_year')::int IS NULL OR budget_year = sqlc.narg('budget_year')::int)
AND (sqlc.narg('quarter')::varchar IS NULL OR quarter = sqlc.narg('quarter')::varchar);

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
    (SELECT COUNT(DISTINCT project_identity_id) FROM bb_project WHERE status = 'active')::bigint AS total_bb_projects,
    (SELECT COUNT(DISTINCT gb_project_identity_id) FROM gb_project WHERE status = 'active')::bigint AS total_gb_projects,
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
SELECT
    bp.id,
    bp.blue_book_id,
    bp.project_identity_id,
    bp.bb_code,
    bp.project_name,
    CONCAT(
        'BB ',
        p.name,
        CASE WHEN bb.revision_number > 0 THEN CONCAT(' Revisi ke-', bb.revision_number) ELSE '' END,
        CASE WHEN bb.revision_year IS NOT NULL THEN CONCAT(' Tahun ', bb.revision_year) ELSE '' END
    )::text AS blue_book_revision_label,
    (bp.id = latest.id)::boolean AS is_latest,
    (bp.id <> latest.id)::boolean AS has_newer_revision,
    latest.id AS latest_bb_project_id,
    CONCAT(
        'BB ',
        latest_period.name,
        CASE WHEN latest_bb.revision_number > 0 THEN CONCAT(' Revisi ke-', latest_bb.revision_number) ELSE '' END,
        CASE WHEN latest_bb.revision_year IS NOT NULL THEN CONCAT(' Tahun ', latest_bb.revision_year) ELSE '' END
    )::text AS latest_blue_book_revision_label
FROM bb_project bp
JOIN blue_book bb ON bb.id = bp.blue_book_id
JOIN period p ON p.id = bb.period_id
JOIN LATERAL (
    SELECT latest_bp.*
    FROM bb_project latest_bp
    JOIN blue_book latest_bb_order ON latest_bb_order.id = latest_bp.blue_book_id
    WHERE latest_bp.project_identity_id = bp.project_identity_id
      AND latest_bp.status = 'active'
    ORDER BY latest_bb_order.revision_number DESC, COALESCE(latest_bb_order.revision_year, 0) DESC, latest_bb_order.created_at DESC
    LIMIT 1
) latest ON TRUE
JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
JOIN period latest_period ON latest_period.id = latest_bb.period_id
WHERE bp.id = $1;

-- name: ListJourneyLenderIndicationsByBBProject :many
SELECT
    li.id,
    li.bb_project_id,
    li.lender_id,
    li.remarks,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM lender_indication li
JOIN lender l ON l.id = li.lender_id
WHERE li.bb_project_id = $1
ORDER BY COALESCE(l.short_name, l.name) ASC;

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
SELECT
    gp.id,
    gp.green_book_id,
    gp.gb_project_identity_id,
    gp.gb_code,
    gp.project_name,
    gp.status,
    CONCAT(
        'GB ',
        gb.publish_year,
        CASE WHEN gb.revision_number > 0 THEN CONCAT(' Revisi ke-', gb.revision_number) ELSE '' END
    )::text AS green_book_revision_label,
    (gp.id = latest.id)::boolean AS is_latest,
    (gp.id <> latest.id)::boolean AS has_newer_revision,
    latest.id AS latest_gb_project_id,
    CONCAT(
        'GB ',
        latest_gb.publish_year,
        CASE WHEN latest_gb.revision_number > 0 THEN CONCAT(' Revisi ke-', latest_gb.revision_number) ELSE '' END
    )::text AS latest_green_book_revision_label
FROM gb_project_bb_project gbp
JOIN gb_project gp ON gp.id = gbp.gb_project_id
JOIN green_book gb ON gb.id = gp.green_book_id
JOIN LATERAL (
    SELECT latest_gp.*
    FROM gb_project latest_gp
    JOIN green_book latest_gb_order ON latest_gb_order.id = latest_gp.green_book_id
    WHERE latest_gp.gb_project_identity_id = gp.gb_project_identity_id
      AND latest_gp.status = 'active'
    ORDER BY latest_gb_order.revision_number DESC, latest_gb_order.created_at DESC
    LIMIT 1
) latest ON TRUE
JOIN green_book latest_gb ON latest_gb.id = latest.green_book_id
WHERE gbp.bb_project_id = $1
ORDER BY gp.gb_code ASC;

-- name: ListJourneyFundingSourcesByGBProjects :many
SELECT
    gfs.id,
    gfs.gb_project_id,
    gfs.lender_id,
    gfs.institution_id,
    gfs.currency,
    gfs.loan_original,
    gfs.grant_original,
    gfs.local_original,
    gfs.loan_usd,
    gfs.grant_usd,
    gfs.local_usd,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name,
    i.name AS institution_name,
    i.short_name AS institution_short_name
FROM gb_funding_source gfs
JOIN lender l ON l.id = gfs.lender_id
LEFT JOIN institution i ON i.id = gfs.institution_id
WHERE gfs.gb_project_id = ANY(sqlc.arg('gb_project_ids')::uuid[])
ORDER BY gfs.gb_project_id ASC, COALESCE(l.short_name, l.name) ASC;

-- name: ListJourneyDKProjectsByGBProjects :many
SELECT
    dpg.gb_project_id,
    dp.id,
    dp.project_name,
    dp.objectives,
    dk.id AS dk_id,
    dk.subject AS dk_subject,
    dk.date AS dk_date,
    dk.letter_number AS dk_letter_number
FROM dk_project_gb_project dpg
JOIN dk_project dp ON dp.id = dpg.dk_project_id
JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
WHERE dpg.gb_project_id = ANY(sqlc.arg('gb_project_ids')::uuid[])
ORDER BY dpg.gb_project_id ASC, dk.date ASC, dp.project_name ASC, dp.id ASC;

-- name: ListJourneyLoanAgreementsByDKProjects :many
SELECT
    la.id,
    la.dk_project_id,
    la.lender_id,
    la.loan_code,
    la.effective_date,
    la.original_closing_date,
    la.closing_date,
    la.agreement_date,
    la.currency,
    la.amount_original,
    la.amount_usd,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE la.dk_project_id = ANY(sqlc.arg('dk_project_ids')::uuid[])
ORDER BY la.dk_project_id ASC;

-- name: ListJourneyMonitoringByLAs :many
SELECT
    id,
    loan_agreement_id,
    budget_year,
    quarter,
    planned_usd,
    realized_usd
FROM monitoring_disbursement
WHERE loan_agreement_id = ANY(sqlc.arg('loan_agreement_ids')::uuid[])
ORDER BY loan_agreement_id ASC, budget_year ASC, quarter ASC;
