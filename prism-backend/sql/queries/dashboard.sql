-- ===== DASHBOARD FOUNDATION =====

-- name: GetDashboardSummary :one
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM bb_project_institution bpi
              WHERE bpi.bb_project_id = bp.id
                AND bpi.institution_id = sqlc.narg('institution_id')::uuid
          )
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM lender_indication li
              WHERE li.bb_project_id = bp.id
                AND li.lender_id = sqlc.narg('lender_id')::uuid
          )
      )
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_institution gpi
              WHERE gpi.gb_project_id = gp.id
                AND gpi.institution_id = sqlc.narg('institution_id')::uuid
          )
          OR EXISTS (
              SELECT 1
              FROM gb_funding_source gfs
              WHERE gfs.gb_project_id = gp.id
                AND gfs.institution_id = sqlc.narg('institution_id')::uuid
          )
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_funding_source gfs
              WHERE gfs.gb_project_id = gp.id
                AND gfs.lender_id = sqlc.narg('lender_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
bb_costs AS (
    SELECT COALESCE(SUM(bpc.amount_usd), 0)::numeric AS bb_pipeline_usd
    FROM selected_bb sb
    JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id
    WHERE bpc.funding_type = 'Foreign'
),
gb_costs AS (
    SELECT
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS gb_pipeline_usd,
        COALESCE(SUM(gfs.local_usd), 0)::numeric AS gb_local_usd
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
),
dk_costs AS (
    SELECT
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS dk_financing_usd,
        COALESCE(SUM(dfd.counterpart_usd), 0)::numeric AS dk_counterpart_usd
    FROM dk_financing_detail dfd
    JOIN dk_project dp ON dp.id = dfd.dk_project_id
    WHERE (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
      AND (
          sqlc.narg('publish_year')::int IS NULL
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project gp ON gp.id = dpg.gb_project_id
              JOIN green_book gb ON gb.id = gp.green_book_id
              WHERE dpg.dk_project_id = dp.id
                AND gb.publish_year = sqlc.narg('publish_year')::int
          )
      )
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM dk_project_gb_project dpg
              JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE dpg.dk_project_id = dp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
la_costs AS (
    SELECT
        COALESCE(SUM(la.amount_usd), 0)::numeric AS la_commitment_usd,
        COUNT(*)::bigint AS total_loan_agreements
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
)
SELECT
    (SELECT COUNT(*) FROM selected_bb)::bigint AS total_bb_projects,
    (SELECT COUNT(*) FROM selected_gb)::bigint AS total_gb_projects,
    la_costs.total_loan_agreements,
    bb_costs.bb_pipeline_usd,
    gb_costs.gb_pipeline_usd,
    gb_costs.gb_local_usd,
    dk_costs.dk_financing_usd,
    dk_costs.dk_counterpart_usd,
    la_costs.la_commitment_usd
FROM bb_costs, gb_costs, dk_costs, la_costs;

-- name: GetDashboardStageCounts :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
selected_la AS (
    SELECT la.id
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
)
SELECT 'BB'::text AS stage, COUNT(*)::bigint AS project_count FROM selected_bb
UNION ALL
SELECT 'GB'::text AS stage, COUNT(*)::bigint AS project_count FROM selected_gb
UNION ALL
SELECT 'DK'::text AS stage, COUNT(*)::bigint AS project_count FROM selected_dk
UNION ALL
SELECT 'LA'::text AS stage, COUNT(*)::bigint AS project_count FROM selected_la;

-- name: GetDashboardStageAmounts :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
selected_la AS (
    SELECT la.id, la.amount_usd
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
)
SELECT 'BB'::text AS stage, COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
FROM selected_bb sb
LEFT JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id AND bpc.funding_type = 'Foreign'
UNION ALL
SELECT 'GB'::text AS stage, COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS amount_usd
FROM selected_gb sg
LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
UNION ALL
SELECT 'DK'::text AS stage, COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS amount_usd
FROM selected_dk sd
LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
UNION ALL
SELECT 'LA'::text AS stage, COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM selected_la;

-- name: ListDashboardFilterOptions :many
SELECT 'period'::text AS option_type, p.id::text AS value, p.name::text AS label
FROM period p
UNION ALL
SELECT 'publish_year'::text AS option_type, gb.publish_year::text AS value, gb.publish_year::text AS label
FROM green_book gb
GROUP BY gb.publish_year
UNION ALL
SELECT 'green_book'::text AS option_type, gb.id::text AS value, ('GB ' || gb.publish_year::text || ' Revisi ke-' || gb.revision_number::text)::text AS label
FROM green_book gb
UNION ALL
SELECT 'lender'::text AS option_type, l.id::text AS value, COALESCE(l.short_name, l.name)::text AS label
FROM lender l
UNION ALL
SELECT 'currency'::text AS option_type, c.code::text AS value, c.code::text AS label
FROM currency c
WHERE c.is_active = true
UNION ALL
SELECT 'institution'::text AS option_type, i.id::text AS value, COALESCE(i.short_name, i.name)::text AS label
FROM institution i
UNION ALL
SELECT 'institution_role'::text AS option_type, r.role AS value, r.role AS label
FROM (VALUES ('Executing Agency'), ('Implementing Agency')) AS r(role)
ORDER BY option_type ASC, label ASC;

-- name: GetDashboardExecutiveFunnel :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN selected_gb sg ON sg.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = dp.id
    )
),
selected_la AS (
    SELECT la.id, la.amount_usd
    FROM loan_agreement la
    WHERE EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN selected_gb sg ON sg.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = la.dk_project_id
    )
)
SELECT 'BB'::text AS stage, COUNT(DISTINCT sb.id)::bigint AS project_count, COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
FROM selected_bb sb
LEFT JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id AND bpc.funding_type = 'Foreign'
UNION ALL
SELECT 'GB'::text AS stage, COUNT(DISTINCT sg.id)::bigint AS project_count, COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS amount_usd
FROM selected_gb sg
LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
UNION ALL
SELECT 'DK'::text AS stage, COUNT(DISTINCT sd.id)::bigint AS project_count, COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS amount_usd
FROM selected_dk sd
LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
UNION ALL
SELECT 'LA'::text AS stage, COUNT(*)::bigint AS project_count, COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM selected_la;

-- name: GetDashboardExecutiveFunnelByProgramTitle :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.program_title_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id, program_title_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.program_title_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT
        rg.id,
        COALESCE(
            rg.program_title_id,
            (
                SELECT bp.program_title_id
                FROM gb_project_bb_project gbp
                JOIN bb_project bp ON bp.id = gbp.bb_project_id
                WHERE gbp.gb_project_id = rg.id
                  AND bp.program_title_id IS NOT NULL
                ORDER BY bp.created_at DESC
                LIMIT 1
            )
        ) AS program_title_id
    FROM ranked_gb rg
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
selected_dk AS (
    SELECT DISTINCT
        dp.id,
        COALESCE(
            dp.program_title_id,
            (
                SELECT sg.program_title_id
                FROM dk_project_gb_project dpg
                JOIN selected_gb sg ON sg.id = dpg.gb_project_id
                WHERE dpg.dk_project_id = dp.id
                  AND sg.program_title_id IS NOT NULL
                ORDER BY sg.id
                LIMIT 1
            )
        ) AS program_title_id
    FROM dk_project dp
    WHERE EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN selected_gb sg ON sg.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = dp.id
    )
),
selected_la AS (
    SELECT
        la.id,
        la.amount_usd,
        COALESCE(
            dp.program_title_id,
            (
                SELECT sg.program_title_id
                FROM dk_project_gb_project dpg
                JOIN selected_gb sg ON sg.id = dpg.gb_project_id
                WHERE dpg.dk_project_id = dp.id
                  AND sg.program_title_id IS NOT NULL
                ORDER BY sg.id
                LIMIT 1
            )
        ) AS program_title_id
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN selected_gb sg ON sg.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = la.dk_project_id
    )
),
stage_rows AS (
    SELECT
        'BB'::text AS stage,
        COALESCE(pt.id::text, '')::text AS program_title_id,
        COALESCE(pt.title, 'Tanpa Program Title')::text AS program_title,
        COUNT(DISTINCT sb.id)::bigint AS project_count,
        COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    LEFT JOIN program_title pt ON pt.id = sb.program_title_id
    LEFT JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id AND bpc.funding_type = 'Foreign'
    GROUP BY COALESCE(pt.id::text, ''), COALESCE(pt.title, 'Tanpa Program Title')

    UNION ALL
    SELECT
        'GB'::text AS stage,
        COALESCE(pt.id::text, '')::text AS program_title_id,
        COALESCE(pt.title, 'Tanpa Program Title')::text AS program_title,
        COUNT(DISTINCT sg.id)::bigint AS project_count,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS amount_usd
    FROM selected_gb sg
    LEFT JOIN program_title pt ON pt.id = sg.program_title_id
    LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    GROUP BY COALESCE(pt.id::text, ''), COALESCE(pt.title, 'Tanpa Program Title')

    UNION ALL
    SELECT
        'DK'::text AS stage,
        COALESCE(pt.id::text, '')::text AS program_title_id,
        COALESCE(pt.title, 'Tanpa Program Title')::text AS program_title,
        COUNT(DISTINCT sd.id)::bigint AS project_count,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    LEFT JOIN program_title pt ON pt.id = sd.program_title_id
    LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    GROUP BY COALESCE(pt.id::text, ''), COALESCE(pt.title, 'Tanpa Program Title')

    UNION ALL
    SELECT
        'LA'::text AS stage,
        COALESCE(pt.id::text, '')::text AS program_title_id,
        COALESCE(pt.title, 'Tanpa Program Title')::text AS program_title,
        COUNT(DISTINCT sl.id)::bigint AS project_count,
        COALESCE(SUM(sl.amount_usd), 0)::numeric AS amount_usd
    FROM selected_la sl
    LEFT JOIN program_title pt ON pt.id = sl.program_title_id
    GROUP BY COALESCE(pt.id::text, ''), COALESCE(pt.title, 'Tanpa Program Title')
)
SELECT
    stage,
    program_title_id,
    program_title,
    project_count,
    amount_usd
FROM stage_rows
ORDER BY
    program_title ASC,
    CASE stage WHEN 'BB' THEN 1 WHEN 'GB' THEN 2 WHEN 'DK' THEN 3 ELSE 4 END ASC;

-- name: GetDashboardExecutiveTopInstitutions :many
WITH RECURSIVE institution_ancestors AS (
    SELECT i.id AS institution_id, i.id AS ancestor_id, i.parent_id, i.name, i.short_name, i.level
    FROM institution i
    UNION ALL
    SELECT ia.institution_id, parent.id AS ancestor_id, parent.parent_id, parent.name, parent.short_name, parent.level
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
institution_roots AS (
    SELECT DISTINCT ON (institution_id)
        institution_id,
        ancestor_id AS root_id,
        COALESCE(short_name, name)::text AS root_label
    FROM institution_ancestors
    WHERE parent_id IS NULL OR level = 'Kementerian/Badan/Lembaga'
    ORDER BY institution_id, CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END
)
SELECT
    ir.root_id AS id,
    ir.root_label AS label,
    COUNT(DISTINCT dp.id)::bigint AS item_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd
FROM dk_project dp
JOIN institution_roots ir ON ir.institution_id = dp.institution_id
LEFT JOIN loan_agreement la ON la.dk_project_id = dp.id
WHERE (
    sqlc.narg('publish_year')::int IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project gp ON gp.id = dpg.gb_project_id
        JOIN green_book gb ON gb.id = gp.green_book_id
        WHERE dpg.dk_project_id = dp.id
          AND gb.publish_year = sqlc.narg('publish_year')::int
    )
)
AND (
    sqlc.narg('period_id')::uuid IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE dpg.dk_project_id = dp.id
          AND bb.period_id = sqlc.narg('period_id')::uuid
    )
)
GROUP BY ir.root_id, ir.root_label
ORDER BY amount_usd DESC, item_count DESC, label ASC
LIMIT 10;

