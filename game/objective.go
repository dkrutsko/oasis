package game

import (
	"time"

	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// BombInfo holds the state of the C4 bomb on defuse maps.
type BombInfo struct {
	Carrier int          // Entity slot of the carrier, -1 when not carried
	Dropped bool         // Bomb is on the ground
	Origin  math.Vector3 // Position of the bomb

	Planted      bool  // Bomb has been planted
	BombSite     int32 // Site the bomb was planted at (0 = A, 1 = B)
	Ticking      bool  // Planted bomb is counting down
	BeingDefused bool  // Someone is defusing the bomb
	Defused      bool  // Bomb has been defused
	Exploded     bool  // Bomb has exploded
}

////////////////////////////////////////////////////////////////////////////////

// HostageInfo holds the state of a single hostage on hostage
// maps.
type HostageInfo struct {
	Origin  math.Vector3 // Hostage position
	Rescued bool         // Hostage has been rescued
	Grabber int          // Entity slot of the carrier, -1 when not carried
}

////////////////////////////////////////////////////////////////////////////////

const (
	c4ClassName            = "C_C4"
	playerResClassName     = "C_CSPlayerResource"
	playerResDesignerName  = "cs_player_manager"
	entityScanInterval     = time.Second
)

////////////////////////////////////////////////////////////////////////////////

// readObjectives reads bomb and hostage state from game
// memory. Called after the entity loop to populate the
// objective fields of the action state.
func (g *Game) readObjectives(
	memory *leech.Memory,
	client uintptr,
	entListBase uintptr,
	pawnHandles [64]int32,
) (*BombInfo, []HostageInfo) {

	//----------------------------------------------------------------------------//

	offGameRules, ok := g.GetOffsetsInt("offsets", "client.dll", "dwGameRules")
	if !ok {
		return nil, nil
	}

	gameRules, err := memory.ReadPtr(client + offGameRules)
	if err != nil || gameRules == 0 {
		return nil, nil
	}

	//----------------------------------------------------------------------------//

	offMapHasBombTarget, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_CSGameRules", "fields", "m_bMapHasBombTarget",
	)
	if !ok {
		return nil, nil
	}

	offMapHasRescueZone, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_CSGameRules", "fields", "m_bMapHasRescueZone",
	)
	if !ok {
		return nil, nil
	}

	hasBombTarget, _ := memory.ReadBool(gameRules + offMapHasBombTarget)
	hasRescueZone, _ := memory.ReadBool(gameRules + offMapHasRescueZone)

	//----------------------------------------------------------------------------//

	// Periodic entity list scan for non-player entities
	// (C4 weapon, player resource). The scan is throttled so
	// the per-frame cost is just the cached address reads.
	now := time.Now()
	if now.Sub(g.lastEntityScan) >= entityScanInterval {
		g.lastEntityScan = now

		offHighest, _ := g.GetOffsetsInt(
			"offsets", "client.dll",
			"dwGameEntitySystem_highestEntityIndex",
		)
		highest, _ := memory.ReadInt32(entListBase + offHighest)

		g.scanEntities(memory, entListBase, hasBombTarget, hasRescueZone, int(highest))
	}

	//----------------------------------------------------------------------------//

	var bomb *BombInfo
	if hasBombTarget {
		bomb = g.readBombInfo(memory, client, gameRules, pawnHandles)
	}

	var hostages []HostageInfo
	if hasRescueZone {
		hostages = g.readHostageInfo(memory, entListBase, pawnHandles)
	}

	return bomb, hostages

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readBombInfo reads the C4 weapon and planted bomb state.
func (g *Game) readBombInfo(
	memory *leech.Memory,
	client uintptr,
	gameRules uintptr,
	pawnHandles [64]int32,
) *BombInfo {

	//----------------------------------------------------------------------------//

	offWeaponC4, ok := g.GetOffsetsInt("offsets", "client.dll", "dwWeaponC4")
	if !ok {
		return nil
	}

	offPlantedC4, ok := g.GetOffsetsInt("offsets", "client.dll", "dwPlantedC4")
	if !ok {
		return nil
	}

	offOwnerEntity, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_BaseEntity", "fields", "m_hOwnerEntity",
	)
	if !ok {
		return nil
	}

	offGameSceneNode, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_BaseEntity", "fields", "m_pGameSceneNode",
	)
	if !ok {
		return nil
	}

	offAbsOrigin, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"CGameSceneNode", "fields", "m_vecAbsOrigin",
	)
	if !ok {
		return nil
	}

	offBombTicking, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_PlantedC4", "fields", "m_bBombTicking",
	)
	if !ok {
		return nil
	}

	offBombSite, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_PlantedC4", "fields", "m_nBombSite",
	)
	if !ok {
		return nil
	}

	offBeingDefused, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_PlantedC4", "fields", "m_bBeingDefused",
	)
	if !ok {
		return nil
	}

	offBombDefused, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_PlantedC4", "fields", "m_bBombDefused",
	)
	if !ok {
		return nil
	}

	offHasExploded, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_PlantedC4", "fields", "m_bHasExploded",
	)
	if !ok {
		return nil
	}

	offBombDropped, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_CSGameRules", "fields", "m_bBombDropped",
	)
	if !ok {
		return nil
	}

	//----------------------------------------------------------------------------//

	info := &BombInfo{Carrier: -1}

	// Check for a planted bomb
	plantedAddr, err := memory.ReadPtr(client + offPlantedC4)
	if err == nil && plantedAddr != 0 {
		info.Planted = true
		info.Ticking, _ = memory.ReadBool(plantedAddr + offBombTicking)
		info.BombSite, _ = memory.ReadInt32(plantedAddr + offBombSite)
		info.BeingDefused, _ = memory.ReadBool(plantedAddr + offBeingDefused)
		info.Defused, _ = memory.ReadBool(plantedAddr + offBombDefused)
		info.Exploded, _ = memory.ReadBool(plantedAddr + offHasExploded)

		sceneNode, err := memory.ReadPtr(plantedAddr + offGameSceneNode)
		if err == nil && sceneNode != 0 {
			info.Origin, _ = ReadVector3(memory, sceneNode+offAbsOrigin)
		}

		// Resolve the planter so they can be highlighted
		ownerHandle, err := memory.ReadInt32(plantedAddr + offOwnerEntity)
		if err == nil && ownerHandle != 0 && ownerHandle != -1 {
			info.Carrier = resolveEntitySlot(ownerHandle, pawnHandles)
			if info.Carrier == -1 {
				ownerIdx := int(ownerHandle & 0x7FFF)
				if ownerIdx >= 1 && ownerIdx <= 64 {
					info.Carrier = ownerIdx - 1
				}
			}
		}

		return info
	}

	//----------------------------------------------------------------------------//

	// Find the C4 weapon entity. Prefer the address from the
	// entity list scan. Fall back to dwWeaponC4 which only
	// works when the local player holds the bomb.
	weaponAddr := g.c4Addr
	if weaponAddr == 0 {
		weaponAddr, _ = memory.ReadPtr(client + offWeaponC4)
	}

	if weaponAddr == 0 {
		bombDropped, _ := memory.ReadBool(gameRules + offBombDropped)
		info.Dropped = bombDropped
		return info
	}

	//----------------------------------------------------------------------------//

	ownerHandle, err := memory.ReadInt32(weaponAddr + offOwnerEntity)
	if err == nil && ownerHandle != 0 && ownerHandle != -1 {
		info.Carrier = resolveEntitySlot(ownerHandle, pawnHandles)

		// The owner handle may reference the controller
		// (entity indices 1-64) instead of the pawn
		if info.Carrier == -1 {
			ownerIdx := int(ownerHandle & 0x7FFF)
			if ownerIdx >= 1 && ownerIdx <= 64 {
				info.Carrier = ownerIdx - 1
			}
		}
	} else {
		info.Dropped = true
	}

	// Read bomb position
	sceneNode, err := memory.ReadPtr(weaponAddr + offGameSceneNode)
	if err == nil && sceneNode != 0 {
		info.Origin, _ = ReadVector3(memory, sceneNode+offAbsOrigin)
	}

	return info

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readHostageInfo reads hostage state using entity handles
// from the `C_CSPlayerResource` singleton. The player resource
// is always networked and provides `m_iHostageEntityIDs` which
// can be resolved to entity addresses even when a brute force
// scan of class names misses them.
func (g *Game) readHostageInfo(
	memory *leech.Memory,
	entListBase uintptr,
	pawnHandles [64]int32,
) []HostageInfo {

	//----------------------------------------------------------------------------//

	if g.playerResAddr == 0 {
		if g.hostageFbCount == 0 {
			return nil
		}

		result := make([]HostageInfo, g.hostageFbCount)
		copy(result, g.hostageFallback[:g.hostageFbCount])
		return result
	}

	//----------------------------------------------------------------------------//

	offHostageAlive, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_CSPlayerResource", "fields", "m_bHostageAlive",
	)
	if !ok {
		return nil
	}

	offHostageEntityIds, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_CSPlayerResource", "fields", "m_iHostageEntityIDs",
	)
	if !ok {
		return nil
	}

	offGameSceneNode, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_BaseEntity", "fields", "m_pGameSceneNode",
	)
	if !ok {
		return nil
	}

	offAbsOrigin, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"CGameSceneNode", "fields", "m_vecAbsOrigin",
	)
	if !ok {
		return nil
	}

	offIsRescued, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_Hostage", "fields", "m_isRescued",
	)
	if !ok {
		return nil
	}

	offHostageGrabber, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"C_Hostage", "fields", "m_hHostageGrabber",
	)
	if !ok {
		return nil
	}

	//----------------------------------------------------------------------------//

	// Walk the player resource hostage arrays. CS2 supports up
	// to 12 hostages but typical maps use 2.
	for i := uintptr(0); i < 2; i++ {
		alive, _ := memory.ReadBool(g.playerResAddr + offHostageAlive + i)
		if !alive {
			continue
		}

		entityId, _ := memory.ReadInt32(g.playerResAddr + offHostageEntityIds + i*4)
		if entityId <= 0 {
			continue
		}

		// Resolve entity handle to address
		entityAddr := resolveEntityAddr(memory, entListBase, entityId)
		if entityAddr == 0 {
			continue
		}

		info := HostageInfo{Grabber: -1}

		sceneNode, err := memory.ReadPtr(entityAddr + offGameSceneNode)
		if err != nil || sceneNode == 0 {
			continue
		}

		info.Origin, _ = ReadVector3(memory, sceneNode+offAbsOrigin)
		if info.Origin.IsZero() {
			continue
		}

		info.Rescued, _ = memory.ReadBool(entityAddr + offIsRescued)

		grabberHandle, err := memory.ReadInt32(entityAddr + offHostageGrabber)
		if err == nil && grabberHandle != 0 && grabberHandle != -1 {
			info.Grabber = resolveEntitySlot(grabberHandle, pawnHandles)
		}

		g.updateHostageFallback(info)
	}

	//----------------------------------------------------------------------------//

	if g.hostageFbCount == 0 {
		return nil
	}

	result := make([]HostageInfo, g.hostageFbCount)
	copy(result, g.hostageFallback[:g.hostageFbCount])
	return result

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// updateHostageFallback updates the fallback cache with live
// hostage data. Matches by proximity (within 200 units) to
// avoid duplicates when entity addresses change across rounds.
func (g *Game) updateHostageFallback(info HostageInfo) {

	for i := 0; i < g.hostageFbCount; i++ {
		dx := g.hostageFallback[i].Origin.X - info.Origin.X
		dy := g.hostageFallback[i].Origin.Y - info.Origin.Y
		dz := g.hostageFallback[i].Origin.Z - info.Origin.Z
		if dx*dx+dy*dy+dz*dz < 200*200 {
			g.hostageFallback[i] = info
			return
		}
	}

	if g.hostageFbCount < len(g.hostageFallback) {
		g.hostageFallback[g.hostageFbCount] = info
		g.hostageFbCount++
	}
}

