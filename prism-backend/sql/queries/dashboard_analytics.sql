-- ===== DASHBOARD ANALYTICS =====
-- Query foundation + aggregation untuk DA-02.

-- ===== OVERVIEW =====

-- name: GetDashboardAnalyticsOverview :one
SELECT
    (SELECT COUNT(DISTINCT project_identity_id) FROM bb_project WHERE status = 'active')::bigint AS total_projects,
    (SELECT COUNT(*) FROM loan_agreement)::bigint AS total_loan_agreements,
    COALESCE((SELECT SUM(amount_usd) FROM loan_agreement), 0)::numeric AS agreement_amount_usd,
    COALESCE((SELECT SUM(planned_usd) FROM monitoring_disbursement md
     WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
       AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)), 0)::numeric AS total_planned_usd,
    COALESCE((SELECT SUM(realized_usd) FROM monitoring_disbursement md
     WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
       AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)), 0)::numeric AS total_realized_usd,
    (SELECT COUNT(DISTINCT la.id) FROM monitoring_disbursement md
     JOIN loan_agreement la ON la.id = md.loan_agreement_id
     WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
       AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar))::bigint AS active_monitoring;

-- name: GetDashboardAnalyticsPipelineFunnel :many
WITH latest_bb_project AS (
    SELECT bp.id, bp.project_identity_id
    FROM bb_project bp
    WHERE bp.status = 'active'
      AND (sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1))
)
SELECT 'BB'::text AS stage,
    COUNT(DISTINCT lbp.project_identity_id)::bigint AS project_count,
    COALESCE((SELECT SUM(pc.amount_usd) FROM bb_project_cost pc
     WHERE pc.funding_type = 'Foreign' AND pc.funding_category = 'Loan'
       AND pc.bb_project_id IN (SELECT lbp2.id FROM latest_bb_project lbp2)), 0)::numeric AS total_loan_usd
FROM latest_bb_project lbp
UNION ALL
SELECT 'GB'::text AS stage,
    COUNT(DISTINCT gp.gb_project_identity_id)::bigint AS project_count,
    COALESCE(SUM(gfs.loan_usd), 0)::numeric AS total_loan_usd
FROM gb_project gp
LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = gp.id
WHERE gp.status = 'active'
UNION ALL
SELECT 'DK'::text AS stage,
    COUNT(DISTINCT dp.id)::bigint AS project_count,
    COALESCE(SUM(dfd.amount_usd), 0)::numeric AS total_loan_usd
FROM dk_project dp
JOIN dk_financing_detail dfd ON dfd.dk_project_id = dp.id
UNION ALL
SELECT 'LA'::text AS stage,
    COUNT(*)::bigint AS project_count,
    COALESCE(SUM(amount_usd), 0)::numeric AS total_loan_usd
FROM loan_agreement
UNION ALL
SELECT 'Monitoring'::text AS stage,
    COUNT(DISTINCT la.id)::bigint AS project_count,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS total_loan_usd
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
ORDER BY stage;

-- name: GetDashboardAnalyticsTopInstitutions :many
SELECT
    i.id AS institution_id,
    i.name AS institution_name,
    i.short_name AS institution_short_name,
    i.level AS institution_level,
    COUNT(DISTINCT dp.id)::bigint AS project_count,
    COUNT(DISTINCT la.id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT md.id)::bigint AS monitoring_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM institution i
JOIN dk_project dp ON dp.institution_id = i.id
JOIN loan_agreement la ON la.dk_project_id = dp.id
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
WHERE i.parent_id IS NULL
  AND (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR i.id = ANY(sqlc.arg('institution_ids')::uuid[]))
GROUP BY i.id, i.name, i.short_name, i.level
ORDER BY agreement_amount_usd DESC
LIMIT CASE WHEN sqlc.arg('limit')::int > 0 THEN sqlc.arg('limit')::int ELSE 10 END;

-- name: GetDashboardAnalyticsTopLenders :many
SELECT
    l.id AS lender_id,
    l.name AS lender_name,
    l.short_name AS lender_short_name,
    l.type AS lender_type,
    COUNT(DISTINCT la.id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT la.dk_project_id)::bigint AS project_count,
    COUNT(DISTINCT dp.institution_id)::bigint AS institution_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM lender l
