package main

import (
	"fmt"
	"image/color"
	"time"
)

func painter(colors []color.NRGBA) {
	img := make([]color.NRGBA, width*height)
	branches := make([]Point, 0, width*height)
	branch_map := make(map[int]bool, width*height)
	painted_map := make(map[int]bool, width*height)
	neighbour_list := init_neighbours()
	fmt.Printf("Initialized. ")
	mem_usage()

	current := new_point(center_x, center_y)

	start_time := time.Now()
	speed := 0
	for i := 0; i < width*height; i++ {
		if i%step == 0 {
			speed = i / max(int(time.Since(start_time).Seconds()), 1)
			fmt.Printf(
				"%6.2f%%, branch: %5d, speed: %5d px/sec | ",
				float64(100*i)/float64(width*height), len(branches), speed,
			)
			mem_usage()
			go save_img(fmt.Sprintf("smoke-progress-%05d.png", i/step), img)
		}

		if len(branches) != 0 {
			idx := select_func(colors[i], branches, neighbour_list, img)
			current = branches[idx]

			branches[len(branches)-1], branches[idx] = branches[idx], branches[len(branches)-1]
			branches = branches[:len(branches)-1]
			delete(branch_map, current.Index)
		}

		painted_map[current.Index] = true

		img[current.Index] = colors[i]

		for _, neighbour := range neighbour_list[current.Index] {
			if !branch_map[neighbour.Index] && !painted_map[neighbour.Index] {
				branches = append(branches, neighbour)
				branch_map[neighbour.Index] = true
			}
		}
	}

	save_img(fmt.Sprintf("rainbow-smoke-%dx%d.png", width, height), img)
	fmt.Printf("100.00%% Rendered in %.2f seconds\n", time.Since(start_time).Seconds())
	mem_usage()
}
