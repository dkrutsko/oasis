package logger

import (
	"log/slog"
	"os"

	"github.com/fatih/color"
)

////////////////////////////////////////////////////////////////////////////////

// LevelConfig defines the output configuration for a single log level.
type LevelConfig struct {
	// Writer is the output destination for this level.
	Writer *os.File

	// Prefix is the string prepended to each log message.
	// It can be set to an empty string to disable the prefix.
	Prefix string

	// Color is the color applied to output for this level.
	Color *color.Color
}

////////////////////////////////////////////////////////////////////////////////

// Options configures the behavior and appearance of a logger instance.
type Options struct {
	// Level is the minimum log level that will be output.
	Level slog.Level

	// Json controls whether output is formatted as JSON lines.
	Json bool

	// Dbg configures the output for debug-level messages.
	Dbg LevelConfig

	// Info configures the output for info-level messages.
	Info LevelConfig

	// Warn configures the output for warning-level messages.
	Warn LevelConfig

	// Err configures the output for error-level messages.
	Err LevelConfig
}

////////////////////////////////////////////////////////////////////////////////

// NewOptions returns an `Options` with sensible defaults for logging.
func NewOptions() *Options {

	return &Options{
		Level: slog.LevelInfo,
		Json:  false,

		Dbg: LevelConfig{
			Writer: os.Stdout,
			Prefix: "[Debug] ",
			Color:  color.New(color.FgHiBlack),
		},

		Info: LevelConfig{
			Writer: os.Stdout,
			Prefix: "[Info] ",
			Color:  color.New(),
		},

		Warn: LevelConfig{
			Writer: os.Stderr,
			Prefix: "[Warn] ",
			Color:  color.New(color.FgYellow),
		},

		Err: LevelConfig{
			Writer: os.Stderr,
			Prefix: "[Error] ",
			Color:  color.New(color.FgRed),
		},
	}
}
