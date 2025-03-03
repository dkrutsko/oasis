package game

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

type Options struct {
	Group *errgroup.Group
	Gctx  context.Context
}

////////////////////////////////////////////////////////////////////////////////

type Game struct {
	options *Options
}

////////////////////////////////////////////////////////////////////////////////

func New(opts *Options) *Game {

	// Create game
	return &Game{
		options: opts,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) Create() error {

	//----------------------------------------------------------------------------//

	opts := g.options

	//----------------------------------------------------------------------------//

	opts.Group.Go(func() error {
		logger.Dbg("starting scanner")

		for {
			timer := time.NewTimer(1 * time.Second)

			select {
			// Wait for cancel call
			case <-opts.Gctx.Done():
				logger.Dbg("stopping scanner")
				timer.Stop()
				return nil

			// Handle timer signal
			case <-timer.C:
				// Perform update
				g.updateScanner()
			}
		}
	})

	//----------------------------------------------------------------------------//

	opts.Group.Go(func() error {
		logger.Dbg("starting camera updater")

		for {
			// Check if stopping the app
			if opts.Gctx.Err() != nil {
				logger.Dbg("stopping camera updater")
				return nil
			}

			// Get a start time
			start := time.Now()

			// Perform update
			g.updateCamera()

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

	opts.Group.Go(func() error {
		logger.Dbg("starting content updater")

		for {
			// Check if stopping the app
			if opts.Gctx.Err() != nil {
				logger.Dbg("stopping content updater")
				return nil
			}

			// Get a start time
			start := time.Now()

			// Perform update
			g.updateContent()

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
