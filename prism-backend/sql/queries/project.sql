-- ===== PROJECT MASTER =====

-- name: ListDashboardStageSummaries :many
WITH latest_bb_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
),
stage_entities AS (
    SELECT
        'BB'::text AS stage,
        lbp.id AS entity_id,
        COALESCE((
            SELECT SUM(pc.amount_usd)
            FROM bb_project_cost pc
            WHERE pc.bb_project_id = lbp.id
              AND pc.funding_type = 'Foreign'
              AND pc.funding_category = 'Loan'
        ), 0)::numeric AS amount_usd
    FROM latest_bb_projects lbp

    UNION ALL

    SELECT
        'GB'::text AS stage,
        gp.id AS entity_id,
        COALESCE((
            SELECT SUM(gfs.loan_usd)
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        ), 0)::numeric AS amount_usd
    FROM gb_project gp
    WHERE gp.status = 'active'
      AND (
          COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
          )
      )

    UNION ALL

    SELECT
        'DK'::text AS stage,
        dp.id AS entity_id,
        COALESCE((
            SELECT SUM(dfd.amount_usd)
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
        ), 0)::numeric AS amount_usd
    FROM dk_project dp
    WHERE (
        COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
        )
    )

    UNION ALL

    SELECT
        'LA'::text AS stage,
        la.id AS entity_id,
        la.amount_usd::numeric AS amount_usd
    FROM loan_agreement la
    WHERE (
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
stage_order AS (
    SELECT 'BB'::text AS stage, 1::int AS sort_order
    UNION ALL
    SELECT 'GB'::text AS stage, 2::int AS sort_order
    UNION ALL
    SELECT 'DK'::text AS stage, 3::int AS sort_order
    UNION ALL
    SELECT 'LA'::text AS stage, 4::int AS sort_order
)
SELECT
    so.stage,
    COUNT(se.entity_id)::int AS project_count,
    COALESCE(SUM(se.amount_usd), 0)::numeric AS total_loan_usd
FROM stage_order so
LEFT JOIN stage_entities se ON se.stage = so.stage
GROUP BY so.stage, so.sort_order
ORDER BY so.sort_order;

-- name: ListDashboardStageRegionGroups :many
WITH latest_bb_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
),
stage_entities AS (
    SELECT
        'BB'::text AS stage,
        lbp.id AS entity_id,
        lbp.id AS location_owner_id,
        COALESCE((
            SELECT SUM(pc.amount_usd)
            FROM bb_project_cost pc
            WHERE pc.bb_project_id = lbp.id
              AND pc.funding_type = 'Foreign'
              AND pc.funding_category = 'Loan'
        ), 0)::numeric AS amount_usd
    FROM latest_bb_projects lbp

    UNION ALL

    SELECT
        'GB'::text AS stage,
        gp.id AS entity_id,
        gp.id AS location_owner_id,
        COALESCE((
            SELECT SUM(gfs.loan_usd)
            FROM gb_funding_source gfs
            WHERE gfs.gb_project_id = gp.id
        ), 0)::numeric AS amount_usd
    FROM gb_project gp
    WHERE gp.status = 'active'
      AND (
          COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
          OR EXISTS (
              SELECT 1
              FROM gb_project_bb_project gbp
              JOIN bb_project bp ON bp.id = gbp.bb_project_id
              JOIN blue_book bb ON bb.id = bp.blue_book_id
              WHERE gbp.gb_project_id = gp.id
                AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
          )
      )

    UNION ALL

    SELECT
        'DK'::text AS stage,
        dp.id AS entity_id,
        dp.id AS location_owner_id,
        COALESCE((
            SELECT SUM(dfd.amount_usd)
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
        ), 0)::numeric AS amount_usd
    FROM dk_project dp
    WHERE (
        COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = dp.id
              AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
        )
    )

    UNION ALL

    SELECT
        'LA'::text AS stage,
        la.id AS entity_id,
        ladp.dk_project_id AS location_owner_id,
        ladp.allocation_usd::numeric AS amount_usd
    FROM loan_agreement_dk_project ladp
    JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
    WHERE (
        COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM dk_project_gb_project dpg
            JOIN gb_project_bb_project gbp ON gbp.gb_project_id = dpg.gb_project_id
            JOIN bb_project bp ON bp.id = gbp.bb_project_id
            JOIN blue_book bb ON bb.id = bp.blue_book_id
            WHERE dpg.dk_project_id = ladp.dk_project_id
              AND bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
        )
    )
),
stage_locations AS (
    SELECT se.stage, se.entity_id, se.location_owner_id, se.amount_usd, bpl.region_id
    FROM stage_entities se
    JOIN bb_project_location bpl ON bpl.bb_project_id = se.location_owner_id
    WHERE se.stage = 'BB'

    UNION ALL

    SELECT se.stage, se.entity_id, se.location_owner_id, se.amount_usd, gpl.region_id
    FROM stage_entities se
    JOIN gb_project_location gpl ON gpl.gb_project_id = se.location_owner_id
    WHERE se.stage = 'GB'

    UNION ALL

    SELECT se.stage, se.entity_id, se.location_owner_id, se.amount_usd, dkpl.region_id
    FROM stage_entities se
    JOIN dk_project_location dkpl ON dkpl.dk_project_id = se.location_owner_id
    WHERE se.stage IN ('DK', 'LA')
),
normalized_region_groups AS (
    SELECT
        stage,
        entity_id,
        label,
        level,
        SUM(amount_usd)::numeric AS amount_usd
    FROM (
        SELECT DISTINCT
            sl.stage,
            sl.entity_id,
            sl.location_owner_id,
            CASE
                WHEN r.type = 'COUNTRY' THEN 'Indonesia'
                WHEN r.type = 'PROVINCE' THEN COALESCE(r.region_group, r.name)
                ELSE COALESCE(parent_province.region_group, r.name)
            END::text AS label,
            CASE
                WHEN r.type = 'COUNTRY' THEN 'Nasional'
                ELSE 'Region Group'
            END::text AS level,
            sl.amount_usd
        FROM stage_locations sl
        JOIN region r ON r.id = sl.region_id
        LEFT JOIN region parent_province ON parent_province.code = r.parent_code
            AND parent_province.type = 'PROVINCE'
    ) normalized
    GROUP BY stage, entity_id, label, level
),
aggregated_region_groups AS (
    SELECT
        stage,
        ''::text AS id,
        label,
        level,
        COUNT(*)::int AS project_count,
        COALESCE(SUM(amount_usd), 0)::numeric AS foreign_loan_usd
    FROM normalized_region_groups
    WHERE label IS NOT NULL
    GROUP BY stage, label, level
),
ranked_region_groups AS (
    SELECT
        stage,
        id,
        label,
        level,
        project_count,
        foreign_loan_usd,
        ROW_NUMBER() OVER (
            PARTITION BY stage
            ORDER BY project_count DESC, label ASC
        ) AS rank_number
    FROM aggregated_region_groups
)
SELECT
    stage,
    id,
    label,
    level,
    project_count,
    foreign_loan_usd
