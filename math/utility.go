package math

import (
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

// IsNanOrInf returns whether the value is NaN or infinity.
func IsNanOrInf(value float64) bool {

	return sysMath.IsNaN(value) || sysMath.IsInf(value, 0)
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

// WrapPI wraps a radian angle to the range [-pi, pi].
func WrapPI(angle float64) float64 {

	const twoPi = 2 * sysMath.Pi

	angle = sysMath.Mod(angle+sysMath.Pi, twoPi)
	if angle < 0 {
		angle += twoPi
	}

	return angle - sysMath.Pi
}

////////////////////////////////////////////////////////////////////////////////

// WrapTwoPI wraps a radian angle to the range [0, 2*pi].
func WrapTwoPI(angle float64) float64 {

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

// SmoothStep performs Hermite interpolation between two values with
// smoothing at the edges. The amount is clamped to the range [0, 1].
func SmoothStep(value1, value2, amount float64) float64 {

	amount = Clamp(amount, 0, 1)
	amount = amount * amount * (3 - 2*amount)

	return value1 + (value2-value1)*amount
}

////////////////////////////////////////////////////////////////////////////////

// IsAngleEqual returns whether the target angle is within the
// specified tolerance of the source angle in degrees. Handles
// circular wrapping across the 0/360 boundary.
func IsAngleEqual(source, target, tolerance float64) bool {

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

// IsInsideSlice returns whether the point (tx, ty) falls
// within a directional cone originating from (sx, sy) with
// the given direction and angular size in degrees.
func IsInsideSlice(sx, sy, dir, size, tx, ty float64) bool {

	// If full circle
	if size == 360 {
		return true
	}

	const radToDeg = 180.0 / sysMath.Pi

	// Calculate angle between s-t
	angle := sysMath.Atan2(ty-sy, tx-sx) * radToDeg

	// Determine if angle is equal
	return IsAngleEqual(dir, angle, size*0.5)
}

////////////////////////////////////////////////////////////////////////////////

// ComputePitchYaw computes pitch and yaw angles in degrees from a source
// position to a target position.
func ComputePitchYaw(
	sx, sy, sz float64,
	tx, ty, tz float64,
) (rx, ry float64) {

	dx := sx - tx
	dy := sy - ty
	dz := sz - tz
	di := sysMath.Sqrt(dx*dx + dy*dy)

	const radToDeg = 180.0 / sysMath.Pi

	rx = sysMath.Atan2(dz, di) * radToDeg
	ry = sysMath.Atan2(dy, dx)*radToDeg + 180.0

	return rx, ry
}

////////////////////////////////////////////////////////////////////////////////

// ProjectPoint projects a 3D world position onto 2D screen
// coordinates using the given view-projection transform.
// Returns the screen position and whether it is visible.
func ProjectPoint(
	transform Matrix,
	x, y, z float64,
	width, height int,
) (screenX, screenY float64, visible bool) {

	ww := transform.M41 * x
	ww += transform.M42 * y
	ww += transform.M43 * z
	ww += transform.M44 * 1

	// If visible
	if ww < 0.01 {
		return 0, 0, false
	}

	// Project position on two dimensions
	xx := (transform.M11*x + transform.M12*y + transform.M13*z + transform.M14) / ww
	yy := (transform.M21*x + transform.M22*y + transform.M23*z + transform.M24) / ww

	w := float64(width)
	h := float64(height)

	// Convert to screen coordinates
	screenX = (xx + 1) * +0.5 * w
	screenY = (yy - 1) * -0.5 * h

	return screenX, screenY, true
}

////////////////////////////////////////////////////////////////////////////////

// CastRay performs a ray-sphere intersection test and returns the distance to
// the nearest hit point. Returns `math.MaxFloat64` if no intersection occurs.
func CastRay(origin, target, direction Vector3, radius float64) float64 {

	// https://www.ccs.neu.edu/home/fell/CS4300/Lectures/Ray-TracingFormulas.pdf

	a := direction.X * direction.X
	a += direction.Y * direction.Y
	a += direction.Z * direction.Z

	b := 2 * direction.X * (origin.X - target.X)
	b += 2 * direction.Y * (origin.Y - target.Y)
	b += 2 * direction.Z * (origin.Z - target.Z)

	c := origin.X*origin.X + target.X*target.X
	c += origin.Y*origin.Y + target.Y*target.Y
	c += origin.Z*origin.Z + target.Z*target.Z
	c -= 2 * (origin.X*target.X + origin.Y*target.Y + origin.Z*target.Z)
	c -= radius * radius

	discriminant := b*b - 4*a*c

	// Whether intersected
	if discriminant >= 0 {

		// Find a location where sphere intersection was made
		result := (-b - sysMath.Sqrt(discriminant)) / (2 * a)

		// Ensure valid
		if result >= 0 {
			return result
		}
	}

	return sysMath.MaxFloat64
}
