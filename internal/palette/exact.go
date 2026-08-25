// Package palette builds the fixed deterministic color sequence consumed by the renderer.
package palette

import (
	"fmt"
	"image/color"
	"math"
)

const (
	hueBuckets = 360
	maxColors  = 256 * 256 * 256
)

// Exact returns at least pixelCount colors in deterministic HCL-bucket order.
// Seed controls the hue rotation and direction; invalid pixel counts return an error.
func Exact(pixelCount int, seed int64) ([]color.NRGBA, error) {
	if pixelCount < 1 || pixelCount > maxColors {
		return nil, fmt.Errorf("pixel count must be between 1 and %d, got %d", maxColors, pixelCount)
	}

	levels := cubeCeiling(pixelCount)
	if levels > 255 {
		levels = 255
	}
	axisEntries := levels + 1
	generatedCount := axisEntries * axisEntries * axisEntries
	if generatedCount < pixelCount {
		return nil, fmt.Errorf("exact palette has %d entries for %d pixels", generatedCount, pixelCount)
	}

	random := splitMix64{state: uint64(seed)}
	descending := random.next()&1 != 0
	offset := float64(random.next()>>11) * (1.0 / (1 << 53)) * hueBuckets
	linear := linearRGBTable()

	colors := make([]color.NRGBA, generatedCount)
	keys := make([]uint16, generatedCount)
	var counts [hueBuckets]int
	denominator := max(levels-1, 1)
	position := 0
	for red := 0; red <= levels; red++ {
		for green := 0; green <= levels; green++ {
			for blue := 0; blue <= levels; blue++ {
				value := color.NRGBA{
					R: uint8(red * 255 / denominator),
					G: uint8(green * 255 / denominator),
					B: uint8(blue * 255 / denominator),
					A: 255,
				}
				bucket := int(hclHue(value, linear)+offset) % hueBuckets
				colors[position] = value
				keys[position] = uint16(bucket)
				counts[bucket]++
				position++
			}
		}
	}

	var starts [hueBuckets]int
	next := 0
	if descending {
		for bucket := hueBuckets - 1; bucket >= 0; bucket-- {
			starts[bucket] = next
			next += counts[bucket]
		}
	} else {
		for bucket := 0; bucket < hueBuckets; bucket++ {
			starts[bucket] = next
			next += counts[bucket]
		}
	}

	ordered := make([]color.NRGBA, generatedCount)
	for index, value := range colors {
		bucket := int(keys[index])
		ordered[starts[bucket]] = value
		starts[bucket]++
	}

	exact := make([]color.NRGBA, pixelCount)
	copy(exact, ordered[:pixelCount])
	return exact, nil
}

type splitMix64 struct {
	state uint64
}

func (random *splitMix64) next() uint64 {
	random.state += 0x9e3779b97f4a7c15
	value := random.state
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func cubeCeiling(value int) int {
	root := int(math.Cbrt(float64(value)))
	for root*root*root < value {
		root++
	}
	for root > 1 && (root-1)*(root-1)*(root-1) >= value {
		root--
	}
	return root
}

func linearRGBTable() [256]float64 {
	var table [256]float64
	for value := range table {
		srgb := float64(value) / 255
		if srgb <= 0.04045 {
			table[value] = srgb / 12.92
		} else {
			table[value] = math.Pow((srgb+0.055)/1.055, 2.4)
		}
	}
	return table
}

func hclHue(value color.NRGBA, linear [256]float64) float64 {
	red := linear[value.R]
	green := linear[value.G]
	blue := linear[value.B]

	x := (red*0.4124564 + green*0.3575761 + blue*0.1804375) / 0.95047
	y := red*0.2126729 + green*0.7151522 + blue*0.0721750
	z := (red*0.0193339 + green*0.1191920 + blue*0.9503041) / 1.08883

	fx := labTransform(x)
	fy := labTransform(y)
	fz := labTransform(z)
	a := 500 * (fx - fy)
	b := 200 * (fy - fz)
	hue := math.Atan2(b, a) * 180 / math.Pi
	if hue < 0 {
		hue += hueBuckets
	}
	return hue
}

func labTransform(value float64) float64 {
	delta := 6.0 / 29.0
	if value > delta*delta*delta {
		return math.Cbrt(value)
	}
	return value/(3*delta*delta) + 4.0/29.0
}
