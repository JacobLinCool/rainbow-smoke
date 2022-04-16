package main

import (
	"image/color"
)

func diff_rgb(a, b color.NRGBA) int {
	diff_r := int(a.R) - int(b.R)
	diff_g := int(a.G) - int(b.G)
	diff_b := int(a.B) - int(b.B)

	return diff_r*diff_r + diff_g*diff_g + diff_b*diff_b
}