-- name: GetDashboardExecutiveTopLenders :many
SELECT
    l.id,
    COALESCE(l.short_name, l.name)::text AS label,
    COUNT(DISTINCT la.id)::bigint AS item_count,
    COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd
FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id
JOIN dk_project dp ON dp.id = la.dk_project_id
WHERE (
    sqlc.narg('publish_year')::int IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project gp ON gp.id = dpg.gb_project_id
        JOIN green_book gb ON gb.id = gp.green_book_id
        WHERE dpg.dk_project_id = dp.id
          AND gb.publish_year = sqlc.narg('publish_year')::int
    )
)
AND (
    sqlc.narg('period_id')::uuid IS NULL
    OR EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN blue_book bb ON bb.id = bp.blue_book_id
        WHERE dpg.dk_project_id = dp.id
          AND bb.period_id = sqlc.narg('period_id')::uuid
    )
)
GROUP BY l.id, COALESCE(l.short_name, l.name)
ORDER BY amount_usd DESC, item_count DESC, label ASC
LIMIT 10;

-- name: ListDashboardExecutiveRiskItems :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.gb_code,
        gp.project_name,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id, gb_code, project_name
    FROM ranked_gb
    WHERE sqlc.arg('include_history')::boolean OR rn = 1
),
la_base AS (
    SELECT
        la.id,
        la.loan_code,
        la.dk_project_id,
        la.effective_date,
        la.closing_date,
        la.amount_usd,
        dp.project_name,
        latest_bp.id AS journey_bb_project_id
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    LEFT JOIN LATERAL (
        SELECT latest.id
        FROM dk_project_gb_project dpg
        JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
        JOIN bb_project bp ON bp.id = gbp.bb_project_id
        JOIN bb_project latest ON latest.project_identity_id = bp.project_identity_id
        JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
        WHERE dpg.dk_project_id = dp.id
          AND latest.status = 'active'
        ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC, latest.created_at DESC
        LIMIT 1
    ) latest_bp ON TRUE
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
    AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
risks AS (
    SELECT
        'LA_CLOSING_12_MONTHS'::text AS risk_type,
        'high'::text AS severity,
        la_base.id AS reference_id,
        'loan_agreement'::text AS reference_type,
        la_base.journey_bb_project_id,
        la_base.loan_code::text AS code,
        la_base.project_name::text AS title,
        ('Loan Agreement ' || la_base.loan_code || ' closing pada ' || la_base.closing_date::text)::text AS description,
        la_base.amount_usd::numeric AS amount_usd,
        (la_base.closing_date - CURRENT_DATE)::int AS days_until_closing,
        90::numeric AS score
    FROM la_base
    WHERE la_base.closing_date BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '12 months')::date
    UNION ALL
    SELECT
        'HIGH_ELAPSED_LA'::text AS risk_type,
        'medium'::text AS severity,
        la_base.id AS reference_id,
        'loan_agreement'::text AS reference_type,
        la_base.journey_bb_project_id,
        la_base.loan_code::text AS code,
        la_base.project_name::text AS title,
        ('Loan Agreement ' || la_base.loan_code || ' sudah berjalan lama dan perlu perhatian')::text AS description,
        la_base.amount_usd::numeric AS amount_usd,
        0::int AS days_until_closing,
        75::numeric AS score
    FROM la_base
    WHERE la_base.effective_date < CURRENT_DATE
      AND la_base.closing_date > la_base.effective_date
      AND ((CURRENT_DATE - la_base.effective_date)::numeric / NULLIF((la_base.closing_date - la_base.effective_date)::numeric, 0)) >= 0.7
    UNION ALL
    SELECT
        'GB_WITHOUT_DK'::text AS risk_type,
        'medium'::text AS severity,
        sg.id AS reference_id,
        'gb_project'::text AS reference_type,
        latest_bp.id AS journey_bb_project_id,
        sg.gb_code::text AS code,
        sg.project_name::text AS title,
        ('Green Book project ' || sg.gb_code || ' belum masuk Daftar Kegiatan')::text AS description,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd), 0)::numeric AS amount_usd,
        0::int AS days_until_closing,
        60::numeric AS score
    FROM selected_gb sg
    LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    LEFT JOIN LATERAL (
        SELECT sb.id
        FROM gb_project_bb_project gbp
        JOIN selected_bb sb ON sb.id = gbp.bb_project_id
        WHERE gbp.gb_project_id = sg.id
        ORDER BY sb.id
        LIMIT 1
    ) latest_bp ON TRUE
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = sg.id)
    GROUP BY sg.id, sg.gb_code, sg.project_name, latest_bp.id
    UNION ALL
    SELECT
        'DK_WITHOUT_LA'::text AS risk_type,
        'medium'::text AS severity,
        dp.id AS reference_id,
        'dk_project'::text AS reference_type,
        latest_bp.id AS journey_bb_project_id,
        COALESCE(dk.letter_number, dp.id::text)::text AS code,
        dp.project_name::text AS title,
        ('Daftar Kegiatan project belum memiliki Loan Agreement')::text AS description,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd), 0)::numeric AS amount_usd,
        0::int AS days_until_closing,
        65::numeric AS score
    FROM dk_project dp
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
    JOIN dk_project_gb_project dpg ON dpg.dk_project_id = dp.id
    JOIN selected_gb sg ON sg.id = dpg.gb_project_id
    LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = dp.id
    LEFT JOIN LATERAL (
        SELECT sb.id
        FROM gb_project_bb_project gbp
        JOIN selected_bb sb ON sb.id = gbp.bb_project_id
        WHERE gbp.gb_project_id = sg.id
        ORDER BY sb.id
        LIMIT 1
    ) latest_bp ON TRUE
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id)
    GROUP BY dp.id, dk.letter_number, dp.project_name, latest_bp.id
)
SELECT
    risk_type,
    severity,
    reference_id,
    reference_type,
    journey_bb_project_id,
    code,
    title,
    description,
    amount_usd,
    days_until_closing,
    score
FROM risks
ORDER BY score DESC, amount_usd DESC, title ASC
LIMIT 25;

-- name: GetDashboardPipelineBottleneckStageSummary :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id
    FROM ranked_gb
    WHERE rn = 1
),
worklist AS (
    SELECT
        'BB_NO_LENDER'::text AS stage,
        sb.id AS project_id,
        bp.project_name::text AS project_name,
        COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - bp.created_at::date), 0)::int AS age_days,
        bp.created_at AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name,
        ARRAY[]::text[] AS lender_names
    FROM selected_bb sb
    JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND sqlc.narg('lender_id')::uuid IS NULL
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'INDICATION_NO_LOI'::text AS stage,
        sb.id AS project_id,
        bp.project_name::text AS project_name,
        COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(li.created_at)::date FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int AS age_days,
        COALESCE((SELECT MAX(li.created_at) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at) AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id JOIN lender l ON l.id = li.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[]) AS lender_names
    FROM selected_bb sb
    JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND li.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'LOI_NO_GB'::text AS stage,
        sb.id AS project_id,
        bp.project_name::text AS project_name,
        COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(lo.date) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int AS age_days,
        COALESCE((SELECT MAX(lo.date)::timestamptz FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at) AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id JOIN lender l ON l.id = lo.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[]) AS lender_names
    FROM selected_bb sb
    JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL
      AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND lo.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'GB_NO_DK'::text AS stage,
        sg.id AS project_id,
        gp.project_name::text AS project_name,
        COALESCE((SELECT SUM(gfs.loan_usd + gfs.grant_usd) FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - gp.created_at::date), 0)::int AS age_days,
        gp.created_at AS relevant_at,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = gp.id AND gpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = gp.id), ARRAY[]::text[]) AS lender_names
    FROM selected_gb sg
    JOIN gb_project gp ON gp.id = sg.id
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = gp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid) OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT
        'DK_NO_LA'::text AS stage,
        dp.id AS project_id,
        dp.project_name::text AS project_name,
        COALESCE((SELECT SUM(dfd.amount_usd + dfd.grant_usd) FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id), 0)::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - COALESCE(dk.date, dp.created_at::date)), 0)::int AS age_days,
        COALESCE(dk.date::timestamptz, dp.created_at) AS relevant_at,
        COALESCE(i.short_name, i.name, '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM dk_financing_detail dfd JOIN lender l ON l.id = dfd.lender_id WHERE dfd.dk_project_id = dp.id), ARRAY[]::text[]) AS lender_names
    FROM dk_project dp
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
    LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id)
      AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id AND dfd.lender_id = sqlc.narg('lender_id')::uuid))
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT
        'LA_NOT_EFFECTIVE'::text AS stage,
        la.id AS project_id,
        dp.project_name::text AS project_name,
        la.amount_usd::numeric AS amount_usd,
        GREATEST((CURRENT_DATE - la.agreement_date), 0)::int AS age_days,
        la.agreement_date::timestamptz AS relevant_at,
        COALESCE(i.short_name, i.name, '')::text AS institution_name,
        ARRAY[COALESCE(l.short_name, l.name)]::text[] AS lender_names
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    JOIN lender l ON l.id = la.lender_id
    LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date > CURRENT_DATE
      AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
filtered AS (
    SELECT *
    FROM worklist
    WHERE (sqlc.narg('stage')::text IS NULL OR stage = sqlc.narg('stage')::text)
      AND (sqlc.narg('min_age_days')::int IS NULL OR age_days >= sqlc.narg('min_age_days')::int)
      AND (
          sqlc.narg('search')::text IS NULL
          OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
          OR institution_name ILIKE '%' || sqlc.narg('search')::text || '%'
          OR EXISTS (SELECT 1 FROM unnest(lender_names) lender_name WHERE lender_name ILIKE '%' || sqlc.narg('search')::text || '%')
      )
)
SELECT
    stage,
    COUNT(*)::bigint AS project_count,
    COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd,
    COALESCE(AVG(age_days), 0)::numeric AS avg_age_days
FROM filtered
GROUP BY stage
ORDER BY CASE stage
    WHEN 'BB_NO_LENDER' THEN 1
    WHEN 'INDICATION_NO_LOI' THEN 2
    WHEN 'LOI_NO_GB' THEN 3
    WHEN 'GB_NO_DK' THEN 4
    WHEN 'DK_NO_LA' THEN 5
    WHEN 'LA_NOT_EFFECTIVE' THEN 6
    ELSE 99
END;

