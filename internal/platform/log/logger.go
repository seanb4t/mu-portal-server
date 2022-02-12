// Package log configures a new logger for an application.
package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	zapadaptor "logur.dev/adapter/zerolog"
	"logur.dev/logur"
	"os"
)

// NewLogger creates a new logger.
func NewLogger(config Config) logur.LoggerFacade {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFieldName = "ts"
	zerolog.MessageFieldName = "msg"

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// TODO: add config support, log format, level, etc.

	var logger zerolog.Logger
	if config.Format == "logfmt" {
		 logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		logger = zerolog.New(os.Stdout).
			With().Timestamp().
			Stack().
			Logger()
	}


	return zapadaptor.New(logger)
}
