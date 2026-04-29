package game

import (
	sysMath "math"
	"sync/atomic"
	"time"

	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/maps"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Radius for the head sphere.
	TriggerHeadRadius = 5.5

	// Radius for the body capsule that extends from neck
	// to pelvis.
	TriggerBodyRadius = 8.0

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
func (t *Trigger) Evaluate(action *ActionState, m *maps.Map) TriggerResult {

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
	// using their view angles. Use the actual head bone
	// when available for accuracy.
	var eyePos math.Vector3
	if player.Bones.Valid && !player.Bones.Pos[BoneHead].IsZero() {
		eyePos = player.Bones.Pos[BoneHead]
	} else {
		eyePos = math.Vector3{
			X: player.Origin.X,
			Y: player.Origin.Y,
			Z: player.Origin.Z + 64.0,
		}
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

		headPos := entity.Bones.Pos[BoneHead]
		neckPos := entity.Bones.Pos[BoneNeck]
		pelvisPos := entity.Bones.Pos[BonePelvis]

		// Head sphere
		if !headPos.IsZero() {
			dist, hit := ray.IntersectSphere(geometry.Sphere{
				Center: headPos,
				Radius: TriggerHeadRadius,
			})
			if hit && dist < best.Dist {
				best.Active = true
				best.Target = entity
				best.Bone = BoneHead
				best.Dist = dist
			}
		}

		// Body capsule from neck to pelvis
		if !neckPos.IsZero() && !pelvisPos.IsZero() {
			dist, hit := ray.IntersectCapsule(geometry.Capsule{
				Start:  neckPos,
				End:    pelvisPos,
				Radius: TriggerBodyRadius,
			})
			if hit && dist < best.Dist {
				best.Active = true
				best.Target = entity
				best.Bone = BoneNeck
				best.Dist = dist
			}
		}
	}

	//----------------------------------------------------------------------------//

	// Verify line of sight through map geometry. If a
	// wall blocks the path to the hit point, suppress
	// the trigger to avoid firing through walls.
	if best.Active && m != nil {
		hitPoint := ray.GetPoint(best.Dist)
		if !m.IsVisible(eyePos, hitPoint) {
			best.Active = false
			best.Target = nil
		}
	}

	//----------------------------------------------------------------------------//

	if best.Active {
		t.lastFire = time.Now()
	}

	return best

	//----------------------------------------------------------------------------//
}
