package main

import "image/color"

func fitness_min(pixel color.NRGBA, neighbours []Point, img []color.NRGBA) int {
	min_diff := MAX_COLOR_SIZE

	for _, neighbour := range neighbours {
		diff := diff_func(pixel, img[neighbour.Index])
		min_diff = min(min_diff, diff)
	}

	return min_diff
}

func fitness_sum(pixel color.NRGBA, neighbours []Point, img []color.NRGBA) int {
	sum_diff := 0

	for _, neighbour := range neighbours {
		sum_diff += diff_func(pixel, img[neighbour.Index])
	}

	return sum_diff
}