-- name: CountDashboardPipelineBottleneckItems :one
WITH ranked_bb AS (
    SELECT bp.id, bp.project_identity_id, ROW_NUMBER() OVER (PARTITION BY bp.project_identity_id ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (SELECT id, project_identity_id FROM ranked_bb WHERE rn = 1),
ranked_gb AS (
    SELECT gp.id, gp.gb_project_identity_id, ROW_NUMBER() OVER (PARTITION BY gp.gb_project_identity_id ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (sqlc.narg('period_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_bb_project gbp JOIN bb_project bp ON bp.id = gbp.bb_project_id JOIN blue_book bb ON bb.id = bp.blue_book_id WHERE gbp.gb_project_id = gp.id AND bb.period_id = sqlc.narg('period_id')::uuid))
),
selected_gb AS (SELECT id, gb_project_identity_id FROM ranked_gb WHERE rn = 1),
worklist AS (
    SELECT 'BB_NO_LENDER'::text AS stage, sb.id AS project_id, bp.project_name::text AS project_name, GREATEST((CURRENT_DATE - bp.created_at::date), 0)::int AS age_days, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name, ARRAY[]::text[] AS lender_names
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND sqlc.narg('lender_id')::uuid IS NULL AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'INDICATION_NO_LOI'::text, sb.id, bp.project_name::text, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(li.created_at)::date FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id JOIN lender l ON l.id = li.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND li.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'LOI_NO_GB'::text, sb.id, bp.project_name::text, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(lo.date) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id JOIN lender l ON l.id = lo.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND lo.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'GB_NO_DK'::text, sg.id, gp.project_name::text, GREATEST((CURRENT_DATE - gp.created_at::date), 0)::int, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ') FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = gp.id AND gpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = gp.id), ARRAY[]::text[])
    FROM selected_gb sg JOIN gb_project gp ON gp.id = sg.id
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = gp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid) OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'DK_NO_LA'::text, dp.id, dp.project_name::text, GREATEST((CURRENT_DATE - COALESCE(dk.date, dp.created_at::date)), 0)::int, COALESCE(i.short_name, i.name, '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name)) FROM dk_financing_detail dfd JOIN lender l ON l.id = dfd.lender_id WHERE dfd.dk_project_id = dp.id), ARRAY[]::text[])
    FROM dk_project dp JOIN daftar_kegiatan dk ON dk.id = dp.dk_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id) AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id AND dfd.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT 'LA_NOT_EFFECTIVE'::text, la.id, dp.project_name::text, GREATEST((CURRENT_DATE - la.agreement_date), 0)::int, COALESCE(i.short_name, i.name, '')::text, ARRAY[COALESCE(l.short_name, l.name)]::text[]
    FROM loan_agreement la JOIN dk_project dp ON dp.id = la.dk_project_id JOIN lender l ON l.id = la.lender_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date > CURRENT_DATE AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
filtered AS (
    SELECT * FROM worklist
    WHERE (sqlc.narg('stage')::text IS NULL OR stage = sqlc.narg('stage')::text)
      AND (sqlc.narg('min_age_days')::int IS NULL OR age_days >= sqlc.narg('min_age_days')::int)
      AND (sqlc.narg('search')::text IS NULL OR project_name ILIKE '%' || sqlc.narg('search')::text || '%' OR institution_name ILIKE '%' || sqlc.narg('search')::text || '%' OR EXISTS (SELECT 1 FROM unnest(lender_names) lender_name WHERE lender_name ILIKE '%' || sqlc.narg('search')::text || '%'))
)
SELECT COUNT(*)::bigint FROM filtered;

-- name: ListDashboardPipelineBottleneckItems :many
WITH ranked_bb AS (
    SELECT bp.id, bp.project_identity_id, ROW_NUMBER() OVER (PARTITION BY bp.project_identity_id ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
),
selected_bb AS (SELECT id, project_identity_id FROM ranked_bb WHERE rn = 1),
ranked_gb AS (
    SELECT gp.id, gp.gb_project_identity_id, ROW_NUMBER() OVER (PARTITION BY gp.gb_project_identity_id ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (sqlc.narg('period_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_bb_project gbp JOIN bb_project bp ON bp.id = gbp.bb_project_id JOIN blue_book bb ON bb.id = bp.blue_book_id WHERE gbp.gb_project_id = gp.id AND bb.period_id = sqlc.narg('period_id')::uuid))
),
selected_gb AS (SELECT id, gb_project_identity_id FROM ranked_gb WHERE rn = 1),
worklist AS (
    SELECT 'BB_NO_LENDER'::text AS stage, 'bb_project'::text AS reference_type, sb.id AS project_id, sb.id AS journey_bb_project_id, bp.bb_code::text AS code, bp.project_name::text AS project_name, COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric AS amount_usd, GREATEST((CURRENT_DATE - bp.created_at::date), 0)::int AS age_days, bp.created_at AS relevant_at, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text AS institution_name, ARRAY[]::text[] AS lender_names
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND sqlc.narg('lender_id')::uuid IS NULL AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'INDICATION_NO_LOI'::text, 'bb_project'::text, sb.id, sb.id, bp.bb_code::text, bp.project_name::text, COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(li.created_at)::date FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT MAX(li.created_at) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at), COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id JOIN lender l ON l.id = li.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN lender_indication li ON li.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND li.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'LOI_NO_GB'::text, 'bb_project'::text, sb.id, sb.id, bp.bb_code::text, bp.project_name::text, COALESCE((SELECT SUM(bpc.amount_usd) FROM bb_project_cost bpc WHERE bpc.bb_project_id = bp.id AND bpc.funding_type = 'Foreign'), 0)::numeric, GREATEST((CURRENT_DATE - COALESCE((SELECT MAX(lo.date) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at::date)), 0)::int, COALESCE((SELECT MAX(lo.date)::timestamptz FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id), bp.created_at), COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM bb_project_institution bpi JOIN institution i ON i.id = bpi.institution_id WHERE bpi.bb_project_id = bp.id AND bpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id JOIN lender l ON l.id = lo.lender_id WHERE bp2.project_identity_id = sb.project_identity_id), ARRAY[]::text[])
    FROM selected_bb sb JOIN bb_project bp ON bp.id = sb.id
    WHERE sqlc.narg('publish_year')::int IS NULL AND NOT EXISTS (SELECT 1 FROM bb_project bp2 JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project bp2 JOIN loi lo ON lo.bb_project_id = bp2.id WHERE bp2.project_identity_id = sb.project_identity_id AND lo.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM bb_project_institution bpi WHERE bpi.bb_project_id = bp.id AND bpi.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'GB_NO_DK'::text, 'gb_project'::text, sg.id, COALESCE((SELECT sb.id FROM gb_project_bb_project gbp JOIN selected_bb sb ON sb.id = gbp.bb_project_id WHERE gbp.gb_project_id = gp.id LIMIT 1), (SELECT bp.id FROM gb_project_bb_project gbp JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE gbp.gb_project_id = gp.id LIMIT 1)), gp.gb_code::text, gp.project_name::text, COALESCE((SELECT SUM(gfs.loan_usd + gfs.grant_usd) FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id), 0)::numeric, GREATEST((CURRENT_DATE - gp.created_at::date), 0)::int, gp.created_at, COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = gp.id AND gpi.role = 'Executing Agency'), '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = gp.id), ARRAY[]::text[])
    FROM selected_gb sg JOIN gb_project gp ON gp.id = sg.id
    WHERE NOT EXISTS (SELECT 1 FROM dk_project_gb_project dpg WHERE dpg.gb_project_id = gp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid) OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid))
    UNION ALL
    SELECT 'DK_NO_LA'::text, 'dk_project'::text, dp.id, (SELECT bp.id FROM dk_project_gb_project dpg JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE dpg.dk_project_id = dp.id LIMIT 1), COALESCE(dk.letter_number, dp.id::text)::text, dp.project_name::text, COALESCE((SELECT SUM(dfd.amount_usd + dfd.grant_usd) FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id), 0)::numeric, GREATEST((CURRENT_DATE - COALESCE(dk.date, dp.created_at::date)), 0)::int, COALESCE(dk.date::timestamptz, dp.created_at), COALESCE(i.short_name, i.name, '')::text, COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM dk_financing_detail dfd JOIN lender l ON l.id = dfd.lender_id WHERE dfd.dk_project_id = dp.id), ARRAY[]::text[])
    FROM dk_project dp JOIN daftar_kegiatan dk ON dk.id = dp.dk_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id) AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id AND dfd.lender_id = sqlc.narg('lender_id')::uuid)) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
    UNION ALL
    SELECT 'LA_NOT_EFFECTIVE'::text, 'loan_agreement'::text, la.id, (SELECT bp.id FROM dk_project_gb_project dpg JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id JOIN bb_project bp ON bp.id = gbp.bb_project_id WHERE dpg.dk_project_id = dp.id LIMIT 1), la.loan_code::text, dp.project_name::text, la.amount_usd::numeric, GREATEST((CURRENT_DATE - la.agreement_date), 0)::int, la.agreement_date::timestamptz, COALESCE(i.short_name, i.name, '')::text, ARRAY[COALESCE(l.short_name, l.name)]::text[]
    FROM loan_agreement la JOIN dk_project dp ON dp.id = la.dk_project_id JOIN lender l ON l.id = la.lender_id LEFT JOIN institution i ON i.id = dp.institution_id
    WHERE la.effective_date > CURRENT_DATE AND EXISTS (SELECT 1 FROM dk_project_gb_project dpg JOIN selected_gb sg ON sg.id = dpg.gb_project_id WHERE dpg.dk_project_id = dp.id) AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid) AND (sqlc.narg('institution_id')::uuid IS NULL OR dp.institution_id = sqlc.narg('institution_id')::uuid)
),
filtered AS (
    SELECT * FROM worklist
    WHERE (sqlc.narg('stage')::text IS NULL OR stage = sqlc.narg('stage')::text)
      AND (sqlc.narg('min_age_days')::int IS NULL OR age_days >= sqlc.narg('min_age_days')::int)
      AND (sqlc.narg('search')::text IS NULL OR project_name ILIKE '%' || sqlc.narg('search')::text || '%' OR code ILIKE '%' || sqlc.narg('search')::text || '%' OR institution_name ILIKE '%' || sqlc.narg('search')::text || '%' OR EXISTS (SELECT 1 FROM unnest(lender_names) lender_name WHERE lender_name ILIKE '%' || sqlc.narg('search')::text || '%'))
)
SELECT
    project_id,
    reference_type,
    journey_bb_project_id,
    code,
    project_name,
    stage AS current_stage,
    age_days,
    amount_usd,
    institution_name,
    lender_names,
    relevant_at
FROM filtered
ORDER BY
    CASE WHEN sqlc.arg('sort')::text = 'stage' AND sqlc.arg('order')::text = 'asc' THEN stage END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'stage' AND sqlc.arg('order')::text = 'desc' THEN stage END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'asc' THEN project_name END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'desc' THEN project_name END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'amount_usd' AND sqlc.arg('order')::text = 'asc' THEN amount_usd END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'amount_usd' AND sqlc.arg('order')::text = 'desc' THEN amount_usd END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'age_days' AND sqlc.arg('order')::text = 'asc' THEN age_days END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'age_days' AND sqlc.arg('order')::text = 'desc' THEN age_days END DESC,
    age_days DESC,
    amount_usd DESC,
    project_name ASC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int;

-- ===== DASHBOARD GREEN BOOK READINESS =====

