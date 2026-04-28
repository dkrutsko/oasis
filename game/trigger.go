package game

import (
	sysMath "math"
	"sync/atomic"
	"time"

	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Radius used for ray-sphere intersection on head and
	// neck bones. Smaller than body for precision.
	triggerHeadRadius = 4.0
	triggerNeckRadius = 4.5

	// Radius used for ray-sphere intersection on body
	// bones (spine). Larger to be more forgiving.
	triggerBodyRadius = 6.0

	// Minimum time between trigger fires to avoid spamming
	// clicks faster than the game can register.
	triggerCooldown = 80 * time.Millisecond
)

////////////////////////////////////////////////////////////////////////////////

// TriggerResult holds the result of a single trigger
// evaluation frame.
type TriggerResult struct {
	Active bool          // Whether the trigger should fire
	Target *ActionEntity // Entity the crosshair is on
	Bone   int           // Bone index that was hit
	Dist   float64       // Distance along the ray to hit
}

////////////////////////////////////////////////////////////////////////////////

// Trigger evaluates whether the local player's crosshair
// is aimed at an enemy bone and signals when to fire. It
// does not perform input synthesis directly. The caller
// is responsible for reading the result and sending input.
type Trigger struct {
	// Enabled is toggled by the activation key. When false,
	// evaluation is skipped entirely.
	Enabled atomic.Bool

	lastFire time.Time
}

////////////////////////////////////////////////////////////////////////////////

// NewTrigger creates a trigger ready for use.
func NewTrigger() *Trigger {

	return &Trigger{}
}

////////////////////////////////////////////////////////////////////////////////

// Evaluate tests whether the local player's aim direction
// intersects any enemy bone hitbox. Returns a result
// indicating whether to fire and which entity was hit.
func (t *Trigger) Evaluate(action *ActionState) TriggerResult {

	//----------------------------------------------------------------------------//

	if !t.Enabled.Load() {
		return TriggerResult{}
	}

	if action == nil ||
		action.Result != ActionResultSuccess ||
		action.Player == nil {
		return TriggerResult{}
	}

	// Enforce cooldown between fires
	if time.Since(t.lastFire) < triggerCooldown {
		return TriggerResult{}
	}

	//----------------------------------------------------------------------------//

	player := action.Player

	// Build the aim ray from the player's eye position
	// using their view angles.
	eyePos := math.Vector3{
		X: player.Origin.X,
		Y: player.Origin.Y,
		Z: player.Origin.Z + 64.0,
	}

	yawRad := player.Angles.Y * sysMath.Pi / 180.0
	pitchRad := player.Angles.X * sysMath.Pi / 180.0

	direction := math.Vector3{
		X: sysMath.Cos(pitchRad) * sysMath.Cos(yawRad),
		Y: sysMath.Cos(pitchRad) * sysMath.Sin(yawRad),
		Z: -sysMath.Sin(pitchRad),
	}

	ray := geometry.Ray{
		Origin:    eyePos,
		Direction: direction,
	}

	//----------------------------------------------------------------------------//

	// Test each enemy entity
	var best TriggerResult
	best.Dist = sysMath.MaxFloat64

	for i := range action.Entities {
		entity := &action.Entities[i]

		if !entity.Valid || entity == player {
			continue
		}

		// Skip teammates and spectators
		if entity.Team == player.Team || entity.Team <= 1 {
			continue
		}

		if !entity.Bones.Valid {
			continue
		}

		// Test bone hitboxes from most to least valuable
		targets := []struct {
			bone   int
			radius float64
		}{
			{BoneHead, triggerHeadRadius},
			{BoneNeck, triggerNeckRadius},
			{BoneSpine3, triggerBodyRadius},
			{BoneSpine2, triggerBodyRadius},
			{BoneSpine1, triggerBodyRadius},
		}

		for _, tgt := range targets {
			pos := entity.Bones.Pos[tgt.bone]
			if pos.X == 0 && pos.Y == 0 && pos.Z == 0 {
				continue
			}

			sphere := geometry.Sphere{
				Center: pos,
				Radius: tgt.radius,
			}

			dist, hit := ray.IntersectSphere(sphere)
			if hit && dist < best.Dist {
				best.Active = true
				best.Target = entity
				best.Bone = tgt.bone
				best.Dist = dist
			}
		}
	}

	//----------------------------------------------------------------------------//

	if best.Active {
		t.lastFire = time.Now()
	}

	return best

	//----------------------------------------------------------------------------//
}
