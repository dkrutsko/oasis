package logger

import (
	"log/slog"
	"strconv"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

// Bool returns an `Attr` for a bool.
func Bool(key string, val bool) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.BoolValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Int converts an int to an int64 and returns an `Attr`.
func Int(key string, val int) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Int8 converts an int8 to an int64 and returns an `Attr`.
func Int8(key string, val int8) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Int16 converts an int16 to an int64 and returns an `Attr`.
func Int16(key string, val int16) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Int32 converts an int32 to an int64 and returns an `Attr`.
func Int32(key string, val int32) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Int64 returns an `Attr` for an int64.
func Int64(key string, val int64) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Byte converts a byte to a uint64 and returns an `Attr`.
func Byte(key string, val byte) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Uint converts a uint to a uint64 and returns an `Attr`.
func Uint(key string, val uint) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Uint8 converts a uint8 to a uint64 and returns an `Attr`.
func Uint8(key string, val uint8) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Uint16 converts a uint16 to a uint64 and returns an `Attr`.
func Uint16(key string, val uint16) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Uint32 converts a uint32 to a uint64 and returns an `Attr`.
func Uint32(key string, val uint32) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Uint64 returns an `Attr` for a uint64.
func Uint64(key string, val uint64) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Uintptr converts a uintptr to a uint64 and returns an `Attr`.
func Uintptr(key string, val uintptr) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Float32 converts a float32 to a float64 and returns an `Attr`.
func Float32(key string, val float32) slog.Attr {

	v := float64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Float64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Float64 returns an `Attr` for a float64.
func Float64(key string, val float64) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.Float64Value(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// String returns an `Attr` for a string.
func String(key string, val string) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Rune converts a rune to a string and returns an `Attr`.
func Rune(key string, val rune) slog.Attr {

	v := string(val)
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Time returns an `Attr` for a `time.Time`. It discards the monotonic portion.
func Time(key string, val time.Time) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.TimeValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Duration returns an `Attr` for a `time.Duration`.
func Duration(key string, val time.Duration) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.DurationValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Elapsed returns an `Attr` for the time elapsed since `val` in a
// formatted way.
func Elapsed(key string, val time.Time) slog.Attr {

	// Calculate elapsed time
	elapsed := time.Since(val)

	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(elapsed.Milliseconds()),
	}
}

////////////////////////////////////////////////////////////////////////////////

func expandError(err error) slog.Value {

	//----------------------------------------------------------------------------//

	// Nil error
	if err == nil {
		return slog.AnyValue(nil)
	}

	//----------------------------------------------------------------------------//

	// Made with errors.Join
	uj, ok := err.(interface {
		Unwrap() []error
	})

	if ok {
		// Unwrap the errors
		errs := uj.Unwrap()

		var res []slog.Attr
		// Loop through all errors
		for i, err := range errs {

			res = append(
				res,
				slog.Attr{
					Key:   strconv.Itoa(i),
					Value: expandError(err),
				},
			)
		}

		// Return result as group value
		return slog.GroupValue(res...)
	}

	//----------------------------------------------------------------------------//

	var wrapped error
	// If an error is wrapped
	ue, ok := err.(interface {
		Unwrap() error
	})

	if ok {
		// Unwrap the error
		wrapped = ue.Unwrap()
	}

	//----------------------------------------------------------------------------//

	// If contains attributes
	a, ok := err.(interface {
		Attributes() []slog.Attr
	})

	if ok {
		var res []slog.Attr
		// Retrieve attributes
		attrs := a.Attributes()

		// Add message
		res = append(
			res,
			slog.Attr{
				Key:   "msg",
				Value: slog.StringValue(err.Error()),
			},
		)

		// Handle wrapped
		if wrapped != nil {

			res = append(
				res,
				slog.Attr{
					Key:   "wrapped",
					Value: expandError(wrapped),
				},
			)
		}

		// Add remaining attributes
		res = append(res, attrs...)

		// Return result as group value
		return slog.GroupValue(res...)
	}

	//----------------------------------------------------------------------------//

	if wrapped != nil {

		// Combine into a group
		return slog.GroupValue(

			slog.Attr{
				Key:   "msg",
				Value: slog.StringValue(err.Error()),
			},

			slog.Attr{
				Key:   "wrapped",
				Value: expandError(wrapped),
			},
		)
	}

	//----------------------------------------------------------------------------//

	return slog.StringValue(err.Error())

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Error returns an `Attr` for an error. If the error is wrapped, it will log
// both the outer error message and the wrapped error message recursively.
// If the error is nil, the `Attr` will have a nil value.
func Error(key string, err error) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: expandError(err),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Any returns an `Attr` for the supplied value.
func Any(key string, val any) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.AnyValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Group returns an `Attr` for a group. Use `Group` to collect several
// key-value pairs under a single key on a log line, or as the result of
// `LogValue` to log a single value as multiple `Attr` values.
func Group(key string, args ...slog.Attr) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.GroupValue(args...),
	}
}

////////////////////////////////////////////////////////////////////////////////

type skipFrames struct {
	frames int
}

////////////////////////////////////////////////////////////////////////////////

// SkipFrames returns a special `Attr` for skipping stack frames during output.
func SkipFrames(n int) slog.Attr {

	v := skipFrames{n}
	return slog.Attr{
		Key:   "",
		Value: slog.AnyValue(v),
	}
}