-- name: GetDashboardGreenBookReadinessSummary :one
WITH ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (
          (sqlc.narg('green_book_id')::uuid IS NOT NULL AND gb.id = sqlc.narg('green_book_id')::uuid)
          OR (sqlc.narg('green_book_id')::uuid IS NULL AND gb.status = 'active')
      )
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_institution gpi
              WHERE gpi.gb_project_id = gp.id
                AND gpi.institution_id = sqlc.narg('institution_id')::uuid
          )
          OR EXISTS (
              SELECT 1
              FROM gb_funding_source gfs
              WHERE gfs.gb_project_id = gp.id
                AND gfs.institution_id = sqlc.narg('institution_id')::uuid
          )
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_funding_source gfs
              WHERE gfs.gb_project_id = gp.id
                AND gfs.lender_id = sqlc.narg('lender_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
readiness AS (
    SELECT
        sg.id,
        CASE WHEN EXISTS (SELECT 1 FROM gb_project_bb_project gbp WHERE gbp.gb_project_id = sg.id) THEN 20 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Executing Agency')
                AND EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Implementing Agency') THEN 15 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_project_location gpl WHERE gpl.gb_project_id = sg.id) THEN 10 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id) THEN 20 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_activity ga WHERE ga.gb_project_id = sg.id) THEN 15 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_disbursement_plan gdp WHERE gdp.gb_project_id = sg.id) THEN 10 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_activity ga JOIN gb_funding_allocation gfa ON gfa.gb_activity_id = ga.id WHERE ga.gb_project_id = sg.id) THEN 10 ELSE 0 END AS readiness_score,
        COALESCE((SELECT COUNT(DISTINCT gfs.lender_id) FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id), 0)::int AS lender_count
    FROM selected_gb sg
),
classified AS (
    SELECT
        id,
        readiness_score,
        lender_count,
        lender_count > 1 AS is_cofinancing,
        CASE
            WHEN readiness_score >= 85 THEN 'READY'
            WHEN readiness_score >= 60 THEN 'PARTIAL'
            ELSE 'INCOMPLETE'
        END AS readiness_status
    FROM readiness
),
filtered AS (
    SELECT *
    FROM classified c
    WHERE (
        sqlc.narg('readiness_status')::text IS NULL
        OR (sqlc.narg('readiness_status')::text = 'COFINANCING' AND c.is_cofinancing)
        OR c.readiness_status = sqlc.narg('readiness_status')::text
    )
),
funding AS (
    SELECT
        f.id,
        COALESCE(SUM(gfs.loan_usd), 0)::numeric AS loan_usd,
        COALESCE(SUM(gfs.grant_usd), 0)::numeric AS grant_usd,
        COALESCE(SUM(gfs.local_usd), 0)::numeric AS local_usd
    FROM filtered f
    LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = f.id
    GROUP BY f.id
)
SELECT
    COUNT(*)::bigint AS total_projects,
    COALESCE(SUM(funding.loan_usd), 0)::numeric AS total_loan_usd,
    COALESCE(SUM(funding.grant_usd), 0)::numeric AS total_grant_usd,
    COALESCE(SUM(funding.local_usd), 0)::numeric AS total_local_usd,
    COUNT(*) FILTER (WHERE filtered.is_cofinancing)::bigint AS projects_with_cofinancing,
    COUNT(*) FILTER (WHERE filtered.readiness_status = 'INCOMPLETE')::bigint AS projects_incomplete,
    COUNT(*) FILTER (WHERE filtered.readiness_status = 'READY')::bigint AS projects_ready,
    COUNT(*) FILTER (WHERE filtered.readiness_status = 'PARTIAL')::bigint AS projects_partial
FROM filtered
LEFT JOIN funding ON funding.id = filtered.id;

-- name: GetDashboardGreenBookReadinessDisbursementByYear :many
WITH ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (
          (sqlc.narg('green_book_id')::uuid IS NOT NULL AND gb.id = sqlc.narg('green_book_id')::uuid)
          OR (sqlc.narg('green_book_id')::uuid IS NULL AND gb.status = 'active')
      )
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid)
          OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid)
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
readiness AS (
    SELECT
        sg.id,
        CASE WHEN EXISTS (SELECT 1 FROM gb_project_bb_project gbp WHERE gbp.gb_project_id = sg.id) THEN 20 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Executing Agency')
                AND EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Implementing Agency') THEN 15 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_project_location gpl WHERE gpl.gb_project_id = sg.id) THEN 10 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id) THEN 20 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_activity ga WHERE ga.gb_project_id = sg.id) THEN 15 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_disbursement_plan gdp WHERE gdp.gb_project_id = sg.id) THEN 10 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_activity ga JOIN gb_funding_allocation gfa ON gfa.gb_activity_id = ga.id WHERE ga.gb_project_id = sg.id) THEN 10 ELSE 0 END AS readiness_score,
        COALESCE((SELECT COUNT(DISTINCT gfs.lender_id) FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id), 0)::int AS lender_count
    FROM selected_gb sg
),
classified AS (
    SELECT
        id,
        lender_count > 1 AS is_cofinancing,
        CASE
            WHEN readiness_score >= 85 THEN 'READY'
            WHEN readiness_score >= 60 THEN 'PARTIAL'
            ELSE 'INCOMPLETE'
        END AS readiness_status
    FROM readiness
),
filtered AS (
    SELECT *
    FROM classified c
    WHERE (
        sqlc.narg('readiness_status')::text IS NULL
        OR (sqlc.narg('readiness_status')::text = 'COFINANCING' AND c.is_cofinancing)
        OR c.readiness_status = sqlc.narg('readiness_status')::text
    )
)
SELECT
    gdp.year,
    COALESCE(SUM(gdp.amount_usd), 0)::numeric AS amount_usd
FROM filtered f
JOIN gb_disbursement_plan gdp ON gdp.gb_project_id = f.id
GROUP BY gdp.year
ORDER BY gdp.year ASC;

-- name: GetDashboardGreenBookReadinessFundingAllocation :one
WITH ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (
          (sqlc.narg('green_book_id')::uuid IS NOT NULL AND gb.id = sqlc.narg('green_book_id')::uuid)
          OR (sqlc.narg('green_book_id')::uuid IS NULL AND gb.status = 'active')
      )
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid)
          OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid)
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
readiness AS (
    SELECT
        sg.id,
        CASE WHEN EXISTS (SELECT 1 FROM gb_project_bb_project gbp WHERE gbp.gb_project_id = sg.id) THEN 20 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Executing Agency')
                AND EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Implementing Agency') THEN 15 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_project_location gpl WHERE gpl.gb_project_id = sg.id) THEN 10 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id) THEN 20 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_activity ga WHERE ga.gb_project_id = sg.id) THEN 15 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_disbursement_plan gdp WHERE gdp.gb_project_id = sg.id) THEN 10 ELSE 0 END
        + CASE WHEN EXISTS (SELECT 1 FROM gb_activity ga JOIN gb_funding_allocation gfa ON gfa.gb_activity_id = ga.id WHERE ga.gb_project_id = sg.id) THEN 10 ELSE 0 END AS readiness_score,
        COALESCE((SELECT COUNT(DISTINCT gfs.lender_id) FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id), 0)::int AS lender_count
    FROM selected_gb sg
),
classified AS (
    SELECT
        id,
        lender_count > 1 AS is_cofinancing,
        CASE
            WHEN readiness_score >= 85 THEN 'READY'
            WHEN readiness_score >= 60 THEN 'PARTIAL'
            ELSE 'INCOMPLETE'
        END AS readiness_status
    FROM readiness
),
filtered AS (
    SELECT *
    FROM classified c
    WHERE (
        sqlc.narg('readiness_status')::text IS NULL
        OR (sqlc.narg('readiness_status')::text = 'COFINANCING' AND c.is_cofinancing)
        OR c.readiness_status = sqlc.narg('readiness_status')::text
    )
)
SELECT
    COALESCE(SUM(gfa.services), 0)::numeric AS services,
    COALESCE(SUM(gfa.constructions), 0)::numeric AS constructions,
    COALESCE(SUM(gfa.goods), 0)::numeric AS goods,
    COALESCE(SUM(gfa.trainings), 0)::numeric AS trainings,
    COALESCE(SUM(gfa.other), 0)::numeric AS other
FROM filtered f
JOIN gb_activity ga ON ga.gb_project_id = f.id
JOIN gb_funding_allocation gfa ON gfa.gb_activity_id = ga.id;

-- name: ListDashboardGreenBookReadinessItems :many
WITH ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.gb_code,
        gp.project_name,
        gb.id AS green_book_id,
        gb.publish_year,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND (
          (sqlc.narg('green_book_id')::uuid IS NOT NULL AND gb.id = sqlc.narg('green_book_id')::uuid)
          OR (sqlc.narg('green_book_id')::uuid IS NULL AND gb.status = 'active')
      )
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = gp.id AND gpi.institution_id = sqlc.narg('institution_id')::uuid)
          OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.institution_id = sqlc.narg('institution_id')::uuid)
      )
      AND (
          sqlc.narg('lender_id')::uuid IS NULL
          OR EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id AND gfs.lender_id = sqlc.narg('lender_id')::uuid)
      )
),
selected_gb AS (
    SELECT id, gb_code, project_name, green_book_id, publish_year
    FROM ranked_gb
    WHERE rn = 1
),
checks AS (
    SELECT
        sg.id,
        sg.gb_code,
        sg.project_name,
        sg.green_book_id,
        sg.publish_year,
        EXISTS (SELECT 1 FROM gb_project_bb_project gbp WHERE gbp.gb_project_id = sg.id) AS has_bb_reference,
        EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Executing Agency')
            AND EXISTS (SELECT 1 FROM gb_project_institution gpi WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Implementing Agency') AS has_ea_ia,
        EXISTS (SELECT 1 FROM gb_project_location gpl WHERE gpl.gb_project_id = sg.id) AS has_location,
        EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id) AS has_funding_source,
        EXISTS (SELECT 1 FROM gb_activity ga WHERE ga.gb_project_id = sg.id) AS has_activities,
        EXISTS (SELECT 1 FROM gb_disbursement_plan gdp WHERE gdp.gb_project_id = sg.id) AS has_disbursement_plan,
        EXISTS (SELECT 1 FROM gb_activity ga JOIN gb_funding_allocation gfa ON gfa.gb_activity_id = ga.id WHERE ga.gb_project_id = sg.id) AS has_funding_allocation,
        COALESCE((SELECT COUNT(DISTINCT gfs.lender_id) FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id), 0)::int AS lender_count,
        COALESCE((SELECT SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd) FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id), 0)::numeric AS total_funding_usd,
        COALESCE((SELECT STRING_AGG(DISTINCT COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name)) FROM gb_project_institution gpi JOIN institution i ON i.id = gpi.institution_id WHERE gpi.gb_project_id = sg.id AND gpi.role = 'Executing Agency'), '')::text AS institution_name,
        COALESCE((SELECT ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)) FROM gb_funding_source gfs JOIN lender l ON l.id = gfs.lender_id WHERE gfs.gb_project_id = sg.id), ARRAY[]::text[]) AS lender_names
    FROM selected_gb sg
),
scored AS (
    SELECT
        c.*,
        (CASE WHEN has_bb_reference THEN 20 ELSE 0 END
        + CASE WHEN has_ea_ia THEN 15 ELSE 0 END
        + CASE WHEN has_location THEN 10 ELSE 0 END
        + CASE WHEN has_funding_source THEN 20 ELSE 0 END
        + CASE WHEN has_activities THEN 15 ELSE 0 END
        + CASE WHEN has_disbursement_plan THEN 10 ELSE 0 END
        + CASE WHEN has_funding_allocation THEN 10 ELSE 0 END)::int AS readiness_score
    FROM checks c
),
classified AS (
    SELECT
        s.*,
        lender_count > 1 AS is_cofinancing,
        CASE
            WHEN readiness_score >= 85 THEN 'READY'
            WHEN readiness_score >= 60 THEN 'PARTIAL'
            ELSE 'INCOMPLETE'
        END AS readiness_status,
        ARRAY_REMOVE(ARRAY[
            CASE WHEN NOT has_bb_reference THEN 'BB reference' END,
            CASE WHEN NOT has_ea_ia THEN 'Executing/Implementing Agency' END,
            CASE WHEN NOT has_location THEN 'Location' END,
            CASE WHEN NOT has_funding_source THEN 'Funding source' END,
            CASE WHEN NOT has_activities THEN 'Activities' END,
            CASE WHEN NOT has_disbursement_plan THEN 'Disbursement plan' END,
            CASE WHEN NOT has_funding_allocation THEN 'Funding allocation' END
        ]::text[], NULL) AS missing_fields
    FROM scored s
)
SELECT
    id AS project_id,
    green_book_id,
    gb_code,
    project_name,
    publish_year,
    readiness_score,
    readiness_status,
    is_cofinancing,
    missing_fields::text[] AS missing_fields,
    total_funding_usd,
    institution_name,
    lender_names::text[] AS lender_names
