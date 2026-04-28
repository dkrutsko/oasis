//go:build darwin || linux

package input

////////////////////////////////////////////////////////////////////////////////

// Button represents a mouse button. The values match
// Robot's Button enumeration (sequential 0-4). These are
// mapped to Moonlight's bitfield format internally.
type Button uint8

////////////////////////////////////////////////////////////////////////////////

const (
	ButtonLeft   Button = 0
	ButtonMid    Button = 1
	ButtonMiddle Button = 1
	ButtonRight  Button = 2
	ButtonX1     Button = 3
	ButtonX2     Button = 4
)
