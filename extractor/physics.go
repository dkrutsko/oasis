package main

// Conversion of decoded world physics into collision triangles. The hull
// and mesh conversion follows CS2-Phys-Extractor (MIT License, Copyright
// (c) 2025 LAITHCOOL). See NOTICE for the license text.

import (
	"encoding/binary"
	sysMath "math"
	"slices"
	"strings"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Size of a vertex in hull and mesh data, which is three
	// float32 values
	vertexSize = 12

	// Size of a hull half-edge, which is four byte indices: the
	// next half-edge, its twin, its origin vertex and its face
	halfEdgeSize   = 4
	halfEdgeNext   = 0
	halfEdgeOrigin = 2

	// Size of a mesh triangle, which is three int32 vertex indices
	meshTriangleSize = 12

	// Limit on the half-edges walked around a single hull face,
	// which stops a corrupt loop from running forever
	maxFaceEdges = 1000

	// Resolution of generated capsules and spheres. Segments go
	// around the axis and each end has this many rings between its
	// pole and its equator. The generated surface lies inside the
	// true one, at most 14% of the radius below it.
	capsuleSegments = 8
	capsuleRings    = 2
)

////////////////////////////////////////////////////////////////////////////////

// PhysicsKeys are the object fields `ConvertPhysics` reads.
// Passing them to `DecodeKv3` avoids storing the rest of the
// physics data.
var PhysicsKeys = map[string]bool{
	"m_bindPose":                 true,
	"m_collisionAttributes":      true,
	"m_CollisionGroupString":     true,
	"m_parts":                    true,
	"m_rnShape":                  true,
	"m_hulls":                    true,
	"m_meshes":                   true,
	"m_spheres":                  true,
	"m_capsules":                 true,
	"m_nCollisionAttributeIndex": true,
	"m_Hull":                     true,
	"m_VertexPositions":          true,
	"m_Vertices":                 true,
	"m_Faces":                    true,
	"m_Edges":                    true,
	"m_Mesh":                     true,
	"m_Triangles":                true,
	"m_Sphere":                   true,
	"m_Capsule":                  true,
	"m_vCenter":                  true,
	"m_flRadius":                 true,
}

////////////////////////////////////////////////////////////////////////////////

// TriangleFunc receives each triangle of the collision geometry.
type TriangleFunc func(v0, v1, v2 math.Vector3)

////////////////////////////////////////////////////////////////////////////////

// PhysicsResult holds the number of shapes of each kind that were
// converted.
type PhysicsResult struct {
	Hulls    int
	Meshes   int
	Spheres  int
	Capsules int
}

////////////////////////////////////////////////////////////////////////////////

