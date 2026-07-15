INSERT INTO users (id, email, password, name, role, team_id, is_active, created_at, updated_at)
VALUES (gen_random_uuid(), 'admin@leadsales.local',
    '$2a$10$ovc.rgngVMT.JEh71Mes4uoTsyGLyM47O799RNazDVZOJ93GFwoqa',
    'Admin SU', 'SU', NULL, true, now(), now());