////////////////////////////////////////////////////////////////////////////////

// scanEntities walks the entity list looking for non-player
// entities (C4 weapon, player resource) and caches their
// addresses.
func (g *Game) scanEntities(
	memory *leech.Memory,
	entListBase uintptr,
	scanC4 bool,
	scanHostages bool,
	highestIdx int,
) {

	//----------------------------------------------------------------------------//

	offEntity, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"CEntityInstance", "fields", "m_pEntity",
	)
	if !ok {
		return
	}

	offDesignerName, ok := g.GetOffsetsInt(
		"client_dll", "client.dll", "classes",
		"CEntityIdentity", "fields", "m_designerName",
	)
	if !ok {
		return
	}

	//----------------------------------------------------------------------------//

	foundC4 := !scanC4
	foundPlayerRes := !scanHostages || g.playerResAddr != 0

	scanLimit := highestIdx + 1
	if scanLimit < 512 {
		scanLimit = 512
	}

	// Scan entity list chunks past the player controller
	// slots (indices 1-64). Non-player entities live at higher
	// indices, possibly across multiple chunks of 512 each.
	for idx := 65; idx < scanLimit; idx++ {

		if foundC4 && foundPlayerRes {
			break
		}

		chunkIdx := uintptr(idx) >> 9
		entryIdx := uintptr(idx) & 0x1FF

		chunk, err := memory.ReadPtr(entListBase + 0x10 + 8*chunkIdx)
		if err != nil || chunk == 0 {
			continue
		}

		entityAddr, err := memory.ReadPtr(chunk + 0x70*entryIdx)
		if err != nil || entityAddr == 0 {
			continue
		}

		identity, err := memory.ReadPtr(entityAddr + offEntity)
		if err != nil || identity == 0 {
			continue
		}

		className := g.readClassName(memory, identity)

		if scanC4 && className == c4ClassName {
			g.c4Addr = entityAddr
			foundC4 = true
		}

		// The player resource has class name `C_CSPlayerResource`
		// and designer name `cs_player_manager`. Try both since
		// the class info chain can fail for some entities.
		if !foundPlayerRes {
			if className == playerResClassName {
				g.playerResAddr = entityAddr
				foundPlayerRes = true
			} else if className == "" {
				name := g.readDesignerName(memory, identity, offDesignerName)
				if name == playerResDesignerName {
					g.playerResAddr = entityAddr
					foundPlayerRes = true
				}
			}
		}
	}

	if g.c4Addr != 0 {
		logger.Dbg("found c4 weapon entity")
	}

	if g.playerResAddr != 0 {
		logger.Dbg("found player resource entity")
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readClassName reads the schema class name for an entity
// identity via the class info chain. Returns an empty string
// if any pointer in the chain is invalid.
func (g *Game) readClassName(
	memory *leech.Memory,
	identity uintptr,
) string {

	classInfo, err := memory.ReadPtr(identity + 0x08)
	if err != nil || classInfo == 0 {
		return ""
	}

	binding, err := memory.ReadPtr(classInfo + 0x58)
	if err != nil || binding == 0 {
		return ""
	}

	namePtr, err := memory.ReadPtr(binding + 0x08)
	if err != nil || namePtr == 0 {
		return ""
	}

	name, err := memory.ReadString(namePtr, 48)
	if err != nil {
		return ""
	}

	return name
}

////////////////////////////////////////////////////////////////////////////////

// readDesignerName reads the designer name string from an
// entity identity. Returns an empty string if the pointer is
// null (common for entities like hostages).
func (g *Game) readDesignerName(
	memory *leech.Memory,
	identity uintptr,
	offDesignerName uintptr,
) string {

	namePtr, err := memory.ReadPtr(identity + offDesignerName)
	if err != nil || namePtr == 0 {
		return ""
	}

	name, err := memory.ReadString(namePtr, 48)
	if err != nil {
		return ""
	}

	return name
}

////////////////////////////////////////////////////////////////////////////////

// resolveEntityAddr resolves an entity handle (CEntityIndex)
// to an entity instance address through the entity list.
// Returns 0 when the handle is invalid or the entity slot
// is empty.
func resolveEntityAddr(
	memory *leech.Memory,
	entListBase uintptr,
	handle int32,
) uintptr {

	idx := uintptr(handle & 0x7FFF)
	chunkIdx := idx >> 9
	entryIdx := idx & 0x1FF

	chunk, err := memory.ReadPtr(entListBase + 0x10 + 8*chunkIdx)
	if err != nil || chunk == 0 {
		return 0
	}

	entityAddr, err := memory.ReadPtr(chunk + 0x70*entryIdx)
	if err != nil {
		return 0
	}

	return entityAddr
}

////////////////////////////////////////////////////////////////////////////////

// IsObjectiveCarrier reports whether the entity at the given
// slot is carrying the bomb or a hostage.
func IsObjectiveCarrier(action *ActionState, slot int) bool {

	if action.Bomb != nil && action.Bomb.Carrier == slot {
		return true
	}

	for i := range action.Hostages {
		if action.Hostages[i].Grabber == slot {
			return true
		}
	}

	return false
}

////////////////////////////////////////////////////////////////////////////////

// resolveEntitySlot finds the entity list slot (0-63) whose
// pawn handle matches the given entity handle. Returns -1
// when no match is found.
func resolveEntitySlot(handle int32, pawnHandles [64]int32) int {

	handleIdx := handle & 0x7FFF
	for e := 0; e < 64; e++ {
		if pawnHandles[e] != 0 && (pawnHandles[e]&0x7FFF) == handleIdx {
			return e
		}
	}

	return -1
}
