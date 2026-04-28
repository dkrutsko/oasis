package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/config"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/overlay"
)

////////////////////////////////////////////////////////////////////////////////

type exitCodeType int

const (
	exitCodeSuccess exitCodeType = iota
	exitCodeLoadConfig
	exitCodeCreateLeech
	exitCodeForceExit
	exitCodeCreateGame
	exitCodeDaemonError
	exitCodeCloseLeech
	exitCodeCreateOverlay
	exitCodeViewerError
)

////////////////////////////////////////////////////////////////////////////////

func main() {

	//----------------------------------------------------------------------------//

	// Try loading application config
	cfg, cfgErr := config.LoadConfig()
	if cfg == nil {
		// This should not be possible
		// but just in case do a check
		panic(cfgErr)
	}

	{
		opts := logger.NewOptions()

		if cfg.Debug {
			opts.Level = slog.LevelDebug
		} else {
			opts.Level = slog.LevelInfo
		}

		opts.Json = cfg.Json
		logger.SetLogger("", logger.New(opts))
	}

	if cfg.Version {
		// Attempt to retrieve version
		version := config.GetVersion()

		if cfg.Json {
			// Try encoding the version struct
			data, err := json.Marshal(version)
			if err != nil {
				// Panic here is fine because
				// failure here is quite rare
				panic(err)
			}

			// Output JSON structure
			fmt.Println(string(data))

		} else {
			// Output using string format
			fmt.Println(version.String())
		}

		return
	}

	// Delay handling LoadConfig error
	// until after the logger setup is
	// complete and version is handled
	if cfgErr != nil {
		logger.Err(
			"failed to load app config",
			logger.Error("error", cfgErr),
		)
		os.Exit(int(exitCodeLoadConfig))
	}

	//----------------------------------------------------------------------------//

	// Attempt to get splash screen
	splash := config.GetSplash(cfg)

	if cfg.Json {
		logger.Info(
			"launching "+config.AppName,
			splash.SlogAttrs()...,
		)

	} else {
		// Output the result with colors
		fmt.Println(splash.Format(true))
	}

	//----------------------------------------------------------------------------//

	// Start pprof server if requested
	if cfg.Pprof {
		go func() {
			logger.Info("pprof server listening on localhost:6060")
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	//----------------------------------------------------------------------------//

	// Run in viewer mode if requested
	if cfg.Viewer {
		v := overlay.NewViewer()

		err := v.Run()
		v.Close()

		if err != nil {
			logger.Err(
				"failed to run viewer",
				logger.Error("error", err),
			)
			os.Exit(int(exitCodeViewerError))
		}

		return
	}

	//----------------------------------------------------------------------------//

	l := leech.New(&leech.Options{
		Args: []string{
			"-device", "fpga",

			// Disable all background refresh threads. We
			// manage our own read cache and process scanner,
			// so automatic TLB walks and process enumeration
			// just cause DMA stalls. Refreshes are triggered
			// manually when the scanner needs them.
			"-norefresh",
		},
	})

	err := l.Create()
	if err != nil {
		logger.Err(
			"failed to create leechcore",
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeCreateLeech))
	}

	// Tune FPGA device and VMM settings for low-latency
	// repeated reads of the same process.
	for _, opt := range []struct {
		name   string
		option uint64
		value  uint64
	}{
		// Disable paging - we don't need page fault resolution
		{"paging_enabled", leech.ConfigPagingEnabled, 0},

		// Disable read retries at the device level. We handle
		// failures with ZEROPAD_ON_FAIL, so retries just add
		// latency on transient errors.
		{"fpga_retry_on_error", leech.ConfigFpgaRetryOnError, 0},

		// Lower the FPGA read delay from the default 300-400us.
		// This controls how long LeechCore waits between sending
		// a TLP and reading the USB response buffer.
		{"fpga_delay_read", leech.ConfigFpgaDelayRead, 150},
	} {
		if err := l.SetConfig(opt.option, opt.value); err != nil {
			logger.Warn("failed to set leech config",
				logger.String("name", opt.name),
				logger.Error("error", err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	// Enable graceful shutdowns
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
		os.Exit(int(exitCodeCreateGame))
	}

	//----------------------------------------------------------------------------//

	// Start the overlay renderer
	o := overlay.NewOverlay(g)
	defer o.Close()
	o.Start(group, gctx)

	//----------------------------------------------------------------------------//

	exitCode := exitCodeSuccess

	// Await shutdown
	err = group.Wait()
	if err != nil {
		exitCode = exitCodeDaemonError

		logger.Err(
			"failed during main loop",
			logger.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	logger.Info("performing shutdown")

	// Release handle
	err = l.Close()
	if err != nil {
		exitCode = exitCodeCloseLeech

		logger.Err(
			"failed to close leechcore",
			logger.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Check if specifying exit code
	if exitCode != exitCodeSuccess {
		os.Exit(int(exitCode))
	}

	logger.Info("shutdown was clean")

	//----------------------------------------------------------------------------//
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

		// Force exit if goroutines don't stop within 2 seconds.
		// l.Close takes an exclusive lock that blocks if an
		// abandoned DMA goroutine still holds a read lock.
		go func() {
			time.Sleep(2 * time.Second)
			logger.Warn("graceful shutdown timed out")
			os.Exit(int(exitCodeForceExit))
		}()

		<-quit // Termination on subsequent requests
		logger.Warn("forceful termination requested")
		os.Exit(int(exitCodeForceExit))
	}()

	return ctx, cancel
}
