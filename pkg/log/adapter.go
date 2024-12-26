package log

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/hashicorp/go-retryablehttp"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)

type Adapter struct {
	logger *slog.Logger
	level  slog.Level
}

func NewAdapter(logger *slog.Logger, level slog.Level) *Adapter {
	result := &Adapter{
		logger: logger,
		level:  level,
	}

	return result
}

func (a *Adapter) Enabled() bool {
	return a.logger.Enabled(context.Background(), a.level)
}

// Error satisfies the API contact of [retryablehttp#LeveledLogger]
func (a *Adapter) Error(msg string, v ...interface{}) {
	a.logger.Error(msg, v...)
}

// Info satisfies the API contact of [retryablehttp#LeveledLogger]
func (a *Adapter) Info(msg string, v ...interface{}) {
	a.logger.Info(msg, v...)
}

// Debug satisfies the API contact of [retryablehttp#LeveledLogger]
func (a *Adapter) Debug(msg string, v ...interface{}) {
	a.logger.Debug(msg, v...)
}

// Warn satisfies the API contact of [retryablehttp#LeveledLogger]
func (a *Adapter) Warn(msg string, v ...interface{}) {
	a.logger.Warn(msg, v...)
}

// Println satisfies the API contact of [promhttp#Logger]
func (a *Adapter) Println(v ...interface{}) {
	line := fmt.Sprint(v...)
	a.logger.Log(context.Background(), a.level, line)
}

// Printf satisfies the API contact of [retryablehttp#Logger]
func (a *Adapter) Printf(msg string, v ...interface{}) {
	line := fmt.Sprintf(msg, v...)
	a.logger.Log(context.Background(), a.level, line)
}
