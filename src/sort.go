package main

import (
	"image/color"
	"math/rand"
	"sort"
)

func sort_hcl(colors []color.NRGBA) {
	offset := rand.Float64() * 360

	sort.Slice(colors, func(i, j int) bool {
		h1, _, _ := create_colorful(colors[i]).Hcl()
		h2, _, _ := create_colorful(colors[j]).Hcl()

		return int(h1+offset)%360 < int(h2+offset)%360
	})
}

func sort_hsv(colors []color.NRGBA) {
	offset := rand.Float64() * 360

	sort.Slice(colors, func(i, j int) bool {
		h1, _, _ := create_colorful(colors[i]).Hsv()
		h2, _, _ := create_colorful(colors[j]).Hsv()

		return int(h1+offset)%360 < int(h2+offset)%360
	})
}

func sort_random(colors []color.NRGBA) {
	sort.Slice(colors, func(i, j int) bool {
		return rand.Intn(2) == 0
	})
}

func sort_none(colors []color.NRGBA) {
	// Do nothing
}
