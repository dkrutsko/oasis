package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/config"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/server"
)

////////////////////////////////////////////////////////////////////////////////

func main() {

	//----------------------------------------------------------------------------//

	// Load application config
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Get application config
	cfg := config.GetConfig()

	//----------------------------------------------------------------------------//

	// Print version
	if cfg.Version {
		printVersion()
		return
	}

	//----------------------------------------------------------------------------//

	// Setup the app logger
	setupLogger()

	// Output splash screen
	logSplash()

	//----------------------------------------------------------------------------//

	l := leech.New("-device", "fpga")

	err = l.Create()
	if err != nil {
		logger.Err(
			"failed to initialize leech",
			logger.Error("error", err),
		)
		return
	}

	//----------------------------------------------------------------------------//

	// Enable a graceful shutdown
	ctx, cancel := setupSignals()
	defer cancel()

	// Executing services in the background
	group, gctx := errgroup.WithContext(ctx)

	//----------------------------------------------------------------------------//

	g := game.New(
		&game.Options{
			Group: group,
			Gctx:  gctx,
			Leech: l,
		},
	)

	err = g.Create()
	if err != nil {
		logger.Err(
			"failed to create game",
			logger.Error("error", err),
		)
		return
	}

	//----------------------------------------------------------------------------//

	s := server.New(
		&server.Options{
			Group: group,
			Gctx:  gctx,
			Addr:  cfg.Addr,
			Port:  cfg.Port,
			Game:  g,
		},
	)

	err = s.Create()
	if err != nil {
		logger.Err(
			"failed to create server",
			logger.Error("error", err),
		)
		return
	}

	//----------------------------------------------------------------------------//

	// Await shutdown
	err = group.Wait()
	if err != nil {
		logger.Err(
			"failed during main loop",
			logger.Error("error", err),
		)
		return
	}

	//----------------------------------------------------------------------------//

	// Release handle
	err = l.Close()
	if err != nil {
		logger.Err(
			"failed to close leech",
			logger.Error("error", err),
		)
		return
	}

	//----------------------------------------------------------------------------//

	logger.Info("shutdown was clean")

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func printVersion() {

	// Try to retrieve the version
	version := config.GetVersion()

	// Encode version into the resulting JSON
	result, err := json.MarshalIndent(version, "", "\t")
	if err != nil {
		panic(err)
	}

	// Output result to console
	fmt.Println(string(result))
}

////////////////////////////////////////////////////////////////////////////////

func setupLogger() {

	// Get application config
	cfg := config.GetConfig()

	// If enabling debug
	var level slog.Level
	if cfg.Debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	logger.SetLogger(
		// Setup logger for the app
		logger.New(level, cfg.Json),
	)
}

////////////////////////////////////////////////////////////////////////////////

func logSplash() {

	// Try to retrieve the version
	version := config.GetVersion()

	logger.Info(
		"launching oasis",
		logger.Time("date", version.Date),
		logger.Uint16("build", version.Build),
		logger.Uint16("rev", version.Rev),
	)
}

////////////////////////////////////////////////////////////////////////////////

func setupSignals() (context.Context, context.CancelFunc) {

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		quit := make(chan os.Signal, 2)
		// Offer us an opportunity to gracefully shutdown
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(quit)

		<-quit // Graceful shutdown on first request
		logger.Info("attempting a graceful shutdown")
		cancel()

		<-quit // Termination on subsequent requests
		logger.Warn("forceful termination requested")
		os.Exit(1)
	}()

	return ctx, cancel
}
