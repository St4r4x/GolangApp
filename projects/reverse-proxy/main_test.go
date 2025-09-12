package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadBalancer_AddBackend(t *testing.T) {
	lb := NewLoadBalancer(RoundRobin)

	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 2)

	if len(lb.backends) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(lb.backends))
	}

	if lb.backends[0].URL != "http://backend1:8080" {
		t.Errorf("Expected first backend URL to be 'http://backend1:8080', got '%s'", lb.backends[0].URL)
	}

	if lb.backends[1].Weight != 2 {
		t.Errorf("Expected second backend weight to be 2, got %d", lb.backends[1].Weight)
	}
}

func TestLoadBalancer_RoundRobin(t *testing.T) {
	lb := NewLoadBalancer(RoundRobin)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 1)
	lb.AddBackend("http://backend3:8080", 1)

	// Test round robin distribution
	backends := make(map[string]int)
	for i := 0; i < 6; i++ {
		backend := lb.GetBackend("127.0.0.1")
		if backend != nil {
			backends[backend.URL]++
		}
	}

	// Each backend should be selected exactly twice
	expectedCount := 2
	for url, count := range backends {
		if count != expectedCount {
			t.Errorf("Backend %s was selected %d times, expected %d", url, count, expectedCount)
		}
	}
}

func TestLoadBalancer_Random(t *testing.T) {
	lb := NewLoadBalancer(Random)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 1)
	lb.AddBackend("http://backend3:8080", 1)

	// Test that random selection works (statistical test)
	backends := make(map[string]int)
	requests := 300

	for i := 0; i < requests; i++ {
		backend := lb.GetBackend("127.0.0.1")
		if backend != nil {
			backends[backend.URL]++
		}
	}

	// Check that all backends were selected at least once
	if len(backends) != 3 {
		t.Errorf("Expected all 3 backends to be selected, got %d", len(backends))
	}

	// Check that distribution is somewhat even (within 40% variance)
	expectedAvg := requests / 3
	tolerance := expectedAvg * 40 / 100 // 40% tolerance

	for url, count := range backends {
		if count < expectedAvg-tolerance || count > expectedAvg+tolerance {
			t.Logf("Warning: Backend %s selection count %d is outside expected range %d±%d",
				url, count, expectedAvg, tolerance)
		}
	}
}

func TestLoadBalancer_WeightedRoundRobin(t *testing.T) {
	lb := NewLoadBalancer(WeightedRoundRobin)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 3)
	lb.AddBackend("http://backend3:8080", 2)

	// Test weighted distribution over 12 requests (total weight = 6, so 2 cycles)
	backends := make(map[string]int)
	for i := 0; i < 12; i++ {
		backend := lb.GetBackend("127.0.0.1")
		if backend != nil {
			backends[backend.URL]++
		}
	}

	// Check distribution matches weights
	expected := map[string]int{
		"http://backend1:8080": 2, // weight 1, 2 cycles = 2
		"http://backend2:8080": 6, // weight 3, 2 cycles = 6
		"http://backend3:8080": 4, // weight 2, 2 cycles = 4
	}

	for url, expectedCount := range expected {
		if backends[url] != expectedCount {
			t.Errorf("Backend %s was selected %d times, expected %d", url, backends[url], expectedCount)
		}
	}
}

func TestLoadBalancer_LeastConnections(t *testing.T) {
	lb := NewLoadBalancer(LeastConnections)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 1)
	lb.AddBackend("http://backend3:8080", 1)

	// Simulate different connection loads
	lb.backends[0].ActiveRequests = 5
	lb.backends[1].ActiveRequests = 2
	lb.backends[2].ActiveRequests = 8

	// Should select backend with least connections (backend2 with 2 connections)
	backend := lb.GetBackend("127.0.0.1")
	if backend == nil || backend.URL != "http://backend2:8080" {
		t.Errorf("Expected backend2 to be selected (least connections), got %v", backend)
	}

	// Test with all backends having equal connections
	for _, b := range lb.backends {
		b.ActiveRequests = 3
	}

	// Should still return a valid backend
	backend = lb.GetBackend("127.0.0.1")
	if backend == nil {
		t.Error("Expected a backend to be selected when all have equal connections")
	}
}

func TestLoadBalancer_IPHash(t *testing.T) {
	lb := NewLoadBalancer(IPHash)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 1)
	lb.AddBackend("http://backend3:8080", 1)

	// Test that same IP consistently maps to same backend
	ip1 := "192.168.1.10"
	ip2 := "192.168.1.20"

	backend1_ip1 := lb.GetBackend(ip1)
	backend2_ip1 := lb.GetBackend(ip1)
	_ = lb.GetBackend(ip2) // Different IP for distribution testing

	// Same IP should always get same backend
	if backend1_ip1.URL != backend2_ip1.URL {
		t.Errorf("Same IP should get same backend consistently")
	}

	// Test multiple IPs to ensure distribution
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
	backends := make(map[string]string)

	for _, ip := range ips {
		backend := lb.GetBackend(ip)
		backends[ip] = backend.URL
	}

	// Verify consistency - same IP should always get same backend
	for _, ip := range ips {
		backend := lb.GetBackend(ip)
		if backends[ip] != backend.URL {
			t.Errorf("IP %s got different backend on second call", ip)
		}
	}
}

