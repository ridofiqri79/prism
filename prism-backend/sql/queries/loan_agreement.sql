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
        COALESCE(primary_rel.dk_project_id, la.dk_project_id) AS dk_project_id,
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
    LEFT JOIN LATERAL (
        SELECT ladp.dk_project_id
        FROM loan_agreement_dk_project ladp
        WHERE ladp.loan_agreement_id = la.id
        ORDER BY ladp.dk_project_id
        LIMIT 1
    ) primary_rel ON true
    WHERE (
        sqlc.narg('search')::text IS NULL
        OR la.loan_code ILIKE '%' || sqlc.narg('search')::text || '%'
        OR l.name ILIKE '%' || sqlc.narg('search')::text || '%'
        OR COALESCE(l.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
        OR EXISTS (
            SELECT 1
            FROM loan_agreement_dk_project ladp
            JOIN dk_project dp ON dp.id = ladp.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE ladp.loan_agreement_id = la.id
              AND (
                  dp.project_name ILIKE '%' || sqlc.narg('search')::text || '%'
                  OR dk.subject ILIKE '%' || sqlc.narg('search')::text || '%'
                  OR COALESCE(dk.letter_number, '') ILIKE '%' || sqlc.narg('search')::text || '%'
              )
        )
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
            FROM loan_agreement_dk_project ladp
            JOIN dk_project_gb_project dpg ON dpg.dk_project_id = ladp.dk_project_id
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE ladp.loan_agreement_id = la.id
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
    OR EXISTS (
        SELECT 1
        FROM loan_agreement_dk_project ladp
        JOIN dk_project dp ON dp.id = ladp.dk_project_id
        JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
        WHERE ladp.loan_agreement_id = la.id
          AND (
              dp.project_name ILIKE '%' || sqlc.narg('search')::text || '%'
              OR dk.subject ILIKE '%' || sqlc.narg('search')::text || '%'
              OR COALESCE(dk.letter_number, '') ILIKE '%' || sqlc.narg('search')::text || '%'
          )
    )
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
        FROM loan_agreement_dk_project ladp
        JOIN dk_project_gb_project dpg ON dpg.dk_project_id = ladp.dk_project_id
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE ladp.loan_agreement_id = la.id
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
        COALESCE(primary_rel.dk_project_id, la.dk_project_id) AS dk_project_id,
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
    LEFT JOIN LATERAL (
        SELECT ladp.dk_project_id
        FROM loan_agreement_dk_project ladp
        WHERE ladp.loan_agreement_id = la.id
        ORDER BY ladp.dk_project_id
        LIMIT 1
    ) primary_rel ON true
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
SELECT
    la.id,
    la.loan_code,
    la.currency,
    la.amount_original,
    la.amount_usd,
    la.cumulative_disbursement,
    ladp.allocation_original,
    ladp.allocation_usd
FROM loan_agreement_dk_project ladp
JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
WHERE ladp.dk_project_id = $1
ORDER BY la.created_at DESC, la.loan_code ASC;

-- name: ListLoanAgreementDKProjectsByLoanAgreements :many
SELECT
    ladp.loan_agreement_id,
    ladp.dk_project_id,
    ladp.allocation_original,
    ladp.allocation_usd,
    dp.dk_id,
    dp.project_name,
    dp.objectives,
    dk.subject AS dk_subject,
    dk.date AS dk_date,
    dk.letter_number AS dk_letter_number,
    COALESCE(string_agg(DISTINCT gp.gb_code, ', ' ORDER BY gp.gb_code), '')::text AS gb_codes
FROM loan_agreement_dk_project ladp
JOIN dk_project dp ON dp.id = ladp.dk_project_id
JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
LEFT JOIN dk_project_gb_project dkgb ON dkgb.dk_project_id = dp.id
LEFT JOIN gb_project gp ON gp.id = dkgb.gb_project_id
WHERE ladp.loan_agreement_id = ANY(sqlc.arg('loan_agreement_ids')::uuid[])
GROUP BY
    ladp.loan_agreement_id,
    ladp.dk_project_id,
    ladp.allocation_original,
    ladp.allocation_usd,
    dp.dk_id,
    dp.project_name,
    dp.objectives,
    dk.subject,
    dk.date,
    dk.letter_number
ORDER BY dk.date DESC, COALESCE(dk.letter_number, '') ASC, dp.project_name ASC;

-- name: CountMissingDKProjectsForLA :one
SELECT COUNT(*)::bigint
FROM unnest(sqlc.arg('dk_project_ids')::uuid[]) AS requested(id)
LEFT JOIN dk_project dp ON dp.id = requested.id
WHERE dp.id IS NULL;

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
        SELECT COUNT(DISTINCT la.id)::bigint
        FROM loan_agreement_dk_project ladp
        JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
        WHERE ladp.dk_project_id = dp.id
    ) AS loan_agreement_count,
    COALESCE((
        SELECT string_agg(DISTINCT la.loan_code, ', ' ORDER BY la.loan_code)
        FROM loan_agreement_dk_project ladp
        JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
        WHERE ladp.dk_project_id = dp.id
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

-- name: GetAllowedLenderIDsForLAProjects :many
SELECT dfd.lender_id
FROM dk_financing_detail dfd
WHERE dfd.dk_project_id = ANY(sqlc.arg('dk_project_ids')::uuid[])
  AND dfd.lender_id IS NOT NULL
GROUP BY dfd.lender_id
HAVING COUNT(DISTINCT dfd.dk_project_id) = cardinality(sqlc.arg('dk_project_ids')::uuid[]);

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
SET dk_project_id = $2,
    lender_id = $3,
    loan_code = $4,
    agreement_date = $5,
    effective_date = $6,
    original_closing_date = $7,
    closing_date = $8,
    currency = $9,
    amount_original = $10,
    amount_usd = $11,
    cumulative_disbursement = $12,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AddLoanAgreementDKProject :exec
INSERT INTO loan_agreement_dk_project (
    loan_agreement_id,
    dk_project_id,
    allocation_original,
    allocation_usd
)
VALUES ($1, $2, $3, $4)
ON CONFLICT (loan_agreement_id, dk_project_id) DO UPDATE
SET allocation_original = EXCLUDED.allocation_original,
    allocation_usd = EXCLUDED.allocation_usd,
    updated_at = NOW();

-- name: DeleteLoanAgreementDKProjects :exec
DELETE FROM loan_agreement_dk_project
WHERE loan_agreement_id = $1;

-- name: DeleteLoanAgreement :exec
DELETE FROM loan_agreement
WHERE id = $1;

-- name: GetAllowedLenderIDsForLA :many
SELECT DISTINCT lender_id
FROM dk_financing_detail
WHERE dk_project_id = $1
  AND lender_id IS NOT NULL;
