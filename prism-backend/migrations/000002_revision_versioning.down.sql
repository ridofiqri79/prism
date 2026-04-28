DROP INDEX IF EXISTS idx_gb_project_book_code;
DROP INDEX IF EXISTS idx_gb_project_identity;
DROP INDEX IF EXISTS idx_bb_project_book_code;
DROP INDEX IF EXISTS idx_bb_project_identity;

ALTER TABLE gb_project DROP CONSTRAINT IF EXISTS gb_project_green_book_id_gb_code_key;
ALTER TABLE bb_project DROP CONSTRAINT IF EXISTS bb_project_blue_book_id_bb_code_key;

ALTER TABLE gb_project
    ADD CONSTRAINT gb_project_gb_code_key UNIQUE (gb_code);

ALTER TABLE bb_project
    ADD CONSTRAINT bb_project_bb_code_key UNIQUE (bb_code);

ALTER TABLE gb_project DROP COLUMN IF EXISTS gb_project_identity_id;
ALTER TABLE bb_project DROP COLUMN IF EXISTS project_identity_id;
ALTER TABLE green_book DROP COLUMN IF EXISTS replaces_green_book_id;
ALTER TABLE blue_book DROP COLUMN IF EXISTS replaces_blue_book_id;

DROP TABLE IF EXISTS gb_project_identity;
DROP TABLE IF EXISTS project_identity;
