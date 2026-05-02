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

-- ===== DASHBOARD ANALYTICS DA-03 RISK AND DATA QUALITY =====

-- name: ListDashboardAnalyticsRiskWatchlist :many
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
filter_state AS (
    SELECT
        (sqlc.narg('budget_year')::int IS NOT NULL OR sqlc.narg('quarter')::varchar IS NOT NULL)::boolean AS has_monitoring_period_filter
),
current_budget_period AS (
    SELECT
        CASE
            WHEN EXTRACT(MONTH FROM CURRENT_DATE)::int BETWEEN 4 AND 6 THEN EXTRACT(YEAR FROM CURRENT_DATE)::int * 4 + 1
            WHEN EXTRACT(MONTH FROM CURRENT_DATE)::int BETWEEN 7 AND 9 THEN EXTRACT(YEAR FROM CURRENT_DATE)::int * 4 + 2
            WHEN EXTRACT(MONTH FROM CURRENT_DATE)::int BETWEEN 10 AND 12 THEN EXTRACT(YEAR FROM CURRENT_DATE)::int * 4 + 3
            ELSE (EXTRACT(YEAR FROM CURRENT_DATE)::int - 1) * 4 + 4
        END AS period_index
),
monitoring_filtered_rows AS (
    SELECT
        md.loan_agreement_id,
        md.budget_year,
        md.quarter,
        CASE md.quarter
            WHEN 'TW1' THEN 1
            WHEN 'TW2' THEN 2
            WHEN 'TW3' THEN 3
            WHEN 'TW4' THEN 4
            ELSE 0
        END AS quarter_index,
        md.planned_usd,
        md.realized_usd
    FROM monitoring_disbursement md
    WHERE (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
),
monitoring_by_filter AS (
    SELECT
        loan_agreement_id,
        COALESCE(SUM(planned_usd), 0)::numeric AS planned_usd,
        COALESCE(SUM(realized_usd), 0)::numeric AS realized_usd,
        COUNT(*)::bigint AS monitoring_count,
        (ARRAY_AGG(budget_year ORDER BY budget_year DESC, quarter_index DESC))[1]::int AS latest_budget_year,
        (ARRAY_AGG(quarter ORDER BY budget_year DESC, quarter_index DESC))[1]::varchar AS latest_quarter
    FROM monitoring_filtered_rows
    GROUP BY loan_agreement_id
),
monitoring_all AS (
    SELECT
        md.loan_agreement_id,
        COUNT(*)::bigint AS monitoring_count,
        MAX(
            md.budget_year * 4
            + CASE md.quarter
                WHEN 'TW1' THEN 1
                WHEN 'TW2' THEN 2
                WHEN 'TW3' THEN 3
                WHEN 'TW4' THEN 4
                ELSE 0
            END
        )::int AS latest_period_index
    FROM monitoring_disbursement md
    GROUP BY md.loan_agreement_id
),
loan_agreement_fact_rows AS (
    SELECT
        la.id AS loan_agreement_id,
        la.loan_code,
        la.effective_date,
        la.original_closing_date,
        la.closing_date,
        la.amount_usd AS agreement_amount_usd,
        la.lender_id,
        l.name AS lender_name,
        l.short_name AS lender_short_name,
        l.type AS lender_type,
        dp.id AS dk_project_id,
        dp.project_name,
        dp.program_title_id,
        dp.institution_id,
        i.name AS institution_name,
        i.short_name AS institution_short_name,
        i.level AS institution_level,
        ir.root_institution_id,
        COALESCE(mf.planned_usd, 0)::numeric AS planned_usd,
        COALESCE(mf.realized_usd, 0)::numeric AS realized_usd,
        COALESCE(mf.monitoring_count, 0)::bigint AS filtered_monitoring_count,
        COALESCE(ma.monitoring_count, 0)::bigint AS all_monitoring_count,
        mf.latest_budget_year,
        mf.latest_quarter,
        CASE
            WHEN COALESCE(mf.planned_usd, 0) = 0 THEN 0::double precision
            ELSE (COALESCE(mf.realized_usd, 0)::double precision / NULLIF(mf.planned_usd::double precision, 0)) * 100
        END AS absorption_pct,
        GREATEST((SELECT period_index FROM current_budget_period) - COALESCE(ma.latest_period_index, (SELECT period_index FROM current_budget_period)), 0)::int AS stale_quarters
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution i ON i.id = dp.institution_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    LEFT JOIN monitoring_by_filter mf ON mf.loan_agreement_id = la.id
    LEFT JOIN monitoring_all ma ON ma.loan_agreement_id = la.id
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
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
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_location gpl ON gpl.gb_project_id = dpg.gb_project_id
              WHERE dpg.dk_project_id = dp.id
                AND gpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project_location bpl ON bpl.bb_project_id = gbp.bb_project_id
              WHERE dpg.dk_project_id = dp.id
                AND bpl.region_id = ANY(sqlc.arg('region_ids')::uuid[])
          )
      )
),
risk_rows AS (
    SELECT
        1::int AS risk_order,
        'LOW_ABSORPTION'::text AS risk_code,
        'Penyerapan rendah'::text AS risk_label,
        'warning'::text AS severity,
        f.*,
        f.latest_budget_year AS budget_year,
        f.latest_quarter AS quarter,
        'has_monitoring'::text AS monitoring_status,
        (CURRENT_DATE - f.effective_date)::int AS days_since_effective,
        (f.closing_date - CURRENT_DATE)::int AS days_to_closing,
        GREATEST(CEIL((f.closing_date - CURRENT_DATE)::double precision / 30.0), 0)::int AS months_to_closing,
        (f.closing_date - f.original_closing_date)::int AS extension_days
    FROM loan_agreement_fact_rows f
    WHERE f.filtered_monitoring_count > 0
      AND f.planned_usd > 0
      AND f.absorption_pct < sqlc.arg('low_absorption_threshold')::double precision

    UNION ALL

    SELECT
        2::int AS risk_order,
        'EFFECTIVE_WITHOUT_MONITORING'::text AS risk_code,
        'Loan Agreement efektif tanpa monitoring terkini'::text AS risk_label,
        'danger'::text AS severity,
        f.*,
        CASE WHEN (SELECT has_monitoring_period_filter FROM filter_state) THEN sqlc.narg('budget_year')::int ELSE f.latest_budget_year END AS budget_year,
        CASE WHEN (SELECT has_monitoring_period_filter FROM filter_state) THEN sqlc.narg('quarter')::varchar ELSE f.latest_quarter END AS quarter,
        CASE
            WHEN (SELECT has_monitoring_period_filter FROM filter_state) THEN 'period_missing'
            WHEN f.all_monitoring_count = 0 THEN 'no_monitoring'
            ELSE 'stale_monitoring'
        END::text AS monitoring_status,
        (CURRENT_DATE - f.effective_date)::int AS days_since_effective,
        (f.closing_date - CURRENT_DATE)::int AS days_to_closing,
        GREATEST(CEIL((f.closing_date - CURRENT_DATE)::double precision / 30.0), 0)::int AS months_to_closing,
        (f.closing_date - f.original_closing_date)::int AS extension_days
    FROM loan_agreement_fact_rows f
    WHERE f.effective_date <= CURRENT_DATE
      AND (
          ((SELECT has_monitoring_period_filter FROM filter_state) AND f.filtered_monitoring_count = 0)
          OR (
              NOT (SELECT has_monitoring_period_filter FROM filter_state)
              AND (f.all_monitoring_count = 0 OR f.stale_quarters >= sqlc.arg('stale_monitoring_quarters')::int)
          )
      )

    UNION ALL

    SELECT
        3::int AS risk_order,
        'CLOSING_RISK'::text AS risk_code,
        'Closing risk'::text AS risk_label,
        'danger'::text AS severity,
        f.*,
        f.latest_budget_year AS budget_year,
        f.latest_quarter AS quarter,
        CASE WHEN f.filtered_monitoring_count > 0 THEN 'has_monitoring' ELSE 'no_monitoring' END::text AS monitoring_status,
        (CURRENT_DATE - f.effective_date)::int AS days_since_effective,
        (f.closing_date - CURRENT_DATE)::int AS days_to_closing,
        GREATEST(CEIL((f.closing_date - CURRENT_DATE)::double precision / 30.0), 0)::int AS months_to_closing,
        (f.closing_date - f.original_closing_date)::int AS extension_days
    FROM loan_agreement_fact_rows f
    WHERE f.effective_date <= CURRENT_DATE
      AND f.closing_date <= (CURRENT_DATE + make_interval(months => sqlc.arg('closing_months_threshold')::int))::date
      AND f.absorption_pct < 80::double precision
      AND (
          NOT (SELECT has_monitoring_period_filter FROM filter_state)
          OR f.filtered_monitoring_count > 0
      )

    UNION ALL

    SELECT
        4::int AS risk_order,
        'EXTENDED_LOAN'::text AS risk_code,
        'Loan Agreement diperpanjang'::text AS risk_label,
        'info'::text AS severity,
        f.*,
        f.latest_budget_year AS budget_year,
        f.latest_quarter AS quarter,
        CASE WHEN f.filtered_monitoring_count > 0 THEN 'has_monitoring' ELSE 'no_monitoring' END::text AS monitoring_status,
        (CURRENT_DATE - f.effective_date)::int AS days_since_effective,
        (f.closing_date - CURRENT_DATE)::int AS days_to_closing,
        GREATEST(CEIL((f.closing_date - CURRENT_DATE)::double precision / 30.0), 0)::int AS months_to_closing,
        (f.closing_date - f.original_closing_date)::int AS extension_days
    FROM loan_agreement_fact_rows f
    WHERE f.closing_date > f.original_closing_date
)
SELECT
    risk_code,
    risk_label,
    severity,
    dk_project_id,
    project_name,
    loan_agreement_id,
    loan_code,
    lender_id,
    lender_name,
    lender_short_name,
    lender_type,
    institution_id,
    institution_name,
    institution_short_name,
    institution_level,
    effective_date,
    closing_date,
    original_closing_date,
    budget_year,
    quarter,
    planned_usd,
    realized_usd,
    absorption_pct::double precision AS absorption_pct,
    agreement_amount_usd,
    days_since_effective,
    days_to_closing,
    months_to_closing,
    extension_days,
    stale_quarters,
    monitoring_status
