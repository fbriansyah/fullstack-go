-- Remove indexes
DROP INDEX IF EXISTS idx_sessions_user_active;
DROP INDEX IF EXISTS idx_sessions_is_active;

-- Remove is_active column from sessions table
ALTER TABLE sessions DROP COLUMN IF EXISTS is_active;