-- Add authentication fields to members table

-- Add email column with unique constraint
ALTER TABLE members
ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE;

-- Add password_hash column
ALTER TABLE members
ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Add role column with default value
ALTER TABLE members
ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'user';

-- Add last_login_at timestamp
ALTER TABLE members
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP;

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_members_email ON members(email);

-- Create index on role for role-based queries
CREATE INDEX IF NOT EXISTS idx_members_role ON members(role);

-- Update existing members to have a default email and role
-- (This is for development/testing - in production you'd handle this differently)
UPDATE members
SET email = CONCAT('member_', id::text, '@library.local'),
    role = 'user'
WHERE email IS NULL;

-- Now make email NOT NULL after populating existing records
ALTER TABLE members
ALTER COLUMN email SET NOT NULL;

-- Make password_hash NOT NULL as well (but only for new records)
-- Existing records without password won't be able to login
ALTER TABLE members
ALTER COLUMN password_hash SET DEFAULT '';

-- Add constraint to ensure role is valid
ALTER TABLE members
ADD CONSTRAINT check_member_role
CHECK (role IN ('user', 'admin'));