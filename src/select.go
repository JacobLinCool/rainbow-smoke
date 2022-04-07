package main

import (
	"image/color"
)

func select_smallest(pixel color.NRGBA, branches []Point, neighbour_list [][]Point, img []color.NRGBA) int {
	best_fitness := MAX_COLOR_SIZE
	best_index := 0

	for index, point := range branches {
		fitness := fit_func(pixel, neighbour_list[point.Index], img)
		if fitness < best_fitness {
			best_index = index
			best_fitness = fitness
		}
	}

	return best_index
}

func select_greatest(pixel color.NRGBA, branches []Point, neighbour_list [][]Point, img []color.NRGBA) int {
	best_fitness := 0
	best_index := 0

	for index, point := range branches {
		fitness := fit_func(pixel, neighbour_list[point.Index], img)
		if fitness > best_fitness {
			best_index = index
			best_fitness = fitness
		}
	}

	return best_index
}
