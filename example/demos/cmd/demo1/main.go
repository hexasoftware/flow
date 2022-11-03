package main

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gohxs/prettylog"
	"github.com/gohxs/webu"
	"github.com/gohxs/webu/chain"
	"github.com/hexasoftware/flow/example/demos/cmd/demo1/assets"
	"github.com/hexasoftware/flow/example/demos/ops/decodeops"
	"github.com/hexasoftware/flow/example/demos/ops/defaultops"
	"github.com/hexasoftware/flow/example/demos/ops/devops"
	"github.com/hexasoftware/flow/example/demos/ops/genericops"
	"github.com/hexasoftware/flow/example/demos/ops/ml"
	"github.com/hexasoftware/flow/example/demos/ops/stringops"
	"github.com/hexasoftware/flow/example/demos/ops/webops"
	"github.com/hexasoftware/flow/flowserver"
)

//go:generate go get github.com/gohxs/folder2go
//go:generate folder2go -handler -nobackup static assets assets

func main() {
	prettylog.Global()
	log.Println("Running version:", flowserver.Version)

	addr := ":2015"
	log.Println("Starting server  at:", addr)

	c := chain.New(webu.ChainLogger(prettylog.New("req")))

	mux := http.NewServeMux()
	mux.HandleFunc("/", assetFunc)

	defops := defaultops.New()
	defops.Merge(genericops.New())
	defops.Merge(stringops.New())

	mux.Handle("/default/", c.Build(
		http.StripPrefix("/default", flowserver.New(defops, "default")),
	))

	mlReg := ml.New()
	mlReg.Merge(genericops.New())
	mlReg.Merge(stringops.New())
	mlReg.Merge(webops.New())
	mlReg.Merge(decodeops.New())

	mux.Handle("/machinelearning/", c.Build(
		http.StripPrefix("/machinelearning", flowserver.New(mlReg, "ml")),
	))

	mux.Handle("/devops/", c.Build(
		http.StripPrefix("/devops", flowserver.New(devops.New(), "devops")),
	))

	// Serve UI here

	// Context registry
	http.ListenAndServe(addr, mux)
}

func assetFunc(w http.ResponseWriter, r *http.Request) {
	urlPath := ""

	// func that handles mux
	server := r.Context().Value(http.ServerContextKey).(*http.Server)
	mux, ok := server.Handler.(*http.ServeMux)
	if ok {
		_, handlerPath := mux.Handler(r)
		urlPath = strings.TrimPrefix(r.URL.String(), handlerPath)
	}
	if urlPath == "" { // Auto index
		urlPath = "index.html"
	}
	data, ok := assets.Data[urlPath]
	if !ok {
		http.Redirect(w, r, "/default", 302)
		// w.WriteHeader(404)
	}

	w.Header().Set("Content-type", mime.TypeByExtension(filepath.Ext(urlPath)))
	w.Write(data)
}
