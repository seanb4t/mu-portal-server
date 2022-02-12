// Package log configures a new logger for an application.
package log

import (
	"go.uber.org/zap"
	zapadaptor "logur.dev/adapter/zap"
	"logur.dev/logur"
)

// NewLogger creates a new logger.
func NewLogger(config Config) logur.LoggerFacade {
	logger, _ := zap.NewDevelopment()

	// TODO: add config support, log format, level, etc.

	return zapadaptor.New(logger)
}
