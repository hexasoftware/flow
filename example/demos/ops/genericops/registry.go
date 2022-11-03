package genericops

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/registry"
)

// New create new registry with generic operations
func New() *registry.R {

	r := registry.New()
	// Test functions
	r.Add(testErrorPanic, testErrorDelayed, testRandomError).
		Tags("testing")

	registry.Describer(
		r.Add("wait", wait),
		r.Add("waitRandom", waitRandom),
	).Tags("testing").Extra("style", map[string]string{"color": "#8a5"})

	r.Add(strings.Split, strings.Join).Tags("strings")

	return r
}
func wait(data flow.Data, n int) flow.Data {
	time.Sleep(time.Duration(n) * time.Second) // Simulate
	return data
}
func waitRandom(data flow.Data) flow.Data {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return data
}

////////////////////////
// testOps
////////////////
func testErrorPanic(n int) flow.Data {
	dur := time.Duration(n)
	time.Sleep(dur * time.Second)
	panic("I panicked")
}

func testErrorDelayed(n int) (flow.Data, error) {
	dur := time.Duration(n)
	time.Sleep(dur * time.Second)
	return nil, fmt.Errorf("I got an error %v", dur)
}
func testRandomError(d flow.Data) (flow.Data, error) {
	r := rand.Intn(10)
	if r > 5 {
		return nil, errors.New("I failed on purpose")
	}
	return d, nil
}
