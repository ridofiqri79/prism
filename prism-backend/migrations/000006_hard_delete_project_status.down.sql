ALTER TABLE gb_project
    DROP CONSTRAINT IF EXISTS gb_project_status_check;

ALTER TABLE gb_project
    ADD CONSTRAINT gb_project_status_check
    CHECK (status IN ('active', 'deleted'));

ALTER TABLE bb_project
    DROP CONSTRAINT IF EXISTS bb_project_status_check;

ALTER TABLE bb_project
    ADD CONSTRAINT bb_project_status_check
    CHECK (status IN ('active', 'deleted'));
