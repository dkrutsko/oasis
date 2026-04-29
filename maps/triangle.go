package maps

import "github.com/dkrutsko/oasis/math"

////////////////////////////////////////////////////////////////////////////////

// Triangle represents a single triangle in world space
// defined by three vertex positions.
type Triangle struct {
	V0 math.Vector3
	V1 math.Vector3
	V2 math.Vector3
}

////////////////////////////////////////////////////////////////////////////////

// GetCentroid returns the center point of the triangle.
func (t Triangle) GetCentroid() math.Vector3 {

	return math.Vector3{
		X: (t.V0.X + t.V1.X + t.V2.X) / 3,
		Y: (t.V0.Y + t.V1.Y + t.V2.Y) / 3,
		Z: (t.V0.Z + t.V1.Z + t.V2.Z) / 3,
	}
}
