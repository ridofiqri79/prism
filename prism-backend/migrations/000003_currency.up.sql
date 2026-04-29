
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
