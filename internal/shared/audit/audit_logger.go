package audit

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/jmoiron/sqlx"
)

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	LogEvent(ctx context.Context, event *AuditEvent) error
	GetEvents(ctx context.Context, filter *AuditFilter) ([]*AuditEvent, error)
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	UserID        string                 `json:"user_id"`
	Action        string                 `json:"action"`
	Resource      string                 `json:"resource"`
	ResourceID    string                 `json:"resource_id"`
	Details       map[string]interface{} `json:"details"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Metadata      events.EventMetadata   `json:"metadata"`
}

// AuditFilter defines filters for querying audit events
type AuditFilter struct {
	EventID    string    `json:"event_id,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	Action     string    `json:"action,omitempty"`
	Resource   string    `json:"resource,omitempty"`
	ResourceID string    `json:"resource_id,omitempty"`
	EventType  string    `json:"event_type,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
}

// auditLoggerImpl implements the AuditLogger interface using PostgreSQL
type auditLoggerImpl struct {
	db *database.DB
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(db *database.DB) AuditLogger {
	return &auditLoggerImpl{
		db: db,
	}
}

// auditEventRecord represents the database record for audit events
type auditEventRecord struct {
	ID            int64             `db:"id"`
	EventID       string            `db:"event_id"`
	EventType     string            `db:"event_type"`
	AggregateID   string            `db:"aggregate_id"`
	AggregateType string            `db:"aggregate_type"`
	UserID        string            `db:"user_id"`
	Action        string            `db:"action"`
	Resource      string            `db:"resource"`
	ResourceID    string            `db:"resource_id"`
	Details       JSONMap           `db:"details"`
	OccurredAt    time.Time         `db:"occurred_at"`
	Metadata      EventMetadataJSON `db:"metadata"`
	CreatedAt     time.Time         `db:"created_at"`
}

// JSONMap is a custom type for handling JSON data in PostgreSQL
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database retrieval
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// EventMetadataJSON is a custom type for handling EventMetadata as JSON
type EventMetadataJSON events.EventMetadata

// Value implements the driver.Valuer interface for database storage
func (e EventMetadataJSON) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements the sql.Scanner interface for database retrieval
func (e *EventMetadataJSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into EventMetadataJSON", value)
	}

	return json.Unmarshal(bytes, e)
}

// LogEvent logs an audit event to the database
func (a *auditLoggerImpl) LogEvent(ctx context.Context, event *AuditEvent) error {
	query := `
		INSERT INTO audit_events (
			event_id, event_type, aggregate_id, aggregate_type, user_id,
			action, resource, resource_id, details, occurred_at, metadata, created_at
		) VALUES (
			:event_id, :event_type, :aggregate_id, :aggregate_type, :user_id,
			:action, :resource, :resource_id, :details, :occurred_at, :metadata, :created_at
		)`

	record := &auditEventRecord{
		EventID:       event.EventID,
		EventType:     event.EventType,
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		UserID:        event.UserID,
		Action:        event.Action,
		Resource:      event.Resource,
		ResourceID:    event.ResourceID,
		Details:       JSONMap(event.Details),
		OccurredAt:    event.OccurredAt,
		Metadata:      EventMetadataJSON(event.Metadata),
		CreatedAt:     time.Now().UTC(),
	}

	_, err := a.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return fmt.Errorf("failed to insert audit event: %w", err)
	}

	return nil
}

// GetEvents retrieves audit events based on the provided filter
func (a *auditLoggerImpl) GetEvents(ctx context.Context, filter *AuditFilter) ([]*AuditEvent, error) {
	query := `
		SELECT 
			id, event_id, event_type, aggregate_id, aggregate_type, user_id,
			action, resource, resource_id, details, occurred_at, metadata, created_at
		FROM audit_events
		WHERE 1=1`

	args := make(map[string]interface{})

	if filter.EventID != "" {
		query += " AND event_id = :event_id"
		args["event_id"] = filter.EventID
	}

	if filter.UserID != "" {
		query += " AND user_id = :user_id"
		args["user_id"] = filter.UserID
	}

	if filter.Action != "" {
		query += " AND action = :action"
		args["action"] = filter.Action
	}

	if filter.Resource != "" {
		query += " AND resource = :resource"
		args["resource"] = filter.Resource
	}

	if filter.ResourceID != "" {
		query += " AND resource_id = :resource_id"
		args["resource_id"] = filter.ResourceID
	}

	if filter.EventType != "" {
		query += " AND event_type = :event_type"
		args["event_type"] = filter.EventType
	}

	if !filter.StartTime.IsZero() {
		query += " AND occurred_at >= :start_time"
		args["start_time"] = filter.StartTime
	}

	if !filter.EndTime.IsZero() {
		query += " AND occurred_at <= :end_time"
		args["end_time"] = filter.EndTime
	}

	query += " ORDER BY occurred_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT :limit"
		args["limit"] = filter.Limit
	}

	if filter.Offset > 0 {
		query += " OFFSET :offset"
		args["offset"] = filter.Offset
	}

	var records []auditEventRecord
	nstmt, err := a.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer nstmt.Close()

	err = nstmt.SelectContext(ctx, &records, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}

	auditEvents := make([]*AuditEvent, len(records))
	for i, record := range records {
		auditEvents[i] = &AuditEvent{
			EventID:       record.EventID,
			EventType:     record.EventType,
			AggregateID:   record.AggregateID,
			AggregateType: record.AggregateType,
			UserID:        record.UserID,
			Action:        record.Action,
			Resource:      record.Resource,
			ResourceID:    record.ResourceID,
			Details:       map[string]interface{}(record.Details),
			OccurredAt:    record.OccurredAt,
			Metadata:      events.EventMetadata(record.Metadata),
		}
	}

	return auditEvents, nil
}

// CreateAuditEventsTable creates the audit_events table if it doesn't exist
func CreateAuditEventsTable(ctx context.Context, db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS audit_events (
			id BIGSERIAL PRIMARY KEY,
			event_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			aggregate_id VARCHAR(255) NOT NULL,
			aggregate_type VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			action VARCHAR(255) NOT NULL,
			resource VARCHAR(255) NOT NULL,
			resource_id VARCHAR(255) NOT NULL,
			details JSONB,
			occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_audit_events_user_id ON audit_events(user_id);
		CREATE INDEX IF NOT EXISTS idx_audit_events_action ON audit_events(action);
		CREATE INDEX IF NOT EXISTS idx_audit_events_resource ON audit_events(resource);
		CREATE INDEX IF NOT EXISTS idx_audit_events_resource_id ON audit_events(resource_id);
		CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON audit_events(event_type);
		CREATE INDEX IF NOT EXISTS idx_audit_events_occurred_at ON audit_events(occurred_at);
		CREATE INDEX IF NOT EXISTS idx_audit_events_event_id ON audit_events(event_id);
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create audit_events table: %w", err)
	}

	return nil
}
