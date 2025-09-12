package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var version string = "0.0.0-local"

// LoadBalancingStrategy defines the strategy enum
type LoadBalancingStrategy int

const (
	RoundRobin LoadBalancingStrategy = iota
	Random
	WeightedRoundRobin
	LeastConnections
	IPHash
)

// Backend represents a backend server with metadata
type Backend struct {
	URL             string
	Weight          int
	ActiveRequests  int
	TotalRequests   int64
	LastRequestTime time.Time
	Healthy         bool
	mutex           sync.RWMutex
}

// LoadBalancer manages backends and load balancing strategies
type LoadBalancer struct {
	backends         []*Backend
	strategy         LoadBalancingStrategy
	roundRobinIndex  int
	weightedIndex    int
	mutex            sync.RWMutex
}

// NewLoadBalancer creates a new load balancer with specified strategy
func NewLoadBalancer(strategy LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		backends: make([]*Backend, 0),
		strategy: strategy,
	}
}

// AddBackend adds a backend to the load balancer
func (lb *LoadBalancer) AddBackend(url string, weight int) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	backend := &Backend{
		URL:     url,
		Weight:  weight,
		Healthy: true,
	}
	lb.backends = append(lb.backends, backend)
	log.Printf("Added backend: %s (weight: %d)", url, weight)
}

// GetBackend returns the next backend based on the load balancing strategy
func (lb *LoadBalancer) GetBackend(clientIP string) *Backend {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	healthyBackends := lb.getHealthyBackends()
	if len(healthyBackends) == 0 {
		return nil
	}
	
	switch lb.strategy {
	case RoundRobin:
		return lb.roundRobinSelect(healthyBackends)
	case Random:
		return lb.randomSelect(healthyBackends)
	case WeightedRoundRobin:
		return lb.weightedRoundRobinSelect(healthyBackends)
	case LeastConnections:
		return lb.leastConnectionsSelect(healthyBackends)
	case IPHash:
		return lb.ipHashSelect(healthyBackends, clientIP)
	default:
		return lb.roundRobinSelect(healthyBackends)
	}
}

// getHealthyBackends returns only healthy backends
func (lb *LoadBalancer) getHealthyBackends() []*Backend {
	var healthy []*Backend
	for _, backend := range lb.backends {
		backend.mutex.RLock()
		if backend.Healthy {
			healthy = append(healthy, backend)
		}
		backend.mutex.RUnlock()
	}
	return healthy
}

// Round Robin implementation
func (lb *LoadBalancer) roundRobinSelect(backends []*Backend) *Backend {
	if len(backends) == 0 {
		return nil
	}
	backend := backends[lb.roundRobinIndex%len(backends)]
	lb.roundRobinIndex++
	return backend
}

// Random implementation
func (lb *LoadBalancer) randomSelect(backends []*Backend) *Backend {
	if len(backends) == 0 {
		return nil
	}
	return backends[rand.Intn(len(backends))]
}

// Weighted Round Robin implementation
func (lb *LoadBalancer) weightedRoundRobinSelect(backends []*Backend) *Backend {
	if len(backends) == 0 {
		return nil
	}
	
	// Calculate total weight
	totalWeight := 0
	for _, backend := range backends {
		backend.mutex.RLock()
		totalWeight += backend.Weight
		backend.mutex.RUnlock()
	}
	
	if totalWeight == 0 {
		return lb.roundRobinSelect(backends)
	}
	
	// Find backend based on weighted distribution
	lb.weightedIndex = (lb.weightedIndex + 1) % totalWeight
	currentWeight := 0
	
	for _, backend := range backends {
		backend.mutex.RLock()
		currentWeight += backend.Weight
		if lb.weightedIndex < currentWeight {
			backend.mutex.RUnlock()
			return backend
		}
		backend.mutex.RUnlock()
	}
	
	return backends[0] // Fallback
}

// Least Connections implementation
func (lb *LoadBalancer) leastConnectionsSelect(backends []*Backend) *Backend {
	if len(backends) == 0 {
		return nil
	}
	
	var selected *Backend
	minConnections := int(^uint(0) >> 1) // Max int value
	
	for _, backend := range backends {
		backend.mutex.RLock()
		if backend.ActiveRequests < minConnections {
			minConnections = backend.ActiveRequests
			selected = backend
		}
		backend.mutex.RUnlock()
	}
	
	return selected
}

// IP Hash implementation (consistent hashing)
func (lb *LoadBalancer) ipHashSelect(backends []*Backend, clientIP string) *Backend {
	if len(backends) == 0 {
		return nil
	}
	
	hasher := fnv.New32a()
	hasher.Write([]byte(clientIP))
	hash := hasher.Sum32()
	
	return backends[int(hash)%len(backends)]
}

// IncrementConnections increments active request count for a backend
func (b *Backend) IncrementConnections() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.ActiveRequests++
	b.TotalRequests++
	b.LastRequestTime = time.Now()
}

// DecrementConnections decrements active request count for a backend
func (b *Backend) DecrementConnections() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if b.ActiveRequests > 0 {
		b.ActiveRequests--
	}
}

// GetStats returns backend statistics
func (b *Backend) GetStats() (int, int64, time.Time) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.ActiveRequests, b.TotalRequests, b.LastRequestTime
}

// Global load balancer instance
var loadBalancer *LoadBalancer

