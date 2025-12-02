package migration

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigrateToMultiSubscription migrates existing data from single subscription to multi-subscription model
// This migration:
// 1. Checks if migration is needed (if UserID column exists in todos table)
// 2. For each todo, finds the corresponding subscription
// 3. Updates todos to use SubscriptionID instead of UserID
// 4. Creates default subscriptions for todos without valid subscriptions
func MigrateToMultiSubscription(db *gorm.DB) error {
	logger.Info("Starting migration to multi-subscription model")

	// Check if migration is needed
	if !db.Migrator().HasColumn(&model.Todo{}, "user_id") {
		logger.Info("Migration already completed (user_id column not found in todos table)")
		return nil
	}

	// Check if SubscriptionID column exists
	if !db.Migrator().HasColumn(&model.Todo{}, "subscription_id") {
		logger.Info("Adding subscription_id column to todos table")
		if err := db.Migrator().AddColumn(&model.Todo{}, "SubscriptionID"); err != nil {
			return fmt.Errorf("failed to add subscription_id column: %w", err)
		}
	}

	// Define a temporary struct to read old todo data
	type OldTodo struct {
		ID     uint
		UserID uint
	}

	// Get all todos that need migration (where subscription_id is NULL)
	var oldTodos []OldTodo
	if err := db.Table("todos").
		Select("id, user_id").
		Where("subscription_id IS NULL OR subscription_id = 0").
		Scan(&oldTodos).Error; err != nil {
		return fmt.Errorf("failed to query todos for migration: %w", err)
	}

	if len(oldTodos) == 0 {
		logger.Info("No todos need migration")
		return finalizeMigration(db)
	}

	logger.Info("Found todos to migrate", zap.Int("count", len(oldTodos)))

	// Group todos by UserID for batch processing
	todosByUser := make(map[uint][]uint) // userID -> []todoID
	for _, todo := range oldTodos {
		todosByUser[todo.UserID] = append(todosByUser[todo.UserID], todo.ID)
	}

	// Process each user
	migratedCount := 0
	defaultSubsCreated := 0

	for userID, todoIDs := range todosByUser {
		// Find user's active subscription
		var subscription model.Subscription
		err := db.Where("user_id = ? AND active = ?", userID, true).
			Order("created_at DESC").
			First(&subscription).Error

		if err == gorm.ErrRecordNotFound {
			// No active subscription, try to find any subscription
			err = db.Where("user_id = ?", userID).
				Order("created_at DESC").
				First(&subscription).Error
		}

		if err == gorm.ErrRecordNotFound {
			// No subscription found, create a default one
			logger.Warn("No subscription found for user, creating default subscription",
				zap.Uint("user_id", userID))

			subscription = model.Subscription{
				UserID:       userID,
				City:         "默认",
				ReminderTime: "08:00",
				Active:       false, // Set to inactive as it's a default subscription
			}

			if err := db.Create(&subscription).Error; err != nil {
				logger.Error("Failed to create default subscription",
					zap.Uint("user_id", userID),
					zap.Error(err))
				continue
			}
			defaultSubsCreated++
		} else if err != nil {
			logger.Error("Failed to find subscription for user",
				zap.Uint("user_id", userID),
				zap.Error(err))
			continue
		}

		// Update todos to use this subscription
		result := db.Table("todos").
			Where("id IN ?", todoIDs).
			Update("subscription_id", subscription.ID)

		if result.Error != nil {
			logger.Error("Failed to update todos",
				zap.Uint("user_id", userID),
				zap.Uint("subscription_id", subscription.ID),
				zap.Error(result.Error))
			continue
		}

		migratedCount += int(result.RowsAffected)
		logger.Debug("Migrated todos for user",
			zap.Uint("user_id", userID),
			zap.Uint("subscription_id", subscription.ID),
			zap.Int("todo_count", int(result.RowsAffected)))
	}

	logger.Info("Migration completed successfully",
		zap.Int("migrated_todos", migratedCount),
		zap.Int("default_subscriptions_created", defaultSubsCreated))

	return finalizeMigration(db)
}

// finalizeMigration performs final cleanup steps
func finalizeMigration(db *gorm.DB) error {
	// Verify all todos have valid subscription_id
	var orphanCount int64
	if err := db.Table("todos").
		Where("subscription_id IS NULL OR subscription_id = 0").
		Count(&orphanCount).Error; err != nil {
		return fmt.Errorf("failed to verify migration: %w", err)
	}

	if orphanCount > 0 {
		logger.Warn("Found orphan todos after migration",
			zap.Int64("count", orphanCount))
		// Don't fail the migration, just warn
	}

	// Drop user_id column if it exists
	if db.Migrator().HasColumn(&model.Todo{}, "user_id") {
		logger.Info("Dropping user_id column from todos table")
		if err := db.Migrator().DropColumn(&model.Todo{}, "user_id"); err != nil {
			logger.Warn("Failed to drop user_id column (non-critical)",
				zap.Error(err))
			// Don't fail the migration if we can't drop the column
		}
	}

	logger.Info("Migration finalization completed")
	return nil
}
