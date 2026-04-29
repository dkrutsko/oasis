//go:build darwin || linux

package overlay

import (
	"image/color"
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

// Canvas provides direct pixel drawing into an RGBA byte
// buffer. All coordinates are in pixels. Drawing outside
// the buffer bounds is silently clipped.
type Canvas struct {
	Pix    []byte
	Width  int
	Height int
	Stride int
}

////////////////////////////////////////////////////////////////////////////////

// Clear sets all pixels to transparent black.
func (c *Canvas) Clear() {

	clear(c.Pix)
}

////////////////////////////////////////////////////////////////////////////////

// SetPixel sets a single pixel at (x, y).
func (c *Canvas) SetPixel(x, y int, col color.NRGBA) {

	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}

	off := y*c.Stride + x*4
	c.Pix[off+0] = col.R
	c.Pix[off+1] = col.G
	c.Pix[off+2] = col.B
	c.Pix[off+3] = col.A
}

////////////////////////////////////////////////////////////////////////////////

// DrawLine draws a line from (x0, y0) to (x1, y1) using
// Bresenham's algorithm with the given thickness and color.
func (c *Canvas) DrawLine(x0, y0, x1, y1 float64, thickness int, col color.NRGBA) {

	ix0 := int(sysMath.Round(x0))
	iy0 := int(sysMath.Round(y0))
	ix1 := int(sysMath.Round(x1))
	iy1 := int(sysMath.Round(y1))

	dx := abs(ix1 - ix0)
	dy := -abs(iy1 - iy0)
	sx := 1
	sy := 1
	if ix0 > ix1 {
		sx = -1
	}
	if iy0 > iy1 {
		sy = -1
	}
	err := dx + dy

	half := thickness / 2

	for {
		// Draw a square of pixels for thickness
		for ty := -half; ty <= half; ty++ {
			for tx := -half; tx <= half; tx++ {
				c.SetPixel(ix0+tx, iy0+ty, col)
			}
		}

		if ix0 == ix1 && iy0 == iy1 {
			break
		}

		e2 := 2 * err
		if e2 >= dy {
			err += dy
			ix0 += sx
		}
		if e2 <= dx {
			err += dx
			iy0 += sy
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// DrawCircle draws a filled circle at (cx, cy) with the
// given radius and color.
func (c *Canvas) DrawCircle(cx, cy float64, radius int, col color.NRGBA) {

	icx := int(sysMath.Round(cx))
	icy := int(sysMath.Round(cy))
	r2 := radius * radius

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= r2 {
				c.SetPixel(icx+dx, icy+dy, col)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func abs(x int) int {

	if x < 0 {
		return -x
	}
	return x
}
