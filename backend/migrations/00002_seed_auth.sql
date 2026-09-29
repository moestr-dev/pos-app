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
('inventory.opname', 'Stock opname');

INSERT INTO auth.roles (name, description) VALUES
('admin', 'Full access'),
('manager', 'Manage inventory and staff'),
('staff', 'Basic operational access');

INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p WHERE r.name = 'admin';

INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p
WHERE r.name = 'manager' AND p.code IN (
    'user.read', 'user.create', 'user.update',
    'inventory.read', 'inventory.manage', 'inventory.transfer', 'inventory.opname'
);

INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p
WHERE r.name = 'staff' AND p.code IN ('inventory.read');

-- +goose Down
DELETE FROM auth.role_permissions;
DELETE FROM auth.roles;
DELETE FROM auth.permissions;