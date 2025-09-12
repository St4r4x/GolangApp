package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var version string = "0.0.0-local"

// Backend service URLs - these will be discovered dynamically
var backends []string
var backendIndex = 0

var pingTemplate, _ = template.New("").Parse(`<html>
<title>Load Balancer Gateway</title>
<style>
html
{
	background: linear-gradient(0.25turn, rgb(2,0,36) 0%, rgb(59,9,121) 50%, rgb(0,21,66) 100%);
	color: #fafafa;
	margin: 0;
	padding: 10px;
}
</style>
<h1>Load Balancer v{{.}}</h1>
<p>Available backends: {{len .Backends}}</p>
</html>`)

// Discover all cats-api instances
func discoverBackends() []string {
	var discoveredBackends []string
	
	// Try to resolve cats-api service to get all IPs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	resolver := &net.Resolver{}
	ips, err := resolver.LookupIPAddr(ctx, "cats-api")
	if err != nil {
		log.Printf("Error discovering backends: %v", err)
		// Fallback to single service name
		return []string{"http://cats-api:8080"}
	}
	
	for _, ip := range ips {
		backend := fmt.Sprintf("http://%s:8080", ip.IP.String())
		discoveredBackends = append(discoveredBackends, backend)
		log.Printf("Discovered backend: %s", backend)
	}
	
	if len(discoveredBackends) == 0 {
		// Fallback
		discoveredBackends = []string{"http://cats-api:8080"}
	}
	
	return discoveredBackends
}

func main() {
	// Discover backends on startup
	backends = discoverBackends()
	log.Printf("Load balancer starting with %d backends", len(backends))

	server := &http.Server{
		Addr:    ":8080",
		Handler: logReq(mainHandler()),
	}

	log.Printf("Server started, listening on %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

// High-order function to generate a log for each incoming request
func logReq(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New request from '%s', endpoint: '%s %s'", r.RemoteAddr, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func proxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := backends[backendIndex]
		backendIndex = (backendIndex + 1) % len(backends)
		target, _ := url.Parse(backend)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	}
}

func mainHandler() http.Handler {

	// For testing the proxy health. A better approach is with ExecuteTemplate form the html/template package
	pingHandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		pingTemplate.ExecuteTemplate(res, "", version)
	})

	mux := http.NewServeMux()

	// Self hosted
	mux.Handle("/ping", pingHandler)

	// Proxy to backends
	mux.Handle("/", proxyHandler())

	return mux
}
