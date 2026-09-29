-- +goose Up
INSERT INTO auth.permissions (code, description) VALUES
('user.read', 'View users'),
('user.create', 'Create users'),
('user.update', 'Update users'),
('user.delete', 'Deactivate users'),
('role.read', 'View roles'),
('role.manage', 'Manage roles and permissions'),
('inventory.read', 'View inventory'),
('inventory.manage', 'Manage products and branches'),
('inventory.transfer', 'Transfer stock between branches'),
('inventory.opname', 'Stock opname')
ON CONFLICT (code) DO NOTHING;

INSERT INTO auth.roles (name, description) VALUES
('admin', 'Full access'),
('manager', 'Manage inventory and staff'),
('staff', 'Basic operational access')
ON CONFLICT (name) DO NOTHING;

INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p
WHERE r.name = 'manager' AND p.code IN (
    'user.read', 'user.create', 'user.update',
    'inventory.read', 'inventory.manage', 'inventory.transfer', 'inventory.opname'
)
ON CONFLICT DO NOTHING;

INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p
WHERE r.name = 'staff' AND p.code IN ('inventory.read')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM auth.role_permissions
WHERE role_id IN (SELECT id FROM auth.roles WHERE name IN ('admin', 'manager', 'staff'));

DELETE FROM auth.roles
WHERE name IN ('admin', 'manager', 'staff');

DELETE FROM auth.permissions
WHERE code IN (
    'user.read', 'user.create', 'user.update', 'user.delete',
    'role.read', 'role.manage',
    'inventory.read', 'inventory.manage', 'inventory.transfer', 'inventory.opname'
);
