-- ===== DASHBOARD FOUNDATION =====

-- name: GetDashboardSummary :one
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM bb_project_institution bpi
              WHERE bpi.bb_project_id = bp.id
                AND bpi.institution_id = sqlc.narg('institution_id')::uuid
          )
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM lender_indication li
              WHERE li.bb_project_id = bp.id
                AND li.lender_id = sqlc.narg('lender_id')::uuid
          )
      )
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_institution gpi
              WHERE gpi.gb_project_id = gp.id
                AND gpi.institution_id = sqlc.narg('institution_id')::uuid
          )
          OR EXISTS (
              SELECT 1
              FROM gb_funding_source gfs
              WHERE gfs.gb_project_id = gp.id
                AND gfs.institution_id = sqlc.narg('institution_id')::uuid
          )
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_funding_source gfs
              WHERE gfs.gb_project_id = gp.id
                AND gfs.lender_id = sqlc.narg('lender_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
bb_costs AS (
    SELECT COALESCE(SUM(bpc.amount_usd), 0)::numeric AS bb_pipeline_usd
    FROM selected_bb sb
    JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id
    WHERE bpc.funding_type = 'Foreign'
),
gb_costs AS (
    SELECT
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS gb_pipeline_usd,
        COALESCE(SUM(gfs.local_usd), 0)::numeric AS gb_local_usd
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
),
dk_costs AS (
    SELECT
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS dk_financing_usd,
        COALESCE(SUM(dfd.counterpart_usd), 0)::numeric AS dk_counterpart_usd
    FROM dk_financing_detail dfd
    JOIN dk_project dp ON dp.id = dfd.dk_project_id
    WHERE (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
      AND (
          sqlc.narg('publish_year')::int IS NULL
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              JOIN green_book gb ON gb.id = gp.green_book_id
              WHERE dpg.dk_project_id = dp.id
                AND gb.publish_year = sqlc.narg('publish_year')::int
          )
      )
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE dpg.dk_project_id = dp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
la_costs AS (
    SELECT
        COALESCE(SUM(la.amount_usd), 0)::numeric AS la_commitment_usd,
        COUNT(*)::bigint AS total_loan_agreements
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
monitoring_costs AS (
    SELECT
        COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_disbursement_usd,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_disbursement_usd
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
)
SELECT
    (SELECT COUNT(*) FROM selected_bb)::bigint AS total_bb_projects,
    (SELECT COUNT(*) FROM selected_gb)::bigint AS total_gb_projects,
    la_costs.total_loan_agreements,
    bb_costs.bb_pipeline_usd,
    gb_costs.gb_pipeline_usd,
    gb_costs.gb_local_usd,
    dk_costs.dk_financing_usd,
    dk_costs.dk_counterpart_usd,
    la_costs.la_commitment_usd,
    monitoring_costs.planned_disbursement_usd,
    monitoring_costs.realized_disbursement_usd,
    CASE
        WHEN monitoring_costs.planned_disbursement_usd = 0 THEN 0
        ELSE (monitoring_costs.realized_disbursement_usd / monitoring_costs.planned_disbursement_usd * 100)
    END::numeric AS absorption_pct,
    CASE
        WHEN la_costs.la_commitment_usd = 0 THEN 0
        ELSE (monitoring_costs.realized_disbursement_usd / la_costs.la_commitment_usd * 100)
    END::numeric AS la_absorption_pct,
    (la_costs.la_commitment_usd - monitoring_costs.realized_disbursement_usd)::numeric AS undisbursed_usd
FROM bb_costs, gb_costs, dk_costs, la_costs, monitoring_costs;

-- name: GetDashboardStageCounts :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
stage_source AS (
    SELECT
        sb.id,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.effective_date <= CURRENT_DATE
            ) THEN 'MONITORING_ACTIVE'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.effective_date <= CURRENT_DATE
            ) THEN 'LA_EFFECTIVE_NO_MONITORING'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.effective_date > CURRENT_DATE
            ) THEN 'LA_SIGNED_NOT_EFFECTIVE'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                LEFT JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.id IS NULL
            ) THEN 'DK'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                LEFT JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND dpg.dk_project_id IS NULL
            ) THEN 'GB'
            WHEN EXISTS (
                SELECT 1
                FROM loi
                WHERE loi.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
            ) THEN 'BB_WITH_LOI'
            WHEN EXISTS (
                SELECT 1
                FROM lender_indication li
                WHERE li.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
            ) THEN 'BB_WITH_LENDER_INDICATION'
            ELSE 'BB_ONLY'
        END::text AS stage
    FROM selected_bb sb
)
SELECT stage, COUNT(*)::bigint AS project_count
FROM stage_source
GROUP BY stage
ORDER BY CASE stage
    WHEN 'BB_ONLY' THEN 1
    WHEN 'BB_WITH_LENDER_INDICATION' THEN 2
    WHEN 'BB_WITH_LOI' THEN 3
    WHEN 'GB' THEN 4
    WHEN 'DK' THEN 5
    WHEN 'LA_SIGNED_NOT_EFFECTIVE' THEN 6
    WHEN 'LA_EFFECTIVE_NO_MONITORING' THEN 7
    WHEN 'MONITORING_ACTIVE' THEN 8
    ELSE 99
