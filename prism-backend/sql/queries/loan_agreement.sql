-- ===== LOAN AGREEMENT =====

-- name: ListLoanAgreements :many
WITH latest_kurs AS (
    SELECT DISTINCT ON (UPPER(TRIM(c.code)))
        UPPER(TRIM(c.code)) AS currency_code,
        kt.kurs_tengah_bi,
        kt.cut_off_date
    FROM kurs_tengah kt
    JOIN currency c ON c.id = kt.currency_id
    ORDER BY UPPER(TRIM(c.code)), kt.cut_off_date DESC, kt.updated_at DESC, kt.id DESC
),
latest_cutoff AS (
    SELECT MAX(cut_off_date)::date AS cut_off_date
    FROM kurs_tengah
),
base AS (
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
        la.amount_usd AS stored_amount_usd,
        la.cumulative_disbursement,
        la.created_at,
        la.updated_at,
        l.name AS lender_name,
        l.type AS lender_type,
        l.short_name AS lender_short_name,
        CASE
            WHEN UPPER(TRIM(la.currency)) = 'USD' THEN 1::numeric
            ELSE lk.kurs_tengah_bi
        END AS kurs_tengah_bi,
        CASE
            WHEN UPPER(TRIM(la.currency)) = 'USD' THEN COALESCE(lc.cut_off_date, CURRENT_DATE)
            ELSE lk.cut_off_date
        END AS kurs_cut_off_date
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    LEFT JOIN latest_kurs lk ON lk.currency_code = UPPER(TRIM(la.currency))
    CROSS JOIN latest_cutoff lc
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
    AND (
        COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = la.dk_project_id
              AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
        )
    )
),
converted AS (
    SELECT
        base.*,
        CASE
            WHEN UPPER(TRIM(base.currency)) = 'USD' THEN base.amount_original
            WHEN base.kurs_tengah_bi IS NULL OR base.kurs_tengah_bi = 0 THEN base.stored_amount_usd
            ELSE base.amount_original / base.kurs_tengah_bi
        END AS amount_usd_calculated,
        CASE
            WHEN UPPER(TRIM(base.currency)) = 'USD' THEN base.cumulative_disbursement
            WHEN base.kurs_tengah_bi IS NULL OR base.kurs_tengah_bi = 0 THEN NULL::numeric
            ELSE base.cumulative_disbursement / base.kurs_tengah_bi
        END AS cumulative_disbursement_usd,
        CASE
            WHEN base.kurs_cut_off_date IS NULL OR base.closing_date <= base.effective_date THEN NULL::numeric
            ELSE ((base.kurs_cut_off_date - base.effective_date)::numeric / NULLIF((base.closing_date - base.effective_date)::numeric, 0)) * 100
        END AS estimated_time_ratio
    FROM base
),
metrics AS (
    SELECT
        converted.*,
        CASE
            WHEN converted.amount_usd_calculated IS NULL OR converted.amount_usd_calculated = 0 THEN NULL::numeric
            WHEN converted.cumulative_disbursement_usd IS NULL THEN NULL::numeric
            ELSE GREATEST(0::numeric, LEAST(100::numeric, (converted.cumulative_disbursement_usd / converted.amount_usd_calculated) * 100))
        END AS disbursement_ratio
    FROM converted
),
final AS (
    SELECT
        metrics.*,
        CASE
            WHEN metrics.disbursement_ratio IS NULL OR metrics.estimated_time_ratio IS NULL OR metrics.estimated_time_ratio = 0 THEN NULL::numeric
            ELSE metrics.disbursement_ratio / metrics.estimated_time_ratio
        END AS performance_value,
        CASE
            WHEN metrics.disbursement_ratio IS NULL OR metrics.estimated_time_ratio IS NULL THEN NULL::text
            WHEN metrics.disbursement_ratio <> 0 AND metrics.estimated_time_ratio = 0 THEN NULL::text
            WHEN metrics.disbursement_ratio <> 0 THEN
                CASE
                    WHEN (metrics.disbursement_ratio / metrics.estimated_time_ratio) <= 0.3 THEN 'At-Risk'
                    WHEN (metrics.disbursement_ratio / metrics.estimated_time_ratio) < 1 THEN 'Behind Schedule'
                    ELSE 'On Schedule'
                END
            ELSE
                CASE
                    WHEN metrics.estimated_time_ratio > 71 THEN 'At-Risk'
                    ELSE 'Behind Schedule'
                END
        END AS performance_status
    FROM metrics
)
SELECT
    id,
    dk_project_id,
    lender_id,
    loan_code,
    agreement_date,
    effective_date,
    original_closing_date,
    closing_date,
    currency,
    amount_original,
    amount_usd_calculated::numeric AS amount_usd,
    cumulative_disbursement,
    cumulative_disbursement_usd::numeric AS cumulative_disbursement_usd,
    disbursement_ratio::numeric AS disbursement_ratio,
    estimated_time_ratio::numeric AS estimated_time_ratio,
    performance_value::numeric AS performance_value,
    COALESCE(performance_status::text, '') AS performance_status,
    kurs_tengah_bi::numeric AS kurs_tengah_bi,
    kurs_cut_off_date::date AS kurs_cut_off_date,
    created_at,
    updated_at,
    lender_name,
    lender_type,
    lender_short_name
