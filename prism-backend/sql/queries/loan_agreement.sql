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
    la.cumulative_disbursement,
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
    OR (la.original_closing_date IS NOT NULL AND la.closing_date <> la.original_closing_date) = sqlc.narg('is_extended')::boolean
)
AND (
    sqlc.narg('closing_date_before')::date IS NULL
    OR la.closing_date <= sqlc.narg('closing_date_before')::date
)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'loan_code' AND sqlc.arg('sort_order')::text = 'asc' THEN la.loan_code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'loan_code' AND sqlc.arg('sort_order')::text = 'desc' THEN la.loan_code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'lender' AND sqlc.arg('sort_order')::text = 'asc' THEN COALESCE(l.short_name, l.name) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'lender' AND sqlc.arg('sort_order')::text = 'desc' THEN COALESCE(l.short_name, l.name) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'effective_date' AND sqlc.arg('sort_order')::text = 'asc' THEN la.effective_date END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'effective_date' AND sqlc.arg('sort_order')::text = 'desc' THEN la.effective_date END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'closing_date' AND sqlc.arg('sort_order')::text = 'asc' THEN la.closing_date END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'closing_date' AND sqlc.arg('sort_order')::text = 'desc' THEN la.closing_date END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'currency' AND sqlc.arg('sort_order')::text = 'asc' THEN la.currency END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'currency' AND sqlc.arg('sort_order')::text = 'desc' THEN la.currency END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'amount_usd' AND sqlc.arg('sort_order')::text = 'asc' THEN la.amount_usd END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'amount_usd' AND sqlc.arg('sort_order')::text = 'desc' THEN la.amount_usd END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'cumulative_disbursement' AND sqlc.arg('sort_order')::text = 'asc' THEN la.cumulative_disbursement END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'cumulative_disbursement' AND sqlc.arg('sort_order')::text = 'desc' THEN la.cumulative_disbursement END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'asc' THEN (la.original_closing_date IS NOT NULL AND la.closing_date <> la.original_closing_date) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'desc' THEN (la.original_closing_date IS NOT NULL AND la.closing_date <> la.original_closing_date) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN la.created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN la.created_at END DESC,
    la.created_at DESC,
    la.id ASC
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
    OR (la.original_closing_date IS NOT NULL AND la.closing_date <> la.original_closing_date) = sqlc.narg('is_extended')::boolean
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
    la.cumulative_disbursement,
    la.created_at,
    la.updated_at,
    l.name AS lender_name,
    l.type AS lender_type,
    l.short_name AS lender_short_name
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
WHERE la.id = $1;

-- name: ListLoanAgreementsByDKProject :many
SELECT *
FROM loan_agreement
WHERE dk_project_id = $1
ORDER BY created_at DESC, loan_code ASC;

-- name: GetLoanAgreementByLoanCode :one
SELECT *
FROM loan_agreement
WHERE loan_code = $1;

-- name: ListLoanAgreementImportDKProjectReferences :many
SELECT
    dp.id,
    dk.id AS dk_id,
    COALESCE(dk.letter_number, '') AS letter_number,
    dk.subject,
    dp.project_name,
    COALESCE(string_agg(DISTINCT gp.gb_code, ', ' ORDER BY gp.gb_code), '')::text AS gb_codes,
    EXISTS (
        SELECT 1
        FROM dk_financing_detail dfd
        WHERE dfd.dk_project_id = dp.id
          AND dfd.lender_id IS NOT NULL
    ) AS has_financing_detail,
    (
        SELECT COUNT(*)::bigint
        FROM loan_agreement la
        WHERE la.dk_project_id = dp.id
    ) AS loan_agreement_count,
    COALESCE((
        SELECT string_agg(la.loan_code, ', ' ORDER BY la.loan_code)
        FROM loan_agreement la
        WHERE la.dk_project_id = dp.id
    ), '')::text AS existing_loan_codes
FROM dk_project dp
JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
LEFT JOIN dk_project_gb_project dkgb ON dkgb.dk_project_id = dp.id
LEFT JOIN gb_project gp ON gp.id = dkgb.gb_project_id
GROUP BY
    dp.id,
    dk.id,
    dk.letter_number,
    dk.subject,
    dp.project_name
ORDER BY dk.date DESC, COALESCE(dk.letter_number, '') ASC, dp.project_name ASC;

-- name: ListLoanAgreementAllowedLenderReferences :many
SELECT DISTINCT
    dfd.dk_project_id,
    dfd.lender_id,
    l.name AS lender_name,
    l.short_name AS lender_short_name,
    l.type AS lender_type,
    dfd.currency,
    dfd.amount_original,
    dfd.amount_usd
FROM dk_financing_detail dfd
JOIN lender l ON l.id = dfd.lender_id
WHERE dfd.lender_id IS NOT NULL
ORDER BY dfd.dk_project_id, l.name ASC, dfd.currency ASC;

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
    amount_usd,
    cumulative_disbursement
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
    cumulative_disbursement = $11,
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
