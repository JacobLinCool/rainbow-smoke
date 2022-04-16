package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

var (
	width         int   = -1
	height        int   = -1
	center_x      int   = -1
	center_y      int   = -1
	step          int   = -1
	seed          int64 = math.MinInt64
	sort_algo     string
	dist          string
	cpu_profile   string
	mem_profile   string
	creation_name string
	pipe_only     bool
	json_progress bool
	color_scale   float64
	config_file   string
	sort_func     SortFunc
)

func init() {
	var (
		help           bool
		_width         int
		_height        int
		_center_x      int
		_center_y      int
		_step          int
		_seed          int64
		_sort_algo     string
		_dist          string
		_cpu_profile   string
		_mem_profile   string
		_creation_name string
		_pipe_only     bool
		_json_progress bool
		_color_scale   float64
	)

	flag.BoolVar(&help, "help", false, "Show this help")
	flag.StringVar(&config_file, "config", "", "Configuration file")
	flag.IntVar(&_width, "width", -1, "Rendered image width, must be at least 2")
	flag.IntVar(&_height, "height", -1, "Rendered image height, must be at least 2")
	flag.IntVar(&_center_x, "x", -1, "X coordinate of the center of the smoke")
	flag.IntVar(&_center_y, "y", -1, "Y coordinate of the center of the smoke")
	flag.IntVar(&_step, "step", -1, "Step for progress image")
	flag.Int64Var(&_seed, "seed", math.MinInt64, "Seed for random number generator")
	flag.StringVar(&_sort_algo, "sort", "hcl", "Sorting algorithm, can be: hcl, hsv, random, or none")
	flag.StringVar(&_dist, "dist", "img", "Output directory")
	flag.StringVar(&_cpu_profile, "cpu", "", "Write CPU profile to the given file")
	flag.StringVar(&_mem_profile, "mem", "", "Write memory profile to the given file")
	flag.StringVar(&_creation_name, "name", "", "Name of the creation")
	flag.BoolVar(&_pipe_only, "pipe", false, "Only pipe output to stdout")
	flag.BoolVar(&_json_progress, "json", false, "Output progress as JSON")
	flag.Float64Var(&_color_scale, "scale", 1.0, "Color scale")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if config_file != "" {
		fmt.Printf("Reading configuration from %s\n", config_file)
		read_config()
	}

	if _width >= 2 {
		width = _width
	}
	if _height >= 2 {
		height = _height
	}
	if _center_x >= 0 {
		center_x = _center_x
	}
	if _center_y >= 0 {
		center_y = _center_y
	}
	if _step > 0 {
		step = _step
	}
	if _seed != math.MinInt64 {
		seed = _seed
	}
	if _sort_algo != "" {
		sort_algo = _sort_algo
	}
	if _dist != "" {
		dist = _dist
	}
	if _cpu_profile != "" {
		cpu_profile = _cpu_profile
	}
	if _mem_profile != "" {
		mem_profile = _mem_profile
	}
	if _creation_name != "" {
		creation_name = _creation_name
	}
	if _pipe_only {
		pipe_only = _pipe_only
	}
	if _json_progress {
		json_progress = _json_progress
	}
	if _color_scale != 1.0 {
		color_scale = _color_scale
	}

	// #region Defaults
	if width < 2 {
		width = 256
	}
	if height < 2 {
		height = 256
	}
	if center_x < 0 || center_x >= width {
		center_x = width / 2
	}
	if center_y < 0 || center_y >= height {
		center_y = height / 2
	}
	if step < 1 {
		step = 1024
	}
	if seed == math.MinInt64 {
		seed = time.Now().UnixNano()
	}
	if sort_algo == "" {
		sort_algo = "hcl"
	}
	if dist == "" {
		dist = "img"
	}
	if creation_name == "" {
		creation_name = fmt.Sprintf("creation-%d", time.Now().Unix())
	}
	// #endregion

	rand.Seed(seed)
	debug.SetGCPercent(200)
}

func main() {
	color_size := int(math.Ceil(math.Cbrt(float64(width) * float64(height))))

	if !pipe_only && !json_progress {
		fmt.Printf(
			"Rendering a %dx%d image with %d colors\n",
			width, height, color_size*color_size*color_size,
		)
	}

	switch sort_algo {
	case "hcl":
		sort_func = sort_hcl
	case "hsv":
		sort_func = sort_hsv
	case "random":
		sort_func = sort_random
	default:
		sort_func = sort_none
	}

	if cpu_profile != "" {
		profile, err := os.Create(cpu_profile)
		if err != nil {
			log.Fatal("Couldn't create CPU profile: ", err.Error())
		}
		defer profile.Close()
		if err := pprof.StartCPUProfile(profile); err != nil {
			log.Fatal("Couldn't start CPU profile: ", err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	config, err := json.MarshalIndent(
		Config{
			Width:      width,
			Height:     height,
			X:          center_x,
			Y:          center_y,
			Step:       step,
			Seed:       seed,
			Sort:       sort_algo,
			Dist:       dist,
			CpuProfile: cpu_profile,
			MemProfile: mem_profile,
			Name:       creation_name,
			ColorScale: color_scale,
		},
		"",
		"    ",
	)
	if err != nil {
		log.Fatal("Couldn't save config: ", err.Error())
	}
	if !pipe_only {
		save_config("config.json", config)
	}

	color_size = min(max(color_size, int(float64(color_size)*color_scale)), 16777216)

	color_list := create_color_list(color_size)
	sort_func(color_list)

	painter(color_list)

	if mem_profile != "" {
		profile, err := os.Create(mem_profile)
		if err != nil {
			log.Fatal("Couldn't create memory profile: ", err.Error())
		}
		defer profile.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(profile); err != nil {
			log.Fatal("Couldn't write memory profile: ", err.Error())
		}
	}
}

func read_config() {
	configBytes, err := ioutil.ReadFile(config_file)
	if err != nil {
		log.Fatal("Couldn't read config: ", err.Error())
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal("Couldn't parse config: ", err.Error())
	}

	width = config.Width
	height = config.Height
	center_x = config.X
	center_y = config.Y
	step = config.Step
	seed = config.Seed
	sort_algo = config.Sort
	dist = config.Dist
	cpu_profile = config.CpuProfile
	mem_profile = config.MemProfile
	creation_name = config.Name
	color_scale = config.ColorScale
}
