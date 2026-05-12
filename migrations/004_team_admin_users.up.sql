-- Team admin accounts (password set via bcrypt only — rotate via DB if needed).
-- Applies same credential for all listed emails; upserts on conflict.

INSERT INTO admin_users (email, password_hash, name, role)
VALUES
    (
        'andriy@glassbase.com',
        '$2a$10$AKL6AKXRvyjHjfMYsPQ9SeJPf2rG.cLV97P3XBggjWnCHzINxl3oC',
        'Andriy',
        'admin'
    ),
    (
        'eugenec@affiniti.io',
        '$2a$10$AKL6AKXRvyjHjfMYsPQ9SeJPf2rG.cLV97P3XBggjWnCHzINxl3oC',
        'Eugene',
        'admin'
    ),
    (
        'anton@glassbase.com',
        '$2a$10$AKL6AKXRvyjHjfMYsPQ9SeJPf2rG.cLV97P3XBggjWnCHzINxl3oC',
        'Anton',
        'admin'
    )
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    name            = EXCLUDED.name,
    role            = EXCLUDED.role;
