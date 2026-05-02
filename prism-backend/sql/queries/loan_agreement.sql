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
AND (
    COALESCE(cardinality(sqlc.arg('risk_codes')::text[]), 0) = 0
    OR (
        'EXTENDED_LOAN' = ANY(sqlc.arg('risk_codes')::text[])
        AND la.closing_date <> la.original_closing_date
    )
    OR (
        'CLOSING_RISK' = ANY(sqlc.arg('risk_codes')::text[])
        AND la.effective_date <= CURRENT_DATE
        AND la.closing_date <= CURRENT_DATE + INTERVAL '12 months'
        AND COALESCE((
            SELECT CASE
                WHEN SUM(md.planned_usd) > 0 THEN SUM(md.realized_usd) / SUM(md.planned_usd) * 100
                ELSE 0
            END
            FROM monitoring_disbursement md
            WHERE md.loan_agreement_id = la.id
        ), 0) < 80
    )
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
)
AND (
    COALESCE(cardinality(sqlc.arg('risk_codes')::text[]), 0) = 0
    OR (
        'EXTENDED_LOAN' = ANY(sqlc.arg('risk_codes')::text[])
        AND la.closing_date <> la.original_closing_date
    )
    OR (
        'CLOSING_RISK' = ANY(sqlc.arg('risk_codes')::text[])
        AND la.effective_date <= CURRENT_DATE
        AND la.closing_date <= CURRENT_DATE + INTERVAL '12 months'
        AND COALESCE((
            SELECT CASE
                WHEN SUM(md.planned_usd) > 0 THEN SUM(md.realized_usd) / SUM(md.planned_usd) * 100
                ELSE 0
            END
            FROM monitoring_disbursement md
            WHERE md.loan_agreement_id = la.id
        ), 0) < 80
    )
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
    la.id AS existing_loan_agreement_id,
    la.loan_code AS existing_loan_code
FROM dk_project dp
JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
LEFT JOIN dk_project_gb_project dkgb ON dkgb.dk_project_id = dp.id
LEFT JOIN gb_project gp ON gp.id = dkgb.gb_project_id
LEFT JOIN loan_agreement la ON la.dk_project_id = dp.id
GROUP BY
    dp.id,
    dk.id,
    dk.letter_number,
    dk.subject,
    dp.project_name,
    la.id,
    la.loan_code
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