FROM ranked_region_groups
WHERE rank_number <= 6
ORDER BY
    CASE stage
        WHEN 'BB' THEN 1
        WHEN 'GB' THEN 2
        WHEN 'DK' THEN 3
        WHEN 'LA' THEN 4
        ELSE 5
    END,
    project_count DESC,
    label ASC;

-- name: ListProjectMaster :many
WITH RECURSIVE institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
project_rows AS (
    SELECT
        bp.id,
        bp.blue_book_id,
        bb.period_id,
        bp.project_identity_id,
        bp.bb_code,
        bp.project_name,
        bp.program_title_id,
        COALESCE(pt.title, '')::text AS program_title,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(la_alloc.allocation_usd)
                FROM (
                    SELECT DISTINCT ladp.loan_agreement_id, ladp.dk_project_id, ladp.allocation_usd
                    FROM gb_project_bb_project gbp
                    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                    WHERE gbp.bb_project_id = bp.id
                ) la_alloc
            ), 0)
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(dk_financing.amount_usd)
                FROM (
                    SELECT DISTINCT dfd.id, dfd.amount_usd
                    FROM gb_project_bb_project gbp
                    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                    JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
                    WHERE gbp.bb_project_id = bp.id
                ) dk_financing
            ), 0)
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Monitoring'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'LA'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'DK'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'GB'
            ELSE 'BB'
        END::text AS pipeline_status,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Ongoing'
            ELSE 'Pipeline'
        END::text AS project_status,
        ARRAY(
            SELECT DISTINCT type_label
            FROM (
                SELECT l.type AS type_label
                FROM lender_indication li
                JOIN lender l ON l.id = li.lender_id
                WHERE li.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                JOIN lender l ON l.id = gfs.lender_id
                WHERE gbp.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
                JOIN lender l ON l.id = dfd.lender_id
                WHERE gbp.bb_project_id = bp.id
                  AND dfd.lender_id IS NOT NULL
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                JOIN lender l ON l.id = la.lender_id
                WHERE gbp.bb_project_id = bp.id
            ) loan_types
            WHERE type_label IS NOT NULL
            ORDER BY type_label
        )::text[] AS loan_types,
        ARRAY(
            SELECT DISTINCT li.lender_id
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
            ORDER BY li.lender_id
        )::uuid[] AS indication_lender_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS indication_lenders,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT root.ancestor_id
            FROM bb_project_institution bpi
            JOIN LATERAL (
                SELECT ia.ancestor_id
                FROM institution_ancestors ia
                WHERE ia.institution_id = bpi.institution_id
                  AND ia.parent_id IS NULL
                ORDER BY ia.depth DESC
                LIMIT 1
            ) root ON TRUE
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY root.ancestor_id
        )::uuid[] AS executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(i.short_name, i.name)
            FROM bb_project_institution bpi
            JOIN institution i ON i.id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY COALESCE(i.short_name, i.name)
        )::text[] AS executing_agencies,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT dfd.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
              AND dfd.lender_id IS NOT NULL
            ORDER BY dfd.lender_id
        )::uuid[] AS dk_lender_ids,
        ARRAY(
            SELECT DISTINCT la.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY la.lender_id
        )::uuid[] AS loan_agreement_lender_ids,
        ARRAY(
            SELECT DISTINCT dp.institution_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
              AND dp.institution_id IS NOT NULL
            ORDER BY dp.institution_id
        )::uuid[] AS dk_executing_agency_ids,
        ARRAY(
            SELECT DISTINCT root.ancestor_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN LATERAL (
                SELECT ia.ancestor_id
                FROM institution_ancestors ia
                WHERE ia.institution_id = dp.institution_id
                  AND ia.parent_id IS NULL
                ORDER BY ia.depth DESC
                LIMIT 1
            ) root ON TRUE
            WHERE gbp.bb_project_id = bp.id
              AND dp.institution_id IS NOT NULL
            ORDER BY root.ancestor_id
        )::uuid[] AS dk_executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS fixed_lenders,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids,
        ARRAY(
            SELECT DISTINCT r.name
            FROM bb_project_location bpl
            JOIN region r ON r.id = bpl.region_id
            WHERE bpl.bb_project_id = bp.id
            ORDER BY r.name
        )::text[] AS locations,
        ARRAY(
            SELECT DISTINCT dk.date::text
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY dk.date::text
        )::text[] AS dk_dates,
        EXISTS (
            SELECT 1
            FROM loi loi_flag
            WHERE loi_flag.bb_project_id = bp.id
        )::boolean AS has_loi,
        EXISTS (
            SELECT 1
            FROM lender_indication li_flag
            WHERE li_flag.bb_project_id = bp.id
        )::boolean AS has_lender_indication,
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
            FROM bb_project newer
            JOIN blue_book newer_bb ON newer_bb.id = newer.blue_book_id
            JOIN blue_book current_bb ON current_bb.id = bp.blue_book_id
            WHERE newer.project_identity_id = bp.project_identity_id
              AND newer.status = 'active'
              AND (
                  newer_bb.revision_number > current_bb.revision_number
                  OR (
                      newer_bb.revision_number = current_bb.revision_number
                      AND newer_bb.created_at > current_bb.created_at
                  )
              )
        )::boolean AS has_newer_revision,
        CONCAT(
            'BB ',
            p.name,
            CASE
                WHEN bb.revision_number > 0 THEN CONCAT(' Revisi ke-', bb.revision_number)
                ELSE ''
            END
        )::text AS blue_book_revision_label,
        ARRAY(
            SELECT DISTINCT gp.gb_code
            FROM gb_project_bb_project gbp
            JOIN gb_project gp ON gp.id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gp.gb_code
        )::text[] AS gb_codes,
        ARRAY(
            SELECT DISTINCT CONCAT(
                'GB ',
                gb_pub.publish_year,
                CASE
                    WHEN gb_pub.revision_number > 0 THEN CONCAT(' Revisi ke-', gb_pub.revision_number)
                    ELSE ''
                END
            )
            FROM gb_project_bb_project gbp
            JOIN gb_project gp ON gp.id = gbp.gb_project_id
            JOIN green_book gb_pub ON gb_pub.id = gp.green_book_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY 1
        )::text[] AS green_book_revision_labels
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    JOIN period p ON p.id = bb.period_id
    LEFT JOIN program_title pt ON pt.id = bp.program_title_id
    WHERE bp.status = 'active'
      AND (
          sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id
              FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id
                AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1
          )
      )
),
filtered_projects AS (
    SELECT *
    FROM project_rows pr
    WHERE (COALESCE(cardinality(sqlc.arg('loan_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('loan_types')::text[])
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('indication_lender_ids')::uuid[]), 0) = 0 OR indication_lender_ids && sqlc.arg('indication_lender_ids')::uuid[])
      AND (
          COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
          OR executing_agency_ids && sqlc.arg('executing_agency_ids')::uuid[]
          OR executing_agency_root_ids && sqlc.arg('executing_agency_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('fixed_lender_ids')::uuid[]), 0) = 0 OR fixed_lender_ids && sqlc.arg('fixed_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('dk_lender_ids')::uuid[]), 0) = 0 OR dk_lender_ids && sqlc.arg('dk_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('loan_agreement_lender_ids')::uuid[]), 0) = 0 OR loan_agreement_lender_ids && sqlc.arg('loan_agreement_lender_ids')::uuid[])
      AND (
          COALESCE(cardinality(sqlc.arg('dk_executing_agency_ids')::uuid[]), 0) = 0
          OR dk_executing_agency_ids && sqlc.arg('dk_executing_agency_ids')::uuid[]
          OR dk_executing_agency_root_ids && sqlc.arg('dk_executing_agency_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('reached_stages')::text[]), 0) = 0
          OR 'BB' = ANY(sqlc.arg('reached_stages')::text[])
          OR ('GB' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('GB', 'DK', 'LA', 'Monitoring'))
          OR ('DK' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('DK', 'LA', 'Monitoring'))
          OR ('LA' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('LA', 'Monitoring'))
          OR ('Monitoring' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status = 'Monitoring')
      )
      AND (
          COALESCE(cardinality(sqlc.arg('missing_stages')::text[]), 0) = 0
          OR ('GB' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status = 'BB')
          OR ('DK' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB'))
          OR ('LA' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB', 'DK'))
          OR ('Monitoring' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB', 'DK', 'LA'))
      )
      AND (
          sqlc.narg('has_loi')::boolean IS NULL
          OR EXISTS (
              SELECT 1
              FROM loi loi_filter
              WHERE loi_filter.bb_project_id = pr.id
          ) = sqlc.narg('has_loi')::boolean
      )
      AND (
          sqlc.narg('has_lender_indication')::boolean IS NULL
          OR EXISTS (
              SELECT 1
              FROM lender_indication li_filter
              WHERE li_filter.bb_project_id = pr.id
          ) = sqlc.narg('has_lender_indication')::boolean
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          (sqlc.narg('dk_date_from')::date IS NULL AND sqlc.narg('dk_date_to')::date IS NULL)
          OR EXISTS (
              SELECT 1
              FROM unnest(dk_dates) AS item(dk_date)
              WHERE (sqlc.narg('dk_date_from')::date IS NULL OR item.dk_date::date >= sqlc.narg('dk_date_from')::date)
                AND (sqlc.narg('dk_date_to')::date IS NULL OR item.dk_date::date <= sqlc.narg('dk_date_to')::date)
          )
      )
      AND (
          sqlc.narg('search')::text IS NULL
          OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR EXISTS (
              SELECT 1
              FROM unnest(indication_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(fixed_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(executing_agencies) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
      )
)
SELECT
    id,
    blue_book_id,
    project_identity_id,
    bb_code,
    project_name,
    loan_types,
    indication_lenders,
    executing_agencies,
    fixed_lenders,
    project_status,
    pipeline_status,
    program_title,
    locations,
    foreign_loan_usd,
    dk_dates,
    has_loi,
    has_lender_indication,
    is_latest,
    has_newer_revision,
    blue_book_revision_label,
    gb_codes,
    green_book_revision_labels
FROM filtered_projects
ORDER BY
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'asc' THEN LOWER(project_name) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'project_name' AND sqlc.arg('order')::text = 'desc' THEN LOWER(project_name) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'bb_code' AND sqlc.arg('order')::text = 'asc' THEN LOWER(bb_code) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'bb_code' AND sqlc.arg('order')::text = 'desc' THEN LOWER(bb_code) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'loan_types' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(loan_types, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'loan_types' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(loan_types, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'indication_lenders' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(indication_lenders, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'indication_lenders' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(indication_lenders, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'executing_agencies' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(executing_agencies, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'executing_agencies' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(executing_agencies, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'fixed_lenders' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(fixed_lenders, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'fixed_lenders' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(fixed_lenders, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'project_status' AND sqlc.arg('order')::text = 'asc' THEN LOWER(project_status || ' - ' || pipeline_status) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'project_status' AND sqlc.arg('order')::text = 'desc' THEN LOWER(project_status || ' - ' || pipeline_status) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'pipeline_status' AND sqlc.arg('order')::text = 'asc' THEN LOWER(pipeline_status) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'pipeline_status' AND sqlc.arg('order')::text = 'desc' THEN LOWER(pipeline_status) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'program_title' AND sqlc.arg('order')::text = 'asc' THEN LOWER(program_title) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'program_title' AND sqlc.arg('order')::text = 'desc' THEN LOWER(program_title) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'locations' AND sqlc.arg('order')::text = 'asc' THEN LOWER(array_to_string(locations, ', ')) END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'locations' AND sqlc.arg('order')::text = 'desc' THEN LOWER(array_to_string(locations, ', ')) END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'foreign_loan_usd' AND sqlc.arg('order')::text = 'asc' THEN foreign_loan_usd END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'foreign_loan_usd' AND sqlc.arg('order')::text = 'desc' THEN foreign_loan_usd END DESC,
    CASE WHEN sqlc.arg('sort')::text = 'dk_dates' AND sqlc.arg('order')::text = 'asc' THEN array_to_string(dk_dates, ', ') END ASC,
    CASE WHEN sqlc.arg('sort')::text = 'dk_dates' AND sqlc.arg('order')::text = 'desc' THEN array_to_string(dk_dates, ', ') END DESC,
    project_name ASC,
    bb_code ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountProjectMaster :one
WITH RECURSIVE institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
project_rows AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bb.period_id,
        bp.program_title_id,
        bp.bb_code,
        bp.project_name,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(la_alloc.allocation_usd)
                FROM (
                    SELECT DISTINCT ladp.loan_agreement_id, ladp.dk_project_id, ladp.allocation_usd
                    FROM gb_project_bb_project gbp
                    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                    WHERE gbp.bb_project_id = bp.id
                ) la_alloc
            ), 0)
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(dk_financing.amount_usd)
                FROM (
                    SELECT DISTINCT dfd.id, dfd.amount_usd
                    FROM gb_project_bb_project gbp
                    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                    JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
                    WHERE gbp.bb_project_id = bp.id
                ) dk_financing
            ), 0)
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Monitoring'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'LA'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'DK'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'GB'
            ELSE 'BB'
        END::text AS pipeline_status,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Ongoing'
            ELSE 'Pipeline'
        END::text AS project_status,
        ARRAY(
            SELECT DISTINCT type_label
            FROM (
                SELECT l.type AS type_label
                FROM lender_indication li
                JOIN lender l ON l.id = li.lender_id
                WHERE li.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                JOIN lender l ON l.id = gfs.lender_id
                WHERE gbp.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
                JOIN lender l ON l.id = dfd.lender_id
                WHERE gbp.bb_project_id = bp.id
                  AND dfd.lender_id IS NOT NULL
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                JOIN lender l ON l.id = la.lender_id
                WHERE gbp.bb_project_id = bp.id
            ) loan_types
            WHERE type_label IS NOT NULL
            ORDER BY type_label
        )::text[] AS loan_types,
        ARRAY(
            SELECT DISTINCT li.lender_id
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
            ORDER BY li.lender_id
        )::uuid[] AS indication_lender_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS indication_lenders,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT root.ancestor_id
            FROM bb_project_institution bpi
            JOIN LATERAL (
                SELECT ia.ancestor_id
                FROM institution_ancestors ia
                WHERE ia.institution_id = bpi.institution_id
                  AND ia.parent_id IS NULL
                ORDER BY ia.depth DESC
                LIMIT 1
            ) root ON TRUE
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY root.ancestor_id
        )::uuid[] AS executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(i.short_name, i.name)
            FROM bb_project_institution bpi
            JOIN institution i ON i.id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY COALESCE(i.short_name, i.name)
        )::text[] AS executing_agencies,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT dfd.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
              AND dfd.lender_id IS NOT NULL
            ORDER BY dfd.lender_id
        )::uuid[] AS dk_lender_ids,
        ARRAY(
            SELECT DISTINCT la.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY la.lender_id
        )::uuid[] AS loan_agreement_lender_ids,
        ARRAY(
            SELECT DISTINCT dp.institution_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
              AND dp.institution_id IS NOT NULL
            ORDER BY dp.institution_id
        )::uuid[] AS dk_executing_agency_ids,
        ARRAY(
            SELECT DISTINCT root.ancestor_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN LATERAL (
                SELECT ia.ancestor_id
                FROM institution_ancestors ia
                WHERE ia.institution_id = dp.institution_id
                  AND ia.parent_id IS NULL
                ORDER BY ia.depth DESC
                LIMIT 1
            ) root ON TRUE
            WHERE gbp.bb_project_id = bp.id
              AND dp.institution_id IS NOT NULL
            ORDER BY root.ancestor_id
        )::uuid[] AS dk_executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS fixed_lenders,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids,
        ARRAY(
            SELECT DISTINCT dk.date::text
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY dk.date::text
        )::text[] AS dk_dates
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (
          sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id
              FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id
                AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1
          )
      )
)
SELECT COUNT(*)::bigint
FROM project_rows pr
WHERE (COALESCE(cardinality(sqlc.arg('loan_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('loan_types')::text[])
  AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR period_id = ANY(sqlc.arg('period_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('indication_lender_ids')::uuid[]), 0) = 0 OR indication_lender_ids && sqlc.arg('indication_lender_ids')::uuid[])
  AND (
      COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
      OR executing_agency_ids && sqlc.arg('executing_agency_ids')::uuid[]
      OR executing_agency_root_ids && sqlc.arg('executing_agency_ids')::uuid[]
  )
  AND (COALESCE(cardinality(sqlc.arg('fixed_lender_ids')::uuid[]), 0) = 0 OR fixed_lender_ids && sqlc.arg('fixed_lender_ids')::uuid[])
  AND (COALESCE(cardinality(sqlc.arg('dk_lender_ids')::uuid[]), 0) = 0 OR dk_lender_ids && sqlc.arg('dk_lender_ids')::uuid[])
  AND (COALESCE(cardinality(sqlc.arg('loan_agreement_lender_ids')::uuid[]), 0) = 0 OR loan_agreement_lender_ids && sqlc.arg('loan_agreement_lender_ids')::uuid[])
  AND (
      COALESCE(cardinality(sqlc.arg('dk_executing_agency_ids')::uuid[]), 0) = 0
      OR dk_executing_agency_ids && sqlc.arg('dk_executing_agency_ids')::uuid[]
      OR dk_executing_agency_root_ids && sqlc.arg('dk_executing_agency_ids')::uuid[]
  )
  AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
  AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
  AND (
      COALESCE(cardinality(sqlc.arg('reached_stages')::text[]), 0) = 0
      OR 'BB' = ANY(sqlc.arg('reached_stages')::text[])
      OR ('GB' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('GB', 'DK', 'LA', 'Monitoring'))
      OR ('DK' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('DK', 'LA', 'Monitoring'))
      OR ('LA' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('LA', 'Monitoring'))
      OR ('Monitoring' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status = 'Monitoring')
  )
  AND (
      COALESCE(cardinality(sqlc.arg('missing_stages')::text[]), 0) = 0
      OR ('GB' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status = 'BB')
      OR ('DK' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB'))
      OR ('LA' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB', 'DK'))
      OR ('Monitoring' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB', 'DK', 'LA'))
  )
  AND (
      sqlc.narg('has_loi')::boolean IS NULL
      OR EXISTS (
          SELECT 1
          FROM loi loi_filter
          WHERE loi_filter.bb_project_id = pr.id
      ) = sqlc.narg('has_loi')::boolean
  )
  AND (
      sqlc.narg('has_lender_indication')::boolean IS NULL
      OR EXISTS (
          SELECT 1
          FROM lender_indication li_filter
          WHERE li_filter.bb_project_id = pr.id
      ) = sqlc.narg('has_lender_indication')::boolean
  )
  AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
  AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
  AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
  AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
  AND (
      (sqlc.narg('dk_date_from')::date IS NULL AND sqlc.narg('dk_date_to')::date IS NULL)
      OR EXISTS (
          SELECT 1
          FROM unnest(dk_dates) AS item(dk_date)
          WHERE (sqlc.narg('dk_date_from')::date IS NULL OR item.dk_date::date >= sqlc.narg('dk_date_from')::date)
            AND (sqlc.narg('dk_date_to')::date IS NULL OR item.dk_date::date <= sqlc.narg('dk_date_to')::date)
      )
  )
  AND (
      sqlc.narg('search')::text IS NULL
      OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      OR EXISTS (
          SELECT 1
          FROM unnest(indication_lenders) AS item(label)
          WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
      OR EXISTS (
          SELECT 1
          FROM unnest(fixed_lenders) AS item(label)
          WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
      OR EXISTS (
          SELECT 1
          FROM unnest(executing_agencies) AS item(label)
          WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
      )
  );

-- name: GetProjectMasterFundingSummary :one
WITH RECURSIVE institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
project_rows AS (
    SELECT
        bp.id,
        bp.project_identity_id,
        bb.period_id,
        bp.program_title_id,
        bp.bb_code,
        bp.project_name,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(la_alloc.allocation_usd)
                FROM (
                    SELECT DISTINCT ladp.loan_agreement_id, ladp.dk_project_id, ladp.allocation_usd
                    FROM gb_project_bb_project gbp
                    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                    WHERE gbp.bb_project_id = bp.id
                ) la_alloc
            ), 0)
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(dk_financing.amount_usd)
                FROM (
                    SELECT DISTINCT dfd.id, dfd.amount_usd
                    FROM gb_project_bb_project gbp
                    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                    JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
                    WHERE gbp.bb_project_id = bp.id
                ) dk_financing
            ), 0)
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.grant_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Grant'
            ), 0)
        END::numeric AS foreign_grant_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.local_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Counterpart'
            ), 0)
        END::numeric AS counterpart_usd,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                JOIN monitoring_disbursement md ON md.loan_agreement_id = la.id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Monitoring'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'LA'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'DK'
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'GB'
            ELSE 'BB'
        END::text AS pipeline_status,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                WHERE gbp.bb_project_id = bp.id
            ) THEN 'Ongoing'
            ELSE 'Pipeline'
        END::text AS project_status,
        ARRAY(
            SELECT DISTINCT type_label
            FROM (
                SELECT l.type AS type_label
                FROM lender_indication li
                JOIN lender l ON l.id = li.lender_id
                WHERE li.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                JOIN lender l ON l.id = gfs.lender_id
                WHERE gbp.bb_project_id = bp.id
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
                JOIN lender l ON l.id = dfd.lender_id
                WHERE gbp.bb_project_id = bp.id
                  AND dfd.lender_id IS NOT NULL
                UNION
                SELECT l.type AS type_label
                FROM gb_project_bb_project gbp
                JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
                JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
                JOIN lender l ON l.id = la.lender_id
                WHERE gbp.bb_project_id = bp.id
            ) loan_types
            WHERE type_label IS NOT NULL
            ORDER BY type_label
        )::text[] AS loan_types,
        ARRAY(
            SELECT DISTINCT li.lender_id
            FROM lender_indication li
            WHERE li.bb_project_id = bp.id
            ORDER BY li.lender_id
        )::uuid[] AS indication_lender_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM lender_indication li
            JOIN lender l ON l.id = li.lender_id
            WHERE li.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS indication_lenders,
        ARRAY(
            SELECT DISTINCT bpi.institution_id
            FROM bb_project_institution bpi
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY bpi.institution_id
        )::uuid[] AS executing_agency_ids,
        ARRAY(
            SELECT DISTINCT root.ancestor_id
            FROM bb_project_institution bpi
            JOIN LATERAL (
                SELECT ia.ancestor_id
                FROM institution_ancestors ia
                WHERE ia.institution_id = bpi.institution_id
                  AND ia.parent_id IS NULL
                ORDER BY ia.depth DESC
                LIMIT 1
            ) root ON TRUE
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY root.ancestor_id
        )::uuid[] AS executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(i.short_name, i.name)
            FROM bb_project_institution bpi
            JOIN institution i ON i.id = bpi.institution_id
            WHERE bpi.bb_project_id = bp.id
              AND bpi.role = 'Executing Agency'
            ORDER BY COALESCE(i.short_name, i.name)
        )::text[] AS executing_agencies,
        ARRAY(
            SELECT DISTINCT gfs.lender_id
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY gfs.lender_id
        )::uuid[] AS fixed_lender_ids,
        ARRAY(
            SELECT DISTINCT dfd.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
              AND dfd.lender_id IS NOT NULL
            ORDER BY dfd.lender_id
        )::uuid[] AS dk_lender_ids,
        ARRAY(
            SELECT DISTINCT la.lender_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY la.lender_id
        )::uuid[] AS loan_agreement_lender_ids,
        ARRAY(
            SELECT DISTINCT dp.institution_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            WHERE gbp.bb_project_id = bp.id
              AND dp.institution_id IS NOT NULL
            ORDER BY dp.institution_id
        )::uuid[] AS dk_executing_agency_ids,
        ARRAY(
            SELECT DISTINCT root.ancestor_id
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN LATERAL (
                SELECT ia.ancestor_id
                FROM institution_ancestors ia
                WHERE ia.institution_id = dp.institution_id
                  AND ia.parent_id IS NULL
                ORDER BY ia.depth DESC
                LIMIT 1
            ) root ON TRUE
            WHERE gbp.bb_project_id = bp.id
              AND dp.institution_id IS NOT NULL
            ORDER BY root.ancestor_id
        )::uuid[] AS dk_executing_agency_root_ids,
        ARRAY(
            SELECT DISTINCT COALESCE(l.short_name, l.name)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            JOIN lender l ON l.id = gfs.lender_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY COALESCE(l.short_name, l.name)
        )::text[] AS fixed_lenders,
        ARRAY(
            SELECT DISTINCT bpl.region_id
            FROM bb_project_location bpl
            WHERE bpl.bb_project_id = bp.id
            ORDER BY bpl.region_id
        )::uuid[] AS region_ids,
        ARRAY(
            SELECT DISTINCT dk.date::text
            FROM gb_project_bb_project gbp
            JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
            JOIN dk_project dp ON dp.id = dpg.dk_project_id
            JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
            WHERE gbp.bb_project_id = bp.id
            ORDER BY dk.date::text
        )::text[] AS dk_dates
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (
          sqlc.arg('include_history')::boolean
          OR bp.id = (
              SELECT latest.id
              FROM bb_project latest
              JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
              WHERE latest.project_identity_id = bp.project_identity_id
                AND latest.status = 'active'
              ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
              LIMIT 1
          )
      )
),
filtered_projects AS (
    SELECT *
    FROM project_rows pr
    WHERE (COALESCE(cardinality(sqlc.arg('loan_types')::text[]), 0) = 0 OR loan_types && sqlc.arg('loan_types')::text[])
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('indication_lender_ids')::uuid[]), 0) = 0 OR indication_lender_ids && sqlc.arg('indication_lender_ids')::uuid[])
      AND (
          COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
          OR executing_agency_ids && sqlc.arg('executing_agency_ids')::uuid[]
          OR executing_agency_root_ids && sqlc.arg('executing_agency_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('fixed_lender_ids')::uuid[]), 0) = 0 OR fixed_lender_ids && sqlc.arg('fixed_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('dk_lender_ids')::uuid[]), 0) = 0 OR dk_lender_ids && sqlc.arg('dk_lender_ids')::uuid[])
      AND (COALESCE(cardinality(sqlc.arg('loan_agreement_lender_ids')::uuid[]), 0) = 0 OR loan_agreement_lender_ids && sqlc.arg('loan_agreement_lender_ids')::uuid[])
      AND (
          COALESCE(cardinality(sqlc.arg('dk_executing_agency_ids')::uuid[]), 0) = 0
          OR dk_executing_agency_ids && sqlc.arg('dk_executing_agency_ids')::uuid[]
          OR dk_executing_agency_root_ids && sqlc.arg('dk_executing_agency_ids')::uuid[]
      )
      AND (COALESCE(cardinality(sqlc.arg('project_statuses')::text[]), 0) = 0 OR project_status = ANY(sqlc.arg('project_statuses')::text[]))
      AND (COALESCE(cardinality(sqlc.arg('pipeline_statuses')::text[]), 0) = 0 OR pipeline_status = ANY(sqlc.arg('pipeline_statuses')::text[]))
      AND (
          COALESCE(cardinality(sqlc.arg('reached_stages')::text[]), 0) = 0
          OR 'BB' = ANY(sqlc.arg('reached_stages')::text[])
          OR ('GB' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('GB', 'DK', 'LA', 'Monitoring'))
          OR ('DK' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('DK', 'LA', 'Monitoring'))
          OR ('LA' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status IN ('LA', 'Monitoring'))
          OR ('Monitoring' = ANY(sqlc.arg('reached_stages')::text[]) AND pipeline_status = 'Monitoring')
      )
      AND (
          COALESCE(cardinality(sqlc.arg('missing_stages')::text[]), 0) = 0
          OR ('GB' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status = 'BB')
          OR ('DK' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB'))
          OR ('LA' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB', 'DK'))
          OR ('Monitoring' = ANY(sqlc.arg('missing_stages')::text[]) AND pipeline_status IN ('BB', 'GB', 'DK', 'LA'))
      )
      AND (
          sqlc.narg('has_loi')::boolean IS NULL
          OR EXISTS (
              SELECT 1
              FROM loi loi_filter
              WHERE loi_filter.bb_project_id = pr.id
          ) = sqlc.narg('has_loi')::boolean
      )
      AND (
          sqlc.narg('has_lender_indication')::boolean IS NULL
          OR EXISTS (
              SELECT 1
              FROM lender_indication li_filter
              WHERE li_filter.bb_project_id = pr.id
          ) = sqlc.narg('has_lender_indication')::boolean
      )
      AND (COALESCE(cardinality(sqlc.arg('program_title_ids')::uuid[]), 0) = 0 OR program_title_id = ANY(sqlc.arg('program_title_ids')::uuid[]))
      AND (COALESCE(cardinality(sqlc.arg('region_ids')::uuid[]), 0) = 0 OR region_ids && sqlc.arg('region_ids')::uuid[])
      AND (sqlc.narg('foreign_loan_min')::numeric IS NULL OR foreign_loan_usd >= sqlc.narg('foreign_loan_min')::numeric)
      AND (sqlc.narg('foreign_loan_max')::numeric IS NULL OR foreign_loan_usd <= sqlc.narg('foreign_loan_max')::numeric)
      AND (
          (sqlc.narg('dk_date_from')::date IS NULL AND sqlc.narg('dk_date_to')::date IS NULL)
          OR EXISTS (
              SELECT 1
              FROM unnest(dk_dates) AS item(dk_date)
              WHERE (sqlc.narg('dk_date_from')::date IS NULL OR item.dk_date::date >= sqlc.narg('dk_date_from')::date)
                AND (sqlc.narg('dk_date_to')::date IS NULL OR item.dk_date::date <= sqlc.narg('dk_date_to')::date)
          )
      )
      AND (
          sqlc.narg('search')::text IS NULL
          OR LOWER(project_name) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR LOWER(bb_code) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          OR EXISTS (
              SELECT 1
              FROM unnest(indication_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(fixed_lenders) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
          OR EXISTS (
              SELECT 1
              FROM unnest(executing_agencies) AS item(label)
              WHERE LOWER(item.label) LIKE '%' || LOWER(sqlc.narg('search')::text) || '%'
          )
      )
)
SELECT
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS total_loan_usd,
    COALESCE(SUM(foreign_grant_usd), 0)::numeric AS total_grant_usd,
    COALESCE(SUM(counterpart_usd), 0)::numeric AS total_counterpart_usd
FROM filtered_projects;

-- name: ListBlueBookTopLevelExecutingAgencyGroups :many
WITH RECURSIVE latest_projects AS (
    SELECT
        bp.id,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
),
institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        i.name AS ancestor_name,
        i.short_name AS ancestor_short_name,
        i.level AS ancestor_level,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        parent.name AS ancestor_name,
        parent.short_name AS ancestor_short_name,
        parent.level AS ancestor_level,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
executing_agency_roots AS (
    SELECT DISTINCT
        lp.id AS bb_project_id,
        lp.foreign_loan_usd,
        CASE
            WHEN root.ancestor_level = 'Kementerian/Badan/Lembaga' THEN 'Kementerian/Lembaga'
            WHEN root.ancestor_level IN ('Pemerintah Daerah Tk. I', 'Pemerintah Daerah Tk. II') THEN 'Pemerintah Daerah'
            ELSE root.ancestor_level
        END::text AS group_label
    FROM latest_projects lp
    JOIN bb_project_institution bpi ON bpi.bb_project_id = lp.id
    JOIN LATERAL (
        SELECT
            ia.ancestor_id,
            ia.ancestor_name,
            ia.ancestor_short_name,
            ia.ancestor_level
        FROM institution_ancestors ia
        WHERE ia.institution_id = bpi.institution_id
          AND ia.parent_id IS NULL
        ORDER BY ia.depth DESC
        LIMIT 1
    ) root ON TRUE
    WHERE bpi.role = 'Executing Agency'
),
executing_agency_groups AS (
    SELECT DISTINCT
        bb_project_id,
        foreign_loan_usd,
        group_label
    FROM executing_agency_roots
)
SELECT
    ''::text AS id,
    group_label::text AS label,
    group_label::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM executing_agency_groups
GROUP BY group_label
ORDER BY project_count DESC, label ASC;

-- name: ListBlueBookTopLevelExecutingAgencies :many
WITH RECURSIVE latest_projects AS (
    SELECT
        bp.id,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
),
institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        i.name AS ancestor_name,
        i.short_name AS ancestor_short_name,
        i.level AS ancestor_level,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        parent.name AS ancestor_name,
        parent.short_name AS ancestor_short_name,
        parent.level AS ancestor_level,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
executing_agency_roots AS (
    SELECT DISTINCT
        lp.id AS bb_project_id,
        lp.foreign_loan_usd,
        root.ancestor_id::text AS id,
        COALESCE(root.ancestor_short_name, root.ancestor_name)::text AS label,
        root.ancestor_level::text AS level
    FROM latest_projects lp
    JOIN bb_project_institution bpi ON bpi.bb_project_id = lp.id
    JOIN LATERAL (
        SELECT
            ia.ancestor_id,
            ia.ancestor_name,
            ia.ancestor_short_name,
            ia.ancestor_level
        FROM institution_ancestors ia
        WHERE ia.institution_id = bpi.institution_id
          AND ia.parent_id IS NULL
        ORDER BY ia.depth DESC
        LIMIT 1
    ) root ON TRUE
    WHERE bpi.role = 'Executing Agency'
)
SELECT
    id,
    label,
    level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM executing_agency_roots
GROUP BY id, label, level
ORDER BY project_count DESC, label ASC
LIMIT 6;

-- name: ListDaftarKegiatanLenderTypes :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          WHERE gbp.bb_project_id = bp.id
      )
),
project_lender_types AS (
    SELECT
        lp.id AS bb_project_id,
        l.type::text AS lender_type,
        COALESCE(SUM(dfd.amount_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
    JOIN lender l ON l.id = dfd.lender_id
    WHERE dfd.lender_id IS NOT NULL
    GROUP BY lp.id, l.type
)
SELECT
    ''::text AS id,
    lender_type AS label,
    'Lender Type'::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_lender_types
GROUP BY lender_type
ORDER BY project_count DESC, label ASC;

-- name: ListDaftarKegiatanTopLenders :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          WHERE gbp.bb_project_id = bp.id
      )
),
project_lenders AS (
    SELECT
        lp.id AS bb_project_id,
        l.id AS lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_label,
        l.type::text AS lender_type,
        COALESCE(SUM(dfd.amount_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
    JOIN lender l ON l.id = dfd.lender_id
    WHERE dfd.lender_id IS NOT NULL
    GROUP BY lp.id, l.id, lender_label, l.type
)
SELECT
    lender_id::text AS id,
    lender_label AS label,
    lender_type AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_lenders
GROUP BY lender_id, lender_label, lender_type
ORDER BY project_count DESC, label ASC
LIMIT 6;

-- name: ListDaftarKegiatanTopLevelExecutingAgencies :many
WITH RECURSIVE latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          WHERE gbp.bb_project_id = bp.id
      )
),
institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        i.name AS ancestor_name,
        i.short_name AS ancestor_short_name,
        i.level AS ancestor_level,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        parent.name AS ancestor_name,
        parent.short_name AS ancestor_short_name,
        parent.level AS ancestor_level,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
project_agencies AS (
    SELECT
        lp.id AS bb_project_id,
        dp.id AS dk_project_id,
        dp.institution_id,
        COALESCE((
            SELECT SUM(dfd.amount_usd)
            FROM dk_financing_detail dfd
            WHERE dfd.dk_project_id = dp.id
        ), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN dk_project dp ON dp.id = dpg.dk_project_id
    WHERE dp.institution_id IS NOT NULL
),
executing_agency_roots AS (
    SELECT
        pa.bb_project_id,
        pa.dk_project_id,
        pa.foreign_loan_usd,
        root.ancestor_id::text AS id,
        COALESCE(root.ancestor_short_name, root.ancestor_name)::text AS label,
        root.ancestor_level::text AS level
    FROM project_agencies pa
    JOIN LATERAL (
        SELECT
            ia.ancestor_id,
            ia.ancestor_name,
            ia.ancestor_short_name,
            ia.ancestor_level
        FROM institution_ancestors ia
        WHERE ia.institution_id = pa.institution_id
          AND ia.parent_id IS NULL
        ORDER BY ia.depth DESC
        LIMIT 1
    ) root ON TRUE
)
SELECT
    id,
    label,
    level,
    COUNT(DISTINCT bb_project_id)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM executing_agency_roots
GROUP BY id, label, level
ORDER BY project_count DESC, label ASC
LIMIT 6;

-- name: ListDaftarKegiatanPrograms :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.program_title_id IS NOT NULL
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          WHERE gbp.bb_project_id = bp.id
      )
),
project_programs AS (
    SELECT
        lp.id AS bb_project_id,
        bp.program_title_id,
        pt.title::text AS program_title,
        COALESCE(SUM(dfd.amount_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN bb_project bp ON bp.id = lp.id
    JOIN program_title pt ON pt.id = bp.program_title_id
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    LEFT JOIN dk_financing_detail dfd ON dfd.dk_project_id = dpg.dk_project_id
    GROUP BY lp.id, bp.program_title_id, pt.title
)
SELECT
    program_title_id::text AS id,
    program_title AS label,
    'Program Title'::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_programs
GROUP BY program_title_id, program_title
ORDER BY project_count DESC, label ASC
LIMIT 8;

-- name: ListLoanAgreementLenderTypes :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
          WHERE gbp.bb_project_id = bp.id
      )
),
project_lender_types AS (
    SELECT
        lp.id AS bb_project_id,
        l.type::text AS lender_type,
        COALESCE(SUM(ladp.allocation_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    GROUP BY lp.id, l.type
)
SELECT
    ''::text AS id,
    lender_type AS label,
    'Lender Type'::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_lender_types
GROUP BY lender_type
ORDER BY project_count DESC, label ASC;

-- name: ListLoanAgreementTopLenders :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
          WHERE gbp.bb_project_id = bp.id
      )
),
project_lenders AS (
    SELECT
        lp.id AS bb_project_id,
        l.id AS lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_label,
        l.type::text AS lender_type,
        COALESCE(SUM(ladp.allocation_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
    JOIN lender l ON l.id = la.lender_id
    GROUP BY lp.id, l.id, lender_label, l.type
)
SELECT
    lender_id::text AS id,
    lender_label AS label,
    lender_type AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_lenders
GROUP BY lender_id, lender_label, lender_type
ORDER BY project_count DESC, label ASC
LIMIT 6;

-- name: ListLoanAgreementTopLevelExecutingAgencies :many
WITH RECURSIVE latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
          WHERE gbp.bb_project_id = bp.id
      )
),
institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        i.name AS ancestor_name,
        i.short_name AS ancestor_short_name,
        i.level AS ancestor_level,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        parent.name AS ancestor_name,
        parent.short_name AS ancestor_short_name,
        parent.level AS ancestor_level,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
project_agencies AS (
    SELECT
        lp.id AS bb_project_id,
        dp.id AS dk_project_id,
        dp.institution_id,
        COALESCE(SUM(ladp.allocation_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN dk_project dp ON dp.id = dpg.dk_project_id
    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dp.id
    JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
    WHERE dp.institution_id IS NOT NULL
    GROUP BY lp.id, dp.id, dp.institution_id
),
executing_agency_roots AS (
    SELECT
        pa.bb_project_id,
        pa.dk_project_id,
        pa.foreign_loan_usd,
        root.ancestor_id::text AS id,
        COALESCE(root.ancestor_short_name, root.ancestor_name)::text AS label,
        root.ancestor_level::text AS level
    FROM project_agencies pa
    JOIN LATERAL (
        SELECT
            ia.ancestor_id,
            ia.ancestor_name,
            ia.ancestor_short_name,
            ia.ancestor_level
        FROM institution_ancestors ia
        WHERE ia.institution_id = pa.institution_id
          AND ia.parent_id IS NULL
        ORDER BY ia.depth DESC
        LIMIT 1
    ) root ON TRUE
)
SELECT
    id,
    label,
    level,
    COUNT(DISTINCT bb_project_id)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM executing_agency_roots
GROUP BY id, label, level
ORDER BY project_count DESC, label ASC
LIMIT 6;

-- name: ListLoanAgreementPrograms :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.program_title_id IS NOT NULL
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
          JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
          WHERE gbp.bb_project_id = bp.id
      )
),
project_programs AS (
    SELECT
        lp.id AS bb_project_id,
        bp.program_title_id,
        pt.title::text AS program_title,
        COALESCE(SUM(ladp.allocation_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN bb_project bp ON bp.id = lp.id
    JOIN program_title pt ON pt.id = bp.program_title_id
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = gbp.gb_project_id
    JOIN loan_agreement_dk_project ladp ON ladp.dk_project_id = dpg.dk_project_id
                JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
    GROUP BY lp.id, bp.program_title_id, pt.title
)
SELECT
    program_title_id::text AS id,
    program_title AS label,
    'Program Title'::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_programs
GROUP BY program_title_id, program_title
ORDER BY project_count DESC, label ASC
LIMIT 8;

-- name: ListBlueBookPrograms :many
WITH latest_projects AS (
    SELECT
        bp.id,
        bp.program_title_id,
        pt.title::text AS program_title,
        CASE
            WHEN EXISTS (
                SELECT 1
                FROM gb_project_bb_project gbp
                WHERE gbp.bb_project_id = bp.id
            ) THEN COALESCE((
                SELECT SUM(gfs.loan_usd)
                FROM gb_project_bb_project gbp
                JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
                WHERE gbp.bb_project_id = bp.id
            ), 0)
            ELSE COALESCE((
                SELECT SUM(pc.amount_usd)
                FROM bb_project_cost pc
                WHERE pc.bb_project_id = bp.id
                  AND pc.funding_type = 'Foreign'
                  AND pc.funding_category = 'Loan'
            ), 0)
        END::numeric AS foreign_loan_usd
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    JOIN program_title pt ON pt.id = bp.program_title_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.program_title_id IS NOT NULL
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
)
SELECT
    program_title_id::text AS id,
    program_title::text AS label,
    'Program Title'::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM latest_projects
GROUP BY program_title_id, program_title
ORDER BY project_count DESC, label ASC
LIMIT 8;

-- name: ListGreenBookLenderTypes :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          WHERE gbp.bb_project_id = bp.id
      )
),
project_lender_types AS (
    SELECT
        lp.id AS bb_project_id,
        l.type::text AS lender_type,
        COALESCE(SUM(gfs.loan_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
    JOIN lender l ON l.id = gfs.lender_id
    GROUP BY lp.id, l.type
)
SELECT
    ''::text AS id,
    lender_type AS label,
    'Lender Type'::text AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_lender_types
GROUP BY lender_type
ORDER BY project_count DESC, label ASC;

-- name: ListGreenBookTopLenders :many
WITH latest_projects AS (
    SELECT bp.id
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          WHERE gbp.bb_project_id = bp.id
      )
),
project_lenders AS (
    SELECT
        lp.id AS bb_project_id,
        l.id AS lender_id,
        COALESCE(l.short_name, l.name)::text AS lender_label,
        l.type::text AS lender_type,
        COALESCE(SUM(gfs.loan_usd), 0)::numeric AS foreign_loan_usd
    FROM latest_projects lp
    JOIN gb_project_bb_project gbp ON gbp.bb_project_id = lp.id
    JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
    JOIN lender l ON l.id = gfs.lender_id
    GROUP BY lp.id, l.id, lender_label, l.type
)
SELECT
    lender_id::text AS id,
    lender_label AS label,
    lender_type AS level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM project_lenders
GROUP BY lender_id, lender_label, lender_type
ORDER BY project_count DESC, label ASC
LIMIT 6;

-- name: ListGreenBookTopLevelExecutingAgencies :many
WITH RECURSIVE latest_projects AS (
    SELECT
        bp.id,
        COALESCE((
            SELECT SUM(gfs.loan_usd)
            FROM gb_project_bb_project gbp
            JOIN gb_funding_source gfs ON gfs.gb_project_id = gbp.gb_project_id
            WHERE gbp.bb_project_id = bp.id
        ), 0)::numeric AS foreign_loan_usd
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.status = 'active'
      AND (COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0 OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[]))
      AND bp.id = (
          SELECT latest.id
          FROM bb_project latest
          JOIN blue_book latest_bb ON latest_bb.id = latest.blue_book_id
          WHERE latest.project_identity_id = bp.project_identity_id
            AND latest.status = 'active'
          ORDER BY latest_bb.revision_number DESC, COALESCE(latest_bb.revision_year, 0) DESC, latest_bb.created_at DESC
          LIMIT 1
      )
      AND EXISTS (
          SELECT 1
          FROM gb_project_bb_project gbp
          WHERE gbp.bb_project_id = bp.id
      )
),
institution_ancestors AS (
    SELECT
        i.id AS institution_id,
        i.id AS ancestor_id,
        i.parent_id,
        i.name AS ancestor_name,
        i.short_name AS ancestor_short_name,
        i.level AS ancestor_level,
        0::int AS depth
    FROM institution i
    UNION ALL
    SELECT
        ia.institution_id,
        parent.id AS ancestor_id,
        parent.parent_id,
        parent.name AS ancestor_name,
        parent.short_name AS ancestor_short_name,
        parent.level AS ancestor_level,
        ia.depth + 1 AS depth
    FROM institution_ancestors ia
    JOIN institution parent ON parent.id = ia.parent_id
),
executing_agency_roots AS (
    SELECT DISTINCT
        lp.id AS bb_project_id,
        lp.foreign_loan_usd,
        root.ancestor_id::text AS id,
        COALESCE(root.ancestor_short_name, root.ancestor_name)::text AS label,
        root.ancestor_level::text AS level
    FROM latest_projects lp
    JOIN bb_project_institution bpi ON bpi.bb_project_id = lp.id
    JOIN LATERAL (
        SELECT
            ia.ancestor_id,
            ia.ancestor_name,
            ia.ancestor_short_name,
            ia.ancestor_level
        FROM institution_ancestors ia
        WHERE ia.institution_id = bpi.institution_id
          AND ia.parent_id IS NULL
        ORDER BY ia.depth DESC
        LIMIT 1
    ) root ON TRUE
    WHERE bpi.role = 'Executing Agency'
)
SELECT
    id,
    label,
    level,
    COUNT(*)::int AS project_count,
    COALESCE(SUM(foreign_loan_usd), 0)::numeric AS foreign_loan_usd
FROM executing_agency_roots
GROUP BY id, label, level
ORDER BY project_count DESC, label ASC
LIMIT 6;
