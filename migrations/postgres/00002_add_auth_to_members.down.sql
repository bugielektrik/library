-- Remove authentication fields from members table

-- Drop constraints
ALTER TABLE members
DROP CONSTRAINT IF EXISTS check_member_role;

-- Drop indexes
DROP INDEX IF EXISTS idx_members_email;
DROP INDEX IF EXISTS idx_members_role;

-- Remove columns
ALTER TABLE members
DROP COLUMN IF EXISTS email,
DROP COLUMN IF EXISTS password_hash,
DROP COLUMN IF EXISTS role,
DROP COLUMN IF EXISTS last_login_at;