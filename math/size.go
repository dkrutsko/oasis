package math

import (
	"fmt"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

var (
	// SizeZero represents a size with both dimensions set to zero.
	SizeZero = Size{0, 0}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Size represents 2D dimensions with integer components.
type Size struct {
	W int
	H int
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the size.
func (s Size) String() string {

	return fmt.Sprintf("[%d, %d]", s.W, s.H)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether both dimensions are zero.
func (s Size) IsZero() bool {

	return s.W == 0 && s.H == 0
}

////////////////////////////////////////////////////////////////////////////////

// IsEmpty returns whether either dimension is less than or equal to zero.
func (s Size) IsEmpty() bool {

	return s.W <= 0 || s.H <= 0
}

////////////////////////////////////////////////////////////////////////////////

// ToPoint returns a `Point` with the dimensions as components.
func (s Size) ToPoint() Point {

	return Point{s.W, s.H}
}

////////////////////////////////////////////////////////////////////////////////

// ToVector2 returns a `Vector2` with the dimensions as float64.
func (s Size) ToVector2() Vector2 {

	return Vector2{float64(s.W), float64(s.H)}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two sizes.
func (s Size) Add(value Size) Size {

	return Size{s.W + value.W, s.H + value.H}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two sizes.
func (s Size) Sub(value Size) Size {

	return Size{s.W - value.W, s.H - value.H}
}
