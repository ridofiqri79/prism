-- ===== BLUE BOOK =====

-- name: ListBlueBooks :many
SELECT
    bb.id,
    bb.period_id,
    bb.replaces_blue_book_id,
    bb.publish_date,
    bb.revision_number,
    bb.revision_year,
    bb.status,
    bb.created_at,
    bb.updated_at,
    p.name AS period_name,
    p.year_start,
    p.year_end,
    (
        SELECT COUNT(*)
        FROM bb_project bp
        WHERE bp.blue_book_id = bb.id
    )::BIGINT AS project_count
FROM blue_book bb
JOIN period p ON p.id = bb.period_id
WHERE (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR bb.publish_date::text ILIKE '%' || sqlc.narg('search')::text || '%'
    OR bb.status ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (
    COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
    OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
)
AND (
    COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) = 0
    OR bb.status = ANY(sqlc.arg('statuses')::text[])
)
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'period' AND sqlc.arg('sort_order')::text = 'asc' THEN p.name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'period' AND sqlc.arg('sort_order')::text = 'desc' THEN p.name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'publish_date' AND sqlc.arg('sort_order')::text = 'asc' THEN bb.publish_date END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'publish_date' AND sqlc.arg('sort_order')::text = 'desc' THEN bb.publish_date END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'revision' AND sqlc.arg('sort_order')::text = 'asc' THEN bb.revision_number END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'revision' AND sqlc.arg('sort_order')::text = 'desc' THEN bb.revision_number END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'asc' THEN bb.status END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'status' AND sqlc.arg('sort_order')::text = 'desc' THEN bb.status END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_count' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT COUNT(*)
        FROM bb_project bp
        WHERE bp.blue_book_id = bb.id
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_count' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT COUNT(*)
        FROM bb_project bp
        WHERE bp.blue_book_id = bb.id
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN bb.created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN bb.created_at END DESC,
    bb.created_at DESC,
    bb.id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountBlueBooks :one
