package config

import (
	"bytes"
	"log/slog"

	"github.com/fatih/color"

	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Splash holds the precomputed values displayed at startup.
type Splash struct {
	// cfg is the parsed application configuration.
	cfg *Config

	// version is the resolved build version string.
	version string
}

////////////////////////////////////////////////////////////////////////////////

// SlogAttrs returns structured log attributes for the splash configuration.
func (s *Splash) SlogAttrs() []slog.Attr {

	return []slog.Attr{
		logger.String("version", s.version),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Format renders the splash banner and configuration summary as a string.
func (s *Splash) Format(useColor bool) string {

	yellow := color.New(color.FgHiYellow).Add(color.Bold)
	whiteL := color.New(color.FgHiWhite)
	whiteB := color.New(color.FgHiWhite).Add(color.Bold)

	if !useColor {
		yellow.DisableColor()
		whiteL.DisableColor()
		whiteB.DisableColor()
	}

	var out bytes.Buffer

	out.WriteString(whiteL.Sprintf("  Version: "))
	out.WriteString(whiteB.Sprintf("%s\n", s.version))

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////

// GetSplash builds a `Splash` from the given `Config`.
func GetSplash(cfg *Config) *Splash {

	out := &Splash{
		cfg: cfg,
	}

	// Attempt to retrieve the version
	out.version = GetVersion().String()
	return out
}
