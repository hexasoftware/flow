package registry_test

import (
	"testing"

	"github.com/hexasoftware/flow/internal/assert"
	"github.com/hexasoftware/flow/registry"
)

func TestNewEntryInvalid(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	_, err := registry.NewEntry(r, "string")
	a.Eq(err, registry.ErrNotAFunc, "entry is not a function")
}
func TestNewEntryValid(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	_, err := registry.NewEntry(r, func(a int) int { return 0 })
	a.Eq(err, nil, "fetching an entry")
}

func TestDescription(t *testing.T) {
	a := assert.A(t)
	r := registry.New()

	e, err := registry.NewEntry(r, func(a int) int { return 0 })
	a.Eq(err, nil, "should not fail creating new entry")
	e.Describer().Tags("a", "b")
	a.Eq(len(e.Description.Tags), 2, "should have 2 categories")
	a.Eq(len(e.Description.Inputs), 1, "Should have 2 input description")

	e2, err := registry.NewEntry(r, func(a, b int) int { return 0 })
	a.Eq(err, nil, "should not fail creating new entry")

	a.Eq(len(e2.Description.Inputs), 2, "Should have 2 input description")

	e.Describer().Inputs("input")
	a.Eq(e.Description.Inputs[0].Name, "input", "should have the input description")

	e.Describer().Inputs("input", "2", "3")
	a.Eq(len(e.Description.Inputs), 1, "should have only one input")

	e.Describer().Output("output name")
	a.Eq(e.Description.Output.Name, "output name", "output description should be the same")

	e.Describer().Extra("test", 123)
	a.Eq(e.Description.Extra["test"], 123, "extra text should be as expected")

	e.Describer().Description("test")

}

func TestEntryBatch(t *testing.T) {
	a := assert.A(t)
	r := registry.New()

	d := registry.Describer(
		r.Add(func() int { return 0 }),
		r.Add(func() int { return 0 }),
		r.Add(func() int { return 0 }),
	).Tags("test").Extra("name", 1)

	a.Eq(d.Err, nil, "should not error registering funcs")
	a.Eq(len(d.Entries()), 3, "should have 3 items")
	for _, e := range d.Entries() {
		a.Eq(e.Description.Tags[0], "test", "It should be of category test")
		a.Eq(e.Description.Extra["name"], 1, "It should contain extra")
	}

}