FROM classified c
WHERE (
    sqlc.narg('readiness_status')::text IS NULL
    OR (sqlc.narg('readiness_status')::text = 'COFINANCING' AND c.is_cofinancing)
    OR c.readiness_status = sqlc.narg('readiness_status')::text
)
ORDER BY
    readiness_score ASC,
    is_cofinancing DESC,
    total_funding_usd DESC,
    project_name ASC;

-- ===== DASHBOARD LENDER FINANCING MIX =====

-- name: GetDashboardLenderFinancingMixSummary :one
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
      AND (
          sqlc.narg('publish_year')::int IS NULL
          OR EXISTS (
              SELECT 1
              FROM bb_project bp_related
              JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp_related.id
              JOIN gb_project gp ON gp.id = gbp.gb_project_id
              JOIN green_book gb ON gb.id = gp.green_book_id
              WHERE bp_related.project_identity_id = bp.project_identity_id
                AND gp.status = 'active'
                AND gb.status = 'active'
                AND gb.publish_year = sqlc.narg('publish_year')::int
          )
      )
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
      AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
bb_foreign_cost AS (
    SELECT
        sb.id AS bb_project_id,
        COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id
    WHERE bpc.funding_type = 'Foreign'
    GROUP BY sb.id
),
all_sources AS (
    SELECT
        li.lender_id,
        l.type AS lender_type,
        sb.id AS project_id,
        bfc.amount_usd
    FROM selected_bb sb
    JOIN bb_foreign_cost bfc ON bfc.bb_project_id = sb.id
    JOIN (SELECT DISTINCT bb_project_id, lender_id FROM lender_indication) li ON li.bb_project_id = sb.id
    JOIN lender l ON l.id = li.lender_id
    WHERE sqlc.narg('currency')::text IS NULL
      AND (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR li.lender_id = sqlc.narg('lender_id')::uuid)
    UNION ALL
    SELECT
        loi.lender_id,
        l.type AS lender_type,
        sb.id AS project_id,
        bfc.amount_usd
    FROM selected_bb sb
    JOIN bb_foreign_cost bfc ON bfc.bb_project_id = sb.id
    JOIN (SELECT DISTINCT bb_project_id, lender_id FROM loi) loi ON loi.bb_project_id = sb.id
    JOIN lender l ON l.id = loi.lender_id
    WHERE sqlc.narg('currency')::text IS NULL
      AND (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR loi.lender_id = sqlc.narg('lender_id')::uuid)
    UNION ALL
    SELECT
        gfs.lender_id,
        l.type AS lender_type,
        sg.id AS project_id,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd), 0)::numeric AS amount_usd
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    JOIN lender l ON l.id = gfs.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR gfs.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR gfs.currency = sqlc.narg('currency')::text)
    GROUP BY gfs.lender_id, l.type, sg.id
    UNION ALL
    SELECT
        dfd.lender_id,
        l.type AS lender_type,
        sd.id AS project_id,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    JOIN lender l ON l.id = dfd.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR dfd.currency = sqlc.narg('currency')::text)
    GROUP BY dfd.lender_id, l.type, sd.id
    UNION ALL
    SELECT
        la.lender_id,
        l.type AS lender_type,
        sd.id AS project_id,
        la.amount_usd::numeric AS amount_usd
    FROM selected_dk sd
    JOIN loan_agreement la ON la.dk_project_id = sd.id
    JOIN lender l ON l.id = la.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR la.currency = sqlc.narg('currency')::text)
),
la_type_amounts AS (
    SELECT
        l.type AS lender_type,
        COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    JOIN loan_agreement la ON la.dk_project_id = sd.id
    JOIN lender l ON l.id = la.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR la.currency = sqlc.narg('currency')::text)
    GROUP BY l.type
),
gb_cofinancing AS (
    SELECT sg.id
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    JOIN lender l ON l.id = gfs.lender_id
    GROUP BY sg.id
    HAVING COUNT(DISTINCT gfs.lender_id) > 1
       AND (
           (sqlc.narg('lender_type')::text IS NULL AND sqlc.narg('lender_id')::uuid IS NULL AND sqlc.narg('currency')::text IS NULL)
           OR BOOL_OR((sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
              AND (sqlc.narg('lender_id')::uuid IS NULL OR gfs.lender_id = sqlc.narg('lender_id')::uuid)
              AND (sqlc.narg('currency')::text IS NULL OR gfs.currency = sqlc.narg('currency')::text))
       )
),
dk_cofinancing AS (
    SELECT sd.id
    FROM selected_dk sd
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    JOIN lender l ON l.id = dfd.lender_id
    GROUP BY sd.id
    HAVING COUNT(DISTINCT dfd.lender_id) > 1
       AND (
           (sqlc.narg('lender_type')::text IS NULL AND sqlc.narg('lender_id')::uuid IS NULL AND sqlc.narg('currency')::text IS NULL)
           OR BOOL_OR((sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
              AND (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
              AND (sqlc.narg('currency')::text IS NULL OR dfd.currency = sqlc.narg('currency')::text))
       )
),
cofinancing_projects AS (
    SELECT id FROM gb_cofinancing
    UNION
    SELECT id FROM dk_cofinancing
)
SELECT
    COUNT(DISTINCT all_sources.lender_id)::bigint AS total_lenders,
    COALESCE((SELECT amount_usd FROM la_type_amounts WHERE lender_type = 'Bilateral'), 0)::numeric AS bilateral_usd,
    COALESCE((SELECT amount_usd FROM la_type_amounts WHERE lender_type = 'Multilateral'), 0)::numeric AS multilateral_usd,
    COALESCE((SELECT amount_usd FROM la_type_amounts WHERE lender_type = 'KSA'), 0)::numeric AS ksa_usd,
    (SELECT COUNT(*) FROM cofinancing_projects)::bigint AS cofinancing_projects
FROM all_sources;

-- name: GetDashboardLenderCertaintyLadder :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
      AND (
          sqlc.narg('publish_year')::int IS NULL
          OR EXISTS (
              SELECT 1
              FROM bb_project bp_related
              JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp_related.id
              JOIN gb_project gp ON gp.id = gbp.gb_project_id
              JOIN green_book gb ON gb.id = gp.green_book_id
              WHERE bp_related.project_identity_id = bp.project_identity_id
                AND gp.status = 'active'
                AND gb.status = 'active'
                AND gb.publish_year = sqlc.narg('publish_year')::int
          )
      )
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
      AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
bb_foreign_cost AS (
    SELECT
        sb.id AS bb_project_id,
        COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id
    WHERE bpc.funding_type = 'Foreign'
    GROUP BY sb.id
),
stage_sources AS (
    SELECT
        'LENDER_INDICATION'::text AS stage,
        li.lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_name,
        l.type::text AS lender_type,
        sb.id AS project_id,
        bfc.amount_usd
    FROM selected_bb sb
    JOIN bb_foreign_cost bfc ON bfc.bb_project_id = sb.id
    JOIN (SELECT DISTINCT bb_project_id, lender_id FROM lender_indication) li ON li.bb_project_id = sb.id
    JOIN lender l ON l.id = li.lender_id
    WHERE sqlc.narg('currency')::text IS NULL
      AND (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR li.lender_id = sqlc.narg('lender_id')::uuid)
    UNION ALL
    SELECT
        'LOI'::text AS stage,
        loi.lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_name,
        l.type::text AS lender_type,
        sb.id AS project_id,
        bfc.amount_usd
    FROM selected_bb sb
    JOIN bb_foreign_cost bfc ON bfc.bb_project_id = sb.id
    JOIN (SELECT DISTINCT bb_project_id, lender_id FROM loi) loi ON loi.bb_project_id = sb.id
    JOIN lender l ON l.id = loi.lender_id
    WHERE sqlc.narg('currency')::text IS NULL
      AND (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR loi.lender_id = sqlc.narg('lender_id')::uuid)
    UNION ALL
    SELECT
        'GB_FUNDING_SOURCE'::text AS stage,
        gfs.lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_name,
        l.type::text AS lender_type,
        sg.id AS project_id,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd), 0)::numeric AS amount_usd
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    JOIN lender l ON l.id = gfs.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR gfs.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR gfs.currency = sqlc.narg('currency')::text)
    GROUP BY gfs.lender_id, COALESCE(l.short_name, l.name), l.type, sg.id
    UNION ALL
    SELECT
        'DK_FINANCING'::text AS stage,
        dfd.lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_name,
        l.type::text AS lender_type,
        sd.id AS project_id,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    JOIN lender l ON l.id = dfd.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR dfd.currency = sqlc.narg('currency')::text)
    GROUP BY dfd.lender_id, COALESCE(l.short_name, l.name), l.type, sd.id
    UNION ALL
    SELECT
        'LA'::text AS stage,
        la.lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_name,
        l.type::text AS lender_type,
        sd.id AS project_id,
        la.amount_usd::numeric AS amount_usd
    FROM selected_dk sd
    JOIN loan_agreement la ON la.dk_project_id = sd.id
    JOIN lender l ON l.id = la.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR la.currency = sqlc.narg('currency')::text)
)
SELECT
    stage,
    lender_id,
    lender_name,
    lender_type,
    COUNT(DISTINCT project_id)::bigint AS project_count,
    COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM stage_sources
GROUP BY stage, lender_id, lender_name, lender_type
ORDER BY
    CASE stage
        WHEN 'LENDER_INDICATION' THEN 1
        WHEN 'LOI' THEN 2
        WHEN 'GB_FUNDING_SOURCE' THEN 3
        WHEN 'DK_FINANCING' THEN 4
        WHEN 'LA' THEN 5
        ELSE 99
    END,
    amount_usd DESC,
    lender_name ASC;

