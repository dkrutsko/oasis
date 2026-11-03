package dump

import (
	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Offset of the sensitivity value inside `dwSensitivity`, hardcoded upstream
const sensitivityOffset = 0x58

////////////////////////////////////////////////////////////////////////////////

// offsetPattern finds one offset, which is the RVA in save slot 1. Upstream
// derives a few more offsets in a callback after a pattern matches, `derive`
// does the same here.
type offsetPattern struct {
	name    string
	pattern string
	derive  func(img *image, offsets map[string]uint32, rva uint32) error
}

////////////////////////////////////////////////////////////////////////////////

// Patterns from `src/analysis/offsets.rs` in cs2-dumper. Keep them identical
// to upstream so that fixes can be copied over when a game update breaks one.
var offsetModules = []struct {
	name     string
	patterns []offsetPattern
}{
	{
		name: "client.dll",
		patterns: []offsetPattern{
			{"dwCSGOInput", "488905${'} 0f57c0 0f1105", deriveViewAngles},
			{"dwEntityList", "48890d${'} e9${} cc", nil},
			{"dwGameEntitySystem", "488b1d${'} 48891d[4] 4c63b3", nil},
			{"dwGameEntitySystem_highestEntityIndex", "ff81u4 4885d2", nil},
			{"dwGameRules", "f6c1010f85${} 4c8b05${'} 4d85", nil},
			{"dwGlobalVars", "488915${'} 488942", nil},
			{"dwGlowManager", "488b05${'} c3 cccccccccccccccc 8b41", nil},
			{"dwLocalPlayerController", "488b05${'} 4189be", nil},
			{"dwPlantedC4", "488b1d${'} 4532f6", nil},
			{"dwPrediction", "488d05${'} c3 cccccccccccccccc 405356 4154", deriveLocalPlayerPawn},
			{"dwSensitivity", "488d0d${[8]'} 0f57c90f28f0", deriveSensitivity},
			{"dwServerConfirmedReservation", "488905${'} e8${} 4885db", nil},
			{"dwViewMatrix", "488d0d${'} 48c1e006", nil},
			{"dwViewRender", "488905${'} 488bc8 4885c0", nil},
			{"dwWeaponC4", "488b15${'} 488b5c24? ffc0 8905${} 488bc6 488934ea 80be", nil},
		},
	},
	{
		name: "engine2.dll",
		patterns: []offsetPattern{
			{"dwBuildNumber", "8905${'} 488d0d${} ff15${} 488b0d", nil},
			{"dwNetworkGameClient", "48893d${'} ff87", nil},
			{"dwNetworkGameClient_clientTickCount", "8b81u4 c3 cccccccccccccccccc 8b81${} c3 cccccccccccccccccc 83b9", nil},
			{"dwNetworkGameClient_deltaTick", "4c8db7u4 4c897c24", nil},
			{"dwNetworkGameClient_isBackgroundMap", "0fb681u4 c3 cccccccccccccccc 0fb681${} c3 cccccccccccccccc 4883ec", nil},
			{"dwNetworkGameClient_localPlayer", "428b94d3u4 5b 49ffe3 32c0 5b c3 cccccccccccccccc 4053", nil},
			{"dwNetworkGameClient_maxClients", "8b81u4 c3????????? 8b81[4] c3????????? 8b81", nil},
			{"dwNetworkGameClient_serverTickCount", "8b81u4 c3 cccccccccccccccccc 83b9", nil},
			{"dwNetworkGameClient_signOnState", "448b81u4 488d0d", nil},
			{"dwWindowHeight", "8b05${'} 8903", nil},
			{"dwWindowWidth", "8b05${'} 8907", nil},
		},
	},
	{
		name: "inputsystem.dll",
		patterns: []offsetPattern{
			{"dwInputSystem", "488905${'} 33c0", nil},
		},
	},
	{
		name: "matchmaking.dll",
		patterns: []offsetPattern{
			{"dwGameTypes", "488d0d${'} ff90", nil},
		},
	},
	{
		name: "soundsystem.dll",
		patterns: []offsetPattern{
			{"dwSoundSystem", "488d0d${'} e8${} 488b0d${} [3] 4c8b82", nil},
			{"dwSoundSystem_engineViewData", "0f1147u1 0f104f10 0f114f7c", nil},
		},
	},
}

////////////////////////////////////////////////////////////////////////////////

func dumpOffsets(
	memory *leech.Memory,
	modules []*leech.Module,
) (map[string]map[string]uint32, error) {

	result := make(map[string]map[string]uint32)

	for _, entry := range offsetModules {

		module := findModule(modules, entry.name)
		if module == nil {
			return nil, errors.New(
				"failed to find module",
				errors.String("name", entry.name),
			)
		}

		img, err := readImage(memory, module)
		if err != nil {
			return nil, errors.New(
				"failed to read module image",
				errors.Error("error", err),
			)
		}

		offsets, err := scanOffsets(img, entry.name, entry.patterns)
		if err != nil {
			return nil, errors.New(
				"failed to scan offsets",
				errors.String("name", entry.name),
				errors.Error("error", err),
			)
		}

		result[entry.name] = offsets
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

func scanOffsets(img *image, module string, patterns []offsetPattern) (map[string]uint32, error) {

	result := make(map[string]uint32)

	for _, p := range patterns {

		save, err := img.find(p.pattern)
		if err != nil {
			return nil, errors.New(
				"failed to find offset",
				errors.String("name", p.name),
				errors.Error("error", err),
			)
		}

		// Upstream skips stale patterns and still writes the file
		if save == nil {
			logger.Warn(
				"pattern is outdated",
				logger.String("module", module),
				logger.String("name", p.name),
			)
			continue
		}

		result[p.name] = save[1]

		if p.derive == nil {
			continue
		}

		err = p.derive(img, result, save[1])
		if err != nil {
			return nil, errors.New(
				"failed to derive offset",
				errors.String("name", p.name),
				errors.Error("error", err),
			)
		}
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

func deriveViewAngles(img *image, offsets map[string]uint32, rva uint32) error {

	save, err := img.find("f2420f108428u4")
	if err != nil {
		return errors.New(
			"failed to find offset",
			errors.String("name", "dwViewAngles"),
			errors.Error("error", err),
		)
	}

	// Upstream skips this offset silently, but a warning makes a stale
	// pattern easier to spot
	if save == nil {
		logger.Warn(
			"pattern is outdated",
			logger.String("name", "dwViewAngles"),
		)
		return nil
	}

	offsets["dwViewAngles"] = rva + save[1]
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func deriveLocalPlayerPawn(img *image, offsets map[string]uint32, rva uint32) error {

	save, err := img.find("4c39b6u4 74? 4488be")
	if err != nil {
		return errors.New(
			"failed to find offset",
			errors.String("name", "dwLocalPlayerPawn"),
			errors.Error("error", err),
		)
	}

	// Upstream skips this offset silently, but a warning makes a stale
	// pattern easier to spot
	if save == nil {
		logger.Warn(
			"pattern is outdated",
			logger.String("name", "dwLocalPlayerPawn"),
		)
		return nil
	}

	offsets["dwLocalPlayerPawn"] = rva + save[1]
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func deriveSensitivity(_ *image, offsets map[string]uint32, _ uint32) error {

	offsets["dwSensitivity_sensitivity"] = sensitivityOffset
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func findModule(modules []*leech.Module, name string) *leech.Module {

	for _, module := range modules {
		if module.GetName() == name {
			return module
		}
	}

	return nil
}