JOIN loan_agreement la ON la.lender_id = l.id
JOIN dk_project dp ON dp.id = la.dk_project_id
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
WHERE (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
GROUP BY l.id, l.name, l.short_name, l.type
ORDER BY agreement_amount_usd DESC
LIMIT CASE WHEN sqlc.arg('limit')::int > 0 THEN sqlc.arg('limit')::int ELSE 10 END;

-- name: GetDashboardAnalyticsLenderProportionBB :many
SELECT
    l.type AS lender_type,
    COUNT(DISTINCT lbp.project_identity_id)::bigint AS project_count,
    COUNT(DISTINCT l.id)::bigint AS lender_count,
    COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
FROM lender l
JOIN lender_indication li ON li.lender_id = l.id
JOIN bb_project lbp ON lbp.id = li.bb_project_id AND lbp.status = 'active'
LEFT JOIN bb_project_cost bpc ON bpc.bb_project_id = lbp.id
    AND bpc.funding_type = 'Foreign' AND bpc.funding_category = 'Loan'
WHERE (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
GROUP BY l.type;

-- name: GetDashboardAnalyticsLenderProportionGB :many
SELECT
    l.type AS lender_type,
    COUNT(DISTINCT gp.gb_project_identity_id)::bigint AS project_count,
    COUNT(DISTINCT l.id)::bigint AS lender_count,
    COALESCE(SUM(gfs.loan_usd), 0)::numeric AS amount_usd
FROM lender l
JOIN gb_funding_source gfs ON gfs.lender_id = l.id
JOIN gb_project gp ON gp.id = gfs.gb_project_id AND gp.status = 'active'
WHERE (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
GROUP BY l.type;

-- name: GetDashboardAnalyticsLenderProportionLA :many
SELECT
    l.type AS lender_type,
    COUNT(DISTINCT la.id)::bigint AS project_count,
    COUNT(DISTINCT l.id)::bigint AS lender_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd
FROM lender l
JOIN loan_agreement la ON la.lender_id = l.id
WHERE (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
GROUP BY l.type;

-- name: GetDashboardAnalyticsLenderProportionMonitoring :many
SELECT
    l.type AS lender_type,
    COUNT(DISTINCT md.loan_agreement_id)::bigint AS project_count,
    COUNT(DISTINCT l.id)::bigint AS lender_count,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
GROUP BY l.type;

-- ===== INSTITUTIONS =====

-- name: GetDashboardAnalyticsInstitutionsSummary :one
WITH institution_rollup AS (
    SELECT i.id AS source_institution_id, i.id AS root_institution_id, i.name AS root_institution_name
    FROM institution i WHERE i.parent_id IS NULL
    UNION ALL
    SELECT child.id AS source_institution_id, ir.root_institution_id, ir.root_institution_name
    FROM institution child
    JOIN institution_rollup ir ON ir.source_institution_id = child.parent_id
)
SELECT
    COUNT(DISTINCT ir.root_institution_id)::bigint AS institution_count,
    COUNT(DISTINCT dp.id)::bigint AS project_count,
    COUNT(DISTINCT dp.id || '|' || ir.source_institution_id::text)::bigint AS assignment_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM institution_rollup ir
LEFT JOIN dk_project dp ON dp.institution_id = ir.source_institution_id
LEFT JOIN loan_agreement la ON la.dk_project_id = dp.id
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
WHERE (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[]));

-- name: GetDashboardAnalyticsInstitutions :many
WITH monitoring_fact AS (
    SELECT md.planned_usd, md.realized_usd, la.id AS loan_agreement_id,
           la.amount_usd AS agreement_amount_usd, la.lender_id,
           l.type AS lender_type, dp.id AS dk_project_id, dp.institution_id AS dk_institution_id
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
),
institution_rollup AS (
    SELECT i.id AS source_institution_id, i.id AS root_institution_id,
           i.name AS root_institution_name, i.short_name AS root_institution_short_name,
           i.level AS root_institution_level
    FROM institution i WHERE i.parent_id IS NULL
    UNION ALL
    SELECT child.id AS source_institution_id, ir.root_institution_id,
           ir.root_institution_name, ir.root_institution_short_name, ir.root_institution_level
    FROM institution child
    JOIN institution_rollup ir ON ir.source_institution_id = child.parent_id
)
SELECT
    ir.root_institution_id AS institution_id,
    ir.root_institution_name AS institution_name,
    ir.root_institution_short_name AS institution_short_name,
    ir.root_institution_level AS institution_level,
    COUNT(DISTINCT dp.id)::bigint AS project_count,
    COUNT(DISTINCT dp.id || '|' || ir.source_institution_id::text)::bigint AS assignment_count,
    COUNT(DISTINCT la.id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT mf.loan_agreement_id)::bigint AS monitoring_count,
    COALESCE(SUM(mf.agreement_amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(mf.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(mf.realized_usd), 0)::numeric AS realized_usd,
    ARRAY(
        SELECT DISTINCT mf2.lender_type FROM institution_rollup ir2
        JOIN dk_project dp2 ON dp2.institution_id = ir2.source_institution_id
        JOIN loan_agreement la2 ON la2.id = dp2.id
        JOIN lender l2 ON l2.id = la2.lender_id
        WHERE ir2.root_institution_id = ir.root_institution_id AND l2.type IS NOT NULL
        ORDER BY l2.type
    )::text[] AS loan_types
FROM institution_rollup ir
LEFT JOIN dk_project dp ON dp.institution_id = ir.source_institution_id
LEFT JOIN loan_agreement la ON la.dk_project_id = dp.id
LEFT JOIN monitoring_fact mf ON mf.loan_agreement_id = la.id AND mf.dk_institution_id = dp.institution_id
WHERE (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[]))
GROUP BY ir.root_institution_id, ir.root_institution_name, ir.root_institution_short_name, ir.root_institution_level
HAVING COUNT(DISTINCT dp.id) > 0
ORDER BY agreement_amount_usd DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDashboardAnalyticsInstitutions :one
WITH institution_rollup AS (
    SELECT i.id AS source_institution_id, i.id AS root_institution_id
    FROM institution i WHERE i.parent_id IS NULL
    UNION ALL
    SELECT child.id AS source_institution_id, ir.root_institution_id
    FROM institution child
    JOIN institution_rollup ir ON ir.source_institution_id = child.parent_id
)
SELECT COUNT(DISTINCT ir.root_institution_id)::bigint
FROM institution_rollup ir
JOIN dk_project dp ON dp.institution_id = ir.source_institution_id
WHERE (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[]));

-- ===== LENDERS =====

-- name: GetDashboardAnalyticsLendersSummary :one
SELECT
    COUNT(DISTINCT l.id)::bigint AS lender_count,
    COUNT(DISTINCT la.id)::bigint AS loan_agreement_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM lender l
JOIN loan_agreement la ON la.lender_id = l.id
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
WHERE (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]));

-- name: GetDashboardAnalyticsLenders :many
WITH monitoring_fact AS (
    SELECT md.planned_usd, md.realized_usd, la.id AS loan_agreement_id,
           la.amount_usd AS agreement_amount_usd, la.lender_id
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l2 ON l2.id = la.lender_id
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l2.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
)
SELECT
    l.id AS lender_id, l.name AS lender_name, l.short_name AS lender_short_name, l.type AS lender_type,
    COUNT(DISTINCT la_all.id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT la_all.dk_project_id)::bigint AS project_count,
    COUNT(DISTINCT dp.institution_id)::bigint AS institution_count,
    COALESCE(SUM(mf.agreement_amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(mf.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(mf.realized_usd), 0)::numeric AS realized_usd
FROM lender l
JOIN loan_agreement la_all ON la_all.lender_id = l.id
JOIN dk_project dp ON dp.id = la_all.dk_project_id
LEFT JOIN monitoring_fact mf ON mf.loan_agreement_id = la_all.id AND mf.lender_id = l.id
GROUP BY l.id, l.name, l.short_name, l.type
ORDER BY agreement_amount_usd DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDashboardAnalyticsLenders :one
SELECT COUNT(DISTINCT l.id)::bigint
FROM lender l
JOIN loan_agreement la ON la.lender_id = l.id;

-- name: GetDashboardAnalyticsLenderInstitutionMatrix :many
SELECT
    i.id AS institution_id, i.name AS institution_name, i.short_name AS institution_short_name,
    l.id AS lender_id, l.name AS lender_name, l.short_name AS lender_short_name, l.type AS lender_type,
    COUNT(DISTINCT dp.id)::bigint AS project_count,
    COUNT(DISTINCT la.id)::bigint AS loan_agreement_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM institution i
JOIN dk_project dp ON dp.institution_id = i.id
JOIN loan_agreement la ON la.dk_project_id = dp.id
JOIN lender l ON l.id = la.lender_id
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
WHERE i.parent_id IS NULL
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR i.id = ANY(sqlc.arg('institution_ids')::uuid[]))
GROUP BY i.id, i.name, i.short_name, l.id, l.name, l.short_name, l.type
ORDER BY agreement_amount_usd DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- ===== ABSORPTION =====

-- name: GetDashboardAnalyticsAbsorptionSummary :one
SELECT
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]));

