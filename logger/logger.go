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
)

////////////////////////////////////////////////////////////////////////////////

var (
	instances     = make(map[string]*Logger)
	instancesLock sync.RWMutex
)

////////////////////////////////////////////////////////////////////////////////

// Logger wraps `slog.Logger` with additional metadata and functions.
type Logger struct {
	*slog.Logger
	level slog.Level
	json  bool
}

////////////////////////////////////////////////////////////////////////////////

// GetLevel returns the minimum log level configured for this logger.
func (l *Logger) GetLevel() slog.Level {
	return l.level
}

////////////////////////////////////////////////////////////////////////////////

// IsJson returns whether this logger is configured to output JSON.
func (l *Logger) IsJson() bool {
	return l.json
}

////////////////////////////////////////////////////////////////////////////////

// GetHandler returns the underlying `LogHandler` for this logger.
func (l *Logger) GetHandler() *LogHandler {
	return l.Handler().(*LogHandler)
}

////////////////////////////////////////////////////////////////////////////////

// New creates a new `slog.Logger` from the provided options. The returned
// logger routes messages to the configured writers with color-coded output
// per level. If the JSON option is enabled, output is passed to the default
// JSON handler instead.
func New(opts *Options) *Logger {

	handlerOpts := &slog.HandlerOptions{
		Level: opts.Level,
	}

	var handler slog.Handler
	if opts.Json {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	cwDbg := &ColorWriter{w: opts.Dbg.Writer, color: opts.Dbg.Color}
	cwInfo := &ColorWriter{w: opts.Info.Writer, color: opts.Info.Color}
	cwWarn := &ColorWriter{w: opts.Warn.Writer, color: opts.Warn.Color}
	cwErr := &ColorWriter{w: opts.Err.Writer, color: opts.Err.Color}

	return &Logger{
		level: opts.Level,
		json:  opts.Json,
		Logger: slog.New(
			&LogHandler{
				Handler: handler,
				Dbg:     log.New(cwDbg, opts.Dbg.Prefix, 0),
				Info:    log.New(cwInfo, opts.Info.Prefix, 0),
				Warn:    log.New(cwWarn, opts.Warn.Prefix, 0),
				Err:     log.New(cwErr, opts.Err.Prefix, 0),
			},
		),
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetLogger gets the `Logger` registered under the given key. Returns nil
// if no logger is registered with that key.
func GetLogger(key string) *Logger {

	// Lock when reading
	instancesLock.RLock()
	defer instancesLock.RUnlock()

	return instances[key]
}

////////////////////////////////////////////////////////////////////////////////

// SetLogger registers a `Logger` under the given key. If the logger is nil,
// the key is removed. All registered loggers receive log messages
// independently.
func SetLogger(key string, l *Logger) {

	// Lock when setting
	instancesLock.Lock()
	defer instancesLock.Unlock()

	if l == nil {
		// Delete the instance
		delete(instances, key)
	} else {
		instances[key] = l
	}
}

////////////////////////////////////////////////////////////////////////////////

func internalLog(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {

	// Lock when reading
	instancesLock.RLock()
	defer instancesLock.RUnlock()

	// Nothing to log to
	if len(instances) == 0 {
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

	// Log to all registered loggers
	for _, inst := range instances {
		inst.LogAttrs(ctx, level, msg, res...)
	}
}

////////////////////////////////////////////////////////////////////////////////

// Dbg logs at `slog.LevelDebug`.
func Dbg(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelDebug, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// DbgContext logs at `slog.LevelDebug` with the given context.
func DbgContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelDebug, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// Info logs at `slog.LevelInfo`.
func Info(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelInfo, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// InfoContext logs at `slog.LevelInfo` with the given context.
func InfoContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelInfo, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// Warn logs at `slog.LevelWarn`.
func Warn(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelWarn, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// WarnContext logs at `slog.LevelWarn` with the given context.
func WarnContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelWarn, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// Err logs at `slog.LevelError`.
func Err(msg string, attrs ...slog.Attr) {
	internalLog(context.Background(), slog.LevelError, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// ErrContext logs at `slog.LevelError` with the given context.
func ErrContext(ctx context.Context, msg string, attrs ...slog.Attr) {
	internalLog(ctx, slog.LevelError, msg, attrs...)
}

////////////////////////////////////////////////////////////////////////////////

// Begin logs at `slog.LevelInfo` the start of some operation and returns
// the current time for use later to provide to the `End` function.
func Begin() time.Time {
	internalLog(context.Background(), slog.LevelInfo, "begin")

	// Record runtime
	return time.Now()
}

////////////////////////////////////////////////////////////////////////////////

// BeginContext logs at `slog.LevelInfo` with the given context the start of
// some operation and returns the current time for use later to provide to
// the `End` function.
func BeginContext(ctx context.Context) time.Time {
	internalLog(ctx, slog.LevelInfo, "begin")

	// Record runtime
	return time.Now()
}

////////////////////////////////////////////////////////////////////////////////

// End logs at `slog.LevelInfo` the end of some operation given a previous time.
func End(t time.Time) {
	internalLog(context.Background(), slog.LevelInfo, "end", Elapsed("elapsed", t))
}

////////////////////////////////////////////////////////////////////////////////

// EndContext logs at `slog.LevelInfo` with the given context the end of some
// operation given a previous time.
func EndContext(ctx context.Context, t time.Time) {
	internalLog(ctx, slog.LevelInfo, "end", Elapsed("elapsed", t))
}