END;

-- name: GetDashboardStageAmounts :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
stage_source AS (
    SELECT
        sb.id,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.effective_date <= CURRENT_DATE
            ) THEN 'MONITORING_ACTIVE'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.effective_date <= CURRENT_DATE
            ) THEN 'LA_EFFECTIVE_NO_MONITORING'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN gb_project gp ON gp.id = gbp.gb_project_id
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gp.id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.effective_date > CURRENT_DATE
            ) THEN 'LA_SIGNED_NOT_EFFECTIVE'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                LEFT JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND la.id IS NULL
            ) THEN 'DK'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                LEFT JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
                  AND dpg.dk_project_id IS NULL
            ) THEN 'GB'
            WHEN EXISTS (
                SELECT 1
                FROM loi
                WHERE loi.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
            ) THEN 'BB_WITH_LOI'
            WHEN EXISTS (
                SELECT 1
                FROM lender_indication li
                WHERE li.bb_project_id IN (SELECT bp2.id FROM bb_project bp2 WHERE bp2.project_identity_id = sb.project_identity_id)
            ) THEN 'BB_WITH_LENDER_INDICATION'
            ELSE 'BB_ONLY'
        END::text AS stage
    FROM selected_bb sb
),
stage_amount AS (
    SELECT
        ss.stage,
        ss.id,
        COALESCE(
            (SELECT SUM(la.amount_usd)
             FROM gb_project_bb_project gbp
             JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
             JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
             WHERE gbp.bb_project_id = ss.id),
            (SELECT SUM(dfd.amount_usd + dfd.grant_usd)
             FROM gb_project_bb_project gbp
             JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
             JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
             WHERE gbp.bb_project_id = ss.id),
            (SELECT SUM(gfs.loan_usd + gfs.grant_usd)
             FROM gb_project_bb_project gbp
             JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
             WHERE gbp.bb_project_id = ss.id),
            (SELECT SUM(bpc.amount_usd)
             FROM bb_project_cost bpc
             WHERE bpc.bb_project_id = ss.id
               AND bpc.funding_type = 'Foreign'),
            0
        )::numeric AS amount_usd
    FROM stage_source ss
)
SELECT stage, COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM stage_amount
GROUP BY stage
ORDER BY CASE stage
    WHEN 'BB_ONLY' THEN 1
    WHEN 'BB_WITH_LENDER_INDICATION' THEN 2
    WHEN 'BB_WITH_LOI' THEN 3
    WHEN 'GB' THEN 4
    WHEN 'DK' THEN 5
    WHEN 'LA_SIGNED_NOT_EFFECTIVE' THEN 6
    WHEN 'LA_EFFECTIVE_NO_MONITORING' THEN 7
    WHEN 'MONITORING_ACTIVE' THEN 8
    ELSE 99
END;

-- name: GetDashboardMonitoringRollup :many
SELECT
    md.budget_year,
    md.quarter,
    COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd,
    CASE
        WHEN COALESCE(SUM(md.planned_usd), 0) = 0 THEN 0
        ELSE COALESCE(SUM(md.realized_usd), 0) / COALESCE(SUM(md.planned_usd), 0) * 100
    END::numeric AS absorption_pct
FROM monitoring_disbursement md
JOIN loan_agreement la ON la.id = md.loan_agreement_id
JOIN dk_project dp ON dp.id = la.dk_project_id
WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
  AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
  AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
  AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
GROUP BY md.budget_year, md.quarter
ORDER BY md.budget_year ASC, md.quarter ASC;

-- name: GetDashboardLAExposureRollup :one
WITH monitoring_by_la AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY md.loan_agreement_id
),
exposure AS (
    SELECT
        COALESCE(SUM(la.amount_usd), 0)::numeric AS la_commitment_usd,
        COALESCE(SUM(mbl.realized_usd), 0)::numeric AS realized_disbursement_usd
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN monitoring_by_la mbl ON mbl.loan_agreement_id = la.id
    WHERE (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
)
SELECT
    la_commitment_usd,
    realized_disbursement_usd,
    (la_commitment_usd - realized_disbursement_usd)::numeric AS undisbursed_usd,
    CASE
        WHEN la_commitment_usd = 0 THEN 0
        ELSE realized_disbursement_usd / la_commitment_usd * 100
    END::numeric AS la_absorption_pct
FROM exposure;

-- name: GetDashboardLenderRollup :many
WITH monitoring_by_la AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY md.loan_agreement_id
)
SELECT
    l.id,
    COALESCE(l.short_name, l.name)::text AS label,
    COUNT(DISTINCT la.id)::bigint AS item_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd,
    COALESCE(SUM(mbl.realized_usd), 0)::numeric AS realized_usd
