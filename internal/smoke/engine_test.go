package smoke

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"testing"
)

func TestRenderMatchesFullFrontierOracle(t *testing.T) {
	config := Config{Width: 9, Height: 7, CenterX: 4, CenterY: 3}
	colors := uniqueTestPalette(config.Width * config.Height)
	engine, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	got, stats, err := engine.Render(context.Background(), colors)
	if err != nil {
		t.Fatal(err)
	}
	want := renderScalar(config, colors)
	if !bytes.Equal(got.Pix, want.Pix) {
		t.Fatal("octree placement differs from the full-frontier oracle")
	}
	if stats.Queries != config.Width*config.Height-1 {
		t.Fatalf("queries = %d, want %d", stats.Queries, config.Width*config.Height-1)
	}
	if _, _, err := engine.Render(context.Background(), colors); err == nil {
		t.Fatal("single-use engine accepted a second render")
	}
}

func TestNewRejectsInvalidGeometry(t *testing.T) {
	invalid := []Config{
		{Width: 1, Height: 2, CenterX: 0, CenterY: 0},
		{Width: 2, Height: 2, CenterX: 2, CenterY: 0},
		{Width: 2, Height: 2, CenterX: 0, CenterY: -1},
	}
	for _, config := range invalid {
		if _, err := New(config); err == nil {
			t.Fatalf("New(%+v) unexpectedly succeeded", config)
		}
	}
}

func uniqueTestPalette(count int) []color.NRGBA {
	colors := make([]color.NRGBA, count)
	for index := range colors {
		colors[index] = color.NRGBA{
			R: uint8(index * 37),
			G: uint8(index*73 + 11),
			B: uint8(index*109 + 19),
			A: 255,
		}
	}
	return colors
}

func renderScalar(config Config, colors []color.NRGBA) *image.NRGBA {
	pixelCount := config.Width * config.Height
	painting := image.NewNRGBA(image.Rect(0, 0, config.Width, config.Height))
	painted := make([]bool, pixelCount)
	candidate := make([]bool, pixelCount)
	candidates := make([]int, 0, pixelCount)
	current := config.CenterY*config.Width + config.CenterX

	neighbours := func(pixel int, result *[8]int) int {
		x := pixel % config.Width
		y := pixel / config.Width
		count := 0
		for _, direction := range neighbourDirections {
			nextX := x + direction[0]
			nextY := y + direction[1]
			if nextX < 0 || nextX >= config.Width || nextY < 0 || nextY >= config.Height {
				continue
			}
			result[count] = nextY*config.Width + nextX
			count++
		}
		return count
	}

	for _, target := range colors[:pixelCount] {
		if len(candidates) > 0 {
			bestDistance := int(^uint(0) >> 1)
			bestRank := int(^uint(0) >> 1)
			bestCandidate := 0
			var adjacent [8]int
			for candidateIndex, pixel := range candidates {
				count := neighbours(pixel, &adjacent)
				for slot := 0; slot < count; slot++ {
					offset := adjacent[slot] * 4
					deltaR := int(target.R) - int(painting.Pix[offset])
					deltaG := int(target.G) - int(painting.Pix[offset+1])
					deltaB := int(target.B) - int(painting.Pix[offset+2])
					distance := deltaR*deltaR + deltaG*deltaG + deltaB*deltaB
					rank := candidateIndex*8 + slot
					if distance < bestDistance || distance == bestDistance && rank < bestRank {
						bestDistance = distance
						bestRank = rank
						bestCandidate = candidateIndex
					}
				}
			}

			current = candidates[bestCandidate]
			last := len(candidates) - 1
			candidates[bestCandidate] = candidates[last]
			candidates = candidates[:last]
			candidate[current] = false
		}

		offset := current * 4
		painting.Pix[offset] = target.R
		painting.Pix[offset+1] = target.G
		painting.Pix[offset+2] = target.B
		painting.Pix[offset+3] = 255
		painted[current] = true

		var adjacent [8]int
		count := neighbours(current, &adjacent)
		for index := 0; index < count; index++ {
			pixel := adjacent[index]
			if !candidate[pixel] && !painted[pixel] {
				candidates = append(candidates, pixel)
				candidate[pixel] = true
			}
		}
	}
	return painting
}
