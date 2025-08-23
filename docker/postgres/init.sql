-- Initialize database for go-templ-template
-- This file is executed when the PostgreSQL container starts for the first time

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC';

-- Create initial schemas (will be populated by migrations)
-- Users and sessions tables will be created by migrations