FROM lender l
LEFT JOIN loan_agreement la ON la.lender_id = l.id
LEFT JOIN dk_project dp ON dp.id = la.dk_project_id
LEFT JOIN monitoring_by_la mbl ON mbl.loan_agreement_id = la.id
WHERE (sqlc.narg('lender_id')::uuid IS NULL OR l.id = sqlc.narg('lender_id')::uuid)
  AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
GROUP BY l.id, COALESCE(l.short_name, l.name)
HAVING COUNT(DISTINCT la.id) > 0
ORDER BY amount_usd DESC, label ASC;

-- name: GetDashboardInstitutionRollup :many
WITH monitoring_by_la AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY md.loan_agreement_id
)
SELECT
    i.id,
    COALESCE(i.short_name, i.name)::text AS label,
    COUNT(DISTINCT dp.id)::bigint AS item_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd,
    COALESCE(SUM(mbl.realized_usd), 0)::numeric AS realized_usd
FROM institution i
LEFT JOIN dk_project dp ON dp.institution_id = i.id
LEFT JOIN loan_agreement la ON la.dk_project_id = dp.id
LEFT JOIN monitoring_by_la mbl ON mbl.loan_agreement_id = la.id
WHERE (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
  AND (sqlc.narg('institution_id')::uuid IS NULL OR i.id = sqlc.narg('institution_id')::uuid)
GROUP BY i.id, COALESCE(i.short_name, i.name)
HAVING COUNT(DISTINCT dp.id) > 0
ORDER BY amount_usd DESC, label ASC;

-- name: ListDashboardFilterOptions :many
SELECT 'period'::text AS option_type, p.id::text AS value, p.name::text AS label
FROM period p
UNION ALL
SELECT 'publish_year'::text AS option_type, gb.publish_year::text AS value, gb.publish_year::text AS label
FROM green_book gb
GROUP BY gb.publish_year
UNION ALL
SELECT 'budget_year'::text AS option_type, md.budget_year::text AS value, md.budget_year::text AS label
FROM monitoring_disbursement md
GROUP BY md.budget_year
UNION ALL
SELECT 'quarter'::text AS option_type, q.quarter AS value, q.quarter AS label
FROM (VALUES ('TW1'), ('TW2'), ('TW3'), ('TW4')) AS q(quarter)
UNION ALL
SELECT 'lender'::text AS option_type, l.id::text AS value, COALESCE(l.short_name, l.name)::text AS label
FROM lender l
UNION ALL
SELECT 'institution'::text AS option_type, i.id::text AS value, COALESCE(i.short_name, i.name)::text AS label
FROM institution i
ORDER BY option_type ASC, label ASC;

-- name: GetDashboardExecutiveFunnel :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
bb_stage AS (
    SELECT sb.id
    FROM selected_bb sb
    WHERE NOT EXISTS (
        SELECT 1
        FROM bb_project bp_any
        JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp_any.id
        WHERE bp_any.project_identity_id = sb.project_identity_id
    )
),
gb_stage AS (
    SELECT sg.id
    FROM selected_gb sg
    WHERE NOT EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        WHERE dpg.gb_project_id = sg.id
    )
),
dk_stage AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    JOIN dk_project_gb_project dpg ON dpg.dk_project_id = dp.id
    JOIN selected_gb sg ON sg.id = dpg.gb_project_id
    WHERE NOT EXISTS (
        SELECT 1
        FROM loan_agreement la
        WHERE la.dk_project_id = dp.id
    )
),
la_stage AS (
    SELECT la.id, la.amount_usd
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN selected_gb sg ON sg.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = dp.id
    )
    AND NOT EXISTS (
        SELECT 1
        FROM monitoring_disbursement md
        WHERE md.loan_agreement_id = la.id
          AND (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
          AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    )
),
monitoring_stage AS (
    SELECT
        la.id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
    WHERE EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN selected_gb sg ON sg.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = dp.id
    )
      AND (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY la.id
)
SELECT 'BB'::text AS stage, COUNT(DISTINCT bs.id)::bigint AS project_count, COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
FROM bb_stage bs
LEFT JOIN bb_project_cost bpc ON bpc.bb_project_id = bs.id AND bpc.funding_type = 'Foreign'
UNION ALL
SELECT 'GB'::text AS stage, COUNT(DISTINCT gs.id)::bigint AS project_count, COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS amount_usd
FROM gb_stage gs
LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = gs.id
UNION ALL
SELECT 'DK'::text AS stage, COUNT(DISTINCT ds.id)::bigint AS project_count, COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS amount_usd
FROM dk_stage ds
LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = ds.id
UNION ALL
SELECT 'LA'::text AS stage, COUNT(*)::bigint AS project_count, COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM la_stage
UNION ALL
SELECT 'MONITORING'::text AS stage, COUNT(*)::bigint AS project_count, COALESCE(SUM(realized_usd), 0)::numeric AS amount_usd
FROM monitoring_stage;

