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
	diff_algo     string
	fit_algo      string
	select_algo   string
	dist          string
	cpu_profile   string
	mem_profile   string
	creation_name string
	config_file   string
	sort_func     SortFunc
	diff_func     DiffFunc
	fit_func      FitnessFunc
	select_func   SelectFunc
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		help           bool
		_width         int
		_height        int
		_center_x      int
		_center_y      int
		_step          int
		_seed          int64
		_sort_algo     string
		_diff_algo     string
		_fit_algo      string
		_select_algo   string
		_dist          string
		_cpu_profile   string
		_mem_profile   string
		_creation_name string
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
	flag.StringVar(&_diff_algo, "diff", "rgb", "Difference algorithm, can be: rgb or lab")
	flag.StringVar(&_fit_algo, "fit", "min", "Fitness algorithm, can be: min or sum")
	flag.StringVar(&_select_algo, "select", "smallest", "Selection algorithm, can be: smallest or greatest")
	flag.StringVar(&_dist, "dist", "img", "Output directory")
	flag.StringVar(&_cpu_profile, "cpu", "", "Write CPU profile to the given file")
	flag.StringVar(&_mem_profile, "mem", "", "Write memory profile to the given file")
	flag.StringVar(&_creation_name, "name", "", "Name of the creation")

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
	if _diff_algo != "" {
		diff_algo = _diff_algo
	}
	if _fit_algo != "" {
		fit_algo = _fit_algo
	}
	if _select_algo != "" {
		select_algo = _select_algo
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
	if diff_algo == "" {
		diff_algo = "rgb"
	}
	if fit_algo == "" {
		fit_algo = "min"
	}
	if select_algo == "" {
		select_algo = "smallest"
	}
	if dist == "" {
		dist = "img"
	}
	if creation_name == "" {
		creation_name = fmt.Sprintf("creation-%d", time.Now().Unix())
	}
	// #endregion

	rand.Seed(seed)
}

func main() {
	color_size := int(math.Ceil(math.Cbrt(float64(width) * float64(height))))
	fmt.Printf(
		"Rendering a %dx%d image with %d colors\n",
		width, height, color_size*color_size*color_size,
	)

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

	switch diff_algo {
	case "rgb":
		diff_func = diff_rgb
	case "lab":
		diff_func = diff_lab
	default:
		diff_func = diff_rgb
	}

	switch fit_algo {
	case "min":
		fit_func = fitness_min
	case "sum":
		fit_func = fitness_sum
	default:
		fit_func = fitness_min
	}

	switch select_algo {
	case "smallest":
		select_func = select_smallest
	case "greatest":
		select_func = select_greatest
	default:
		select_func = select_smallest
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
			Width:  width,
			Height: height,
			X:      center_x,
			Y:      center_y,
			Step:   step,
			Seed:   seed,
			Sort:   sort_algo,
			Diff:   diff_algo,
			Fit:    fit_algo,
			Select: select_algo,
			Dist:   dist,
			Cpu:    cpu_profile,
			Mem:    mem_profile,
			Name:   creation_name,
		},
		"",
		"    ",
	)
	if err != nil {
		log.Fatal("Couldn't save config: ", err.Error())
	}
	save_config("config.json", config)

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
	diff_algo = config.Diff
	fit_algo = config.Fit
	select_algo = config.Select
	dist = config.Dist
	cpu_profile = config.Cpu
	mem_profile = config.Mem
	creation_name = config.Name
}
