package main

import (
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
)

func scale_color(c, color_max int) uint8 {
	return uint8((c * 255) / (color_max - 1))
}

func create_color(r, g, b, color_max int) color.NRGBA {
	return color.NRGBA{
		scale_color(r, color_max),
		scale_color(g, color_max),
		scale_color(b, color_max),
		255,
	}
}

func create_colorful(color color.NRGBA) colorful.Color {
	return colorful.Color{
		R: float64(color.R) / 255.0,
		G: float64(color.G) / 255.0,
		B: float64(color.B) / 255.0,
	}
}

func create_color_list(colors int) []color.NRGBA {
	color_list := make([]color.NRGBA, 0, colors*colors*colors)

	for r := 0; r <= colors; r++ {
		for g := 0; g <= colors; g++ {
			for b := 0; b <= colors; b++ {
				color_list = append(color_list, create_color(r, g, b, colors))
			}
		}
	}

	return color_list
}
