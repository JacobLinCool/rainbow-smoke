package main

import (
	"image/color"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

func diff_rgb(a, b color.NRGBA) int {
	diff_r := int(a.R) - int(b.R)
	diff_g := int(a.G) - int(b.G)
	diff_b := int(a.B) - int(b.B)

	return diff_r*diff_r + diff_g*diff_g + diff_b*diff_b
}

func diff_lab(a, b color.NRGBA) int {
	// Ugly color space hacking.

	// RGB255 -> sRGB -> Linear RGB
	// This step is needed because go-colorful uses floats internally
	// rather than uint8. Additionally, RGB is a relative measurement, so
	// cannot be used directly.
	ar, ag, ab := create_colorful(a).FastLinearRgb()
	br, bg, bb := create_colorful(b).FastLinearRgb()

	// Linear RGB -> CIE XYZ
	// Here we transform the relative RGB system to absolute XYZ co-ordinates.
	ax, ay, az := colorful.LinearRgbToXyz(ar, ag, ab)
	bx, by, bz := colorful.LinearRgbToXyz(br, bg, bb)

	// CIE XYZ -> CIE L*a*b*
	// And finally, here we transform absolute XYZ to L*a*b*, which is a
	// perception-based color space.
	a_l, a_a, a_b := colorful.XyzToLab(ax, ay, az)
	b_l, b_a, b_b := colorful.XyzToLab(bx, by, bz)

	// And finally we can calculate the perceived difference in color.
	diff := math.Sqrt((a_l-b_l)*(a_l-b_l) + (a_a-b_a)*(a_a-b_a) + (a_b - b_b))

	// Yep, we just went through four color spaces to get what we wanted.

	return int(65535.0 * diff)
}
