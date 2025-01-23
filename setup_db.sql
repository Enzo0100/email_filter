-- Clean up existing data
DELETE FROM users;
DELETE FROM tenants;

-- Add password_hash column
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255) NOT NULL;

-- Insert tenant
INSERT INTO tenants (id, name, plan, active)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Default Tenant',
    'free',
    true
);

-- Insert admin user
INSERT INTO users (id, tenant_id, email, name, role, password_hash, active)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'admin@example.com',
    'Admin User',
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',  -- password: test123
    true
);
