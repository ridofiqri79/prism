-- ===== BLUE BOOK =====

-- name: ListBlueBooks :many
SELECT
    bb.id,
    bb.period_id,
    bb.replaces_blue_book_id,
    bb.publish_date,
    bb.revision_number,
    bb.revision_year,
    bb.status,
    bb.created_at,
    bb.updated_at,
    p.name AS period_name,
    p.year_start,
    p.year_end
FROM blue_book bb
JOIN period p ON p.id = bb.period_id
ORDER BY bb.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountBlueBooks :one
SELECT COUNT(*) FROM blue_book;

-- name: GetBlueBook :one
SELECT
    bb.id,
    bb.period_id,
    bb.replaces_blue_book_id,
    bb.publish_date,
    bb.revision_number,
    bb.revision_year,
    bb.status,
    bb.created_at,
    bb.updated_at,
    p.name AS period_name,
    p.year_start,
    p.year_end
FROM blue_book bb
JOIN period p ON p.id = bb.period_id
WHERE bb.id = $1;

-- name: CreateBlueBook :one
INSERT INTO blue_book (period_id, replaces_blue_book_id, publish_date, revision_number, revision_year, status)
VALUES ($1, $2, $3, $4, $5, 'active')
RETURNING *;

-- name: UpdateBlueBook :one
UPDATE blue_book
SET publish_date = $2,
    revision_number = $3,
    revision_year = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SupersedeBlueBooksByPeriod :exec
UPDATE blue_book
SET status = 'superseded',
    updated_at = NOW()
WHERE period_id = $1
  AND status = 'active';

-- name: GetActiveBlueBookByPeriod :one
SELECT *
FROM blue_book
WHERE period_id = $1
  AND status = 'active'
ORDER BY revision_number DESC, created_at DESC
LIMIT 1;

-- name: SupersedeBlueBook :one
UPDATE blue_book
SET status = 'superseded',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- ===== BB PROJECT =====

-- name: CreateProjectIdentity :one
INSERT INTO project_identity DEFAULT VALUES
RETURNING *;

-- name: GetProjectIdentity :one
SELECT *
FROM project_identity
WHERE id = $1;

-- name: ListBBProjectsByBlueBook :many
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND status = 'active'
ORDER BY bb_code ASC
LIMIT $2 OFFSET $3;

-- name: CountBBProjectsByBlueBook :one
SELECT COUNT(*)
FROM bb_project
WHERE blue_book_id = $1
  AND status = 'active';

-- name: GetBBProject :one
SELECT *
FROM bb_project
WHERE id = $1;

-- name: GetActiveBBProjectByBlueBook :one
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND id = $2
  AND status = 'active';

-- name: GetBBProjectByCode :one
SELECT *
FROM bb_project
WHERE bb_code = $1;

-- name: GetBBProjectByBlueBookAndCode :one
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND LOWER(bb_code) = LOWER($2)
LIMIT 1;

-- name: FindPreviousBBProjectByCodeForBlueBook :one
SELECT bp.*
FROM bb_project bp
JOIN blue_book source_bb ON source_bb.id = bp.blue_book_id
JOIN blue_book target_bb ON target_bb.id = $1
WHERE LOWER(bp.bb_code) = LOWER($2)
  AND source_bb.period_id = target_bb.period_id
  AND source_bb.id <> target_bb.id
  AND (
      source_bb.revision_number < target_bb.revision_number
      OR (
          source_bb.revision_number = target_bb.revision_number
          AND source_bb.created_at < target_bb.created_at
      )
  )
ORDER BY source_bb.revision_number DESC, source_bb.created_at DESC
LIMIT 1;

-- name: CreateBBProject :one
INSERT INTO bb_project (
    blue_book_id,
    project_identity_id,
    program_title_id,
    bappenas_partner_id,
    bb_code,
    project_name,
    duration,
    objective,
    scope_of_work,
    outputs,
    outcomes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListBBProjectsForClone :many
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND status = 'active'
ORDER BY bb_code ASC;

-- name: GetLatestBBProjectByIdentity :one
SELECT bp.*
FROM bb_project bp
JOIN blue_book bb ON bb.id = bp.blue_book_id
WHERE bp.project_identity_id = $1
  AND bp.status = 'active'
ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC
LIMIT 1;

-- name: GetLatestBBProjectByProject :one
SELECT latest.*
FROM bb_project current_project
JOIN LATERAL (
    SELECT bp.*
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.project_identity_id = current_project.project_identity_id
      AND bp.status = 'active'
    ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC
    LIMIT 1
) latest ON TRUE
WHERE current_project.id = $1;

-- name: ListBBProjectHistoryByIdentity :many
SELECT
    bp.id,
    bp.project_identity_id,
    bp.blue_book_id,
    bp.bb_code,
    bp.project_name,
    p.name AS period_name,
    bb.revision_number,
    bb.revision_year,
    bb.status AS book_status,
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
        FROM gb_project_bb_project gbp
        WHERE gbp.bb_project_id = bp.id
    )::boolean AS used_by_downstream
FROM bb_project bp
JOIN blue_book bb ON bb.id = bp.blue_book_id
JOIN period p ON p.id = bb.period_id
WHERE bp.project_identity_id = $1
ORDER BY bb.revision_number ASC, COALESCE(bb.revision_year, 0) ASC, bb.created_at ASC;

-- name: ListBBProjectHistoryByProject :many
SELECT history.*
FROM bb_project current_project
JOIN LATERAL (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.blue_book_id,
        bp.bb_code,
        bp.project_name,
        p.name AS period_name,
        bb.revision_number,
        bb.revision_year,
        bb.status AS book_status,
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
            FROM gb_project_bb_project gbp
            WHERE gbp.bb_project_id = bp.id
        )::boolean AS used_by_downstream
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    JOIN period p ON p.id = bb.period_id
    WHERE bp.project_identity_id = current_project.project_identity_id
    ORDER BY bb.revision_number ASC, COALESCE(bb.revision_year, 0) ASC, bb.created_at ASC
) history ON TRUE
WHERE current_project.id = $1;

