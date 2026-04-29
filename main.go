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
	"runtime"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/config"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/input"
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/overlay"
	"github.com/dkrutsko/oasis/viewer3d"
)

////////////////////////////////////////////////////////////////////////////////

func init() {

	// Lock the main goroutine to the OS main thread. macOS
	// requires Cocoa (and therefore GLFW) operations to run
	// on thread 0. This is harmless when --viewer3d is not
	// used.
	runtime.LockOSThread()
}

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

	// Create FPGA leech instances. When --action and --camera
	// point to the same device index, a single leech is shared.
	// When they differ, each gets its own leech on a separate
	// USB bus so reads never contend.
	// LeechCore FPGA device string with devindex parameter
	// for addressing specific FTDI FT601 devices.
	actionDevice := fmt.Sprintf("fpga://devindex=%d", cfg.Action)
	cameraDevice := fmt.Sprintf("fpga://devindex=%d", cfg.Camera)

	l := createLeech(actionDevice, nil)
	logger.Info("action using fpga device", logger.Int("devindex", cfg.Action))

	var l2 *leech.Leech
	if cfg.Action != cfg.Camera {
		// The secondary FPGA may have a limited PCIe memory
		// view and fail to auto-detect the DTB. Pass the DTB
		// discovered by the primary FPGA so it can initialize.
		l2 = createLeech(cameraDevice, l)
		logger.Info("camera using fpga device", logger.Int("devindex", cfg.Camera))
	} else {
		logger.Info("camera using fpga device", logger.Int("devindex", cfg.Action))
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
			Group:       group,
			Gctx:        gctx,
			Leech:       l,
			CameraLeech: l2,
			RateAction:  cfg.RateAction,
			RateCamera:  cfg.RateCamera,
			Maps:        cfg.Maps,
		},
	)

	err := g.Create()
	if err != nil {
		logger.Err(
			"failed to create game",
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeCreateGame))
	}

	//----------------------------------------------------------------------------//

	// Start the health monitor
	health := g.GetHealthMonitor()
	health.Start(group, gctx)

	//----------------------------------------------------------------------------//

	// Start the overlay renderer
	o := overlay.NewOverlay(g, g.GetTrigger())
	defer o.Close()
	o.Start(group, gctx)

	//----------------------------------------------------------------------------//

	// Start the input synthesizer
	inp := input.NewInput()
	defer inp.Close()
	inp.Start(group, gctx)

	// Start the key state reader for hotkey detection
	keys := input.NewKeys()
	defer keys.Close()
	keys.Start(group, gctx)

	//----------------------------------------------------------------------------//

	// Start the triggerbot evaluation loop. The trigger is
	// active only while middle mouse button is held down,
	// detected via the /oasis_keys shared memory segment
	// written by Moonlight.
	trigger := g.GetTrigger()
	triggerHb := health.Register("trigger")

	group.Go(func() error {
		logger.Dbg("starting trigger evaluator")

		for {
			triggerHb.Beat()

			if gctx.Err() != nil {
				logger.Dbg("stopping trigger evaluator")
				return nil
			}

			// Toggle trigger based on middle mouse hold state
			trigger.Enabled.Store(keys.IsMouseDown(input.KeysMouseMiddle))

			action := g.GetActionState()
			result := trigger.Evaluate(action, g.GetCurrentMap())

			if result.Active && inp.IsConnected() {
				logger.Dbg("trigger fired",
					logger.Int("bone", result.Bone),
					logger.Float64("dist", result.Dist),
					logger.Uint32("pending", inp.GetPending()),
				)
				inp.MousePress(input.ButtonLeft)
				time.Sleep(15 * time.Millisecond)
				inp.MouseRelease(input.ButtonLeft)
			}

			// Wait for new game data before evaluating again
			select {
			case <-gctx.Done():
				logger.Dbg("stopping trigger evaluator")
				return nil
			case <-g.GetUpdated():
			}
		}
	})

	//----------------------------------------------------------------------------//

	exitCode := exitCodeSuccess

	// When the 3D viewer is active, run its GLFW loop on the
	// main goroutine (required by macOS) and wait for the
	// errgroup in the background. Otherwise use the normal
	// blocking wait.
	if cfg.Viewer3d {
		v3d := viewer3d.NewViewer3d(g, keys)

		errCh := make(chan error, 1)
		go func() { errCh <- group.Wait() }()

		if vErr := v3d.Run(gctx); vErr != nil {
			logger.Err("failed to run 3d viewer",
				logger.Error("error", vErr),
			)
		}
		v3d.Close()

		cancel()
		err = <-errCh
	} else {
		err = group.Wait()
	}

	if err != nil {
		exitCode = exitCodeDaemonError

		logger.Err(
			"failed during main loop",
			logger.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	logger.Info("performing shutdown")

	// Release handles
	err = l.Close()
	if err != nil {
		exitCode = exitCodeCloseLeech

		logger.Err(
			"failed to close leechcore",
			logger.Error("error", err),
		)
	}

	if l2 != nil {
		if err := l2.Close(); err != nil {
			exitCode = exitCodeCloseLeech

			logger.Err(
				"failed to close secondary leechcore",
				logger.Error("error", err),
			)
		}
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

func createLeech(device string, primary *leech.Leech) *leech.Leech {

	args := []string{
		"-device", device,

		// Disable all background refresh threads. We
		// manage our own read cache and process scanner,
		// so automatic TLB walks and process enumeration
		// just cause DMA stalls. Refreshes are triggered
		// manually when the scanner needs them.
		"-norefresh",
	}

	// When a primary leech is provided, forward its kernel
	// DTB so the secondary can skip the auto-detection scan
	// which may fail if the FPGA has a limited memory view.
	if primary != nil {
		dtb, err := primary.GetProcessDtb(4)
		if err != nil {
			logger.Warn("failed to read system dtb from primary",
				logger.Error("error", err),
			)
		} else {
			args = append(args, "-dtb", fmt.Sprintf("0x%x", dtb))
			logger.Info("forwarding dtb to secondary fpga",
				logger.String("dtb", fmt.Sprintf("0x%x", dtb)),
			)
		}
	}

	l := leech.New(&leech.Options{
		Args: args,
	})

	err := l.Create()
	if err != nil {
		logger.Err(
			"failed to create leechcore",
			logger.String("device", device),
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
		{"paging_enabled", leech.ConfigPagingEnabled, 0},
		{"fpga_retry_on_error", leech.ConfigFpgaRetryOnError, 1},
		{"fpga_delay_read", leech.ConfigFpgaDelayRead, 200},
	} {
		if err := l.SetConfig(opt.option, opt.value); err != nil {
			logger.Warn("failed to set leech config",
				logger.String("name", opt.name),
				logger.String("device", device),
				logger.Error("error", err),
			)
		}
	}

	return l
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
