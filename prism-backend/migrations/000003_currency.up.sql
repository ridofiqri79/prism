CREATE TABLE IF NOT EXISTS currency (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        CHAR(3) NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    symbol      VARCHAR(16),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_audit_currency ON currency;

CREATE TRIGGER trg_audit_currency
    AFTER INSERT OR UPDATE OR DELETE ON currency
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

INSERT INTO currency (code, name, symbol, is_active, sort_order) VALUES
('USD', 'United States Dollar', '$', TRUE, 10),
('EUR', 'Euro', 'EUR', TRUE, 20),
('JPY', 'Japanese Yen', 'JPY', TRUE, 30),
('KRW', 'South Korean Won', 'KRW', TRUE, 40),
('CNY', 'Chinese Yuan', 'CNY', TRUE, 50),
('AUD', 'Australian Dollar', 'A$', TRUE, 60),
('CAD', 'Canadian Dollar', 'C$', TRUE, 70),
('GBP', 'Pound Sterling', 'GBP', TRUE, 80),
('CHF', 'Swiss Franc', 'CHF', TRUE, 90),
('SAR', 'Saudi Riyal', 'SAR', TRUE, 100),
('SGD', 'Singapore Dollar', 'S$', TRUE, 110),
('IDR', 'Indonesian Rupiah', 'Rp', TRUE, 120),
('XDR', 'Special Drawing Rights', 'XDR', FALSE, 130)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    symbol = EXCLUDED.symbol,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