var pingTemplate, _ = template.New("").Parse(`<html>
<title>Advanced Load Balancer Gateway</title>
<style>
html {
	background: linear-gradient(0.25turn, rgb(2,0,36) 0%, rgb(59,9,121) 50%, rgb(0,21,66) 100%);
	color: #fafafa;
	margin: 0;
	padding: 10px;
	font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}
.stats { background: rgba(255,255,255,0.1); padding: 15px; margin: 10px 0; border-radius: 8px; }
.backend { margin: 5px 0; padding: 8px; background: rgba(255,255,255,0.05); border-radius: 4px; }
</style>
<h1>🔄 Advanced Load Balancer v{{.Version}}</h1>
<div class="stats">
	<h3>📊 Load Balancing Strategy: {{.Strategy}}</h3>
	<p><strong>Active Backends:</strong> {{len .Backends}}</p>
	<p><strong>Total Requests:</strong> {{.TotalRequests}}</p>
</div>
<div class="stats">
	<h3>🖥️ Backend Status:</h3>
	{{range .Backends}}
	<div class="backend">
		<strong>{{.URL}}</strong> 
		(Weight: {{.Weight}}, Active: {{.ActiveRequests}}, Total: {{.TotalRequests}})
		{{if .Healthy}}✅{{else}}❌{{end}}
	</div>
	{{end}}
</div>
</html>`)

// TemplateData for the ping template
type TemplateData struct {
	Version       string
	Strategy      string
	Backends      []*Backend
	TotalRequests int64
}

// GetStrategyName returns human-readable strategy name
func GetStrategyName(strategy LoadBalancingStrategy) string {
	switch strategy {
	case RoundRobin:
		return "Round Robin"
	case Random:
		return "Random"
	case WeightedRoundRobin:
		return "Weighted Round Robin"
	case LeastConnections:
		return "Least Connections"
	case IPHash:
		return "IP Hash"
	default:
		return "Unknown"
	}
}

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

// parseStrategy converts string to LoadBalancingStrategy
func parseStrategy(strategyStr string) LoadBalancingStrategy {
	switch strings.ToLower(strings.TrimSpace(strategyStr)) {
	case "roundrobin", "round-robin", "rr":
		return RoundRobin
	case "random", "rand":
		return Random
	case "weighted", "weightedroundrobin", "weighted-round-robin", "wrr":
		return WeightedRoundRobin
	case "leastconnections", "least-connections", "lc":
		return LeastConnections
	case "iphash", "ip-hash", "hash":
		return IPHash
	default:
		log.Printf("Unknown strategy '%s', defaulting to Round Robin", strategyStr)
		return RoundRobin
	}
}

// getConfiguredStrategy returns the strategy from command line or environment variable
func getConfiguredStrategy() LoadBalancingStrategy {
	// Check command line flag first
	var strategyFlag = flag.String("strategy", "", "Load balancing strategy (roundrobin, random, weighted, leastconnections, iphash)")
	flag.Parse()
	
	if *strategyFlag != "" {
		return parseStrategy(*strategyFlag)
	}
	
	// Check environment variable
	if envStrategy := os.Getenv("LB_STRATEGY"); envStrategy != "" {
		return parseStrategy(envStrategy)
	}
	
	// Default to Round Robin
	return RoundRobin
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Get configured load balancing strategy
	strategy := getConfiguredStrategy()
	loadBalancer = NewLoadBalancer(strategy)
	
	// Discover backends and add them to load balancer
	discoveredBackends := discoverBackends()
	for _, backend := range discoveredBackends {
		// Default weight of 1, can be made configurable
		loadBalancer.AddBackend(backend, 1)
	}
	
	log.Printf("Load balancer starting with %d backends using %s strategy", 
		len(discoveredBackends), GetStrategyName(strategy))

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

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func proxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		backend := loadBalancer.GetBackend(clientIP)
		
		if backend == nil {
			http.Error(w, "No healthy backends available", http.StatusServiceUnavailable)
			return
		}
		
		// Track active connections
		backend.IncrementConnections()
		defer backend.DecrementConnections()
		
		target, err := url.Parse(backend.URL)
		if err != nil {
			http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
			return
		}
		
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Add error handling
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error for backend %s: %v", backend.URL, err)
			backend.mutex.Lock()
			backend.Healthy = false
			backend.mutex.Unlock()
			http.Error(w, "Backend temporarily unavailable", http.StatusBadGateway)
		}
		
		proxy.ServeHTTP(w, r)
	}
}

func mainHandler() http.Handler {
	// Health/status endpoint with enhanced information
	pingHandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		loadBalancer.mutex.RLock()
		var totalRequests int64
		for _, backend := range loadBalancer.backends {
			_, total, _ := backend.GetStats()
			totalRequests += total
		}
		
		data := TemplateData{
			Version:       version,
			Strategy:      GetStrategyName(loadBalancer.strategy),
			Backends:      loadBalancer.backends,
			TotalRequests: totalRequests,
		}
		loadBalancer.mutex.RUnlock()
		
		pingTemplate.ExecuteTemplate(res, "", data)
	})

	mux := http.NewServeMux()

	// Self hosted endpoints
	mux.Handle("/ping", pingHandler)
	mux.Handle("/health", pingHandler) // Alias for health checks

	// Proxy to backends
	mux.Handle("/", proxyHandler())

	return mux
}
