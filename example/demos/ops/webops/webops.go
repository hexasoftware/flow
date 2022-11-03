package webops

import (
	"io/ioutil"
	"net/http"

	"github.com/hexasoftware/flow/registry"
)

// New creates web operations for flow
func New() *registry.R {
	r := registry.New()

	r.Add(httpGet).Tags("http").
		Extra("style", registry.M{"color": "#828"})

	return r
}

func httpGet(url string) ([]byte, error) {

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}
