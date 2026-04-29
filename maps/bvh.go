package maps

import (
	sysMath "math"
	"sort"

	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Maximum number of triangles in a BVH leaf node.
// Nodes with this many or fewer triangles are not
// subdivided further.
const bvhLeafThreshold = 4

////////////////////////////////////////////////////////////////////////////////

// BVHNode is a node in a bounding volume hierarchy used
// for accelerating ray-triangle intersection queries.
type BVHNode struct {
	Box       geometry.Box
	Left      *BVHNode
	Right     *BVHNode
	Triangles []Triangle
}

////////////////////////////////////////////////////////////////////////////////

func buildBVH(triangles []Triangle, depth int) *BVHNode {

	if len(triangles) == 0 {
		return nil
	}

	node := &BVHNode{
		Box: computeBounds(triangles),
	}

	// Store directly in leaf when below threshold
	if len(triangles) <= bvhLeafThreshold {
		node.Triangles = triangles
		return node
	}

	// Split along the longest axis of the bounding box
	size := node.Box.GetSize()
	axis := longestAxis(size)

	// Sort triangles by centroid along the split axis
	sort.Slice(triangles, func(i, j int) bool {
		ci := triangles[i].GetCentroid()
		cj := triangles[j].GetCentroid()
		return axisComponent(ci, axis) < axisComponent(cj, axis)
	})

	// Split at median
	mid := len(triangles) / 2
	node.Left = buildBVH(triangles[:mid], depth+1)
	node.Right = buildBVH(triangles[mid:], depth+1)

	return node
}

////////////////////////////////////////////////////////////////////////////////

func computeBounds(triangles []Triangle) geometry.Box {

	min := math.Vector3{
		X: sysMath.MaxFloat64,
		Y: sysMath.MaxFloat64,
		Z: sysMath.MaxFloat64,
	}

	max := math.Vector3{
		X: -sysMath.MaxFloat64,
		Y: -sysMath.MaxFloat64,
		Z: -sysMath.MaxFloat64,
	}

	for i := range triangles {
		tri := &triangles[i]
		min = min.Min(tri.V0).Min(tri.V1).Min(tri.V2)
		max = max.Max(tri.V0).Max(tri.V1).Max(tri.V2)
	}

	return geometry.BoxFromMinMax(min, max)
}

////////////////////////////////////////////////////////////////////////////////

func longestAxis(size math.Vector3) int {

	axis := 0
	if size.Y > size.X {
		axis = 1
	}
	if size.Z > axisComponent(size, axis) {
		axis = 2
	}
	return axis
}

////////////////////////////////////////////////////////////////////////////////

func axisComponent(v math.Vector3, axis int) float64 {

	switch axis {
	case 1:
		return v.Y
	case 2:
		return v.Z
	default:
		return v.X
	}
}
