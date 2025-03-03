package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/config"
	"github.com/dkrutsko/oasis/game"
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

	if cfg.Version {
		// Just print version
		fmt.Printf("%s/%s\n",
			config.AppName(),
			config.Version(),
		)
		return
	}

	// Create application logger
	setupLogger()

	logger.Info(
		"starting",
		logger.String("appName", config.AppName()),
		logger.String("version", config.Version()),
	)

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
		},
	)

	err = g.Create()
	if err != nil {
		panic(err)
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
		panic(err)
	}

	//----------------------------------------------------------------------------//

	// Await shutdown
	err = group.Wait()
	if err != nil {
		panic(err)
	}

	//----------------------------------------------------------------------------//

	logger.Info("shutdown was clean")

	//----------------------------------------------------------------------------//
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
