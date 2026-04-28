-- ============================================================
-- PRISM — Project Loan Integrated Monitoring System
-- SQL DDL (PostgreSQL)
-- ============================================================

-- ============================================================
-- EXTENSIONS
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- ============================================================
-- MASTER DATA
-- ============================================================

CREATE TABLE country (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    code        CHAR(3) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE lender (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    country_id  UUID REFERENCES country(id),           -- NULL untuk Multilateral
    name        VARCHAR(255) NOT NULL,
    short_name  VARCHAR(100),
    type        VARCHAR(20) NOT NULL CHECK (type IN ('Bilateral', 'Multilateral', 'KSA')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_lender_negara CHECK (
        (type IN ('Bilateral', 'KSA') AND country_id IS NOT NULL) OR
        (type = 'Multilateral' AND country_id IS NULL)
    )
);

CREATE TABLE institution (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id   UUID REFERENCES institution(id),       -- NULL untuk Kementerian (level parent)
    name        VARCHAR(255) NOT NULL,
    short_name  VARCHAR(100),    
    level       VARCHAR(50) NOT NULL CHECK (level IN ('Kementerian/Badan/Lembaga', 'Eselon I', 'BUMN', 'Pemerintah Daerah', 'BUMD', 'Lainnya')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE region (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code         VARCHAR(10) NOT NULL UNIQUE,
    name         VARCHAR(255) NOT NULL,
    type         VARCHAR(20) NOT NULL CHECK (type IN ('COUNTRY', 'PROVINCE', 'CITY')),
    parent_code  VARCHAR(10) REFERENCES region(code),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE program_title (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id   UUID REFERENCES program_title(id),     -- NULL untuk Parent Program Title
    title       VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bappenas_partner (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id   UUID REFERENCES bappenas_partner(id),  -- NULL untuk Eselon I
    name        VARCHAR(255) NOT NULL,
    level       VARCHAR(20) NOT NULL CHECK (level IN ('Eselon I', 'Eselon II')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE period (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,                 -- contoh: "2025-2029"
    year_start  INT NOT NULL,
    year_end    INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_period_years CHECK (year_end > year_start)
);

CREATE TABLE national_priority (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    period_id   UUID NOT NULL REFERENCES period(id),
    title       VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================================
-- BLUE BOOK
-- ============================================================

CREATE TABLE blue_book (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    period_id        UUID NOT NULL REFERENCES period(id),
    replaces_blue_book_id UUID REFERENCES blue_book(id),
    publish_date     DATE NOT NULL,
    revision_number  INT NOT NULL DEFAULT 0,
    revision_year    INT,                              -- NULL untuk versi awal
    status           VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'superseded')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_identity (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bb_project (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    blue_book_id         UUID NOT NULL REFERENCES blue_book(id),
    project_identity_id  UUID NOT NULL REFERENCES project_identity(id),
    program_title_id     UUID REFERENCES program_title(id),
    bappenas_partner_id  UUID REFERENCES bappenas_partner(id), -- Eselon II; Eselon I diturunkan dari hierarki
    bb_code              VARCHAR(50) NOT NULL,
    project_name         VARCHAR(500) NOT NULL,
    duration             VARCHAR(100),
    objective            TEXT,
    scope_of_work        TEXT,
    outputs              TEXT,
    outcomes             TEXT,
    status               VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deleted')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blue_book_id, bb_code)
);

-- EA & IA Blue Book (multi-select, shared Institution)
CREATE TABLE bb_project_institution (
    bb_project_id  UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    institution_id UUID NOT NULL REFERENCES institution(id),
    role           VARCHAR(30) NOT NULL CHECK (role IN ('Executing Agency', 'Implementing Agency')),
    PRIMARY KEY (bb_project_id, institution_id, role)
);

-- Location Blue Book (multi-select)
CREATE TABLE bb_project_location (
    bb_project_id  UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    region_id      UUID NOT NULL REFERENCES region(id),
    PRIMARY KEY (bb_project_id, region_id)
);

-- National Priority Blue Book (multi-select)
CREATE TABLE bb_project_national_priority (
    bb_project_id        UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    national_priority_id UUID NOT NULL REFERENCES national_priority(id),
    PRIMARY KEY (bb_project_id, national_priority_id)
);

-- Project Cost Blue Book
CREATE TABLE bb_project_cost (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bb_project_id    UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    funding_type     VARCHAR(20) NOT NULL CHECK (funding_type IN ('Foreign', 'Counterpart')),
    -- Foreign: 'Loan' | 'Grant'
    -- Counterpart: 'Central Government' | 'Regional Government' | 'State-Owned Enterprise' | 'Others'
    funding_category VARCHAR(50) NOT NULL,
    amount_usd       NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Lender Indication (bisa lebih dari satu per BB project)
CREATE TABLE lender_indication (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bb_project_id  UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    lender_id      UUID NOT NULL REFERENCES lender(id),
    remarks        TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Letter of Intent
CREATE TABLE loi (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bb_project_id  UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    lender_id      UUID NOT NULL REFERENCES lender(id),
    subject        VARCHAR(500) NOT NULL,
    date           DATE NOT NULL,
    letter_number  VARCHAR(100),                       -- opsional
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================================
-- GREEN BOOK
-- ============================================================

CREATE TABLE green_book (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    publish_year     INT NOT NULL,
    replaces_green_book_id UUID REFERENCES green_book(id),
    revision_number  INT NOT NULL DEFAULT 0,
    status           VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'superseded')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gb_project_identity (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gb_project (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    green_book_id    UUID NOT NULL REFERENCES green_book(id),
    gb_project_identity_id UUID NOT NULL REFERENCES gb_project_identity(id),
    program_title_id UUID REFERENCES program_title(id),
    gb_code          VARCHAR(50) NOT NULL,
    project_name     VARCHAR(500) NOT NULL,
    duration         VARCHAR(100),
    objective        TEXT,
    scope_of_project TEXT,
    status           VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deleted')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (green_book_id, gb_code)
);

-- Relasi BB Project <-> GB Project (many-to-many)
CREATE TABLE gb_project_bb_project (
    gb_project_id  UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    bb_project_id  UUID NOT NULL REFERENCES bb_project(id) ON DELETE CASCADE,
    PRIMARY KEY (gb_project_id, bb_project_id)
);

-- EA & IA Green Book (multi-select, shared Institution)
CREATE TABLE gb_project_institution (
    gb_project_id  UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    institution_id UUID NOT NULL REFERENCES institution(id),
    role           VARCHAR(30) NOT NULL CHECK (role IN ('Executing Agency', 'Implementing Agency')),
    PRIMARY KEY (gb_project_id, institution_id, role)
);

-- Location Green Book (multi-select)
CREATE TABLE gb_project_location (
    gb_project_id  UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    region_id      UUID NOT NULL REFERENCES region(id),
    PRIMARY KEY (gb_project_id, region_id)
);

-- Activities Green Book
CREATE TABLE gb_activity (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gb_project_id           UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    activity_name           VARCHAR(500) NOT NULL,
    implementation_location TEXT,                      -- teks bebas, bukan entitas Region
    piu                     VARCHAR(255),              -- Project Implementation Units
    sort_order              INT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Funding Source Green Book (cofinancing -- per lender)
CREATE TABLE gb_funding_source (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gb_project_id  UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    lender_id      UUID NOT NULL REFERENCES lender(id),
    institution_id UUID REFERENCES institution(id),    -- Implementing Agency per baris
    loan_usd       NUMERIC(20, 2) NOT NULL DEFAULT 0,
    grant_usd      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    local_usd      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Disbursement Plan Green Book (total proyek, bukan per lender)
CREATE TABLE gb_disbursement_plan (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gb_project_id  UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    year           INT NOT NULL,
    amount_usd     NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (gb_project_id, year)
);

-- Funding Allocation Green Book (mengacu ke gb_activity)
CREATE TABLE gb_funding_allocation (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gb_activity_id UUID NOT NULL REFERENCES gb_activity(id) ON DELETE CASCADE,
    services       NUMERIC(20, 2) NOT NULL DEFAULT 0,
    constructions  NUMERIC(20, 2) NOT NULL DEFAULT 0,
    goods          NUMERIC(20, 2) NOT NULL DEFAULT 0,
    trainings      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    other          NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================================
-- DAFTAR KEGIATAN
-- ============================================================

CREATE TABLE daftar_kegiatan (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    letter_number VARCHAR(100),                         -- opsional
    subject       VARCHAR(500) NOT NULL,
    date          DATE NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE dk_project (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dk_id            UUID NOT NULL REFERENCES daftar_kegiatan(id),
    program_title_id UUID REFERENCES program_title(id),
    institution_id   UUID REFERENCES institution(id),  -- Executing Agency
    duration         VARCHAR(100),
    objectives       TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Relasi DK Project <-> GB Project (many-to-many)
CREATE TABLE dk_project_gb_project (
    dk_project_id  UUID NOT NULL REFERENCES dk_project(id) ON DELETE CASCADE,
    gb_project_id  UUID NOT NULL REFERENCES gb_project(id) ON DELETE CASCADE,
    PRIMARY KEY (dk_project_id, gb_project_id)
);

-- Location Daftar Kegiatan (multi-select)
CREATE TABLE dk_project_location (
    dk_project_id  UUID NOT NULL REFERENCES dk_project(id) ON DELETE CASCADE,
    region_id      UUID NOT NULL REFERENCES region(id),
    PRIMARY KEY (dk_project_id, region_id)
);

-- Financing Detail Daftar Kegiatan (per lender)
-- Mendukung multi-currency: simpan nilai original + ekuivalen USD
CREATE TABLE dk_financing_detail (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dk_project_id        UUID NOT NULL REFERENCES dk_project(id) ON DELETE CASCADE,
    lender_id            UUID REFERENCES lender(id),
    currency             VARCHAR(10) NOT NULL DEFAULT 'USD',
    amount_original      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    grant_original       NUMERIC(20, 2) NOT NULL DEFAULT 0,
    counterpart_original NUMERIC(20, 2) NOT NULL DEFAULT 0,
    amount_usd           NUMERIC(20, 2) NOT NULL DEFAULT 0,
    grant_usd            NUMERIC(20, 2) NOT NULL DEFAULT 0,
    counterpart_usd      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    remarks              TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Loan Allocation Daftar Kegiatan (per Executing Agency)
-- Mendukung multi-currency: simpan nilai original + ekuivalen USD
CREATE TABLE dk_loan_allocation (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dk_project_id        UUID NOT NULL REFERENCES dk_project(id) ON DELETE CASCADE,
    institution_id       UUID REFERENCES institution(id),
    currency             VARCHAR(10) NOT NULL DEFAULT 'USD',
    amount_original      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    grant_original       NUMERIC(20, 2) NOT NULL DEFAULT 0,
    counterpart_original NUMERIC(20, 2) NOT NULL DEFAULT 0,
    amount_usd           NUMERIC(20, 2) NOT NULL DEFAULT 0,
    grant_usd            NUMERIC(20, 2) NOT NULL DEFAULT 0,
    counterpart_usd      NUMERIC(20, 2) NOT NULL DEFAULT 0,
    remarks              TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Activity Details Daftar Kegiatan (teks bebas, bukan dari GB Activities)
CREATE TABLE dk_activity_detail (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dk_project_id   UUID NOT NULL REFERENCES dk_project(id) ON DELETE CASCADE,
    activity_number INT NOT NULL,
    activity_name   VARCHAR(500) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (dk_project_id, activity_number)
);


-- ============================================================
-- LOAN AGREEMENT
-- ============================================================

CREATE TABLE loan_agreement (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dk_project_id         UUID NOT NULL UNIQUE REFERENCES dk_project(id), -- One-to-One
    lender_id             UUID NOT NULL REFERENCES lender(id),
    loan_code             VARCHAR(100) NOT NULL UNIQUE,
    agreement_date        DATE NOT NULL,
    effective_date        DATE NOT NULL,
    original_closing_date DATE NOT NULL,
    closing_date          DATE NOT NULL,               -- sama dengan original jika belum diperpanjang
    currency              VARCHAR(10) NOT NULL,        -- kode ISO mata uang lender
    amount_original       NUMERIC(20, 2) NOT NULL,     -- dalam mata uang lender
    amount_usd            NUMERIC(20, 2) NOT NULL,     -- ekuivalen USD, dikonversi manual
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_closing_date CHECK (closing_date >= original_closing_date)
);

-- View helper: deteksi LA yang diperpanjang
CREATE VIEW loan_agreement_extended AS
SELECT
    id,
    loan_code,
    original_closing_date,
    closing_date,
    (closing_date != original_closing_date) AS is_extended,
    (closing_date - original_closing_date)  AS extension_days
FROM loan_agreement;


-- ============================================================
-- MONITORING DISBURSEMENT
-- ============================================================

CREATE TABLE monitoring_disbursement (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    loan_agreement_id     UUID NOT NULL REFERENCES loan_agreement(id),
    budget_year           INT NOT NULL,
    quarter               VARCHAR(3) NOT NULL CHECK (quarter IN ('TW1', 'TW2', 'TW3', 'TW4')),
    -- TW1: Apr-Jun | TW2: Jul-Sep | TW3: Okt-Des | TW4: Jan-Mar
    exchange_rate_usd_idr NUMERIC(15, 4) NOT NULL,
    exchange_rate_la_idr  NUMERIC(15, 4) NOT NULL,
    planned_la            NUMERIC(20, 2) NOT NULL DEFAULT 0,
    planned_usd           NUMERIC(20, 2) NOT NULL DEFAULT 0,
    planned_idr           NUMERIC(20, 2) NOT NULL DEFAULT 0,
    realized_la           NUMERIC(20, 2) NOT NULL DEFAULT 0,
    realized_usd          NUMERIC(20, 2) NOT NULL DEFAULT 0,
    realized_idr          NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (loan_agreement_id, budget_year, quarter)
);

-- Breakdown per Komponen (opsional)
CREATE TABLE monitoring_komponen (
    id                         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    monitoring_disbursement_id UUID NOT NULL REFERENCES monitoring_disbursement(id) ON DELETE CASCADE,
    component_name             VARCHAR(500) NOT NULL,
    planned_la                 NUMERIC(20, 2) NOT NULL DEFAULT 0,
    planned_usd                NUMERIC(20, 2) NOT NULL DEFAULT 0,
    planned_idr                NUMERIC(20, 2) NOT NULL DEFAULT 0,
    realized_la                NUMERIC(20, 2) NOT NULL DEFAULT 0,
    realized_usd               NUMERIC(20, 2) NOT NULL DEFAULT 0,
    realized_idr               NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================================
-- USER & PERMISSION
-- ============================================================

CREATE TABLE app_user (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(100) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(10) NOT NULL CHECK (role IN ('ADMIN', 'STAFF')),
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Granular CRUD permission per modul per STAFF
CREATE TABLE user_permission (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
    module     VARCHAR(50) NOT NULL,
    -- contoh: 'blue_book' | 'bb_project' | 'green_book' | 'gb_project'
    --         'daftar_kegiatan' | 'dk_project' | 'loan_agreement' | 'monitoring_disbursement'
    --         'institution' | 'lender' | 'region' | 'national_priority' | 'program_title'
    can_create BOOLEAN NOT NULL DEFAULT FALSE,
    can_read   BOOLEAN NOT NULL DEFAULT FALSE,
    can_update BOOLEAN NOT NULL DEFAULT FALSE,
    can_delete BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, module)
);


-- ============================================================
-- INDEXES
-- ============================================================

-- Blue Book
CREATE INDEX idx_bb_project_blue_book      ON bb_project(blue_book_id);
CREATE INDEX idx_bb_project_identity       ON bb_project(project_identity_id);
CREATE INDEX idx_bb_project_book_code      ON bb_project(blue_book_id, bb_code);
CREATE INDEX idx_bb_project_bb_code        ON bb_project(bb_code);
CREATE INDEX idx_bb_project_status         ON bb_project(status);
CREATE UNIQUE INDEX idx_blue_book_period_version_with_year
    ON blue_book(period_id, revision_number, revision_year)
    WHERE revision_year IS NOT NULL;
CREATE UNIQUE INDEX idx_blue_book_period_version_without_year
    ON blue_book(period_id, revision_number)
    WHERE revision_year IS NULL;
CREATE INDEX idx_lender_indication_project ON lender_indication(bb_project_id);
CREATE INDEX idx_loi_project               ON loi(bb_project_id);
CREATE INDEX idx_loi_lender                ON loi(lender_id);

-- Green Book
CREATE INDEX idx_gb_project_green_book     ON gb_project(green_book_id);
CREATE INDEX idx_gb_project_identity       ON gb_project(gb_project_identity_id);
CREATE INDEX idx_gb_project_book_code      ON gb_project(green_book_id, gb_code);
CREATE INDEX idx_gb_project_gb_code        ON gb_project(gb_code);
CREATE INDEX idx_gb_project_status         ON gb_project(status);
CREATE INDEX idx_gb_activity_project       ON gb_activity(gb_project_id);
CREATE INDEX idx_gb_funding_source_project ON gb_funding_source(gb_project_id);
CREATE INDEX idx_gb_funding_source_lender  ON gb_funding_source(lender_id);
CREATE INDEX idx_gb_disbursement_project   ON gb_disbursement_plan(gb_project_id);

-- Daftar Kegiatan
CREATE INDEX idx_dk_project_dk             ON dk_project(dk_id);
CREATE INDEX idx_dk_financing_project      ON dk_financing_detail(dk_project_id);
CREATE INDEX idx_dk_allocation_project     ON dk_loan_allocation(dk_project_id);

-- Loan Agreement
CREATE INDEX idx_la_dk_project             ON loan_agreement(dk_project_id);
CREATE INDEX idx_la_lender                 ON loan_agreement(lender_id);
CREATE INDEX idx_la_effective_date         ON loan_agreement(effective_date);
CREATE INDEX idx_la_closing_date           ON loan_agreement(closing_date);

-- Monitoring
CREATE INDEX idx_monitoring_la             ON monitoring_disbursement(loan_agreement_id);
CREATE INDEX idx_monitoring_budget_year    ON monitoring_disbursement(budget_year);
CREATE INDEX idx_monitoring_komponen       ON monitoring_komponen(monitoring_disbursement_id);

-- User & Permission
CREATE INDEX idx_user_permission_user      ON user_permission(user_id);
CREATE INDEX idx_user_permission_module    ON user_permission(module);


-- ============================================================
-- AUDIT TRAIL
-- ============================================================

CREATE TABLE audit_log (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(100) NOT NULL,
    record_id  UUID         NOT NULL,
    action     VARCHAR(10)  NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_data   JSONB,
    new_data   JSONB,
    changed_by UUID REFERENCES app_user(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_table_record ON audit_log(table_name, record_id);
CREATE INDEX idx_audit_changed_by   ON audit_log(changed_by);
CREATE INDEX idx_audit_changed_at   ON audit_log(changed_at DESC);
CREATE INDEX idx_audit_action       ON audit_log(action);
CREATE INDEX idx_audit_old_data     ON audit_log USING GIN (old_data);
CREATE INDEX idx_audit_new_data     ON audit_log USING GIN (new_data);


-- ============================================================
-- AUDIT TRIGGER FUNCTION
-- ============================================================

CREATE OR REPLACE FUNCTION audit_trigger_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_user_id UUID;
BEGIN
    BEGIN
        v_user_id := current_setting('app.current_user_id', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_user_id := NULL;
    END;

    INSERT INTO audit_log (table_name, record_id, action, old_data, new_data, changed_by)
    VALUES (
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE to_jsonb(OLD) END,
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE to_jsonb(NEW) END,
        v_user_id
    );

    RETURN COALESCE(NEW, OLD);
END;
$$;


-- ============================================================
-- PASANG TRIGGER KE SEMUA TABEL
-- ============================================================

-- Master data
CREATE TRIGGER trg_audit_country
    AFTER INSERT OR UPDATE OR DELETE ON country
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_lender
    AFTER INSERT OR UPDATE OR DELETE ON lender
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_institution
    AFTER INSERT OR UPDATE OR DELETE ON institution
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_region
    AFTER INSERT OR UPDATE OR DELETE ON region
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_program_title
    AFTER INSERT OR UPDATE OR DELETE ON program_title
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_bappenas_partner
    AFTER INSERT OR UPDATE OR DELETE ON bappenas_partner
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_period
    AFTER INSERT OR UPDATE OR DELETE ON period
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_national_priority
    AFTER INSERT OR UPDATE OR DELETE ON national_priority
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

-- Blue Book
CREATE TRIGGER trg_audit_blue_book
    AFTER INSERT OR UPDATE OR DELETE ON blue_book
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_bb_project
    AFTER INSERT OR UPDATE OR DELETE ON bb_project
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_bb_project_cost
    AFTER INSERT OR UPDATE OR DELETE ON bb_project_cost
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_lender_indication
    AFTER INSERT OR UPDATE OR DELETE ON lender_indication
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_loi
    AFTER INSERT OR UPDATE OR DELETE ON loi
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

-- Green Book
CREATE TRIGGER trg_audit_green_book
    AFTER INSERT OR UPDATE OR DELETE ON green_book
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_gb_project
    AFTER INSERT OR UPDATE OR DELETE ON gb_project
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_gb_activity
    AFTER INSERT OR UPDATE OR DELETE ON gb_activity
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_gb_funding_source
    AFTER INSERT OR UPDATE OR DELETE ON gb_funding_source
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_gb_disbursement_plan
    AFTER INSERT OR UPDATE OR DELETE ON gb_disbursement_plan
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_gb_funding_allocation
    AFTER INSERT OR UPDATE OR DELETE ON gb_funding_allocation
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

-- Daftar Kegiatan
CREATE TRIGGER trg_audit_daftar_kegiatan
    AFTER INSERT OR UPDATE OR DELETE ON daftar_kegiatan
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_dk_project
    AFTER INSERT OR UPDATE OR DELETE ON dk_project
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_dk_financing_detail
    AFTER INSERT OR UPDATE OR DELETE ON dk_financing_detail
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_dk_loan_allocation
    AFTER INSERT OR UPDATE OR DELETE ON dk_loan_allocation
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_dk_activity_detail
    AFTER INSERT OR UPDATE OR DELETE ON dk_activity_detail
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

-- Loan Agreement
CREATE TRIGGER trg_audit_loan_agreement
    AFTER INSERT OR UPDATE OR DELETE ON loan_agreement
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

-- Monitoring
CREATE TRIGGER trg_audit_monitoring_disbursement
    AFTER INSERT OR UPDATE OR DELETE ON monitoring_disbursement
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_monitoring_komponen
    AFTER INSERT OR UPDATE OR DELETE ON monitoring_komponen
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

-- User & Permission
CREATE TRIGGER trg_audit_app_user
    AFTER INSERT OR UPDATE OR DELETE ON app_user
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER trg_audit_user_permission
    AFTER INSERT OR UPDATE OR DELETE ON user_permission
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();


-- ============================================================
-- HELPER VIEWS UNTUK AUDIT
-- ============================================================

CREATE VIEW v_audit_record_history AS
SELECT
    al.changed_at,
    al.action,
    al.table_name,
    al.record_id,
    u.username AS changed_by,
    al.old_data,
    al.new_data
FROM audit_log al
LEFT JOIN app_user u ON u.id = al.changed_by
ORDER BY al.changed_at DESC;

CREATE VIEW v_audit_recent_activity AS
SELECT
    u.username,
    al.action,
    al.table_name,
    al.record_id,
    al.changed_at
FROM audit_log al
LEFT JOIN app_user u ON u.id = al.changed_by
WHERE al.changed_at >= NOW() - INTERVAL '30 days'
ORDER BY al.changed_at DESC;


-- ============================================================
-- CATATAN IMPLEMENTASI
-- ============================================================
-- 1. Setiap request dari aplikasi wajib mengeset user aktif di awal transaksi:
--       SET LOCAL app.current_user_id = '<user_uuid>';
--
-- 2. Tabel junction (bb_project_institution, bb_project_location, dst.)
--    tidak dipasang trigger audit karena tidak memiliki kolom 'id' UUID.
--
-- 3. Untuk kebutuhan archiving, pertimbangkan partisi audit_log per bulan.
--
-- 4. Kolom old_data dan new_data menyimpan seluruh row sebagai JSONB —
--    termasuk kolom sensitif seperti password_hash di app_user.
--    Pastikan akses ke audit_log dibatasi hanya untuk ADMIN.
