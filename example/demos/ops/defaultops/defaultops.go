package defaultops

import (
	"math"
	"math/rand"

	"github.com/hexasoftware/flow/registry"
)

// New create a registry
func New() *registry.R {
	r := registry.New()
	// String functions
	// Math functions
	r.Add(
		math.Abs, math.Cos, math.Sin, math.Exp, math.Exp2, math.Tanh, math.Max, math.Min,
	).Tags("math").Extra("style", registry.M{"color": "#386"})

	registry.Describer(
		r.Add(rand.Int, rand.Intn, rand.Float64),
		r.Add("Perm", func(n int) []int {
			if n > 10 { // Limiter for safety
				n = 10
			}
			return rand.Perm(n)
		}),
	).Tags("rand").Extra("style", registry.M{"color": "#486"})

	return r
}
