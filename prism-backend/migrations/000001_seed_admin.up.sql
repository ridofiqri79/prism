-- Seed ADMIN user pertama (password: admin123).
-- Password harus diganti setelah login pertama.
INSERT INTO app_user (username, email, password_hash, role)
VALUES (
    'admin',
    'admin@prism.go.id',
    '$2a$10$KHYKwWSrJSZ0N3FKX/vzt.ENZrGO1sj/vJ9noelFhZIFKOeCm0vAS',
    'ADMIN'
)
ON CONFLICT (username) DO NOTHING;
