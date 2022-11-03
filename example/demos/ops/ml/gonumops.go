// Package ml machine learning operations for flow
package ml

import (
	"math/rand"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/registry"
	"gonum.org/v1/gonum/mat"
)

// Matrix wrapper
type Matrix = mat.Matrix

// New registry
func New() *registry.R {

	r := registry.New()

	registry.Describer(
		r.Add(matNew).Inputs("rows", "columns", "data"),
		r.Add(
			normFloat,
			matNewRand,
			matAdd,
			matSub,
			matMul,
			matMulElem,
			matScale,
			matTranspose,
			matSigmoid,
			matSigmoidPrime,
			toFloatArr,
		),
		r.Add("train", func(a, b, c, d flow.Data) []flow.Data {
			return []flow.Data{a, b, c, d}
		}).Inputs("dummy", "dummy", "dummy", "dummy"),
	).Description("gonum functions").
		Tags("gonum").
		Extra("style", registry.M{"color": "#953"})

	registry.Describer(
		r.Add(imageToGrayMatrix),
		r.Add(displayImg),
		r.Add(toGrayImage),
		r.Add(matConv).Inputs("matrix", "conv"),
		r.Add(displayGrayMat),
	).Tags("experiment")

	return r
}

func normFloat(n int) []float64 {
	data := make([]float64, n)
	for i := range data {
		data[i] = rand.NormFloat64()
	}
	return data
}

func matNew(r, c int, data []float64) Matrix {
	return mat.NewDense(r, c, data)
}
func matNewRand(r, c int) Matrix {
	data := normFloat(r * c)

	return mat.NewDense(r, c, data)

}

func matAdd(a Matrix, b Matrix) Matrix {
	var r mat.Dense
	r.Add(a, b)
	return &r
}
func matSub(a, b Matrix) Matrix {
	var r mat.Dense
	r.Sub(a, b)
	return &r
}

func matMul(a Matrix, b Matrix) Matrix {
	var r mat.Dense
	r.Mul(a, b)
	return &r
}
func matMulElem(a, b Matrix) Matrix {
	var r mat.Dense
	r.MulElem(a, b)
	return &r
}

// Scalar per element multiplication
func matScale(f float64, a Matrix) Matrix {
	var r mat.Dense
	r.Scale(f, a)
	return &r
}

func matTranspose(a Matrix) Matrix {
	return a.T()
}

// sigmoid Activator
func matSigmoid(a Matrix) Matrix {
	ret := &mat.Dense{}
	ret.Apply(func(_, _ int, v float64) float64 { return sigmoid(v) }, a)
	return ret
}

func matSigmoidPrime(a Matrix) Matrix {
	ret := &mat.Dense{}
	ret.Apply(func(_, _ int, v float64) float64 { return sigmoidPrime(v) }, a)
	return ret
}

func toFloatArr(a Matrix) []float64 {
	r, c := a.Dims()
	ret := make([]float64, r*c)
	for i := 0; i < c; i++ {
		for j := 0; j < r; j++ {
			ret[i*r+j] = a.At(j, i)
		}
	}
	return ret
}
