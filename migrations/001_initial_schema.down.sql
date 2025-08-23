-- Rollback initial schema migration

-- Drop functions
DROP FUNCTION IF EXISTS cleanup_expired_sessions();
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Drop tables (CASCADE will drop dependent objects like triggers)
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop extension (only if no other objects depend on it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";