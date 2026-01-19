package main

import "testing"

func TestDownsampleBy3(t *testing.T) {
	input := []float32{0, 3, 6, 9, 12, 15}
	output := downsampleBy3(input)
	if len(output) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(output))
	}
	if output[0] != 3 {
		t.Fatalf("expected first sample 3, got %v", output[0])
	}
	if output[1] != 12 {
		t.Fatalf("expected second sample 12, got %v", output[1])
	}
}

func TestDownsampleBy3ShortInput(t *testing.T) {
	output := downsampleBy3([]float32{1, 2})
	if output != nil {
		t.Fatalf("expected nil output")
	}
}
