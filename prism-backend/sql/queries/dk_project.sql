-- ===== DAFTAR KEGIATAN =====

-- name: ListDaftarKegiatan :many
SELECT *
FROM daftar_kegiatan
ORDER BY date DESC, created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDaftarKegiatan :one
SELECT COUNT(*) FROM daftar_kegiatan;

-- name: GetDaftarKegiatan :one
SELECT *
FROM daftar_kegiatan
WHERE id = $1;

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

-- name: DeleteDaftarKegiatan :exec
DELETE FROM daftar_kegiatan
WHERE id = $1;

-- ===== DK PROJECT =====

-- name: ListDKProjectsByDK :many
SELECT *
FROM dk_project
WHERE dk_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountDKProjectsByDK :one
SELECT COUNT(*)
FROM dk_project
WHERE dk_id = $1;

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
INSERT INTO dk_project (dk_id, program_title_id, institution_id, duration, objectives)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateDKProject :one
UPDATE dk_project
SET program_title_id = $2,
    institution_id = $3,
    duration = $4,
    objectives = $5,
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
