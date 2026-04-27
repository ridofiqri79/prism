-- ===== BLUE BOOK =====

-- name: ListBlueBooks :many
SELECT
    bb.id,
    bb.period_id,
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
INSERT INTO blue_book (period_id, publish_date, revision_number, revision_year, status)
VALUES ($1, $2, $3, $4, 'active')
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

-- name: SupersedeBlueBook :one
UPDATE blue_book
SET status = 'superseded',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- ===== BB PROJECT =====

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

-- name: CreateBBProject :one
INSERT INTO bb_project (
    blue_book_id,
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

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
