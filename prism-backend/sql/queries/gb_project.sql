-- ===== GREEN BOOK =====

-- name: ListGreenBooks :many
SELECT
    gb.id,
    gb.publish_year,
    gb.replaces_green_book_id,
    gb.revision_number,
    gb.status,
    gb.created_at,
    gb.updated_at,
    (
        SELECT COUNT(*)
        FROM gb_project gp
        WHERE gp.green_book_id = gb.id
    )::BIGINT AS project_count
FROM green_book gb
WHERE (
    sqlc.narg('search')::text IS NULL
    OR gb.publish_year::text ILIKE '%' || sqlc.narg('search')::text || '%'
    OR gb.status ILIKE '%' || sqlc.narg('search')::text || '%'
    OR CASE gb.status WHEN 'active' THEN 'Berlaku' WHEN 'superseded' THEN 'Tidak Berlaku' ELSE gb.status END ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (
    COALESCE(cardinality(sqlc.arg('publish_years')::int[]), 0) = 0
    OR gb.publish_year = ANY(sqlc.arg('publish_years')::int[])
)
AND (
    COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) = 0
    OR gb.status = ANY(sqlc.arg('statuses')::text[])
)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'publish_year' AND sqlc.arg('sort_order')::text = 'asc' THEN gb.publish_year END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'publish_year' AND sqlc.arg('sort_order')::text = 'desc' THEN gb.publish_year END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'revision' AND sqlc.arg('sort_order')::text = 'asc' THEN gb.revision_number END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'revision' AND sqlc.arg('sort_order')::text = 'desc' THEN gb.revision_number END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'asc' THEN gb.status END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'desc' THEN gb.status END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_count' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT COUNT(*)
        FROM gb_project gp
        WHERE gp.green_book_id = gb.id
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_count' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT COUNT(*)
        FROM gb_project gp
        WHERE gp.green_book_id = gb.id
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN gb.created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN gb.created_at END DESC,
    gb.publish_year DESC,
    gb.revision_number DESC,
    gb.id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountGreenBooks :one
SELECT COUNT(*)
FROM green_book
WHERE (
    sqlc.narg('search')::text IS NULL
    OR publish_year::text ILIKE '%' || sqlc.narg('search')::text || '%'
    OR status ILIKE '%' || sqlc.narg('search')::text || '%'
    OR CASE status WHEN 'active' THEN 'Berlaku' WHEN 'superseded' THEN 'Tidak Berlaku' ELSE status END ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (
    COALESCE(cardinality(sqlc.arg('publish_years')::int[]), 0) = 0
    OR publish_year = ANY(sqlc.arg('publish_years')::int[])
)
AND (
    COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) = 0
    OR status = ANY(sqlc.arg('statuses')::text[])
);

-- name: GetGreenBook :one
SELECT
    gb.id,
    gb.publish_year,
    gb.replaces_green_book_id,
    gb.revision_number,
    gb.status,
    gb.created_at,
    gb.updated_at,
    (
        SELECT COUNT(*)
        FROM gb_project gp
        WHERE gp.green_book_id = gb.id
    )::BIGINT AS project_count
FROM green_book gb
WHERE gb.id = $1;

