package main

import (
	"runtime"
)

func init_neighbours(width, height int) [][]int {
	neighbours := make([][]int, width*height)
	dirs := [8][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*width + x
			neighbours[index] = make([]int, 0, 8)
			for _, dir := range dirs {
				nx, ny := x+dir[0], y+dir[1]
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					neighbours[index] = append(neighbours[index], ny*width+nx)
				}
			}
		}
	}

	return neighbours
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mem_usage() runtime.MemStats {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return mem
}
