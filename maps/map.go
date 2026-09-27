package maps

import (
	"encoding/binary"
	"io/fs"
	sysMath "math"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"

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
// with no header (36 bytes per triangle). A zstd
// compressed .tri.zst file is used when one exists.
func Load(dir, name string) (*Map, error) {

	//----------------------------------------------------------------------------//

	start := time.Now()

	data, path, err := readTriFile(dir, name)
	if err != nil {
		return nil, err
	}

	if len(data) < triSize {
		return nil, errors.New(
			"tri file too small",
			errors.String("path", path),
			errors.Int("size", len(data)),
		)
	}

	// The format has no header so the size is the only
	// check that the file actually holds packed triangles
	if len(data)%triSize != 0 {
		return nil, errors.New(
			"tri file size is not a multiple of the triangle size",
			errors.String("path", path),
			errors.Int("size", len(data)),
			errors.Int("triangle_size", triSize),
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
	root := buildBVH(triangles)

	logger.Info("loaded map collision data",
		logger.String("name", name),
		logger.String("path", path),
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

// readTriFile returns the contents of a map's .tri file and
// its path. The compressed .tri.zst file that `extract.sh`
// writes is preferred over an uncompressed .tri file.
func readTriFile(dir, name string) ([]byte, string, error) {

	//----------------------------------------------------------------------------//

	path := filepath.Join(dir, name+".tri.zst")

	compressed, err := os.ReadFile(path)
	if err == nil {
		data, err := decompressTriFile(compressed)
		if err != nil {
			return nil, path, errors.New(
				"failed to decompress tri file",
				errors.String("path", path),
				errors.Error("error", err),
			)
		}

		return data, path, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return nil, path, errors.New(
			"failed to read tri file",
			errors.String("path", path),
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	path = filepath.Join(dir, name+".tri")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, errors.New(
			"failed to read tri file",
			errors.String("path", path),
			errors.Error("error", err),
		)
	}

	return data, path, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func decompressTriFile(compressed []byte) ([]byte, error) {

	decoder, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	return decoder.DecodeAll(compressed, nil)
}

////////////////////////////////////////////////////////////////////////////////

// GetTriangles returns a flat slice of all triangles in
// the map by walking the BVH leaf nodes.
func (m *Map) GetTriangles() []Triangle {

	if m == nil || m.Root == nil {
		return nil
	}

	result := make([]Triangle, 0, m.Triangles)
	collectTriangles(m.Root, &result)
	return result
}

////////////////////////////////////////////////////////////////////////////////

// Trace returns the distance along the ray to the nearest
// triangle intersection within maxDist. Returns (0, false)
// when there is no hit.
func (m *Map) Trace(ray geometry.Ray, maxDist float64) (float64, bool) {

	if m == nil || m.Root == nil {
		return 0, false
	}

	dist := traceNearest(m.Root, ray, maxDist)
	if dist >= maxDist {
		return 0, false
	}
	return dist, true
}

////////////////////////////////////////////////////////////////////////////////

// GetBounds returns the map's axis-aligned bounding box.
// Returns a zero box when the map is nil or has no geometry.
func (m *Map) GetBounds() geometry.Box {

	if m == nil || m.Root == nil {
		return geometry.BoxZero
	}
	return m.Root.Box
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

func traceNearest(node *BVHNode, ray geometry.Ray, best float64) float64 {

	if node == nil {
		return best
	}

	// Prune subtree if the ray misses the bounding box or
	// only enters it beyond the nearest hit found so far
	entry, hit := boxEntry(ray, node.Box)
	if !hit || entry >= best {
		return best
	}

	if node.Triangles != nil {
		for i := range node.Triangles {
			tri := &node.Triangles[i]
			d, ok := ray.IntersectTriangle(tri.V0, tri.V1, tri.V2)
			if ok && d > 0 && d < best {
				best = d
			}
		}
		return best
	}

	// Visit the child whose center is nearer along the ray
	// first so its hits can prune the farther child
	near, far := node.Left, node.Right
	if near != nil && far != nil {
		nearDist := near.Box.Center.Sub(ray.Origin).Dot(ray.Direction)
		farDist := far.Box.Center.Sub(ray.Origin).Dot(ray.Direction)
		if farDist < nearDist {
			near, far = far, near
		}
	}

	best = traceNearest(near, ray, best)
	best = traceNearest(far, ray, best)
	return best
}

////////////////////////////////////////////////////////////////////////////////

func intersects(node *BVHNode, ray geometry.Ray, maxDist float64) bool {

	if node == nil {
		return false
	}

	// Prune subtree if the ray misses the bounding box or
	// only enters it beyond the end of the segment
	entry, hit := boxEntry(ray, node.Box)
	if !hit || entry >= maxDist {
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

// boxEntry returns the distance at which the ray enters the
// box, or zero when the ray starts inside it. In that case
// `IntersectBox` returns the exit distance instead, which
// cannot be used for pruning.
func boxEntry(ray geometry.Ray, box geometry.Box) (float64, bool) {

	dist, hit := ray.IntersectBox(box)
	if !hit {
		return 0, false
	}

	if box.Contains(ray.Origin) {
		return 0, true
	}
	return dist, true
}

////////////////////////////////////////////////////////////////////////////////

func collectTriangles(node *BVHNode, result *[]Triangle) {

	if node == nil {
		return
	}

	if node.Triangles != nil {
		*result = append(*result, node.Triangles...)
		return
	}

	collectTriangles(node.Left, result)
	collectTriangles(node.Right, result)
}

////////////////////////////////////////////////////////////////////////////////

func decodeVec3(data []byte) math.Vector3 {

	return math.Vector3{
		X: float64(sysMath.Float32frombits(binary.LittleEndian.Uint32(data[0:4]))),
		Y: float64(sysMath.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))),
		Z: float64(sysMath.Float32frombits(binary.LittleEndian.Uint32(data[8:12]))),
	}
}
