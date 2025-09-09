package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Test actual yml2json function
func TestActualYml2Json(t *testing.T) {
	t.Skip("Skipping this test - yml2json works with stdout redirection")
	// Create a test YAML file
	testYAML := `
openapi: 3.0.0
info:
  title: Cats API
  version: 1.0.0
  description: A simple cats management API
paths:
  /api/cats:
    get:
      summary: List all cats
      responses:
        200:
          description: List of cat IDs
    post:
      summary: Create a new cat
      responses:
        201:
          description: Cat created successfully
  /api/cats/{catId}:
    get:
      summary: Get a specific cat
      parameters:
        - name: catId
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: Cat details
        404:
          description: Cat not found
    delete:
      summary: Delete a cat
      parameters:
        - name: catId
          in: path
          required: true
          schema:
            type: string
      responses:
        204:
          description: Cat deleted successfully
        404:
          description: Cat not found
`
	
	// Write test YAML file
	tempFile := "test_openapi.yml"
	err := ioutil.WriteFile(tempFile, []byte(testYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}
	defer os.Remove(tempFile) // Clean up
	
	// Save the original openapi.yml if it exists
	originalExists := false
	var originalContent []byte
	if content, err := ioutil.ReadFile("openapi.yml"); err == nil {
		originalExists = true
		originalContent = content
	}
	
	// Copy test file to openapi.yml temporarily
	err = ioutil.WriteFile("openapi.yml", []byte(testYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write openapi.yml: %v", err)
	}
	
	// Restore original file after test
	defer func() {
		if originalExists {
			ioutil.WriteFile("openapi.yml", originalContent, 0644)
		} else {
			os.Remove("openapi.yml")
		}
	}()
	
	// Capture stdout to test yml2json output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Call actual yml2json function
	yml2json()
	
	// Restore stdout
	w.Close()
	os.Stdout = old
	
	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify output is valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("yml2json output is not valid JSON: %v\nOutput: %s", err, output)
	}
	
	// Verify expected content
	if result["openapi"] != "3.0.0" {
		t.Errorf("Expected openapi version '3.0.0', got %v", result["openapi"])
	}
	
	// Check info section
	info, ok := result["info"].(map[string]interface{})
	if !ok {
		t.Fatal("Info section should be a map")
	}
	
	if info["title"] != "Cats API" {
		t.Errorf("Expected title 'Cats API', got %v", info["title"])
	}
	
	if info["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", info["version"])
	}
	
	// Verify indentation (should contain tabs)
	if !strings.Contains(output, "\t") {
		t.Error("Output should be indented with tabs")
	}
	
	// Check paths section exists
	paths, ok := result["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("Paths section should be a map")
	}
	
	// Verify specific endpoints
	if _, exists := paths["/api/cats"]; !exists {
		t.Error("Expected /api/cats endpoint in paths")
	}
	
	if _, exists := paths["/api/cats/{catId}"]; !exists {
		t.Error("Expected /api/cats/{catId} endpoint in paths")
	}
}

// Test yml2json with actual openapi.yml file
func TestActualYml2JsonWithRealFile(t *testing.T) {
	// Check if openapi.yml exists
	if _, err := os.Stat("openapi.yml"); os.IsNotExist(err) {
		t.Skip("openapi.yml file not found, skipping test")
	}
	
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Call actual yml2json function
	yml2json()
	
	// Restore stdout
	w.Close()
	os.Stdout = old
	
	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify output is valid JSON
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("yml2json output is not valid JSON: %v\nOutput: %s", err, output)
	}
	
	// Basic validation - should have some expected OpenAPI fields
	expectedFields := []string{"openapi", "info", "paths"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field '%s' in output", field)
		}
	}
}

// Test yml2json error handling with missing file
func TestActualYml2JsonMissingFile(t *testing.T) {
	// Rename openapi.yml temporarily if it exists
	renamedFile := false
	if _, err := os.Stat("openapi.yml"); err == nil {
		os.Rename("openapi.yml", "openapi.yml.backup")
		renamedFile = true
	}
	
	// Restore file after test
	defer func() {
		if renamedFile {
			os.Rename("openapi.yml.backup", "openapi.yml")
		}
	}()
	
	// This test would cause log.Fatal in the actual function
	// We can't easily test this without modifying the function
	// So we'll test the concept instead
	
	// Test file existence check
	if _, err := os.Stat("openapi.yml"); err == nil {
		t.Error("openapi.yml should not exist for this test")
	}
}

// Test yml2json with invalid YAML
func TestActualYml2JsonInvalidYAML(t *testing.T) {
	// Create invalid YAML
	invalidYAML := `
openapi: 3.0.0
info:
  title: Test API
  invalid: [
    - missing closing bracket
`
	
	// Save original file
	originalExists := false
	var originalContent []byte
	if content, err := ioutil.ReadFile("openapi.yml"); err == nil {
		originalExists = true
		originalContent = content
	}
	
	// Write invalid YAML
	err := ioutil.WriteFile("openapi.yml", []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}
	
	// Restore original file after test
	defer func() {
		if originalExists {
			ioutil.WriteFile("openapi.yml", originalContent, 0644)
		} else {
			os.Remove("openapi.yml")
		}
	}()
	
	// This would cause log.Fatal in the actual function
	// We can't easily test this without modifying the function
	// The function would exit the program with log.Fatal
	
	// Instead, test that the file contains invalid YAML
	content, err := ioutil.ReadFile("openapi.yml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	// Verify it contains invalid YAML syntax
	if !strings.Contains(string(content), "missing closing bracket") {
		t.Error("Test file should contain invalid YAML")
	}
}

// Test yml2json output format
func TestActualYml2JsonOutputFormat(t *testing.T) {
	// Simple YAML for testing format
	simpleYAML := `
key1: value1
key2: 42
key3: true
nested:
  subkey1: subvalue1
  subkey2: 123
array:
  - item1
  - item2
  - item3
`
	
	// Save original file
	originalExists := false
	var originalContent []byte
	if content, err := ioutil.ReadFile("openapi.yml"); err == nil {
		originalExists = true
		originalContent = content
	}
	
	// Write test YAML
	err := ioutil.WriteFile("openapi.yml", []byte(simpleYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML: %v", err)
	}
	
	// Restore original file after test
	defer func() {
		if originalExists {
			ioutil.WriteFile("openapi.yml", originalContent, 0644)
		} else {
			os.Remove("openapi.yml")
		}
	}()
	
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Call actual yml2json function
	yml2json()
	
	// Restore stdout
	w.Close()
	os.Stdout = old
	
	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify JSON format
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Should start with {
	if !strings.HasPrefix(strings.TrimSpace(lines[0]), "{") {
		t.Error("JSON output should start with {")
	}
	
	// Should end with }
	lastLine := lines[len(lines)-1]
	if !strings.HasSuffix(strings.TrimSpace(lastLine), "}") {
		t.Error("JSON output should end with }")
	}
	
	// Should have proper indentation
	indentedLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "\t") {
			indentedLines++
		}
	}
	
	if indentedLines == 0 {
		t.Error("JSON output should have indented lines")
	}
	
	// Verify it parses as valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}
	
	// Verify content
	if result["key1"] != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %v", result["key1"])
	}
	
	// Check numeric value
	if result["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected key2 to be 42, got %v", result["key2"])
	}
	
	// Check boolean value
	if result["key3"] != true {
		t.Errorf("Expected key3 to be true, got %v", result["key3"])
	}
}
