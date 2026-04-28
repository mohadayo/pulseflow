package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", body["status"])
	}
	if body["service"] != "event-processor" {
		t.Fatalf("expected service event-processor, got %s", body["service"])
	}
}

func TestProcessHandler(t *testing.T) {
	mu.Lock()
	results = nil
	mu.Unlock()

	event := Event{ID: "test-1", Type: "click", Payload: map[string]interface{}{"x": float64(10)}}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result ProcessResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.EventID != "test-1" {
		t.Fatalf("expected event_id test-1, got %s", result.EventID)
	}
	if result.Status != "processed" {
		t.Fatalf("expected status processed, got %s", result.Status)
	}
}

func TestProcessHandlerMissingFields(t *testing.T) {
	body := []byte(`{"id":"","type":""}`)
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProcessHandlerInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProcessHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/process", nil)
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestResultsHandler(t *testing.T) {
	mu.Lock()
	results = []ProcessResult{{EventID: "r-1", Status: "processed", ProcessedAt: "2025-01-01T00:00:00Z"}}
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/results", nil)
	w := httptest.NewRecorder()
	resultsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res []ProcessResult
	json.Unmarshal(w.Body.Bytes(), &res)
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
}

func TestResultsHandlerEmpty(t *testing.T) {
	mu.Lock()
	results = nil
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/results", nil)
	w := httptest.NewRecorder()
	resultsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
