package logger

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
)

////////////////////////////////////////////////////////////////////////////////

type ColorWriter struct {
	w     *os.File
	color *color.Color
}

////////////////////////////////////////////////////////////////////////////////

func (cw *ColorWriter) Write(p []byte) (n int, err error) {
	return cw.w.Write([]byte(cw.color.Sprint(string(p))))
}

////////////////////////////////////////////////////////////////////////////////

type LogHandler struct {
	slog.Handler
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

////////////////////////////////////////////////////////////////////////////////

func (log *LogHandler) prependSrc(a slog.Attr, result *[]string) bool {

	// Check for source and make sure it's a kind of group
	if a.Key == "src" && a.Value.Kind() == slog.KindGroup {

		var fn string = ""
		var ln int64 = -1

		// Iterate through attrs in the group
		for _, sub := range a.Value.Group() {

			// Grab function name
			if sub.Key == "func" {

				parts := strings.Split(
					sub.Value.String(),
					"/",
				)

				fn = parts[len(parts)-1]
			}

			// Grab function line
			if sub.Key == "line" {
				ln = sub.Value.Int64()
			}
		}

		// Prepend all the parts
		if fn != "" && ln >= 0 {

			// Prepend parts
			*result = append(
				[]string{fmt.Sprintf("[%s:%d]", fn, ln)},
				*result...,
			)

			return true
		}
	}

	return false
}

////////////////////////////////////////////////////////////////////////////////

func (log *LogHandler) formatInner(v slog.Value) string {

	// Try parsing value as a group
	if v.Kind() == slog.KindGroup {

		var result []string
		// Loop through the attributes
		for _, a := range v.Group() {

			// Check if src can be prepended
			if !log.prependSrc(a, &result) {

				// Format other attributes recursively with same filtering
				result = append(result, a.Key+"="+log.formatInner(a.Value))
			}
		}

		return "[" + strings.Join(result, " ") + "]"
	}

	// Try parsing value as a string
	if v.Kind() == slog.KindString {

		// Wrap the string in quotations
		return "\"" + v.String() + "\""
	}

	// Default format
	return v.String()
}

////////////////////////////////////////////////////////////////////////////////

func (log *LogHandler) formatMessage(r slog.Record) string {

	result := []string{r.Message}
	// Handle attrs when writing text
	r.Attrs(func(a slog.Attr) bool {

		// Check if src can be prepended
		if !log.prependSrc(a, &result) {

			// Format the other attributes and filter out extra metadata
			result = append(result, a.Key+"="+log.formatInner(a.Value))
		}

		return true
	})

	return strings.Join(result, " ")
}

////////////////////////////////////////////////////////////////////////////////

func (log *LogHandler) Handle(ctx context.Context, r slog.Record) error {

	// Pass to the default embedded handler if writing JSON
	if handler, ok := log.Handler.(*slog.JSONHandler); ok {
		return handler.Handle(ctx, r)
	}

	// Format current message
	msg := log.formatMessage(r)

	switch r.Level {
	case slog.LevelDebug:
		log.Debug.Println(msg)

	case slog.LevelInfo:
		log.Info.Println(msg)

	case slog.LevelWarn:
		log.Warn.Println(msg)

	case slog.LevelError:
		log.Error.Println(msg)
	}

	return nil
}
