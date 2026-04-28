package game

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

type Options struct {
	Group *errgroup.Group
	Gctx  context.Context
	Leech *leech.Leech
}

////////////////////////////////////////////////////////////////////////////////

type Game struct {
	options *Options

	scanner      *ScannerState
	memory       *leech.Memory
	cameraMemory *leech.Memory
	action       *ActionState
	camera       *CameraState

	// Number of DMA goroutines currently abandoned and
	// stuck in a cgo call. Shared across updaters.
	abandoned atomic.Int32

	offsets   map[string][]byte
	strCache  map[string]string
	intCache  map[string]uintptr
	cacheLock sync.Mutex
}

////////////////////////////////////////////////////////////////////////////////

func New(opts *Options) *Game {

	g := &Game{
		options: opts,

		scanner: NewScannerState(),
		action:  NewActionState(),
		camera:  NewCameraState(),

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

		logger.Info("looking for the game")

		for {
			// Trigger a process list refresh before scanning.
			// With -norefresh, MemProcFS does not enumerate
			// processes automatically.
			g.options.Leech.SetConfig(leech.ConfigRefreshFreqMedium, 1)

			// Retrieve the new scanner state
			next := g.updateScanner(g.scanner)

			if next != g.scanner {
				logger.Dbg("scan complete", logger.String("status", next.Result.String()))
			}

			// Whether the state changed
			if next.Changed(g.scanner) {

				// If attached or detached from the game
				if next.Result == ScannerResultSuccess {

					attached = true

					// Create cached memory for entity reads
					mem := next.Process.GetMemory()
					mem.CreateCache(16384, 4096, 5242880, 1048576, 10485760)
					g.memory = mem

					// Separate uncached memory for camera reads
					g.cameraMemory = next.Process.GetMemory()

					logger.Info(
						"attached",
						logger.Uint32("pid", next.Process.GetPid()),
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

		var actionFlight atomic.Bool

		// Rolling average tracking
		var totalDur time.Duration
		var frameCount int
		lastLog := time.Now()

		for {
			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping action updater")
				return nil
			}

			start := time.Now()

			// Skip this frame if the previous DMA call is
			// still stuck. The overlay renders last-known
			// data until DMA recovers.
			if actionFlight.Load() {
				select {
				case <-g.options.Gctx.Done():
					logger.Dbg("stopping action updater")
					return nil
				case <-time.After(time.Second / 120):
				}
				continue
			}

			actionFlight.Store(true)

			done := make(chan *ActionState, 1)
			go func() {
				result := g.updateAction(g.scanner)
				actionFlight.Store(false)
				done <- result
			}()

			timer := time.NewTimer(100 * time.Millisecond)

			select {
			case <-g.options.Gctx.Done():
				timer.Stop()
				logger.Dbg("stopping action updater")
				return nil

			case next := <-done:
				timer.Stop()
				dur := time.Since(start)
				totalDur += dur
				frameCount++

				if next.Result == ActionResultSuccess {
					g.action = next
				}

			case <-timer.C:
				count := g.abandoned.Add(1)
				logger.Warn("action dma timed out",
					logger.Int("abandoned", int(count)),
				)

				// The stuck goroutine clears the flag and
				// decrements when it eventually completes
				go func() {
					<-done
					g.abandoned.Add(-1)
				}()

				if s := g.scanner; s != nil &&
					s.Result == ScannerResultSuccess {

					mem := s.Process.GetMemory()
					mem.CreateCache(
						16384, 4096, 5242880, 1048576, 10485760,
					)
					g.memory = mem
				}
			}

			// Log rolling averages once per second
			if frameCount > 0 && time.Since(lastLog) >= time.Second {
				avg := totalDur / time.Duration(frameCount)
				logger.Dbg("action stats",
					logger.Duration("avg", avg),
					logger.Int("frames", frameCount),
				)
				totalDur = 0
				frameCount = 0
				lastLog = time.Now()
			}

			// Rate limit to target frame rate
			elapsed := time.Since(start)
			rem := (time.Second / 120) - elapsed
			if rem > 0 {
				select {
				case <-g.options.Gctx.Done():
					logger.Dbg("stopping action updater")
					return nil
				case <-time.After(rem):
				}
			}
		}
	})

	//----------------------------------------------------------------------------//

	g.options.Group.Go(func() error {
		logger.Dbg("starting camera updater")

		var cameraFlight atomic.Bool

		var totalDur time.Duration
		var frameCount int
		lastLog := time.Now()

		for {
			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping camera updater")
				return nil
			}

			start := time.Now()

			if cameraFlight.Load() {
				select {
				case <-g.options.Gctx.Done():
					logger.Dbg("stopping camera updater")
					return nil
				case <-time.After(time.Second / 120):
				}
				continue
			}

			cameraFlight.Store(true)

			done := make(chan *CameraState, 1)
			go func() {
				result := g.updateCamera(g.scanner)
				cameraFlight.Store(false)
				done <- result
			}()

			timer := time.NewTimer(50 * time.Millisecond)

			select {
			case <-g.options.Gctx.Done():
				timer.Stop()
				logger.Dbg("stopping camera updater")
				return nil

			case next := <-done:
				timer.Stop()
				dur := time.Since(start)
				totalDur += dur
				frameCount++

				if next.Result == CameraResultSuccess {
					g.camera = next
				}

			case <-timer.C:
				count := g.abandoned.Add(1)
				logger.Warn("camera dma timed out",
					logger.Int("abandoned", int(count)),
				)

				go func() {
					<-done
					g.abandoned.Add(-1)
				}()

				if s := g.scanner; s != nil &&
					s.Result == ScannerResultSuccess {

					g.cameraMemory = s.Process.GetMemory()
				}
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

			// Rate limit to target frame rate
			elapsed := time.Since(start)
			rem := (time.Second / 120) - elapsed
			if rem > 0 {
				select {
				case <-g.options.Gctx.Done():
					logger.Dbg("stopping camera updater")
					return nil
				case <-time.After(rem):
				}
			}
		}
	})

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}
