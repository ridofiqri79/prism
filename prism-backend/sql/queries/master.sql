-- ===== COUNTRY =====
-- name: ListCountries :many
SELECT *
FROM country
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountCountries :one
SELECT COUNT(*)
FROM country;

-- name: GetCountry :one
SELECT *
FROM country
WHERE id = $1;

-- name: CreateCountry :one
INSERT INTO country (name, code)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateCountry :one
UPDATE country
SET name = $2,
    code = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCountry :exec
DELETE FROM country
WHERE id = $1;

-- ===== LENDER =====
-- name: ListLenders :many
SELECT
    l.id,
    l.country_id,
    l.name,
    l.short_name,
    l.type,
    l.created_at,
    l.updated_at,
    c.name AS country_name,
    c.code AS country_code
FROM lender l
LEFT JOIN country c ON c.id = l.country_id
WHERE (sqlc.narg('type_filter')::text IS NULL OR l.type = sqlc.narg('type_filter'))
ORDER BY l.name ASC
LIMIT $1 OFFSET $2;

-- name: CountLenders :one
SELECT COUNT(*)
FROM lender
WHERE (sqlc.narg('type_filter')::text IS NULL OR type = sqlc.narg('type_filter'));

-- name: GetLender :one
SELECT
    l.id,
    l.country_id,
    l.name,
    l.short_name,
    l.type,
    l.created_at,
    l.updated_at,
    c.name AS country_name,
    c.code AS country_code
FROM lender l
LEFT JOIN country c ON c.id = l.country_id
WHERE l.id = $1;

-- name: CreateLender :one
INSERT INTO lender (country_id, name, short_name, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateLender :one
UPDATE lender
SET country_id = $2,
    name = $3,
    short_name = $4,
    type = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLender :exec
DELETE FROM lender
WHERE id = $1;

-- ===== INSTITUTION =====
-- name: ListInstitutions :many
SELECT *
FROM institution
WHERE (sqlc.narg('level_filter')::text IS NULL OR level = sqlc.narg('level_filter'))
  AND (sqlc.narg('parent_id_filter')::uuid IS NULL OR parent_id = sqlc.narg('parent_id_filter'))
ORDER BY level ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: CountInstitutions :one
SELECT COUNT(*)
FROM institution
WHERE (sqlc.narg('level_filter')::text IS NULL OR level = sqlc.narg('level_filter'))
  AND (sqlc.narg('parent_id_filter')::uuid IS NULL OR parent_id = sqlc.narg('parent_id_filter'));

-- name: GetInstitution :one
SELECT
    i.id,
    i.parent_id,
    i.name,
    i.short_name,
    i.level,
    i.created_at,
    i.updated_at,
    p.name AS parent_name
FROM institution i
LEFT JOIN institution p ON p.id = i.parent_id
WHERE i.id = $1;

-- name: CreateInstitution :one
INSERT INTO institution (parent_id, name, short_name, level)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateInstitution :one
UPDATE institution
SET parent_id = $2,
    name = $3,
    short_name = $4,
    level = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteInstitution :exec
DELETE FROM institution
WHERE id = $1;

-- ===== REGION =====
-- name: ListRegions :many
SELECT *
FROM region
WHERE (sqlc.narg('type_filter')::text IS NULL OR type = sqlc.narg('type_filter'))
  AND (sqlc.narg('parent_code_filter')::text IS NULL OR parent_code = sqlc.narg('parent_code_filter'))
ORDER BY type ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: CountRegions :one
SELECT COUNT(*)
FROM region
WHERE (sqlc.narg('type_filter')::text IS NULL OR type = sqlc.narg('type_filter'))
  AND (sqlc.narg('parent_code_filter')::text IS NULL OR parent_code = sqlc.narg('parent_code_filter'));

-- name: GetRegion :one
SELECT *
FROM region
WHERE id = $1;

-- name: CreateRegion :one
INSERT INTO region (code, name, type, parent_code)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateRegion :one
UPDATE region
SET code = $2,
    name = $3,
    type = $4,
    parent_code = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRegion :exec
DELETE FROM region
WHERE id = $1;

-- ===== PROGRAM TITLE =====
-- name: ListProgramTitles :many
SELECT *
FROM program_title
ORDER BY title ASC
LIMIT $1 OFFSET $2;

-- name: CountProgramTitles :one
SELECT COUNT(*)
FROM program_title;

-- name: GetProgramTitle :one
SELECT *
FROM program_title
WHERE id = $1;

-- name: CreateProgramTitle :one
INSERT INTO program_title (parent_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateProgramTitle :one
UPDATE program_title
SET parent_id = $2,
    title = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProgramTitle :exec
DELETE FROM program_title
WHERE id = $1;

-- ===== BAPPENAS PARTNER =====
-- name: ListBappenasPartners :many
SELECT *
FROM bappenas_partner
ORDER BY level ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: CountBappenasPartners :one
SELECT COUNT(*)
FROM bappenas_partner;

-- name: GetBappenasPartner :one
SELECT *
FROM bappenas_partner
WHERE id = $1;

-- name: CreateBappenasPartner :one
INSERT INTO bappenas_partner (parent_id, name, level)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateBappenasPartner :one
UPDATE bappenas_partner
SET parent_id = $2,
    name = $3,
    level = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBappenasPartner :exec
DELETE FROM bappenas_partner
WHERE id = $1;

-- ===== PERIOD =====
-- name: ListPeriods :many
SELECT *
FROM period
ORDER BY year_start DESC
LIMIT $1 OFFSET $2;

-- name: CountPeriods :one
SELECT COUNT(*)
FROM period;

-- name: GetPeriod :one
SELECT *
FROM period
WHERE id = $1;

-- name: CreatePeriod :one
INSERT INTO period (name, year_start, year_end)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdatePeriod :one
UPDATE period
SET name = $2,
    year_start = $3,
    year_end = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePeriod :exec
DELETE FROM period
WHERE id = $1;

-- ===== NATIONAL PRIORITY =====
-- name: ListNationalPriorities :many
SELECT
    np.id,
    np.period_id,
    np.title,
    np.created_at,
    np.updated_at,
    p.name AS period_name
FROM national_priority np
JOIN period p ON p.id = np.period_id
WHERE (sqlc.narg('period_id_filter')::uuid IS NULL OR np.period_id = sqlc.narg('period_id_filter'))
ORDER BY np.title ASC
LIMIT $1 OFFSET $2;

-- name: CountNationalPriorities :one
SELECT COUNT(*)
FROM national_priority
WHERE (sqlc.narg('period_id_filter')::uuid IS NULL OR period_id = sqlc.narg('period_id_filter'));

-- name: GetNationalPriority :one
SELECT
    np.id,
    np.period_id,
    np.title,
    np.created_at,
    np.updated_at,
    p.name AS period_name
FROM national_priority np
JOIN period p ON p.id = np.period_id
WHERE np.id = $1;

-- name: CreateNationalPriority :one
INSERT INTO national_priority (period_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateNationalPriority :one
UPDATE national_priority
SET period_id = $2,
    title = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteNationalPriority :exec
DELETE FROM national_priority
WHERE id = $1;
