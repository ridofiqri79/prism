CREATE TABLE IF NOT EXISTS kurs_tengah (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    currency_id     UUID NOT NULL REFERENCES currency(id),
    kurs            NUMERIC(20, 6) NOT NULL CHECK (kurs > 0),
    kurs_tengah_bi  NUMERIC(20, 6) NOT NULL CHECK (kurs_tengah_bi > 0),
    cut_off_date    DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_kurs_tengah_currency_cutoff UNIQUE (currency_id, cut_off_date)
);

CREATE INDEX IF NOT EXISTS idx_kurs_tengah_currency_cutoff
    ON kurs_tengah(currency_id, cut_off_date DESC);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trg_audit_kurs_tengah'
    ) THEN
        CREATE TRIGGER trg_audit_kurs_tengah
            AFTER INSERT OR UPDATE OR DELETE ON kurs_tengah
            FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();
    END IF;
END $$;
