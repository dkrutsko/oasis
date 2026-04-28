package math

import (
	"fmt"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

var (
	// PointZero represents a point with both components set to zero.
	PointZero = Point{0, 0}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Point represents a 2D position with integer components.
type Point struct {
	X int
	Y int
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the point.
func (p Point) String() string {

	return fmt.Sprintf("[%d, %d]", p.X, p.Y)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether both components are zero.
func (p Point) IsZero() bool {

	return p.X == 0 && p.Y == 0
}

////////////////////////////////////////////////////////////////////////////////

// ToSize returns a `Size` with the point components as dimensions.
func (p Point) ToSize() Size {

	return Size{p.X, p.Y}
}

////////////////////////////////////////////////////////////////////////////////

// ToVector2 returns a `Vector2` with the point components as float64.
func (p Point) ToVector2() Vector2 {

	return Vector2{float64(p.X), float64(p.Y)}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two points.
func (p Point) Add(value Point) Point {

	return Point{p.X + value.X, p.Y + value.Y}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two points.
func (p Point) Sub(value Point) Point {

	return Point{p.X - value.X, p.Y - value.Y}
}

////////////////////////////////////////////////////////////////////////////////

// Neg returns the negation of the point.
func (p Point) Neg() Point {

	return Point{-p.X, -p.Y}
}
