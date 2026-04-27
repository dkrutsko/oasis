package math

import (
	"fmt"
	sysMath "math"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

var (
	// ColorZero represents a color with all components set to zero.
	ColorZero = Color{0, 0, 0, 0}

	// ColorBlack represents opaque black.
	ColorBlack = Color{0, 0, 0, 1}

	// ColorWhite represents opaque white.
	ColorWhite = Color{1, 1, 1, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Color represents an RGBA color with float64 components in the range [0, 1].
type Color struct {
	R float64
	G float64
	B float64
	A float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the color.
func (c Color) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f]", c.R, c.G, c.B, c.A)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all components are zero.
func (c Color) IsZero() bool {

	return c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0
}

////////////////////////////////////////////////////////////////////////////////

// IsHdr returns whether any component is outside the [0, 1]
// range.
func (c Color) IsHdr() bool {

	return c.R < 0 || c.R > 1 ||
		c.G < 0 || c.G > 1 ||
		c.B < 0 || c.B > 1 ||
		c.A < 0 || c.A > 1
}

////////////////////////////////////////////////////////////////////////////////

// Clamp returns the color with all components restricted to the
// [0, 1] range.
func (c Color) Clamp() Color {

	return Color{
		Clamp(c.R, 0, 1),
		Clamp(c.G, 0, 1),
		Clamp(c.B, 0, 1),
		Clamp(c.A, 0, 1),
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetBrightness returns the WCAG perceived brightness in the
// range [0, 1]. Based on the YIQ color space weighting.
func (c Color) GetBrightness() float64 {

	return c.R*0.299 + c.G*0.587 + c.B*0.114
}

////////////////////////////////////////////////////////////////////////////////

// GetLuminance returns the WCAG 2.0 relative luminance in the
// range [0, 1]. Uses the sRGB linearization formula.
func (c Color) GetLuminance() float64 {

	lr := c.R / 12.92
	if c.R > 0.04045 {
		lr = sysMath.Pow((c.R+0.055)/1.055, 2.4)
	}

	lg := c.G / 12.92
	if c.G > 0.04045 {
		lg = sysMath.Pow((c.G+0.055)/1.055, 2.4)
	}

	lb := c.B / 12.92
	if c.B > 0.04045 {
		lb = sysMath.Pow((c.B+0.055)/1.055, 2.4)
	}

	return 0.2126*lr + 0.7152*lg + 0.0722*lb
}

////////////////////////////////////////////////////////////////////////////////

// IsDark returns whether the color is perceived as dark.
func (c Color) IsDark() bool {

	return c.R*0.299+c.G*0.587+c.B*0.114 < 0.5
}

////////////////////////////////////////////////////////////////////////////////

// IsLight returns whether the color is perceived as light.
func (c Color) IsLight() bool {

	return c.R*0.299+c.G*0.587+c.B*0.114 >= 0.5
}

////////////////////////////////////////////////////////////////////////////////

// Lighten returns the color with lightness increased by the given amount
// in HSL space. The amount is added directly to the lightness value, e.g.
// 0.1 shifts lightness from 0.3 to 0.4.
func (c Color) Lighten(amount float64) Color {

	h, s, l, a := c.ToHsl()
	l = Clamp(l+amount, 0, 1)
	return ColorFromHsl(h, s, l, a)
}

////////////////////////////////////////////////////////////////////////////////

// Darken returns the color with lightness decreased by the given amount
// in HSL space. The amount is subtracted directly from the lightness value,
// e.g. 0.1 shifts lightness from 0.4 to 0.3.
func (c Color) Darken(amount float64) Color {

	h, s, l, a := c.ToHsl()
	l = Clamp(l-amount, 0, 1)
	return ColorFromHsl(h, s, l, a)
}

////////////////////////////////////////////////////////////////////////////////

// Saturate returns the color with saturation increased by the given amount
// in HSL space. The amount is added directly to the saturation value.
func (c Color) Saturate(amount float64) Color {

	h, s, l, a := c.ToHsl()
	s = Clamp(s+amount, 0, 1)
	return ColorFromHsl(h, s, l, a)
}

////////////////////////////////////////////////////////////////////////////////

// Desaturate returns the color with saturation decreased by the given amount
// in HSL space. The amount is subtracted directly from the saturation value.
func (c Color) Desaturate(amount float64) Color {

	h, s, l, a := c.ToHsl()
	s = Clamp(s-amount, 0, 1)
	return ColorFromHsl(h, s, l, a)
}

////////////////////////////////////////////////////////////////////////////////

// Grayscale returns the fully desaturated color.
func (c Color) Grayscale() Color {

	h, _, l, a := c.ToHsl()
	return ColorFromHsl(h, 0, l, a)
}

////////////////////////////////////////////////////////////////////////////////

// Invert returns the color with RGB channels inverted.
// The alpha channel is preserved.
func (c Color) Invert() Color {

	return Color{1 - c.R, 1 - c.G, 1 - c.B, c.A}
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs linear interpolation toward the target color.
func (c Color) Lerp(target Color, amount float64) Color {

	return Color{
		c.R + (target.R-c.R)*amount,
		c.G + (target.G-c.G)*amount,
		c.B + (target.B-c.B)*amount,
		c.A + (target.A-c.A)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SmoothStep performs Hermite interpolation toward the target color
// with smoothing at the edges.
func (c Color) SmoothStep(target Color, amount float64) Color {

	amount = Clamp(amount, 0, 1)
	amount = amount * amount * (3 - 2*amount)

	return Color{
		c.R + (target.R-c.R)*amount,
		c.G + (target.G-c.G)*amount,
		c.B + (target.B-c.B)*amount,
		c.A + (target.A-c.A)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToRgba returns the color as 8-bit RGBA components in the range
// [0, 255]. Components are clamped to [0, 1] before conversion.
func (c Color) ToRgba() (r, g, b, a uint8) {

	return uint8(sysMath.Round(Clamp(c.R, 0, 1) * 255)),
		uint8(sysMath.Round(Clamp(c.G, 0, 1) * 255)),
		uint8(sysMath.Round(Clamp(c.B, 0, 1) * 255)),
		uint8(sysMath.Round(Clamp(c.A, 0, 1) * 255))
}

////////////////////////////////////////////////////////////////////////////////

// ToPacked returns the color as a packed ARGB uint32. Components
// are clamped to [0, 1] before conversion.
func (c Color) ToPacked() uint32 {

	r := uint32(sysMath.Round(Clamp(c.R, 0, 1) * 255))
	g := uint32(sysMath.Round(Clamp(c.G, 0, 1) * 255))
	b := uint32(sysMath.Round(Clamp(c.B, 0, 1) * 255))
	a := uint32(sysMath.Round(Clamp(c.A, 0, 1) * 255))

	return a<<24 | r<<16 | g<<8 | b
}

////////////////////////////////////////////////////////////////////////////////

// ToHsv returns the color converted to HSV with hue in [0, 360]
// and saturation/value/alpha in [0, 1].
func (c Color) ToHsv() (h, s, v, a float64) {

	mx := sysMath.Max(c.R, sysMath.Max(c.G, c.B))
	mn := sysMath.Min(c.R, sysMath.Min(c.G, c.B))
	delta := mx - mn

	v = mx
	a = c.A

	if mx == 0 {
		return 0, 0, v, a
	}

	s = delta / mx

	if delta == 0 {
		return 0, s, v, a
	}

	switch mx {
	case c.R:
		h = 60 * (c.G - c.B) / delta
	case c.G:
		h = 60*(c.B-c.R)/delta + 120
	case c.B:
		h = 60*(c.R-c.G)/delta + 240
	}

	if h < 0 {
		h += 360
	}

	return h, s, v, a
}

////////////////////////////////////////////////////////////////////////////////

// ToHsl returns the color converted to HSL with hue in [0, 360]
// and saturation/lightness/alpha in [0, 1].
func (c Color) ToHsl() (h, s, l, a float64) {

	mx := sysMath.Max(c.R, sysMath.Max(c.G, c.B))
	mn := sysMath.Min(c.R, sysMath.Min(c.G, c.B))
	delta := mx - mn

	l = (mx + mn) / 2
	a = c.A

	if delta == 0 {
		return 0, 0, l, a
	}

	if l < 0.5 {
		s = delta / (mx + mn)
	} else {
		s = delta / (2 - mx - mn)
	}

	switch mx {
	case c.R:
		h = 60 * (c.G - c.B) / delta
	case c.G:
		h = 60*(c.B-c.R)/delta + 120
	case c.B:
		h = 60*(c.R-c.G)/delta + 240
	}

	if h < 0 {
		h += 360
	}

	return h, s, l, a
}

////////////////////////////////////////////////////////////////////////////////

// ToHex returns the color as a hex string. When shorthand is
// true, uses the short format (#RGB or #RGBA) if all channels
// have repeated nibbles. When alpha is true, the alpha channel
// is always included. Otherwise alpha is only included when it
// is less than 255.
func (c Color) ToHex(shorthand bool, alpha bool) string {

	r := uint8(sysMath.Round(Clamp(c.R, 0, 1) * 255))
	g := uint8(sysMath.Round(Clamp(c.G, 0, 1) * 255))
	b := uint8(sysMath.Round(Clamp(c.B, 0, 1) * 255))
	a := uint8(sysMath.Round(Clamp(c.A, 0, 1) * 255))

	showAlpha := alpha || a != 255

	if shorthand &&
		r>>4 == r&0xF &&
		g>>4 == g&0xF &&
		b>>4 == b&0xF &&
		(!showAlpha || a>>4 == a&0xF) {

		if showAlpha {
			return fmt.Sprintf("#%x%x%x%x", r&0xF, g&0xF, b&0xF, a&0xF)
		}

		return fmt.Sprintf("#%x%x%x", r&0xF, g&0xF, b&0xF)
	}

	if showAlpha {
		return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

////////////////////////////////////////////////////////////////////////////////

// ToVector4 returns a `Vector4` with R, G, B, A mapped to X, Y, Z, W.
func (c Color) ToVector4() Vector4 {

	return Vector4{c.R, c.G, c.B, c.A}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the components as a float32 slice.
func (c Color) ToSlice32() []float32 {

	return []float32{
		float32(c.R),
		float32(c.G),
		float32(c.B),
		float32(c.A),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the components as a float64 slice.
func (c Color) ToSlice64() []float64 {

	return []float64{c.R, c.G, c.B, c.A}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// ColorFromRgba creates a color from 8-bit RGBA components in the range [0, 255].
func ColorFromRgba(r, g, b, a uint8) Color {

	return Color{
		float64(r) / 255,
		float64(g) / 255,
		float64(b) / 255,
		float64(a) / 255,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ColorFromPacked creates a color from a packed ARGB uint32.
func ColorFromPacked(argb uint32) Color {

	return ColorFromRgba(
		uint8((argb>>16)&0xFF),
		uint8((argb>>8)&0xFF),
		uint8((argb>>0)&0xFF),
		uint8((argb>>24)&0xFF),
	)
}

////////////////////////////////////////////////////////////////////////////////

// ColorFromHsv creates a color from HSV with hue in [0, 360] and
// saturation/value/alpha in [0, 1].
func ColorFromHsv(h, s, v, a float64) Color {

	h = sysMath.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	s = Clamp(s, 0, 1)
	v = Clamp(v, 0, 1)

	c := v * s
	x := c * (1 - sysMath.Abs(sysMath.Mod(h/60, 2)-1))
	m := v - c

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return Color{r + m, g + m, b + m, a}
}

////////////////////////////////////////////////////////////////////////////////

// ColorFromHsl creates a color from HSL with hue in [0, 360] and
// saturation/lightness/alpha in [0, 1].
func ColorFromHsl(h, s, l, a float64) Color {

	h = sysMath.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	s = Clamp(s, 0, 1)
	l = Clamp(l, 0, 1)

	c := (1 - sysMath.Abs(2*l-1)) * s
	x := c * (1 - sysMath.Abs(sysMath.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return Color{r + m, g + m, b + m, a}
}

////////////////////////////////////////////////////////////////////////////////

// ColorFromHex creates a color from a hex string. Supports
// "#RGB", "#RGBA", "#RRGGBB", and "#RRGGBBAA" formats. The
// "#" prefix is optional.
func ColorFromHex(hex string) (Color, error) {

	hex = strings.TrimPrefix(hex, "#")

	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return ColorZero, ErrInvalidHex
	}

	switch len(hex) {
	case 3:
		r := uint8((v>>8)&0xF) * 17
		g := uint8((v>>4)&0xF) * 17
		b := uint8((v>>0)&0xF) * 17
		return ColorFromRgba(r, g, b, 255), nil

	case 4:
		r := uint8((v>>12)&0xF) * 17
		g := uint8((v>>8)&0xF) * 17
		b := uint8((v>>4)&0xF) * 17
		a := uint8((v>>0)&0xF) * 17
		return ColorFromRgba(r, g, b, a), nil

	case 6:
		r := uint8((v >> 16) & 0xFF)
		g := uint8((v >> 8) & 0xFF)
		b := uint8((v >> 0) & 0xFF)
		return ColorFromRgba(r, g, b, 255), nil

	case 8:
		r := uint8((v >> 24) & 0xFF)
		g := uint8((v >> 16) & 0xFF)
		b := uint8((v >> 8) & 0xFF)
		a := uint8((v >> 0) & 0xFF)
		return ColorFromRgba(r, g, b, a), nil

	default:
		return ColorZero, ErrInvalidHex
	}
}

////////////////////////////////////////////////////////////////////////////////

// ColorFromSlice32 creates a Color from a float32 slice.
func ColorFromSlice32(values []float32) (Color, error) {

	if len(values) != 4 {
		return ColorZero, ErrInvalidLength
	}

	return Color{
		float64(values[0]),
		float64(values[1]),
		float64(values[2]),
		float64(values[3]),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// ColorFromSlice64 creates a Color from a float64 slice.
func ColorFromSlice64(values []float64) (Color, error) {

	if len(values) != 4 {
		return ColorZero, ErrInvalidLength
	}

	return Color{
		values[0],
		values[1],
		values[2],
		values[3],
	}, nil
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two colors.
func (c Color) Add(value Color) Color {

	return Color{c.R + value.R, c.G + value.G, c.B + value.B, c.A + value.A}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two colors.
func (c Color) Sub(value Color) Color {

	return Color{c.R - value.R, c.G - value.G, c.B - value.B, c.A - value.A}
}

////////////////////////////////////////////////////////////////////////////////

// Mul returns the component-wise product of two colors.
func (c Color) Mul(value Color) Color {

	return Color{c.R * value.R, c.G * value.G, c.B * value.B, c.A * value.A}
}

////////////////////////////////////////////////////////////////////////////////

// Div returns the component-wise quotient of two colors.
func (c Color) Div(value Color) Color {

	return Color{c.R / value.R, c.G / value.G, c.B / value.B, c.A / value.A}
}

////////////////////////////////////////////////////////////////////////////////

// AddScalar adds a scalar to each component.
func (c Color) AddScalar(scalar float64) Color {

	return Color{c.R + scalar, c.G + scalar, c.B + scalar, c.A + scalar}
}

////////////////////////////////////////////////////////////////////////////////

// SubScalar subtracts a scalar from each component.
func (c Color) SubScalar(scalar float64) Color {

	return Color{c.R - scalar, c.G - scalar, c.B - scalar, c.A - scalar}
}

////////////////////////////////////////////////////////////////////////////////

// MulScalar multiplies each component by a scalar.
func (c Color) MulScalar(scalar float64) Color {

	return Color{c.R * scalar, c.G * scalar, c.B * scalar, c.A * scalar}
}

////////////////////////////////////////////////////////////////////////////////

// DivScalar divides each component by a scalar.
func (c Color) DivScalar(scalar float64) Color {

	return Color{c.R / scalar, c.G / scalar, c.B / scalar, c.A / scalar}
}