-- name: GetDashboardExecutiveTopInstitutions :many
WITH RECURSIVE institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        i.name,
        i.short_name,
        i.level
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        parent.name,
        parent.short_name,
        parent.level
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
institution_roots AS (
    SELECT DISTINCT ON (institution_id)
        institution_id,
        ancestor_id AS root_id,
        COALESCE(short_name, name)::text AS root_label
    FROM institution_ancestors
    WHERE parent_id IS NULL OR level = 'Kementerian/Badan/Lembaga'
    ORDER BY institution_id, CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END
),
monitoring_by_la AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY md.loan_agreement_id
)
SELECT
    ir.root_id AS id,
    ir.root_label AS label,
    COUNT(DISTINCT dp.id)::bigint AS item_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd,
    COALESCE(SUM(mbl.realized_usd), 0)::numeric AS realized_usd
FROM dk_project dp
JOIN institution_roots ir ON ir.institution_id = dp.institution_id
LEFT JOIN loan_agreement la ON la.dk_project_id = dp.id
LEFT JOIN monitoring_by_la mbl ON mbl.loan_agreement_id = la.id
WHERE (
    sqlc.narg('publish_year')::int IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project gp ON gp.id = dpg.gb_project_id
        JOIN green_book gb ON gb.id = gp.green_book_id
        WHERE dpg.dk_project_id = dp.id
          AND gb.publish_year = sqlc.narg('publish_year')::int
    )
)
AND (
    sqlc.narg('period_id')::uuid IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE dpg.dk_project_id = dp.id
          AND bb.period_id = sqlc.narg('period_id')::uuid
    )
)
GROUP BY ir.root_id, ir.root_label
ORDER BY amount_usd DESC, item_count DESC, label ASC
LIMIT 10;

-- name: GetDashboardExecutiveTopLenders :many
WITH monitoring_by_la AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY md.loan_agreement_id
)
SELECT
    l.id,
    COALESCE(l.short_name, l.name)::text AS label,
    COUNT(DISTINCT la.id)::bigint AS item_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd,
    COALESCE(SUM(mbl.realized_usd), 0)::numeric AS realized_usd
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
JOIN dk_project dp ON dp.id = la.dk_project_id
LEFT JOIN monitoring_by_la mbl ON mbl.loan_agreement_id = la.id
WHERE (
    sqlc.narg('publish_year')::int IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project gp ON gp.id = dpg.gb_project_id
        JOIN green_book gb ON gb.id = gp.green_book_id
        WHERE dpg.dk_project_id = dp.id
          AND gb.publish_year = sqlc.narg('publish_year')::int
    )
)
AND (
    sqlc.narg('period_id')::uuid IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE dpg.dk_project_id = dp.id
          AND bb.period_id = sqlc.narg('period_id')::uuid
    )
)
GROUP BY l.id, COALESCE(l.short_name, l.name)
ORDER BY amount_usd DESC, item_count DESC, label ASC
LIMIT 10;

