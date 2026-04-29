package game

import (
	sysMath "math"
	"strconv"
	"time"

	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
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

	Bones BonePositions // Full skeleton bone data

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

	// Periodically refresh VMM's TLB cache. With -norefresh,
	// page table translations go stale when the OS remaps
	// pages. This causes reads to return zeros for valid
	// addresses and entities silently disappear. The refresh
	// takes ~100ms so we limit it to every 3 seconds.
	now := time.Now()
	if now.Sub(g.lastTlbRefresh) >= 3*time.Second {
		g.lastTlbRefresh = now
		g.options.Leech.SetConfig(leech.ConfigRefreshFreqTlb, 1)
	}

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

	offGameSceneNode, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "C_BaseEntity", "fields", "m_pGameSceneNode")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	offModelState, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "CSkeletonInstance", "fields", "m_modelState")
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
		if now.Sub(g.lastEntityLog) >= time.Second {
			g.lastEntityLog = now
			logger.Dbg("not in game, local pawn is null")
		}
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

	// Use scatter for batched entity reads when available.
	// This reduces ~30-40 individual DMA round-trips to ~5
	// batched executions.
	scatter := g.scatter
	pid := scanner.Process.GetPid()

	// Internal CModelState offset to the bone array pointer.
	// This is not exposed in schema dumps and comes from SDK
	// reverse engineering.
	const offBoneArray uintptr = 0x80

	if scatter == nil {
		return g.updateActionSequential(
			result, memory, localPawnAddr, entListBase, listEntry,
			offPawnHandle, offHealth, offTeamNum, offOrigin, offEyeAngles,
			offGameSceneNode, offModelState, offBoneArray,
		)
	}

	//----------------------------------------------------------------------------//

	// Pass 1: Read all 64 controller pointers
	for e := 0; e < 64; e++ {
		scatter.Prepare(listEntry+uintptr(e+1)*0x70, 8)
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	var controllers [64]uintptr
	for e := 0; e < 64; e++ {
		controllers[e], _ = scatter.ReadPtr(listEntry + uintptr(e+1)*0x70)
	}

	//----------------------------------------------------------------------------//

	// Pass 2: Read pawn handles from valid controllers
	scatter.Clear(pid, leech.ScatterFlagDefault)

	for e := 0; e < 64; e++ {
		if controllers[e] != 0 {
			scatter.Prepare(controllers[e]+offPawnHandle, 4)
		}
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	var pawnHandles [64]int32
	var chunkIdxs [64]uintptr
	var entryIdxs [64]uintptr

	for e := 0; e < 64; e++ {
		if controllers[e] == 0 {
			continue
		}

		h, _ := scatter.ReadInt32(controllers[e] + offPawnHandle)
		if h == 0 || h == -1 {
			continue
		}

		pawnHandles[e] = h
		entityIndex := uintptr(h) & 0x7FFF
		chunkIdxs[e] = entityIndex >> 9
		entryIdxs[e] = entityIndex & 0x1FF
	}

	//----------------------------------------------------------------------------//

	// Pass 3: Read pawn chunk pointers
	scatter.Clear(pid, leech.ScatterFlagDefault)

	for e := 0; e < 64; e++ {
		if pawnHandles[e] == 0 {
			continue
		}
		scatter.Prepare(entListBase+0x10+8*chunkIdxs[e], 8)
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	var pawnChunks [64]uintptr
	for e := 0; e < 64; e++ {
		if pawnHandles[e] == 0 {
			continue
		}
		pawnChunks[e], _ = scatter.ReadPtr(entListBase + 0x10 + 8*chunkIdxs[e])
	}

	//----------------------------------------------------------------------------//

	// Pass 4: Read pawn pointers from chunks
	scatter.Clear(pid, leech.ScatterFlagDefault)

	for e := 0; e < 64; e++ {
		if pawnChunks[e] == 0 {
			continue
		}
		scatter.Prepare(pawnChunks[e]+0x70*entryIdxs[e], 8)
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	var pawns [64]uintptr
	for e := 0; e < 64; e++ {
		if pawnChunks[e] == 0 {
			continue
		}
		pawns[e], _ = scatter.ReadPtr(pawnChunks[e] + 0x70*entryIdxs[e])
	}

	//----------------------------------------------------------------------------//

	// Pass 5: Read entity data and scene node pointers for all valid pawns
	scatter.Clear(pid, leech.ScatterFlagDefault)

	for e := 0; e < 64; e++ {
		if pawns[e] == 0 {
			continue
		}
		scatter.Prepare(pawns[e]+offHealth, 4)
		scatter.Prepare(pawns[e]+offTeamNum, 4)
		scatter.Prepare(pawns[e]+offOrigin, 12)
		scatter.Prepare(pawns[e]+offEyeAngles, 8)
		scatter.Prepare(pawns[e]+offGameSceneNode, 8)
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	//----------------------------------------------------------------------------//

	// Decode results and build entity list
	entities := make([]ActionEntity, 0, 64)
	playerIdx := -1

	// Track scene node pointers and entity-to-slot mapping
	// for the bone reading passes that follow.
	var sceneNodes [64]uintptr
	entitySlots := make([]int, 0, 64)

	for e := 0; e < 64; e++ {
		if pawns[e] == 0 {
			continue
		}

		health, _ := scatter.ReadInt32(pawns[e] + offHealth)
		if health <= 0 {
			continue
		}

		team, _ := scatter.ReadInt32(pawns[e] + offTeamNum)

		originData, err := scatter.Read(pawns[e]+offOrigin, 12)
		if err != nil {
			continue
		}
		origin, _ := math.Vector3FromBytes32(originData)

		if origin.X == 0 && origin.Y == 0 && origin.Z == 0 {
			continue
		}

		anglesData, _ := scatter.Read(pawns[e]+offEyeAngles, 8)
		angles, _ := math.Vector2FromBytes32(anglesData)

		sceneNodes[e], _ = scatter.ReadPtr(pawns[e] + offGameSceneNode)

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
		entitySlots = append(entitySlots, e)

		if pawns[e] == localPawnAddr {
			playerIdx = len(entities) - 1
		}
	}

	//----------------------------------------------------------------------------//

	// Pass 6: Read bone array pointers from scene nodes
	scatter.Clear(pid, leech.ScatterFlagDefault)

	for _, slot := range entitySlots {
		if sceneNodes[slot] != 0 {
			scatter.Prepare(sceneNodes[slot]+offModelState+offBoneArray, 8)
		}
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	var boneArrays [64]uintptr
	for _, slot := range entitySlots {
		if sceneNodes[slot] != 0 {
			boneArrays[slot], _ = scatter.ReadPtr(sceneNodes[slot] + offModelState + offBoneArray)
		}
	}

	//----------------------------------------------------------------------------//

	// Pass 7: Read bone positions for all entities with valid
	// bone arrays. We read the full contiguous block up to
	// BoneCount so every indexed bone is covered in one read.
	scatter.Clear(pid, leech.ScatterFlagDefault)

	readSize := uint32(BoneCount * BoneDataSize)
	for _, slot := range entitySlots {
		if boneArrays[slot] != 0 {
			scatter.Prepare(boneArrays[slot], readSize)
		}
	}

	if err := scatter.ExecuteRead(); err != nil {
		result.Result = ActionResultReadFail
		return result
	}

	for i, slot := range entitySlots {
		if boneArrays[slot] == 0 {
			continue
		}

		boneData, err := scatter.Read(boneArrays[slot], readSize)
		if err != nil {
			continue
		}

		bones := decodeBonePositions(boneData)
		entities[i].Bones = bones

		if bones.Valid {
			entities[i].Head = bones.Pos[BoneHead]
			entities[i].Neck = bones.Pos[BoneNeck]
			entities[i].Body = bones.Pos[BoneSpine2]
		}
	}

	//----------------------------------------------------------------------------//

	// Reset scatter for next frame
	scatter.Clear(pid, leech.ScatterFlagDefault)

	//----------------------------------------------------------------------------//

	result.Entities = entities

	if playerIdx >= 0 {
		result.Player = &result.Entities[playerIdx]
	}

	// Debug: log entity count once per second
	if now := time.Now(); now.Sub(g.lastEntityLog) >= time.Second {
		g.lastEntityLog = now
		alive := 0
		for i := range entities {
			if entities[i].Valid && entities[i].Health > 0 {
				alive++
			}
		}
		logger.Dbg("entity count",
			logger.Int("total", len(entities)),
			logger.Int("alive", alive),
			logger.Bool("has_player", playerIdx >= 0),
		)
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

////////////////////////////////////////////////////////////////////////////////

func (g *Game) updateActionSequential(
	result *ActionState,
	memory *leech.Memory,
	localPawnAddr uintptr,
	entListBase uintptr,
	listEntry uintptr,
	offPawnHandle uintptr,
	offHealth uintptr,
	offTeamNum uintptr,
	offOrigin uintptr,
	offEyeAngles uintptr,
	offGameSceneNode uintptr,
	offModelState uintptr,
	offBoneArray uintptr,
) *ActionState {

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

		// Read bone data through the skeleton instance
		entity.Bones = readBonesSequential(
			memory, pawn,
			offGameSceneNode, offModelState, offBoneArray,
		)

		if entity.Bones.Valid {
			entity.Head = entity.Bones.Pos[BoneHead]
			entity.Neck = entity.Bones.Pos[BoneNeck]
			entity.Body = entity.Bones.Pos[BoneSpine2]
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
