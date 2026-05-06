-- ===== DAFTAR KEGIATAN =====

-- name: ListDaftarKegiatan :many
SELECT
    dk.id,
    dk.letter_number,
    dk.subject,
    dk.date,
    dk.created_at,
    dk.updated_at,
    (
        SELECT COUNT(*)
        FROM dk_project dkp
        WHERE dkp.dk_id = dk.id
    )::BIGINT AS project_count
FROM daftar_kegiatan dk
WHERE (
    sqlc.narg('search')::text IS NULL
    OR dk.subject ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(dk.letter_number, '') ILIKE '%' || sqlc.narg('search')::text || '%'
    OR dk.date::text ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (sqlc.narg('date_from')::date IS NULL OR dk.date >= sqlc.narg('date_from')::date)
AND (sqlc.narg('date_to')::date IS NULL OR dk.date <= sqlc.narg('date_to')::date)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'subject' AND sqlc.arg('sort_order')::text = 'asc' THEN dk.subject END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'subject' AND sqlc.arg('sort_order')::text = 'desc' THEN dk.subject END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'date' AND sqlc.arg('sort_order')::text = 'asc' THEN dk.date END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'date' AND sqlc.arg('sort_order')::text = 'desc' THEN dk.date END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'letter_number' AND sqlc.arg('sort_order')::text = 'asc' THEN dk.letter_number END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'letter_number' AND sqlc.arg('sort_order')::text = 'desc' THEN dk.letter_number END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_count' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT COUNT(*)
        FROM dk_project dkp
        WHERE dkp.dk_id = dk.id
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_count' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT COUNT(*)
        FROM dk_project dkp
        WHERE dkp.dk_id = dk.id
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN dk.created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN dk.created_at END DESC,
    dk.date DESC,
    dk.created_at DESC,
    dk.id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDaftarKegiatan :one
SELECT COUNT(*)
FROM daftar_kegiatan
WHERE (
    sqlc.narg('search')::text IS NULL
    OR subject ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(letter_number, '') ILIKE '%' || sqlc.narg('search')::text || '%'
    OR date::text ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (sqlc.narg('date_from')::date IS NULL OR date >= sqlc.narg('date_from')::date)
AND (sqlc.narg('date_to')::date IS NULL OR date <= sqlc.narg('date_to')::date);

-- name: GetDaftarKegiatan :one
SELECT
    dk.id,
    dk.letter_number,
    dk.subject,
    dk.date,
    dk.created_at,
    dk.updated_at,
    (
        SELECT COUNT(*)
        FROM dk_project dkp
        WHERE dkp.dk_id = dk.id
    )::BIGINT AS project_count
FROM daftar_kegiatan dk
WHERE dk.id = $1;

-- name: GetDaftarKegiatanByLetterNumber :one
SELECT *
FROM daftar_kegiatan
WHERE LOWER(letter_number) = LOWER($1)
LIMIT 1;

-- name: CreateDaftarKegiatan :one
INSERT INTO daftar_kegiatan (letter_number, subject, date)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateDaftarKegiatan :one
UPDATE daftar_kegiatan
SET letter_number = $2,
    subject = $3,
    date = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountAnyDKProjectsByDaftarKegiatan :one
SELECT COUNT(*)
FROM dk_project
WHERE dk_id = $1;

-- name: HardDeleteDaftarKegiatan :one
DELETE FROM daftar_kegiatan
WHERE daftar_kegiatan.id = $1
  AND NOT EXISTS (
      SELECT 1
      FROM dk_project dkp
      WHERE dkp.dk_id = daftar_kegiatan.id
  )
RETURNING *;

-- ===== DK PROJECT =====

-- name: ListDKProjectsByDK :many
SELECT *
FROM dk_project
WHERE dk_id = sqlc.arg('dk_id')
AND (
    sqlc.narg('search')::text IS NULL
    OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(objectives, '') ILIKE '%' || sqlc.narg('search')::text || '%'
    OR EXISTS (
        SELECT 1
        FROM institution i
        WHERE i.id = dk_project.institution_id
          AND (
              i.name ILIKE '%' || sqlc.narg('search')::text || '%'
              OR COALESCE(i.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
          )
    )
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dkgb
        JOIN gb_project gp ON gp.id = dkgb.gb_project_id
        WHERE dkgb.dk_project_id = dk_project.id
          AND (
              gp.gb_code ILIKE '%' || sqlc.narg('search')::text || '%'
              OR gp.project_name ILIKE '%' || sqlc.narg('search')::text || '%'
          )
    )
)
AND (
    COALESCE(cardinality(sqlc.arg('gb_project_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dkgb
        WHERE dkgb.dk_project_id = dk_project.id
          AND dkgb.gb_project_id = ANY(sqlc.arg('gb_project_ids')::uuid[])
    )
)
AND (
    COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
    OR institution_id = ANY(sqlc.arg('executing_agency_ids')::uuid[])
)
AND (
    COALESCE(cardinality(sqlc.arg('location_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_project_location dkpl
        WHERE dkpl.dk_project_id = dk_project.id
          AND dkpl.region_id = ANY(sqlc.arg('location_ids')::uuid[])
    )
)
AND (
    COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_financing_detail dfd
        WHERE dfd.dk_project_id = dk_project.id
          AND dfd.lender_id = ANY(sqlc.arg('lender_ids')::uuid[])
    )
)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'project_name' AND sqlc.arg('sort_order')::text = 'asc' THEN project_name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_name' AND sqlc.arg('sort_order')::text = 'desc' THEN project_name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'executing_agency' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT COALESCE(i.short_name, i.name)
        FROM institution i
        WHERE i.id = dk_project.institution_id
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'executing_agency' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT COALESCE(i.short_name, i.name)
        FROM institution i
        WHERE i.id = dk_project.institution_id
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'duration' AND sqlc.arg('sort_order')::text = 'asc' THEN duration END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'duration' AND sqlc.arg('sort_order')::text = 'desc' THEN duration END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN created_at END DESC,
    created_at DESC,
    id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDKProjectsByDK :one
SELECT COUNT(*)
FROM dk_project
WHERE dk_id = sqlc.arg('dk_id')
AND (
    sqlc.narg('search')::text IS NULL
    OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(objectives, '') ILIKE '%' || sqlc.narg('search')::text || '%'
    OR EXISTS (
        SELECT 1
        FROM institution i
        WHERE i.id = dk_project.institution_id
          AND (
              i.name ILIKE '%' || sqlc.narg('search')::text || '%'
              OR COALESCE(i.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
          )
    )
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dkgb
        JOIN gb_project gp ON gp.id = dkgb.gb_project_id
        WHERE dkgb.dk_project_id = dk_project.id
          AND (
              gp.gb_code ILIKE '%' || sqlc.narg('search')::text || '%'
              OR gp.project_name ILIKE '%' || sqlc.narg('search')::text || '%'
          )
    )
)
AND (
    COALESCE(cardinality(sqlc.arg('gb_project_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dkgb
        WHERE dkgb.dk_project_id = dk_project.id
          AND dkgb.gb_project_id = ANY(sqlc.arg('gb_project_ids')::uuid[])
    )
)
AND (
    COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
    OR institution_id = ANY(sqlc.arg('executing_agency_ids')::uuid[])
)
AND (
    COALESCE(cardinality(sqlc.arg('location_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_project_location dkpl
        WHERE dkpl.dk_project_id = dk_project.id
          AND dkpl.region_id = ANY(sqlc.arg('location_ids')::uuid[])
    )
)
AND (
    COALESCE(cardinality(sqlc.arg('lender_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_financing_detail dfd
        WHERE dfd.dk_project_id = dk_project.id
          AND dfd.lender_id = ANY(sqlc.arg('lender_ids')::uuid[])
    )
);

-- name: GetDKProject :one
SELECT *
FROM dk_project
WHERE id = $1;

-- name: GetDKProjectByDK :one
SELECT *
FROM dk_project
WHERE dk_id = $1
  AND id = $2;

-- name: CreateDKProject :one
INSERT INTO dk_project (dk_id, program_title_id, institution_id, project_name, duration, objectives)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateDKProject :one
UPDATE dk_project
SET program_title_id = $2,
    institution_id = $3,
    project_name = $4,
    duration = $5,
    objectives = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDKProject :exec
DELETE FROM dk_project
WHERE id = $1;

-- name: ListActiveGBProjectReferences :many
SELECT
    gp.id,
    gp.green_book_id,
    gp.gb_project_identity_id,
    gp.program_title_id,
    gp.gb_code,
    gp.project_name,
    gb.publish_year,
    gb.revision_number,
    gp.duration,
    gp.objective,
    gp.scope_of_project,
    gp.created_at,
    gp.updated_at
FROM gb_project gp
JOIN green_book gb ON gb.id = gp.green_book_id
WHERE gp.status = 'active'
  AND gp.id = (
      SELECT latest.id
      FROM gb_project latest
      JOIN green_book latest_gb ON latest_gb.id = latest.green_book_id
      WHERE latest.gb_project_identity_id = gp.gb_project_identity_id
        AND latest.status = 'active'
      ORDER BY latest_gb.revision_number DESC, latest_gb.created_at DESC
      LIMIT 1
  )
ORDER BY gp.gb_code ASC;

-- ===== DK PROJECT GB PROJECT =====

-- name: GetDKProjectGBProjects :many
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
    gp.updated_at
FROM dk_project_gb_project dkgb
JOIN gb_project gp ON gp.id = dkgb.gb_project_id
WHERE dkgb.dk_project_id = $1
ORDER BY gp.gb_code;

-- name: ResolveLatestGBProjectForDK :one
SELECT latest.*
FROM gb_project current_project
JOIN LATERAL (
    SELECT gp.*
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.gb_project_identity_id = current_project.gb_project_identity_id
      AND gp.status = 'active'
    ORDER BY gb.revision_number DESC, gb.created_at DESC
    LIMIT 1
) latest ON TRUE
WHERE current_project.id = $1;

-- name: AddDKProjectGBProject :exec
INSERT INTO dk_project_gb_project (dk_project_id, gb_project_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteDKProjectGBProjects :exec
DELETE FROM dk_project_gb_project
WHERE dk_project_id = $1;

-- ===== DK PROJECT BAPPENAS PARTNERS =====

-- name: GetDKProjectBappenasPartners :many
SELECT bp.*
FROM dk_project_bappenas_partner dpbp
JOIN bappenas_partner bp ON bp.id = dpbp.bappenas_partner_id
WHERE dpbp.dk_project_id = $1
ORDER BY bp.name;

-- name: AddDKProjectBappenasPartner :exec
INSERT INTO dk_project_bappenas_partner (dk_project_id, bappenas_partner_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteDKProjectBappenasPartners :exec
DELETE FROM dk_project_bappenas_partner
WHERE dk_project_id = $1;

-- ===== DK PROJECT LOCATION =====

-- name: GetDKProjectLocations :many
SELECT
    r.id,
    r.code,
    r.name,
    r.type,
    r.parent_code,
    r.created_at,
    r.updated_at
FROM dk_project_location dkpl
JOIN region r ON r.id = dkpl.region_id
WHERE dkpl.dk_project_id = $1
ORDER BY r.code;

-- name: AddDKProjectLocation :exec
INSERT INTO dk_project_location (dk_project_id, region_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteDKProjectLocations :exec
DELETE FROM dk_project_location
WHERE dk_project_id = $1;

-- ===== FINANCING DETAIL =====

-- name: GetDKFinancingDetails :many
SELECT
    df.id,
    df.dk_project_id,
    df.lender_id,
    df.currency,
    df.amount_original,
    df.grant_original,
    df.counterpart_original,
    df.amount_usd,
    df.grant_usd,
    df.counterpart_usd,
    df.remarks,
    df.created_at,
    df.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM dk_financing_detail df
LEFT JOIN lender l ON l.id = df.lender_id
WHERE df.dk_project_id = $1
ORDER BY l.name;

-- name: CreateDKFinancingDetail :one
INSERT INTO dk_financing_detail (
    dk_project_id,
    lender_id,
    currency,
    amount_original,
    grant_original,
    counterpart_original,
    amount_usd,
    grant_usd,
    counterpart_usd,
    remarks
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: DeleteDKFinancingDetails :exec
DELETE FROM dk_financing_detail
WHERE dk_project_id = $1;

-- ===== LOAN ALLOCATION =====

-- name: GetDKLoanAllocations :many
SELECT
    dla.id,
    dla.dk_project_id,
    dla.institution_id,
    dla.currency,
    dla.amount_original,
    dla.grant_original,
    dla.counterpart_original,
    dla.amount_usd,
    dla.grant_usd,
    dla.counterpart_usd,
    dla.remarks,
    dla.created_at,
    dla.updated_at,
    i.name AS institution_name,
    i.short_name AS institution_short_name,
    i.level AS institution_level
FROM dk_loan_allocation dla
LEFT JOIN institution i ON i.id = dla.institution_id
WHERE dla.dk_project_id = $1
ORDER BY i.name;

-- name: CreateDKLoanAllocation :one
INSERT INTO dk_loan_allocation (
    dk_project_id,
    institution_id,
    currency,
    amount_original,
    grant_original,
    counterpart_original,
    amount_usd,
    grant_usd,
    counterpart_usd,
    remarks
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: DeleteDKLoanAllocations :exec
DELETE FROM dk_loan_allocation
WHERE dk_project_id = $1;

-- ===== ACTIVITY DETAIL =====

-- name: GetDKActivityDetails :many
SELECT *
FROM dk_activity_detail
WHERE dk_project_id = $1
ORDER BY activity_number ASC;

-- name: CreateDKActivityDetail :one
INSERT INTO dk_activity_detail (dk_project_id, activity_number, activity_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteDKActivityDetails :exec
DELETE FROM dk_activity_detail
WHERE dk_project_id = $1;

-- name: GetAllowedLenderIDsForDK :many
SELECT DISTINCT lender_id
FROM (
    SELECT li.lender_id
    FROM dk_project_gb_project dkgb
    JOIN gb_project_bb_project gbbb ON gbbb.gb_project_id = dkgb.gb_project_id
    JOIN lender_indication li ON li.bb_project_id = gbbb.bb_project_id
    WHERE dkgb.dk_project_id = $1
    UNION
    SELECT gfs.lender_id
    FROM dk_project_gb_project dkgb
    JOIN gb_funding_source gfs ON gfs.gb_project_id = dkgb.gb_project_id
    WHERE dkgb.dk_project_id = $1
) allowed_lenders
WHERE lender_id IS NOT NULL;

-- name: ListAllowedLenderReferencesByGBProject :many
SELECT DISTINCT
    allowed_lenders.gb_project_id,
    allowed_lenders.lender_id,
    l.name AS lender_name,
    l.short_name AS lender_short_name,
    l.type AS lender_type
FROM (
    SELECT gfs.gb_project_id, gfs.lender_id
    FROM gb_funding_source gfs
    WHERE gfs.gb_project_id = $1
    UNION
    SELECT gbp.gb_project_id, li.lender_id
    FROM gb_project_bb_project gbp
    JOIN lender_indication li ON li.bb_project_id = gbp.bb_project_id
    WHERE gbp.gb_project_id = $1
) allowed_lenders
JOIN lender l ON l.id = allowed_lenders.lender_id
WHERE allowed_lenders.lender_id IS NOT NULL
ORDER BY l.name ASC;