-- name: GetDashboardLenderConversion :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
      AND (
          sqlc.narg('publish_year')::int IS NULL
          OR EXISTS (
              SELECT 1
              FROM bb_project bp_related
              JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp_related.id
              JOIN gb_project gp ON gp.id = gbp.gb_project_id
              JOIN green_book gb ON gb.id = gp.green_book_id
              WHERE bp_related.project_identity_id = bp.project_identity_id
                AND gp.status = 'active'
                AND gb.status = 'active'
                AND gb.publish_year = sqlc.narg('publish_year')::int
          )
      )
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
      AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
bb_foreign_cost AS (
    SELECT
        sb.id AS bb_project_id,
        COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id
    WHERE bpc.funding_type = 'Foreign'
    GROUP BY sb.id
),
indication AS (
    SELECT
        li.lender_id,
        COUNT(DISTINCT sb.id)::bigint AS project_count,
        COALESCE(SUM(bfc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    JOIN bb_foreign_cost bfc ON bfc.bb_project_id = sb.id
    JOIN (SELECT DISTINCT bb_project_id, lender_id FROM lender_indication) li ON li.bb_project_id = sb.id
    JOIN lender l ON l.id = li.lender_id
    WHERE sqlc.narg('currency')::text IS NULL
      AND (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR li.lender_id = sqlc.narg('lender_id')::uuid)
    GROUP BY li.lender_id
),
loi_stage AS (
    SELECT
        loi.lender_id,
        COUNT(DISTINCT sb.id)::bigint AS project_count,
        COALESCE(SUM(bfc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    JOIN bb_foreign_cost bfc ON bfc.bb_project_id = sb.id
    JOIN (SELECT DISTINCT bb_project_id, lender_id FROM loi) loi ON loi.bb_project_id = sb.id
    JOIN lender l ON l.id = loi.lender_id
    WHERE sqlc.narg('currency')::text IS NULL
      AND (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR loi.lender_id = sqlc.narg('lender_id')::uuid)
    GROUP BY loi.lender_id
),
gb_stage AS (
    SELECT
        gfs.lender_id,
        COUNT(DISTINCT sg.id)::bigint AS project_count,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd), 0)::numeric AS amount_usd
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    JOIN lender l ON l.id = gfs.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR gfs.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR gfs.currency = sqlc.narg('currency')::text)
    GROUP BY gfs.lender_id
),
dk_stage AS (
    SELECT
        dfd.lender_id,
        COUNT(DISTINCT sd.id)::bigint AS project_count,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    JOIN lender l ON l.id = dfd.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR dfd.currency = sqlc.narg('currency')::text)
    GROUP BY dfd.lender_id
),
la_stage AS (
    SELECT
        la.lender_id,
        COUNT(DISTINCT sd.id)::bigint AS project_count,
        COALESCE(SUM(la.amount_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    JOIN loan_agreement la ON la.dk_project_id = sd.id
    JOIN lender l ON l.id = la.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR la.currency = sqlc.narg('currency')::text)
    GROUP BY la.lender_id
),
all_lenders AS (
    SELECT lender_id FROM indication
    UNION
    SELECT lender_id FROM loi_stage
    UNION
    SELECT lender_id FROM gb_stage
    UNION
    SELECT lender_id FROM dk_stage
    UNION
    SELECT lender_id FROM la_stage
)
SELECT
    l.id AS lender_id,
    COALESCE(l.short_name, l.name)::text AS lender_name,
    l.type::text AS lender_type,
    COALESCE(indication.project_count, 0)::bigint AS indication_count,
    COALESCE(loi_stage.project_count, 0)::bigint AS loi_count,
    COALESCE(gb_stage.project_count, 0)::bigint AS gb_count,
    COALESCE(dk_stage.project_count, 0)::bigint AS dk_count,
    COALESCE(la_stage.project_count, 0)::bigint AS la_count,
    COALESCE(indication.amount_usd, 0)::numeric AS indication_usd,
    COALESCE(la_stage.amount_usd, 0)::numeric AS la_usd
FROM all_lenders al
JOIN lender l ON l.id = al.lender_id
LEFT JOIN indication ON indication.lender_id = al.lender_id
LEFT JOIN loi_stage ON loi_stage.lender_id = al.lender_id
LEFT JOIN gb_stage ON gb_stage.lender_id = al.lender_id
LEFT JOIN dk_stage ON dk_stage.lender_id = al.lender_id
LEFT JOIN la_stage ON la_stage.lender_id = al.lender_id
ORDER BY la_usd DESC, indication_usd DESC, lender_name ASC;

-- name: GetDashboardCurrencyExposure :many
WITH ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id
    FROM ranked_gb
    WHERE rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
      AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
currency_sources AS (
    SELECT
        'GB_FUNDING_SOURCE'::text AS stage,
        gfs.currency::text AS currency,
        sg.id AS project_id,
        (gfs.loan_original + gfs.grant_original + gfs.local_original)::numeric AS amount_original,
        (gfs.loan_usd + gfs.grant_usd + gfs.local_usd)::numeric AS amount_usd
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    JOIN lender l ON l.id = gfs.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR gfs.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR gfs.currency = sqlc.narg('currency')::text)
    UNION ALL
    SELECT
        'DK_FINANCING'::text AS stage,
        dfd.currency::text AS currency,
        sd.id AS project_id,
        (dfd.amount_original + dfd.grant_original + dfd.counterpart_original)::numeric AS amount_original,
        (dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd)::numeric AS amount_usd
    FROM selected_dk sd
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    JOIN lender l ON l.id = dfd.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR dfd.currency = sqlc.narg('currency')::text)
    UNION ALL
    SELECT
        'LA'::text AS stage,
        la.currency::text AS currency,
        sd.id AS project_id,
        la.amount_original::numeric AS amount_original,
        la.amount_usd::numeric AS amount_usd
    FROM selected_dk sd
    JOIN loan_agreement la ON la.dk_project_id = sd.id
    JOIN lender l ON l.id = la.lender_id
    WHERE (sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
      AND (sqlc.narg('lender_id')::uuid IS NULL OR la.lender_id = sqlc.narg('lender_id')::uuid)
      AND (sqlc.narg('currency')::text IS NULL OR la.currency = sqlc.narg('currency')::text)
)
SELECT
    currency,
    stage,
    COUNT(DISTINCT project_id)::bigint AS project_count,
    COALESCE(SUM(amount_original), 0)::numeric AS amount_original,
    COALESCE(SUM(amount_usd), 0)::numeric AS amount_usd
FROM currency_sources
GROUP BY currency, stage
ORDER BY amount_usd DESC, currency ASC, stage ASC;

-- name: ListDashboardCofinancingItems :many
WITH ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.gb_code,
        gp.project_name,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_code, project_name
    FROM ranked_gb
    WHERE rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id, dp.project_name, dk.letter_number
    FROM dk_project dp
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
      AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
gb_items AS (
    SELECT
        sg.id AS project_id,
        'GB'::text AS reference_type,
        sg.gb_code::text AS project_code,
        sg.project_name::text AS project_name,
        COUNT(DISTINCT gfs.lender_id)::bigint AS lender_count,
        COALESCE(ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)), ARRAY[]::text[])::text[] AS lender_names,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd), 0)::numeric AS amount_usd,
        BOOL_OR((sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
            AND (sqlc.narg('lender_id')::uuid IS NULL OR gfs.lender_id = sqlc.narg('lender_id')::uuid)
            AND (sqlc.narg('currency')::text IS NULL OR gfs.currency = sqlc.narg('currency')::text)) AS matches_filter
    FROM selected_gb sg
    JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    JOIN lender l ON l.id = gfs.lender_id
    GROUP BY sg.id, sg.gb_code, sg.project_name
),
dk_items AS (
    SELECT
        sd.id AS project_id,
        'DK'::text AS reference_type,
        COALESCE(sd.letter_number, '')::text AS project_code,
        sd.project_name::text AS project_name,
        COUNT(DISTINCT dfd.lender_id)::bigint AS lender_count,
        COALESCE(ARRAY_AGG(DISTINCT COALESCE(l.short_name, l.name) ORDER BY COALESCE(l.short_name, l.name)), ARRAY[]::text[])::text[] AS lender_names,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd), 0)::numeric AS amount_usd,
        BOOL_OR((sqlc.narg('lender_type')::text IS NULL OR l.type = sqlc.narg('lender_type')::text)
            AND (sqlc.narg('lender_id')::uuid IS NULL OR dfd.lender_id = sqlc.narg('lender_id')::uuid)
            AND (sqlc.narg('currency')::text IS NULL OR dfd.currency = sqlc.narg('currency')::text)) AS matches_filter
    FROM selected_dk sd
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    JOIN lender l ON l.id = dfd.lender_id
    GROUP BY sd.id, sd.letter_number, sd.project_name
)
SELECT
    project_id,
    reference_type,
    project_code,
    project_name,
    lender_count,
    lender_names,
    amount_usd
FROM (
    SELECT * FROM gb_items
    UNION ALL
    SELECT * FROM dk_items
) items
WHERE lender_count > 1
  AND (
      (sqlc.narg('lender_type')::text IS NULL AND sqlc.narg('lender_id')::uuid IS NULL AND sqlc.narg('currency')::text IS NULL)
      OR matches_filter
  )
ORDER BY amount_usd DESC, project_name ASC
LIMIT 50;

-- ===== DASHBOARD K/L PORTFOLIO PERFORMANCE =====

