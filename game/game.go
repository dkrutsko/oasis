package game

import (
	"context"
	"fmt"
	"os"
	"sync"
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

	scanner *ScannerState
	action  *ActionState
	camera  *CameraState

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
					logger.Info(
						"attached",
						logger.Uint32("pid", next.Process.GetPid()),
						logger.String("engine", fmt.Sprintf("%08X", next.Engine.GetBase())),
						logger.String("client", fmt.Sprintf("%08X", next.Client.GetBase())),
					)

				} else if attached {

					attached = false
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

		for {
			// Check whether stopping the app
			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping action updater")
				return nil
			}

			// Get a start time
			start := time.Now()

			// Retrieve the new action state
			next := g.updateAction(g.scanner)

			// Set state if update was successful
			if next.Result == ActionResultSuccess {
				g.action = next
			}

			// Calculate time for update
			elapsed := time.Since(start)

			// Calculate remaining time for frame
			rem := (time.Second / 120) - elapsed
			if rem > 0 {
				time.Sleep(rem)
			}
		}
	})

	//----------------------------------------------------------------------------//

	g.options.Group.Go(func() error {
		logger.Dbg("starting camera updater")

		for {
			// Check whether stopping the app
			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping camera updater")
				return nil
			}

			// Get a start time
			start := time.Now()

			// Retrieve the new camera state
			next := g.updateCamera(g.scanner)

			// Set state if update was successful
			if next.Result == CameraResultSuccess {
				g.camera = next
				//fmt.Println(g.camera.View.String()) // TODO: REMOVE THIS LINE
			}

			// Calculate time for update
			elapsed := time.Since(start)

			// Calculate remaining time for frame
			rem := (time.Second / 120) - elapsed
			if rem > 0 {
				time.Sleep(rem)
			}
		}
	})

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}
