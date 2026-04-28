package game

import (
	sysMath "math"
	"strconv"
	"time"

	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

type ActionResult uint8

const (
	ActionResultSuccess ActionResult = 0x00
	ActionResultNoValue ActionResult = 0x01

	ActionResultNoProcess ActionResult = 0x10
	ActionResultNoOffset  ActionResult = 0x11
	ActionResultReadFail  ActionResult = 0x12

	ActionResultNoEntList ActionResult = 0x20
)

////////////////////////////////////////////////////////////////////////////////

func (s ActionResult) String() string {

	switch s {
	case ActionResultSuccess:
		return "Success"
	case ActionResultNoValue:
		return "NoValue"

	case ActionResultNoProcess:
		return "NoProcess"
	case ActionResultNoOffset:
		return "NoOffset"
	case ActionResultReadFail:
		return "ReadFail"

	case ActionResultNoEntList:
		return "NoEntList"
	}

	return strconv.FormatUint(uint64(s), 10)
}

////////////////////////////////////////////////////////////////////////////////

type ActionEntity struct {
	Valid bool // If entity is valid
	Index int  // Index in entity list

	Origin math.Vector3 // Position of entity
	Angles math.Vector2 // Entity eye angles
	View   math.Vector3 // Entity view offset

	Body math.Vector3 // Entity body position
	Neck math.Vector3 // Entity neck position
	Head math.Vector3 // Entity head position

	Health int32 // Health of entity
	//	Flags  Flags_t // Some entity flags
	Scoped bool  // If entity scoped
	Team   int32 // Entity team number

	Distance float64 // Distance to player
}

////////////////////////////////////////////////////////////////////////////////

type ActionState struct {
	Result ActionResult
	Time   time.Time

	//	ClientState ClientState_t // State of the client

	Entities []ActionEntity // Game entities list
	Player   *ActionEntity  // Local player entity
}

////////////////////////////////////////////////////////////////////////////////

func NewActionState() *ActionState {

	return &ActionState{
		Result: ActionResultNoValue,
		Time:   time.Now(),
	}
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) updateAction(scanner *ScannerState) *ActionState {

	//----------------------------------------------------------------------------//

	// Create the final result
	result := NewActionState()

	//----------------------------------------------------------------------------//

	// Check if a game has been selected and is currently valid
	if scanner == nil || scanner.Result != ScannerResultSuccess {
		result.Result = ActionResultNoProcess
		return result
	}

	// Grab the client module base address
	client := scanner.Client.GetBase()

	// Use the cached memory for entity reads. Clear at the
	// start of each frame so all reads within the frame are
	// fresh but nearby reads coalesce into single DMA calls.
	memory := g.memory
	if memory == nil {
		memory = scanner.Process.GetMemory()
	}
	memory.ClearCache()

	//----------------------------------------------------------------------------//

	// Get module-level offsets
	offLocalPlayerPawn, ok := g.GetOffsetsInt("offsets", "client.dll", "dwLocalPlayerPawn")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	offEntityList, ok := g.GetOffsetsInt("offsets", "client.dll", "dwEntityList")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	//----------------------------------------------------------------------------//

	// Get field offsets from schema data
	offPawnHandle, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "CCSPlayerController", "fields", "m_hPlayerPawn")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	offHealth, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "C_BaseEntity", "fields", "m_iHealth")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	offTeamNum, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "C_BaseEntity", "fields", "m_iTeamNum")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	offOrigin, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "C_BasePlayerPawn", "fields", "m_vOldOrigin")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	offEyeAngles, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "C_CSPlayerPawn", "fields", "m_angEyeAngles")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	//----------------------------------------------------------------------------//

	// Check if the player is currently in-game
	localPawnAddr, err := memory.ReadPtr(client + offLocalPlayerPawn)
	if err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	// Not in game
	if localPawnAddr == 0 {
		result.Result = ActionResultSuccess
		return result
	}

	//----------------------------------------------------------------------------//

	// Read entity list base pointer
	entListBase, err := memory.ReadPtr(client + offEntityList)
	if err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	if entListBase == 0 {
		result.Result = ActionResultNoEntList
		return result
	}

	// Read first chunk of entity list (chunk pointer at +0x10)
	listEntry, err := memory.ReadPtr(entListBase + 0x10)
	if err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	if listEntry == 0 {
		result.Result = ActionResultNoEntList
		return result
	}

	//----------------------------------------------------------------------------//

	// Read all game entities
	entities := make([]ActionEntity, 0, 64)
	playerIdx := -1

	for e := 0; e < 64; e++ {

		// Read controller from chunk (entity index e+1, skip world at 0)
		controller, err := memory.ReadPtr(listEntry + uintptr(e+1)*0x70)
		if err != nil || controller == 0 {
			continue
		}

		// Read pawn handle from controller
		pawnHandle, err := memory.ReadInt32(controller + offPawnHandle)
		if err != nil || pawnHandle == 0 || pawnHandle == -1 {
			continue
		}

		// Decode entity handle to get pawn address
		entityIndex := uintptr(pawnHandle) & 0x7FFF
		chunkIdx := entityIndex >> 9
		entryIdx := entityIndex & 0x1FF

		pawnChunk, err := memory.ReadPtr(entListBase + 0x10 + 8*chunkIdx)
		if err != nil || pawnChunk == 0 {
			continue
		}

		pawn, err := memory.ReadPtr(pawnChunk + 0x70*entryIdx)
		if err != nil || pawn == 0 {
			continue
		}

		// Read health and skip dead entities
		health, err := memory.ReadInt32(pawn + offHealth)
		if err != nil || health <= 0 {
			continue
		}

		// Read team number
		team, err := memory.ReadInt32(pawn + offTeamNum)
		if err != nil {
			continue
		}

		// Read position
		origin, err := ReadVector3(memory, pawn+offOrigin)
		if err != nil {
			continue
		}

		// Skip entities at world origin
		if origin.X == 0 && origin.Y == 0 && origin.Z == 0 {
			continue
		}

		// Read eye angles
		angles, _ := ReadVector2(memory, pawn+offEyeAngles)

		entity := ActionEntity{
			Valid:  true,
			Index:  e,
			Origin: origin,
			Angles: angles,
			Health: health,
			Team:   team,
			Head:   math.Vector3{X: origin.X, Y: origin.Y, Z: origin.Z + 72},
		}

		entities = append(entities, entity)

		// Check if this pawn is the local player
		if pawn == localPawnAddr {
			playerIdx = len(entities) - 1
		}
	}

	//----------------------------------------------------------------------------//

	result.Entities = entities

	if playerIdx >= 0 {
		result.Player = &result.Entities[playerIdx]
	}

	// Calculate distances from local player
	if result.Player != nil {
		for i := range result.Entities {
			if &result.Entities[i] == result.Player {
				continue
			}

			dx := result.Entities[i].Origin.X - result.Player.Origin.X
			dy := result.Entities[i].Origin.Y - result.Player.Origin.Y
			dz := result.Entities[i].Origin.Z - result.Player.Origin.Z
			result.Entities[i].Distance = sysMath.Sqrt(dx*dx + dy*dy + dz*dz)
		}
	}

	//----------------------------------------------------------------------------//

	// Specify that scan was successful
	result.Result = ActionResultSuccess
	return result

	//----------------------------------------------------------------------------//
}
