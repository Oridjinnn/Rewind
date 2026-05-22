package ranking

import "testing"

func TestCosineSimilarity_UnequalLengths(t *testing.T) {
	got := Cosine([]float64{1, 2, 3}, []float64{1, 2})
	if got != 0 {
		t.Fatalf("Cosine() unequal lengths = %v, want 0", got)
	}
}

