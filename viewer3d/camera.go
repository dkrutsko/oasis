//go:build viewer3d

package viewer3d

import (
	sysMath "math"
	"time"

	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Distance behind and above the player in auto mode.
	autoDistance = 1200.0
	autoPitch   = 0.35 // ~20 degrees above horizontal

	// Movement speed in free mode (units per second).
	freeMoveSpeed = 800.0
)

////////////////////////////////////////////////////////////////////////////////

// OrbitCamera provides two modes:
//
// Auto mode: orbits behind the local player, tracks their
// position and facing. Scroll to zoom, left-drag to adjust
// viewing angle, double-click to reset.
//
// Free mode: WASD first-person movement through the scene.
// Pressing any movement key exits auto mode. Double-click
// to re-enter auto mode.
type OrbitCamera struct {
	center   math.Vector3
	distance float64
	yaw      float64
	pitch    float64

	fov      float64
	viewport math.Size
	near     float64
	far      float64

	prevX    float32
	prevY    float32
	hasInput bool

	// Accumulated scroll delta from events.
	pendingScroll float64

	// Auto mode state.
	auto          bool
	autoYawOffset float64
	autoPitchOvr  float64
	autoDistOvr   float64
	lastClickTime time.Time
}

////////////////////////////////////////////////////////////////////////////////

