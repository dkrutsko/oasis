package math

import (
	"errors"
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

var (
	// ErrInvalidLength is returned by FromSlice constructors when
	// the input slice does not have the required number of elements.
	ErrInvalidLength = errors.New("invalid slice length")

	// ErrInvalidHex is returned by `ColorFromHex` when the input
	// string contains invalid hex characters or has an unsupported
	// length.
	ErrInvalidHex = errors.New("invalid hex color")
)

////////////////////////////////////////////////////////////////////////////////

// IsNanOrInf returns whether the value is NaN or infinity.
func IsNanOrInf(value float64) bool {

	return sysMath.IsNaN(value) || sysMath.IsInf(value, 0)
}

////////////////////////////////////////////////////////////////////////////////

// RoundN rounds a value to the specified number of decimal places.
func RoundN(value float64, precision int) float64 {

	p := sysMath.Pow(10, float64(precision))
	return sysMath.Round(value*p) / p
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the absolute distance between two values.
func Distance(value1, value2 float64) float64 {

	return sysMath.Abs(value1 - value2)
}

////////////////////////////////////////////////////////////////////////////////

// Clamp restricts a value to the specified range.
func Clamp(value, min, max float64) float64 {

	if value > max {
		return max
	}

	if value < min {
		return min
	}

	return value
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs linear interpolation between two values.
func Lerp(value1, value2, amount float64) float64 {

	return value1 + (value2-value1)*amount
}

////////////////////////////////////////////////////////////////////////////////

// InverseLerp returns the normalized position of a value within
// the range [value1, value2]. Returns 0 when value equals value1
// and 1 when value equals value2. This is the inverse of `Lerp`.
func InverseLerp(value1, value2, value float64) float64 {

	if value1 == value2 {
		return 0
	}

	return (value - value1) / (value2 - value1)
}

////////////////////////////////////////////////////////////////////////////////

// SmoothStep performs Hermite interpolation between two values with
// smoothing at the edges. The amount is clamped to the range [0, 1].
func SmoothStep(value1, value2, amount float64) float64 {

	amount = Clamp(amount, 0, 1)
	amount = amount * amount * (3 - 2*amount)

	return value1 + (value2-value1)*amount
}

////////////////////////////////////////////////////////////////////////////////

// ToDegrees converts radians to degrees.
func ToDegrees(radians float64) float64 {

	return radians * (180.0 / sysMath.Pi)
}

////////////////////////////////////////////////////////////////////////////////

// ToRadians converts degrees to radians.
func ToRadians(degrees float64) float64 {

	return degrees * (sysMath.Pi / 180.0)
}

////////////////////////////////////////////////////////////////////////////////

// WrapPi wraps a radian angle to the range [-pi, pi].
func WrapPi(angle float64) float64 {

	const twoPi = 2 * sysMath.Pi

	angle = sysMath.Mod(angle+sysMath.Pi, twoPi)
	if angle < 0 {
		angle += twoPi
	}

	return angle - sysMath.Pi
}

////////////////////////////////////////////////////////////////////////////////

// WrapTwoPi wraps a radian angle to the range [0, 2*pi].
func WrapTwoPi(angle float64) float64 {

	const twoPi = 2 * sysMath.Pi

	angle = sysMath.Mod(angle, twoPi)
	if angle < 0 {
		angle += twoPi
	}

	return angle
}

////////////////////////////////////////////////////////////////////////////////

// Wrap180 wraps a degree angle to the range [-180, 180].
func Wrap180(angle float64) float64 {

	angle = sysMath.Mod(angle+180, 360)
	if angle < 0 {
		angle += 360
	}

	return angle - 180
}

////////////////////////////////////////////////////////////////////////////////

// Wrap360 wraps a degree angle to the range [0, 360].
func Wrap360(angle float64) float64 {

	angle = sysMath.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}

	return angle
}

////////////////////////////////////////////////////////////////////////////////

// IsRadEqual returns whether the target angle is within the
// specified tolerance of the source angle in radians. Handles
// circular wrapping across the 0/2*Pi boundary.
func IsRadEqual(source, target, tolerance float64) bool {

	twoPi := 2 * sysMath.Pi

	angle := sysMath.Mod(twoPi+sysMath.Mod(target, twoPi), twoPi)
	angleMin := sysMath.Mod(twoPi*1000000+source-tolerance, twoPi)
	angleMax := sysMath.Mod(twoPi*1000000+source+tolerance, twoPi)

	if angleMin < angleMax {
		return angleMin <= angle && angle <= angleMax
	} else {
		return angleMin <= angle || angle <= angleMax
	}
}

////////////////////////////////////////////////////////////////////////////////

// IsDegEqual returns whether the target angle is within the
// specified tolerance of the source angle in degrees. Handles
// circular wrapping across the 0/360 boundary.
func IsDegEqual(source, target, tolerance float64) bool {

	angle := sysMath.Mod(360+sysMath.Mod(target, 360), 360)
	angleMin := sysMath.Mod(3600000+source-tolerance, 360)
	angleMax := sysMath.Mod(3600000+source+tolerance, 360)

	if angleMin < angleMax {
		return angleMin <= angle && angle <= angleMax
	} else {
		return angleMin <= angle || angle <= angleMax
	}
}

////////////////////////////////////////////////////////////////////////////////

// IsRadInsideSlice returns whether the target point falls within a
// directional cone originating from the source point with the given
// direction and angular size in radians.
func IsRadInsideSlice(source Vector2, direction, size float64, target Vector2) bool {

	twoPi := 2 * sysMath.Pi

	// If full circle
	if size == twoPi {
		return true
	}

	angle := sysMath.Atan2(
		target.Y-source.Y,
		target.X-source.X,
	)

	return IsRadEqual(direction, angle, size*0.5)
}

////////////////////////////////////////////////////////////////////////////////

// IsDegInsideSlice returns whether the target point falls within a
// directional cone originating from the source point with the given
// direction and angular size in degrees.
func IsDegInsideSlice(source Vector2, direction, size float64, target Vector2) bool {

	// If full circle
	if size == 360 {
		return true
	}

	const radToDeg = 180.0 / sysMath.Pi

	angle := sysMath.Atan2(
		target.Y-source.Y,
		target.X-source.X,
	) * radToDeg

	return IsDegEqual(direction, angle, size*0.5)
}

////////////////////////////////////////////////////////////////////////////////

// ComputeRadPitchYaw computes pitch and yaw angles in radians
// from a source position to a target position. Returns the
// result as a `Vector2` where X is pitch and Y is yaw.
func ComputeRadPitchYaw(source, target Vector3) Vector2 {

	dx := source.X - target.X
	dy := source.Y - target.Y
	dz := source.Z - target.Z
	di := sysMath.Sqrt(dx*dx + dy*dy)

	return Vector2{
		sysMath.Atan2(dz, di),
		sysMath.Atan2(dy, dx) + sysMath.Pi,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ComputeDegPitchYaw computes pitch and yaw angles in degrees
// from a source position to a target position. Returns the
// result as a `Vector2` where X is pitch and Y is yaw.
func ComputeDegPitchYaw(source, target Vector3) Vector2 {

	dx := source.X - target.X
	dy := source.Y - target.Y
	dz := source.Z - target.Z
	di := sysMath.Sqrt(dx*dx + dy*dy)

	const radToDeg = 180.0 / sysMath.Pi

	return Vector2{
		sysMath.Atan2(dz, di) * radToDeg,
		sysMath.Atan2(dy, dx)*radToDeg + 180.0,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ProjectToScreen projects a 3D position to screen coordinates using
// separate model, view, and projection matrices. Returns the screen
// position with depth in Z and whether the point is visible.
func ProjectToScreen(
	position Vector3, viewport Size,
	model, view, projection Matrix4,
) (result Vector3, visible bool) {

	return ProjectToScreenMvp(position, viewport, model.Mul(view).Mul(projection))
}

////////////////////////////////////////////////////////////////////////////////

// ProjectToScreenMvp projects a 3D position to screen coordinates using
// a precomputed model-view-projection matrix. Returns the screen position
// with depth in Z and whether the point is visible.
func ProjectToScreenMvp(
	position Vector3, viewport Size, mvp Matrix4,
) (result Vector3, visible bool) {

	const minZ = 0.0
	const maxZ = 1.0

	transform := Vector4TransformVector3(mvp, position)

	if transform.W < 0.01 {
		return Vector3Zero, false
	}

	w := float64(viewport.W)
	h := float64(viewport.H)

	return Vector3{
		(1 + transform.X/transform.W) * w * 0.5,
		(1 - transform.Y/transform.W) * h * 0.5,
		minZ + (transform.Z/transform.W)*(maxZ-minZ),
	}, true
}
