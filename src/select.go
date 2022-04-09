package main

func select_smallest(nodes *[]Node) int {
	best_fitness := MAX_COLOR_SIZE
	best_index := 0

	for index, node := range *nodes {
		if node.Fitness < best_fitness {
			best_fitness = node.Fitness
			best_index = index
		}
	}

	return best_index
}

func select_greatest(nodes *[]Node) int {
	best_fitness := 0
	best_index := 0

	for index, node := range *nodes {
		if node.Fitness > best_fitness {
			best_fitness = node.Fitness
			best_index = index
		}
	}

	return best_index
}