-- name: GetDashboardKLPortfolioPerformanceItems :many
WITH RECURSIVE institution_tree AS (
    SELECT
        i.id,
        i.id AS root_id,
        COALESCE(i.short_name, i.name)::text AS root_name
    FROM institution i
    WHERE i.parent_id IS NULL
    UNION ALL
    SELECT
        child.id,
        parent.root_id,
        parent.root_name
    FROM institution child
    JOIN institution_tree parent ON parent.id = child.parent_id
),
selected_institution AS (
    SELECT it.root_id
    FROM institution_tree it
    WHERE it.id = sqlc.narg('institution_id')::uuid
),
ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (sqlc.narg('period_id')::uuid IS NULL OR bb.period_id = sqlc.narg('period_id')::uuid)
      AND (
          sqlc.narg('publish_year')::int IS NULL
          OR EXISTS (
              SELECT 1
              FROM bb_project bp_related
              JOIN gb_project_bb_project gbp ON gbp.bb_project_id = bp_related.id
              JOIN gb_project gp ON gp.id = gbp.gb_project_id
              JOIN green_book gb ON gb.id = gp.green_book_id
              WHERE bp_related.project_identity_id = bp.project_identity_id
                AND gp.status = 'active'
                AND gb.status = 'active'
                AND gb.publish_year = sqlc.narg('publish_year')::int
          )
      )
),
selected_bb AS (
    SELECT id, project_identity_id
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
      AND (sqlc.narg('publish_year')::int IS NULL OR gb.publish_year = sqlc.narg('publish_year')::int)
      AND (
          sqlc.narg('period_id')::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = sqlc.narg('period_id')::uuid
          )
      )
),
selected_gb AS (
    SELECT id, gb_project_identity_id
    FROM ranked_gb
    WHERE rn = 1
),
selected_dk AS (
    SELECT DISTINCT dp.id
    FROM dk_project dp
    WHERE (
        sqlc.narg('publish_year')::int IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project gp ON gp.id = dpg.gb_project_id
            JOIN green_book gb ON gb.id = gp.green_book_id
            WHERE dpg.dk_project_id = dp.id
              AND gb.publish_year = sqlc.narg('publish_year')::int
        )
    )
      AND (
        sqlc.narg('period_id')::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = sqlc.narg('period_id')::uuid
        )
    )
),
bb_amount AS (
    SELECT
        sb.id AS project_id,
        COALESCE(SUM(bpc.amount_usd), 0)::numeric AS amount_usd
    FROM selected_bb sb
    LEFT JOIN bb_project_cost bpc ON bpc.bb_project_id = sb.id AND bpc.funding_type = 'Foreign'
    GROUP BY sb.id
),
bb_by_institution_project AS (
    SELECT
        it.root_id AS institution_id,
        it.root_name AS institution_name,
        sb.id AS project_id,
        MAX(ba.amount_usd)::numeric AS amount_usd
    FROM selected_bb sb
    JOIN bb_project_institution bpi ON bpi.bb_project_id = sb.id
    JOIN institution_tree it ON it.id = bpi.institution_id
    JOIN bb_amount ba ON ba.project_id = sb.id
    WHERE (sqlc.narg('institution_role')::text IS NULL OR bpi.role = sqlc.narg('institution_role')::text)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR it.root_id = (SELECT root_id FROM selected_institution)
      )
    GROUP BY it.root_id, it.root_name, sb.id
),
gb_amount AS (
    SELECT
        sg.id AS project_id,
        COALESCE(SUM(gfs.loan_usd + gfs.grant_usd + gfs.local_usd), 0)::numeric AS amount_usd
    FROM selected_gb sg
    LEFT JOIN gb_funding_source gfs ON gfs.gb_project_id = sg.id
    GROUP BY sg.id
),
gb_by_institution_project AS (
    SELECT
        it.root_id AS institution_id,
        it.root_name AS institution_name,
        sg.id AS project_id,
        MAX(ga.amount_usd)::numeric AS amount_usd,
        BOOL_OR(EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = sg.id)) AS has_funding_source,
        BOOL_OR(EXISTS (SELECT 1 FROM gb_activity act WHERE act.gb_project_id = sg.id)) AS has_activities,
        BOOL_OR(EXISTS (SELECT 1 FROM gb_disbursement_plan plan WHERE plan.gb_project_id = sg.id)) AS has_disbursement_plan
    FROM selected_gb sg
    JOIN gb_project_institution gpi ON gpi.gb_project_id = sg.id
    JOIN institution_tree it ON it.id = gpi.institution_id
    JOIN gb_amount ga ON ga.project_id = sg.id
    WHERE (sqlc.narg('institution_role')::text IS NULL OR gpi.role = sqlc.narg('institution_role')::text)
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR it.root_id = (SELECT root_id FROM selected_institution)
      )
    GROUP BY it.root_id, it.root_name, sg.id
),
dk_amount AS (
    SELECT
        sd.id AS project_id,
        COALESCE(SUM(dfd.amount_usd + dfd.grant_usd + dfd.counterpart_usd), 0)::numeric AS amount_usd
    FROM selected_dk sd
    LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = sd.id
    GROUP BY sd.id
),
dk_institution_source AS (
    SELECT DISTINCT
        it.root_id AS institution_id,
        it.root_name AS institution_name,
        sd.id AS project_id
    FROM selected_dk sd
    JOIN dk_project dp ON dp.id = sd.id
    JOIN institution_tree it ON it.id = dp.institution_id
    WHERE (sqlc.narg('institution_role')::text IS NULL OR sqlc.narg('institution_role')::text = 'Executing Agency')
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR it.root_id = (SELECT root_id FROM selected_institution)
      )
    UNION
    SELECT DISTINCT
        it.root_id AS institution_id,
        it.root_name AS institution_name,
        sd.id AS project_id
    FROM selected_dk sd
    JOIN dk_project_gb_project dpg ON dpg.dk_project_id = sd.id
    JOIN gb_project_institution gpi ON gpi.gb_project_id = dpg.gb_project_id AND gpi.role = 'Implementing Agency'
    JOIN institution_tree it ON it.id = gpi.institution_id
    WHERE (sqlc.narg('institution_role')::text IS NULL OR sqlc.narg('institution_role')::text = 'Implementing Agency')
      AND (
          sqlc.narg('institution_id')::uuid IS NULL
          OR it.root_id = (SELECT root_id FROM selected_institution)
      )
),
dk_by_institution_project AS (
    SELECT
        dis.institution_id,
        dis.institution_name,
        dis.project_id,
        MAX(da.amount_usd)::numeric AS amount_usd
    FROM dk_institution_source dis
    JOIN dk_amount da ON da.project_id = dis.project_id
    GROUP BY dis.institution_id, dis.institution_name, dis.project_id
),
la_by_institution_project AS (
    SELECT
        dk.institution_id,
        dk.institution_name,
        dk.project_id,
        la.id AS loan_agreement_id,
        la.effective_date,
        la.closing_date,
        la.amount_usd::numeric AS amount_usd
    FROM dk_by_institution_project dk
    JOIN loan_agreement la ON la.dk_project_id = dk.project_id
),

bb_agg AS (
    SELECT
        institution_id,
        MAX(institution_name)::text AS institution_name,
        COUNT(DISTINCT project_id)::bigint AS bb_project_count,
        COALESCE(SUM(amount_usd), 0)::numeric AS bb_pipeline_usd
    FROM bb_by_institution_project
    GROUP BY institution_id
),
gb_agg AS (
    SELECT
        institution_id,
        MAX(institution_name)::text AS institution_name,
        COUNT(DISTINCT project_id)::bigint AS gb_project_count,
        COALESCE(SUM(amount_usd), 0)::numeric AS gb_pipeline_usd,
        COALESCE(AVG(
            (CASE WHEN has_funding_source THEN 1 ELSE 0 END
            + CASE WHEN has_activities THEN 1 ELSE 0 END
            + CASE WHEN has_disbursement_plan THEN 1 ELSE 0 END)::numeric / 3 * 100
        ), 100)::numeric AS data_completeness_score
    FROM gb_by_institution_project
    GROUP BY institution_id
),
dk_agg AS (
    SELECT
        institution_id,
        MAX(institution_name)::text AS institution_name,
        COUNT(DISTINCT project_id)::bigint AS dk_project_count,
        COALESCE(SUM(amount_usd), 0)::numeric AS dk_pipeline_usd
    FROM dk_by_institution_project
    GROUP BY institution_id
),
la_agg AS (
    SELECT
        institution_id,
        MAX(institution_name)::text AS institution_name,
        COUNT(DISTINCT loan_agreement_id)::bigint AS la_count,
        COALESCE(SUM(amount_usd), 0)::numeric AS la_commitment_usd
    FROM la_by_institution_project
    GROUP BY institution_id
),
gb_risk AS (
    SELECT
        gb.institution_id,
        COUNT(DISTINCT gb.project_id)::bigint AS risk_count
    FROM gb_by_institution_project gb
    WHERE NOT EXISTS (
        SELECT 1
        FROM dk_project_gb_project dpg
        WHERE dpg.gb_project_id = gb.project_id
    )
    GROUP BY gb.institution_id
),
dk_risk AS (
    SELECT
        dk.institution_id,
        COUNT(DISTINCT dk.project_id)::bigint AS risk_count
    FROM dk_by_institution_project dk
    WHERE NOT EXISTS (
        SELECT 1
        FROM loan_agreement la
        WHERE la.dk_project_id = dk.project_id
    )
    GROUP BY dk.institution_id
),
la_risk AS (
    SELECT
        la.institution_id,
        COUNT(DISTINCT la.loan_agreement_id) FILTER (
            WHERE la.closing_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '12 months'
        )
        + COUNT(DISTINCT la.loan_agreement_id) FILTER (
            WHERE la.effective_date < CURRENT_DATE
              AND la.closing_date > la.effective_date
              AND ((CURRENT_DATE - la.effective_date)::numeric / NULLIF((la.closing_date - la.effective_date)::numeric, 0)) >= 0.7
        ) AS risk_count
    FROM la_by_institution_project la
    GROUP BY la.institution_id
),
risk_agg AS (
    SELECT institution_id, COALESCE(SUM(risk_count), 0)::bigint AS risk_count
    FROM (
        SELECT institution_id, risk_count FROM gb_risk
        UNION ALL
        SELECT institution_id, risk_count FROM dk_risk
        UNION ALL
        SELECT institution_id, risk_count FROM la_risk
    ) risks
    GROUP BY institution_id
),
all_institutions AS (
    SELECT institution_id, institution_name FROM bb_agg
    UNION
    SELECT institution_id, institution_name FROM gb_agg
    UNION
    SELECT institution_id, institution_name FROM dk_agg
    UNION
    SELECT institution_id, institution_name FROM la_agg
),
scored AS (
    SELECT
        ai.institution_id,
        ai.institution_name,
        COALESCE(bb.bb_project_count, 0)::bigint AS bb_project_count,
        COALESCE(gb.gb_project_count, 0)::bigint AS gb_project_count,
        COALESCE(dk.dk_project_count, 0)::bigint AS dk_project_count,
        COALESCE(la.la_count, 0)::bigint AS la_count,
        (COALESCE(bb.bb_pipeline_usd, 0) + COALESCE(gb.gb_pipeline_usd, 0) + COALESCE(dk.dk_pipeline_usd, 0))::numeric AS pipeline_usd,
        COALESCE(la.la_commitment_usd, 0)::numeric AS la_commitment_usd,
        COALESCE(risk.risk_count, 0)::bigint AS risk_count,
        CASE
            WHEN COALESCE(gb.gb_project_count, 0) = 0 THEN 100
            ELSE LEAST(100, COALESCE(dk.dk_project_count, 0)::numeric / NULLIF(gb.gb_project_count, 0)::numeric * 100)
        END::numeric AS pipeline_progress_score,
        COALESCE(gb.data_completeness_score, 100)::numeric AS data_completeness_score
    FROM all_institutions ai
    LEFT JOIN bb_agg bb ON bb.institution_id = ai.institution_id
    LEFT JOIN gb_agg gb ON gb.institution_id = ai.institution_id
    LEFT JOIN dk_agg dk ON dk.institution_id = ai.institution_id
    LEFT JOIN la_agg la ON la.institution_id = ai.institution_id
    LEFT JOIN risk_agg risk ON risk.institution_id = ai.institution_id
),
final_items AS (
    SELECT
        s.*,
        (
            s.pipeline_progress_score * 0.45
            + s.data_completeness_score * 0.35
            + GREATEST(0, 100 - s.risk_count::numeric * 20) * 0.20
        )::numeric AS performance_score
    FROM scored s
)
SELECT
    institution_id,
    institution_name,
    bb_project_count,
    gb_project_count,
    dk_project_count,
    la_count,
    pipeline_usd,
    la_commitment_usd,
    risk_count,
    performance_score,
    CASE
        WHEN performance_score >= 80 THEN 'Good'
        WHEN performance_score >= 60 THEN 'Watch'
        ELSE 'High Risk'
    END::text AS performance_category
FROM final_items
ORDER BY
    CASE WHEN sqlc.narg('sort_by')::text = 'pipeline_usd' THEN pipeline_usd END DESC,
    CASE WHEN sqlc.narg('sort_by')::text = 'la_commitment_usd' THEN la_commitment_usd END DESC,
    CASE WHEN sqlc.narg('sort_by')::text = 'risk_count' THEN risk_count END DESC,
    pipeline_usd DESC,
    institution_name ASC;

