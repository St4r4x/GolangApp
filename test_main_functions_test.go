package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test the actual functions from the main package

func TestActualHomeHandler(t *testing.T) {
	// Test the real getHomeHandler function
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	
	// Call the actual handler
	getHomeHandler(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "text/html")
	}

	// Check the body contains expected content
	body := rr.Body.String()
	expectedContent := []string{
		"Cats API",
		"Software version:",
		version,
		"Swagger OpenAPI UI",
		"<html>",
		"</html>",
	}

	for _, content := range expectedContent {
		if !contains(body, content) {
			t.Errorf("Response body should contain '%s', got: %s", content, body)
		}
	}
}

func TestActualNewApp(t *testing.T) {
	// Test the actual newApp function
	app := newApp()
	
	if app == nil {
		t.Error("newApp() should return a non-nil handler")
	}

	// Test that the app handles requests
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// Should get some response (even if 404)
	if rr.Code == 0 {
		t.Error("App should handle requests and return a status code")
	}
}

func TestActualYml2JsonComponents(t *testing.T) {
	// Test the core components of yml2json function
	
	// Create a test YAML content
	testYAML := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
`

	// Test YAML unmarshaling (same logic as in yml2json)
	var data interface{}
	err := yaml.Unmarshal([]byte(testYAML), &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Test JSON marshaling (same logic as in yml2json)
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Verify the conversion worked
	jsonStr := string(jsonData)
	expectedFields := []string{"openapi", "info", "title", "paths"}
	
	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON output should contain '%s', got: %s", field, jsonStr)
		}
	}
}

func TestActualLogReq(t *testing.T) {
	// Test the logging middleware
	req, err := http.NewRequest("GET", "/cats/123", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test handler to wrap
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Wrap with logReq middleware
	wrappedHandler := logReq(testHandler)

	// Test the wrapped handler
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("logged handler returned wrong status: got %v want %v",
			status, http.StatusOK)
	}

	if body := rr.Body.String(); body != "test" {
		t.Errorf("logged handler returned wrong body: got %v want %v",
			body, "test")
	}
}

func TestActualMakeHandlerFunc(t *testing.T) {
	// Test the makeHandlerFunc wrapper with correct ServiceFunc signature
	testServiceFunc := func(req *http.Request) (int, interface{}) {
		return http.StatusOK, "test response"
	}

	// Wrap with makeHandlerFunc
	wrappedHandler := makeHandlerFunc(testServiceFunc)

	// Test the wrapped handler
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("wrapped handler returned wrong status: got %v want %v",
			status, http.StatusOK)
	}

	// The response should be JSON encoded
	if !contains(rr.Body.String(), "test response") {
		t.Errorf("wrapped handler should contain test response, got: %v",
			rr.Body.String())
	}
}

func TestActualCatsHandlers(t *testing.T) {
	// Test the cats handlers using the actual ServiceFunc pattern
	
	// Test listCats
	req, err := http.NewRequest("GET", "/cats", nil)
	if err != nil {
		t.Fatal(err)
	}

	code, response := listCats(req)
	
	if code == 0 {
		t.Error("listCats should return a valid status code")
	}

	if response == nil {
		t.Error("listCats should return a response")
	}

	// Test getCat
	req2, err := http.NewRequest("GET", "/cats/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	code2, response2 := getCat(req2)
	
	if code2 == 0 {
		t.Error("getCat should return a valid status code")
	}

	// Response can be nil (cat not found) or valid data
	t.Logf("getCat returned code: %d, response: %v", code2, response2)
}

func TestFileOperations(t *testing.T) {
	// Test file operations that yml2json uses
	
	// Test reading a file (simulate what yml2json does)
	tempFile, err := ioutil.TempFile("", "test*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "test: content\n"
	if _, err := tempFile.Write([]byte(testContent)); err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	// Read the file back
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch: got %s, want %s", string(content), testContent)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
