package main

// Conversion of decoded physics data into packed triangles. This follows
// the original CS2-Phys-Extractor (MIT License, Copyright (c) 2025
// LAITHCOOL) exactly, including where each loop stops and which errors skip
// a hull or keep a partial mesh, so the .tri output is byte for byte the
// same. See NOTICE.

import (
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Safety limit on edges walked per hull face, from the original
const maxHullFaceEdges = 1000

////////////////////////////////////////////////////////////////////////////////

// PhysicsKeys are the object fields ConvertPhysics reads. Passing them to
// `DecodeKv3` avoids storing the rest of the physics data.
var PhysicsKeys = map[string]bool{
	"m_collisionAttributes":      true,
	"m_CollisionGroupString":     true,
	"m_parts":                    true,
	"m_rnShape":                  true,
	"m_hulls":                    true,
	"m_meshes":                   true,
	"m_nCollisionAttributeIndex": true,
	"m_Hull":                     true,
	"m_VertexPositions":          true,
	"m_Vertices":                 true,
	"m_Faces":                    true,
	"m_Edges":                    true,
	"m_Mesh":                     true,
	"m_Triangles":                true,
}

////////////////////////////////////////////////////////////////////////////////

// TriangleFunc receives each triangle as three 12 byte vertices, in the
// order they belong in the .tri file.
type TriangleFunc func(v1, v2, v3 []byte)

// PhysicsResult holds the number of triangles and processed shapes.
type PhysicsResult struct {
	Triangles int
	Hulls     int
	Meshes    int
}

////////////////////////////////////////////////////////////////////////////////

// ConvertPhysics turns the hulls and meshes of the "default" collision group
// into triangles. Hull triangles come first, then mesh triangles.
func ConvertPhysics(phys *Value, emit TriangleFunc) PhysicsResult {

	var result PhysicsResult

	count := func(v1, v2, v3 []byte) {
		result.Triangles++
		emit(v1, v2, v3)
	}

	indices := getCollisionAttributeIndices(phys)

	// Hulls and meshes both live on the first part's shape
	shape := phys.GetChild("m_parts").GetElement(0).GetChild("m_rnShape")

	result.Hulls = convertHulls(shape, indices, count)
	result.Meshes = convertMeshes(shape, indices, count)

	return result
}

////////////////////////////////////////////////////////////////////////////////

// getCollisionAttributeIndices returns the indices of the collision
// attributes in the "default" group. Like the original it stops at the first
// attribute without a group name and falls back to index 0.
func getCollisionAttributeIndices(phys *Value) []int {

	attributes := phys.GetChild("m_collisionAttributes")
	var indices []int

	for index := 0; ; index++ {
		group := attributes.GetElement(index).GetChild("m_CollisionGroupString").GetText()
		if group == "" {
			break
		}

		clean := strings.TrimSpace(strings.Trim(group, `"`))
		if strings.EqualFold(clean, "default") || clean == "0" {
			indices = append(indices, index)
		}
	}

	if len(indices) == 0 {
		indices = append(indices, 0)
	}

	return indices
}

////////////////////////////////////////////////////////////////////////////////

// convertHulls appends the triangles of every convex hull in the default
// collision group. Hull faces are stored as half-edge loops and each face
// is split into a triangle fan.
func convertHulls(shape *Value, indices []int, emit TriangleFunc) int {

	hulls := shape.GetChild("m_hulls")
	processed := 0

	for index := 0; ; index++ {
		hull := hulls.GetElement(index)

		// The original stops at the first hull without a collision index
		text := hull.GetChild("m_nCollisionAttributeIndex").GetText()
		if text == "" {
			break
		}
		if !hasCollisionIndex(indices, text) {
			continue
		}

		hullData := hull.GetChild("m_Hull")

		// Newer data names the vertices `m_VertexPositions`, older `m_Vertices`
		vertexData := hullData.GetChild("m_VertexPositions")
		if isEmpty(vertexData) {
			vertexData = hullData.GetChild("m_Vertices")
		}
		if isEmpty(vertexData) {
			continue
		}

		vertices, vertexCount, err := getVertices(vertexData.GetBlob())
		if err != nil {
			logger.Warn(
				"failed to process hull",
				logger.Int("hull", index),
				logger.Error("error", err),
			)
			continue
		}

		faces := hullData.GetChild("m_Faces").GetBlob()
		edges := hullData.GetChild("m_Edges").GetBlob()

		// Each half-edge is 4 bytes: next, twin, origin vertex, face
		if len(edges)%4 != 0 {
			logger.Warn(
				"hull edge data is not a multiple of 4 bytes",
				logger.Int("hull", index),
				logger.Int("size", len(edges)),
			)
			continue
		}
		edgeCount := len(edges) / 4

		if vertexCount > 0 && len(faces) > 0 && edgeCount > 0 {
			appendHullTriangles(vertices, vertexCount, faces, edges, edgeCount, emit)
			processed++
		}
	}

	return processed
}

////////////////////////////////////////////////////////////////////////////////

// appendHullTriangles walks each face's edge loop and emits a fan of
// triangles anchored at the face's first edge.
func appendHullTriangles(vertices []byte, vertexCount int, faces, edges []byte, edgeCount int, emit TriangleFunc) {

	for _, startEdge := range faces {
		start := int(startEdge)
		if start >= edgeCount {
			continue
		}

		edge := int(edges[start*4])

		for iterations := 0; edge != start && iterations < maxHullFaceEdges; iterations++ {
			if edge >= edgeCount {
				break
			}

			next := int(edges[edge*4])
			if next >= edgeCount {
				break
			}

			o1 := int(edges[start*4+2])
			o2 := int(edges[edge*4+2])
			o3 := int(edges[next*4+2])

			if o1 < vertexCount && o2 < vertexCount && o3 < vertexCount {
				emit(vertices[o1*12:o1*12+12], vertices[o2*12:o2*12+12], vertices[o3*12:o3*12+12])
			}

			edge = next
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// convertMeshes appends the triangles of every indexed triangle mesh in the
// default collision group.
func convertMeshes(shape *Value, indices []int, emit TriangleFunc) int {

	meshes := shape.GetChild("m_meshes")
	processed := 0

	for index := 0; ; index++ {
		mesh := meshes.GetElement(index)

		text := mesh.GetChild("m_nCollisionAttributeIndex").GetText()
		if text == "" {
			break
		}
		if !hasCollisionIndex(indices, text) {
			continue
		}

		meshData := mesh.GetChild("m_Mesh")
		triangles := meshData.GetChild("m_Triangles").GetBlob()
		indexCount := len(triangles) / 4

		vertices, vertexCount, err := getVertices(meshData.GetChild("m_Vertices").GetBlob())
		if err != nil {
			logger.Warn(
				"failed to process mesh",
				logger.Int("mesh", index),
				logger.Error("error", err),
			)
			continue
		}

		if vertexCount > 0 && indexCount > 0 {
			// Like the original, triangles added before an invalid index
			// are kept
			if err := appendMeshTriangles(vertices, vertexCount, triangles, indexCount, emit); err != nil {
				logger.Warn(
					"failed to process mesh",
					logger.Int("mesh", index),
					logger.Error("error", err),
				)
				continue
			}
			processed++
		}
	}

	return processed
}

////////////////////////////////////////////////////////////////////////////////

func appendMeshTriangles(vertices []byte, vertexCount int, triangles []byte, indexCount int, emit TriangleFunc) error {

	vertexAt := func(i int) (int, bool) {
		if i >= indexCount {
			return 0, false
		}
		v := int(int32(binary.LittleEndian.Uint32(triangles[i*4:])))
		return v, v >= 0 && v < vertexCount
	}

	for i := 0; i < indexCount; i += 3 {
		a, okA := vertexAt(i)
		b, okB := vertexAt(i + 1)
		c, okC := vertexAt(i + 2)

		if !okA || !okB || !okC {
			return errors.New(
				"triangle index is out of range",
				errors.Int("index", i),
			)
		}

		emit(vertices[a*12:a*12+12], vertices[b*12:b*12+12], vertices[c*12:c*12+12])
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// getVertices checks that a blob holds whole xyz float32 triplets. Trailing
// bytes that do not form a whole float are ignored, as in the original. The
// vertices are copied as raw bytes so the output bits are exactly the input
// bits.
func getVertices(blob []byte) ([]byte, int, error) {

	floats := len(blob) / 4
	if floats%3 != 0 {
		return nil, 0, errors.New(
			"vertex data is not a multiple of 3 floats",
			errors.Int("size", len(blob)),
		)
	}
	return blob, floats / 3, nil
}

////////////////////////////////////////////////////////////////////////////////

// hasCollisionIndex parses the collision index the way C# `int.TryParse`
// did and checks it against the default group indices.
func hasCollisionIndex(indices []int, text string) bool {

	n, err := strconv.ParseInt(strings.TrimSpace(text), 10, 32)
	if err != nil {
		return false
	}

	for _, index := range indices {
		if index == int(n) {
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

// isEmpty matches the original's empty check: missing values, empty blobs
// and empty strings are empty, anything else is not.
func isEmpty(v *Value) bool {

	if v == nil {
		return true
	}

	switch v.Kind {
	case KindBlob:
		return len(v.blob) == 0
	case KindString:
		return v.str == ""
	default:
		return false
	}
}
