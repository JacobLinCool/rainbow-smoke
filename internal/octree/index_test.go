package octree

import (
	"image/color"
	"math/rand"
	"testing"
)

func TestNearestMatchesScalarOracle(t *testing.T) {
	index := New()
	active := make(map[uint32]uint32)
	random := rand.New(rand.NewSource(7))
	for rank := uint32(0); rank < 500; rank++ {
		rgb := uint32(random.Intn(1 << 24))
		active[rgb] = rank
		index.Update(rgb, rank)
	}

	for query := 0; query < 1_000; query++ {
		target := color.NRGBA{
			R: uint8(random.Intn(256)),
			G: uint8(random.Intn(256)),
			B: uint8(random.Intn(256)),
			A: 255,
		}
		got, ok := index.Nearest(target)
		if !ok {
			t.Fatal("Nearest reported an empty index")
		}
		wantColor, wantRank := scalarNearest(active, target)
		if got.Color != wantColor || got.Rank != wantRank {
			t.Fatalf("Nearest(%v) = (%06x,%d), want (%06x,%d)", target, got.Color, got.Rank, wantColor, wantRank)
		}
	}
}

func TestNearestUsesEdgeRankForDistanceTies(t *testing.T) {
	index := New()
	index.Update(0x000000, 12)
	index.Update(0x020000, 4)

	got, ok := index.Nearest(color.NRGBA{R: 1, A: 255})
	if !ok {
		t.Fatal("Nearest reported an empty index")
	}
	if got.Color != 0x020000 || got.Rank != 4 {
		t.Fatalf("Nearest distance tie = (%06x,%d), want (020000,4)", got.Color, got.Rank)
	}
}

func TestUpdateRemovesLeaves(t *testing.T) {
	index := New()
	index.Update(0x123456, 9)
	index.Update(0x123456, EmptyRank)
	if _, ok := index.Nearest(color.NRGBA{}); ok {
		t.Fatal("Nearest reported a match after removing the only leaf")
	}
}

func TestMemoryBytes(t *testing.T) {
	const nodeCount = 19_173_961
	if got, want := New().MemoryBytes(), uint64(nodeCount*4); got != want {
		t.Fatalf("MemoryBytes() = %d, want %d", got, want)
	}
}

func scalarNearest(active map[uint32]uint32, target color.NRGBA) (uint32, uint32) {
	bestDistance := int(^uint(0) >> 1)
	bestRank := EmptyRank
	bestColor := uint32(0)
	for rgb, rank := range active {
		red := int(rgb >> 16)
		green := int(rgb >> 8 & 0xff)
		blue := int(rgb & 0xff)
		deltaR := int(target.R) - red
		deltaG := int(target.G) - green
		deltaB := int(target.B) - blue
		distance := deltaR*deltaR + deltaG*deltaG + deltaB*deltaB
		if distance < bestDistance || distance == bestDistance && rank < bestRank {
			bestDistance = distance
			bestRank = rank
			bestColor = rgb
		}
	}
	return bestColor, bestRank
}