func TestBackend_ConnectionTracking(t *testing.T) {
	backend := &Backend{
		URL:     "http://backend1:8080",
		Weight:  1,
		Healthy: true,
	}

	// Test initial state
	active, total, _ := backend.GetStats()
	if active != 0 || total != 0 {
		t.Errorf("Expected initial stats to be 0, got active=%d, total=%d", active, total)
	}

	// Test increment
	backend.IncrementConnections()
	active, total, lastRequest := backend.GetStats()
	if active != 1 || total != 1 {
		t.Errorf("Expected stats after increment to be 1, got active=%d, total=%d", active, total)
	}

	// Check that timestamp was updated
	if lastRequest.IsZero() {
		t.Error("Expected LastRequestTime to be updated")
	}

	// Test multiple increments
	backend.IncrementConnections()
	backend.IncrementConnections()
	active, total, _ = backend.GetStats()
	if active != 3 || total != 3 {
		t.Errorf("Expected stats after 3 increments to be 3, got active=%d, total=%d", active, total)
	}

	// Test decrement
	backend.DecrementConnections()
	active, total, _ = backend.GetStats()
	if active != 2 || total != 3 {
		t.Errorf("Expected stats after decrement to be active=2, total=3, got active=%d, total=%d", active, total)
	}

	// Test decrement doesn't go below zero
	for i := 0; i < 5; i++ {
		backend.DecrementConnections()
	}
	active, _, _ = backend.GetStats()
	if active != 0 {
		t.Errorf("Expected active connections to not go below 0, got %d", active)
	}
}

func TestLoadBalancer_HealthyBackends(t *testing.T) {
	lb := NewLoadBalancer(RoundRobin)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 1)
	lb.AddBackend("http://backend3:8080", 1)

	// Mark one backend as unhealthy
	lb.backends[1].Healthy = false

	// Should only return healthy backends
	healthyBackends := lb.getHealthyBackends()
	if len(healthyBackends) != 2 {
		t.Errorf("Expected 2 healthy backends, got %d", len(healthyBackends))
	}

	// GetBackend should skip unhealthy backends
	backends := make(map[string]int)
	for i := 0; i < 10; i++ {
		backend := lb.GetBackend("127.0.0.1")
		if backend != nil {
			backends[backend.URL]++
		}
	}

	// Should not select the unhealthy backend
	if _, exists := backends["http://backend2:8080"]; exists {
		t.Error("Unhealthy backend should not be selected")
	}

	// Should distribute between healthy backends
	if len(backends) != 2 {
		t.Errorf("Expected 2 backends to be selected, got %d", len(backends))
	}
}

func TestLoadBalancer_NoHealthyBackends(t *testing.T) {
	lb := NewLoadBalancer(RoundRobin)
	lb.AddBackend("http://backend1:8080", 1)
	lb.AddBackend("http://backend2:8080", 1)

	// Mark all backends as unhealthy
	for _, backend := range lb.backends {
		backend.Healthy = false
	}

	// Should return nil when no healthy backends
	backend := lb.GetBackend("127.0.0.1")
	if backend != nil {
		t.Error("Expected nil when no healthy backends available")
	}
}

func TestGetStrategyName(t *testing.T) {
	tests := []struct {
		strategy LoadBalancingStrategy
		expected string
	}{
		{RoundRobin, "Round Robin"},
		{Random, "Random"},
		{WeightedRoundRobin, "Weighted Round Robin"},
		{LeastConnections, "Least Connections"},
		{IPHash, "IP Hash"},
		{LoadBalancingStrategy(999), "Unknown"},
	}

	for _, test := range tests {
		result := GetStrategyName(test.strategy)
		if result != test.expected {
			t.Errorf("GetStrategyName(%v) = %s, expected %s", test.strategy, result, test.expected)
		}
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expectedIP   string
	}{
		{
			name: "X-Forwarded-For header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-Forwarded-For", "192.168.1.100")
				req.RemoteAddr = "10.0.0.1:12345"
				return req
			},
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Real-IP header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-Real-IP", "192.168.1.200")
				req.RemoteAddr = "10.0.0.1:12345"
				return req
			},
			expectedIP: "192.168.1.200",
		},
		{
			name: "RemoteAddr fallback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "10.0.0.1:12345"
				return req
			},
			expectedIP: "10.0.0.1:12345",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := test.setupRequest()
			ip := getClientIP(req)
			if ip != test.expectedIP {
				t.Errorf("getClientIP() = %s, expected %s", ip, test.expectedIP)
			}
		})
	}
}

// Benchmark tests
func BenchmarkLoadBalancer_RoundRobin(b *testing.B) {
	lb := NewLoadBalancer(RoundRobin)
	for i := 0; i < 10; i++ {
		lb.AddBackend(fmt.Sprintf("http://backend%d:8080", i), 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetBackend("127.0.0.1")
	}
}

func BenchmarkLoadBalancer_Random(b *testing.B) {
	lb := NewLoadBalancer(Random)
	for i := 0; i < 10; i++ {
		lb.AddBackend(fmt.Sprintf("http://backend%d:8080", i), 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetBackend("127.0.0.1")
	}
}

func BenchmarkLoadBalancer_LeastConnections(b *testing.B) {
	lb := NewLoadBalancer(LeastConnections)
	for i := 0; i < 10; i++ {
		lb.AddBackend(fmt.Sprintf("http://backend%d:8080", i), 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetBackend("127.0.0.1")
	}
}

func BenchmarkLoadBalancer_IPHash(b *testing.B) {
	lb := NewLoadBalancer(IPHash)
	for i := 0; i < 10; i++ {
		lb.AddBackend(fmt.Sprintf("http://backend%d:8080", i), 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetBackend(fmt.Sprintf("192.168.1.%d", i%256))
	}
}

func BenchmarkLoadBalancer_WeightedRoundRobin(b *testing.B) {
	lb := NewLoadBalancer(WeightedRoundRobin)
	for i := 0; i < 10; i++ {
		lb.AddBackend(fmt.Sprintf("http://backend%d:8080", i), i+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetBackend("127.0.0.1")
	}
}
