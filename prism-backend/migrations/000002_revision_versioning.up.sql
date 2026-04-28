CREATE TABLE IF NOT EXISTS project_identity (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gb_project_identity (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE blue_book
    ADD COLUMN IF NOT EXISTS replaces_blue_book_id UUID REFERENCES blue_book(id);

ALTER TABLE green_book
    ADD COLUMN IF NOT EXISTS replaces_green_book_id UUID REFERENCES green_book(id);

ALTER TABLE bb_project
    ADD COLUMN IF NOT EXISTS project_identity_id UUID REFERENCES project_identity(id);

DO $$
DECLARE
    r RECORD;
    new_identity UUID;
BEGIN
    FOR r IN SELECT id FROM bb_project WHERE project_identity_id IS NULL LOOP
        INSERT INTO project_identity DEFAULT VALUES RETURNING id INTO new_identity;
        UPDATE bb_project SET project_identity_id = new_identity WHERE id = r.id;
    END LOOP;
END $$;

ALTER TABLE bb_project
    ALTER COLUMN project_identity_id SET NOT NULL;

ALTER TABLE gb_project
    ADD COLUMN IF NOT EXISTS gb_project_identity_id UUID REFERENCES gb_project_identity(id);

DO $$
DECLARE
    r RECORD;
    new_identity UUID;
BEGIN
    FOR r IN SELECT id FROM gb_project WHERE gb_project_identity_id IS NULL LOOP
        INSERT INTO gb_project_identity DEFAULT VALUES RETURNING id INTO new_identity;
        UPDATE gb_project SET gb_project_identity_id = new_identity WHERE id = r.id;
    END LOOP;
END $$;

ALTER TABLE gb_project
    ALTER COLUMN gb_project_identity_id SET NOT NULL;

ALTER TABLE bb_project DROP CONSTRAINT IF EXISTS bb_project_bb_code_key;
ALTER TABLE gb_project DROP CONSTRAINT IF EXISTS gb_project_gb_code_key;

ALTER TABLE bb_project
    ADD CONSTRAINT bb_project_blue_book_id_bb_code_key UNIQUE (blue_book_id, bb_code);

ALTER TABLE gb_project
    ADD CONSTRAINT gb_project_green_book_id_gb_code_key UNIQUE (green_book_id, gb_code);

CREATE INDEX IF NOT EXISTS idx_bb_project_identity ON bb_project(project_identity_id);
CREATE INDEX IF NOT EXISTS idx_bb_project_book_code ON bb_project(blue_book_id, bb_code);
CREATE INDEX IF NOT EXISTS idx_gb_project_identity ON gb_project(gb_project_identity_id);
CREATE INDEX IF NOT EXISTS idx_gb_project_book_code ON gb_project(green_book_id, gb_code);
