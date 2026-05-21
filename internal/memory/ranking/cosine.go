package ranking

import "math"

func Cosine(a, b []float64) float64 {
	// Bounds check to prevent panic and handle empty vectors
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, magA, magB float64

	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}
