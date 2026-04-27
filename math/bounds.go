package math

import (
	"fmt"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

var (
	// BoundsZero represents a bounds with all components set to zero.
	BoundsZero = Bounds{0, 0, 0, 0}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Bounds represents a 2D axis-aligned rectangle defined by a position and
// dimensions with integer components.
type Bounds struct {
	X int
	Y int
	W int
	H int
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the bounds.
func (bn Bounds) String() string {

	return fmt.Sprintf("[%d, %d, %d, %d]", bn.X, bn.Y, bn.W, bn.H)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all components are zero.
func (bn Bounds) IsZero() bool {

	return bn.X == 0 && bn.Y == 0 && bn.W == 0 && bn.H == 0
}

////////////////////////////////////////////////////////////////////////////////

// IsEmpty returns whether either dimension is less than or equal to zero.
func (bn Bounds) IsEmpty() bool {

	return bn.W <= 0 || bn.H <= 0
}

////////////////////////////////////////////////////////////////////////////////

// IsValid returns whether both dimensions are greater than zero.
func (bn Bounds) IsValid() bool {

	return bn.W > 0 && bn.H > 0
}

////////////////////////////////////////////////////////////////////////////////

// GetLeft returns the left edge of the bounds.
func (bn Bounds) GetLeft() int {

	return bn.X
}

////////////////////////////////////////////////////////////////////////////////

// SetLeft returns the bounds with the left edge moved to the given value.
// The right edge stays fixed, so the width adjusts.
func (bn Bounds) SetLeft(value int) Bounds {

	return Bounds{value, bn.Y, bn.W + (bn.X - value), bn.H}
}

////////////////////////////////////////////////////////////////////////////////

// GetTop returns the top edge of the bounds.
func (bn Bounds) GetTop() int {

	return bn.Y
}

////////////////////////////////////////////////////////////////////////////////

// SetTop returns the bounds with the top edge moved to the given value.
// The bottom edge stays fixed, so the height adjusts.
func (bn Bounds) SetTop(value int) Bounds {

	return Bounds{bn.X, value, bn.W, bn.H + (bn.Y - value)}
}

////////////////////////////////////////////////////////////////////////////////

// GetRight returns the right edge of the bounds.
func (bn Bounds) GetRight() int {

	return bn.X + bn.W
}

////////////////////////////////////////////////////////////////////////////////

// SetRight returns the bounds with the right edge moved to the
// given value. The left edge stays fixed, so the width adjusts.
func (bn Bounds) SetRight(value int) Bounds {

	return Bounds{bn.X, bn.Y, value - bn.X, bn.H}
}

////////////////////////////////////////////////////////////////////////////////

// GetBottom returns the bottom edge of the bounds.
func (bn Bounds) GetBottom() int {

	return bn.Y + bn.H
}

////////////////////////////////////////////////////////////////////////////////

// SetBottom returns the bounds with the bottom edge moved to the
// given value. The top edge stays fixed, so the height adjusts.
func (bn Bounds) SetBottom(value int) Bounds {

	return Bounds{bn.X, bn.Y, bn.W, value - bn.Y}
}

////////////////////////////////////////////////////////////////////////////////

// GetLtrb returns the left, top, right, and bottom edge values.
func (bn Bounds) GetLtrb() (l, t, r, b int) {

	return bn.X, bn.Y, bn.X + bn.W, bn.Y + bn.H
}

////////////////////////////////////////////////////////////////////////////////

// SetLtrb returns the bounds set from left, top, right, and bottom edge values.
func (bn Bounds) SetLtrb(l, t, r, b int) Bounds {

	return Bounds{l, t, r - l, b - t}
}

////////////////////////////////////////////////////////////////////////////////

// Normalize returns the bounds with negative dimensions resolved by
// adjusting the position and making dimensions positive.
func (bn Bounds) Normalize() Bounds {

	x := bn.X
	y := bn.Y
	w := bn.W
	h := bn.H

	if w < 0 {
		x += w
		w = -w
	}

	if h < 0 {
		y += h
		h = -h
	}

	return Bounds{x, y, w, h}
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the given point is inside the bounds.
// When inclusive is true, the edges are included. Handles negative
// dimensions.
func (bn Bounds) Contains(point Point, inclusive bool) bool {

	l := bn.X
	r := bn.X + bn.W
	if bn.W < 0 {
		l, r = r, l
	}

	t := bn.Y
	b := bn.Y + bn.H
	if bn.H < 0 {
		t, b = b, t
	}

	if inclusive {
		return point.X >= l && point.X <= r && point.Y >= t && point.Y <= b
	}

	return point.X > l && point.X < r && point.Y > t && point.Y < b
}

////////////////////////////////////////////////////////////////////////////////

// ContainsBounds returns whether the given bounds is entirely inside
// this bounds. When inclusive is true, the edges are included.
// Handles negative dimensions on both bounds.
func (bn Bounds) ContainsBounds(other Bounds, inclusive bool) bool {

	l1 := bn.X
	r1 := bn.X + bn.W
	if bn.W < 0 {
		l1, r1 = r1, l1
	}

	t1 := bn.Y
	b1 := bn.Y + bn.H
	if bn.H < 0 {
		t1, b1 = b1, t1
	}

	l2 := other.X
	r2 := other.X + other.W
	if other.W < 0 {
		l2, r2 = r2, l2
	}

	t2 := other.Y
	b2 := other.Y + other.H
	if other.H < 0 {
		t2, b2 = b2, t2
	}

	if inclusive {
		return l2 >= l1 && r2 <= r1 && t2 >= t1 && b2 <= b1
	}

	return l2 > l1 && r2 < r1 && t2 > t1 && b2 < b1
}

////////////////////////////////////////////////////////////////////////////////

// Intersects returns whether this bounds overlaps the given bounds.
// When inclusive is true, touching edges are considered
// intersecting. Handles negative dimensions on both bounds.
func (bn Bounds) Intersects(other Bounds, inclusive bool) bool {

	l1 := bn.X
	r1 := bn.X + bn.W
	if bn.W < 0 {
		l1, r1 = r1, l1
	}

	t1 := bn.Y
	b1 := bn.Y + bn.H
	if bn.H < 0 {
		t1, b1 = b1, t1
	}

	l2 := other.X
	r2 := other.X + other.W
	if other.W < 0 {
		l2, r2 = r2, l2
	}

	t2 := other.Y
	b2 := other.Y + other.H
	if other.H < 0 {
		t2, b2 = b2, t2
	}

	if inclusive {
		return l1 <= r2 && r1 >= l2 && t1 <= b2 && b1 >= t2
	}

	return l1 < r2 && r1 > l2 && t1 < b2 && b1 > t2
}

////////////////////////////////////////////////////////////////////////////////

// GetPoint returns the position of the bounds as a `Point`.
func (bn Bounds) GetPoint() Point {

	return Point{bn.X, bn.Y}
}

////////////////////////////////////////////////////////////////////////////////

// SetPoint returns the bounds with the position set to the given point.
// The dimensions are preserved.
func (bn Bounds) SetPoint(value Point) Bounds {

	return Bounds{value.X, value.Y, bn.W, bn.H}
}

////////////////////////////////////////////////////////////////////////////////

// GetSize returns the dimensions of the bounds as a `Size`.
func (bn Bounds) GetSize() Size {

	return Size{bn.W, bn.H}
}

////////////////////////////////////////////////////////////////////////////////

// SetSize returns the bounds with the dimensions set to the given size.
// The position is preserved.
func (bn Bounds) SetSize(value Size) Bounds {

	return Bounds{bn.X, bn.Y, value.W, value.H}
}

////////////////////////////////////////////////////////////////////////////////

// GetCenter returns the center point of the bounds.
func (bn Bounds) GetCenter() Point {

	return Point{bn.X + bn.W/2, bn.Y + bn.H/2}
}

////////////////////////////////////////////////////////////////////////////////

// Unite returns the smallest bounds that contains both this bounds and the
// given bounds. Handles negative dimensions on both bounds.
func (bn Bounds) Unite(other Bounds) Bounds {

	l1 := bn.X
	r1 := bn.X + bn.W
	if bn.W < 0 {
		l1, r1 = r1, l1
	}

	t1 := bn.Y
	b1 := bn.Y + bn.H
	if bn.H < 0 {
		t1, b1 = b1, t1
	}

	l2 := other.X
	r2 := other.X + other.W
	if other.W < 0 {
		l2, r2 = r2, l2
	}

	t2 := other.Y
	b2 := other.Y + other.H
	if other.H < 0 {
		t2, b2 = b2, t2
	}

	l := l1
	if l2 < l {
		l = l2
	}

	t := t1
	if t2 < t {
		t = t2
	}

	r := r1
	if r2 > r {
		r = r2
	}

	b := b1
	if b2 > b {
		b = b2
	}

	return Bounds{l, t, r - l, b - t}
}

////////////////////////////////////////////////////////////////////////////////

// Intersect returns the overlapping region of this bounds and the given
// bounds. Returns `BoundsZero` if there is no overlap. Handles negative
// dimensions on both bounds.
func (bn Bounds) Intersect(other Bounds) Bounds {

	l1 := bn.X
	r1 := bn.X + bn.W
	if bn.W < 0 {
		l1, r1 = r1, l1
	}

	t1 := bn.Y
	b1 := bn.Y + bn.H
	if bn.H < 0 {
		t1, b1 = b1, t1
	}

	l2 := other.X
	r2 := other.X + other.W
	if other.W < 0 {
		l2, r2 = r2, l2
	}

	t2 := other.Y
	b2 := other.Y + other.H
	if other.H < 0 {
		t2, b2 = b2, t2
	}

	l := l1
	if l2 > l {
		l = l2
	}

	t := t1
	if t2 > t {
		t = t2
	}

	r := r1
	if r2 < r {
		r = r2
	}

	b := b1
	if b2 < b {
		b = b2
	}

	w := r - l
	h := b - t

	if w <= 0 || h <= 0 {
		return BoundsZero
	}

	return Bounds{l, t, w, h}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// BoundsFromPointSize creates a bounds from a position and size.
func BoundsFromPointSize(point Point, size Size) Bounds {

	return Bounds{point.X, point.Y, size.W, size.H}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Eq returns whether the bounds are equal.
func (bn Bounds) Eq(value Bounds) bool {

	return bn.X == value.X && bn.Y == value.Y && bn.W == value.W && bn.H == value.H
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns whether the bounds are not equal.
func (bn Bounds) Ne(value Bounds) bool {

	return bn.X != value.X || bn.Y != value.Y || bn.W != value.W || bn.H != value.H
}
