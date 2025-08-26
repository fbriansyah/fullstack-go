-- Drop audit_events table and its indexes
DROP INDEX IF EXISTS idx_audit_events_occurred_at_desc;
DROP INDEX IF EXISTS idx_audit_events_resource_action;
DROP INDEX IF EXISTS idx_audit_events_user_action;
DROP INDEX IF EXISTS idx_audit_events_event_id;
DROP INDEX IF EXISTS idx_audit_events_occurred_at;
DROP INDEX IF EXISTS idx_audit_events_event_type;
DROP INDEX IF EXISTS idx_audit_events_resource_id;
DROP INDEX IF EXISTS idx_audit_events_resource;
DROP INDEX IF EXISTS idx_audit_events_action;
DROP INDEX IF EXISTS idx_audit_events_user_id;

DROP TABLE IF EXISTS audit_events;