// ConvertPhysics turns the shapes of the default collision group
// into triangles. Hull triangles come first, then those of the
// meshes, spheres and capsules.
func ConvertPhysics(phys *Kv3Value, emit TriangleFunc) PhysicsResult {

	//----------------------------------------------------------------------------//

	var result PhysicsResult
	indices := getDefaultCollisionIndices(phys)

	// World physics has no bind pose in any current map, so its
	// shapes are already in world space
	bindPose := phys.GetField("m_bindPose").GetElements()
	if len(bindPose) > 0 {
		logger.Warn(
			"world physics has a bind pose, which is ignored",
			logger.Int("bones", len(bindPose)),
		)
	}

	//----------------------------------------------------------------------------//

	for _, part := range phys.GetField("m_parts").GetElements() {
		shapes := part.GetField("m_rnShape")

		hulls := shapes.GetField("m_hulls")
		meshes := shapes.GetField("m_meshes")
		spheres := shapes.GetField("m_spheres")
		capsules := shapes.GetField("m_capsules")

		result.Hulls += convertShapes(hulls, "m_Hull", indices, convertHull, emit)
		result.Meshes += convertShapes(meshes, "m_Mesh", indices, convertMesh, emit)
		result.Spheres += convertShapes(spheres, "m_Sphere", indices, convertSphere, emit)
		result.Capsules += convertShapes(capsules, "m_Capsule", indices, convertCapsule, emit)
	}

	return result

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// getDefaultCollisionIndices returns the indices of the collision
// attributes in the "default" group, which shapes refer to by
// index. Like CS2-Phys-Extractor it falls back to index 0 when no
// group is named default.
func getDefaultCollisionIndices(phys *Kv3Value) []int64 {

	var result []int64
	attributes := phys.GetField("m_collisionAttributes").GetElements()

	for index, attribute := range attributes {
		group := attribute.GetField("m_CollisionGroupString").GetString()

		if strings.EqualFold(group, "default") {
			result = append(result, int64(index))
		}
	}

	if len(result) == 0 {
		result = append(result, 0)
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

// convertShapes converts the shapes of a list that are in the
// default collision group and returns how many were converted.
// Each shape keeps its geometry in the named field, which
// `convert` turns into triangles.
func convertShapes(
	shapes *Kv3Value,
	geometryField string,
	indices []int64,
	convert func(geometry *Kv3Value, emit TriangleFunc) error,
	emit TriangleFunc,
) int {

	converted := 0

	for index, shape := range shapes.GetElements() {
		collision, ok := shape.GetField("m_nCollisionAttributeIndex").GetInt()
		if !ok || !slices.Contains(indices, collision) {
			continue
		}

		err := convert(shape.GetField(geometryField), emit)
		if err != nil {
			logger.Warn(
				"failed to convert shape",
				logger.String("type", geometryField),
				logger.Int("index", index),
				logger.Error("error", err),
			)
			continue
		}

		converted++
	}

	return converted
}

////////////////////////////////////////////////////////////////////////////////

// convertHull splits each face of a convex hull into a fan of
// triangles. Faces are loops of half-edges, and each face holds
// the index of one of its half-edges.
func convertHull(hull *Kv3Value, emit TriangleFunc) error {

	//----------------------------------------------------------------------------//

	// Newer hulls name their vertices `m_VertexPositions`, older
	// ones `m_Vertices`
	vertexData := hull.GetField("m_VertexPositions").GetBlob()
	if len(vertexData) == 0 {
		vertexData = hull.GetField("m_Vertices").GetBlob()
	}

	vertices, err := decodeVertices(vertexData)
	if err != nil {
		return err
	}

	faces := hull.GetField("m_Faces").GetBlob()
	edges := hull.GetField("m_Edges").GetBlob()

	if len(edges)%halfEdgeSize != 0 {
		return errors.New(
			"hull edge data is not a multiple of the half-edge size",
			errors.Int("size", len(edges)),
		)
	}

	edgeCount := len(edges) / halfEdgeSize

	//----------------------------------------------------------------------------//

	for _, face := range faces {
		start := int(face)
		if start >= edgeCount {
			continue
		}

		// Every triangle of the fan starts at the origin of the
		// face's own half-edge
		origin := int(edges[start*halfEdgeSize+halfEdgeOrigin])
		edge := int(edges[start*halfEdgeSize+halfEdgeNext])

		for i := 0; edge != start && i < maxFaceEdges; i++ {
			if edge >= edgeCount {
				break
			}

			// The half-edge that leads back to the start would
			// give a triangle with zero area, so the fan ends
			next := int(edges[edge*halfEdgeSize+halfEdgeNext])
			if next >= edgeCount || next == start {
				break
			}

			v1 := int(edges[edge*halfEdgeSize+halfEdgeOrigin])
			v2 := int(edges[next*halfEdgeSize+halfEdgeOrigin])

			if origin < len(vertices) && v1 < len(vertices) && v2 < len(vertices) {
				emit(vertices[origin], vertices[v1], vertices[v2])
			}

			edge = next
		}
	}

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// convertMesh emits the triangles of an indexed triangle mesh.
// Every index is checked first, so a bad mesh adds no triangles.
func convertMesh(mesh *Kv3Value, emit TriangleFunc) error {

	//----------------------------------------------------------------------------//

	vertices, err := decodeVertices(mesh.GetField("m_Vertices").GetBlob())
	if err != nil {
		return err
	}

	triangles := mesh.GetField("m_Triangles").GetBlob()

	if len(triangles)%meshTriangleSize != 0 {
		return errors.New(
			"mesh triangle data is not a multiple of the triangle size",
			errors.Int("size", len(triangles)),
		)
	}

	for i := 0; i < len(triangles); i += 4 {
		index := int32(binary.LittleEndian.Uint32(triangles[i : i+4]))

		if index < 0 || int(index) >= len(vertices) {
			return errors.New(
				"mesh vertex index is out of range",
				errors.Int32("index", index),
				errors.Int("vertices", len(vertices)),
			)
		}
	}

	//----------------------------------------------------------------------------//

	for i := 0; i < len(triangles); i += meshTriangleSize {
		a := binary.LittleEndian.Uint32(triangles[i : i+4])
		b := binary.LittleEndian.Uint32(triangles[i+4 : i+8])
		c := binary.LittleEndian.Uint32(triangles[i+8 : i+12])

		emit(vertices[a], vertices[b], vertices[c])
	}

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// convertSphere emits a sphere as a capsule whose ends are both
// at its center.
func convertSphere(sphere *Kv3Value, emit TriangleFunc) error {

	center, err := getVector3(sphere.GetField("m_vCenter"))
	if err != nil {
		return err
	}

	radius, err := getRadius(sphere.GetField("m_flRadius"))
	if err != nil {
		return err
	}

	emitCapsule(center, center, radius, emit)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// convertCapsule emits a capsule. Its two centers are the ends
// of the line segment the capsule is built around.
func convertCapsule(capsule *Kv3Value, emit TriangleFunc) error {

	centers := capsule.GetField("m_vCenter").GetElements()
	if len(centers) != 2 {
		return errors.New(
			"capsule does not have two centers",
			errors.Int("centers", len(centers)),
		)
	}

	start, err := getVector3(&centers[0])
	if err != nil {
		return err
	}

	end, err := getVector3(&centers[1])
	if err != nil {
		return err
	}

	radius, err := getRadius(capsule.GetField("m_flRadius"))
	if err != nil {
		return err
	}

	emitCapsule(start, end, radius, emit)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// emitCapsule emits a capsule around the line segment from
// `start` to `end` as rings of vertices joined by triangles. The
// rings run from the pole below `start` to the pole above `end`.
// When both ends are the same point the capsule is a sphere.
func emitCapsule(start, end math.Vector3, radius float64, emit TriangleFunc) {

	//----------------------------------------------------------------------------//

	// A sphere has no axis of its own, so any direction works
	isSphere := start == end

	axis := math.Vector3UnitZ
	if !isSphere {
		axis = end.Sub(start).Normalize()
	}

	// Two directions at right angles to the axis span each ring.
	// The helper only needs to not be parallel to the axis.
	helper := math.Vector3UnitX
	if sysMath.Abs(axis.X) > 0.9 {
		helper = math.Vector3UnitY
	}

	tangent := axis.Cross(helper).Normalize()
	bitangent := axis.Cross(tangent)

	//----------------------------------------------------------------------------//

	// Rings around `start` go from its pole to its equator, and
	// rings around `end` from its equator to its pole. A sphere
	// only needs its equator once.
	var rings [][]math.Vector3

	for i := 1; i <= capsuleRings; i++ {
		angle := sysMath.Pi / 2 * float64(i) / capsuleRings
		center := start.Sub(axis.MulScalar(radius * sysMath.Cos(angle)))

		ring := makeRing(center, tangent, bitangent, radius*sysMath.Sin(angle))
		rings = append(rings, ring)
	}

	first := capsuleRings
	if isSphere {
		first = capsuleRings - 1
	}

	for i := first; i >= 1; i-- {
		angle := sysMath.Pi / 2 * float64(i) / capsuleRings
		center := end.Add(axis.MulScalar(radius * sysMath.Cos(angle)))

		ring := makeRing(center, tangent, bitangent, radius*sysMath.Sin(angle))
		rings = append(rings, ring)
	}

	//----------------------------------------------------------------------------//

	// Triangles face outward, fanning from each pole to its ring
	// and joining neighboring rings with two triangles per segment
	bottom := start.Sub(axis.MulScalar(radius))
	top := end.Add(axis.MulScalar(radius))
	last := len(rings) - 1

	for k := 0; k < capsuleSegments; k++ {
		next := (k + 1) % capsuleSegments
		emit(bottom, rings[0][next], rings[0][k])
	}

	for i := 0; i < last; i++ {
		for k := 0; k < capsuleSegments; k++ {
			next := (k + 1) % capsuleSegments

			emit(rings[i][k], rings[i][next], rings[i+1][next])
			emit(rings[i][k], rings[i+1][next], rings[i+1][k])
		}
	}

	for k := 0; k < capsuleSegments; k++ {
		next := (k + 1) % capsuleSegments
		emit(top, rings[last][k], rings[last][next])
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func makeRing(center, tangent, bitangent math.Vector3, radius float64) []math.Vector3 {

	result := make([]math.Vector3, capsuleSegments)

	for k := range result {
		angle := 2 * sysMath.Pi * float64(k) / capsuleSegments

		offset := tangent.MulScalar(sysMath.Cos(angle))
		offset = offset.Add(bitangent.MulScalar(sysMath.Sin(angle)))

		result[k] = center.Add(offset.MulScalar(radius))
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

func decodeVertices(data []byte) ([]math.Vector3, error) {

	if len(data)%vertexSize != 0 {
		return nil, errors.New(
			"vertex data is not a multiple of the vertex size",
			errors.Int("size", len(data)),
		)
	}

	result := make([]math.Vector3, len(data)/vertexSize)

	for i := range result {
		vertex, err := math.Vector3FromBytes32(data[i*vertexSize:])
		if err != nil {
			return nil, err
		}

		result[i] = vertex
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

// getVector3 reads an array of three numbers, such as a center.
func getVector3(value *Kv3Value) (math.Vector3, error) {

	elements := value.GetElements()
	if len(elements) != 3 {
		return math.Vector3Zero, errors.New(
			"vector does not have three components",
			errors.Int("components", len(elements)),
		)
	}

	var components [3]float64

	for i := range elements {
		component, ok := elements[i].GetFloat()
		if !ok {
			return math.Vector3Zero, errors.New(
				"vector component is not a number",
				errors.Int("component", i),
			)
		}

		components[i] = component
	}

	return math.Vector3{
		X: components[0],
		Y: components[1],
		Z: components[2],
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// getRadius reads a radius, which must be a positive number.
func getRadius(value *Kv3Value) (float64, error) {

	radius, ok := value.GetFloat()
	if !ok || radius <= 0 {
		return 0, errors.New(
			"radius is not a positive number",
			errors.Float64("radius", radius),
		)
	}

	return radius, nil
}