SELECT COUNT(*)
FROM blue_book bb
JOIN period p ON p.id = bb.period_id
WHERE (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR bb.publish_date::text ILIKE '%' || sqlc.narg('search')::text || '%'
    OR bb.status ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (
    COALESCE(cardinality(sqlc.arg('period_ids')::uuid[]), 0) = 0
    OR bb.period_id = ANY(sqlc.arg('period_ids')::uuid[])
)
AND (
    COALESCE(cardinality(sqlc.arg('statuses')::text[]), 0) = 0
    OR bb.status = ANY(sqlc.arg('statuses')::text[])
);

-- name: GetBlueBook :one
SELECT
    bb.id,
    bb.period_id,
    bb.replaces_blue_book_id,
    bb.publish_date,
    bb.revision_number,
    bb.revision_year,
    bb.status,
    bb.created_at,
    bb.updated_at,
    p.name AS period_name,
    p.year_start,
    p.year_end,
    (
        SELECT COUNT(*)
        FROM bb_project bp
        WHERE bp.blue_book_id = bb.id
    )::BIGINT AS project_count
FROM blue_book bb
JOIN period p ON p.id = bb.period_id
WHERE bb.id = $1;

-- name: CreateBlueBook :one
INSERT INTO blue_book (period_id, replaces_blue_book_id, publish_date, revision_number, revision_year, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateBlueBook :one
UPDATE blue_book
SET publish_date = $2,
    revision_number = $3,
    revision_year = $4,
    status = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetActiveBlueBookByPeriod :one
SELECT *
FROM blue_book
WHERE period_id = $1
  AND status = 'active'
ORDER BY revision_number DESC, created_at DESC
LIMIT 1;

-- name: CountBlueBooksByPeriodAndVersion :one
SELECT COUNT(*)
FROM blue_book
WHERE period_id = sqlc.arg('period_id')
  AND revision_number = sqlc.arg('revision_number')
  AND (
      (sqlc.arg('revision_year_valid')::BOOLEAN AND revision_year = sqlc.arg('revision_year')::INT)
      OR (NOT sqlc.arg('revision_year_valid')::BOOLEAN AND revision_year IS NULL)
  );

-- name: CountBlueBooksByPeriodAndVersionExcept :one
SELECT COUNT(*)
FROM blue_book
WHERE period_id = sqlc.arg('period_id')
  AND revision_number = sqlc.arg('revision_number')
  AND (
      (sqlc.arg('revision_year_valid')::BOOLEAN AND revision_year = sqlc.arg('revision_year')::INT)
      OR (NOT sqlc.arg('revision_year_valid')::BOOLEAN AND revision_year IS NULL)
  )
  AND id <> sqlc.arg('id');

-- name: CountAnyBBProjectsByBlueBook :one
SELECT COUNT(*)
FROM bb_project
WHERE blue_book_id = $1;

-- name: CountBlueBookRevisionsReplacing :one
SELECT COUNT(*)
FROM blue_book
WHERE replaces_blue_book_id = $1;

-- name: HardDeleteBlueBook :one
DELETE FROM blue_book
WHERE blue_book.id = $1
  AND NOT EXISTS (
      SELECT 1
      FROM bb_project bp
      WHERE bp.blue_book_id = blue_book.id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM blue_book child
      WHERE child.replaces_blue_book_id = blue_book.id
  )
RETURNING *;

-- ===== BB PROJECT =====

-- name: CreateProjectIdentity :one
INSERT INTO project_identity DEFAULT VALUES
RETURNING *;

-- name: GetProjectIdentity :one
SELECT *
FROM project_identity
WHERE id = $1;

-- name: ListBBProjectsByBlueBook :many
SELECT *
FROM bb_project
WHERE blue_book_id = sqlc.arg('blue_book_id')
  AND status = 'active'
  AND (
      sqlc.narg('search')::text IS NULL
      OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
      OR EXISTS (
          SELECT 1
          FROM bb_project_institution bpi
          JOIN institution i ON i.id = bpi.institution_id
          WHERE bpi.bb_project_id = bb_project.id
            AND bpi.role = 'Executing Agency'
            AND (
                i.name ILIKE '%' || sqlc.narg('search')::text || '%'
                OR COALESCE(i.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
            )
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM bb_project_institution bpi
          WHERE bpi.bb_project_id = bb_project.id
            AND bpi.role = 'Executing Agency'
            AND bpi.institution_id = ANY(sqlc.arg('executing_agency_ids')::uuid[])
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('location_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM bb_project_location bpl
          WHERE bpl.bb_project_id = bb_project.id
            AND bpl.region_id = ANY(sqlc.arg('location_ids')::uuid[])
      )
  )
ORDER BY
    CASE WHEN sqlc.arg('sort_field')::text = 'bb_code' AND sqlc.arg('sort_order')::text = 'asc' THEN bb_code END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'bb_code' AND sqlc.arg('sort_order')::text = 'desc' THEN bb_code END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_name' AND sqlc.arg('sort_order')::text = 'asc' THEN project_name END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'project_name' AND sqlc.arg('sort_order')::text = 'desc' THEN project_name END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'executing_agency' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT string_agg(COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name))
        FROM bb_project_institution bpi
        JOIN institution i ON i.id = bpi.institution_id
        WHERE bpi.bb_project_id = bb_project.id
          AND bpi.role = 'Executing Agency'
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'executing_agency' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT string_agg(COALESCE(i.short_name, i.name), ', ' ORDER BY COALESCE(i.short_name, i.name))
        FROM bb_project_institution bpi
        JOIN institution i ON i.id = bpi.institution_id
        WHERE bpi.bb_project_id = bb_project.id
          AND bpi.role = 'Executing Agency'
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'location' AND sqlc.arg('sort_order')::text = 'asc' THEN (
        SELECT string_agg(r.name, ', ' ORDER BY r.name)
        FROM bb_project_location bpl
        JOIN region r ON r.id = bpl.region_id
        WHERE bpl.bb_project_id = bb_project.id
    ) END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'location' AND sqlc.arg('sort_order')::text = 'desc' THEN (
        SELECT string_agg(r.name, ', ' ORDER BY r.name)
        FROM bb_project_location bpl
        JOIN region r ON r.id = bpl.region_id
        WHERE bpl.bb_project_id = bb_project.id
    ) END DESC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('sort_field')::text = 'created_at' AND sqlc.arg('sort_order')::text = 'desc' THEN created_at END DESC,
    bb_code ASC,
    id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountBBProjectsByBlueBook :one
SELECT COUNT(*)
FROM bb_project
WHERE blue_book_id = sqlc.arg('blue_book_id')
  AND status = 'active'
  AND (
      sqlc.narg('search')::text IS NULL
      OR project_name ILIKE '%' || sqlc.narg('search')::text || '%'
      OR EXISTS (
          SELECT 1
          FROM bb_project_institution bpi
          JOIN institution i ON i.id = bpi.institution_id
          WHERE bpi.bb_project_id = bb_project.id
            AND bpi.role = 'Executing Agency'
            AND (
                i.name ILIKE '%' || sqlc.narg('search')::text || '%'
                OR COALESCE(i.short_name, '') ILIKE '%' || sqlc.narg('search')::text || '%'
            )
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('executing_agency_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM bb_project_institution bpi
          WHERE bpi.bb_project_id = bb_project.id
            AND bpi.role = 'Executing Agency'
            AND bpi.institution_id = ANY(sqlc.arg('executing_agency_ids')::uuid[])
      )
  )
  AND (
      COALESCE(cardinality(sqlc.arg('location_ids')::uuid[]), 0) = 0
      OR EXISTS (
          SELECT 1
          FROM bb_project_location bpl
          WHERE bpl.bb_project_id = bb_project.id
            AND bpl.region_id = ANY(sqlc.arg('location_ids')::uuid[])
      )
  );

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

-- name: GetBBProjectByBlueBookAndCode :one
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND LOWER(bb_code) = LOWER($2)
LIMIT 1;

-- name: FindPreviousBBProjectByCodeForBlueBook :one
SELECT bp.*
FROM bb_project bp
JOIN blue_book source_bb ON source_bb.id = bp.blue_book_id
JOIN blue_book target_bb ON target_bb.id = $1
WHERE LOWER(bp.bb_code) = LOWER($2)
  AND source_bb.period_id = target_bb.period_id
  AND source_bb.id <> target_bb.id
  AND (
      source_bb.revision_number < target_bb.revision_number
      OR (
          source_bb.revision_number = target_bb.revision_number
          AND source_bb.created_at < target_bb.created_at
      )
  )
ORDER BY source_bb.revision_number DESC, source_bb.created_at DESC
LIMIT 1;

-- name: CreateBBProject :one
INSERT INTO bb_project (
    blue_book_id,
    project_identity_id,
    program_title_id,
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

-- name: ListBBProjectsForClone :many
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND status = 'active'
ORDER BY bb_code ASC;

-- name: ListBBProjectsForCloneByIDs :many
SELECT *
FROM bb_project
WHERE blue_book_id = $1
  AND status = 'active'
  AND id = ANY(sqlc.arg('project_ids')::uuid[])
ORDER BY bb_code ASC;

-- name: GetLatestBBProjectByIdentity :one
SELECT bp.*
FROM bb_project bp
JOIN blue_book bb ON bb.id = bp.blue_book_id
WHERE bp.project_identity_id = $1
  AND bp.status = 'active'
ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC
LIMIT 1;

-- name: GetLatestBBProjectByProject :one
SELECT latest.*
FROM bb_project current_project
JOIN LATERAL (
    SELECT bp.*
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    WHERE bp.project_identity_id = current_project.project_identity_id
      AND bp.status = 'active'
    ORDER BY bb.revision_number DESC, COALESCE(bb.revision_year, 0) DESC, bb.created_at DESC
    LIMIT 1
) latest ON TRUE
WHERE current_project.id = $1;

-- name: ListBBProjectHistoryByIdentity :many
SELECT
    bp.id,
    bp.project_identity_id,
    bp.blue_book_id,
    bp.bb_code,
    bp.project_name,
    p.name AS period_name,
    bb.revision_number,
    bb.revision_year,
    bb.status AS book_status,
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
        FROM gb_project_bb_project gbp
        WHERE gbp.bb_project_id = bp.id
    )::boolean AS used_by_downstream
FROM bb_project bp
JOIN blue_book bb ON bb.id = bp.blue_book_id
JOIN period p ON p.id = bb.period_id
WHERE bp.project_identity_id = $1
ORDER BY bb.revision_number ASC, COALESCE(bb.revision_year, 0) ASC, bb.created_at ASC;

-- name: ListBBProjectHistoryByProject :many
SELECT history.*
FROM bb_project current_project
JOIN LATERAL (
    SELECT
        bp.id,
        bp.project_identity_id,
        bp.blue_book_id,
        bp.bb_code,
        bp.project_name,
        p.name AS period_name,
        bb.revision_number,
        bb.revision_year,
        bb.status AS book_status,
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
            FROM gb_project_bb_project gbp
            WHERE gbp.bb_project_id = bp.id
        )::boolean AS used_by_downstream
    FROM bb_project bp
    JOIN blue_book bb ON bb.id = bp.blue_book_id
    JOIN period p ON p.id = bb.period_id
    WHERE bp.project_identity_id = current_project.project_identity_id
    ORDER BY bb.revision_number ASC, COALESCE(bb.revision_year, 0) ASC, bb.created_at ASC
) history ON TRUE
WHERE current_project.id = $1;

-- name: CloneBBProjectInstitutions :exec
INSERT INTO bb_project_institution (bb_project_id, institution_id, role)
SELECT $2, source.institution_id, source.role
FROM bb_project_institution source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectBappenasPartners :exec
INSERT INTO bb_project_bappenas_partner (bb_project_id, bappenas_partner_id)
SELECT $2, source.bappenas_partner_id
FROM bb_project_bappenas_partner source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectLocations :exec
INSERT INTO bb_project_location (bb_project_id, region_id)
SELECT $2, source.region_id
FROM bb_project_location source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectNationalPriorities :exec
INSERT INTO bb_project_national_priority (bb_project_id, national_priority_id)
SELECT $2, source.national_priority_id
FROM bb_project_national_priority source
WHERE source.bb_project_id = $1
ON CONFLICT DO NOTHING;

-- name: CloneBBProjectCosts :exec
INSERT INTO bb_project_cost (bb_project_id, funding_type, funding_category, amount_usd)
SELECT $2, source.funding_type, source.funding_category, source.amount_usd
FROM bb_project_cost source
WHERE source.bb_project_id = $1;

-- name: CloneLenderIndications :exec
INSERT INTO lender_indication (bb_project_id, lender_id, remarks)
SELECT $2, source.lender_id, source.remarks
FROM lender_indication source
WHERE source.bb_project_id = $1;

-- name: CloneLoIs :exec
INSERT INTO loi (bb_project_id, lender_id, subject, date, letter_number)
SELECT $2, source.lender_id, source.subject, source.date, source.letter_number
FROM loi source
WHERE source.bb_project_id = $1;

-- name: UpdateBBProject :one
UPDATE bb_project
SET program_title_id = $2,
    project_name = $3,
    duration = $4,
    objective = $5,
    scope_of_work = $6,
    outputs = $7,
    outcomes = $8,
    updated_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING *;

-- name: ListBBProjectDeletionDependencies :many
WITH related_gb AS (
    SELECT DISTINCT
        gp.id,
        gp.gb_code,
        gp.project_name,
        gb.publish_year,
        gb.revision_number
    FROM gb_project_bb_project gbp
    JOIN gb_project gp ON gp.id = gbp.gb_project_id
    JOIN green_book gb ON gb.id = gp.green_book_id
    WHERE gbp.bb_project_id = $1
),
related_dk AS (
    SELECT DISTINCT
        dp.id,
        dk.subject,
        dk.letter_number,
        dk.date,
        rg.gb_code
    FROM related_gb rg
    JOIN dk_project_gb_project dpg ON dpg.gb_project_id = rg.id
    JOIN dk_project dp ON dp.id = dpg.dk_project_id
    JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
),
related_la AS (
    SELECT DISTINCT
        la.id,
        la.loan_code,
        rd.subject,
        rd.letter_number,
        rd.gb_code
    FROM related_dk rd
    JOIN loan_agreement la ON la.dk_project_id = rd.id
),
related_monitoring AS (
    SELECT DISTINCT
        md.id,
        md.budget_year,
        md.quarter,
        rla.loan_code,
        rla.subject,
        rla.letter_number,
        rla.gb_code
    FROM related_la rla
    JOIN monitoring_disbursement md ON md.loan_agreement_id = rla.id
)
SELECT
    'green_book_project'::text AS relation_type,
    id AS relation_id,
    format('%s - %s', gb_code, project_name)::text AS relation_label,
    format('Green Book %s Revisi %s', publish_year, revision_number)::text AS relation_path
FROM related_gb
UNION ALL
SELECT
    'daftar_kegiatan_project'::text AS relation_type,
    id AS relation_id,
    COALESCE(letter_number, subject)::text AS relation_label,
    format('Green Book Project %s -> Daftar Kegiatan %s', gb_code, COALESCE(letter_number, subject))::text AS relation_path
FROM related_dk
UNION ALL
SELECT
    'loan_agreement'::text AS relation_type,
    id AS relation_id,
    loan_code::text AS relation_label,
    format('Green Book Project %s -> Daftar Kegiatan %s -> Loan Agreement %s', gb_code, COALESCE(letter_number, subject), loan_code)::text AS relation_path
FROM related_la
UNION ALL
SELECT
    'monitoring_disbursement'::text AS relation_type,
    id AS relation_id,
    format('%s %s', budget_year, quarter)::text AS relation_label,
    format('Green Book Project %s -> Daftar Kegiatan %s -> Loan Agreement %s -> Monitoring %s %s', gb_code, COALESCE(letter_number, subject), loan_code, budget_year, quarter)::text AS relation_path
FROM related_monitoring
ORDER BY relation_type, relation_label;

-- name: HardDeleteBBProject :one
DELETE FROM bb_project bp
WHERE bp.blue_book_id = $1
  AND bp.id = $2
  AND NOT EXISTS (
      SELECT 1
      FROM gb_project_bb_project gbp
      WHERE gbp.bb_project_id = bp.id
  )
RETURNING *;

-- name: DeleteOrphanProjectIdentity :exec
DELETE FROM project_identity pi
WHERE pi.id = $1
  AND NOT EXISTS (
      SELECT 1
      FROM bb_project bp
      WHERE bp.project_identity_id = pi.id
  );

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

-- name: DeleteBBProjectInstitution :exec
DELETE FROM bb_project_institution
WHERE bb_project_id = $1
  AND institution_id = $2
  AND role = $3;

-- ===== BB BAPPENAS PARTNERS =====

-- name: GetBBProjectBappenasPartners :many
SELECT bp.*
FROM bb_project_bappenas_partner bpbp
JOIN bappenas_partner bp ON bp.id = bpbp.bappenas_partner_id
WHERE bpbp.bb_project_id = $1
ORDER BY bp.name;

-- name: AddBBProjectBappenasPartner :exec
INSERT INTO bb_project_bappenas_partner (bb_project_id, bappenas_partner_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectBappenasPartners :exec
DELETE FROM bb_project_bappenas_partner
WHERE bb_project_id = $1;

-- name: DeleteBBProjectBappenasPartner :exec
DELETE FROM bb_project_bappenas_partner
WHERE bb_project_id = $1
  AND bappenas_partner_id = $2;

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

-- name: DeleteBBProjectLocation :exec
DELETE FROM bb_project_location
WHERE bb_project_id = $1
  AND region_id = $2;

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

-- name: DeleteBBProjectNationalPriority :exec
DELETE FROM bb_project_national_priority
WHERE bb_project_id = $1
  AND national_priority_id = $2;

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

-- name: UpdateBBProjectCost :one
UPDATE bb_project_cost
SET funding_type = $3,
    funding_category = $4,
    amount_usd = $5,
    updated_at = NOW()
WHERE id = $1
  AND bb_project_id = $2
RETURNING *;

-- name: DeleteBBProjectCosts :exec
DELETE FROM bb_project_cost
WHERE bb_project_id = $1;

-- name: DeleteBBProjectCost :exec
DELETE FROM bb_project_cost
WHERE id = $1
  AND bb_project_id = $2;

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

-- name: UpdateLenderIndication :one
UPDATE lender_indication
SET lender_id = $3,
    remarks = $4,
    updated_at = NOW()
WHERE id = $1
  AND bb_project_id = $2
RETURNING *;

-- name: DeleteLenderIndications :exec
DELETE FROM lender_indication
WHERE bb_project_id = $1;

-- name: DeleteLenderIndication :exec
DELETE FROM lender_indication
WHERE id = $1
  AND bb_project_id = $2;

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
