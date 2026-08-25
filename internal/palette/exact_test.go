package palette

import (
	"slices"
	"testing"
)

func TestExactIsDeterministicAndSized(t *testing.T) {
	first, err := Exact(4_096, 42)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Exact(4_096, 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 4_096 {
		t.Fatalf("Exact length = %d, want 4096", len(first))
	}
	if !slices.Equal(first, second) {
		t.Fatal("Exact returned different palettes for the same seed")
	}
}

func TestExactSeedChangesOrder(t *testing.T) {
	first, err := Exact(4_096, 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Exact(4_096, 2)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Equal(first, second) {
		t.Fatal("Exact returned the same palette for distinct seeds")
	}
}

func TestExactRejectsInvalidPixelCounts(t *testing.T) {
	for _, count := range []int{0, maxColors + 1} {
		if _, err := Exact(count, 1); err == nil {
			t.Fatalf("Exact(%d) unexpectedly succeeded", count)
		}
	}
}
