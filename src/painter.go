package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime"
	"time"
)

var (
	GOROUTINE_EACH = 4096
)

func painter(colors []color.NRGBA) {
	neighbours := init_neighbours(width, height)

	img := make([]color.NRGBA, width*height)
	candidates := make([]int, 0, width*height)

	is_painted := make([]bool, width*height)
	is_candidate := make([]bool, width*height)
	for i := 0; i < width*height; i++ {
		is_painted[i] = false
		is_candidate[i] = false
	}

	mem := mem_usage()
	if !pipe_only && !json_progress {
		fmt.Printf("Initialized. Memory: %d MB\n", mem.Alloc>>20)
	}

	current := center_y*width + center_x

	start_time, candidates_n := time.Now(), 0
	prev_time, prev_n := start_time, candidates_n
	for i := 0; i < width*height; i++ {
		candidates_n += len(candidates)
		if !pipe_only && i%step == 0 {
			duration := max(int(time.Since(start_time).Seconds()*1000), 1)
			span := max(int(time.Since(prev_time).Seconds()*1000), 1)
			mem := mem_usage()
			info := ProgressInfo{
				Done:   i,
				Total:  width * height,
				Time:   duration,
				Node:   len(candidates),
				Power:  int((candidates_n - prev_n) * 1000 / span),
				Speed:  i * 1000 / duration,
				Memory: mem.Alloc >> 20,
				GC:     mem.NumGC,
			}
			if json_progress {
				output, _ := json.Marshal(info)
				fmt.Printf("%s\n", output)
			} else {
				fmt.Printf(
					"%6.2f%%, node: %5d, speed: %5d px/sec (%9d) | Memory: %4d MB (GC: %3d)\n",
					float64(info.Done)*100/float64(info.Total), info.Node, info.Speed, info.Power, info.Memory, info.GC,
				)
			}
			prev_n, prev_time = candidates_n, time.Now()

			saved := make([]color.NRGBA, len(img))
			for i := 0; i < len(img); i++ {
				copy(saved[i:i+1], img[i:i+1])
			}
			go save_img(fmt.Sprintf("smoke-progress-%05d.png", i/step), saved)
		}

		if len(candidates) != 0 {
			index := select_best(candidates, neighbours, img, colors[i])
			current = candidates[index]

			candidates[len(candidates)-1], candidates[index] = candidates[index], candidates[len(candidates)-1]
			candidates = candidates[:len(candidates)-1]
			is_candidate[current] = false
		}

		img[current] = colors[i]
		is_painted[current] = true

		for _, neighbour := range neighbours[current] {
			if !is_candidate[neighbour] && !is_painted[neighbour] {
				candidates = append(candidates, neighbour)
				is_candidate[neighbour] = true
			}
		}
	}

	if !pipe_only {
		save_img(creation_name+".png", img)
		duration := max(int(time.Since(start_time).Seconds()*1000), 1)
		span := max(int(time.Since(prev_time).Seconds()*1000), 1)
		mem = mem_usage()
		info := ProgressInfo{
			Done:   width * height,
			Total:  width * height,
			Time:   duration,
			Node:   len(candidates),
			Power:  int((candidates_n - prev_n) * 1000 / span),
			Speed:  width * height * 1000 / duration,
			Memory: mem.Alloc >> 20,
			GC:     mem.NumGC,
		}
		if json_progress {
			output, _ := json.Marshal(info)
			fmt.Printf("%s\n", output)
		} else {
			fmt.Printf(
				"%6.2f%%, rendered in %.2f seconds.\n",
				float64(info.Done)*100/float64(info.Total), float64(info.Time)/1000,
			)
		}
	} else {
		painting := image.NewNRGBA(image.Rect(0, 0, width, height))
		for index, color := range img {
			painting.SetNRGBA(index%width, index/width, color)
		}

		err := png.Encode(os.Stdout, painting)
		if err != nil {
			panic(err)
		}
	}
}

// Use the race condition of goroutines to make randomize-effect
func select_best(candidates []int, neighbours [][]int, img []color.NRGBA, color color.NRGBA) int {
	subtasks := min((len(candidates)+GOROUTINE_EACH-1)/GOROUTINE_EACH, runtime.NumCPU())
	channel := make(chan int, subtasks)

	best_fitness, best_index := MAX_COLOR_SIZE, 0

	for i := 0; i < subtasks; i++ {
		go func(i int) {
			for index := i; index < len(candidates); index += subtasks {
				for idx := range neighbours[candidates[index]] {
					// inline diff_rgb
					// fitness := diff_rgb(color, img[neighbours[candidates[index]][idx]])

					diff_r := int(color.R) - int(img[neighbours[candidates[index]][idx]].R)
					diff_g := int(color.G) - int(img[neighbours[candidates[index]][idx]].G)
					diff_b := int(color.B) - int(img[neighbours[candidates[index]][idx]].B)

					fitness := diff_r*diff_r + diff_g*diff_g + diff_b*diff_b

					if fitness < best_fitness {
						best_fitness = fitness
						best_index = index
					}
				}
			}

			channel <- i
		}(i)
	}

	for i := 0; i < subtasks; i++ {
		<-channel
	}

	return best_index
}
