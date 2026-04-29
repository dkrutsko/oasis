package maps

import (
	"encoding/binary"
	sysMath "math"
	"os"
	"path/filepath"
	"time"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Size of a single triangle in the .tri file format.
// Each triangle is 9 consecutive float32 values (3
// vertices x 3 components) stored in little-endian.
const triSize = 36

////////////////////////////////////////////////////////////////////////////////

// Map holds the parsed collision geometry and spatial
// acceleration structure for a single game map. Use
// Load to create a Map from a .tri file on disk.
type Map struct {
	Name      string
	Root      *BVHNode
	Triangles int
}

////////////////////////////////////////////////////////////////////////////////

// Load reads a .tri file from the given directory and
// builds a BVH for fast ray intersection queries. The
// .tri format is a flat array of packed float32 triplets
// with no header (36 bytes per triangle).
func Load(dir, name string) (*Map, error) {

	//----------------------------------------------------------------------------//

	path := filepath.Join(dir, name+".tri")
	start := time.Now()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New(
			"failed to read tri file",
			errors.String("path", path),
			errors.Error("error", err),
		)
	}

	if len(data) < triSize {
		return nil, errors.New(
			"tri file too small",
			errors.String("path", path),
			errors.Int("size", len(data)),
		)
	}

	//----------------------------------------------------------------------------//

	// Decode packed float32 triangles
	count := len(data) / triSize
	triangles := make([]Triangle, count)

	for i := 0; i < count; i++ {
		off := i * triSize
		triangles[i] = Triangle{
			V0: decodeVec3(data[off:]),
			V1: decodeVec3(data[off+12:]),
			V2: decodeVec3(data[off+24:]),
		}
	}

	//----------------------------------------------------------------------------//

	// Build spatial acceleration structure
	root := buildBVH(triangles, 0)

	logger.Info("loaded map collision data",
		logger.String("name", name),
		logger.Int("triangles", count),
		logger.Duration("elapsed", time.Since(start)),
	)

	return &Map{
		Name:      name,
		Root:      root,
		Triangles: count,
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// IsVisible returns true if the line segment from `from`
// to `to` does not intersect any triangle in the map
// collision geometry. Returns true when the map is nil
// or has no geometry loaded.
func (m *Map) IsVisible(from, to math.Vector3) bool {

	if m == nil || m.Root == nil {
		return true
	}

	ray := geometry.RayFromPoints(from, to)
	maxDist := from.Distance(to)

	return !intersects(m.Root, ray, maxDist)
}

////////////////////////////////////////////////////////////////////////////////

func intersects(node *BVHNode, ray geometry.Ray, maxDist float64) bool {

	if node == nil {
		return false
	}

	// Prune subtree if the ray misses the bounding box
	_, hit := ray.IntersectBox(node.Box)
	if !hit {
		return false
	}

	// Leaf node - test each triangle
	if node.Triangles != nil {
		for i := range node.Triangles {
			tri := &node.Triangles[i]
			d, ok := ray.IntersectTriangle(tri.V0, tri.V1, tri.V2)
			if ok && d > 0 && d < maxDist {
				return true
			}
		}
		return false
	}

	// Internal node - early exit on first hit
	if intersects(node.Left, ray, maxDist) {
		return true
	}
	return intersects(node.Right, ray, maxDist)
}

////////////////////////////////////////////////////////////////////////////////

func decodeVec3(data []byte) math.Vector3 {

	return math.Vector3{
		X: float64(sysMath.Float32frombits(binary.LittleEndian.Uint32(data[0:4]))),
		Y: float64(sysMath.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))),
		Z: float64(sysMath.Float32frombits(binary.LittleEndian.Uint32(data[8:12]))),
	}
}
