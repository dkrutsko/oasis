package game

import (
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Bone indices for the CS2 player skeleton. These map to the
// engine bone array returned by CModelState and correspond to
// the standard player model rig.
const (
	BoneHead          = 6
	BoneNeck          = 5
	BoneSpine3        = 4
	BoneSpine2        = 3
	BoneSpine1        = 2
	BonePelvis        = 1
	BoneLeftShoulder  = 13
	BoneLeftElbow     = 14
	BoneLeftHand      = 15
	BoneRightShoulder = 8
	BoneRightElbow    = 9
	BoneRightHand     = 10
	BoneLeftHip       = 25
	BoneLeftKnee      = 26
	BoneLeftFoot      = 27
	BoneRightHip      = 22
	BoneRightKnee     = 23
	BoneRightFoot     = 24

	// Total number of bones we read per entity. This is the
	// highest index plus one so we can read a single contiguous
	// block from the bone array.
	BoneCount = 28

	// Size of a single CBoneData entry in the engine bone
	// array. Each entry is 32 bytes: position (float x, y, z),
	// scale (float), rotation quaternion (float x, y, z, w).
	BoneDataSize = 32
)

////////////////////////////////////////////////////////////////////////////////

// BonePositions holds world-space positions for each bone in
// the player skeleton. Only the indices listed above are
// populated. Valid indicates whether the bone data was
// successfully read from the game.
type BonePositions struct {
	Valid bool
	Pos   [BoneCount]math.Vector3
}

////////////////////////////////////////////////////////////////////////////////

// decodeBonePositions extracts bone positions from a raw byte
// buffer containing BoneCount contiguous CBoneData entries.
// Each entry is BoneDataSize bytes with the position stored
// as three little-endian float32 values at the start.
func decodeBonePositions(data []byte) BonePositions {

	var bones BonePositions

	if len(data) < BoneCount*BoneDataSize {
		return bones
	}

	for i := 0; i < BoneCount; i++ {
		off := i * BoneDataSize
		pos, err := math.Vector3FromBytes32(data[off : off+12])
		if err != nil {
			return bones
		}
		bones.Pos[i] = pos
	}

	// Sanity check: head position at world origin likely
	// means the bone data is not yet initialized.
	head := bones.Pos[BoneHead]
	if head.X == 0 && head.Y == 0 && head.Z == 0 {
		return bones
	}

	bones.Valid = true
	return bones
}

////////////////////////////////////////////////////////////////////////////////

// readBonesSequential reads bone positions for a single pawn
// using sequential memory reads. Used as the fallback path
// when scatter is not available.
func readBonesSequential(
	memory *leech.Memory,
	pawn uintptr,
	offGameSceneNode uintptr,
	offModelState uintptr,
	offBoneArray uintptr,
) BonePositions {

	//----------------------------------------------------------------------------//

	sceneNode, err := memory.ReadPtr(pawn + offGameSceneNode)
	if err != nil || sceneNode == 0 {
		return BonePositions{}
	}

	//----------------------------------------------------------------------------//

	boneArrayPtr, err := memory.ReadPtr(sceneNode + offModelState + offBoneArray)
	if err != nil || boneArrayPtr == 0 {
		return BonePositions{}
	}

	//----------------------------------------------------------------------------//

	data, err := memory.ReadData(boneArrayPtr, BoneCount*BoneDataSize)
	if err != nil {
		return BonePositions{}
	}

	return decodeBonePositions(data)

	//----------------------------------------------------------------------------//
}
