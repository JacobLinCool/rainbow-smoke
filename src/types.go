package main

import (
	"image/color"
)

const MAX_COLOR_SIZE = 255 * 255 * 3
const DIR_PERMISSION = 0755

type SortFunc func([]color.NRGBA)
type DiffFunc func(a, b color.NRGBA) int
type FitnessFunc func(diffs []int) int
type SelectFunc func(nodes []int, fitnesses []int) int

type Config struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Step       int     `json:"step"`
	Seed       int64   `json:"seed"`
	Sort       string  `json:"sort"`
	Dist       string  `json:"dist"`
	CpuProfile string  `json:"cpu"`
	MemProfile string  `json:"mem"`
	Name       string  `json:"name"`
	PipeOnly   bool    `json:"pipe_only"`
	Json       bool    `json:"json"`
	ColorScale float64 `json:"color_scale"`
}

type ProgressInfo struct {
	Done   int    `json:"done"`
	Total  int    `json:"total"`
	Time   int    `json:"time"`
	Node   int    `json:"node"`
	Power  int    `json:"power"`
	Speed  int    `json:"pixel"`
	Memory uint64 `json:"mem"`
	GC     uint32 `json:"gc"`
}
