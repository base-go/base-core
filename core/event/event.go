package event

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JsonMap is a custom type for handling JSON data in the database
type JsonMap map[string]interface{}

// Value implements the driver.Valuer interface for JsonMap
func (j JsonMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JsonMap
func (j *JsonMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}

	result := make(JsonMap)
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}
	*j = result
	return nil
}

// Event represents a system event
type Event struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Type        string         `gorm:"index;not null;type:varchar(191)" json:"type"`
	Category    string         `gorm:"index;not null;type:varchar(191)" json:"category"`
	Actor       string         `gorm:"index;type:varchar(191)" json:"actor"`
	ActorID     string         `gorm:"index;type:varchar(191)" json:"actor_id"`
	Target      string         `gorm:"index;type:varchar(191)" json:"target"`
	TargetID    string         `gorm:"index;type:varchar(191)" json:"target_id"`
	Action      string         `gorm:"index;not null;type:varchar(191)" json:"action"`
	Status      string         `gorm:"index;type:varchar(191)" json:"status"`
	Description string         `gorm:"type:text" json:"description"`
	Metadata    JsonMap        `gorm:"type:JSON" json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// EventService handles event tracking and storage
type EventService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewEventService creates a new event service
func NewEventService(db *gorm.DB, logger *zap.Logger) *EventService {
	// Auto-migrate the events table
	err := db.AutoMigrate(&Event{})
	if err != nil {
		logger.Error("Failed to migrate events table", zap.Error(err))
	}

	return &EventService{
		db:     db,
		logger: logger,
	}
}

// EventOptions contains options for creating an event
type EventOptions struct {
	Type        string
	Category    string
	Actor       string
	ActorID     string
	Target      string
	TargetID    string
	Action      string
	Status      string
	Description string
	Metadata    map[string]interface{}
}

// Track creates and stores a new event
func (s *EventService) Track(ctx context.Context, opts EventOptions) (*Event, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	event := &Event{
		Type:        opts.Type,
		Category:    opts.Category,
		Actor:       opts.Actor,
		ActorID:     opts.ActorID,
		Target:      opts.Target,
		TargetID:    opts.TargetID,
		Action:      opts.Action,
		Status:      opts.Status,
		Description: opts.Description,
		Metadata:    opts.Metadata,
	}

	// Use a transaction for better consistency
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(event).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create event",
			zap.Error(err),
			zap.String("type", opts.Type),
			zap.String("action", opts.Action))
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit event transaction",
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Event tracked successfully",
		zap.String("type", event.Type),
		zap.String("action", event.Action),
		zap.Uint("event_id", event.ID))

	return event, nil
}

// Query represents search criteria for events
type Query struct {
	Types      []string
	Categories []string
	Actors     []string
	ActorIDs   []string
	Targets    []string
	TargetIDs  []string
	Actions    []string
	Statuses   []string
	StartDate  *time.Time
	EndDate    *time.Time
	PageSize   int
	PageNumber int
}

// FindEvents searches for events based on the provided criteria
func (s *EventService) FindEvents(ctx context.Context, query Query) ([]Event, int64, error) {
	db := s.db.WithContext(ctx).Model(&Event{})

	// Apply filters
	if len(query.Types) > 0 {
		db = db.Where("type IN ?", query.Types)
	}
	if len(query.Categories) > 0 {
		db = db.Where("category IN ?", query.Categories)
	}
	if len(query.Actors) > 0 {
		db = db.Where("actor IN ?", query.Actors)
	}
	if len(query.ActorIDs) > 0 {
		db = db.Where("actor_id IN ?", query.ActorIDs)
	}
	if len(query.Targets) > 0 {
		db = db.Where("target IN ?", query.Targets)
	}
	if len(query.TargetIDs) > 0 {
		db = db.Where("target_id IN ?", query.TargetIDs)
	}
	if len(query.Actions) > 0 {
		db = db.Where("action IN ?", query.Actions)
	}
	if len(query.Statuses) > 0 {
		db = db.Where("status IN ?", query.Statuses)
	}
	if query.StartDate != nil {
		db = db.Where("created_at >= ?", query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("created_at <= ?", query.EndDate)
	}

	// Count total records
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if query.PageSize > 0 {
		offset := (query.PageNumber - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	// Get results
	var events []Event
	if err := db.Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventByID retrieves an event by its ID
func (s *EventService) GetEventByID(ctx context.Context, id uint) (*Event, error) {
	var event Event
	if err := s.db.WithContext(ctx).First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// DeleteEvent soft deletes an event
func (s *EventService) DeleteEvent(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&Event{}, id).Error
}

// Cleanup deletes events older than the specified duration
func (s *EventService) Cleanup(ctx context.Context, olderThan time.Duration) error {
	cutoffDate := time.Now().Add(-olderThan)

	result := s.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&Event{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup old events",
			zap.Error(result.Error),
			zap.Time("cutoff_date", cutoffDate))
		return result.Error
	}

	s.logger.Info("Successfully cleaned up old events",
		zap.Int64("deleted_count", result.RowsAffected),
		zap.Time("cutoff_date", cutoffDate))

	return nil
}
