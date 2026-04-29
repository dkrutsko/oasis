package game

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/maps"
)

////////////////////////////////////////////////////////////////////////////////

type Options struct {
	Group *errgroup.Group
	Gctx  context.Context
	Leech *leech.Leech

	// CameraLeech is an optional secondary FPGA handle used
	// exclusively for camera reads. When nil, camera reads
	// use the primary Leech handle.
	CameraLeech *leech.Leech

	// Target frame rate for action and camera reads (Hz).
	// Zero or negative means uncapped.
	RateAction int
	RateCamera int

	// Directory containing .tri map collision files.
	// When empty, line-of-sight checks are disabled.
	Maps string
}

////////////////////////////////////////////////////////////////////////////////

type Game struct {
	options *Options

	scanner      *ScannerState
	memory       *leech.Memory
	cameraMemory *leech.Memory
	scatter      *leech.Scatter
	action       *ActionState
	camera       *CameraState
	trigger      *Trigger
	health       *HealthMonitor
	currentMap   *maps.Map
	mapName      string

	// Signals the action updater to pause while the
	// scanner is using the primary FPGA.
	scanning atomic.Bool

	// Signaled when action or camera state is updated.
	// The overlay listens on this to render only when
	// new data is available.
	updated chan struct{}

	offsets       map[string][]byte
	strCache      map[string]string
	intCache      map[string]uintptr
	cacheLock     sync.Mutex
	lastEntityLog        time.Time
	lastTlbRefresh       time.Time
	lastCameraTlbRefresh time.Time
	lastMapCheck         time.Time
}

////////////////////////////////////////////////////////////////////////////////

