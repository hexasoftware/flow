package registry_test

import (
	"flow/registry"
	"strings"
	"testing"

	"github.com/hexasoftware/flow/internal/assert"
)

func TestMakeBatch(t *testing.T) {
	a := assert.A(t)
	r := registry.New()

	d := registry.Describer(
		r.Add(strings.Split, strings.Join),
		r.Add(strings.Compare),
		r.Add("named", strings.Compare),
	)
	a.Eq(len(d.Entries()), 4, "should have 4 entries in batch")

	d.Inputs("str")
	d.Output("result")

	for _, en := range d.Entries() {
		a.Eq(en.Description.Inputs[0].Name, "str", "first input should be string")
		a.Eq(en.Description.Output.Name, "result", "output should be equal")
	}
}

/*func TestNilDescription(t *testing.T) {
	a := assert.A(t)
	r := registry.New()

	r.Add("test", strings.Compare)
	e, _ := r.Entry("test")
	e.Description = nil

	d := registry.Describer(e)
	a.Eq(len(d.Entries()), 1, "should have 3 entries in batch")

	// Testing adding to a nil description
	d.Tags("")
	a.NotEq(d.Err, nil, "err should not be nil while adding tags")
	d.Description("")
	a.NotEq(d.Err, nil, "err should not be nil while adding a description")
	d.Inputs("str")
	a.NotEq(d.Err, nil, "err should not be nil describing inputs")
	d.Output("result")
	a.NotEq(d.Err, nil, "err should not be nil describing the output")
	d.Extra("v", "v")
	a.NotEq(d.Err, nil, "err should not be nil setting extra")

}*/
