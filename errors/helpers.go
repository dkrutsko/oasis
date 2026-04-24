package errors

import (
	"log/slog"
	"strconv"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

func Bool(key string, val bool) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.BoolValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Int(key string, val int) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Int8(key string, val int8) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Int16(key string, val int16) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Int32(key string, val int32) slog.Attr {

	v := int64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Int64(key string, val int64) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Byte(key string, val byte) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Uint(key string, val uint) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Uint8(key string, val uint8) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Uint16(key string, val uint16) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Uint32(key string, val uint32) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Uint64(key string, val uint64) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Uintptr(key string, val uintptr) slog.Attr {

	v := uint64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Uint64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Float32(key string, val float32) slog.Attr {

	v := float64(val)
	return slog.Attr{
		Key:   key,
		Value: slog.Float64Value(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Float64(key string, val float64) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.Float64Value(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

func String(key string, val string) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Rune(key string, val rune) slog.Attr {

	v := string(val)
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(v),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Time(key string, val time.Time) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.TimeValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Duration(key string, val time.Duration) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.DurationValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func Error(key string, err error) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: expandError(err),
	}
}

////////////////////////////////////////////////////////////////////////////////

func Any(key string, val any) slog.Attr {

	return slog.Attr{
		Key:   key,
		Value: slog.AnyValue(val),
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func SkipFrames(n int) slog.Attr {

	v := skipFrames{n}
	return slog.Attr{
		Key:   "",
		Value: slog.AnyValue(v),
	}
}
