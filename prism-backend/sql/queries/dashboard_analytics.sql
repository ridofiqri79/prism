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

-- ===== DASHBOARD ANALYTICS DA-02 AGGREGATES =====

-- name: GetDashboardAnalyticsOverviewPortfolio :one
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
    COUNT(DISTINCT project_identity_id)::bigint AS project_count,
    COALESCE(SUM(cardinality(executing_agency_root_ids)), 0)::bigint AS assignment_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS total_pipeline_loan_usd
FROM filtered_projects;

-- name: ListDashboardAnalyticsPipelineFunnel :many
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
stages AS (
    SELECT 'BB'::text AS pipeline_stage, 1::int AS sort_order
    UNION ALL SELECT 'GB'::text, 2::int
    UNION ALL SELECT 'DK'::text, 3::int
    UNION ALL SELECT 'LA'::text, 4::int
    UNION ALL SELECT 'Monitoring'::text, 5::int
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
    s.pipeline_stage AS stage,
    COUNT(DISTINCT fp.project_identity_id)::bigint AS project_count,
    COALESCE(SUM(fp.foreign_loan_usd), 0)::numeric AS total_loan_usd
FROM stages s
LEFT JOIN filtered_projects fp ON fp.pipeline_status = s.pipeline_stage
GROUP BY s.pipeline_stage, s.sort_order
ORDER BY s.sort_order ASC;

-- name: GetDashboardAnalyticsAgreementPerformanceSummary :one
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
monitoring_by_loan_agreement AS (
    SELECT
        md.loan_agreement_id,
        COUNT(md.id)::bigint AS monitoring_count,
        COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
    GROUP BY md.loan_agreement_id
),
loan_agreement_rows AS (
    SELECT
        la.id AS loan_agreement_id,
        la.dk_project_id,
        dp.project_name,
        dp.program_title_id,
        la.lender_id,
        l.name AS lender_name,
        l.short_name AS lender_short_name,
        l.type AS lender_type,
        la.amount_usd AS agreement_amount_usd,
        dp.institution_id,
        ir.root_institution_id,
        ir.root_institution_name,
        ir.root_institution_level,
        COALESCE(m.monitoring_count, 0)::bigint AS monitoring_count,
        COALESCE(m.planned_usd, 0)::numeric AS planned_usd,
        COALESCE(m.realized_usd, 0)::numeric AS realized_usd
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    LEFT JOIN monitoring_by_loan_agreement m ON m.loan_agreement_id = la.id
    WHERE (
        (sqlc.narg('budget_year')::int IS NULL AND sqlc.narg('quarter')::varchar IS NULL)
        OR COALESCE(m.monitoring_count, 0) > 0
    )
),
filtered_loan_agreements AS (
    SELECT *
    FROM loan_agreement_rows r
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR r.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR r.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR r.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR r.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (
          COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0
          OR 'LA' = ANY(sqlc.arg('pipeline_statuses')::text[])
          OR (r.monitoring_count > 0 AND 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      )
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR r.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR r.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR r.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              LEFT JOIN gb_project_bb_project gbp ON gbp.gb_project_id = gp.id
              LEFT JOIN bb_project bp ON bp.id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
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
              WHERE dpl.dk_project_id = r.dk_project_id
                AND dpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
)
SELECT
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count,
    COALESCE(SUM(monitoring_count), 0)::bigint AS monitoring_count,
    COALESCE(SUM(agreement_amount_usd), 0)::numeric AS total_agreement_amount_usd,
    COALESCE(SUM(planned_usd), 0)::numeric AS total_planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS total_realized_usd
FROM filtered_loan_agreements;

-- name: ListDashboardAnalyticsInstitutions :many
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
),
portfolio_by_institution AS (
    SELECT
        ea.root_institution_id,
        COUNT(DISTINCT fp.project_identity_id)::bigint AS portfolio_project_count,
        COUNT(*)::bigint AS portfolio_assignment_count,
        COUNT(DISTINCT fp.project_identity_id) FILTER (WHERE fp.pipeline_status = 'BB')::bigint AS bb_count,
        COUNT(DISTINCT fp.project_identity_id) FILTER (WHERE fp.pipeline_status = 'GB')::bigint AS gb_count,
        COUNT(DISTINCT fp.project_identity_id) FILTER (WHERE fp.pipeline_status = 'DK')::bigint AS dk_count,
        COUNT(DISTINCT fp.project_identity_id) FILTER (WHERE fp.pipeline_status = 'LA')::bigint AS la_count,
        COUNT(DISTINCT fp.project_identity_id) FILTER (WHERE fp.pipeline_status = 'Monitoring')::bigint AS monitoring_pipeline_count
    FROM filtered_projects fp
    CROSS JOIN LATERAL unnest(fp.executing_agency_root_ids) AS ea(root_institution_id)
    GROUP BY ea.root_institution_id
),
monitoring_by_loan_agreement AS (
    SELECT
        md.loan_agreement_id,
        COUNT(md.id)::bigint AS monitoring_count,
        COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
    GROUP BY md.loan_agreement_id
),
loan_agreement_rows AS (
    SELECT
        la.id AS loan_agreement_id,
        la.dk_project_id,
        dp.project_name,
        dp.program_title_id,
        la.lender_id,
        l.type AS lender_type,
        la.amount_usd AS agreement_amount_usd,
        dp.institution_id,
        ir.root_institution_id,
        COALESCE(m.monitoring_count, 0)::bigint AS monitoring_count,
        COALESCE(m.planned_usd, 0)::numeric AS planned_usd,
        COALESCE(m.realized_usd, 0)::numeric AS realized_usd
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    LEFT JOIN monitoring_by_loan_agreement m ON m.loan_agreement_id = la.id
    WHERE (
        (sqlc.narg('budget_year')::int IS NULL AND sqlc.narg('quarter')::varchar IS NULL)
        OR COALESCE(m.monitoring_count, 0) > 0
    )
),
filtered_loan_agreements AS (
    SELECT *
    FROM loan_agreement_rows r
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR r.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR r.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR r.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR r.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (
          COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0
          OR 'LA' = ANY(sqlc.arg('pipeline_statuses')::text[])
          OR (r.monitoring_count > 0 AND 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      )
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR r.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR r.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR r.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              LEFT JOIN gb_project_bb_project gbp ON gbp.gb_project_id = gp.id
              LEFT JOIN bb_project bp ON bp.id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
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
              WHERE dpl.dk_project_id = r.dk_project_id
                AND dpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
),
performance_by_institution AS (
    SELECT
        root_institution_id,
        COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
        COUNT(DISTINCT dk_project_id)::bigint AS monitoring_project_count,
        COALESCE(SUM(monitoring_count), 0)::bigint AS monitoring_count,
        COALESCE(SUM(agreement_amount_usd), 0)::numeric AS agreement_amount_usd,
        COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd
    FROM filtered_loan_agreements
    WHERE root_institution_id IS NOT NULL
    GROUP BY root_institution_id
),
institution_ids AS (
    SELECT root_institution_id FROM portfolio_by_institution
    UNION
    SELECT root_institution_id FROM performance_by_institution
)
SELECT
    i.id AS institution_id,
    i.name AS institution_name,
    i.short_name AS institution_short_name,
    i.level AS institution_level,
    COALESCE(p.portfolio_project_count, 0)::bigint AS portfolio_project_count,
    COALESCE(p.portfolio_assignment_count, 0)::bigint AS portfolio_assignment_count,
    COALESCE(perf.loan_agreement_count, 0)::bigint AS loan_agreement_count,
    COALESCE(perf.monitoring_project_count, 0)::bigint AS monitoring_project_count,
    COALESCE(perf.monitoring_count, 0)::bigint AS monitoring_count,
    COALESCE(perf.agreement_amount_usd, 0)::numeric AS agreement_amount_usd,
    COALESCE(perf.planned_usd, 0)::numeric AS planned_usd,
    COALESCE(perf.realized_usd, 0)::numeric AS realized_usd,
    COALESCE(p.bb_count, 0)::bigint AS bb_count,
    COALESCE(p.gb_count, 0)::bigint AS gb_count,
    COALESCE(p.dk_count, 0)::bigint AS dk_count,
    COALESCE(p.la_count, 0)::bigint AS la_count,
    COALESCE(p.monitoring_pipeline_count, 0)::bigint AS monitoring_pipeline_count
FROM institution_ids ids
JOIN institution i ON i.id = ids.root_institution_id
LEFT JOIN portfolio_by_institution p ON p.root_institution_id = ids.root_institution_id
LEFT JOIN performance_by_institution perf ON perf.root_institution_id = ids.root_institution_id
ORDER BY agreement_amount_usd DESC, portfolio_project_count DESC, i.name ASC;

-- name: ListDashboardAnalyticsLenders :many
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
monitoring_by_loan_agreement AS (
    SELECT
        md.loan_agreement_id,
        COUNT(md.id)::bigint AS monitoring_count,
        COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
    GROUP BY md.loan_agreement_id
),
loan_agreement_rows AS (
    SELECT
        la.id AS loan_agreement_id,
        la.dk_project_id,
        dp.project_name,
        dp.program_title_id,
        la.lender_id,
        l.name AS lender_name,
        l.short_name AS lender_short_name,
        l.type AS lender_type,
        la.amount_usd AS agreement_amount_usd,
        dp.institution_id,
        ir.root_institution_id,
        COALESCE(m.monitoring_count, 0)::bigint AS monitoring_count,
        COALESCE(m.planned_usd, 0)::numeric AS planned_usd,
        COALESCE(m.realized_usd, 0)::numeric AS realized_usd
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    LEFT JOIN monitoring_by_loan_agreement m ON m.loan_agreement_id = la.id
    WHERE (
        (sqlc.narg('budget_year')::int IS NULL AND sqlc.narg('quarter')::varchar IS NULL)
        OR COALESCE(m.monitoring_count, 0) > 0
    )
),
filtered_loan_agreements AS (
    SELECT *
    FROM loan_agreement_rows r
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR r.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR r.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR r.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR r.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (
          COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0
          OR 'LA' = ANY(sqlc.arg('pipeline_statuses')::text[])
          OR (r.monitoring_count > 0 AND 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      )
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR r.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR r.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR r.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              LEFT JOIN gb_project_bb_project gbp ON gbp.gb_project_id = gp.id
              LEFT JOIN bb_project bp ON bp.id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
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
              WHERE dpl.dk_project_id = r.dk_project_id
                AND dpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
)
SELECT
    lender_id,
    lender_name,
    lender_short_name,
    lender_type,
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count,
    COUNT(DISTINCT root_institution_id) FILTER (WHERE root_institution_id IS NOT NULL)::bigint AS institution_count,
    COALESCE(SUM(monitoring_count), 0)::bigint AS monitoring_count,
    COALESCE(SUM(agreement_amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd
FROM filtered_loan_agreements
GROUP BY lender_id, lender_name, lender_short_name, lender_type
ORDER BY agreement_amount_usd DESC, lender_name ASC;

-- name: ListDashboardAnalyticsLenderInstitutionMatrix :many
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id,
        i.name AS root_institution_name,
        i.short_name AS root_institution_short_name,
        i.level AS root_institution_level
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id,
        parent.root_institution_name,
        parent.root_institution_short_name,
        parent.root_institution_level
    FROM institution child
    JOIN institution_rollup parent ON parent.source_institution_id = child.parent_id
),
monitoring_by_loan_agreement AS (
    SELECT
        md.loan_agreement_id,
        COUNT(md.id)::bigint AS monitoring_count,
        COALESCE(SUM(md.planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
    GROUP BY md.loan_agreement_id
),
loan_agreement_rows AS (
    SELECT
        la.id AS loan_agreement_id,
        la.dk_project_id,
        dp.project_name,
        dp.program_title_id,
        la.lender_id,
        l.name AS lender_name,
        l.short_name AS lender_short_name,
        l.type AS lender_type,
        la.amount_usd AS agreement_amount_usd,
        dp.institution_id,
        ir.root_institution_id,
        ir.root_institution_name,
        ir.root_institution_short_name,
        ir.root_institution_level,
        COALESCE(m.monitoring_count, 0)::bigint AS monitoring_count,
        COALESCE(m.planned_usd, 0)::numeric AS planned_usd,
        COALESCE(m.realized_usd, 0)::numeric AS realized_usd
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    LEFT JOIN monitoring_by_loan_agreement m ON m.loan_agreement_id = la.id
    WHERE (
        (sqlc.narg('budget_year')::int IS NULL AND sqlc.narg('quarter')::varchar IS NULL)
        OR COALESCE(m.monitoring_count, 0) > 0
    )
),
filtered_loan_agreements AS (
    SELECT *
    FROM loan_agreement_rows r
    WHERE r.root_institution_id IS NOT NULL
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR r.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR r.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR r.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR r.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (
          COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0
          OR 'LA' = ANY(sqlc.arg('pipeline_statuses')::text[])
          OR (r.monitoring_count > 0 AND 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      )
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR r.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR r.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR r.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              LEFT JOIN gb_project_bb_project gbp ON gbp.gb_project_id = gp.id
              LEFT JOIN bb_project bp ON bp.id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
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
              WHERE dpl.dk_project_id = r.dk_project_id
                AND dpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE dpg.dk_project_id = r.dk_project_id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
)
SELECT
    root_institution_id AS institution_id,
    root_institution_name AS institution_name,
    root_institution_short_name AS institution_short_name,
    root_institution_level AS institution_level,
    lender_id,
    lender_name,
    lender_short_name,
    lender_type,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count,
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COALESCE(SUM(monitoring_count), 0)::bigint AS monitoring_count,
    COALESCE(SUM(agreement_amount_usd), 0)::numeric AS agreement_amount_usd,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd
FROM filtered_loan_agreements
GROUP BY root_institution_id, root_institution_name, root_institution_short_name, root_institution_level, lender_id, lender_name, lender_short_name, lender_type
ORDER BY agreement_amount_usd DESC, root_institution_name ASC, lender_name ASC;

-- name: ListDashboardAnalyticsAbsorptionByInstitution :many
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id,
        i.name AS root_institution_name,
        i.short_name AS root_institution_short_name,
        i.level AS root_institution_level
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id,
        parent.root_institution_name,
        parent.root_institution_short_name,
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
        dp.project_name AS dk_project_name,
        dp.program_title_id,
        dp.institution_id,
        ir.root_institution_id,
        ir.root_institution_name,
        ir.root_institution_short_name,
        ir.root_institution_level
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
),
filtered_monitoring_facts AS (
    SELECT *
    FROM monitoring_fact_rows m
    WHERE m.root_institution_id IS NOT NULL
      AND (sqlc.narg('budget_year')::int IS NULL OR m.budget_year = sqlc.narg('budget_year')::int)
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
          OR m.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
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
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = m.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
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
)
SELECT
    root_institution_id AS id,
    root_institution_name AS name,
    'institution'::text AS dimension,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd,
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count
FROM filtered_monitoring_facts
GROUP BY root_institution_id, root_institution_name
ORDER BY planned_usd DESC, realized_usd DESC, root_institution_name ASC;

-- name: ListDashboardAnalyticsAbsorptionByProject :many
WITH monitoring_fact_rows AS (
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
        dp.project_name AS dk_project_name,
        dp.program_title_id,
        dp.institution_id
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
),
filtered_monitoring_facts AS (
    SELECT *
    FROM monitoring_fact_rows m
    WHERE (sqlc.narg('budget_year')::int IS NULL OR m.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR m.quarter = sqlc.narg('quarter')::varchar)
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR m.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR m.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR m.institution_id = ANY(sqlc.arg('institution_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR m.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR m.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR m.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
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
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = m.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
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
)
SELECT
    dk_project_id AS id,
    dk_project_name AS name,
    'project'::text AS dimension,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd,
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count
FROM filtered_monitoring_facts
GROUP BY dk_project_id, dk_project_name
ORDER BY planned_usd DESC, realized_usd DESC, dk_project_name ASC;

-- name: ListDashboardAnalyticsAbsorptionByLender :many
WITH monitoring_fact_rows AS (
    SELECT
        md.id AS monitoring_id,
        md.budget_year,
        md.quarter,
        md.planned_usd,
        md.realized_usd,
        la.id AS loan_agreement_id,
        la.amount_usd AS agreement_amount_usd,
        la.lender_id,
        l.name AS lender_name,
        l.type AS lender_type,
        dp.id AS dk_project_id,
        dp.program_title_id,
        dp.institution_id
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
),
filtered_monitoring_facts AS (
    SELECT *
    FROM monitoring_fact_rows m
    WHERE (sqlc.narg('budget_year')::int IS NULL OR m.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR m.quarter = sqlc.narg('quarter')::varchar)
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR m.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR m.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0 OR m.institution_id = ANY(sqlc.arg('institution_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR m.agreement_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR m.agreement_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0
          OR m.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
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
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = m.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
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
)
SELECT
    lender_id AS id,
    lender_name AS name,
    'lender'::text AS dimension,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd,
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count
FROM filtered_monitoring_facts
GROUP BY lender_id, lender_name
ORDER BY planned_usd DESC, realized_usd DESC, lender_name ASC;

-- name: ListDashboardAnalyticsYearlyPerformance :many
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id
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
        dp.program_title_id,
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
          OR m.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[])
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
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = m.dk_project_id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
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
)
SELECT
    budget_year,
    quarter,
    COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
    COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd,
    COUNT(DISTINCT loan_agreement_id)::bigint AS loan_agreement_count,
    COUNT(DISTINCT dk_project_id)::bigint AS project_count
FROM filtered_monitoring_facts
GROUP BY budget_year, quarter
ORDER BY budget_year ASC,
    CASE quarter
        WHEN 'TW1' THEN 1
        WHEN 'TW2' THEN 2
        WHEN 'TW3' THEN 3
        WHEN 'TW4' THEN 4
        ELSE 5
    END ASC;

-- name: ListDashboardAnalyticsLenderProportion :many
WITH RECURSIVE institution_rollup AS (
    SELECT
        i.id AS source_institution_id,
        i.id AS root_institution_id
    FROM institution i
    WHERE i.parent_id IS NULL

    UNION ALL

    SELECT
        child.id AS source_institution_id,
        parent.root_institution_id
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
latest_gb_project_rows AS (
    SELECT *
    FROM (
        SELECT
            gp.*,
            ROW_NUMBER() OVER (
                PARTITION BY gp.gb_project_identity_id
                ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
            ) AS latest_rank
        FROM gb_project gp
        JOIN green_book gb ON gb.id = gp.green_book_id
        WHERE gp.status = 'active'
    ) ranked
    WHERE sqlc.arg('include_history')::boolean OR latest_rank = 1
),
bb_project_amounts AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.program_title_id,
        COALESCE(SUM(pc.amount_usd) FILTER (WHERE pc.funding_type = 'Foreign' AND pc.funding_category = 'Loan'), 0)::numeric AS amount_usd,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT ir.root_institution_id
            FROM bb_project_institution bpi
            LEFT JOIN institution_rollup ir ON ir.source_institution_id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
              AND ir.root_institution_id IS NOT NULL
        )::uuid[] AS executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
        )::uuid[] AS region_ids
    FROM latest_bb_project_rows bp
    LEFT JOIN bb_project_cost pc ON pc.bb_project_id = bp.id
    GROUP BY bp.id, bp.project_identity_id, bp.program_title_id
),
indication_rows AS (
    SELECT
        'Lender Indication'::text AS stage,
        l.type AS lender_type,
        bpa.project_identity_id::uuid AS project_key,
        li.lender_id,
        bpa.amount_usd
    FROM bb_project_amounts bpa
    JOIN lender_indication li ON li.bb_project_id = bpa.id
    JOIN lender l ON l.id = li.lender_id
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR li.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR bpa.executing_agency_ids && sqlc.arg('institution_ids')::uuid[]
          OR bpa.executing_agency_root_ids && sqlc.arg('institution_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'BB' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Pipeline' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR bpa.region_ids && sqlc.arg('region_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR bpa.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR bpa.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR bpa.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
),
funding_source_rows AS (
    SELECT
        'Green Book Funding Source'::text AS stage,
        l.type AS lender_type,
        gp.gb_project_identity_id::uuid AS project_key,
        gfs.lender_id,
        gfs.loan_usd AS amount_usd
    FROM latest_gb_project_rows gp
    JOIN gb_funding_source gfs ON gfs.gb_project_id = gp.id
    JOIN lender l ON l.id = gfs.lender_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = gfs.institution_id
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR gfs.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR gfs.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'GB' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Pipeline' = ANY(sqlc.arg('project_statuses')::text[]) OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR gp.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR gfs.loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR gfs.loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0
          OR EXISTS (
              SELECT 1
              FROM gb_project_location gpl
              WHERE gpl.gb_project_id = gp.id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE gbp.gb_project_id = gp.id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
),
monitoring_by_loan_agreement AS (
    SELECT
        md.loan_agreement_id,
        COALESCE(SUM(md.realized_usd), 0)::numeric AS realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
    GROUP BY md.loan_agreement_id
),
loan_agreement_rows AS (
    SELECT
        'Loan Agreement'::text AS stage,
        l.type AS lender_type,
        dp.id::uuid AS project_key,
        la.lender_id,
        la.amount_usd AS amount_usd,
        la.id AS loan_agreement_id,
        dp.program_title_id,
        dp.institution_id,
        ir.root_institution_id,
        COALESCE(m.realized_usd, 0)::numeric AS realized_usd
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    LEFT JOIN monitoring_by_loan_agreement m ON m.loan_agreement_id = la.id
    WHERE (
        (sqlc.narg('budget_year')::int IS NULL AND sqlc.narg('quarter')::varchar IS NULL)
        OR m.loan_agreement_id IS NOT NULL
    )
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR dp.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'LA' = ANY(sqlc.arg('pipeline_statuses')::text[]) OR 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR la.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR la.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR dp.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (
          COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0
          OR EXISTS (
              SELECT 1
              FROM dk_project_location dpl
              WHERE dpl.dk_project_id = dp.id
                AND dpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
),
stage_rows AS (
    SELECT stage, lender_type, project_key, lender_id, amount_usd FROM indication_rows
    UNION ALL
    SELECT stage, lender_type, project_key, lender_id, amount_usd FROM funding_source_rows
    UNION ALL
    SELECT stage, lender_type, project_key, lender_id, amount_usd FROM loan_agreement_rows
    UNION ALL
    SELECT
        'Monitoring Realization'::text AS stage,
        lender_type,
        project_key,
        lender_id,
        realized_usd AS amount_usd
    FROM loan_agreement_rows
    WHERE realized_usd > 0
)
SELECT
    stage,
    lender_type,
    COUNT(DISTINCT project_key)::bigint AS project_count,
    COUNT(DISTINCT lender_id)::bigint AS lender_count,
    COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM stage_rows
GROUP BY stage, lender_type
ORDER BY
    CASE stage
        WHEN 'Lender Indication' THEN 1
        WHEN 'Green Book Funding Source' THEN 2
        WHEN 'Loan Agreement' THEN 3
        WHEN 'Monitoring Realization' THEN 4
        ELSE 5
    END,
    lender_type ASC;
