-- Add is_active column to sessions table
ALTER TABLE sessions ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Create index for better performance on active sessions
CREATE INDEX idx_sessions_is_active ON sessions(is_active);

-- Create composite index for user active sessions
CREATE INDEX idx_sessions_user_active ON sessions(user_id, is_active, expires_at);