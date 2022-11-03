package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/hexasoftware/flow/flowserver"
	"github.com/hexasoftware/flow/registry"
)

func main() {
	r := registry.New()
	r.Add(strings.Split, strings.Join, toString, rand.Float64)
	http.ListenAndServe(":8080", flowserver.New(r, "mystore"))
}

func toString(a interface{}) string {
	return fmt.Sprintf("%v", a)
}