FROM final
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'loan_code' AND sqlc.arg('sort_order')::text = 'asc' THEN loan_code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'loan_code' AND sqlc.arg('sort_order')::text = 'desc' THEN loan_code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'lender' AND sqlc.arg('sort_order')::text = 'asc' THEN COALESCE(lender_short_name, lender_name) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'lender' AND sqlc.arg('sort_order')::text = 'desc' THEN COALESCE(lender_short_name, lender_name) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'effective_date' AND sqlc.arg('sort_order')::text = 'asc' THEN effective_date END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'effective_date' AND sqlc.arg('sort_order')::text = 'desc' THEN effective_date END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'closing_date' AND sqlc.arg('sort_order')::text = 'asc' THEN closing_date END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'closing_date' AND sqlc.arg('sort_order')::text = 'desc' THEN closing_date END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'currency' AND sqlc.arg('sort_order')::text = 'asc' THEN currency END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'currency' AND sqlc.arg('sort_order')::text = 'desc' THEN currency END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'amount_usd' AND sqlc.arg('sort_order')::text = 'asc' THEN amount_usd_calculated END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'amount_usd' AND sqlc.arg('sort_order')::text = 'desc' THEN amount_usd_calculated END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'cumulative_disbursement_usd' AND sqlc.arg('sort_order')::text = 'asc' THEN cumulative_disbursement_usd END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'cumulative_disbursement_usd' AND sqlc.arg('sort_order')::text = 'desc' THEN cumulative_disbursement_usd END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'disbursement_ratio' AND sqlc.arg('sort_order')::text = 'asc' THEN disbursement_ratio END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'disbursement_ratio' AND sqlc.arg('sort_order')::text = 'desc' THEN disbursement_ratio END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'estimated_time_ratio' AND sqlc.arg('sort_order')::text = 'asc' THEN estimated_time_ratio END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'estimated_time_ratio' AND sqlc.arg('sort_order')::text = 'desc' THEN estimated_time_ratio END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'performance_value' AND sqlc.arg('sort_order')::text = 'asc' THEN performance_value END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'performance_value' AND sqlc.arg('sort_order')::text = 'desc' THEN performance_value END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'performance_status' AND sqlc.arg('sort_order')::text = 'asc' THEN performance_status END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'performance_status' AND sqlc.arg('sort_order')::text = 'desc' THEN performance_status END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'asc' THEN (original_closing_date IS NOT NULL AND closing_date <> original_closing_date) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'desc' THEN (original_closing_date IS NOT NULL AND closing_date <> original_closing_date) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN created_at END DESC,
    created_at DESC,
    id ASC
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
)
AND (
    COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE dpg.dk_project_id = la.dk_project_id
          AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
    )
);

