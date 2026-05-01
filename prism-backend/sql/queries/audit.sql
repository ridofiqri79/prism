-- ===== PROJECT AUDIT RAIL =====

-- name: ListBBProjectAuditEntries :many
WITH relevant_audit AS (
    SELECT al.*
    FROM audit_log al
    WHERE (
        al.table_name IN (
            'bb_project',
            'bb_project_institution',
            'bb_project_bappenas_partner',
            'bb_project_location',
            'bb_project_national_priority'
        )
        AND al.record_id = sqlc.arg('project_id')::uuid
    )
    OR (
        al.table_name IN ('bb_project_cost', 'lender_indication', 'loi')
        AND COALESCE(al.new_data ->> 'bb_project_id', al.old_data ->> 'bb_project_id') = sqlc.arg('project_id')::text
    )
)
SELECT
    al.id,
    al.table_name,
    al.record_id,
    al.action,
    al.changed_by,
    COALESCE(u.username, 'Sistem')::text AS changed_by_username,
    al.changed_at,
    COALESCE(diff.changed_fields, ARRAY[]::text[])::text[] AS changed_fields
FROM relevant_audit al
LEFT JOIN app_user u ON u.id = al.changed_by
LEFT JOIN LATERAL (
    SELECT ARRAY_AGG(keys.key ORDER BY keys.key)::text[] AS changed_fields
    FROM jsonb_object_keys(COALESCE(al.new_data, al.old_data, '{}'::jsonb)) AS keys(key)
    WHERE keys.key NOT IN ('id', 'created_at', 'updated_at')
      AND (
          al.action <> 'UPDATE'
          OR (al.old_data -> keys.key) IS DISTINCT FROM (al.new_data -> keys.key)
      )
) diff ON TRUE
ORDER BY al.changed_at DESC, al.id DESC
LIMIT 200;

-- name: ListGBProjectAuditEntries :many
WITH project_activities AS (
    SELECT ga.id::text AS id
    FROM gb_activity ga
    WHERE ga.gb_project_id = sqlc.arg('project_id')::uuid
    UNION
    SELECT COALESCE(al.new_data ->> 'id', al.old_data ->> 'id') AS id
    FROM audit_log al
    WHERE al.table_name = 'gb_activity'
      AND COALESCE(al.new_data ->> 'gb_project_id', al.old_data ->> 'gb_project_id') = sqlc.arg('project_id')::text
),
relevant_audit AS (
    SELECT al.*
    FROM audit_log al
    WHERE (
        al.table_name IN (
            'gb_project',
            'gb_project_bb_project',
            'gb_project_bappenas_partner',
            'gb_project_institution',
            'gb_project_location'
        )
        AND al.record_id = sqlc.arg('project_id')::uuid
    )
    OR (
        al.table_name IN ('gb_activity', 'gb_funding_source', 'gb_disbursement_plan')
        AND COALESCE(al.new_data ->> 'gb_project_id', al.old_data ->> 'gb_project_id') = sqlc.arg('project_id')::text
    )
    OR (
        al.table_name = 'gb_funding_allocation'
        AND COALESCE(al.new_data ->> 'gb_activity_id', al.old_data ->> 'gb_activity_id') IN (
            SELECT pa.id
            FROM project_activities pa
            WHERE pa.id IS NOT NULL
        )
    )
)
SELECT
    al.id,
    al.table_name,
    al.record_id,
    al.action,
    al.changed_by,
    COALESCE(u.username, 'Sistem')::text AS changed_by_username,
    al.changed_at,
    COALESCE(diff.changed_fields, ARRAY[]::text[])::text[] AS changed_fields
FROM relevant_audit al
LEFT JOIN app_user u ON u.id = al.changed_by
LEFT JOIN LATERAL (
    SELECT ARRAY_AGG(keys.key ORDER BY keys.key)::text[] AS changed_fields
    FROM jsonb_object_keys(COALESCE(al.new_data, al.old_data, '{}'::jsonb)) AS keys(key)
    WHERE keys.key NOT IN ('id', 'created_at', 'updated_at')
      AND (
          al.action <> 'UPDATE'
          OR (al.old_data -> keys.key) IS DISTINCT FROM (al.new_data -> keys.key)
      )
) diff ON TRUE
ORDER BY al.changed_at DESC, al.id DESC
LIMIT 200;
