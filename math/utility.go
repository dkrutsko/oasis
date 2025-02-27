package math

import (
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

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

	return sysMath.MaxFloat32
}
