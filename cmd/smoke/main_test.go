package main

import "testing"

func TestCheckedPixelCount(t *testing.T) {
	got, err := checkedPixelCount(1_024, 1_024)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1_048_576 {
		t.Fatalf("checkedPixelCount = %d, want 1048576", got)
	}
	if _, err := checkedPixelCount(1, 2); err == nil {
		t.Fatal("checkedPixelCount accepted a one-pixel-wide image")
	}
	if _, err := checkedPixelCount(4_097, 4_096); err == nil {
		t.Fatal("checkedPixelCount accepted more than the 24-bit RGB space")
	}
}
