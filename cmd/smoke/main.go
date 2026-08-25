package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/JacobLinCool/rainbow-smoke-go/internal/palette"
	"github.com/JacobLinCool/rainbow-smoke-go/internal/smoke"
)

type benchmarkReport struct {
	Width           int         `json:"width"`
	Height          int         `json:"height"`
	Seed            int64       `json:"seed"`
	PaletteMS       float64     `json:"palette_ms"`
	EngineInitMS    float64     `json:"engine_init_ms"`
	PlacementMS     float64     `json:"placement_ms"`
	PNGMS           float64     `json:"png_ms"`
	TotalMS         float64     `json:"total_ms"`
	PixelsPerSecond float64     `json:"pixels_per_second"`
	AllocatedMiB    float64     `json:"allocated_mib"`
	Stats           smoke.Stats `json:"stats"`
}

func main() {
	log.SetFlags(0)
	width := flag.Int("width", 256, "image width")
	height := flag.Int("height", 256, "image height")
	centerX := flag.Int("x", -1, "initial pixel x coordinate; defaults to image center")
	centerY := flag.Int("y", -1, "initial pixel y coordinate; defaults to image center")
	seed := flag.Int64("seed", 20260825, "deterministic exact palette seed")
	output := flag.String("output", "smoke.png", "PNG output path")
	benchmark := flag.Bool("benchmark", false, "measure rendering without PNG encoding")
	flag.Parse()

	if *centerX < 0 {
		*centerX = *width / 2
	}
	if *centerY < 0 {
		*centerY = *height / 2
	}

	started := time.Now()
	pixelCount, err := checkedPixelCount(*width, *height)
	if err != nil {
		log.Fatal(err)
	}

	phase := time.Now()
	colors, err := palette.Exact(pixelCount, *seed)
	if err != nil {
		log.Fatal(err)
	}
	paletteDuration := time.Since(phase)

	phase = time.Now()
	engine, err := smoke.New(smoke.Config{
		Width:   *width,
		Height:  *height,
		CenterX: *centerX,
		CenterY: *centerY,
	})
	if err != nil {
		log.Fatal(err)
	}
	engineInitDuration := time.Since(phase)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	phase = time.Now()
	painting, stats, err := engine.Render(ctx, colors)
	if err != nil {
		log.Fatal(err)
	}
	placementDuration := time.Since(phase)

	pngDuration := time.Duration(0)
	if !*benchmark {
		phase = time.Now()
		if err := writePNG(*output, painting); err != nil {
			log.Fatal(err)
		}
		pngDuration = time.Since(phase)
	}

	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	report := benchmarkReport{
		Width:           *width,
		Height:          *height,
		Seed:            *seed,
		PaletteMS:       milliseconds(paletteDuration),
		EngineInitMS:    milliseconds(engineInitDuration),
		PlacementMS:     milliseconds(placementDuration),
		PNGMS:           milliseconds(pngDuration),
		TotalMS:         milliseconds(time.Since(started)),
		PixelsPerSecond: float64(pixelCount) / placementDuration.Seconds(),
		AllocatedMiB:    float64(memory.Alloc) / (1024 * 1024),
		Stats:           stats,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Fatal(err)
	}
}

func checkedPixelCount(width, height int) (int, error) {
	if width < 2 || height < 2 {
		return 0, fmt.Errorf("width and height must both be at least 2, got %dx%d", width, height)
	}
	const maximum = 256 * 256 * 256
	if width > maximum/height {
		return 0, fmt.Errorf("image exceeds the %d-pixel RGB space", maximum)
	}
	return width * height, nil
}

func writePNG(path string, painting image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create PNG %q: %w", path, err)
	}
	defer file.Close()
	if err := png.Encode(file, painting); err != nil {
		return fmt.Errorf("encode PNG %q: %w", path, err)
	}
	return nil
}

func milliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}
