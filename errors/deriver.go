package errors

import (
	"log/slog"
)

////////////////////////////////////////////////////////////////////////////////

type Deriver struct {
	attrs []slog.Attr
}

////////////////////////////////////////////////////////////////////////////////

func With(attrs ...slog.Attr) Deriver {

	// Create a copy of the attributes
	a := make([]slog.Attr, len(attrs))
	copy(a, attrs)

	return Deriver{
		attrs: a,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (d Deriver) With(attrs ...slog.Attr) Deriver {

	final := make([]slog.Attr, 0, len(d.attrs)+len(attrs))

	// Combine the two attribute sets
	final = append(final, d.attrs...)
	final = append(final, attrs...)

	return Deriver{
		attrs: final,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (d Deriver) New(msg string, attrs ...slog.Attr) error {

	final := make([]slog.Attr, 0, 1+len(d.attrs)+len(attrs))

	// Frame should point to the caller
	final = append(final, SkipFrames(1))

	// Combine the two attribute sets
	final = append(final, d.attrs...)
	final = append(final, attrs...)

	return New(msg, final...)
}