-- name: CloneBBProjectInstitutions :exec
INSERT INTO bb_project_institution (bb_project_id, institution_id, role)
SELECT $2, source.institution_id, source.role
FROM bb_project_institution source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectLocations :exec
INSERT INTO bb_project_location (bb_project_id, region_id)
SELECT $2, source.region_id
FROM bb_project_location source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectNationalPriorities :exec
INSERT INTO bb_project_national_priority (bb_project_id, national_priority_id)
SELECT $2, source.national_priority_id
FROM bb_project_national_priority source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectCosts :exec
INSERT INTO bb_project_cost (bb_project_id, funding_type, funding_category, amount_usd)
SELECT $2, source.funding_type, source.funding_category, source.amount_usd
FROM bb_project_cost source
WHERE source.bb_project_id = $1;

-- name: CloneLenderIndications :exec
INSERT INTO lender_indication (bb_project_id, lender_id, remarks)
SELECT $2, source.lender_id, source.remarks
FROM lender_indication source
WHERE source.bb_project_id = $1;

-- name: CloneLoIs :exec
INSERT INTO loi (bb_project_id, lender_id, subject, date, letter_number)
SELECT $2, source.lender_id, source.subject, source.date, source.letter_number
FROM loi source
WHERE source.bb_project_id = $1;

-- name: UpdateBBProject :one
UPDATE bb_project
SET program_title_id = $2,
    bappenas_partner_id = $3,
    project_name = $4,
    duration = $5,
    objective = $6,
    scope_of_work = $7,
    outputs = $8,
    outcomes = $9,
    updated_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING *;

-- name: SoftDeleteBBProject :one
UPDATE bb_project
SET status = 'deleted',
    updated_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING *;

-- ===== BB INSTITUTIONS =====

-- name: GetBBProjectInstitutions :many
SELECT
    bpi.role,
    i.id,
    i.parent_id,
    i.name,
    i.short_name,
    i.level,
    i.created_at,
    i.updated_at
FROM bb_project_institution bpi
JOIN institution i ON i.id = bpi.institution_id
WHERE bpi.bb_project_id = $1
ORDER BY bpi.role, i.name;

-- name: AddBBProjectInstitution :exec
INSERT INTO bb_project_institution (bb_project_id, institution_id, role)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectInstitutions :exec
DELETE FROM bb_project_institution
WHERE bb_project_id = $1;

-- ===== BB LOCATIONS =====

-- name: GetBBProjectLocations :many
SELECT
    r.id,
    r.code,
    r.name,
    r.type,
    r.parent_code,
    r.created_at,
    r.updated_at
FROM bb_project_location bpl
JOIN region r ON r.id = bpl.region_id
WHERE bpl.bb_project_id = $1
ORDER BY r.code;

-- name: AddBBProjectLocation :exec
INSERT INTO bb_project_location (bb_project_id, region_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectLocations :exec
DELETE FROM bb_project_location
WHERE bb_project_id = $1;

-- ===== BB NATIONAL PRIORITIES =====

-- name: GetBBProjectNationalPriorities :many
SELECT
    np.id,
    np.period_id,
    np.title,
    np.created_at,
    np.updated_at
FROM bb_project_national_priority bpnp
JOIN national_priority np ON np.id = bpnp.national_priority_id
WHERE bpnp.bb_project_id = $1
ORDER BY np.title;

-- name: AddBBProjectNationalPriority :exec
INSERT INTO bb_project_national_priority (bb_project_id, national_priority_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectNationalPriorities :exec
DELETE FROM bb_project_national_priority
WHERE bb_project_id = $1;

-- name: CountMismatchedBBProjectNationalPriorities :one
SELECT COUNT(*)
FROM national_priority np
JOIN blue_book bb ON bb.id = $1
WHERE np.id = ANY($2::uuid[])
  AND np.period_id <> bb.period_id;

-- ===== PROJECT COSTS =====

-- name: GetBBProjectCosts :many
SELECT *
FROM bb_project_cost
WHERE bb_project_id = $1
ORDER BY created_at ASC;

-- name: CreateBBProjectCost :one
INSERT INTO bb_project_cost (bb_project_id, funding_type, funding_category, amount_usd)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteBBProjectCosts :exec
DELETE FROM bb_project_cost
WHERE bb_project_id = $1;

-- ===== LENDER INDICATION =====

-- name: GetLenderIndications :many
SELECT
    li.id,
    li.bb_project_id,
    li.lender_id,
    li.remarks,
    li.created_at,
    li.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM lender_indication li
JOIN lender l ON l.id = li.lender_id
WHERE li.bb_project_id = $1
ORDER BY l.name;

-- name: CreateLenderIndication :one
INSERT INTO lender_indication (bb_project_id, lender_id, remarks)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteLenderIndications :exec
DELETE FROM lender_indication
WHERE bb_project_id = $1;

-- ===== LoI =====

-- name: GetLoIsByBBProject :many
SELECT
    loi.id,
    loi.bb_project_id,
    loi.lender_id,
    loi.subject,
    loi.date,
    loi.letter_number,
    loi.created_at,
    loi.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loi
JOIN lender l ON l.id = loi.lender_id
WHERE loi.bb_project_id = $1
ORDER BY loi.date DESC;

-- name: CreateLoI :one
INSERT INTO loi (bb_project_id, lender_id, subject, date, letter_number)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteLoI :exec
DELETE FROM loi
WHERE id = $1
  AND bb_project_id = $2;
