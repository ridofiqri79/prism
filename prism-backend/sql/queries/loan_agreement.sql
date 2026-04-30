-- ===== LOAN AGREEMENT =====

-- name: ListLoanAgreements :many
SELECT
    la.id,
    la.dk_project_id,
    la.lender_id,
    la.loan_code,
    la.agreement_date,
    la.effective_date,
    la.original_closing_date,
    la.closing_date,
    la.currency,
    la.amount_original,
    la.amount_usd,
    la.created_at,
    la.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE (
    sqlc.narg('search')::text IS NULL
    OR la.loan_code ILIKE '%' || sqlc.narg('search')::text || '%'
    OR l.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(l.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
AND (
    sqlc.narg('is_extended')::boolean IS NULL
    OR (la.closing_date <> la.original_closing_date) = sqlc.narg('is_extended')::boolean
)
AND (
    sqlc.narg('closing_date_before')::date IS NULL
    OR la.closing_date <= sqlc.narg('closing_date_before')::date
)
ORDER BY la.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountLoanAgreements :one
SELECT COUNT(*)
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE (
    sqlc.narg('search')::text IS NULL
    OR la.loan_code ILIKE '%' || sqlc.narg('search')::text || '%'
    OR l.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(l.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
AND (
    sqlc.narg('is_extended')::boolean IS NULL
    OR (la.closing_date <> la.original_closing_date) = sqlc.narg('is_extended')::boolean
)
AND (
    sqlc.narg('closing_date_before')::date IS NULL
    OR la.closing_date <= sqlc.narg('closing_date_before')::date
);

-- name: GetLoanAgreement :one
SELECT
    la.id,
    la.dk_project_id,
    la.lender_id,
    la.loan_code,
    la.agreement_date,
    la.effective_date,
    la.original_closing_date,
    la.closing_date,
    la.currency,
    la.amount_original,
    la.amount_usd,
    la.created_at,
    la.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE la.id = $1;

-- name: GetLoanAgreementByDKProject :one
SELECT *
FROM loan_agreement
WHERE dk_project_id = $1;

-- name: CreateLoanAgreement :one
INSERT INTO loan_agreement (
    dk_project_id,
    lender_id,
    loan_code,
    agreement_date,
    effective_date,
    original_closing_date,
    closing_date,
    currency,
    amount_original,
    amount_usd
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateLoanAgreement :one
UPDATE loan_agreement
SET lender_id = $2,
    loan_code = $3,
    agreement_date = $4,
    effective_date = $5,
    original_closing_date = $6,
    closing_date = $7,
    currency = $8,
    amount_original = $9,
    amount_usd = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLoanAgreement :exec
DELETE FROM loan_agreement
WHERE id = $1;

-- name: GetAllowedLenderIDsForLA :many
SELECT DISTINCT lender_id
FROM dk_financing_detail
WHERE dk_project_id = $1
  AND lender_id IS NOT NULL;
