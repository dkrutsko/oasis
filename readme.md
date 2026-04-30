### Warning: AI Slop Ahead. Use at Your Own Risk!

# Oasis

Oasis is a real-time game assistance tool that uses FPGA-based DMA (Direct Memory Access) to read game memory, analyze player positions and skeletal data, and provide intelligent triggerbot and overlay functionality. It integrates hardware-level memory reading with 3D spatial geometry calculations and input synthesis via shared memory.

## Features

- **DMA Memory Reading** - Direct hardware access via FPGA (FT601 USB) to read process memory without kernel hooks
- **Real-time Entity Tracking** - Reads and caches player positions, view angles, and bone data at configurable frame rates
- **Intelligent Triggerbot** - Ray-cast intersection with head and body hitboxes, with configurable cooldowns
- **Line-of-Sight Verification** - Uses map collision geometry (.tri files) with BVH spatial acceleration to prevent firing through walls
- **Dual FPGA Support** - Separate devices for action and camera reads to avoid USB contention
- **Overlay Rendering** - Outputs entity data to shared memory for use with Moonlight streaming
- **3D Map Viewer** - GPU-accelerated viewer using WebGPU (pure Go, zero CGO) for visualizing collision geometry and entities
- **Cross-Platform** - Targets macOS (Intel + ARM + Universal), Linux (x86_64/ARM64), and Windows (x86_64)
- **Structured Logging** - JSON or colored text output with debug mode
- **Health Monitoring** - Goroutine heartbeat detection for service health
- **Built-in Profiling** - Optional pprof server on localhost:6060

## Requirements

- Go 1.26+
- FPGA hardware (FT601-based) with LeechCore/VMMDLL drivers
- Make (for build commands)

## Getting Started

## Building

- Use this [modified version](https://github.com/dkrutsko/moonlight-qt) of Moonlight streaming software to stream the game.
- Download memory offsets using `./dump.sh` or by downloading `offsets.json` and `client_dll.json` from [this repository](https://github.com/a2x/cs2-dumper).
- Download [pcieleech](https://github.com/ufrisk/pcileech/releases/tag/v4.19) for the operating system you're planning on running this software.
- Use [CS2-Phys-Extractor](https://github.com/itzlaith/cs2-phys-extractor) to extract and convert maps into `.tri` files for `--maps` to work.
- Run `./patch.sh` on MacOS to fix a bug with `gogpu`, optional but you'll get an annoying key press sound if you don't.

```
# Run from source code
cd `./bin/`
CGO_ENABLED=0 go run -mod=vendor -tags "viewer3d nofakecgo" .. --camera 0 --action 0 --viewer3d --maps "/path/to/tri"
```

```
# Show available commands
make help

# Build release binary
make build

# Build debug binary (with race detection and GDB support)
make debug

# Clean build artifacts
make clean

# Cross-compile for all platforms
make publish
```

The release binary is output to `./bin/oasis`.

## Usage

```
oasis [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--version` | | Print version info and exit |
| `--debug` | | Enable verbose debug logging |
| `--json` | | Enable JSON-formatted output |
| `--viewer` | | Launch the debug overlay viewer |
| `--viewer3d` | | Launch the GPU-accelerated 3D map viewer |
| `--pprof` | | Start pprof server on localhost:6060 |
| `--action` | `0` | FPGA device index for action/entity reads |
| `--camera` | `0` | FPGA device index for camera reads |
| `--rate` | | Set both action and camera frame rates (Hz) |
| `--rate-action` | `60` | Target frame rate for action reads (Hz) |
| `--rate-camera` | `90` | Target frame rate for camera reads (Hz) |
| `--maps` | | Directory containing .tri map collision files |

### Examples

```bash
# Run with default settings (single FPGA)
oasis

# Run with dual FPGAs and map collision data
oasis --action 0 --camera 1 --maps ./maps

# Run with higher frame rates
oasis --rate 120

# Launch the 3D map viewer for debugging
oasis --viewer3d --maps ./maps

# Run with JSON logging and debug output
oasis --debug --json
```

## Project Structure

```
oasis/
	main.go          Entry point and service orchestration
	config/          CLI argument parsing, versioning, splash screen
	game/            Core game state, entity tracking, triggerbot, health monitor
	geometry/        3D collision primitives (Ray, Box, Sphere, Capsule, Cylinder)
	input/           Input synthesis via shared memory (keyboard/mouse)
	leech/           FPGA/DMA memory access wrapper (LeechCore/VMMDLL)
	logger/          Structured JSON/text logging
	maps/            Map collision geometry parsing (.tri files) and BVH acceleration
	math/            Vector/matrix math library (Vector2/3/4, Matrix3/4, Quaternion)
	overlay/         Overlay renderer (shared memory output for Moonlight)
	shm/             Cross-platform shared memory (SysV on Unix, Windows API)
	viewer3d/        GPU-accelerated 3D map viewer with FPS camera
	errors/          Custom structured error handling
	utility/         Git metadata embedding
```

## Architecture

The application runs as a set of concurrent services managed by an `errgroup`:

1. **Game** - Reads entity state and camera data from game memory via DMA
2. **Overlay** - Renders entity positions to shared memory for the streaming overlay
3. **Input** - Listens on shared memory for input synthesis commands
4. **Keys** - Reads key/mouse state for hotkey detection (trigger activation)
5. **Trigger** - Evaluates crosshair-to-hitbox intersection and fires when conditions are met
6. **Health Monitor** - Tracks goroutine heartbeats and detects stalls
7. **Viewer3d** (optional) - GPU-accelerated 3D visualization on the main thread

Graceful shutdown is handled via SIGINT/SIGTERM with a 2-second timeout before force exit.
