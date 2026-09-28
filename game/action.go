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

// entityChain holds the pointers followed from an entity list
// slot to the player data. The chain ends at the first null
// pointer, so every pointer after it is zero.
type entityChain struct {
	controller uintptr // Player controller
	handle     int32   // Pawn handle, zero while dead
	chunk      uintptr // Entity list chunk of the pawn
	pawn       uintptr // Player pawn
	sceneNode  uintptr // Game scene node of the pawn
	boneArray  uintptr // Bone array of the skeleton
}

////////////////////////////////////////////////////////////////////////////////

// entityCache holds the entity list and the chains of every
// slot resolved by earlier action frames.
type entityCache struct {
	pid    uint32
	client uintptr

	entListBase uintptr
	listEntry   uintptr
	chains      [64]entityChain

	batches int // Scatter batches since the last entity log
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

	// Check for map changes (throttled to once per second)
	if now.Sub(g.lastMapCheck) >= time.Second {
		g.lastMapCheck = now
		g.checkMapChange(memory, client)
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

	offPawnIsAlive, ok := g.GetOffsetsInt("client_dll", "client.dll", "classes", "CCSPlayerController", "fields", "m_bPawnIsAlive")
	if !ok {
		result.Result = ActionResultNoOffset
		return result
	}

	//----------------------------------------------------------------------------//

	// Use scatter for batched entity reads when available. The
	// pointer chains it caches are only valid in the process
	// they were read from.
	scatter := g.scatter
	pid := scanner.Process.GetPid()

	cache := &g.entityCache
	if cache.pid != pid || cache.client != client {
		*cache = entityCache{pid: pid, client: client}
	}

	// Internal CModelState offset to the bone array pointer.
	// This is not exposed in schema dumps and comes from SDK
	// reverse engineering.
	const offBoneArray uintptr = 0x80

	//----------------------------------------------------------------------------//

	// Resolve the entity list when it is not cached. Without
	// scatter, it is resolved and read one entity at a time on
	// every frame.
	if scatter == nil || cache.listEntry == 0 {

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

		if scatter == nil {
			return g.updateActionSequential(
				result, memory, localPawnAddr, entListBase, listEntry,
				offPawnHandle, offHealth, offTeamNum, offOrigin, offEyeAngles,
				offGameSceneNode, offModelState, offBoneArray,
			)
		}

		cache.entListBase = entListBase
		cache.listEntry = listEntry
	}

	//----------------------------------------------------------------------------//

	// Read the entities through the pointer chains cached by
	// earlier frames. Each pass reads every pointer of a chain
	// again along with its data, and the data is only used when
	// none of them changed. A changed chain is followed one
	// pointer further each pass, so a new chain takes a pass
	// per pointer plus one for its data. Chains that are still
	// changing after that are skipped for this frame.
	const maxPasses = 7

	// The FPGA keeps a limited number of bytes of reads in
	// flight, and a batch that needs more waits several
	// milliseconds for each extra round. Every request also has
	// a small chance of stalling for as long, so reads are made
	// in whole pages, which take one request each. Passes are
	// split into batches that stay below that limit, using the
	// pages each part of a batch reads: the entity list pages
	// and globals, a controller and a pawn with its bones.
	const maxBatchPages = 36
	const batchPages = 6
	const controllerPages = 1
	const pawnPages = 4

	readSize := uint32(BoneCount * BoneDataSize)

	var localPawnAddr uintptr
	var settled [64]bool
	entities := make([]ActionEntity, 64)

	pass := 0
	first := 0
	changed := false

	for pass < maxPasses {

		cache.batches++
		scatter.Clear(pid, leech.ScatterFlagDefault)

		// The entity list is read again in every batch to check
		// that it has not moved since it was cached
		scatter.Prepare(client+offLocalPlayerPawn, 8)
		scatter.Prepare(client+offEntityList, 8)
		scatter.Prepare(cache.entListBase+0x10, 8)

		// Queue the chains from `first` until the batch is full
		end := first
		pages := batchPages

		for end < 64 {
			e := end
			chain := &cache.chains[e]

			if !settled[e] {
				cost := 0
				if chain.controller != 0 {
					cost += controllerPages
				}
				if chain.pawn != 0 {
					cost += pawnPages
				}

				if pages+cost > maxBatchPages {
					break
				}
				pages += cost
			}

			end++

			if settled[e] {
				continue
			}

			index := uintptr(chain.handle) & 0x7FFF

			scatter.Prepare(cache.listEntry+uintptr(e+1)*0x70, 8)

			if chain.controller != 0 {
				scatter.Prepare(chain.controller+offPawnHandle, 4)
				scatter.Prepare(chain.controller+offPawnIsAlive, 1)
			}

			if chain.handle != 0 {
				scatter.Prepare(cache.entListBase+0x10+8*(index>>9), 8)
			}

			if chain.chunk != 0 {
				scatter.Prepare(chain.chunk+0x70*(index&0x1FF), 8)
			}

			if chain.pawn != 0 {
				scatter.Prepare(chain.pawn+offHealth, 4)
				scatter.Prepare(chain.pawn+offTeamNum, 4)
				scatter.Prepare(chain.pawn+offOrigin, 12)
				scatter.Prepare(chain.pawn+offEyeAngles, 8)
				scatter.Prepare(chain.pawn+offGameSceneNode, 8)
			}

			if chain.sceneNode != 0 {
				scatter.Prepare(chain.sceneNode+offModelState+offBoneArray, 8)
			}

			// VMM reads a lone read of up to 0x400 bytes as 128
			// byte requests, so the bones are read as the whole
			// pages they are on
			if chain.boneArray != 0 {
				pageStart := chain.boneArray &^ 0xFFF
				pageEnd := (chain.boneArray + uintptr(readSize) + 0xFFF) &^ 0xFFF
				scatter.Prepare(pageStart, uint32(pageEnd-pageStart))
			}
		}

		if err := scatter.ExecuteRead(); err != nil {
			result.Result = ActionResultReadFail
			return result
		}

		localPawnAddr, _ = scatter.ReadPtr(client + offLocalPlayerPawn)
		entListBase, _ := scatter.ReadPtr(client + offEntityList)
		listEntry, _ := scatter.ReadPtr(cache.entListBase + 0x10)

		// Left the game, the next frame starts over
		if localPawnAddr == 0 {
			*cache = entityCache{pid: pid, client: client}
			result.Result = ActionResultSuccess
			return result
		}

		// The entity list moved, which invalidates every chain,
		// so the next frame resolves it again
		if entListBase != cache.entListBase || listEntry != cache.listEntry {
			*cache = entityCache{pid: pid, client: client}
			result.Result = ActionResultNoEntList
			return result
		}

		for e := first; e < end; e++ {
			if settled[e] {
				continue
			}

			chain := &cache.chains[e]
			index := uintptr(chain.handle) & 0x7FFF

			// Read each pointer again through the cached pointer
			// before it, ending the chain at the first null one
			var fresh entityChain
			fresh.controller, _ = scatter.ReadPtr(cache.listEntry + uintptr(e+1)*0x70)

			if chain.controller != 0 && fresh.controller != 0 {

				// Pawns of dead and disconnected players never
				// become entities, so their chains end here and
				// their pawns do not add to the batch
				alive, _ := scatter.Read(chain.controller+offPawnIsAlive, 1)
				handle, _ := scatter.ReadInt32(chain.controller + offPawnHandle)

				if len(alive) == 1 && alive[0] != 0 && handle != 0 && handle != -1 {
					fresh.handle = handle
				}
			}

			if chain.handle != 0 && fresh.handle != 0 {
				fresh.chunk, _ = scatter.ReadPtr(cache.entListBase + 0x10 + 8*(index>>9))
			}

			if chain.chunk != 0 && fresh.chunk != 0 {
				fresh.pawn, _ = scatter.ReadPtr(chain.chunk + 0x70*(index&0x1FF))
			}

			if chain.pawn != 0 && fresh.pawn != 0 {
				fresh.sceneNode, _ = scatter.ReadPtr(chain.pawn + offGameSceneNode)
			}

			if chain.sceneNode != 0 && fresh.sceneNode != 0 {
				fresh.boneArray, _ = scatter.ReadPtr(chain.sceneNode + offModelState + offBoneArray)
			}

			// Data read through a changed chain may be stale
			if fresh != *chain {
				*chain = fresh
				changed = true
				continue
			}

			settled[e] = true

			if chain.pawn == 0 {
				continue
			}

			health, _ := scatter.ReadInt32(chain.pawn + offHealth)
			if health <= 0 {
				continue
			}

			team, _ := scatter.ReadInt32(chain.pawn + offTeamNum)

			originData, err := scatter.Read(chain.pawn+offOrigin, 12)
			if err != nil {
				continue
			}
			origin, _ := math.Vector3FromBytes32(originData)

			if origin.X == 0 && origin.Y == 0 && origin.Z == 0 {
				continue
			}

			anglesData, _ := scatter.Read(chain.pawn+offEyeAngles, 8)
			angles, _ := math.Vector2FromBytes32(anglesData)

			entity := ActionEntity{
				Valid:  true,
				Index:  e,
				Origin: origin,
				Angles: angles,
				Health: health,
				Team:   team,
				Head:   math.Vector3{X: origin.X, Y: origin.Y, Z: origin.Z + 72},
			}

			// The full contiguous block up to BoneCount is read
			// so every indexed bone is covered in one read
			if chain.boneArray != 0 {
				boneData, err := scatter.Read(chain.boneArray, readSize)
				if err == nil {
					entity.Bones = decodeBonePositions(boneData)
				}
			}

			if entity.Bones.Valid {
				entity.Head = entity.Bones.Pos[BoneHead]
				entity.Neck = entity.Bones.Pos[BoneNeck]
				entity.Body = entity.Bones.Pos[BoneSpine2]
			}

			entities[e] = entity
		}

		// Continue with the chains after this batch that are not
		// settled yet
		first = end
		for first < 64 && settled[first] {
			first++
		}

		if first < 64 {
			continue
		}

		// Start another pass when a chain changed in this one
		if !changed {
			break
		}

		pass++
		first = 0
		changed = false
	}

	//----------------------------------------------------------------------------//

	// Keep the entities in slot order and find the local player
	count := 0
	playerIdx := -1

	for e := range entities {
		if !entities[e].Valid {
			continue
		}

		if cache.chains[e].pawn == localPawnAddr {
			playerIdx = count
		}

		entities[count] = entities[e]
		count++
	}

	entities = entities[:count]

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
			logger.Int("batches", cache.batches),
		)
		cache.batches = 0
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
