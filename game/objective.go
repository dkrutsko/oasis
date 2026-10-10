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
	c4DesignerName      = "weapon_c4"
	hostageDesignerName = "hostage_entity"
	entityScanInterval  = 5 * time.Second
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
	// (C4 weapon, hostages). The scan is throttled so the
	// per-frame cost is just the cached address reads.
	now := time.Now()
	if now.Sub(g.lastEntityScan) >= entityScanInterval {
		g.lastEntityScan = now
		g.scanEntities(memory, entListBase, hasBombTarget, hasRescueZone)
	}

	//----------------------------------------------------------------------------//

	var bomb *BombInfo
	if hasBombTarget {
		bomb = g.readBombInfo(memory, client, gameRules, pawnHandles)
	}

	var hostages []HostageInfo
	if hasRescueZone {
		hostages = g.readHostageInfo(memory, pawnHandles)
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

// readHostageInfo reads the state of cached hostage entities.
// The entity scan that populates the cache runs from
// `readObjectives` before this is called.
func (g *Game) readHostageInfo(
	memory *leech.Memory,
	pawnHandles [64]int32,
) []HostageInfo {

	//----------------------------------------------------------------------------//

	if g.hostageCount == 0 {
		return nil
	}

	//----------------------------------------------------------------------------//

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

	hostages := make([]HostageInfo, 0, g.hostageCount)

	for i := 0; i < g.hostageCount; i++ {
		addr := g.hostageAddrs[i]
		info := HostageInfo{Grabber: -1}

		// Read position through the scene node
		sceneNode, err := memory.ReadPtr(addr + offGameSceneNode)
		if err != nil || sceneNode == 0 {
			continue
		}

		info.Origin, _ = ReadVector3(memory, sceneNode+offAbsOrigin)
		info.Rescued, _ = memory.ReadBool(addr + offIsRescued)

		// Check if a player is carrying this hostage
		grabberHandle, err := memory.ReadInt32(addr + offHostageGrabber)
		if err == nil && grabberHandle != 0 && grabberHandle != -1 {
			info.Grabber = resolveEntitySlot(grabberHandle, pawnHandles)
		}

		hostages = append(hostages, info)
	}

	return hostages

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// scanEntities walks the entity list looking for non-player
// entities (C4 weapon, hostages) and caches their addresses.
func (g *Game) scanEntities(
	memory *leech.Memory,
	entListBase uintptr,
	scanC4 bool,
	scanHostages bool,
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

	g.c4Addr = 0
	g.hostageCount = 0

	foundC4 := !scanC4
	foundHostages := !scanHostages

	// Scan entity list chunks past the player controller
	// slots (indices 1-64). Non-player entities such as the
	// C4 weapon and hostages live at higher indices, possibly
	// across multiple chunks of 512 entries each.
	for idx := 65; idx < 2048; idx++ {

		if foundC4 && foundHostages {
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

		namePtr, err := memory.ReadPtr(identity + offDesignerName)
		if err != nil || namePtr == 0 {
			continue
		}

		name, err := memory.ReadString(namePtr, 32)
		if err != nil {
			continue
		}

		if scanC4 && name == c4DesignerName {
			g.c4Addr = entityAddr
			foundC4 = true
		}

		if scanHostages && name == hostageDesignerName {
			g.hostageAddrs[g.hostageCount] = entityAddr
			g.hostageCount++
			if g.hostageCount >= len(g.hostageAddrs) {
				foundHostages = true
			}
		}
	}

	if g.c4Addr != 0 {
		logger.Dbg("found c4 weapon entity")
	}

	if g.hostageCount > 0 {
		logger.Dbg("found hostage entities",
			logger.Int("count", g.hostageCount),
		)
	}

	//----------------------------------------------------------------------------//
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