FROM risk_rows
ORDER BY
    risk_order ASC,
    CASE WHEN risk_code = 'LOW_ABSORPTION' THEN absorption_pct END ASC,
    CASE WHEN risk_code = 'LOW_ABSORPTION' THEN planned_usd END DESC,
    CASE WHEN risk_code = 'CLOSING_RISK' THEN days_to_closing END ASC,
    CASE WHEN risk_code = 'EXTENDED_LOAN' THEN extension_days END DESC,
    project_name ASC,
    loan_code ASC;

-- name: ListDashboardAnalyticsPipelineBottlenecks :many
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
bb_rows AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.program_title_id,
        bp.created_at,
        COALESCE((
            SELECT SUM(pc.amount_usd)
            FROM bb_project_cost pc
            WHERE pc.bb_project_id = bp.id
              AND pc.funding_type = 'Foreign'
              AND pc.funding_category = 'Loan'
        ), 0)::numeric AS amount_usd,
        ARRAY(
            SELECT DISTINCT li.lender_id
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
        )::uuid[] AS lender_ids,
        ARRAY(
            SELECT DISTINCT l.type
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
        )::text[] AS lender_types,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
        )::uuid[] AS institution_ids,
        ARRAY(
            SELECT DISTINCT ir.root_institution_id
            FROM bb_project_institution bpi
            LEFT JOIN institution_rollup ir ON ir.source_institution_id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
              AND ir.root_institution_id IS NOT NULL
        )::uuid[] AS root_institution_ids,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
        )::uuid[] AS region_ids
    FROM latest_bb_project_rows bp
),
gb_rows AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.program_title_id,
        gp.created_at,
        COALESCE((
            SELECT SUM(gfs.loan_usd)
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        ), 0)::numeric AS amount_usd,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        )::uuid[] AS lender_ids,
        ARRAY(
            SELECT DISTINCT l.type
            FROM gb_funding_source gfs
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gfs.gb_project_id = gp.id
        )::text[] AS lender_types,
        ARRAY(
            SELECT DISTINCT gpi.institution_id
            FROM gb_project_institution gpi
            WHERE gpi.gb_project_id = gp.id
              AND gpi.role = 'Executing Agency'
        )::uuid[] AS institution_ids,
        ARRAY(
            SELECT DISTINCT ir.root_institution_id
            FROM gb_project_institution gpi
            LEFT JOIN institution_rollup ir ON ir.source_institution_id = gpi.institution_id
            WHERE gpi.gb_project_id = gp.id
              AND gpi.role = 'Executing Agency'
              AND ir.root_institution_id IS NOT NULL
        )::uuid[] AS root_institution_ids,
        ARRAY(
            SELECT DISTINCT gpl.region_id
            FROM gb_project_location gpl
            WHERE gpl.gb_project_id = gp.id
        )::uuid[] AS region_ids
    FROM latest_gb_project_rows gp
),
dk_rows AS (
    SELECT
        dp.id,
        dp.program_title_id,
        dp.institution_id,
        ir.root_institution_id,
        dp.created_at,
        COALESCE((
            SELECT SUM(dfd.amount_usd)
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
        ), 0)::numeric AS amount_usd,
        ARRAY(
            SELECT DISTINCT dfd.lender_id
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
              AND dfd.lender_id IS NOT NULL
        )::uuid[] AS lender_ids,
        ARRAY(
            SELECT DISTINCT l.type
            FROM dk_financing_detail dfd
            JOIN lender l ON l.id = dfd.lender_id
            WHERE dfd.dk_project_id = dp.id
        )::text[] AS lender_types,
        ARRAY(
            SELECT DISTINCT dpl.region_id
            FROM dk_project_location dpl
            WHERE dpl.dk_project_id = dp.id
        )::uuid[] AS region_ids
    FROM dk_project dp
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
),
la_rows AS (
    SELECT
        la.id,
        la.lender_id,
        l.type AS lender_type,
        la.amount_usd,
        la.effective_date,
        la.created_at,
        dp.id AS dk_project_id,
        dp.program_title_id,
        dp.institution_id,
        ir.root_institution_id
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
)
SELECT
    bucket.stage,
    bucket.label,
    COUNT(DISTINCT bucket.project_key)::bigint AS project_count,
    COALESCE(SUM(bucket.amount_usd), 0)::numeric AS total_loan_usd,
    MIN(bucket.oldest_date)::date AS oldest_date,
    bucket.severity
