package main

import (
	"fmt"
	"image/color"
	"sync"
	"time"
)

var GOROUTINE_EACH = 4096

func painter(colors []color.NRGBA) {
	img := make([]color.NRGBA, width*height)
	painted := make(map[int]bool, width*height)
	nodes := make([]Node, 0, width*height)
	existing_node := make(map[int]bool, width*height)
	neighbour_list := init_neighbours()
	fmt.Printf("Initialized. ")
	mem_usage()

	current := new_node(center_x, center_y, len(neighbour_list[center_y*width+center_x]))

	start_time := time.Now()
	nodes_n := 0
	for i := 0; i < width*height; i++ {
		nodes_n += len(nodes)
		if i%step == 0 {
			duration := max(int(time.Since(start_time).Seconds()), 1)
			fmt.Printf(
				"%6.2f%%, node: %5d, speed: %5d px/sec (%8d) | ",
				float64(100*i)/float64(width*height), len(nodes), i/duration, nodes_n/duration,
			)
			mem_usage()

			saved := make([]color.NRGBA, len(img))
			for i := 0; i < len(img); i++ {
				copy(saved[i:i+1], img[i:i+1])
			}
			go save_img(fmt.Sprintf("smoke-progress-%05d.png", i/step), saved)
		}

		if len(nodes) != 0 {
			update(nodes, neighbour_list, colors[i], img)

			index := select_func(&nodes)
			current = nodes[index]

			nodes[len(nodes)-1], nodes[index] = nodes[index], nodes[len(nodes)-1]
			nodes = nodes[:len(nodes)-1]
			delete(existing_node, current.Point.Index)
		}

		img[current.Point.Index] = colors[i]
		painted[current.Point.Index] = true

		for _, neighbour := range neighbour_list[current.Point.Index] {
			if !existing_node[neighbour.Index] && !painted[neighbour.Index] {
				nodes = append(nodes, new_node(neighbour.X, neighbour.Y, len(neighbour_list[neighbour.Index])))
				existing_node[neighbour.Index] = true
			}
		}
	}

	save_img(creation_name+".png", img)
	fmt.Printf("100.00%% Rendered in %.2f seconds\n", time.Since(start_time).Seconds())
	mem_usage()
}

func update(nodes []Node, neighbour_list [][]Point, color color.NRGBA, img []color.NRGBA) {
	if len(nodes) <= GOROUTINE_EACH {
		for index := 0; index < len(nodes); index++ {
			for neighbour_index, neighbour := range neighbour_list[nodes[index].Point.Index] {
				nodes[index].Diffs[neighbour_index] = diff_func(color, img[neighbour.Index])
			}
			nodes[index].Fitness = fit_func(&nodes[index])
		}
	} else {
		wg := new(sync.WaitGroup)

		for i := 0; i < len(nodes); i += GOROUTINE_EACH {
			wg.Add(1)
			go func(i int) {
				for index := i; index < min(i+GOROUTINE_EACH, len(nodes)); index++ {
					for neighbour_index, neighbour := range neighbour_list[nodes[index].Point.Index] {
						nodes[index].Diffs[neighbour_index] = diff_func(color, img[neighbour.Index])
					}
					nodes[index].Fitness = fit_func(&nodes[index])
				}
				wg.Done()
			}(i)
		}

		wg.Wait()
	}
}
