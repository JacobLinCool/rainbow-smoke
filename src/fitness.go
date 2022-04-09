package main

func fitness_min(node *Node) int {
	min_diff := MAX_COLOR_SIZE

	for _, diff := range node.Diffs {
		if diff < min_diff {
			min_diff = diff
		}
	}

	return min_diff
}

func fitness_sum(node *Node) int {
	sum_diff := 0

	for _, diff := range node.Diffs {
		sum_diff += diff
	}

	return sum_diff
}
