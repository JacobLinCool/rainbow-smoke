package main

import (
	"fmt"
	"runtime"
)

func get_neighbours(point *Point) []Point {
	neighbours := make([]Point, 0, 8)

	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			if x == 0 && y == 0 {
				continue
			}

			neighbour := new_point(point.X+x, point.Y+y)
			if neighbour.X >= 0 && neighbour.X < width && neighbour.Y >= 0 && neighbour.Y < height {
				neighbours = append(neighbours, neighbour)
			}
		}
	}

	return neighbours
}

func init_neighbours() [][]Point {
	neighbour_list := make([][]Point, width*height)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			point := new_point(x, y)
			neighbour_list[point.Index] = get_neighbours(&point)
		}
	}

	return neighbour_list
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

func mem_usage() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	fmt.Printf("Memory: %4v MB (GC Cycle: %3v)\n", mem.Alloc>>20, mem.NumGC)
}
