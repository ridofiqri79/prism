-- ===== SPATIAL DISTRIBUTION =====

-- name: ListSpatialRegionMetrics :many
WITH project_rows AS (
    SELECT
        bp.id,
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
        )::text[] AS loan_types
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
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (
          sqlc.narg('search')::text IS NULL
          OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
),
display_regions AS (
    SELECT r.id, r.code, r.name, r.type, r.parent_code
    FROM region r
    WHERE (
        sqlc.arg('level')::text = 'city'
        AND r.type = 'CITY'
        AND r.parent_code = sqlc.narg('province_code')::text
    )
    OR (
        sqlc.arg('level')::text <> 'city'
        AND r.type = 'PROVINCE'
    )
),
region_footprints AS (
    SELECT DISTINCT fp.id AS project_id, matched.code AS region_code
    FROM filtered_projects fp
    JOIN bb_project_location bpl ON bpl.bb_project_id = fp.id
    JOIN region loc ON loc.id = bpl.region_id
    JOIN region matched ON (
        sqlc.arg('level')::text = 'province'
        AND (
            (loc.type = 'PROVINCE' AND matched.code = loc.code)
            OR (loc.type = 'CITY' AND matched.code = loc.parent_code)
            OR (loc.type = 'COUNTRY' AND matched.type = 'PROVINCE' AND matched.parent_code = loc.code)
        )
    )
    UNION
    SELECT DISTINCT fp.id AS project_id, city.code AS region_code
    FROM filtered_projects fp
    JOIN bb_project_location bpl ON bpl.bb_project_id = fp.id
    JOIN region loc ON loc.id = bpl.region_id
    JOIN region city ON (
        sqlc.arg('level')::text = 'city'
        AND city.type = 'CITY'
        AND city.parent_code = sqlc.narg('province_code')::text
        AND loc.type = 'CITY'
        AND city.code = loc.code
    )
)
SELECT
    dr.id,
    dr.code,
    dr.name,
    dr.type,
    dr.parent_code,
    COUNT(DISTINCT rf.project_id)::bigint AS project_count,
    COALESCE(SUM(fp.foreign_loan_usd), 0)::numeric AS total_loan_usd
FROM display_regions dr
LEFT JOIN region_footprints rf ON rf.region_code = dr.code
LEFT JOIN filtered_projects fp ON fp.id = rf.project_id
GROUP BY dr.id, dr.code, dr.name, dr.type, dr.parent_code
ORDER BY dr.name ASC;

-- name: GetSpatialRegionSummary :one
WITH project_rows AS (
    SELECT
        bp.id,
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
        )::text[] AS loan_types
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
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (
          sqlc.narg('search')::text IS NULL
          OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
),
region_footprints AS (
    SELECT DISTINCT fp.id AS project_id
    FROM filtered_projects fp
    JOIN bb_project_location bpl ON bpl.bb_project_id = fp.id
    JOIN region loc ON loc.id = bpl.region_id
    JOIN region matched ON (
        sqlc.arg('level')::text = 'province'
        AND (
            (loc.type = 'PROVINCE' AND matched.code = loc.code)
            OR (loc.type = 'CITY' AND matched.code = loc.parent_code)
            OR (loc.type = 'COUNTRY' AND matched.type = 'PROVINCE' AND matched.parent_code = loc.code)
        )
    )
    UNION
    SELECT DISTINCT fp.id AS project_id
    FROM filtered_projects fp
    JOIN bb_project_location bpl ON bpl.bb_project_id = fp.id
    JOIN region loc ON loc.id = bpl.region_id
    JOIN region city ON (
        sqlc.arg('level')::text = 'city'
        AND city.type = 'CITY'
        AND city.parent_code = sqlc.narg('province_code')::text
        AND loc.type = 'CITY'
        AND city.code = loc.code
    )
)
SELECT
    COUNT(DISTINCT fp.id)::bigint AS project_count,
    COALESCE(SUM(fp.foreign_loan_usd), 0)::numeric AS total_loan_usd
FROM filtered_projects fp
JOIN region_footprints rf ON rf.project_id = fp.id;

-- name: GetSpatialRegionByCode :one
SELECT id, code, name, type, parent_code
FROM region
WHERE code = $1;

-- name: ListSpatialRegionFilterIDs :many
WITH selected AS (
    SELECT id, code, name, type, parent_code
    FROM region
    WHERE code = sqlc.arg('region_code')::text
),
ids AS (
    SELECT r.id
    FROM region r
    WHERE r.type = 'COUNTRY'
      AND sqlc.arg('level')::text = 'province'

    UNION
    SELECT selected.id
    FROM selected

    UNION
    SELECT child.id
    FROM selected
    JOIN region child ON sqlc.arg('level')::text = 'province'
        AND selected.type = 'PROVINCE'
        AND child.type = 'CITY'
        AND child.parent_code = selected.code
)
SELECT id
FROM ids;
