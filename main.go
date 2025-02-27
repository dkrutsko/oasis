package main

import (
	"log/slog"
	"os"

	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	setupLogger()
}

////////////////////////////////////////////////////////////////////////////////

func setupLogger() {

	var level slog.Level
	if os.Getenv("OASIS_LOG_DEBUG") == "" {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	var outputJson bool
	if os.Getenv("OASIS_LOG_JSON") == "" {
		outputJson = false
	} else {
		outputJson = true
	}

	logger.SetLogger(
		// Create a logger for the app
		logger.New(level, outputJson),
	)
}
