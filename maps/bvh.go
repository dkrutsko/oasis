package maps

import (
	sysMath "math"
	"sync"

	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Maximum number of triangles in a BVH leaf node.
// Nodes with this many or fewer triangles are not
// subdivided further.
const bvhLeafThreshold = 4

// Minimum number of triangles in a BVH node for its
// subtrees to be built in parallel. Smaller subtrees
// are built faster than a goroutine is worth.
const bvhParallelThreshold = 32768

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

func buildBVH(triangles []Triangle) *BVHNode {

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

	// Split at the median centroid along the split axis.
	// Only the median has to be in place, so the triangles
	// are partitioned around it instead of sorted.
	mid := len(triangles) / 2
	selectTriangle(triangles, mid, axis)

	if len(triangles) < bvhParallelThreshold {
		node.Left = buildBVH(triangles[:mid])
		node.Right = buildBVH(triangles[mid:])
		return node
	}

	// The halves share no triangles, so large subtrees are
	// built in parallel
	var wg sync.WaitGroup
	wg.Go(func() {
		node.Left = buildBVH(triangles[:mid])
	})

	node.Right = buildBVH(triangles[mid:])
	wg.Wait()

	return node
}

////////////////////////////////////////////////////////////////////////////////

// selectTriangle reorders the triangles so that the one at
// `k` is the one sorting by centroid along `axis` would put
// there. No triangle before it has a greater centroid and
// none after it has a smaller one.
func selectTriangle(triangles []Triangle, k, axis int) {

	lo, hi := 0, len(triangles)-1

	for lo < hi {
		pivot := medianOfThree(
			centroidKey(&triangles[lo], axis),
			centroidKey(&triangles[lo+(hi-lo)/2], axis),
			centroidKey(&triangles[hi], axis),
		)

		// Triangles equal to the pivot stop both scans, so
		// many equal centroids still split evenly
		i, j := lo, hi
		for i <= j {
			for centroidKey(&triangles[i], axis) < pivot {
				i++
			}
			for centroidKey(&triangles[j], axis) > pivot {
				j--
			}
			if i <= j {
				triangles[i], triangles[j] = triangles[j], triangles[i]
				i++
				j--
			}
		}

		// Continue in the side holding `k`. Triangles between
		// the sides are equal to the pivot.
		switch {
		case k <= j:
			hi = j
		case k >= i:
			lo = i
		default:
			return
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// centroidKey returns the triangle centroid along the axis
// scaled by three. Leaving out the division does not change
// the order of the triangles.
func centroidKey(tri *Triangle, axis int) float64 {

	switch axis {
	case 1:
		return tri.V0.Y + tri.V1.Y + tri.V2.Y
	case 2:
		return tri.V0.Z + tri.V1.Z + tri.V2.Z
	default:
		return tri.V0.X + tri.V1.X + tri.V2.X
	}
}

////////////////////////////////////////////////////////////////////////////////

func medianOfThree(a, b, c float64) float64 {

	if a > b {
		a, b = b, a
	}
	if b > c {
		b = c
	}
	if a > b {
		b = a
	}
	return b
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
