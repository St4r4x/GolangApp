package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test actual createCat function
func TestActualCreateCat(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()
	
	// Clear database for test
	catsDatabase = make(map[string]Cat)
	
	// Create test cat
	testCat := Cat{
		Name:      "TestCat",
		Color:     "Orange",
		BirthDate: "2023-01-01",
	}
	
	jsonData, err := json.Marshal(testCat)
	if err != nil {
		t.Fatalf("Failed to marshal test cat: %v", err)
	}
	
	// Create request
	req := httptest.NewRequest("POST", "/api/cats", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Call actual function
	statusCode, response := createCat(req)
	
	// Assertions
	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}
	
	// Check response is a string (cat ID)
	responseStr, ok := response.(string)
	if !ok {
		t.Errorf("Expected string response, got %T", response)
		return
	}
	
	if responseStr == "" {
		t.Error("Expected non-empty cat ID")
	}
	
	// Check cat was saved to database
	if len(catsDatabase) != 1 {
		t.Errorf("Expected 1 cat in database, got %d", len(catsDatabase))
	}
	
	// Verify the cat in database
	savedCat, exists := catsDatabase[responseStr]
	if !exists {
		t.Error("Created cat not found in database")
		return
	}
	
	if savedCat.Name != testCat.Name {
		t.Errorf("Expected cat name %s, got %s", testCat.Name, savedCat.Name)
	}
	
	if savedCat.Color != testCat.Color {
		t.Errorf("Expected cat color %s, got %s", testCat.Color, savedCat.Color)
	}
}

// Test actual createCat function with invalid JSON
func TestActualCreateCatInvalidJSON(t *testing.T) {
	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/cats", strings.NewReader("{ invalid json }"))
	req.Header.Set("Content-Type", "application/json")
	
	// Call actual function
	statusCode, response := createCat(req)
	
	// Assertions
	if statusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, statusCode)
	}
	
	if response != "Invalid JSON input" {
		t.Errorf("Expected 'Invalid JSON input', got %v", response)
	}
}

// Test actual deleteCat function with existing cat
func TestActualDeleteCatExists(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()
	
	// Set up test cat in database
	testCatID := "test-cat-id-123"
	testCat := Cat{
		Name: "TestCat",
		ID:   testCatID,
	}
	catsDatabase = map[string]Cat{
		testCatID: testCat,
	}
	
	// Create request with path parameter
	req := httptest.NewRequest("DELETE", "/api/cats/"+testCatID, nil)
	req.SetPathValue("catId", testCatID)
	
	// Call actual function
	statusCode, response := deleteCat(req)
	
	// Assertions
	if statusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, statusCode)
	}
	
	if response != nil {
		t.Errorf("Expected nil response, got %v", response)
	}
	
	// Check cat was deleted from database
	if _, exists := catsDatabase[testCatID]; exists {
		t.Error("Cat should have been deleted from database")
	}
	
	if len(catsDatabase) != 0 {
		t.Errorf("Expected empty database, got %d items", len(catsDatabase))
	}
}

// Test actual deleteCat function with non-existent cat
func TestActualDeleteCatNotExists(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()
	
	// Clear database
	catsDatabase = make(map[string]Cat)
	
	nonExistentID := "non-existent-cat-id"
	
	// Create request
	req := httptest.NewRequest("DELETE", "/api/cats/"+nonExistentID, nil)
	req.SetPathValue("catId", nonExistentID)
	
	// Call actual function
	statusCode, response := deleteCat(req)
	
	// Assertions
	if statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, statusCode)
	}
	
	if response != "Cat not found" {
		t.Errorf("Expected 'Cat not found', got %v", response)
	}
}

// Test createCat with empty request body
func TestActualCreateCatEmptyBody(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()
	
	// Clear database
	catsDatabase = make(map[string]Cat)
	
	// Create request with empty body
	req := httptest.NewRequest("POST", "/api/cats", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	
	// Call actual function
	statusCode, response := createCat(req)
	
	// Should still create a cat with empty fields
	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}
	
	// Check response is a string (cat ID)
	responseStr, ok := response.(string)
	if !ok {
		t.Errorf("Expected string response, got %T", response)
		return
	}
	
	if responseStr == "" {
		t.Error("Expected non-empty cat ID")
	}
	
	// Check cat was saved
	if len(catsDatabase) != 1 {
		t.Errorf("Expected 1 cat in database, got %d", len(catsDatabase))
	}
}

// Test createCat with partial data
func TestActualCreateCatPartialData(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()
	
	// Clear database
	catsDatabase = make(map[string]Cat)
	
	// Create test cat with only name
	testData := `{"name": "PartialCat"}`
	
	// Create request
	req := httptest.NewRequest("POST", "/api/cats", strings.NewReader(testData))
	req.Header.Set("Content-Type", "application/json")
	
	// Call actual function
	statusCode, response := createCat(req)
	
	// Assertions
	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}
	
	responseStr, ok := response.(string)
	if !ok {
		t.Errorf("Expected string response, got %T", response)
		return
	}
	
	// Check saved cat
	savedCat, exists := catsDatabase[responseStr]
	if !exists {
		t.Error("Created cat not found in database")
		return
	}
	
	if savedCat.Name != "PartialCat" {
		t.Errorf("Expected cat name 'PartialCat', got %s", savedCat.Name)
	}
	
	// Other fields should be empty
	if savedCat.Color != "" {
		t.Errorf("Expected empty color, got %s", savedCat.Color)
	}
	
	if savedCat.BirthDate != "" {
		t.Errorf("Expected empty birthDate, got %s", savedCat.BirthDate)
	}
}

// Test multiple operations
func TestActualCRUDOperations(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()
	
	// Clear database
	catsDatabase = make(map[string]Cat)
	
	// Create cat
	testCat := Cat{
		Name:      "CRUDCat",
		Color:     "Blue",
		BirthDate: "2023-01-01",
	}
	
	jsonData, _ := json.Marshal(testCat)
	createReq := httptest.NewRequest("POST", "/api/cats", bytes.NewBuffer(jsonData))
	createReq.Header.Set("Content-Type", "application/json")
	
	statusCode, response := createCat(createReq)
	if statusCode != http.StatusCreated {
		t.Fatalf("Failed to create cat: status %d", statusCode)
	}
	
	catID := response.(string)
	
	// Verify cat exists with getCat
	getReq := httptest.NewRequest("GET", "/api/cats/"+catID, nil)
	getReq.SetPathValue("catId", catID)
	
	statusCode, response = getCat(getReq)
	if statusCode != http.StatusOK {
		t.Errorf("Failed to get cat: status %d", statusCode)
	}
	
	// Delete cat
	deleteReq := httptest.NewRequest("DELETE", "/api/cats/"+catID, nil)
	deleteReq.SetPathValue("catId", catID)
	
	statusCode, response = deleteCat(deleteReq)
	if statusCode != http.StatusNoContent {
		t.Errorf("Failed to delete cat: status %d", statusCode)
	}
	
	// Verify cat is gone
	getReq2 := httptest.NewRequest("GET", "/api/cats/"+catID, nil)
	getReq2.SetPathValue("catId", catID)
	
	statusCode, response = getCat(getReq2)
	if statusCode != http.StatusNotFound {
		t.Errorf("Expected cat to be deleted, got status %d", statusCode)
	}
}
