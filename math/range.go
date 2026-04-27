package math

import (
	"fmt"
	"math/rand"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

var (
	// RangeZero represents a range with both components set to zero.
	RangeZero = Range{0, 0}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Range represents a scalar interval defined by minimum and
// maximum float64 values.
type Range struct {
	Min float64
	Max float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the range.
func (r Range) String() string {

	return fmt.Sprintf("[%.2f, %.2f)", r.Min, r.Max)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether both components are zero.
func (r Range) IsZero() bool {

	return r.Min == 0 && r.Max == 0
}

////////////////////////////////////////////////////////////////////////////////

// GetSpan returns the distance between the minimum and maximum.
func (r Range) GetSpan() float64 {

	return r.Max - r.Min
}

////////////////////////////////////////////////////////////////////////////////

// GetRandom returns a random value within [Min, Max). Uses
// the global `math/rand` source. Returns `Min` if the range
// has zero or negative span.
func (r Range) GetRandom() float64 {

	span := r.Max - r.Min
	if span <= 0 {
		return r.Min
	}

	return r.Min + rand.Float64()*span
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the given value is within the range.
// When inclusive is true, the endpoints are included.
func (r Range) Contains(value float64, inclusive bool) bool {

	if inclusive {
		return value >= r.Min && value <= r.Max
	}

	return value > r.Min && value < r.Max
}

