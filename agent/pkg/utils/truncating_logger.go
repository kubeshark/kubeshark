package utils

import (
	"context"
	"fmt"
	"time"

	loggerShared "github.com/up9inc/mizu/shared/logger"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// TruncatingLogger implements the gorm logger.Interface interface. Its purpose is to act as gorm's logger while truncating logs to a max of 50 characters to minimise the performance impact
type TruncatingLogger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func (truncatingLogger *TruncatingLogger) LogMode(logLevel logger.LogLevel) logger.Interface {
	truncatingLogger.LogLevel = logLevel
	return truncatingLogger
}

func (truncatingLogger *TruncatingLogger) Info(_ context.Context, message string, __ ...interface{}) {
	if truncatingLogger.LogLevel < logger.Info {
		return
	}
	loggerShared.Log.Errorf("gorm info: %.150s", message)
}

func (truncatingLogger *TruncatingLogger) Warn(_ context.Context, message string, __ ...interface{}) {
	if truncatingLogger.LogLevel < logger.Warn {
		return
	}
	loggerShared.Log.Errorf("gorm warning: %.150s", message)
}

func (truncatingLogger *TruncatingLogger) Error(_ context.Context, message string, __ ...interface{}) {
	if truncatingLogger.LogLevel < logger.Error {
		return
	}
	loggerShared.Log.Errorf("gorm error: %.150s", message)
}

func (truncatingLogger *TruncatingLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if truncatingLogger.LogLevel == logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	if err != nil {
		sql, rows := fc() // copied into every condition as this is a potentially heavy operation best done only when necessary
		truncatingLogger.Error(ctx, fmt.Sprintf("Error in %s: %v - elapsed: %fs affected rows: %d, sql: %s", utils.FileWithLineNum(), err, elapsed.Seconds(), rows, sql))
	} else if truncatingLogger.LogLevel >= logger.Warn && elapsed > truncatingLogger.SlowThreshold {
		sql, rows := fc()
		truncatingLogger.Warn(ctx, fmt.Sprintf("Slow sql query - elapse: %fs rows: %d, sql: %s", elapsed.Seconds(), rows, sql))
	} else if truncatingLogger.LogLevel >= logger.Info {
		sql, rows := fc()
		truncatingLogger.Info(ctx, fmt.Sprintf("Sql query - elapse: %fs rows: %d, sql: %s", elapsed.Seconds(), rows, sql))
	}
}
