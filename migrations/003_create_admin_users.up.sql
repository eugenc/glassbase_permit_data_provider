CREATE TABLE IF NOT EXISTS admin_users (
    id             SERIAL PRIMARY KEY,
    email          TEXT NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    name           TEXT NOT NULL,
    role           TEXT NOT NULL DEFAULT 'admin',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at  TIMESTAMPTZ
);

INSERT INTO admin_users (email, password_hash, name, role)
VALUES (
    'admin@glassbase.io',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lh9i',
    'Admin',
    'admin'
) ON CONFLICT DO NOTHING;
