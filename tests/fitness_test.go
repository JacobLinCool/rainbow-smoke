package smoke_test

import "testing"

var (
	width  = 4096
	height = 4096
)

func BenchmarkNothing(b *testing.B) {
	nodes := make([]Node, 0, width*height)
	for i := 0; i < width*height; i++ {
		nodes = append(nodes, new_node(i%width, i/width, 8))
		for j := range nodes[i].Diffs {
			nodes[i].Diffs[j] = i - j
		}
	}

	for i := 0; i < b.N; i++ {
		for index := range nodes {
			nodes[index].Fitness = 1
		}
	}
}

func BenchmarkFitnessCopy(b *testing.B) {
	nodes := make([]Node, 0, width*height)
	for i := 0; i < width*height; i++ {
		nodes = append(nodes, new_node(i%width, i/width, 8))
		for j := range nodes[i].Diffs {
			nodes[i].Diffs[j] = i - j
		}
	}

	for i := 0; i < b.N; i++ {
		for index := range nodes {
			nodes[index].Fitness = fitness_min_copy(&nodes[index])
		}
	}
}

func BenchmarkFitness(b *testing.B) {
	nodes := make([]Node, 0, width*height)
	for i := 0; i < width*height; i++ {
		nodes = append(nodes, new_node(i%width, i/width, 8))
		for j := range nodes[i].Diffs {
			nodes[i].Diffs[j] = i - j
		}
	}

	for i := 0; i < b.N; i++ {
		for index := range nodes {
			nodes[index].Fitness = fitness_min(&nodes[index])
		}
	}
}

func fitness_min_copy(node *Node) int {
	min_diff := 195075

	for _, diff := range node.Diffs {
		if diff < min_diff {
			min_diff = diff
		}
	}

	return min_diff
}

func fitness_min(node *Node) int {
	min_diff := 195075

	for index := range node.Diffs {
		if node.Diffs[index] < min_diff {
			min_diff = node.Diffs[index]
		}
	}

	return min_diff
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
