package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type MockBackend struct {
	metricsReceived int64
	requestsHandled int64
	bytesReceived   int64
	mu              sync.Mutex
}

type MetricsPayload struct {
	CollectorID string `json:"collector_id"`
	Hostname    string `json:"hostname"`
	Version     string `json:"version"`
	Metrics     []struct {
		Name      string            `json:"name"`
		Value     float64           `json:"value"`
		Timestamp int64             `json:"timestamp"`
		Labels    map[string]string `json:"labels"`
	} `json:"metrics"`
}

func (mb *MockBackend) handleMetricsJSON(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&mb.requestsHandled, 1)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	atomic.AddInt64(&mb.bytesReceived, int64(len(body)))

	// Parse JSON
	var payload MetricsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Count metrics
	metricCount := int64(len(payload.Metrics))
	atomic.AddInt64(&mb.metricsReceived, metricCount)

	// Log
	log.Printf("[JSON] Received %d metrics from %s (%d bytes)", metricCount, payload.CollectorID, len(body))

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"status":"accepted","metrics_received":%d}`, metricCount)
}

func (mb *MockBackend) handleMetricsBinary(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&mb.requestsHandled, 1)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	atomic.AddInt64(&mb.bytesReceived, int64(len(body)))

	// Estimate metric count from binary payload (rough estimate)
	// Binary protocol is ~40% of JSON size, so 1 metric ≈ ~200 bytes
	estimatedMetrics := int64(len(body) / 150)
	if estimatedMetrics < 1 {
		estimatedMetrics = 1
	}

	atomic.AddInt64(&mb.metricsReceived, estimatedMetrics)

	collectorID := r.Header.Get("X-Collector-ID")
	log.Printf("[Binary] Received ~%d metrics from %s (%d bytes)", estimatedMetrics, collectorID, len(body))

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"status":"accepted","metrics_received":%d}`, estimatedMetrics)
}

func (mb *MockBackend) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","version":"3.0.0"}`)
}

func (mb *MockBackend) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := struct {
		RequestsHandled int64 `json:"requests_handled"`
		MetricsReceived int64 `json:"metrics_received"`
		BytesReceived   int64 `json:"bytes_received"`
	}{
		RequestsHandled: atomic.LoadInt64(&mb.requestsHandled),
		MetricsReceived: atomic.LoadInt64(&mb.metricsReceived),
		BytesReceived:   atomic.LoadInt64(&mb.bytesReceived),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func main() {
	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	mb := &MockBackend{}

	// Routes
	http.HandleFunc("/api/v1/health", mb.handleHealth)
	http.HandleFunc("/api/v1/metrics/push", mb.handleMetricsJSON)
	http.HandleFunc("/api/v1/metrics/push/binary", mb.handleMetricsBinary)
	http.HandleFunc("/api/v1/metrics", mb.handleMetrics)

	// Periodic logging
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			totalReqs := atomic.LoadInt64(&mb.requestsHandled)
			totalMetrics := atomic.LoadInt64(&mb.metricsReceived)
			totalBytes := atomic.LoadInt64(&mb.bytesReceived)
			log.Printf("Stats: %d requests, %d metrics, %d bytes (%.2f MB/min)",
				totalReqs, totalMetrics, totalBytes, float64(totalBytes)/1024/1024)
		}
	}()

	log.Printf("Mock Backend listening on %s", port)
	log.Printf("  POST /api/v1/metrics/push - JSON metrics")
	log.Printf("  POST /api/v1/metrics/push/binary - Binary metrics")
	log.Printf("  GET  /api/v1/health - Health check")
	log.Printf("  GET  /api/v1/metrics - Stats")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