-- name: ListDashboardExecutiveRiskItems :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.gb_code,
        gp.project_name,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id, gb_code, project_name
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
monitoring_by_la AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::text IS NULL OR md.quarter = sqlc.narg('quarter')::text)
    GROUP BY md.loan_agreement_id
),
la_base AS (
    SELECT
        la.id,
        la.loan_code,
        la.dk_project_id,
        la.effective_date,
        la.closing_date,
        la.amount_usd,
        dp.project_name,
        mbl.planned_usd,
        mbl.realized_usd,
        latest_bp.id AS journey_bb_project_id
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN monitoring_by_la mbl ON mbl.loan_agreement_id = la.id
    LEFT JOIN LATERAL (
        SELECT latest.id
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN bb_project latest ON latest.project_identity_id = bp.project_identity_id
        JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
        WHERE dpg.dk_project_id = dp.id
          AND latest.status = 'active'
        ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC, latest.created_at DESC
        LIMIT 1
    ) latest_bp ON TRUE
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
    AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
risks AS (
    SELECT
        'LA_CLOSING_12_MONTHS'::text AS risk_type,
        'high'::text AS severity,
        la_base.id AS reference_id,
        'loan_agreement'::text AS reference_type,
        la_base.journey_bb_project_id,
        la_base.loan_code::text AS code,
        la_base.project_name::text AS title,
        ('Loan Agreement ' || la_base.loan_code || ' closing pada ' || la_base.closing_date::text)::text AS description,
        la_base.amount_usd::numeric AS amount_usd,
        (la_base.closing_date - CURRENT_DATE)::int AS days_until_closing,
        CASE WHEN COALESCE(la_base.planned_usd, 0) = 0 THEN 0 ELSE COALESCE(la_base.realized_usd, 0) / la_base.planned_usd * 100 END::numeric AS absorption_pct,
        90::numeric AS score
    FROM la_base
    WHERE la_base.closing_date BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '12 months')::date
    UNION ALL
    SELECT
        'EFFECTIVE_LA_NO_MONITORING'::text AS risk_type,
        'high'::text AS severity,
        la_base.id AS reference_id,
        'loan_agreement'::text AS reference_type,
        la_base.journey_bb_project_id,
        la_base.loan_code::text AS code,
        la_base.project_name::text AS title,
        ('Loan Agreement ' || la_base.loan_code || ' sudah efektif tanpa monitoring')::text AS description,
        la_base.amount_usd::numeric AS amount_usd,
        0::int AS days_until_closing,
        0::numeric AS absorption_pct,
        85::numeric AS score
    FROM la_base
    WHERE la_base.effective_date <= CURRENT_DATE
      AND NOT EXISTS (SELECT 1 FROM monitoring_disbursement md WHERE md.loan_agreement_id = la_base.id)
    UNION ALL
    SELECT
        'HIGH_ELAPSED_LOW_ABSORPTION'::text AS risk_type,
        'medium'::text AS severity,
        la_base.id AS reference_id,
        'loan_agreement'::text AS reference_type,
        la_base.journey_bb_project_id,
        la_base.loan_code::text AS code,
        la_base.project_name::text AS title,
        ('Waktu berjalan tinggi tetapi serapan masih rendah untuk LA ' || la_base.loan_code)::text AS description,
        la_base.amount_usd::numeric AS amount_usd,
        0::int AS days_until_closing,
        CASE WHEN COALESCE(la_base.planned_usd, 0) = 0 THEN 0 ELSE COALESCE(la_base.realized_usd, 0) / la_base.planned_usd * 100 END::numeric AS absorption_pct,
        75::numeric AS score
    FROM la_base
    WHERE la_base.effective_date < CURRENT_DATE
      AND la_base.closing_date > la_base.effective_date
      AND ((CURRENT_DATE - la_base.effective_date)::numeric / NULLIF((la_base.closing_date - la_base.effective_date)::numeric, 0)) >= 0.7
      AND (CASE WHEN COALESCE(la_base.planned_usd, 0) = 0 THEN 0 ELSE COALESCE(la_base.realized_usd, 0) / la_base.planned_usd * 100 END) < 50
    UNION ALL
    SELECT
        'GB_WITHOUT_DK'::text AS risk_type,
        'medium'::text AS severity,
        sg.id AS reference_id,
        'gb_project'::text AS reference_type,
        latest_bp.id AS journey_bb_project_id,
        sg.gb_code::text AS code,
        sg.project_name::text AS title,
        ('Green Book project ' || sg.gb_code || ' belum masuk Daftar Kegiatan')::text AS description,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS amount_usd,
        0::int AS days_until_closing,
        0::numeric AS absorption_pct,
        60::numeric AS score
    FROM selected_gb sg
    LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    LEFT JOIN LATERAL (
        SELECT sb.id
        FROM gb_project_bb_project gbp
        JOIN selected_bb sb ON sb.id = gbp.bb_project_id
        WHERE gbp.gb_project_id = sg.id
        ORDER BY sb.id
        LIMIT 1
    ) latest_bp ON TRUE
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = sg.id)
    GROUP BY sg.id, sg.gb_code, sg.project_name, latest_bp.id
    UNION ALL
    SELECT
        'DK_WITHOUT_LA'::text AS risk_type,
        'medium'::text AS severity,
        dp.id AS reference_id,
        'dk_project'::text AS reference_type,
        latest_bp.id AS journey_bb_project_id,
        COALESCE(dk.letter_number, dp.id::text)::text AS code,
        dp.project_name::text AS title,
        ('Daftar Kegiatan project belum memiliki Loan Agreement')::text AS description,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS amount_usd,
        0::int AS days_until_closing,
        0::numeric AS absorption_pct,
        65::numeric AS score
    FROM dk_project dp
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
    JOIN dk_project_gb_project dpg ON dpg.dk_project_id = dp.id
    JOIN selected_gb sg ON sg.id = dpg.gb_project_id
    LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = dp.id
    LEFT JOIN LATERAL (
        SELECT sb.id
        FROM gb_project_bb_project gbp
        JOIN selected_bb sb ON sb.id = gbp.bb_project_id
        WHERE gbp.gb_project_id = sg.id
        ORDER BY sb.id
        LIMIT 1
    ) latest_bp ON TRUE
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id)
    GROUP BY dp.id, dk.letter_number, dp.project_name, latest_bp.id
)
SELECT
    risk_type,
    severity,
    reference_id,
    reference_type,
    journey_bb_project_id,
    code,
    title,
    description,
    amount_usd,
    days_until_closing,
    absorption_pct,
    score
FROM risks
ORDER BY score DESC, amount_usd DESC, title ASC
LIMIT 25;