-- name: GetLoanAgreement :one
WITH latest_kurs AS (
    SELECT DISTINCT ON (UPPER(TRIM(c.code)))
        UPPER(TRIM(c.code)) AS currency_code,
        kt.kurs_tengah_bi,
        kt.cut_off_date
    FROM kurs_tengah kt
    JOIN currency c ON c.id = kt.currency_id
    ORDER BY UPPER(TRIM(c.code)), kt.cut_off_date DESC, kt.updated_at DESC, kt.id DESC
),
latest_cutoff AS (
    SELECT MAX(cut_off_date)::date AS cut_off_date
    FROM kurs_tengah
),
base AS (
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
        la.amount_usd AS stored_amount_usd,
        la.cumulative_disbursement,
        la.created_at,
        la.updated_at,
        l.name AS lender_name,
        l.type AS lender_type,
        l.short_name AS lender_short_name,
        CASE
            WHEN UPPER(TRIM(la.currency)) = 'USD' THEN 1::numeric
            ELSE lk.kurs_tengah_bi
        END AS kurs_tengah_bi,
        CASE
            WHEN UPPER(TRIM(la.currency)) = 'USD' THEN COALESCE(lc.cut_off_date, CURRENT_DATE)
            ELSE lk.cut_off_date
        END AS kurs_cut_off_date
    FROM loan_agreement la
    JOIN lender l ON l.id = la.lender_id
    LEFT JOIN latest_kurs lk ON lk.currency_code = UPPER(TRIM(la.currency))
    CROSS JOIN latest_cutoff lc
    WHERE la.id = $1
),
converted AS (
    SELECT
        base.*,
        CASE
            WHEN UPPER(TRIM(base.currency)) = 'USD' THEN base.amount_original
            WHEN base.kurs_tengah_bi IS NULL OR base.kurs_tengah_bi = 0 THEN base.stored_amount_usd
            ELSE base.amount_original / base.kurs_tengah_bi
        END AS amount_usd_calculated,
        CASE
            WHEN UPPER(TRIM(base.currency)) = 'USD' THEN base.cumulative_disbursement
            WHEN base.kurs_tengah_bi IS NULL OR base.kurs_tengah_bi = 0 THEN NULL::numeric
            ELSE base.cumulative_disbursement / base.kurs_tengah_bi
        END AS cumulative_disbursement_usd,
        CASE
            WHEN base.kurs_cut_off_date IS NULL OR base.closing_date <= base.effective_date THEN NULL::numeric
            ELSE ((base.kurs_cut_off_date - base.effective_date)::numeric / NULLIF((base.closing_date - base.effective_date)::numeric, 0)) * 100
        END AS estimated_time_ratio
    FROM base
),
metrics AS (
    SELECT
        converted.*,
        CASE
            WHEN converted.amount_usd_calculated IS NULL OR converted.amount_usd_calculated = 0 THEN NULL::numeric
            WHEN converted.cumulative_disbursement_usd IS NULL THEN NULL::numeric
            ELSE GREATEST(0::numeric, LEAST(100::numeric, (converted.cumulative_disbursement_usd / converted.amount_usd_calculated) * 100))
        END AS disbursement_ratio
    FROM converted
),
final AS (
    SELECT
        metrics.*,
        CASE
            WHEN metrics.disbursement_ratio IS NULL OR metrics.estimated_time_ratio IS NULL OR metrics.estimated_time_ratio = 0 THEN NULL::numeric
            ELSE metrics.disbursement_ratio / metrics.estimated_time_ratio
        END AS performance_value,
        CASE
            WHEN metrics.disbursement_ratio IS NULL OR metrics.estimated_time_ratio IS NULL THEN NULL::text
            WHEN metrics.disbursement_ratio <> 0 AND metrics.estimated_time_ratio = 0 THEN NULL::text
            WHEN metrics.disbursement_ratio <> 0 THEN
                CASE
                    WHEN (metrics.disbursement_ratio / metrics.estimated_time_ratio) <= 0.3 THEN 'At-Risk'
                    WHEN (metrics.disbursement_ratio / metrics.estimated_time_ratio) < 1 THEN 'Behind Schedule'
                    ELSE 'On Schedule'
                END
            ELSE
                CASE
                    WHEN metrics.estimated_time_ratio > 71 THEN 'At-Risk'
                    ELSE 'Behind Schedule'
                END
        END AS performance_status
    FROM metrics
)
SELECT
    id,
    dk_project_id,
    lender_id,
    loan_code,
    agreement_date,
    effective_date,
    original_closing_date,
    closing_date,
    currency,
    amount_original,
    amount_usd_calculated::numeric AS amount_usd,
    cumulative_disbursement,
    cumulative_disbursement_usd::numeric AS cumulative_disbursement_usd,
    disbursement_ratio::numeric AS disbursement_ratio,
    estimated_time_ratio::numeric AS estimated_time_ratio,
    performance_value::numeric AS performance_value,
    COALESCE(performance_status::text, '') AS performance_status,
    kurs_tengah_bi::numeric AS kurs_tengah_bi,
    kurs_cut_off_date::date AS kurs_cut_off_date,
    created_at,
    updated_at,
    lender_name,
    lender_type,
    lender_short_name
FROM final;

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