-- name: CreateGreenBook :one
INSERT INTO green_book (publish_year, replaces_green_book_id, revision_number, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateGreenBook :one
UPDATE green_book
SET publish_year = $2,
    revision_number = $3,
    status = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SupersedeGreenBooksByPublishYear :exec
UPDATE green_book
SET status = 'superseded',
    updated_at = NOW()
WHERE publish_year = $1
  AND status = 'active';

-- name: GetActiveGreenBookByPublishYear :one
SELECT *
FROM green_book
WHERE publish_year = $1
  AND status = 'active'
ORDER BY revision_number DESC, created_at DESC
LIMIT 1;

-- name: CountGreenBooksByPublishYearAndRevisionNumber :one
SELECT COUNT(*)
FROM green_book
WHERE publish_year = sqlc.arg('publish_year')
  AND revision_number = sqlc.arg('revision_number');

-- name: CountGreenBooksByPublishYearAndRevisionNumberExcept :one
SELECT COUNT(*)
FROM green_book
WHERE publish_year = sqlc.arg('publish_year')
  AND revision_number = sqlc.arg('revision_number')
  AND id <> sqlc.arg('id');

-- name: CountAnyGBProjectsByGreenBook :one
SELECT COUNT(*)
FROM gb_project
WHERE green_book_id = $1;

-- name: CountGreenBookRevisionsReplacing :one
SELECT COUNT(*)
FROM green_book
WHERE replaces_green_book_id = $1;

-- name: HardDeleteGreenBook :one
DELETE FROM green_book
WHERE green_book.id = $1
  AND NOT EXISTS (
      SELECT 1
      FROM gb_project gp
      WHERE gp.green_book_id = green_book.id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM green_book child
      WHERE child.replaces_green_book_id = green_book.id
  )
RETURNING *;

-- ===== GB PROJECT =====

-- name: CreateGBProjectIdentity :one
INSERT INTO gb_project_identity DEFAULT VALUES
RETURNING *;

-- name: GetGBProjectIdentity :one
SELECT *
FROM gb_project_identity
WHERE id = $1;

-- name: ListGBProjectsByGreenBook :many
SELECT *
FROM gb_project
WHERE green_book_id = sqlc.arg('green_book_id')
  AND (
      (
          COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) = 0
          AND status = 'active'
      )
      OR (
          COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) > 0
          AND status = ANY(sqlc.arg('statuses')::text[])
      )
  )
  AND (
      sqlc.narg('search')::text IS NULL
      OR gb_code ILIKE '%' || sqlc.narg('search')::text || '%'
      OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
      OR EXISTS (
          SELECT 1
          FROM gb_project_institution gpi
          JOIN institution i ON i.id = gpi.institution_id
          WHERE gpi.gb_project_id = gb_project.id
            AND gpi.role = 'Executing Agency'
            AND (
                i.name ILIKE '%' || sqlc.narg('search')::text || '%'
                OR COALESCE(i.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
            )
      )
      OR EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN bb_project bp ON bp.id = gbp.bb_project_id
          WHERE gbp.gb_project_id = gb_project.id
            AND (
                bp.bb_code ILIKE '%' || sqlc.narg('search')::text || '%'
                OR bp.project_name ILIKE '%' || sqlc.narg('search')::text || '%'
            )
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('bb_project_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          WHERE gbp.gb_project_id = gb_project.id
            AND gbp.bb_project_id = ANY(sqlc.arg('bb_project_ids')::uuid[])
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM gb_project_institution gpi
          WHERE gpi.gb_project_id = gb_project.id
            AND gpi.role = 'Executing Agency'
            AND gpi.institution_id = ANY(sqlc.arg('executing_agency_ids')::uuid[])
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('location_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM gb_project_location gpl
          WHERE gpl.gb_project_id = gb_project.id
            AND gpl.region_id = ANY(sqlc.arg('location_ids')::uuid[])
      )
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'gb_code' AND sqlc.arg('sort_order')::text = 'asc' THEN gb_code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'gb_code' AND sqlc.arg('sort_order')::text = 'desc' THEN gb_code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_name' AND sqlc.arg('sort_order')::text = 'asc' THEN project_name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_name' AND sqlc.arg('sort_order')::text = 'desc' THEN project_name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'bb_projects' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT string_agg(bp.bb_code, ', ' ORDER BY bp.bb_code)
        FROM gb_project_bb_project gbp
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        WHERE gbp.gb_project_id = gb_project.id
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'bb_projects' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT string_agg(bp.bb_code, ', ' ORDER BY bp.bb_code)
        FROM gb_project_bb_project gbp
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        WHERE gbp.gb_project_id = gb_project.id
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'asc' THEN status END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'desc' THEN status END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN created_at END DESC,
    gb_code ASC,
    id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountGBProjectsByGreenBook :one
SELECT COUNT(*)
FROM gb_project
WHERE green_book_id = sqlc.arg('green_book_id')
  AND (
      (
          COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) = 0
          AND status = 'active'
      )
      OR (
          COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) > 0
          AND status = ANY(sqlc.arg('statuses')::text[])
      )
  )
  AND (
      sqlc.narg('search')::text IS NULL
      OR gb_code ILIKE '%' || sqlc.narg('search')::text || '%'
      OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
      OR EXISTS (
          SELECT 1
          FROM gb_project_institution gpi
          JOIN institution i ON i.id = gpi.institution_id
          WHERE gpi.gb_project_id = gb_project.id
            AND gpi.role = 'Executing Agency'
            AND (
                i.name ILIKE '%' || sqlc.narg('search')::text || '%'
                OR COALESCE(i.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
            )
      )
      OR EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN bb_project bp ON bp.id = gbp.bb_project_id
          WHERE gbp.gb_project_id = gb_project.id
            AND (
                bp.bb_code ILIKE '%' || sqlc.narg('search')::text || '%'
                OR bp.project_name ILIKE '%' || sqlc.narg('search')::text || '%'
            )
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('bb_project_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          WHERE gbp.gb_project_id = gb_project.id
            AND gbp.bb_project_id = ANY(sqlc.arg('bb_project_ids')::uuid[])
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM gb_project_institution gpi
          WHERE gpi.gb_project_id = gb_project.id
            AND gpi.role = 'Executing Agency'
            AND gpi.institution_id = ANY(sqlc.arg('executing_agency_ids')::uuid[])
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('location_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM gb_project_location gpl
          WHERE gpl.gb_project_id = gb_project.id
            AND gpl.region_id = ANY(sqlc.arg('location_ids')::uuid[])
      )
  )
;

-- name: GetGBProject :one
SELECT *
FROM gb_project
WHERE id = $1;

-- name: GetActiveGBProjectByGreenBook :one
SELECT *
FROM gb_project
WHERE green_book_id = $1
  AND id = $2
  AND status = 'active';

-- name: GetGBProjectByCode :one
SELECT *
FROM gb_project
WHERE gb_code = $1;

-- name: GetGBProjectByGreenBookAndCode :one
SELECT *
FROM gb_project
WHERE green_book_id = $1
  AND LOWER(gb_code) = LOWER($2)
LIMIT 1;

-- name: FindPreviousGBProjectByCodeForGreenBook :one
SELECT gp.*
FROM gb_project gp
JOIN green_book source_gb ON source_gb.id = gp.green_book_id
JOIN green_book target_gb ON target_gb.id = $1
WHERE LOWER(gp.gb_code) = LOWER($2)
  AND source_gb.publish_year = target_gb.publish_year
  AND source_gb.id <> target_gb.id
  AND (
      source_gb.revision_number < target_gb.revision_number
      OR (
          source_gb.revision_number = target_gb.revision_number
          AND source_gb.created_at < target_gb.created_at
      )
  )
ORDER BY source_gb.revision_number DESC, source_gb.created_at DESC
LIMIT 1;

-- name: GetGBProjectWithRelations :one
SELECT
    gp.id,
    gp.green_book_id,
    gp.gb_project_identity_id,
    gp.program_title_id,
    gp.gb_code,
    gp.project_name,
    gp.duration,
    gp.objective,
    gp.scope_of_project,
    gp.status,
    gp.created_at,
    gp.updated_at,
    COUNT(DISTINCT gbp.bb_project_id) AS bb_project_count,
    COUNT(DISTINCT gpi.institution_id) AS institution_count,
    COUNT(DISTINCT gpl.region_id) AS location_count,
    COUNT(DISTINCT ga.id) AS activity_count,
    COUNT(DISTINCT gfs.id) AS funding_source_count,
    COUNT(DISTINCT gdp.id) AS disbursement_plan_count,
    COUNT(DISTINCT gfa.id) AS funding_allocation_count
FROM gb_project gp
LEFT JOIN gb_project_bb_project gbp ON gbp.gb_project_id = gp.id
LEFT JOIN gb_project_institution gpi ON gpi.gb_project_id = gp.id
LEFT JOIN gb_project_location gpl ON gpl.gb_project_id = gp.id
LEFT JOIN gb_activity ga ON ga.gb_project_id = gp.id
LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = gp.id
LEFT JOIN gb_disbursement_plan gdp ON gdp.gb_project_id = gp.id
LEFT JOIN gb_funding_allocation gfa ON gfa.gb_activity_id = ga.id
WHERE gp.id = $1
GROUP BY gp.id;

-- name: CreateGBProject :one
INSERT INTO gb_project (
    green_book_id,
    gb_project_identity_id,
    program_title_id,
    gb_code,
    project_name,
    duration,
    objective,
    scope_of_project
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListGBProjectsForClone :many
SELECT *
FROM gb_project
WHERE green_book_id = $1
  AND status = 'active'
ORDER BY gb_code ASC;

-- name: ListGBProjectsForCloneByIDs :many
SELECT *
FROM gb_project
WHERE green_book_id = $1
  AND status = 'active'
  AND id = ANY(sqlc.arg('project_ids')::uuid[])
ORDER BY gb_code ASC;

-- name: GetLatestGBProjectByIdentity :one
SELECT gp.*
FROM gb_project gp
JOIN green_book gb ON gb.id = gp.green_book_id
WHERE gp.gb_project_identity_id = $1
  AND gp.status = 'active'
ORDER BY gb.publish_year DESC, gb.revision_number DESC, gb.created_at DESC
LIMIT 1;

-- name: GetLatestGBProjectByProject :one
SELECT latest.*
FROM gb_project current_project
JOIN LATERAL (
    SELECT gp.*
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.gb_project_identity_id = current_project.gb_project_identity_id
      AND gp.status = 'active'
    ORDER BY gb.publish_year DESC, gb.revision_number DESC, gb.created_at DESC
    LIMIT 1
) latest ON TRUE
WHERE current_project.id = $1;

-- name: ListGBProjectHistoryByIdentity :many
SELECT
    gp.id,
    gp.gb_project_identity_id,
    gp.green_book_id,
    gp.gb_code,
    gp.project_name,
    gb.publish_year,
    gb.revision_number,
    gb.status AS book_status,
    (gp.id = (
        SELECT latest.id
        FROM gb_project latest
        JOIN green_book latest_gb ON latest_gb.id = latest.green_book_id
        WHERE latest.gb_project_identity_id = gp.gb_project_identity_id
          AND latest.status = 'active'
        ORDER BY latest_gb.publish_year DESC, latest_gb.revision_number DESC, latest_gb.created_at DESC
        LIMIT 1
    ))::boolean AS is_latest,
    EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        WHERE dpg.gb_project_id = gp.id
    )::boolean AS used_by_downstream
FROM gb_project gp
JOIN green_book gb ON gb.id = gp.green_book_id
WHERE gp.gb_project_identity_id = $1
ORDER BY gb.publish_year ASC, gb.revision_number ASC, gb.created_at ASC;

-- name: ListGBProjectHistoryByProject :many
SELECT history.*
FROM gb_project current_project
JOIN LATERAL (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.green_book_id,
        gp.gb_code,
        gp.project_name,
        gb.publish_year,
        gb.revision_number,
        gb.status AS book_status,
        (gp.id = (
            SELECT latest.id
            FROM gb_project latest
            JOIN green_book latest_gb ON latest_gb.id = latest.green_book_id
            WHERE latest.gb_project_identity_id = gp.gb_project_identity_id
              AND latest.status = 'active'
            ORDER BY latest_gb.publish_year DESC, latest_gb.revision_number DESC, latest_gb.created_at DESC
            LIMIT 1
        ))::boolean AS is_latest,
        EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            WHERE dpg.gb_project_id = gp.id
        )::boolean AS used_by_downstream
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.gb_project_identity_id = current_project.gb_project_identity_id
    ORDER BY gb.publish_year ASC, gb.revision_number ASC, gb.created_at ASC
) history ON TRUE
WHERE current_project.id = $1;

-- name: UpdateGBProject :one
UPDATE gb_project
SET program_title_id = $2,
    project_name = $3,
    duration = $4,
    objective = $5,
    scope_of_project = $6,
    updated_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING *;

-- name: ListGBProjectDeletionDependencies :many
WITH related_dk AS (
    SELECT DISTINCT
        dp.id,
        dk.subject,
        dk.letter_number,
        dk.date
    FROM dk_project_gb_project dpg
    JOIN dk_project dp ON dp.id = dpg.dk_project_id
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
    WHERE dpg.gb_project_id = $1
),
related_la AS (
    SELECT DISTINCT
        la.id,
        la.loan_code,
        rd.subject,
        rd.letter_number
    FROM related_dk rd
    JOIN loan_agreement la ON la.dk_project_id = rd.id
),
related_monitoring AS (
    SELECT DISTINCT
        md.id,
        md.budget_year,
        md.quarter,
        rla.loan_code,
        rla.subject,
        rla.letter_number
    FROM related_la rla
    JOIN monitoring_disbursement md ON md.loan_agreement_id = rla.id
)
SELECT
    'daftar_kegiatan_project'::text AS relation_type,
    id AS relation_id,
    COALESCE(letter_number, subject)::text AS relation_label,
    format('Daftar Kegiatan %s', COALESCE(letter_number, subject))::text AS relation_path
FROM related_dk
UNION ALL
SELECT
    'loan_agreement'::text AS relation_type,
    id AS relation_id,
    loan_code::text AS relation_label,
    format('Daftar Kegiatan %s -> Loan Agreement %s', COALESCE(letter_number, subject), loan_code)::text AS relation_path
FROM related_la
UNION ALL
SELECT
    'monitoring_disbursement'::text AS relation_type,
    id AS relation_id,
    format('%s %s', budget_year, quarter)::text AS relation_label,
    format('Daftar Kegiatan %s -> Loan Agreement %s -> Monitoring %s %s', COALESCE(letter_number, subject), loan_code, budget_year, quarter)::text AS relation_path
FROM related_monitoring
ORDER BY relation_type, relation_label;

-- name: HardDeleteGBProject :one
DELETE FROM gb_project gp
WHERE gp.green_book_id = $1
  AND gp.id = $2
  AND NOT EXISTS (
      SELECT 1
      FROM dk_project_gb_project dpg
      WHERE dpg.gb_project_id = gp.id
  )
RETURNING *;

-- name: DeleteOrphanGBProjectIdentity :exec
DELETE FROM gb_project_identity gpi
WHERE gpi.id = $1
  AND NOT EXISTS (
      SELECT 1
      FROM gb_project gp
      WHERE gp.gb_project_identity_id = gpi.id
  );

-- ===== GB PROJECT BB PROJECT =====

-- name: ListActiveBBProjectReferences :many
SELECT
    bp.id,
    bp.blue_book_id,
    bp.project_identity_id,
    bp.bb_code,
    bp.project_name,
    p.name AS period_name,
    bb.publish_date,
    bb.revision_number,
    bb.revision_year
FROM bb_project bp
JOIN blue_book bb ON bb.id = bp.blue_book_id
JOIN period p ON p.id = bb.period_id
WHERE bp.status = 'active'
  AND bp.id = (
      SELECT latest.id
      FROM bb_project latest
      JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
      WHERE latest.project_identity_id = bp.project_identity_id
        AND latest.status = 'active'
      ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
      LIMIT 1
  )
ORDER BY bp.bb_code ASC;

-- name: GetGBProjectBBProjects :many
SELECT bp.*
FROM gb_project_bb_project gbp
JOIN bb_project bp ON bp.id = gbp.bb_project_id
WHERE gbp.gb_project_id = $1
ORDER BY bp.bb_code;

-- name: CloneGBProjectBBProjectsWithLatestBB :exec
INSERT INTO gb_project_bb_project (gb_project_id, bb_project_id)
SELECT
    $2,
    latest.id
FROM gb_project_bb_project source_link
JOIN bb_project source_bp ON source_bp.id = source_link.bb_project_id
JOIN LATERAL (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.project_identity_id = source_bp.project_identity_id
      AND bp.status = 'active'
    ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC
    LIMIT 1
) latest ON TRUE
WHERE source_link.gb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: AddGBProjectBBProject :exec
INSERT INTO gb_project_bb_project (gb_project_id, bb_project_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteGBProjectBBProjects :exec
DELETE FROM gb_project_bb_project
WHERE gb_project_id = $1;

-- name: CountDistinctBlueBooksForBBProjects :one
SELECT COUNT(DISTINCT blue_book_id)
FROM bb_project
WHERE id = ANY($1::uuid[]);

-- ===== GB BAPPENAS PARTNERS =====

-- name: GetGBProjectBappenasPartners :many
SELECT bp.*
FROM gb_project_bappenas_partner gpbp
JOIN bappenas_partner bp ON bp.id = gpbp.bappenas_partner_id
WHERE gpbp.gb_project_id = $1
ORDER BY bp.name;

-- name: AddGBProjectBappenasPartner :exec
INSERT INTO gb_project_bappenas_partner (gb_project_id, bappenas_partner_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteGBProjectBappenasPartners :exec
DELETE FROM gb_project_bappenas_partner
WHERE gb_project_id = $1;

-- name: CloneGBProjectBappenasPartners :exec
INSERT INTO gb_project_bappenas_partner (gb_project_id, bappenas_partner_id)
SELECT $2, source.bappenas_partner_id
FROM gb_project_bappenas_partner source
WHERE source.gb_project_id = $1
ON CONFLICT DO NOTHING;

-- ===== GB PROJECT INSTITUTIONS =====

-- name: GetGBProjectInstitutions :many
SELECT
    gpi.role,
    i.id,
    i.parent_id,
    i.name,
    i.short_name,
    i.level,
    i.created_at,
    i.updated_at
FROM gb_project_institution gpi
JOIN institution i ON i.id = gpi.institution_id
WHERE gpi.gb_project_id = $1
ORDER BY gpi.role, i.name;

-- name: AddGBProjectInstitution :exec
INSERT INTO gb_project_institution (gb_project_id, institution_id, role)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: DeleteGBProjectInstitutions :exec
DELETE FROM gb_project_institution
WHERE gb_project_id = $1;

-- ===== GB PROJECT LOCATIONS =====

-- name: GetGBProjectLocations :many
SELECT
    r.id,
    r.code,
    r.name,
    r.type,
    r.parent_code,
    r.created_at,
    r.updated_at
FROM gb_project_location gpl
JOIN region r ON r.id = gpl.region_id
WHERE gpl.gb_project_id = $1
ORDER BY r.code;

-- name: AddGBProjectLocation :exec
INSERT INTO gb_project_location (gb_project_id, region_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteGBProjectLocations :exec
DELETE FROM gb_project_location
WHERE gb_project_id = $1;

-- name: CloneGBProjectInstitutions :exec
INSERT INTO gb_project_institution (gb_project_id, institution_id, role)
SELECT $2, source.institution_id, source.role
FROM gb_project_institution source
WHERE source.gb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneGBProjectLocations :exec
INSERT INTO gb_project_location (gb_project_id, region_id)
SELECT $2, source.region_id
FROM gb_project_location source
WHERE source.gb_project_id = $1
ON CONFLICT DO NOTHING;

-- ===== GB ACTIVITY =====

-- name: ListGBActivitiesByProject :many
SELECT *
FROM gb_activity
WHERE gb_project_id = $1
ORDER BY sort_order ASC;

-- name: CreateGBActivity :one
INSERT INTO gb_activity (gb_project_id, activity_name, implementation_location, piu, sort_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateGBActivity :one
UPDATE gb_activity
SET activity_name = $2,
    implementation_location = $3,
    piu = $4,
    sort_order = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGBActivities :exec
DELETE FROM gb_activity
WHERE gb_project_id = $1;

-- name: CloneGBActivity :one
INSERT INTO gb_activity (gb_project_id, activity_name, implementation_location, piu, sort_order)
SELECT $2, source.activity_name, source.implementation_location, source.piu, source.sort_order
FROM gb_activity source
WHERE source.id = $1
RETURNING *;

-- ===== GB FUNDING SOURCE =====

-- name: ListGBFundingSourcesByProject :many
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
    gfs.created_at,
    gfs.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name,
    i.name AS institution_name,
    i.short_name AS institution_short_name,
    i.level AS institution_level
FROM gb_funding_source gfs
JOIN lender l ON l.id = gfs.lender_id
LEFT JOIN institution i ON i.id = gfs.institution_id
WHERE gfs.gb_project_id = $1
ORDER BY l.name;

-- name: CreateGBFundingSource :one
INSERT INTO gb_funding_source (
    gb_project_id,
    lender_id,
    institution_id,
    currency,
    loan_original,
    grant_original,
    local_original,
    loan_usd,
    grant_usd,
    local_usd
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateGBFundingSource :one
UPDATE gb_funding_source
SET lender_id = $2,
    institution_id = $3,
    currency = $4,
    loan_original = $5,
    grant_original = $6,
    local_original = $7,
    loan_usd = $8,
    grant_usd = $9,
    local_usd = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGBFundingSources :exec
DELETE FROM gb_funding_source
WHERE gb_project_id = $1;

-- name: CloneGBFundingSources :exec
INSERT INTO gb_funding_source (
    gb_project_id,
    lender_id,
    institution_id,
    currency,
    loan_original,
    grant_original,
    local_original,
    loan_usd,
    grant_usd,
    local_usd
)
SELECT
    $2,
    source.lender_id,
    source.institution_id,
    source.currency,
    source.loan_original,
    source.grant_original,
    source.local_original,
    source.loan_usd,
    source.grant_usd,
    source.local_usd
FROM gb_funding_source source
WHERE source.gb_project_id = $1;

-- ===== GB DISBURSEMENT PLAN =====

-- name: ListGBDisbursementPlansByProject :many
SELECT *
FROM gb_disbursement_plan
WHERE gb_project_id = $1
ORDER BY year ASC;

-- name: UpsertGBDisbursementPlan :one
INSERT INTO gb_disbursement_plan (gb_project_id, year, amount_usd)
VALUES ($1, $2, $3)
ON CONFLICT (gb_project_id, year) DO UPDATE
SET amount_usd = $3,
    updated_at = NOW()
RETURNING *;

-- name: DeleteGBDisbursementPlans :exec
DELETE FROM gb_disbursement_plan
WHERE gb_project_id = $1;

-- name: CloneGBDisbursementPlans :exec
INSERT INTO gb_disbursement_plan (gb_project_id, year, amount_usd)
SELECT $2, source.year, source.amount_usd
FROM gb_disbursement_plan source
WHERE source.gb_project_id = $1
ON CONFLICT (gb_project_id, year) DO UPDATE
SET amount_usd = EXCLUDED.amount_usd,
    updated_at = NOW();

-- ===== GB FUNDING ALLOCATION =====

-- name: ListGBFundingAllocationsByProject :many
SELECT
    gfa.id,
    gfa.gb_activity_id,
    ga.gb_project_id,
    ga.activity_name,
    ga.sort_order,
    gfa.services,
    gfa.constructions,
    gfa.goods,
    gfa.trainings,
    gfa.other,
    gfa.created_at,
    gfa.updated_at
FROM gb_funding_allocation gfa
JOIN gb_activity ga ON ga.id = gfa.gb_activity_id
WHERE ga.gb_project_id = $1
ORDER BY ga.sort_order ASC;

-- name: CreateGBFundingAllocation :one
INSERT INTO gb_funding_allocation (gb_activity_id, services, constructions, goods, trainings, other)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateGBFundingAllocation :one
UPDATE gb_funding_allocation
SET services = $2,
    constructions = $3,
    goods = $4,
    trainings = $5,
    other = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGBFundingAllocationsByProject :exec
DELETE FROM gb_funding_allocation
WHERE gb_activity_id IN (
    SELECT id
    FROM gb_activity
    WHERE gb_project_id = $1
);

-- name: CloneGBFundingAllocation :exec
INSERT INTO gb_funding_allocation (gb_activity_id, services, constructions, goods, trainings, other)
SELECT $2, source.services, source.constructions, source.goods, source.trainings, source.other
FROM gb_funding_allocation source
WHERE source.gb_activity_id = $1;
