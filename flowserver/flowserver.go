package flowserver

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/hexasoftware/flow/flowserver/flowuiassets"
	"github.com/hexasoftware/flow/registry"

	"github.com/gohxs/prettylog"
	"github.com/gohxs/webu"
)

//go:generate go get github.com/gohxs/genversion
//go:generate genversion -package flowserver -out version.go
//

// FlowServer structure
type FlowServer struct {
	//mux            *http.ServeMux
	sessionHandler http.Handler
	staticHandler  http.Handler
}

// New creates a New flow server
func New(r *registry.R, store string) *FlowServer {
	if r == nil {
		r = registry.Global.Clone()
	}
	var sessionHandler http.Handler
	var staticHandler http.Handler

	sessionHandler = NewFlowSessionManager(r, store)

	if os.Getenv("DEBUG") == "1" {
		//log.Println("DEBUG MODE: reverse proxy localhost:8081")
		proxyURL, err := url.Parse("http://localhost:8081")
		if err != nil {
			return nil
		}

		rp := httputil.NewSingleHostReverseProxy(proxyURL)
		rp.ErrorLog = prettylog.New("rproxy")
		staticHandler = rp
	} else {
		// Check folder web?
		//staticHandler = webu.StaticHandler("web", "index.html")
		staticHandler = webu.MapHandler(flowuiassets.Data, "index.html")
		//staticHandler = flowuiassets.AssetHandleFunc
	}

	/*mux := http.NewServeMux()
	mux.Handle("/conn", sessionHandler)
	mux.Handle("/", staticHandler)*/

	return &FlowServer{
		sessionHandler: sessionHandler,
		staticHandler:  staticHandler,
	}
}

func (f *FlowServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Manual routing
	switch r.URL.Path {
	case "/conn":
		f.sessionHandler.ServeHTTP(w, r)
	default:
		f.staticHandler.ServeHTTP(w, r)
	}
}

// ListenAndServe starts the httpserver
// It will listen on default port 2015 and increase if port is in use
func (f *FlowServer) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, f)
}
