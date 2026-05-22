package memory

import (
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{"Perfect match", []float64{1, 0}, []float64{1, 0}, 1.0},
		{"Orthogonal", []float64{1, 0}, []float64{0, 1}, 0.0},
		{"Opposite", []float64{1, 1}, []float64{-1, -1}, -1.0},
		{"Empty vector", []float64{}, []float64{1, 0}, 0.0},
		{"Nil vector", nil, []float64{1, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			// Use a small epsilon for float comparison
			if got < tt.expected-0.00001 || got > tt.expected+0.00001 {
				t.Errorf("CosineSimilarity() = %v, want %v", got, tt.expected)
			}
		})
	}
}