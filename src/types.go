package main

import "image/color"

const MAX_COLOR_SIZE = 255 * 255 * 3
const DIR_PERMISSION = 0755

type SortFunc func([]color.NRGBA)
type DiffFunc func(a, b color.NRGBA) int
type FitnessFunc func(pixel color.NRGBA, neighbours []Point, img []color.NRGBA) int
type SelectFunc func(pixel color.NRGBA, unfilled []Point, neighbour_list [][]Point, img []color.NRGBA) int

type Point struct {
	X, Y, Index int
}

func new_point(x, y int) Point {
	return Point{X: x, Y: y, Index: y*width + x}
}

type Config struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Step   int    `json:"step"`
	Seed   int64  `json:"seed"`
	Sort   string `json:"sort"`
	Diff   string `json:"diff"`
	Fit    string `json:"fit"`
	Select string `json:"select"`
	Dist   string `json:"dist"`
	Cpu    string `json:"cpu"`
	Mem    string `json:"mem"`
	Name   string `json:"name"`
}
