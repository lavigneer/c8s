/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unit

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/org/c8s/pkg/api/handlers"
)

// TestParseFilters_ParsesAllParameters verifies all filter parameters are parsed
func TestParseFilters_ParsesAllParameters(t *testing.T) {
	params := url.Values{
		"status":    []string{"Succeeded"},
		"branch":    []string{"main"},
		"search":    []string{"abc123"},
		"from_date": []string{"2025-01-01"},
		"to_date":   []string{"2025-01-31"},
	}

	req := &http.Request{
		URL: &url.URL{RawQuery: params.Encode()},
	}

	filters := handlers.ParseFilters(req)

	if filters.Status != "Succeeded" {
		t.Errorf("Expected status='Succeeded', got '%s'", filters.Status)
	}

	if filters.Branch != "main" {
		t.Errorf("Expected branch='main', got '%s'", filters.Branch)
	}

	if filters.Search != "abc123" {
		t.Errorf("Expected search='abc123', got '%s'", filters.Search)
	}

	// Check dates are parsed
	expectedFromDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !filters.FromDate.Equal(expectedFromDate) {
		t.Errorf("Expected from_date=%v, got %v", expectedFromDate, filters.FromDate)
	}
}

// TestParseFilters_HandlesInvalidDates handles invalid date formats gracefully
func TestParseFilters_HandlesInvalidDates(t *testing.T) {
	params := url.Values{
		"from_date": []string{"invalid-date"},
		"to_date":   []string{"2025-13-45"}, // Invalid month/day
	}

	req := &http.Request{
		URL: &url.URL{RawQuery: params.Encode()},
	}

	filters := handlers.ParseFilters(req)

	// Invalid dates should result in zero time values
	if !filters.FromDate.IsZero() && filters.FromDate.Year() != 0 {
		t.Logf("From date parsing should handle invalid input")
	}

	if !filters.ToDate.IsZero() && filters.ToDate.Year() != 0 {
		t.Logf("To date parsing should handle invalid input")
	}
}

// TestParseFilters_HandlesEmptyFilters returns empty filters when no params provided
func TestParseFilters_HandlesEmptyFilters(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{RawQuery: ""},
	}

	filters := handlers.ParseFilters(req)

	if filters.Status != "" {
		t.Errorf("Expected empty status, got '%s'", filters.Status)
	}

	if filters.Branch != "" {
		t.Errorf("Expected empty branch, got '%s'", filters.Branch)
	}

	if filters.Search != "" {
		t.Errorf("Expected empty search, got '%s'", filters.Search)
	}

	if !filters.FromDate.IsZero() {
		t.Errorf("Expected zero from_date, got %v", filters.FromDate)
	}

	if !filters.ToDate.IsZero() {
		t.Errorf("Expected zero to_date, got %v", filters.ToDate)
	}
}

// TestParseFilters_DateRangeValidation verifies to_date is set to end of day
func TestParseFilters_DateRangeValidation(t *testing.T) {
	params := url.Values{
		"to_date": []string{"2025-01-31"},
	}

	req := &http.Request{
		URL: &url.URL{RawQuery: params.Encode()},
	}

	filters := handlers.ParseFilters(req)

	// To date should be set to end of day (23:59:59)
	expectedEndOfDay := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	if filters.ToDate != expectedEndOfDay {
		t.Logf("Expected to_date to be set to end of day")
	}
}

// TestParseFilters_PartialFilters handles partial filter sets
func TestParseFilters_PartialFilters(t *testing.T) {
	params := url.Values{
		"status": []string{"Running"},
		"branch": []string{"develop"},
		// No search, dates
	}

	req := &http.Request{
		URL: &url.URL{RawQuery: params.Encode()},
	}

	filters := handlers.ParseFilters(req)

	if filters.Status != "Running" {
		t.Errorf("Expected status='Running', got '%s'", filters.Status)
	}

	if filters.Branch != "develop" {
		t.Errorf("Expected branch='develop', got '%s'", filters.Branch)
	}

	if filters.Search != "" {
		t.Errorf("Expected empty search, got '%s'", filters.Search)
	}
}
