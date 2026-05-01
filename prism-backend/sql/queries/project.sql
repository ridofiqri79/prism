-- ===== PROJECT MASTER =====

-- name: ListProjectMaster :many
WITH project_rows AS (
    SELECT
        bp.id,
        bp.blue_book_id,
        bp.project_identity_id,
        bp.bb_code,
        bp.project_name,
        bp.program_title_id,
        COALESCE(pt.title, '')::text AS program_title,
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
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS indication_lenders,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(i.short_name, i.name)
            FROM bb_project_institution bpi
            JOIN institution i ON i.id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY COALESCE(i.short_name, i.name)
        )::text[] AS executing_agencies,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS fixed_lenders,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids,
        ARRAY(
            SELECT DISTINCT r.name
            FROM bb_project_location bpl
            JOIN region r ON r.id = bpl.region_id
            WHERE bpl.bb_project_id = bp.id
            ORDER BY r.name
        )::text[] AS locations,
        ARRAY(
            SELECT DISTINCT dk.date::text
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY dk.date::text
        )::text[] AS dk_dates
        ,
        (bp.id = (
            SELECT latest.id
            FROM bb_project latest
            JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
            WHERE latest.project_identity_id = bp.project_identity_id
              AND latest.status = 'active'
            ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
            LIMIT 1
        ))::boolean AS is_latest,
        EXISTS (
            SELECT 1
            FROM bb_project newer
            JOIN blue_book newer_bb ON newer_bb.id = newer.blue_book_id
            JOIN blue_book current_bb ON current_bb.id = bp.blue_book_id
            WHERE newer.project_identity_id = bp.project_identity_id
              AND newer.status = 'active'
              AND (
                  newer_bb.revision_number > current_bb.revision_number
                  OR (
                      newer_bb.revision_number = current_bb.revision_number
                      AND newer_bb.created_at > current_bb.created_at
                  )
              )
        )::boolean AS has_newer_revision,
        CONCAT(
            'BB ',
            p.name,
            CASE
                WHEN bb.revision_number > 0 THEN CONCAT(' Revisi ke-', bb.revision_number)
                ELSE ''
            END
        )::text AS blue_book_revision_label
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    JOIN period p ON p.id = bb.period_id
    LEFT JOIN program_title pt ON pt.id = bp.program_title_id
    WHERE bp.status = 'active'
      AND (
          sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id
              FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id
                AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1
          )
      )
),
filtered_projects AS (
    SELECT *
    FROM project_rows
    WHERE (COALESCE(cardinality(sqlc.arg('loan_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('loan_types')::text[])
      AND (COALESCE(cardinality(sqlc.arg('indication_lender_ids')::uuid[]), 0) = 0 OR indication_lender_ids && sqlc.arg('indication_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0 OR executing_agency_ids && sqlc.arg('executing_agency_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('fixed_lender_ids')::uuid[]), 0) = 0 OR fixed_lender_ids && sqlc.arg('fixed_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          (sqlc.narg('dk_date_from')::date IS NULL AND sqlc.narg('dk_date_to')::date IS NULL)
          OR EXISTS (
              SELECT 1
              FROM unnest(dk_dates) AS item(dk_date)
              WHERE (sqlc.narg('dk_date_from')::date IS NULL OR item.dk_date::date >= sqlc.narg('dk_date_from')::date)
                AND (sqlc.narg('dk_date_to')::date IS NULL OR item.dk_date::date <= sqlc.narg('dk_date_to')::date)
          )
      )
      AND (
          sqlc.narg('search')::text IS NULL
          OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR EXISTS (
              SELECT 1
              FROM unnest(indication_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(fixed_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(executing_agencies) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
      )
)
SELECT
    id,
    blue_book_id,
    project_identity_id,
    bb_code,
    project_name,
    loan_types,
    indication_lenders,
    executing_agencies,
    fixed_lenders,
    project_status,
    pipeline_status,
    program_title,
    locations,
    foreign_loan_usd,
    dk_dates,
    is_latest,
    has_newer_revision,
    blue_book_revision_label
FROM filtered_projects
ORDER BY
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'asc' THEN LOWER(project_name) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'desc' THEN LOWER(project_name) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'bb_code' AND sqlc.arg('order')::text = 'asc' THEN LOWER(bb_code) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'bb_code' AND sqlc.arg('order')::text = 'desc' THEN LOWER(bb_code) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'loan_types' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(loan_types, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'loan_types' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(loan_types, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'indication_lenders' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(indication_lenders, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'indication_lenders' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(indication_lenders, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'executing_agencies' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(executing_agencies, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'executing_agencies' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(executing_agencies, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'fixed_lenders' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(fixed_lenders, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'fixed_lenders' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(fixed_lenders, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'project_status' AND sqlc.arg('order')::text = 'asc' THEN LOWER(project_status || ' - ' || pipeline_status) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'project_status' AND sqlc.arg('order')::text = 'desc' THEN LOWER(project_status || ' - ' || pipeline_status) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'pipeline_status' AND sqlc.arg('order')::text = 'asc' THEN LOWER(pipeline_status) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'pipeline_status' AND sqlc.arg('order')::text = 'desc' THEN LOWER(pipeline_status) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'program_title' AND sqlc.arg('order')::text = 'asc' THEN LOWER(program_title) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'program_title' AND sqlc.arg('order')::text = 'desc' THEN LOWER(program_title) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'locations' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(locations, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'locations' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(locations, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'foreign_loan_usd' AND sqlc.arg('order')::text = 'asc' THEN foreign_loan_usd END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'foreign_loan_usd' AND sqlc.arg('order')::text = 'desc' THEN foreign_loan_usd END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'dk_dates' AND sqlc.arg('order')::text = 'asc' THEN array_to_string(dk_dates, ', ') END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'dk_dates' AND sqlc.arg('order')::text = 'desc' THEN array_to_string(dk_dates, ', ') END DESC,
    project_name ASC,
    bb_code ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountProjectMaster :one
WITH project_rows AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.program_title_id,
        bp.bb_code,
        bp.project_name,
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
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS indication_lenders,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(i.short_name, i.name)
            FROM bb_project_institution bpi
            JOIN institution i ON i.id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY COALESCE(i.short_name, i.name)
        )::text[] AS executing_agencies,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS fixed_lenders,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids,
        ARRAY(
            SELECT DISTINCT dk.date::text
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY dk.date::text
        )::text[] AS dk_dates
    FROM bb_project bp
    WHERE bp.status = 'active'
      AND (
          sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id
              FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id
                AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1
          )
      )
)
SELECT COUNT(*)::bigint
FROM project_rows
WHERE (COALESCE(cardinality(sqlc.arg('loan_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('loan_types')::text[])
  AND (COALESCE(cardinality(sqlc.arg('indication_lender_ids')::uuid[]), 0) = 0 OR indication_lender_ids && sqlc.arg('indication_lender_ids')::uuid[])
  AND (COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0 OR executing_agency_ids && sqlc.arg('executing_agency_ids')::uuid[])
  AND (COALESCE(cardinality(sqlc.arg('fixed_lender_ids')::uuid[]), 0) = 0 OR fixed_lender_ids && sqlc.arg('fixed_lender_ids')::uuid[])
  AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
  AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
  AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
  AND (
      (sqlc.narg('dk_date_from')::date IS NULL AND sqlc.narg('dk_date_to')::date IS NULL)
      OR EXISTS (
          SELECT 1
          FROM unnest(dk_dates) AS item(dk_date)
          WHERE (sqlc.narg('dk_date_from')::date IS NULL OR item.dk_date::date >= sqlc.narg('dk_date_from')::date)
            AND (sqlc.narg('dk_date_to')::date IS NULL OR item.dk_date::date <= sqlc.narg('dk_date_to')::date)
      )
  )
  AND (
      sqlc.narg('search')::text IS NULL
      OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      OR EXISTS (
          SELECT 1
          FROM unnest(indication_lenders) AS item(label)
          WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
      OR EXISTS (
          SELECT 1
          FROM unnest(fixed_lenders) AS item(label)
          WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
      OR EXISTS (
          SELECT 1
          FROM unnest(executing_agencies) AS item(label)
          WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
  );

-- name: GetProjectMasterFundingSummary :one
WITH project_rows AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.program_title_id,
        bp.bb_code,
        bp.project_name,
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
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS indication_lenders,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(i.short_name, i.name)
            FROM bb_project_institution bpi
            JOIN institution i ON i.id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY COALESCE(i.short_name, i.name)
        )::text[] AS executing_agencies,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS fixed_lenders,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids,
        ARRAY(
            SELECT DISTINCT dk.date::text
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY dk.date::text
        )::text[] AS dk_dates
    FROM bb_project bp
    WHERE bp.status = 'active'
      AND (
          sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id
              FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id
                AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1
          )
      )
),
filtered_projects AS (
    SELECT *
    FROM project_rows
    WHERE (COALESCE(cardinality(sqlc.arg('loan_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('loan_types')::text[])
      AND (COALESCE(cardinality(sqlc.arg('indication_lender_ids')::uuid[]), 0) = 0 OR indication_lender_ids && sqlc.arg('indication_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0 OR executing_agency_ids && sqlc.arg('executing_agency_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('fixed_lender_ids')::uuid[]), 0) = 0 OR fixed_lender_ids && sqlc.arg('fixed_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          (sqlc.narg('dk_date_from')::date IS NULL AND sqlc.narg('dk_date_to')::date IS NULL)
          OR EXISTS (
              SELECT 1
              FROM unnest(dk_dates) AS item(dk_date)
              WHERE (sqlc.narg('dk_date_from')::date IS NULL OR item.dk_date::date >= sqlc.narg('dk_date_from')::date)
                AND (sqlc.narg('dk_date_to')::date IS NULL OR item.dk_date::date <= sqlc.narg('dk_date_to')::date)
          )
      )
      AND (
          sqlc.narg('search')::text IS NULL
          OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR EXISTS (
              SELECT 1
              FROM unnest(indication_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(fixed_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(executing_agencies) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
      )
)
SELECT
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS total_loan_usd,
    COALESCE(SUM(foreign_grant_usd), 0)::numeric AS total_grant_usd,
    COALESCE(SUM(counterpart_usd), 0)::numeric AS total_counterpart_usd
FROM filtered_projects;
