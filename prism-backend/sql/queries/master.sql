-- ===== COUNTRY =====
-- name: ListCountries :many
SELECT *
FROM country
WHERE (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'code' AND sqlc.arg('sort_order')::text = 'asc' THEN code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'code' AND sqlc.arg('sort_order')::text = 'desc' THEN code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN name END DESC,
    name ASC
LIMIT $1 OFFSET $2;

-- name: CountCountries :one
SELECT COUNT(*)
FROM country
WHERE (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
);

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

-- ===== CURRENCY =====
-- name: ListCurrencies :many
SELECT *
FROM currency
WHERE (sqlc.narg('active_filter')::boolean IS NULL OR is_active = sqlc.narg('active_filter'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'code' AND sqlc.arg('sort_order')::text = 'asc' THEN code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'code' AND sqlc.arg('sort_order')::text = 'desc' THEN code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'sort_order' AND sqlc.arg('sort_order')::text = 'asc' THEN sort_order END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'sort_order' AND sqlc.arg('sort_order')::text = 'desc' THEN sort_order END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'is_active' AND sqlc.arg('sort_order')::text = 'asc' THEN is_active END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'is_active' AND sqlc.arg('sort_order')::text = 'desc' THEN is_active END DESC,
    sort_order ASC,
    code ASC
LIMIT $1 OFFSET $2;

-- name: CountCurrencies :one
SELECT COUNT(*)
FROM currency
WHERE (sqlc.narg('active_filter')::boolean IS NULL OR is_active = sqlc.narg('active_filter'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: GetCurrency :one
SELECT *
FROM currency
WHERE id = $1;

-- name: GetCurrencyByCode :one
SELECT *
FROM currency
WHERE code = $1;

-- name: CreateCurrency :one
INSERT INTO currency (code, name, symbol, is_active, sort_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCurrency :one
UPDATE currency
SET code = $2,
    name = $3,
    symbol = $4,
    is_active = $5,
    sort_order = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCurrency :exec
DELETE FROM currency
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
WHERE (COALESCE(cardinality(sqlc.arg('type_filters')::text[]), 0) = 0 OR l.type = ANY(sqlc.arg('type_filters')::text[]))
  AND (
    sqlc.narg('search')::text IS NULL
    OR l.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(l.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN l.name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN l.name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'short_name' AND sqlc.arg('sort_order')::text = 'asc' THEN l.short_name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'short_name' AND sqlc.arg('sort_order')::text = 'desc' THEN l.short_name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'type' AND sqlc.arg('sort_order')::text = 'asc' THEN l.type END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'type' AND sqlc.arg('sort_order')::text = 'desc' THEN l.type END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'country' AND sqlc.arg('sort_order')::text = 'asc' THEN c.name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'country' AND sqlc.arg('sort_order')::text = 'desc' THEN c.name END DESC,
    l.name ASC
LIMIT $1 OFFSET $2;

-- name: CountLenders :one
SELECT COUNT(*)
FROM lender
WHERE (COALESCE(cardinality(sqlc.arg('type_filters')::text[]), 0) = 0 OR type = ANY(sqlc.arg('type_filters')::text[]))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
  );

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
WHERE (COALESCE(cardinality(sqlc.arg('level_filters')::text[]), 0) = 0 OR level = ANY(sqlc.arg('level_filters')::text[]))
  AND (sqlc.narg('parent_id_filter')::uuid IS NULL OR parent_id = sqlc.narg('parent_id_filter'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'short_name' AND sqlc.arg('sort_order')::text = 'asc' THEN short_name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'short_name' AND sqlc.arg('sort_order')::text = 'desc' THEN short_name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'level' AND sqlc.arg('sort_order')::text = 'asc' THEN level END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'level' AND sqlc.arg('sort_order')::text = 'desc' THEN level END DESC,
    level ASC,
    name ASC
LIMIT $1 OFFSET $2;

-- name: CountInstitutions :one
SELECT COUNT(*)
FROM institution
WHERE (COALESCE(cardinality(sqlc.arg('level_filters')::text[]), 0) = 0 OR level = ANY(sqlc.arg('level_filters')::text[]))
  AND (sqlc.narg('parent_id_filter')::uuid IS NULL OR parent_id = sqlc.narg('parent_id_filter'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
  );

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

-- name: CountInstitutionsByNameScope :one
SELECT COUNT(*)
FROM institution
WHERE LOWER(BTRIM(name)) = LOWER(BTRIM(sqlc.arg('name')::text))
  AND (
    (sqlc.narg('parent_id')::uuid IS NULL AND parent_id IS NULL)
    OR parent_id = sqlc.narg('parent_id')::uuid
  )
  AND (sqlc.narg('except_id')::uuid IS NULL OR id <> sqlc.narg('except_id')::uuid);

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
WHERE (COALESCE(cardinality(sqlc.arg('type_filters')::text[]), 0) = 0 OR type = ANY(sqlc.arg('type_filters')::text[]))
  AND (sqlc.narg('parent_code_filter')::text IS NULL OR parent_code = sqlc.narg('parent_code_filter'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'code' AND sqlc.arg('sort_order')::text = 'asc' THEN code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'code' AND sqlc.arg('sort_order')::text = 'desc' THEN code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'type' AND sqlc.arg('sort_order')::text = 'asc' THEN type END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'type' AND sqlc.arg('sort_order')::text = 'desc' THEN type END DESC,
    type ASC,
    name ASC
LIMIT $1 OFFSET $2;

-- name: CountRegions :one
SELECT COUNT(*)
FROM region
WHERE (COALESCE(cardinality(sqlc.arg('type_filters')::text[]), 0) = 0 OR type = ANY(sqlc.arg('type_filters')::text[]))
  AND (sqlc.narg('parent_code_filter')::text IS NULL OR parent_code = sqlc.narg('parent_code_filter'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
  );

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
WHERE (
    sqlc.narg('search')::text IS NULL
    OR title ILIKE '%' || sqlc.narg('search')::text || '%'
)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'title' AND sqlc.arg('sort_order')::text = 'asc' THEN title END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'title' AND sqlc.arg('sort_order')::text = 'desc' THEN title END DESC,
    title ASC
LIMIT $1 OFFSET $2;

-- name: CountProgramTitles :one
SELECT COUNT(*)
FROM program_title
WHERE (
    sqlc.narg('search')::text IS NULL
    OR title ILIKE '%' || sqlc.narg('search')::text || '%'
);

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
WHERE (COALESCE(cardinality(sqlc.arg('level_filters')::text[]), 0) = 0 OR level = ANY(sqlc.arg('level_filters')::text[]))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'level' AND sqlc.arg('sort_order')::text = 'asc' THEN level END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'level' AND sqlc.arg('sort_order')::text = 'desc' THEN level END DESC,
    level ASC,
    name ASC
LIMIT $1 OFFSET $2;

-- name: CountBappenasPartners :one
SELECT COUNT(*)
FROM bappenas_partner
WHERE (COALESCE(cardinality(sqlc.arg('level_filters')::text[]), 0) = 0 OR level = ANY(sqlc.arg('level_filters')::text[]))
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
  );

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
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'name' AND sqlc.arg('sort_order')::text = 'desc' THEN name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'year_start' AND sqlc.arg('sort_order')::text = 'asc' THEN year_start END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'year_start' AND sqlc.arg('sort_order')::text = 'desc' THEN year_start END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'year_end' AND sqlc.arg('sort_order')::text = 'asc' THEN year_end END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'year_end' AND sqlc.arg('sort_order')::text = 'desc' THEN year_end END DESC,
    year_start DESC
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
WHERE (COALESCE(cardinality(sqlc.arg('period_id_filters')::uuid[]), 0) = 0 OR np.period_id = ANY(sqlc.arg('period_id_filters')::uuid[]))
  AND (
    sqlc.narg('search')::text IS NULL
    OR np.title ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'title' AND sqlc.arg('sort_order')::text = 'asc' THEN np.title END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'title' AND sqlc.arg('sort_order')::text = 'desc' THEN np.title END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'period' AND sqlc.arg('sort_order')::text = 'asc' THEN p.name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'period' AND sqlc.arg('sort_order')::text = 'desc' THEN p.name END DESC,
    np.title ASC
LIMIT $1 OFFSET $2;

-- name: CountNationalPriorities :one
SELECT COUNT(*)
FROM national_priority
WHERE (COALESCE(cardinality(sqlc.arg('period_id_filters')::uuid[]), 0) = 0 OR period_id = ANY(sqlc.arg('period_id_filters')::uuid[]))
  AND (
    sqlc.narg('search')::text IS NULL
    OR title ILIKE '%' || sqlc.narg('search')::text || '%'
  );

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