FROM (
    SELECT
        'BB'::text AS stage,
        'Blue Book belum berlanjut ke Green Book'::text AS label,
        bb.project_identity_id AS project_key,
        bb.amount_usd,
        bb.created_at::date AS oldest_date,
        'warning'::text AS severity
    FROM bb_rows bb
    WHERE NOT EXISTS (
        SELECT 1
        FROM gb_project_bb_project gbp
        WHERE gbp.bb_project_id = bb.id
    )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'BB' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Pipeline' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR bb.lender_ids && sqlc.arg('lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR bb.lender_types && sqlc.arg('lender_types')::text[])
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR bb.institution_ids && sqlc.arg('institution_ids')::uuid[]
          OR bb.root_institution_ids && sqlc.arg('institution_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR bb.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR bb.region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR bb.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR bb.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)

    UNION ALL

    SELECT
        'GB'::text AS stage,
        'Green Book belum berlanjut ke Daftar Kegiatan'::text AS label,
        gb.gb_project_identity_id AS project_key,
        gb.amount_usd,
        gb.created_at::date AS oldest_date,
        'warning'::text AS severity
    FROM gb_rows gb
    WHERE NOT EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        WHERE dpg.gb_project_id = gb.id
    )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'GB' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Pipeline' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR gb.lender_ids && sqlc.arg('lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR gb.lender_types && sqlc.arg('lender_types')::text[])
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR gb.institution_ids && sqlc.arg('institution_ids')::uuid[]
          OR gb.root_institution_ids && sqlc.arg('institution_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR gb.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR gb.region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR gb.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR gb.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)

    UNION ALL

    SELECT
        'DK'::text AS stage,
        'Daftar Kegiatan belum memiliki Loan Agreement'::text AS label,
        dk.id AS project_key,
        dk.amount_usd,
        dk.created_at::date AS oldest_date,
        'warning'::text AS severity
    FROM dk_rows dk
    WHERE NOT EXISTS (
        SELECT 1
        FROM loan_agreement la
        WHERE la.dk_project_id = dk.id
    )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'DK' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Pipeline' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR dk.lender_ids && sqlc.arg('lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR dk.lender_types && sqlc.arg('lender_types')::text[])
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR dk.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR dk.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR dk.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR dk.region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR dk.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR dk.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)

    UNION ALL

    SELECT
        'LA'::text AS stage,
        'Loan Agreement efektif belum dimonitor'::text AS label,
        la.id AS project_key,
        la.amount_usd,
        la.effective_date AS oldest_date,
        'danger'::text AS severity
    FROM la_rows la
    WHERE la.effective_date <= CURRENT_DATE
      AND NOT EXISTS (
          SELECT 1
          FROM monitoring_disbursement md
          WHERE md.loan_agreement_id = la.id
      )
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR 'LA' = ANY(sqlc.arg('pipeline_statuses')::text[]) OR 'Monitoring' = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR 'Ongoing' = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR la.lender_type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR la.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR la.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR la.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR la.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR la.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
) bucket
GROUP BY bucket.stage, bucket.label, bucket.severity
ORDER BY
    CASE bucket.stage
        WHEN 'BB' THEN 1
        WHEN 'GB' THEN 2
        WHEN 'DK' THEN 3
        WHEN 'LA' THEN 4
        ELSE 5
    END;

