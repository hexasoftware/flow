package ml

import "math"

func sigmoid(v float64) float64 {
	return 1 / (1 + math.Exp(-v))
}
func sigmoidPrime(v float64) float64 {
	return v * (1 - v)
}
