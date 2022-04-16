package range_test

import "testing"

var (
	width  = 4096
	height = 4096
)

func BenchmarkSelect(b *testing.B) {
	nodes := make([]Node, 0, width*height)
	for i := 0; i < width*height; i++ {
		nodes = append(nodes, new_node(i%width, i/width, 8))
	}

	for i := 0; i < b.N; i++ {
		select_smallest(&nodes)
	}
}

func BenchmarkSelectCopy(b *testing.B) {
	nodes := make([]Node, 0, width*height)
	for i := 0; i < width*height; i++ {
		nodes = append(nodes, new_node(i%width, i/width, 8))
	}

	for i := 0; i < b.N; i++ {
		select_smallest_copy(&nodes)
	}
}

func select_smallest(nodes *[]Node) int {
	best_fitness := 195075
	best_index := 0

	for index := range *nodes {
		if (*nodes)[index].Fitness < best_fitness {
			best_fitness = (*nodes)[index].Fitness
			best_index = index
		}
	}

	return best_index
}

func select_smallest_copy(nodes *[]Node) int {
	best_fitness := 195075
	best_index := 0

	for index, node := range *nodes {
		if node.Fitness < best_fitness {
			best_fitness = node.Fitness
			best_index = index
		}
	}

	return best_index
}

type Point struct {
	X, Y, Index int
}

func new_point(x, y int) Point {
	return Point{X: x, Y: y, Index: y*width + x}
}

type Node struct {
	Point   Point
	Fitness int
	Diffs   []int
}

func new_node(x, y, diff_size int) Node {
	return Node{Point: new_point(x, y), Fitness: 0, Diffs: make([]int, diff_size)}
}