-- name: GetDashboardAnalyticsAbsorptionByLender :many
SELECT
    l.id AS lender_id, l.name AS lender_name, l.short_name AS lender_short_name, l.type AS lender_type,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd,
    COUNT(md.id)::bigint AS monitoring_count
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
GROUP BY l.id, l.name, l.short_name, l.type
ORDER BY realized_usd DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetDashboardAnalyticsAbsorptionByInstitution :many
SELECT
    i.id AS institution_id, i.name AS institution_name, i.short_name AS institution_short_name, i.level AS institution_level,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd,
    COUNT(md.id)::bigint AS monitoring_count
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
JOIN dk_project dp ON dp.id = la.dk_project_id
JOIN institution i ON i.id = dp.institution_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR i.id = ANY(sqlc.arg('institution_ids')::uuid[]))
GROUP BY i.id, i.name, i.short_name, i.level
ORDER BY realized_usd DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetDashboardAnalyticsAbsorptionByProject :many
SELECT
    dp.id AS dk_project_id, dp.project_name,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd,
    COUNT(md.id)::bigint AS monitoring_count
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
JOIN dk_project dp ON dp.id = la.dk_project_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
GROUP BY dp.id, dp.project_name
ORDER BY realized_usd DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- ===== YEARLY =====

-- name: GetDashboardAnalyticsYearly :many
SELECT
    md.budget_year,
    md.quarter,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd,
    COUNT(DISTINCT la.id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT la.dk_project_id)::bigint AS project_count
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN lender l ON l.id = la.lender_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
  AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR l.id = ANY(sqlc.arg('lender_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
GROUP BY md.budget_year, md.quarter
ORDER BY md.budget_year ASC, md.quarter ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDashboardAnalyticsYearly :one
SELECT COUNT(DISTINCT budget_year)::bigint FROM monitoring_disbursement;

-- ===== RISKS =====

-- name: GetDashboardAnalyticsClosingRisks :many
SELECT
    la.id AS loan_agreement_id, la.loan_code,
    la.closing_date,
    la.closing_date - CURRENT_DATE AS days_to_closing,
    l.name AS lender_name, l.type AS lender_type,
    CASE WHEN COALESCE(SUM(md.planned_usd), 0) = 0 THEN 0::numeric
         ELSE ROUND((COALESCE(SUM(md.realized_usd), 0) / SUM(md.planned_usd) * 100)::numeric, 2)
    END AS absorption_pct
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
LEFT JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
WHERE la.closing_date > CURRENT_DATE
  AND la.closing_date <= CURRENT_DATE + INTERVAL '365 days'
  AND la.effective_date <= CURRENT_DATE
GROUP BY la.id, la.loan_code, la.closing_date, l.name, l.type
ORDER BY la.closing_date ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDashboardAnalyticsClosingRisks :one
SELECT COUNT(*)::bigint FROM loan_agreement la
WHERE la.closing_date > CURRENT_DATE
  AND la.closing_date <= CURRENT_DATE + INTERVAL '365 days'
  AND la.effective_date <= CURRENT_DATE;

-- name: GetDashboardAnalyticsEffectiveWithoutMonitoring :many
SELECT
    la.id AS loan_agreement_id, la.loan_code,
    la.effective_date, la.amount_usd,
    l.name AS lender_name, l.type AS lender_type
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE la.effective_date <= CURRENT_DATE
  AND NOT EXISTS (SELECT 1 FROM monitoring_disbursement md WHERE md.loan_agreement_id = la.id)
ORDER BY la.effective_date DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDashboardAnalyticsEffectiveWithoutMonitoring :one
SELECT COUNT(*)::bigint FROM loan_agreement la
WHERE la.effective_date <= CURRENT_DATE
  AND NOT EXISTS (SELECT 1 FROM monitoring_disbursement md WHERE md.loan_agreement_id = la.id);

-- name: GetDashboardAnalyticsExtendedLoans :many
SELECT
    la.id AS loan_agreement_id, la.loan_code,
    la.original_closing_date, la.closing_date,
    la.closing_date - la.original_closing_date AS extension_days,
    l.name AS lender_name, l.type AS lender_type
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE la.closing_date != la.original_closing_date
ORDER BY la.closing_date - la.original_closing_date DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDashboardAnalyticsExtendedLoans :one
SELECT COUNT(*)::bigint FROM loan_agreement WHERE closing_date != original_closing_date;

-- name: GetDashboardAnalyticsDataQualityCounts :one
SELECT
    (SELECT COUNT(*)::bigint FROM bb_project bp WHERE bp.status = 'active'
       AND NOT EXISTS (SELECT 1 FROM bb_project_institution bpi
        WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency')) AS missing_executing_agency_count,
    (SELECT COUNT(*)::bigint FROM bb_project bp WHERE bp.status = 'active'
       AND NOT EXISTS (SELECT 1 FROM lender_indication li
        WHERE li.bb_project_id = bp.id)) AS missing_lender_indication_count,
    (SELECT COUNT(DISTINCT bp.project_identity_id)::bigint FROM bb_project bp WHERE bp.status = 'active'
       AND NOT EXISTS (SELECT 1 FROM gb_project_bb_project gpbp
        WHERE gpbp.bb_project_id = bp.id)) AS project_without_gb_count;
