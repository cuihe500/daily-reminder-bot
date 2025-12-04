package repository

import (
	"time"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WarningLogRepository handles database operations for warning logs
type WarningLogRepository struct {
	db *gorm.DB
}

// NewWarningLogRepository creates a new WarningLogRepository
func NewWarningLogRepository(db *gorm.DB) *WarningLogRepository {
	return &WarningLogRepository{db: db}
}

// GetByWarningID retrieves a warning log by its warning ID
func (r *WarningLogRepository) GetByWarningID(warningID string) (*model.WarningLog, error) {
	logger.Debug("WarningLogRepository.GetByWarningID",
		zap.String("warning_id", warningID))

	var log model.WarningLog
	result := r.db.Where("warning_id = ?", warningID).First(&log)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Debug("Warning log not found",
				zap.String("warning_id", warningID))
			return nil, nil
		}
		logger.Error("Failed to get warning log",
			zap.String("warning_id", warningID),
			zap.Error(result.Error))
		return nil, result.Error
	}

	logger.Debug("Warning log retrieved",
		zap.String("warning_id", warningID),
		zap.String("status", log.Status))
	return &log, nil
}

// Create creates a new warning log
func (r *WarningLogRepository) Create(log *model.WarningLog) error {
	logger.Debug("WarningLogRepository.Create",
		zap.String("warning_id", log.WarningID),
		zap.String("city", log.City))

	result := r.db.Create(log)
	if result.Error != nil {
		logger.Error("Failed to create warning log",
			zap.String("warning_id", log.WarningID),
			zap.Error(result.Error))
		return result.Error
	}

	logger.Debug("Warning log created",
		zap.String("warning_id", log.WarningID),
		zap.Uint("id", log.ID))
	return nil
}

// Update updates an existing warning log
func (r *WarningLogRepository) Update(log *model.WarningLog) error {
	logger.Debug("WarningLogRepository.Update",
		zap.String("warning_id", log.WarningID),
		zap.String("status", log.Status))

	result := r.db.Save(log)
	if result.Error != nil {
		logger.Error("Failed to update warning log",
			zap.String("warning_id", log.WarningID),
			zap.Error(result.Error))
		return result.Error
	}

	logger.Debug("Warning log updated",
		zap.String("warning_id", log.WarningID))
	return nil
}

// DeleteOldLogs deletes warning logs older than the specified duration
func (r *WarningLogRepository) DeleteOldLogs(olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	logger.Debug("WarningLogRepository.DeleteOldLogs",
		zap.Time("cutoff_time", cutoffTime))

	result := r.db.Where("created_at < ?", cutoffTime).Delete(&model.WarningLog{})
	if result.Error != nil {
		logger.Error("Failed to delete old warning logs",
			zap.Error(result.Error))
		return result.Error
	}

	logger.Info("Old warning logs deleted",
		zap.Int64("deleted_count", result.RowsAffected))
	return nil
}

// GetActiveWarningsByLocationID retrieves active warnings for a location
func (r *WarningLogRepository) GetActiveWarningsByLocationID(locationID string) ([]model.WarningLog, error) {
	logger.Debug("WarningLogRepository.GetActiveWarningsByLocationID",
		zap.String("location_id", locationID))

	var logs []model.WarningLog
	result := r.db.Where("location_id = ? AND status = ?", locationID, "active").
		Order("start_time DESC").
		Find(&logs)

	if result.Error != nil {
		logger.Error("Failed to get active warnings",
			zap.String("location_id", locationID),
			zap.Error(result.Error))
		return nil, result.Error
	}

	logger.Debug("Active warnings retrieved",
		zap.String("location_id", locationID),
		zap.Int("count", len(logs)))
	return logs, nil
}

// GetUnresolvedWarningsByCity retrieves all unresolved warnings for a city
// Unresolved means status is not 'resolved' (i.e., 'active' or 'update')
func (r *WarningLogRepository) GetUnresolvedWarningsByCity(city string) ([]model.WarningLog, error) {
	logger.Debug("WarningLogRepository.GetUnresolvedWarningsByCity",
		zap.String("city", city))

	var logs []model.WarningLog
	result := r.db.Where("city = ? AND status != ?", city, "resolved").
		Order("start_time DESC").
		Find(&logs)

	if result.Error != nil {
		logger.Error("Failed to get unresolved warnings",
			zap.String("city", city),
			zap.Error(result.Error))
		return nil, result.Error
	}

	logger.Debug("Unresolved warnings retrieved",
		zap.String("city", city),
		zap.Int("count", len(logs)))
	return logs, nil
}

// MarkWarningResolved marks a warning as resolved
func (r *WarningLogRepository) MarkWarningResolved(warningID string) error {
	logger.Debug("WarningLogRepository.MarkWarningResolved",
		zap.String("warning_id", warningID))

	result := r.db.Model(&model.WarningLog{}).
		Where("warning_id = ?", warningID).
		Update("status", "resolved")

	if result.Error != nil {
		logger.Error("Failed to mark warning as resolved",
			zap.String("warning_id", warningID),
			zap.Error(result.Error))
		return result.Error
	}

	logger.Debug("Warning marked as resolved",
		zap.String("warning_id", warningID))
	return nil
}
