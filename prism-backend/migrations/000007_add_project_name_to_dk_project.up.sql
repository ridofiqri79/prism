ALTER TABLE dk_project
    ADD COLUMN project_name VARCHAR(500);

UPDATE dk_project dp
SET project_name = COALESCE(
    (
        SELECT gp.project_name
        FROM dk_project_gb_project dpg
        JOIN gb_project gp ON gp.id = dpg.gb_project_id
        WHERE dpg.dk_project_id = dp.id
        ORDER BY gp.gb_code ASC, gp.created_at ASC
        LIMIT 1
    ),
    NULLIF(BTRIM(dp.objectives), ''),
    'Proyek Daftar Kegiatan'
);

ALTER TABLE dk_project
    ALTER COLUMN project_name SET NOT NULL;
