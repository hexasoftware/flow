package stringops

import (
	"fmt"
	"strings"

	"github.com/hexasoftware/flow/registry"
)

// New create string operations
func New() *registry.R {
	r := registry.New()
	registry.Describer(
		r.Add(strings.Split).Inputs("string", "separator"),
		r.Add(strings.Join).Inputs("", "sep"),
		r.Add(strings.Compare, strings.Contains),
		r.Add("Cat", func(a, b string) string { return a + " " + b }),
		r.Add("ToString", func(a interface{}) string { return fmt.Sprint(a) }),
	).Tags("string").Extra("style", registry.M{"color": "#839"})

	return r
}
