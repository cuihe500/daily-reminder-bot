package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormAdapter adapts zap logger to GORM logger interface
type GormAdapter struct {
	logger                    *zap.Logger
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewGormAdapter creates a new GORM logger adapter
func NewGormAdapter(logger *zap.Logger, slowThreshold time.Duration) gormlogger.Interface {
	return &GormAdapter{
		logger:                    logger,
		logLevel:                  gormlogger.Info,
		slowThreshold:             slowThreshold,
		ignoreRecordNotFoundError: true,
	}
}

func (l *GormAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func (l *GormAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.logger.Sugar().Infof(msg, data...)
	}
}

func (l *GormAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.logger.Sugar().Warnf(msg, data...)
	}
}

func (l *GormAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.logger.Sugar().Errorf(msg, data...)
	}
}

func (l *GormAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	}

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.logger.Error("database error", append(fields, zap.Error(err))...)
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormlogger.Warn:
		l.logger.Warn("slow query", fields...)
	case l.logLevel >= gormlogger.Info:
		l.logger.Debug("query executed", fields...)
	}
}
