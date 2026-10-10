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

const hostageDesignerName = "hostage_entity"
const hostageScanInterval = 5 * time.Second

////////////////////////////////////////////////////////////////////////////////

// readObjectives reads bomb and hostage state from game
// memory. Called after the entity loop to populate the
// objective fields of the action state.
func (g *Game) readObjectives(
	memory *leech.Memory,
	client uintptr,
	listEntry uintptr,
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

	//----------------------------------------------------------------------------//

	var bomb *BombInfo
	hasBombTarget, err := memory.ReadBool(gameRules + offMapHasBombTarget)
	if err == nil && hasBombTarget {
		bomb = g.readBombInfo(memory, client, pawnHandles)
	}

	var hostages []HostageInfo
	hasRescueZone, err := memory.ReadBool(gameRules + offMapHasRescueZone)
	if err == nil && hasRescueZone {
		hostages = g.readHostageInfo(memory, listEntry, pawnHandles)
	}

	return bomb, hostages

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readBombInfo reads the C4 weapon and planted bomb state.
func (g *Game) readBombInfo(
	memory *leech.Memory,
	client uintptr,
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

	// Read weapon C4 for carrier or dropped state
	weaponAddr, err := memory.ReadPtr(client + offWeaponC4)
	if err != nil || weaponAddr == 0 {
		return info
	}

	ownerHandle, err := memory.ReadInt32(weaponAddr + offOwnerEntity)
	if err == nil && ownerHandle != 0 && ownerHandle != -1 {
		info.Carrier = resolveEntitySlot(ownerHandle, pawnHandles)
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

// readHostageInfo reads the state of cached hostage entities
// and periodically rescans the entity list to find them.
func (g *Game) readHostageInfo(
	memory *leech.Memory,
	listEntry uintptr,
	pawnHandles [64]int32,
) []HostageInfo {

	//----------------------------------------------------------------------------//

	// Rescan the entity list for hostage entities periodically
	now := time.Now()
	if now.Sub(g.lastHostageScan) >= hostageScanInterval {
		g.lastHostageScan = now
		g.scanHostages(memory, listEntry)
	}

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

// scanHostages walks the entity list looking for hostage
// entities and caches their addresses. CS2 hostage maps
// have at most two hostages.
func (g *Game) scanHostages(memory *leech.Memory, listEntry uintptr) {

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

	g.hostageCount = 0

	// Scan chunk 0 of the entity list past the player slots.
	// Hostages are among the first non-player entities and
	// are always within the first chunk (indices 0-511).
	for idx := 65; idx < 512; idx++ {

		entityAddr, err := memory.ReadPtr(listEntry + uintptr(idx)*0x70)
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

		if name == hostageDesignerName {
			g.hostageAddrs[g.hostageCount] = entityAddr
			g.hostageCount++
			if g.hostageCount >= len(g.hostageAddrs) {
				break
			}
		}
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