-- name: GetDashboardPipelineBottleneckStageSummary :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id
    FROM ranked_gb
    WHERE rn = 1
),
worklist AS (
    SELECT
        'BB_NO_LENDER'::text AS stage,
        sb.id AS project_id,
        bp.project_name::text AS project_name,
        COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - bp.created_at::date), 0)::int AS age_days,
        bp.created_at AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name,
        ARRAY[]::text[] AS lender_names
    FROM selected_bb sb
    JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND sqlc.narg('lender_id')::uuid IS NULL
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'INDICATION_NO_LOI'::text AS stage,
        sb.id AS project_id,
        bp.project_name::text AS project_name,
        COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(li.created_at)::date FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int AS age_days,
        COALESCE((SELECT MAX(li.created_at) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at) AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id JOIN lender l ON l.id = li.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[]) AS lender_names
    FROM selected_bb sb
    JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND li.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'LOI_NO_GB'::text AS stage,
        sb.id AS project_id,
        bp.project_name::text AS project_name,
        COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(lo.date) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int AS age_days,
        COALESCE((SELECT MAX(lo.date)::timestamptz FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at) AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id JOIN lender l ON l.id = lo.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[]) AS lender_names
    FROM selected_bb sb
    JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND lo.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'GB_NO_DK'::text AS stage,
        sg.id AS project_id,
        gp.project_name::text AS project_name,
        COALESCE((SELECT SUM(gfs.loan_usd + gfs.grant_usd) FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - gp.created_at::date), 0)::int AS age_days,
        gp.created_at AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = gp.id AND gpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = gp.id), ARRAY[]::text[]) AS lender_names
    FROM selected_gb sg
    JOIN gb_project gp ON gp.id = sg.id
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = gp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid) OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'DK_NO_LA'::text AS stage,
        dp.id AS project_id,
        dp.project_name::text AS project_name,
        COALESCE((SELECT SUM(dfd.amount_usd + dfd.grant_usd) FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - COALESCE(dk.date, dp.created_at::date)), 0)::int AS age_days,
        COALESCE(dk.date::timestamptz, dp.created_at) AS relevant_at,
        COALESCE(i.short_name, i.name, '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM dk_financing_detail dfd JOIN lender l ON l.id = dfd.lender_id WHERE dfd.dk_project_id = dp.id), ARRAY[]::text[]) AS lender_names
    FROM dk_project dp
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
    LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id)
      AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id AND dfd.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT
        'LA_NOT_EFFECTIVE'::text AS stage,
        la.id AS project_id,
        dp.project_name::text AS project_name,
        la.amount_usd::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - la.agreement_date), 0)::int AS age_days,
        la.agreement_date::timestamptz AS relevant_at,
        COALESCE(i.short_name, i.name, '')::text AS institution_name,
        ARRAY[COALESCE(l.short_name, l.name)]::text[] AS lender_names
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    JOIN lender l ON l.id = la.lender_id
    LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date > CURRENT_DATE
      AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT
        'EFFECTIVE_NO_MONITORING'::text AS stage,
        la.id AS project_id,
        dp.project_name::text AS project_name,
        la.amount_usd::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - la.effective_date), 0)::int AS age_days,
        la.effective_date::timestamptz AS relevant_at,
        COALESCE(i.short_name, i.name, '')::text AS institution_name,
        ARRAY[COALESCE(l.short_name, l.name)]::text[] AS lender_names
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    JOIN lender l ON l.id = la.lender_id
    LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date <= CURRENT_DATE
      AND NOT EXISTS (SELECT 1 FROM monitoring_disbursement md WHERE md.loan_agreement_id = la.id)
      AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
filtered AS (
    SELECT *
    FROM worklist
    WHERE (sqlc.narg('stage')::text IS NULL OR stage = sqlc.narg('stage')::text)
      AND (sqlc.narg('min_age_days')::int IS NULL OR age_days >= sqlc.narg('min_age_days')::int)
      AND (
          sqlc.narg('search')::text IS NULL
          OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
          OR institution_name ILIKE '%' || sqlc.narg('search')::text || '%'
          OR EXISTS (SELECT 1 FROM unnest(lender_names) lender_name WHERE lender_name ILIKE '%' || sqlc.narg('search')::text || '%')
      )
)
SELECT
    stage,
    COUNT(*)::bigint AS project_count,
    COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd,
    COALESCE(AVG(age_days), 0)::numeric AS avg_age_days
FROM filtered
GROUP BY stage
ORDER BY CASE stage
    WHEN 'BB_NO_LENDER' THEN 1
    WHEN 'INDICATION_NO_LOI' THEN 2
    WHEN 'LOI_NO_GB' THEN 3
    WHEN 'GB_NO_DK' THEN 4
    WHEN 'DK_NO_LA' THEN 5
    WHEN 'LA_NOT_EFFECTIVE' THEN 6
    WHEN 'EFFECTIVE_NO_MONITORING' THEN 7
    ELSE 99
END;

