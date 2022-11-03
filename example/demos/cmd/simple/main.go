package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/flowserver"
	"github.com/hexasoftware/flow/registry"
)

func main() {

	r := registry.New()
	r.Add("hello", func() string {
		return "hello world"
	})

	// Describing functions:

	// utility to apply functions to several entries
	registry.Describer(
		r.Add(strings.Split).Inputs("str", "sep").Output("slice"),
		r.Add(strings.Join).Inputs("slice", "sep").Output("string"),
	).Tags("strings").
		Extra("style", registry.M{"color": "#a77"})

	f := flow.New()
	f.UseRegistry(r)

	op := f.Op("Join",
		f.Op("Split", "hello world", " "),
		",",
	)
	res, err := op.Process()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("res:", res)

	http.ListenAndServe(":5000", flowserver.New(r, "storename"))

}