-- ===== DASHBOARD DATA QUALITY & GOVERNANCE =====

-- name: ListDashboardDataQualityIssues :many
WITH ranked_bb AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.bb_code,
        bp.project_name,
        ROW_NUMBER() OVER (
            PARTITION BY bp.project_identity_id
            ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC, bp.created_at DESC
        ) AS rn
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
),
latest_bb AS (
    SELECT id, project_identity_id, bb_code, project_name
    FROM ranked_bb
    WHERE rn = 1
),
ranked_gb AS (
    SELECT
        gp.id,
        gp.gb_project_identity_id,
        gp.gb_code,
        gp.project_name,
        ROW_NUMBER() OVER (
            PARTITION BY gp.gb_project_identity_id
            ORDER BY gb.revision_number DESC, gb.created_at DESC, gp.created_at DESC
        ) AS rn
    FROM gb_project gp
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gp.status = 'active'
      AND gb.status = 'active'
),
latest_gb AS (
    SELECT id, gb_project_identity_id, gb_code, project_name
    FROM ranked_gb
    WHERE rn = 1
),
issues AS (
    SELECT 'warning'::text AS severity, 'bb_project'::text AS module, 'BB_WITHOUT_BAPPENAS_PARTNER'::text AS issue_type, bp.id AS record_id, (bp.bb_code || ' - ' || bp.project_name)::text AS record_label, 'Blue Book project belum memiliki Mitra Kerja Bappenas'::text AS message, 'Lengkapi Mitra Kerja Bappenas jika sudah tersedia.'::text AS recommended_action, false::boolean AS is_resolved
    FROM latest_bb bp
    WHERE NOT EXISTS (SELECT 1 FROM bb_project_bappenas_partner bpp WHERE bpp.bb_project_id = bp.id)

    UNION ALL
    SELECT 'info'::text, 'bb_project'::text, 'BB_INDICATION_WITHOUT_LOI'::text, li.id, (bp.bb_code || ' - ' || bp.project_name)::text, 'Blue Book project memiliki lender indication tetapi belum memiliki Letter of Intent untuk lender tersebut'::text, 'Follow up lender untuk penerbitan Letter of Intent.'::text, false::boolean
    FROM lender_indication li
    JOIN latest_bb bp ON bp.id = li.bb_project_id
    WHERE NOT EXISTS (SELECT 1 FROM loi l WHERE l.bb_project_id = li.bb_project_id AND l.lender_id = li.lender_id)

    UNION ALL
    SELECT 'warning'::text, 'bb_project'::text, 'LOI_WITHOUT_GB'::text, loi.id, (bp.bb_code || ' - ' || bp.project_name)::text, 'Letter of Intent sudah ada tetapi belum menjadi referensi Green Book'::text, 'Cek readiness dan usulkan proyek ke Green Book.'::text, false::boolean
    FROM loi
    JOIN latest_bb bp ON bp.id = loi.bb_project_id
    WHERE NOT EXISTS (SELECT 1 FROM gb_project_bb_project gbp WHERE gbp.bb_project_id = bp.id)

    UNION ALL
    SELECT 'error'::text, 'gb_project'::text, 'GB_WITHOUT_BB_REFERENCE'::text, gp.id, (gp.gb_code || ' - ' || gp.project_name)::text, 'Green Book project tidak memiliki referensi Blue Book'::text, 'Tambahkan minimal satu referensi Blue Book project.'::text, false::boolean
    FROM latest_gb gp
    WHERE NOT EXISTS (SELECT 1 FROM gb_project_bb_project gbp WHERE gbp.gb_project_id = gp.id)

    UNION ALL
    SELECT 'warning'::text, 'gb_project'::text, 'GB_WITHOUT_FUNDING_SOURCE'::text, gp.id, (gp.gb_code || ' - ' || gp.project_name)::text, 'Green Book project belum memiliki funding source'::text, 'Lengkapi lender dan nilai funding source.'::text, false::boolean
    FROM latest_gb gp
    WHERE NOT EXISTS (SELECT 1 FROM gb_funding_source gfs WHERE gfs.gb_project_id = gp.id)

    UNION ALL
    SELECT 'warning'::text, 'gb_project'::text, 'GB_WITHOUT_DISBURSEMENT_PLAN'::text, gp.id, (gp.gb_code || ' - ' || gp.project_name)::text, 'Green Book project belum memiliki disbursement plan'::text, 'Lengkapi disbursement plan per tahun.'::text, false::boolean
    FROM latest_gb gp
    WHERE NOT EXISTS (SELECT 1 FROM gb_disbursement_plan gdp WHERE gdp.gb_project_id = gp.id)

    UNION ALL
    SELECT 'warning'::text, 'gb_project'::text, 'GB_WITHOUT_ACTIVITY'::text, gp.id, (gp.gb_code || ' - ' || gp.project_name)::text, 'Green Book project belum memiliki activities'::text, 'Lengkapi daftar activities Green Book.'::text, false::boolean
    FROM latest_gb gp
    WHERE NOT EXISTS (SELECT 1 FROM gb_activity ga WHERE ga.gb_project_id = gp.id)

    UNION ALL
    SELECT 'error'::text, 'dk_project'::text, 'DK_WITHOUT_FINANCING_DETAIL'::text, dp.id, dp.project_name::text, 'Daftar Kegiatan project belum memiliki financing detail'::text, 'Lengkapi financing detail sesuai lender yang valid.'::text, false::boolean
    FROM dk_project dp
    WHERE NOT EXISTS (SELECT 1 FROM dk_financing_detail dfd WHERE dfd.dk_project_id = dp.id)

    UNION ALL
    SELECT 'warning'::text, 'dk_project'::text, 'DK_WITHOUT_ACTIVITY_DETAIL'::text, dp.id, dp.project_name::text, 'Daftar Kegiatan project belum memiliki activity detail'::text, 'Lengkapi activity detail Daftar Kegiatan.'::text, false::boolean
    FROM dk_project dp
    WHERE NOT EXISTS (SELECT 1 FROM dk_activity_detail dad WHERE dad.dk_project_id = dp.id)

    UNION ALL
    SELECT 'warning'::text, 'dk_project'::text, 'DK_WITHOUT_LA'::text, dp.id, dp.project_name::text, 'Daftar Kegiatan project belum memiliki Loan Agreement'::text, 'Dorong penyelesaian negosiasi dan legal agreement.'::text, false::boolean
    FROM dk_project dp
    WHERE NOT EXISTS (SELECT 1 FROM loan_agreement la WHERE la.dk_project_id = dp.id)

    UNION ALL
    SELECT 'info'::text, 'loan_agreement'::text, 'LA_NOT_EFFECTIVE'::text, la.id, (la.loan_code || ' - ' || dp.project_name)::text, 'Loan Agreement belum efektif'::text, 'Pantau pemenuhan effectiveness conditions.'::text, false::boolean
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE la.effective_date > CURRENT_DATE

    UNION ALL
    SELECT 'warning'::text, 'gb_project'::text, 'CURRENCY_USD_MISMATCH'::text, gfs.id, (gp.gb_code || ' - ' || gp.project_name)::text, 'Funding source memakai currency USD tetapi nilai original tidak sama dengan nilai USD'::text, 'Samakan nilai original dan USD untuk currency USD.'::text, false::boolean
    FROM gb_funding_source gfs
    JOIN gb_project gp ON gp.id = gfs.gb_project_id
    WHERE gfs.currency = 'USD'
      AND (gfs.loan_original != gfs.loan_usd OR gfs.grant_original != gfs.grant_usd OR gfs.local_original != gfs.local_usd)

    UNION ALL
    SELECT 'warning'::text, 'dk_project'::text, 'CURRENCY_USD_MISMATCH'::text, dfd.id, dp.project_name::text, 'DK financing memakai currency USD tetapi nilai original tidak sama dengan nilai USD'::text, 'Samakan nilai original dan USD untuk currency USD.'::text, false::boolean
    FROM dk_financing_detail dfd
    JOIN dk_project dp ON dp.id = dfd.dk_project_id
    WHERE dfd.currency = 'USD'
      AND (dfd.amount_original != dfd.amount_usd OR dfd.grant_original != dfd.grant_usd OR dfd.counterpart_original != dfd.counterpart_usd)

    UNION ALL
    SELECT 'warning'::text, 'loan_agreement'::text, 'CURRENCY_USD_MISMATCH'::text, la.id, (la.loan_code || ' - ' || dp.project_name)::text, 'Loan Agreement memakai currency USD tetapi amount original tidak sama dengan amount USD'::text, 'Samakan amount original dan amount USD untuk currency USD.'::text, false::boolean
    FROM loan_agreement la
    JOIN dk_project dp ON dp.id = la.dk_project_id
    WHERE la.currency = 'USD'
      AND la.amount_original != la.amount_usd
)
SELECT
    severity,
    module,
    issue_type,
    record_id,
    record_label,
    message,
    recommended_action,
    is_resolved
FROM issues
WHERE (sqlc.narg('severity')::text IS NULL OR severity = sqlc.narg('severity')::text)
  AND (sqlc.narg('module')::text IS NULL OR module = sqlc.narg('module')::text)
  AND (sqlc.narg('issue_type')::text IS NULL OR issue_type = sqlc.narg('issue_type')::text)
  AND (NOT sqlc.arg('only_unresolved')::boolean OR is_resolved = false)
ORDER BY
    CASE severity WHEN 'error' THEN 1 WHEN 'warning' THEN 2 ELSE 3 END,
    module ASC,
    issue_type ASC,
    record_label ASC;

-- name: CountDashboardAuditEvents :one
SELECT COUNT(*)::bigint AS event_count
FROM audit_log
WHERE changed_at >= NOW() - make_interval(days => sqlc.arg('audit_days')::int);

-- name: GetDashboardAuditSummaryByUser :many
SELECT
    COALESCE(u.username, 'system')::text AS label,
    COUNT(*)::bigint AS event_count,
    MAX(al.changed_at)::timestamptz AS last_changed_at
FROM audit_log al
LEFT JOIN app_user u ON u.id = al.changed_by
WHERE al.changed_at >= NOW() - make_interval(days => sqlc.arg('audit_days')::int)
GROUP BY COALESCE(u.username, 'system')
ORDER BY event_count DESC, label ASC
LIMIT 10;

-- name: GetDashboardAuditSummaryByTable :many
SELECT
    al.table_name::text AS label,
    COUNT(*)::bigint AS event_count,
    MAX(al.changed_at)::timestamptz AS last_changed_at
FROM audit_log al
WHERE al.changed_at >= NOW() - make_interval(days => sqlc.arg('audit_days')::int)
GROUP BY al.table_name
ORDER BY event_count DESC, label ASC
LIMIT 10;

-- name: ListDashboardAuditRecentActivity :many
SELECT
    al.id,
    COALESCE(u.username, 'system')::text AS username,
    al.action::text AS action,
    al.table_name::text AS table_name,
    al.record_id,
    al.changed_at
FROM audit_log al
LEFT JOIN app_user u ON u.id = al.changed_by
WHERE al.changed_at >= NOW() - make_interval(days => sqlc.arg('audit_days')::int)
ORDER BY al.changed_at DESC
LIMIT 20;
