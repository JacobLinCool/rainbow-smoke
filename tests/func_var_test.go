package smoke_test

import (
	"math"
	"testing"
)

type point_3d struct {
	x, y, z int
}
type Func func(p1, p2 point_3d) int

func function(p1, p2 point_3d) int {
	return int(math.Pow(float64(p1.x-p2.x), 2) + math.Pow(float64(p1.y-p2.y), 2) + math.Pow(float64(p1.z-p2.z), 2))
}

var f_global Func = function

func BenchmarkInline(b *testing.B) {
	x := 0
	for i := 0; i < b.N; i++ {
		p1, p2 := point_3d{i, i + 1, i + 2}, point_3d{i + 1, i + 2, i + 3}
		x = int(math.Pow(float64(p1.x-p2.x), 2) + math.Pow(float64(p1.y-p2.y), 2) + math.Pow(float64(p1.z-p2.z), 2))
	}
	_ = x
}

func BenchmarkDirect(b *testing.B) {
	x := 0
	for i := 0; i < b.N; i++ {
		x = function(point_3d{i, i + 1, i + 2}, point_3d{i + 1, i + 2, i + 3})
	}
	_ = x
}

func BenchmarkGlobal(b *testing.B) {
	x := 0
	for i := 0; i < b.N; i++ {
		x = f_global(point_3d{i, i + 1, i + 2}, point_3d{i + 1, i + 2, i + 3})
	}
	_ = x
}

func BenchmarkLocal(b *testing.B) {
	f_local := function

	x := 0
	for i := 0; i < b.N; i++ {
		x = f_local(point_3d{i, i + 1, i + 2}, point_3d{i + 1, i + 2, i + 3})
	}
	_ = x
}

func BenchmarkLocalX(b *testing.B) {
	x := 0
	for i := 0; i < b.N; i++ {
		f_local := function
		x = f_local(point_3d{i, i + 1, i + 2}, point_3d{i + 1, i + 2, i + 3})
	}
	_ = x
}

func BenchmarkClosure(b *testing.B) {
	f_closure := func(p1, p2 point_3d) int {
		return int(math.Pow(float64(p1.x-p2.x), 2) + math.Pow(float64(p1.y-p2.y), 2) + math.Pow(float64(p1.z-p2.z), 2))
	}

	x := 0
	for i := 0; i < b.N; i++ {
		x = f_closure(point_3d{i, i + 1, i + 2}, point_3d{i + 1, i + 2, i + 3})
	}
	_ = x
}
