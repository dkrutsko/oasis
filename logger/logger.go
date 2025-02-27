package logger

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
)

////////////////////////////////////////////////////////////////////////////////

var (
	instance     *slog.Logger
	instanceLock sync.RWMutex
)

////////////////////////////////////////////////////////////////////////////////

func New(level slog.Level, outputJson bool) *slog.Logger {

	options := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if outputJson {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	cwDebug := &ColorWriter{w: os.Stdout, color: color.New(color.FgHiBlack)}
	cwInfo := &ColorWriter{w: os.Stdout, color: color.New()}
	cwWarn := &ColorWriter{w: os.Stderr, color: color.New(color.FgYellow)}
	cwError := &ColorWriter{w: os.Stderr, color: color.New(color.FgRed)}

	result := slog.New(&LogHandler{
		Handler: handler,

		Debug: log.New(cwDebug, "", 0),
		Info:  log.New(cwInfo, "", 0),
		Warn:  log.New(cwWarn, "", 0),
		Error: log.New(cwError, "", 0),
	})

	return result
}

////////////////////////////////////////////////////////////////////////////////

func GetLogger() *slog.Logger {

	// Lock when reading
	instanceLock.RLock()
	defer instanceLock.RUnlock()

	return instance
}

////////////////////////////////////////////////////////////////////////////////

func SetLogger(l *slog.Logger) {

	// Lock when setting
	instanceLock.Lock()
	defer instanceLock.Unlock()

	instance = l
}

////////////////////////////////////////////////////////////////////////////////

func internalLog(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {

	// Lock when reading
	instanceLock.RLock()
	defer instanceLock.RUnlock()

	// If instance exists
	if instance == nil {
		return
	}

	// Final attributes
	var res []slog.Attr

	// Default frame skip
	skip := 2

	// Process attributes values
	for _, attr := range attrs {

		// Check for any special config flags
		switch v := attr.Value.Any().(type) {

		case skipFrames:
			// Add some frame skips
			skip += v.frames

		default:
			// Keep other attributes
			res = append(res, attr)
		}
	}

	{
		// Retrieve the specified call stack frame
		pc, file, line, ok := runtime.Caller(skip)

		if ok {
			// Retrieve the full name of the caller
			funcName := runtime.FuncForPC(pc).Name()

			// Extract just the file name
			fileName := filepath.Base(file)

			// Put attributes into a group
			srcAttr := slog.Group(
				"src",
				String("file", fileName),
				String("func", funcName),
				Int("line", line),
			)

			// Prepend source to attributes
			res = append([]slog.Attr{srcAttr}, res...)
		}
	}

	// Pass to the parent LogAttrs with source
	instance.LogAttrs(ctx, level, msg, res...)
}

////////////////////////////////////////////////////////////////////////////////

func Dbg(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelDebug, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func DbgContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelDebug, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func Info(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelInfo, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelInfo, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func Warn(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelWarn, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelWarn, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func Err(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelError, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func ErrContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelError, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

func Begin() time.Time {
	internalLog(context.Background(), slog.LevelInfo, "begin")

	// Record runtime
	return time.Now()
}

////////////////////////////////////////////////////////////////////////////////

func BeginContext(ctx context.Context) time.Time {
	internalLog(ctx, slog.LevelInfo, "begin")

	// Record runtime
	return time.Now()
}

////////////////////////////////////////////////////////////////////////////////

func End(t time.Time) {
	internalLog(context.Background(), slog.LevelInfo, "end", Elapsed("elapsed", t))
}

////////////////////////////////////////////////////////////////////////////////

func EndContext(ctx context.Context, t time.Time) {
	internalLog(ctx, slog.LevelInfo, "end", Elapsed("elapsed", t))
}
