package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp float64                `json:"timestamp"`
	Processed bool                   `json:"processed"`
}

type ProcessResult struct {
	EventID     string `json:"event_id"`
	Status      string `json:"status"`
	ProcessedAt string `json:"processed_at"`
}

var (
	results []ProcessResult
	mu      sync.Mutex
	logger  *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "[event-processor] ", log.LstdFlags)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "event-processor",
	})
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		logger.Printf("Failed to decode event: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if event.ID == "" || event.Type == "" {
		logger.Printf("Rejected event: missing id or type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "fields 'id' and 'type' are required"})
		return
	}

	result := ProcessResult{
		EventID:     event.ID,
		Status:      "processed",
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
	}

	mu.Lock()
	results = append(results, result)
	mu.Unlock()

	logger.Printf("Processed event %s (type=%s)", event.ID, event.Type)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	snapshot := make([]ProcessResult, len(results))
	copy(snapshot, results)
	mu.Unlock()

	logger.Printf("Returning %d results", len(snapshot))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func main() {
	port := os.Getenv("PROCESSOR_PORT")
	if port == "" {
		port = "5002"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/process", processHandler)
	mux.HandleFunc("/results", resultsHandler)

	logger.Printf("Starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
