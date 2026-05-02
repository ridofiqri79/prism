-- ===== DASHBOARD ANALYTICS FOUNDATION =====

-- name: GetDashboardAnalyticsPortfolioFoundation :one
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id,
        i.name AS root_institution_name,
        i.level AS root_institution_level
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id,
        parent.root_institution_name,
        parent.root_institution_level
    FROM institution child
    JOIN institution_rollup parent ON parent.source_institution_id = child.parent_id
),
latest_bb_project_rows AS (
    SELECT *
    FROM (
        SELECT
            bp.*,
            ROW_NUMBER() OVER (
                PARTITION BY bp.project_identity_id
                ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
            ) AS latest_rank
        FROM bb_project bp
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE bp.status = 'active'
    ) ranked
    WHERE sqlc.arg('include_history')::boolean OR latest_rank = 1
),
project_portfolio_rows AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.program_title_id,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Monitoring'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'LA'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'DK'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'GB'
            ELSE 'BB'
        END::text AS pipeline_status,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Ongoing'
            ELSE 'Pipeline'
        END::text AS project_status,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.grant_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Grant'
            ), 0)
        END::numeric AS foreign_grant_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.local_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Counterpart'
            ), 0)
        END::numeric AS counterpart_usd,
        ARRAY(
            SELECT DISTINCT type_label
            FROM (
                SELECT l.type AS type_label
                FROM lender_indication li
                JOIN lender l ON l.id = li.lender_id
                WHERE li.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                JOIN lender l ON l.id = gfs.lender_id
                WHERE gbp.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
                JOIN lender l ON l.id = la.lender_id
                WHERE gbp.bb_project_id = bp.id
            ) loan_types
            WHERE type_label IS NOT NULL
            ORDER BY type_label
        )::text[] AS loan_types,
        ARRAY(
            SELECT DISTINCT li.lender_id
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
            ORDER BY li.lender_id
        )::uuid[] AS indication_lender_ids,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT la.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN loan_agreement la ON la.dk_project_id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY la.lender_id
        )::uuid[] AS agreement_lender_ids,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT ir.root_institution_id
            FROM bb_project_institution bpi
            LEFT JOIN institution_rollup ir ON ir.source_institution_id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
              AND ir.root_institution_id IS NOT NULL
            ORDER BY ir.root_institution_id
        )::uuid[] AS executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids
    FROM latest_bb_project_rows bp
),
filtered_projects AS (
    SELECT *
    FROM project_portfolio_rows
    WHERE (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('lender_types')::text[])
      AND (
          COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0
          OR indication_lender_ids && sqlc.arg('lender_ids')::uuid[]
          OR fixed_lender_ids && sqlc.arg('lender_ids')::uuid[]
          OR agreement_lender_ids && sqlc.arg('lender_ids')::uuid[]
      )
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR executing_agency_ids && sqlc.arg('institution_ids')::uuid[]
          OR executing_agency_root_ids && sqlc.arg('institution_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
)
SELECT
    COUNT(*)::bigint AS project_count,
    COUNT(*) FILTER (WHERE project_status = 'Pipeline')::bigint AS pipeline_project_count,
    COUNT(*) FILTER (WHERE project_status = 'Ongoing')::bigint AS ongoing_project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS total_foreign_loan_usd,
    COALESCE(SUM(foreign_grant_usd), 0)::numeric AS total_grant_usd,
    COALESCE(SUM(counterpart_usd), 0)::numeric AS total_counterpart_usd
FROM filtered_projects;

-- name: GetDashboardAnalyticsMonitoringFoundation :one
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id,
        i.name AS root_institution_name,
        i.level AS root_institution_level
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id,
        parent.root_institution_name,
        parent.root_institution_level
    FROM institution child
    JOIN institution_rollup parent ON parent.source_institution_id = child.parent_id
),
monitoring_fact_rows AS (
    SELECT
        md.id AS monitoring_id,
        md.budget_year,
        md.quarter,
        md.planned_usd,
        md.realized_usd,
        la.id AS loan_agreement_id,
        la.amount_usd AS agreement_amount_usd,
        la.lender_id,
        l.type AS lender_type,
        dp.id AS dk_project_id,
        dp.institution_id,
        ir.root_institution_id
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
),
filtered_monitoring_facts AS (
    SELECT *
    FROM monitoring_fact_rows m
    WHERE (sqlc.narg('budget_year')::int IS NULL OR m.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR m.quarter = sqlc.narg('quarter')::varchar)
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR m.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR m.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR m.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR m.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR m.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR m.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              LEFT JOIN gb_project_bb_project gbp ON gbp.gb_project_id = gp.id
              LEFT JOIN bb_project bp ON bp.id = gbp.bb_project_id
              WHERE dpg.dk_project_id = m.dk_project_id
                AND (
                    gp.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
                    OR bp.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
                )
          )
      )
      AND (
          COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0
          OR EXISTS (
              SELECT 1
              FROM dk_project_location dpl
              WHERE dpl.dk_project_id = m.dk_project_id
                AND dpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE dpg.dk_project_id = m.dk_project_id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
),
unique_monitoring_loan_agreements AS (
    SELECT DISTINCT loan_agreement_id, agreement_amount_usd
    FROM filtered_monitoring_facts
)
SELECT
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(monitoring_id)::bigint AS monitoring_count,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd,
    COALESCE((SELECT SUM(agreement_amount_usd) FROM unique_monitoring_loan_agreements), 0)::numeric AS agreement_amount_usd
FROM filtered_monitoring_facts;

-- name: ListDashboardAnalyticsInstitutionRollup :many
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id,
        i.name AS root_institution_name,
        i.level AS root_institution_level
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id,
        parent.root_institution_name,
        parent.root_institution_level
    FROM institution child
    JOIN institution_rollup parent ON parent.source_institution_id = child.parent_id
)
SELECT
    ir.source_institution_id,
    i.name AS source_institution_name,
    i.level AS source_institution_level,
    ir.root_institution_id,
    ir.root_institution_name,
    ir.root_institution_level
FROM institution_rollup ir
JOIN institution i ON i.id = ir.source_institution_id
ORDER BY ir.root_institution_name ASC, i.name ASC;
