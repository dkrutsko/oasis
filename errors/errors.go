package errors

import (
	sysErrors "errors"
	"log/slog"
	"path/filepath"
	"runtime"
)

////////////////////////////////////////////////////////////////////////////////

type structuredError struct {
	msg   string
	attrs []slog.Attr
	stack []runtime.Frame
}

////////////////////////////////////////////////////////////////////////////////

func New(msg string, attrs ...slog.Attr) error {

	// Final attributes
	var res []slog.Attr

	// Default frame skip
	skip := 1

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

	trace := trace()
	// Check frame count
	if len(trace) > skip {

		// Grab stack frame
		frame := trace[skip]

		// Extract just the file name
		fileName := filepath.Base(frame.File)

		// Retrieve the function name
		funcName := frame.Function

		// Put attributes into a group
		srcAttr := slog.Group(
			"src",
			String("file", fileName),
			String("func", funcName),
			Int("line", frame.Line),
		)

		// Prepend source to attributes
		res = append([]slog.Attr{srcAttr}, res...)
	}

	// Prepare resulting error
	result := &structuredError{
		msg:   msg,
		attrs: res,
		stack: trace,
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

func (e *structuredError) Error() string {
	return e.msg
}

////////////////////////////////////////////////////////////////////////////////

func (e *structuredError) Is(target error) bool {
	t, ok := target.(*structuredError)
	if !ok {
		return false
	}
	return e.msg == t.msg
}

////////////////////////////////////////////////////////////////////////////////

func Unwrap(err error) error {

	// If has required method
	u, ok := err.(interface {
		Unwrap() error
	})

	if !ok {
		return nil
	}

	return u.Unwrap()
}

////////////////////////////////////////////////////////////////////////////////

func (e *structuredError) Attributes() []slog.Attr {
	return e.attrs
}

////////////////////////////////////////////////////////////////////////////////

func Attributes(err error) []slog.Attr {

	// If has required method
	a, ok := err.(interface {
		Attributes() []slog.Attr
	})

	if !ok {
		return nil
	}

	return a.Attributes()
}

////////////////////////////////////////////////////////////////////////////////

func (e *structuredError) StackTrace() []runtime.Frame {
	return e.stack
}

////////////////////////////////////////////////////////////////////////////////

func StackTrace(err error) []runtime.Frame {

	// If has required method
	u, ok := err.(interface {
		StackTrace() []runtime.Frame
	})

	if !ok {
		return nil
	}

	return u.StackTrace()
}

////////////////////////////////////////////////////////////////////////////////

func Join(errs ...error) error {

	// If only one error found
	var first error

	count := 0
	// Iterate all the errors
	for _, err := range errs {

		if err != nil {
			first = err
			count++
		}
	}

	if count == 0 {
		return nil
	}

	if count == 1 {
		return first
	}

	// Use default implementation
	return sysErrors.Join(errs...)
}

////////////////////////////////////////////////////////////////////////////////

func Is(err, target error) bool {
	return sysErrors.Is(err, target)
}

////////////////////////////////////////////////////////////////////////////////

func As(err error, target any) bool {
	return sysErrors.As(err, target)
}

////////////////////////////////////////////////////////////////////////////////

func trace() []runtime.Frame {

	var pcs [512]uintptr
	// Retrieve current call stack
	// Exclude the runtime.Callers
	n := runtime.Callers(1, pcs[:])

	// Convert program counters into frames
	frames := runtime.CallersFrames(pcs[:n])
	result := make([]runtime.Frame, 0, n)

	// Skip extra frame retrieved just to
	// make sure runtime.sigpanic special
	// case is handled
	frame, more := frames.Next()

	for more {
		// Retrieve the next frame
		frame, more = frames.Next()
		result = append(result, frame)
	}

	return result
}