func NewOrbitCamera() OrbitCamera {

	return OrbitCamera{
		distance: 5000,
		yaw:      sysMath.Pi / 4,
		pitch:    sysMath.Pi / 6,
		fov:      math.ToRadians(60),
		viewport: math.Size{W: 1280, H: 720},
		near:     1.0,
		far:      200000.0,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (c *OrbitCamera) GetEyePosition() math.Vector3 {

	return math.Vector3{
		X: c.center.X + c.distance*sysMath.Cos(c.pitch)*sysMath.Cos(c.yaw),
		Y: c.center.Y + c.distance*sysMath.Cos(c.pitch)*sysMath.Sin(c.yaw),
		Z: c.center.Z + c.distance*sysMath.Sin(c.pitch),
	}
}

////////////////////////////////////////////////////////////////////////////////

func (c *OrbitCamera) GetViewMatrix() math.Matrix4 {

	return math.Matrix4CreateView(c.GetEyePosition(), c.center, math.Vector3UnitZ)
}

////////////////////////////////////////////////////////////////////////////////

func (c *OrbitCamera) GetProjectionMatrix() math.Matrix4 {

	if c.viewport.W == 0 || c.viewport.H == 0 {
		return math.Matrix4Identity
	}

	return math.Matrix4CreateProjection(c.fov, c.viewport, c.near, c.far)
}

////////////////////////////////////////////////////////////////////////////////

func (c *OrbitCamera) GetViewProjection() math.Matrix4 {

	return c.GetViewMatrix().Mul(c.GetProjectionMatrix())
}

////////////////////////////////////////////////////////////////////////////////

// FrameMap positions the camera to see the entire map
// bounding box at a comfortable zoom level.
func (c *OrbitCamera) FrameMap(min, max math.Vector3) {

	c.center = min.Add(max).MulScalar(0.5)

	size := max.Sub(min)
	maxExtent := size.X
	if size.Y > maxExtent {
		maxExtent = size.Y
	}
	if size.Z > maxExtent {
		maxExtent = size.Z
	}

	c.distance = maxExtent * 1.2
	c.yaw = sysMath.Pi / 4
	c.pitch = sysMath.Pi / 6
	c.auto = true
	c.autoYawOffset = 0
	c.autoPitchOvr = autoPitch
	c.autoDistOvr = 0
}

////////////////////////////////////////////////////////////////////////////////

func (c *OrbitCamera) SetViewport(width, height int) {

	c.viewport = math.Size{W: width, H: height}
}

////////////////////////////////////////////////////////////////////////////////

// HandleScroll accumulates scroll delta from event
// callbacks. Only used in auto mode.
func (c *OrbitCamera) HandleScroll(dy float64) {

	c.pendingScroll += dy
}

////////////////////////////////////////////////////////////////////////////////

// FollowPlayer positions the camera relative to the
// player. Called each frame when auto mode is active.
func (c *OrbitCamera) FollowPlayer(head math.Vector3, yawDeg, pitchDeg float64) {

	if !c.auto {
		return
	}

	c.center = head

	if c.autoDistOvr > 0 {
		c.distance = c.autoDistOvr
	} else {
		c.distance = autoDistance
	}

	playerPitch := pitchDeg * sysMath.Pi / 180.0
	c.pitch = c.autoPitchOvr + playerPitch

	maxPitch := 89.0 * sysMath.Pi / 180.0
	if c.pitch > maxPitch {
		c.pitch = maxPitch
	}
	if c.pitch < -maxPitch {
		c.pitch = -maxPitch
	}

	playerYaw := yawDeg * sysMath.Pi / 180.0
	c.yaw = playerYaw + sysMath.Pi + c.autoYawOffset
}

////////////////////////////////////////////////////////////////////////////////

// UpdateAuto processes input for auto mode: scroll to
// zoom, left-drag to rotate the viewing angle.
func (c *OrbitCamera) UpdateAuto(
	mx, my float32,
	leftDown bool,
) {

	//----------------------------------------------------------------------------//

	if !c.hasInput {
		c.prevX = mx
		c.prevY = my
		c.hasInput = true
	}

	dx := float64(mx - c.prevX)
	dy := float64(my - c.prevY)
	c.prevX = mx
	c.prevY = my

	//----------------------------------------------------------------------------//

	// Scroll to zoom
	if c.pendingScroll != 0 {
		c.distance *= sysMath.Exp(c.pendingScroll * 0.002)
		if c.distance < 1 {
			c.distance = 1
		}
		if c.distance > 200000 {
			c.distance = 200000
		}
		c.autoDistOvr = c.distance
		c.pendingScroll = 0
	}

	//----------------------------------------------------------------------------//

	// Left-drag to adjust viewing angle
	if leftDown && (dx != 0 || dy != 0) {
		c.autoYawOffset -= dx * 0.005
		c.autoPitchOvr += dy * 0.005

		maxPitch := 89.0 * sysMath.Pi / 180.0
		if c.autoPitchOvr > maxPitch {
			c.autoPitchOvr = maxPitch
		}
		if c.autoPitchOvr < -maxPitch {
			c.autoPitchOvr = -maxPitch
		}
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// UpdateFree processes WASD movement and mouse look for
// free mode. Left-click drag to look around (FPS style).
func (c *OrbitCamera) UpdateFree(
	dt float64,
	mx, my float32, leftDown bool,
	forward, backward, left, right, up, down bool,
	sprint bool,
) {

	//----------------------------------------------------------------------------//

	// Mouse look (left-drag)
	if !c.hasInput {
		c.prevX = mx
		c.prevY = my
		c.hasInput = true
	}

	dmx := float64(mx - c.prevX)
	dmy := float64(my - c.prevY)
	c.prevX = mx
	c.prevY = my

	if leftDown && (dmx != 0 || dmy != 0) {
		eye := c.GetEyePosition()

		c.yaw -= dmx * 0.005
		c.pitch += dmy * 0.005

		maxPitch := 89.0 * sysMath.Pi / 180.0
		if c.pitch > maxPitch {
			c.pitch = maxPitch
		}
		if c.pitch < -maxPitch {
			c.pitch = -maxPitch
		}

		// Keep eye fixed, move center
		c.center = math.Vector3{
			X: eye.X - c.distance*sysMath.Cos(c.pitch)*sysMath.Cos(c.yaw),
			Y: eye.Y - c.distance*sysMath.Cos(c.pitch)*sysMath.Sin(c.yaw),
			Z: eye.Z - c.distance*sysMath.Sin(c.pitch),
		}
	}

	//----------------------------------------------------------------------------//

	// WASD movement
	if !forward && !backward && !left && !right && !up && !down {
		return
	}

	speed := freeMoveSpeed * dt
	if sprint {
		speed *= 3.0
	}

	// Forward along the full look direction (including pitch)
	fwd := math.Vector3{
		X: -sysMath.Cos(c.pitch) * sysMath.Cos(c.yaw),
		Y: -sysMath.Cos(c.pitch) * sysMath.Sin(c.yaw),
		Z: -sysMath.Sin(c.pitch),
	}
	rgt := math.Vector3{
		X: -sysMath.Sin(c.yaw),
		Y: sysMath.Cos(c.yaw),
	}

	var move math.Vector3

	if forward {
		move = move.Add(fwd.MulScalar(speed))
	}
	if backward {
		move = move.Add(fwd.MulScalar(-speed))
	}
	if left {
		move = move.Add(rgt.MulScalar(-speed))
	}
	if right {
		move = move.Add(rgt.MulScalar(speed))
	}
	if up {
		move.Z += speed
	}
	if down {
		move.Z -= speed
	}

	c.center = c.center.Add(move)

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// HandleLeftClick detects double-clicks to enter auto
// mode or reset it.
func (c *OrbitCamera) HandleLeftClick() {

	now := time.Now()
	if now.Sub(c.lastClickTime) < 300*time.Millisecond {
		if c.auto {
			c.autoYawOffset = 0
			c.autoPitchOvr = autoPitch
		} else {
			c.auto = true
			c.autoYawOffset = 0
			c.autoPitchOvr = autoPitch
			c.autoDistOvr = 0
		}
		c.lastClickTime = time.Time{}
	} else {
		c.lastClickTime = now
	}
}