-- name: ListDashboardAnalyticsDataQualityIssues :many
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
bb_quality_rows AS (
    SELECT
        'Blue Book'::text AS stage,
        bp.id AS entity_id,
        bp.program_title_id,
        COALESCE((
            SELECT SUM(pc.amount_usd)
            FROM bb_project_cost pc
            WHERE pc.bb_project_id = bp.id
              AND pc.funding_type = 'Foreign'
        ), 0)::numeric AS funding_amount_usd,
        EXISTS (
            SELECT 1
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
        ) AS has_executing_agency,
        EXISTS (
            SELECT 1
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
        ) AS has_lender,
        EXISTS (
            SELECT 1
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
        ) AS has_region,
        ARRAY(
            SELECT DISTINCT li.lender_id
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
        )::uuid[] AS lender_ids,
        ARRAY(
            SELECT DISTINCT l.type
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
        )::text[] AS lender_types,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
        )::uuid[] AS institution_ids,
        ARRAY(
            SELECT DISTINCT ir.root_institution_id
            FROM bb_project_institution bpi
            LEFT JOIN institution_rollup ir ON ir.source_institution_id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
              AND ir.root_institution_id IS NOT NULL
        )::uuid[] AS root_institution_ids,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
        )::uuid[] AS region_ids
    FROM latest_bb_project_rows bp
),
gb_quality_rows AS (
    SELECT
        'Green Book Funding Source'::text AS stage,
        gp.id AS entity_id,
        gp.program_title_id,
        COALESCE((
            SELECT SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd)
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        ), 0)::numeric AS funding_amount_usd,
        EXISTS (
            SELECT 1
            FROM gb_project_institution gpi
            WHERE gpi.gb_project_id = gp.id
              AND gpi.role = 'Executing Agency'
        ) AS has_executing_agency,
        EXISTS (
            SELECT 1
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        ) AS has_lender,
        EXISTS (
            SELECT 1
            FROM gb_project_location gpl
            WHERE gpl.gb_project_id = gp.id
        ) AS has_region,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        )::uuid[] AS lender_ids,
        ARRAY(
            SELECT DISTINCT l.type
            FROM gb_funding_source gfs
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gfs.gb_project_id = gp.id
        )::text[] AS lender_types,
        ARRAY(
            SELECT DISTINCT gpi.institution_id
            FROM gb_project_institution gpi
            WHERE gpi.gb_project_id = gp.id
              AND gpi.role = 'Executing Agency'
        )::uuid[] AS institution_ids,
        ARRAY(
            SELECT DISTINCT ir.root_institution_id
            FROM gb_project_institution gpi
            LEFT JOIN institution_rollup ir ON ir.source_institution_id = gpi.institution_id
            WHERE gpi.gb_project_id = gp.id
              AND gpi.role = 'Executing Agency'
              AND ir.root_institution_id IS NOT NULL
        )::uuid[] AS root_institution_ids,
        ARRAY(
            SELECT DISTINCT gpl.region_id
            FROM gb_project_location gpl
            WHERE gpl.gb_project_id = gp.id
        )::uuid[] AS region_ids
    FROM latest_gb_project_rows gp
),
dk_quality_rows AS (
    SELECT
        'Daftar Kegiatan Financing'::text AS stage,
        dp.id AS entity_id,
        dp.program_title_id,
        COALESCE((
            SELECT SUM(dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd)
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
        ), 0)::numeric AS funding_amount_usd,
        (dp.institution_id IS NOT NULL) AS has_executing_agency,
        EXISTS (
            SELECT 1
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
              AND dfd.lender_id IS NOT NULL
        ) AS has_lender,
        EXISTS (
            SELECT 1
            FROM dk_project_location dpl
            WHERE dpl.dk_project_id = dp.id
        ) AS has_region,
        ARRAY(
            SELECT DISTINCT dfd.lender_id
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
              AND dfd.lender_id IS NOT NULL
        )::uuid[] AS lender_ids,
        ARRAY(
            SELECT DISTINCT l.type
            FROM dk_financing_detail dfd
            JOIN lender l ON l.id = dfd.lender_id
            WHERE dfd.dk_project_id = dp.id
        )::text[] AS lender_types,
        ARRAY[dp.institution_id]::uuid[] AS institution_ids,
        ARRAY[ir.root_institution_id]::uuid[] AS root_institution_ids,
        ARRAY(
            SELECT DISTINCT dpl.region_id
            FROM dk_project_location dpl
            WHERE dpl.dk_project_id = dp.id
        )::uuid[] AS region_ids
    FROM dk_project dp
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
),
project_quality_rows AS (
    SELECT * FROM bb_quality_rows
    UNION ALL
    SELECT * FROM gb_quality_rows
    UNION ALL
    SELECT * FROM dk_quality_rows
),
filtered_project_quality_rows AS (
    SELECT *
    FROM project_quality_rows p
    WHERE (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR p.lender_ids && sqlc.arg('lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR p.lender_types && sqlc.arg('lender_types')::text[])
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR p.institution_ids && sqlc.arg('institution_ids')::uuid[]
          OR p.root_institution_ids && sqlc.arg('institution_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR p.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR p.region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR p.funding_amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR p.funding_amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
),
effective_without_monitoring AS (
    SELECT COUNT(*)::bigint AS affected_count
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    WHERE la.effective_date <= CURRENT_DATE
      AND NOT EXISTS (
          SELECT 1
          FROM monitoring_disbursement md
          WHERE md.loan_agreement_id = la.id
            AND (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
            AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
      )
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR dp.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR dp.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR la.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR la.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
),
planned_zero_realized_positive AS (
    SELECT COUNT(*)::bigint AS affected_count
    FROM monitoring_disbursement md
    JOIN loan_agreement la ON la.id = md.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN institution_rollup ir ON ir.source_institution_id = dp.institution_id
    WHERE md.planned_usd = 0
      AND md.realized_usd > 0
      AND (sqlc.narg('budget_year')::int IS NULL OR md.budget_year = sqlc.narg('budget_year')::int)
      AND (sqlc.narg('quarter')::varchar IS NULL OR md.quarter = sqlc.narg('quarter')::varchar)
      AND (COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0 OR la.lender_id = ANY(sqlc.arg('lender_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('lender_types')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('lender_types')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('institution_ids')::uuid[]), 0) = 0
          OR dp.institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
          OR ir.root_institution_id = ANY(sqlc.arg('institution_ids')::uuid[])
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR dp.program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR la.amount_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR la.amount_usd <= sqlc.narg('foreign_loan_max')::numeric)
)
SELECT
    issue.code,
    issue.label,
    issue.stage,
    issue.affected_count,
    issue.severity,
    issue.target
FROM (
    SELECT
        'NO_EXECUTING_AGENCY'::text AS code,
        'Project tanpa Executing Agency'::text AS label,
        stage,
        COUNT(*)::bigint AS affected_count,
        'warning'::text AS severity,
        'projects'::text AS target
    FROM filtered_project_quality_rows
    WHERE NOT has_executing_agency
    GROUP BY stage

    UNION ALL

    SELECT
        'NO_LENDER'::text AS code,
        'Project tanpa lender sesuai stage'::text AS label,
        stage,
        COUNT(*)::bigint AS affected_count,
        'warning'::text AS severity,
        'projects'::text AS target
    FROM filtered_project_quality_rows
    WHERE NOT has_lender
    GROUP BY stage

    UNION ALL

    SELECT
        'NO_REGION'::text AS code,
        'Project tanpa location'::text AS label,
        stage,
        COUNT(*)::bigint AS affected_count,
        'info'::text AS severity,
        'projects'::text AS target
    FROM filtered_project_quality_rows
    WHERE NOT has_region
    GROUP BY stage

    UNION ALL

    SELECT
        'NO_FUNDING_AMOUNT'::text AS code,
        'Funding amount USD kosong atau nol'::text AS label,
        stage,
        COUNT(*)::bigint AS affected_count,
        'warning'::text AS severity,
        'projects'::text AS target
    FROM filtered_project_quality_rows
    WHERE funding_amount_usd <= 0
    GROUP BY stage

    UNION ALL

    SELECT
        'EFFECTIVE_NO_MONITORING'::text AS code,
        'Loan Agreement efektif belum dimonitor'::text AS label,
        'Monitoring'::text AS stage,
        affected_count,
        'danger'::text AS severity,
        'monitoring'::text AS target
    FROM effective_without_monitoring
    WHERE affected_count > 0

    UNION ALL

    SELECT
        'PLANNED_ZERO_REALIZED_POSITIVE'::text AS code,
        'Realisasi ada tetapi planned 0'::text AS label,
        'Monitoring'::text AS stage,
        affected_count,
        'warning'::text AS severity,
        'monitoring'::text AS target
    FROM planned_zero_realized_positive
    WHERE affected_count > 0
) issue
WHERE issue.affected_count > 0
ORDER BY
    CASE issue.severity
        WHEN 'danger' THEN 1
        WHEN 'warning' THEN 2
        ELSE 3
    END,
    issue.code ASC,
    issue.stage ASC;
