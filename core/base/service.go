package base

import (
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/types"
	"fmt"
	"math"

	"gorm.io/gorm"
)

// Service provides common functionality for all services
type Service struct {
	DB      *gorm.DB
	Logger  logger.Logger
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
}

// NewService creates a new Service instance
func NewService(db *gorm.DB, logger logger.Logger, emitter *emitter.Emitter, storage *storage.ActiveStorage) *Service {
	return &Service{
		DB:      db,
		Logger:  logger,
		Emitter: emitter,
		Storage: storage,
	}
}

// LogError logs an error with context
func (bs *Service) LogError(operation string, err error, context ...logger.Field) {
	fields := append([]logger.Field{
		logger.String("operation", operation),
		logger.String("error", err.Error()),
	}, context...)

	bs.Logger.Error("Service error", fields...)
}

// LogInfo logs an info message with context
func (bs *Service) LogInfo(operation string, message string, context ...logger.Field) {
	fields := append([]logger.Field{
		logger.String("operation", operation),
		logger.String("message", message),
	}, context...)

	bs.Logger.Info("Service operation", fields...)
}

// EmitEvent emits an event through the event emitter
func (bs *Service) EmitEvent(eventName string, data any) {
	if bs.Emitter != nil {
		bs.Emitter.Emit(eventName, data)
	}
}

// CreatePaginatedResponse creates a paginated response
func (bs *Service) CreatePaginatedResponse(data any, total int64, page int, limit int) *types.PaginatedResponse {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &types.PaginatedResponse{
		Data: data,
		Pagination: types.Pagination{
			Total:      int(total),
			Page:       page,
			PageSize:   limit,
			TotalPages: totalPages,
		},
	}
}

// ValidateID validates that an ID is valid (greater than 0)
func (bs *Service) ValidateID(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid ID: ID must be greater than 0")
	}
	return nil
}

// BeginTransaction starts a database transaction
func (bs *Service) BeginTransaction() *gorm.DB {
	return bs.DB.Begin()
}

// CommitTransaction commits a database transaction
func (bs *Service) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTransaction rolls back a database transaction
func (bs *Service) RollbackTransaction(tx *gorm.DB) {
	tx.Rollback()
}

// WithTransaction executes a function within a database transaction
func (bs *Service) WithTransaction(fn func(*gorm.DB) error) error {
	tx := bs.BeginTransaction()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			bs.RollbackTransaction(tx)
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		bs.RollbackTransaction(tx)
		return err
	}

	return bs.CommitTransaction(tx)
}

// FindByID is a generic function to find a record by ID
func (bs *Service) FindByID(model any, id uint, preloads ...string) error {
	if err := bs.ValidateID(id); err != nil {
		return err
	}

	query := bs.DB
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	return query.First(model, id).Error
}

// Count counts records matching the given conditions
func (bs *Service) Count(model any, conditions ...any) (int64, error) {
	var count int64
	query := bs.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	return count, query.Count(&count).Error
}

// Delete performs a soft delete on a record
func (bs *Service) Delete(model any, id uint) error {
	if err := bs.ValidateID(id); err != nil {
		return err
	}

	return bs.DB.Delete(model, id).Error
}

// HardDelete performs a hard delete on a record
func (bs *Service) HardDelete(model any, id uint) error {
	if err := bs.ValidateID(id); err != nil {
		return err
	}

	return bs.DB.Unscoped().Delete(model, id).Error
}