-- name: CountDashboardPipelineBottleneckItems :one
WITH ranked_bb AS (
    SELECT bp.id, bp.project_identity_id, ROW_NUMBER() OVER (PARTITION BY bp.project_identity_id ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (SELECT id, project_identity_id FROM ranked_bb WHERE rn = 1),
ranked_gb AS (
    SELECT gp.id, gp.gb_project_identity_id, ROW_NUMBER() OVER (PARTITION BY gp.gb_project_identity_id ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (sqlc.narg('period_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_bb_project gbp JOIN bb_project bp ON bp.id = gbp.bb_project_id JOIN blue_book bb ON bb.id = bp.blue_book_id WHERE gbp.gb_project_id = gp.id AND bb.period_id = sqlc.narg('period_id')::uuid))
),
selected_gb AS (SELECT id, gb_project_identity_id FROM ranked_gb WHERE rn = 1),
worklist AS (
    SELECT 'BB_NO_LENDER'::text AS stage, sb.id AS project_id, bp.project_name::text AS project_name, GREATEST((CURRENT_DATE - bp.created_at::date), 0)::int AS age_days, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name, ARRAY[]::text[] AS lender_names
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND sqlc.narg('lender_id')::uuid IS NULL AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'INDICATION_NO_LOI'::text, sb.id, bp.project_name::text, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(li.created_at)::date FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id JOIN lender l ON l.id = li.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND li.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'LOI_NO_GB'::text, sb.id, bp.project_name::text, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(lo.date) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id JOIN lender l ON l.id = lo.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND lo.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'GB_NO_DK'::text, sg.id, gp.project_name::text, GREATEST((CURRENT_DATE - gp.created_at::date), 0)::int, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = gp.id AND gpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = gp.id), ARRAY[]::text[])
    FROM selected_gb sg JOIN gb_project gp ON gp.id = sg.id
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = gp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid) OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'DK_NO_LA'::text, dp.id, dp.project_name::text, GREATEST((CURRENT_DATE - COALESCE(dk.date, dp.created_at::date)), 0)::int, COALESCE(i.short_name, i.name, '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM dk_financing_detail dfd JOIN lender l ON l.id = dfd.lender_id WHERE dfd.dk_project_id = dp.id), ARRAY[]::text[])
    FROM dk_project dp JOIN daftar_kegiatan dk ON dk.id = dp.dk_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id) AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id AND dfd.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT 'LA_NOT_EFFECTIVE'::text, la.id, dp.project_name::text, GREATEST((CURRENT_DATE - la.agreement_date), 0)::int, COALESCE(i.short_name, i.name, '')::text, ARRAY[COALESCE(l.short_name, l.name)]::text[]
    FROM loan_agreement la JOIN dk_project dp ON dp.id = la.dk_project_id JOIN lender l ON l.id = la.lender_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date > CURRENT_DATE AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT 'EFFECTIVE_NO_MONITORING'::text, la.id, dp.project_name::text, GREATEST((CURRENT_DATE - la.effective_date), 0)::int, COALESCE(i.short_name, i.name, '')::text, ARRAY[COALESCE(l.short_name, l.name)]::text[]
    FROM loan_agreement la JOIN dk_project dp ON dp.id = la.dk_project_id JOIN lender l ON l.id = la.lender_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date <= CURRENT_DATE AND NOT EXISTS (SELECT 1 FROM monitoring_disbursement md WHERE md.loan_agreement_id = la.id) AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
filtered AS (
    SELECT * FROM worklist
    WHERE (sqlc.narg('stage')::text IS NULL OR stage = sqlc.narg('stage')::text)
      AND (sqlc.narg('min_age_days')::int IS NULL OR age_days >= sqlc.narg('min_age_days')::int)
      AND (sqlc.narg('search')::text IS NULL OR project_name ILIKE '%' || sqlc.narg('search')::text || '%' OR institution_name ILIKE '%' || sqlc.narg('search')::text || '%' OR EXISTS (SELECT 1 FROM unnest(lender_names) lender_name WHERE lender_name ILIKE '%' || sqlc.narg('search')::text || '%'))
)
SELECT COUNT(*)::bigint FROM filtered;

-- name: ListDashboardPipelineBottleneckItems :many
WITH ranked_bb AS (
    SELECT bp.id, bp.project_identity_id, ROW_NUMBER() OVER (PARTITION BY bp.project_identity_id ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (SELECT id, project_identity_id FROM ranked_bb WHERE rn = 1),
ranked_gb AS (
    SELECT gp.id, gp.gb_project_identity_id, ROW_NUMBER() OVER (PARTITION BY gp.gb_project_identity_id ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (sqlc.narg('period_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_bb_project gbp JOIN bb_project bp ON bp.id = gbp.bb_project_id JOIN blue_book bb ON bb.id = bp.blue_book_id WHERE gbp.gb_project_id = gp.id AND bb.period_id = sqlc.narg('period_id')::uuid))
),
selected_gb AS (SELECT id, gb_project_identity_id FROM ranked_gb WHERE rn = 1),
worklist AS (
    SELECT 'BB_NO_LENDER'::text AS stage, 'bb_project'::text AS reference_type, sb.id AS project_id, sb.id AS journey_bb_project_id, bp.bb_code::text AS code, bp.project_name::text AS project_name, COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd, GREATEST((CURRENT_DATE - bp.created_at::date), 0)::int AS age_days, bp.created_at AS relevant_at, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name, ARRAY[]::text[] AS lender_names
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND sqlc.narg('lender_id')::uuid IS NULL AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'INDICATION_NO_LOI'::text, 'bb_project'::text, sb.id, sb.id, bp.bb_code::text, bp.project_name::text, COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(li.created_at)::date FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT MAX(li.created_at) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at), COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id JOIN lender l ON l.id = li.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND li.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'LOI_NO_GB'::text, 'bb_project'::text, sb.id, sb.id, bp.bb_code::text, bp.project_name::text, COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(lo.date) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT MAX(lo.date)::timestamptz FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at), COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id JOIN lender l ON l.id = lo.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND lo.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'GB_NO_DK'::text, 'gb_project'::text, sg.id, COALESCE((SELECT sb.id FROM gb_project_bb_project gbp JOIN selected_bb sb ON sb.id = gbp.bb_project_id WHERE gbp.gb_project_id = gp.id LIMIT 1), (SELECT bp.id FROM gb_project_bb_project gbp JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE gbp.gb_project_id = gp.id LIMIT 1)), gp.gb_code::text, gp.project_name::text, COALESCE((SELECT SUM(gfs.loan_usd + gfs.grant_usd) FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id), 0)::numeric, GREATEST((CURRENT_DATE - gp.created_at::date), 0)::int, gp.created_at, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = gp.id AND gpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = gp.id), ARRAY[]::text[])
    FROM selected_gb sg JOIN gb_project gp ON gp.id = sg.id
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = gp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid) OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'DK_NO_LA'::text, 'dk_project'::text, dp.id, (SELECT bp.id FROM dk_project_gb_project dpg JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE dpg.dk_project_id = dp.id LIMIT 1), COALESCE(dk.letter_number, dp.id::text)::text, dp.project_name::text, COALESCE((SELECT SUM(dfd.amount_usd + dfd.grant_usd) FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id), 0)::numeric, GREATEST((CURRENT_DATE - COALESCE(dk.date, dp.created_at::date)), 0)::int, COALESCE(dk.date::timestamptz, dp.created_at), COALESCE(i.short_name, i.name, '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM dk_financing_detail dfd JOIN lender l ON l.id = dfd.lender_id WHERE dfd.dk_project_id = dp.id), ARRAY[]::text[])
    FROM dk_project dp JOIN daftar_kegiatan dk ON dk.id = dp.dk_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id) AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id AND dfd.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT 'LA_NOT_EFFECTIVE'::text, 'loan_agreement'::text, la.id, (SELECT bp.id FROM dk_project_gb_project dpg JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE dpg.dk_project_id = dp.id LIMIT 1), la.loan_code::text, dp.project_name::text, la.amount_usd::numeric, GREATEST((CURRENT_DATE - la.agreement_date), 0)::int, la.agreement_date::timestamptz, COALESCE(i.short_name, i.name, '')::text, ARRAY[COALESCE(l.short_name, l.name)]::text[]
    FROM loan_agreement la JOIN dk_project dp ON dp.id = la.dk_project_id JOIN lender l ON l.id = la.lender_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date > CURRENT_DATE AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT 'EFFECTIVE_NO_MONITORING'::text, 'loan_agreement'::text, la.id, (SELECT bp.id FROM dk_project_gb_project dpg JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE dpg.dk_project_id = dp.id LIMIT 1), la.loan_code::text, dp.project_name::text, la.amount_usd::numeric, GREATEST((CURRENT_DATE - la.effective_date), 0)::int, la.effective_date::timestamptz, COALESCE(i.short_name, i.name, '')::text, ARRAY[COALESCE(l.short_name, l.name)]::text[]
    FROM loan_agreement la JOIN dk_project dp ON dp.id = la.dk_project_id JOIN lender l ON l.id = la.lender_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date <= CURRENT_DATE AND NOT EXISTS (SELECT 1 FROM monitoring_disbursement md WHERE md.loan_agreement_id = la.id) AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
filtered AS (
    SELECT * FROM worklist
    WHERE (sqlc.narg('stage')::text IS NULL OR stage = sqlc.narg('stage')::text)
      AND (sqlc.narg('min_age_days')::int IS NULL OR age_days >= sqlc.narg('min_age_days')::int)
      AND (sqlc.narg('search')::text IS NULL OR project_name ILIKE '%' || sqlc.narg('search')::text || '%' OR code ILIKE '%' || sqlc.narg('search')::text || '%' OR institution_name ILIKE '%' || sqlc.narg('search')::text || '%' OR EXISTS (SELECT 1 FROM unnest(lender_names) lender_name WHERE lender_name ILIKE '%' || sqlc.narg('search')::text || '%'))
)
SELECT
    project_id,
    reference_type,
    journey_bb_project_id,
    code,
    project_name,
    stage AS current_stage,
    age_days,
    amount_usd,
    institution_name,
    lender_names,
    relevant_at
FROM filtered
ORDER BY
    CASE WHEN sqlc.arg('sort')::text = 'stage' AND sqlc.arg('order')::text = 'asc' THEN stage END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'stage' AND sqlc.arg('order')::text = 'desc' THEN stage END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'asc' THEN project_name END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'desc' THEN project_name END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'amount_usd' AND sqlc.arg('order')::text = 'asc' THEN amount_usd END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'amount_usd' AND sqlc.arg('order')::text = 'desc' THEN amount_usd END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'age_days' AND sqlc.arg('order')::text = 'asc' THEN age_days END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'age_days' AND sqlc.arg('order')::text = 'desc' THEN age_days END DESC,
    age_days DESC,
    amount_usd DESC,
    project_name ASC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int;