func New(opts *Options) *Game {

	g := &Game{
		options: opts,

		scanner: NewScannerState(),
		action:  NewActionState(),
		camera:  NewCameraState(),
		trigger: NewTrigger(),
		health:  NewHealthMonitor(),

		updated: make(chan struct{}, 1),

		offsets:  make(map[string][]byte),
		strCache: make(map[string]string),
		intCache: make(map[string]uintptr),
	}

	return g
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetScannerState() *ScannerState {
	return g.scanner
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetActionState() *ActionState {
	return g.action
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetCameraState() *CameraState {
	return g.camera
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetTrigger() *Trigger {
	return g.trigger
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetHealthMonitor() *HealthMonitor {
	return g.health
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetCurrentMap() *maps.Map {
	return g.currentMap
}

////////////////////////////////////////////////////////////////////////////////

// Updated returns a channel that receives a signal whenever
// action or camera state is updated. The overlay should
// listen on this to render only when new data is available.
func (g *Game) GetUpdated() <-chan struct{} {
	return g.updated
}

////////////////////////////////////////////////////////////////////////////////

// notifyUpdated sends a non-blocking signal on the updated
// channel. If the overlay hasn't consumed the previous
// signal yet, this is a no-op.
func (g *Game) notifyUpdated() {

	select {
	case g.updated <- struct{}{}:
	default:
	}
}

////////////////////////////////////////////////////////////////////////////////

// checkMapChange reads the current map name from game
// memory and reloads collision data when the map changes.
func (g *Game) checkMapChange(memory *leech.Memory, client uintptr) {

	//----------------------------------------------------------------------------//

	offGlobalVars, ok := g.GetOffsetsInt("offsets", "client.dll", "dwGlobalVars")
	if !ok {
		return
	}

	globalVars, err := memory.ReadPtr(client + offGlobalVars)
	if err != nil || globalVars == 0 {
		return
	}

	// current_map_name is at offset 0x188 in CGlobalVarsBase.
	// It is a pointer to a null-terminated C string containing
	// just the map name (e.g. "de_dust2").
	mapNamePtr, err := memory.ReadPtr(globalVars + 0x188)
	if err != nil || mapNamePtr == 0 {
		return
	}

	mapName, err := memory.ReadString(mapNamePtr, 64)
	if err != nil || mapName == "" {
		return
	}

	//----------------------------------------------------------------------------//

	// No change
	if mapName == g.mapName {
		return
	}

	g.mapName = mapName
	logger.Info("map changed", logger.String("map", mapName))

	// Only load collision data when a maps directory is configured
	if g.options.Maps == "" {
		g.currentMap = nil
		return
	}

	// Try to load collision data for this map
	m, err := maps.Load(g.options.Maps, mapName)
	if err != nil {
		logger.Warn("failed to load map collision data",
			logger.String("map", mapName),
			logger.Error("error", err),
		)
		g.currentMap = nil
		return
	}

	g.currentMap = m

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) Create() error {

	//----------------------------------------------------------------------------//

	{
		// Try and read the contents of offsets file
		offsets, err := os.ReadFile("./offsets.json")
		if err != nil {
			return err
		}

		// Try and read the contents of client_dll file
		client, err := os.ReadFile("./client_dll.json")
		if err != nil {
			return err
		}

		g.offsets["offsets"] = offsets
		g.offsets["client_dll"] = client
	}

	//----------------------------------------------------------------------------//

	// Better logging
	attached := false

	g.options.Group.Go(func() error {
		logger.Dbg("starting game scanner")
		hb := g.health.Register("scanner")

		logger.Info("looking for the game")

		for {
			hb.Beat()

			// Skip the refresh if already attached and the
			// process is still running. The refresh triggers
			// a full process list walk over DMA (~100ms) and
			// blocks the action updater.
			needsScan := !attached ||
				g.scanner == nil ||
				g.scanner.Result != ScannerResultSuccess ||
				g.scanner.Process.HasExited()

			if needsScan {
				// Block the action updater while the scanner
				// uses the primary FPGA.
				g.scanning.Store(true)

				g.options.Leech.SetConfig(leech.ConfigRefreshFreqMedium, 1)
				hb.Beat()
			}

			// Retrieve the new scanner state
			next := g.updateScanner(g.scanner)
			hb.Beat()

			if needsScan {
				g.scanning.Store(false)
			}

			if next != g.scanner {
				logger.Dbg("scan complete", logger.String("status", next.Result.String()))
			}

			// Whether the state changed
			if next.Changed(g.scanner) {

				// If attached or detached from the game
				if next.Result == ScannerResultSuccess {

					attached = true
					pid := next.Process.GetPid()

					// Create cached memory for entity reads
					mem := next.Process.GetMemory()
					mem.CreateCache(16384, 4096, 5242880, 1048576, 10485760)
					g.memory = mem

					// Camera reads use the secondary FPGA when
					// available, otherwise share the primary.
					if g.options.CameraLeech != nil {
						g.options.CameraLeech.SetConfig(
							leech.ConfigRefreshFreqMedium, 1,
						)

						camProc, camErr := g.options.CameraLeech.GetProcess(pid)
						if camErr != nil {
							logger.Warn("failed to get camera process on secondary fpga",
								logger.Error("error", camErr),
							)
							g.cameraMemory = next.Process.GetMemory()
						} else {
							g.cameraMemory = camProc.GetMemory()
							logger.Info("camera using secondary fpga")
						}
					} else {
						g.cameraMemory = next.Process.GetMemory()
					}

					// Create scatter handle for batched entity reads
					scatter, sErr := next.Process.GetScatter(leech.ScatterFlagDefault)
					if sErr != nil {
						logger.Warn("failed to create scatter handle",
							logger.Error("error", sErr),
						)
					}
					g.scatter = scatter

					logger.Info(
						"attached",
						logger.Uint32("pid", pid),
						logger.String("engine", fmt.Sprintf("%08X", next.Engine.GetBase())),
						logger.String("client", fmt.Sprintf("%08X", next.Client.GetBase())),
					)

				} else if attached {

					attached = false

					// Clean up memory instances
					if g.memory != nil {
						g.memory.DeleteCache()
						g.memory = nil
					}
					g.cameraMemory = nil

					if g.scatter != nil {
						g.scatter.Close()
						g.scatter = nil
					}

					logger.Info("detached")
				}

				// Set new state
				g.scanner = next
			}

			// Schedule the next time to do a scan
			timer := time.NewTimer(4 * time.Second)

			select {
			// Wait for the cancel call
			case <-g.options.Gctx.Done():
				logger.Dbg("stopping game scanner")
				timer.Stop()
				return nil

			case <-timer.C:
				// Wait on timer signal
			}
		}
	})

	//----------------------------------------------------------------------------//

	g.options.Group.Go(func() error {
		logger.Dbg("starting action updater")
		hb := g.health.Register("action")

		var totalDur time.Duration
		var frameCount int
		lastLog := time.Now()

		for {
			hb.Beat()

			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping action updater")
				return nil
			}

			start := time.Now()

			// Yield while the scanner is using the primary
			// FPGA to avoid concurrent USB access on the
			// same device.
			if g.scanning.Load() {
				runtime.Gosched()
				continue
			}

			result := g.updateAction(g.scanner)

			dur := time.Since(start)
			totalDur += dur
			frameCount++

			if result.Result == ActionResultSuccess {
				g.action = result
				g.notifyUpdated()
			}

			// Log rolling averages once per second
			if frameCount > 0 && time.Since(lastLog) >= time.Second {
				avg := totalDur / time.Duration(frameCount)
				logger.Dbg("action stats",
					logger.Duration("avg", avg),
					logger.Int("frames", frameCount),
					logger.String("result", result.Result.String()),
				)
				totalDur = 0
				frameCount = 0
				lastLog = time.Now()
			}

			// Rate limit to configured action FPS
			if g.options.RateAction > 0 {
				rem := (time.Second / time.Duration(g.options.RateAction)) - time.Since(start)
				if rem > 0 {
					select {
					case <-g.options.Gctx.Done():
						logger.Dbg("stopping action updater")
						return nil
					case <-time.After(rem):
					}
				}
			}
		}
	})

	//----------------------------------------------------------------------------//

	g.options.Group.Go(func() error {
		logger.Dbg("starting camera updater")
		hb := g.health.Register("camera")

		var totalDur time.Duration
		var frameCount int
		lastLog := time.Now()

		for {
			hb.Beat()

			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping camera updater")
				return nil
			}

			start := time.Now()

			result := g.updateCamera(g.scanner)

			dur := time.Since(start)
			totalDur += dur
			frameCount++

			if result.Result == CameraResultSuccess {
				g.camera = result
				g.notifyUpdated()
			}

			// Log rolling averages once per second
			if frameCount > 0 && time.Since(lastLog) >= time.Second {
				avg := totalDur / time.Duration(frameCount)
				logger.Dbg("camera stats",
					logger.Duration("avg", avg),
					logger.Int("frames", frameCount),
				)
				totalDur = 0
				frameCount = 0
				lastLog = time.Now()
			}

			// Rate limit to configured camera FPS
			if g.options.RateCamera > 0 {
				rem := (time.Second / time.Duration(g.options.RateCamera)) - time.Since(start)
				if rem > 0 {
					select {
					case <-g.options.Gctx.Done():
						logger.Dbg("stopping camera updater")
						return nil
					case <-time.After(rem):
					}
				}
			}
		}
	})

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}
