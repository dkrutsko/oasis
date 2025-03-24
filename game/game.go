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

	scannerLock sync.RWMutex
	scanner     *ScannerState

	offsets   []byte
	strCache  map[string]string
	intCache  map[string]int64
	cacheLock sync.RWMutex
}

////////////////////////////////////////////////////////////////////////////////

func New(opts *Options) *Game {

	g := &Game{
		options: opts,

		offsets:  nil,
		strCache: make(map[string]string),
		intCache: make(map[string]int64),
	}

	// Create empty scanner state
	g.scanner = NewScannerState()
	return g
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetScannerState() *ScannerState {

	// Lock when getting
	g.scannerLock.RLock()
	defer g.scannerLock.RUnlock()

	if g.scanner != nil {
		// Return clone of state
		return g.scanner.Clone()
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) SetScannerState(s *ScannerState) {

	// Lock when setting
	g.scannerLock.Lock()
	defer g.scannerLock.Unlock()

	// Set state
	g.scanner = s
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) Create() error {

	//----------------------------------------------------------------------------//

	var err error
	// Try and read the contents of offsets file
	g.offsets, err = os.ReadFile("./offsets.json")
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	g.options.Group.Go(func() error {
		logger.Dbg("starting game scanner")

		logger.Info("looking for the game")
		for {
			timer := time.NewTimer(1 * time.Second)

			select {
			// Wait for the cancel call
			case <-g.options.Gctx.Done():
				logger.Dbg("stopping game scanner")
				timer.Stop()
				return nil

			// Handle timer signal
			case <-timer.C:
				// Get current scanner state
				prev := g.GetScannerState()

				// Get the new scanner state
				curr := g.updateScanner(prev)

				// If the state changed
				if curr.Changed(prev) {

					// If attached or detached from the game
					if curr.Result == ScannerResultSuccess {
						logger.Info(
							"attached",
							logger.Uint32("pid", curr.PID),
							logger.String("engine", fmt.Sprintf("0x%08X", curr.EngineBase)),
							logger.String("client", fmt.Sprintf("0x%08X", curr.ClientBase)),
						)

					} else {
						logger.Info("detached")
					}

					// Set new scanner state
					g.SetScannerState(curr)
				}
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

			// Get current scanner state
			scanner := g.GetScannerState()

			// Perform the update
			g.updateCamera(scanner)

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
		logger.Dbg("starting content updater")

		for {
			// Check whether stopping the app
			if g.options.Gctx.Err() != nil {
				logger.Dbg("stopping content updater")
				return nil
			}

			// Get a start time
			start := time.Now()

			// Get current scanner state
			scanner := g.GetScannerState()

			// Perform the update
			g.updateContent(scanner)

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
