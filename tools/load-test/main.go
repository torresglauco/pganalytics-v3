package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Load test configuration
type LoadTestConfig struct {
	NumCollectors      int
	Duration           time.Duration
	Interval           time.Duration
	MetricsPerCollector int
	BackendURL         string
	Protocol           string // "json" or "binary"
	TLSVerify          bool
	PrometheusPort     int
}

// Metrics to track
type LoadTestMetrics struct {
	// Prometheus metrics
	collectionsDone    prometheus.Counter
	collectionsErrors  prometheus.Counter
	collectionTime     prometheus.Histogram
	ingestionTime      prometheus.Histogram
	metricsGenerated   prometheus.Counter
	metricsSent        prometheus.Counter
	metricsErrors      prometheus.Counter
	bytesSent          prometheus.Counter
	bytesSavedBinary   prometheus.Counter

	// In-memory tracking
	successCount    int64
	errorCount      int64
	totalMetrics    int64
	totalBytesJSON  int64
	totalBytesBinary int64
	startTime       time.Time
	mu              sync.Mutex
}

// SimulatedCollector represents a single collector instance
type SimulatedCollector struct {
	ID                string
	BackendURL        string
	Protocol          string
	Interval          time.Duration
	MetricsPerRound   int
	metrics           *LoadTestMetrics
	httpClient        *http.Client
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// Metric represents a single metric data point
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Timestamp int64             `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

// MetricsPayload represents the metrics submission payload
type MetricsPayload struct {
	CollectorID string    `json:"collector_id"`
	Hostname    string    `json:"hostname"`
	Version     string    `json:"version"`
	Metrics     []Metric  `json:"metrics"`
}

func main() {
	// Parse flags
	config := parseFlags()

	// Initialize metrics
	metrics := initMetrics()
	metrics.startTime = time.Now()

	// Create HTTP client with TLS configuration
	httpClient := createHTTPClient(config.TLSVerify)

	// Start Prometheus metrics server
	go startPrometheusServer(config.PrometheusPort)

	log.Printf("Starting load test with %d collectors (protocol: %s)", config.NumCollectors, config.Protocol)
	log.Printf("Duration: %v, Interval: %v, Metrics per collector: %d", config.Duration, config.Interval, config.MetricsPerCollector)
	log.Printf("Backend: %s", config.BackendURL)
	log.Printf("Prometheus metrics available at http://localhost:%d/metrics", config.PrometheusPort)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+30*time.Second)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Spawn collectors
	collectors := make([]*SimulatedCollector, config.NumCollectors)
	var wg sync.WaitGroup

	for i := 0; i < config.NumCollectors; i++ {
		collector := &SimulatedCollector{
			ID:              fmt.Sprintf("collector-%03d", i+1),
			BackendURL:      config.BackendURL,
			Protocol:        config.Protocol,
			Interval:        config.Interval,
			MetricsPerRound: config.MetricsPerCollector,
			metrics:         metrics,
			httpClient:      httpClient,
		}
		collector.ctx, collector.cancel = context.WithCancel(ctx)
		collectors[i] = collector

		wg.Add(1)
		go func(c *SimulatedCollector) {
			defer wg.Done()
			c.run()
		}(collector)
	}

	// Wait for completion or signal
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Load test completed")
	case <-sigChan:
		log.Println("Shutdown signal received, stopping collectors...")
		for _, c := range collectors {
			c.cancel()
		}
		wg.Wait()
	}

	// Print final report
	printReport(metrics, config)
}

func parseFlags() LoadTestConfig {
	config := LoadTestConfig{}

	flag.IntVar(&config.NumCollectors, "collectors", 10, "Number of simulated collectors")
	flag.DurationVar(&config.Duration, "duration", 15*time.Minute, "Test duration")
	flag.DurationVar(&config.Interval, "interval", 60*time.Second, "Collection interval per collector")
	flag.IntVar(&config.MetricsPerCollector, "metrics", 50, "Metrics per collector per interval")
	flag.StringVar(&config.BackendURL, "backend", "https://localhost:8080", "Backend API URL")
	flag.StringVar(&config.Protocol, "protocol", "json", "Protocol to use (json or binary)")
	flag.BoolVar(&config.TLSVerify, "tls-verify", false, "Verify TLS certificates")
	flag.IntVar(&config.PrometheusPort, "prometheus-port", 9090, "Prometheus metrics port")

	flag.Parse()

	return config
}

func initMetrics() *LoadTestMetrics {
	return &LoadTestMetrics{
		collectionsDone: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_collections_done",
			Help: "Total collections completed",
		}),
		collectionsErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_collections_errors",
			Help: "Total collection errors",
		}),
		collectionTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "load_test_collection_time_ms",
			Help:    "Time to collect and send metrics (milliseconds)",
			Buckets: []float64{10, 50, 100, 500, 1000, 2000, 5000},
		}),
		ingestionTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "load_test_ingestion_time_ms",
			Help:    "Backend ingestion time (milliseconds)",
			Buckets: []float64{10, 50, 100, 500, 1000, 2000, 5000},
		}),
		metricsGenerated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_metrics_generated",
			Help: "Total metrics generated",
		}),
		metricsSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_metrics_sent",
			Help: "Total metrics sent successfully",
		}),
		metricsErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_metrics_errors",
			Help: "Total metrics send errors",
		}),
		bytesSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_bytes_sent",
			Help: "Total bytes sent to backend",
		}),
		bytesSavedBinary: promauto.NewCounter(prometheus.CounterOpts{
			Name: "load_test_bytes_saved_binary",
			Help: "Bytes saved using binary protocol vs JSON",
		}),
	}
}

func createHTTPClient(tlsVerify bool) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !tlsVerify,
		},
	}

	return &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}
}

func startPrometheusServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting Prometheus metrics server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Prometheus server error: %v", err)
	}
}

// Run executes the collector's periodic collection
func (c *SimulatedCollector) run() {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// Collect generates and sends metrics
func (c *SimulatedCollector) collect() {
	start := time.Now()

	// Generate synthetic metrics
	payload := c.generateMetrics()

	// Send metrics
	var err error
	var bytesUsed int

	if c.Protocol == "binary" {
		bytesUsed, err = c.sendBinary(payload)
	} else {
		bytesUsed, err = c.sendJSON(payload)
	}

	elapsed := time.Since(start).Milliseconds()
	c.metrics.collectionTime.Observe(float64(elapsed))

	if err != nil {
		c.metrics.collectionsErrors.Inc()
		c.metrics.metricsErrors.Add(float64(c.MetricsPerRound))
		atomic.AddInt64(&c.metrics.errorCount, 1)
	} else {
		c.metrics.collectionsDone.Inc()
		c.metrics.metricsSent.Add(float64(c.MetricsPerRound))
		c.metrics.bytesSent.Add(float64(bytesUsed))
		atomic.AddInt64(&c.metrics.successCount, 1)

		// Track bandwidth savings for binary protocol
		if c.Protocol == "binary" {
			// Estimate JSON size (typically 3x larger)
			estimatedJSONSize := bytesUsed * 3
			saved := estimatedJSONSize - bytesUsed
			c.metrics.bytesSavedBinary.Add(float64(saved))
		}
	}
}

// Generate synthetic metrics payload
func (c *SimulatedCollector) generateMetrics() *MetricsPayload {
	c.metrics.metricsGenerated.Add(float64(c.MetricsPerRound))

	metrics := make([]Metric, c.MetricsPerRound)
	now := time.Now().Unix()

	metricNames := []string{
		"cpu_usage_percent",
		"memory_usage_percent",
		"disk_usage_percent",
		"connection_count",
		"queries_per_second",
		"transactions_per_second",
		"cache_hit_ratio",
		"index_scan_count",
		"sequential_scan_count",
	}

	for i := 0; i < c.MetricsPerRound; i++ {
		metrics[i] = Metric{
			Name:      metricNames[i%len(metricNames)],
			Value:     rand.Float64() * 100,
			Timestamp: now,
			Labels: map[string]string{
				"instance": "postgres-prod",
				"job":      "pganalytics",
			},
		}
	}

	return &MetricsPayload{
		CollectorID: c.ID,
		Hostname:    fmt.Sprintf("db-server-%s", c.ID),
		Version:     "1.0.0",
		Metrics:     metrics,
	}
}

// Send JSON protocol
func (c *SimulatedCollector) sendJSON(payload *MetricsPayload) (int, error) {
	// For simplicity, we'll create a JSON representation
	// In reality, this would use json.Marshal
	jsonStr := fmt.Sprintf(
		`{"collector_id":"%s","hostname":"%s","version":"%s","metrics":[]}`,
		payload.CollectorID, payload.Hostname, payload.Version,
	)
	for range payload.Metrics {
		jsonStr += `{"name":"metric","value":50.5,"timestamp":1234567890,"labels":{}}`
	}
	jsonStr += `]}`

	body := []byte(jsonStr)
	return c.sendHTTP("/api/v1/metrics/push", body, "application/json", "gzip")
}

// Send Binary protocol
func (c *SimulatedCollector) sendBinary(payload *MetricsPayload) (int, error) {
	// Simulate binary protocol with smaller payload
	// In reality, this would use the binary_protocol encoding
	binaryPayload := []byte{
		0x01,                                    // Message type: MetricsBatch
		0x01,                                    // Compression: Zstd
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,    // Reserved
		0x00, 0x00, 0x00, 0x80,                 // Payload size (~128 bytes)
		0x00, 0x00, 0x00, 0x00,                 // Timestamp
		0x00, 0x00, 0x00, 0x00,                 // CRC32
		0x01, 0x00,                             // Version
		0x00, 0x00, 0x00, 0x00,                 // Flags
	}

	// Add collector ID and metrics data
	binaryPayload = append(binaryPayload, []byte(payload.CollectorID)...)
	binaryPayload = append(binaryPayload, byte(len(payload.Metrics)))

	return c.sendHTTP("/api/v1/metrics/push/binary", binaryPayload, "application/octet-stream", "zstd")
}

// Send HTTP request
func (c *SimulatedCollector) sendHTTP(endpoint string, body []byte, contentType, encoding string) (int, error) {
	start := time.Now()

	url := c.BackendURL + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("X-Collector-ID", c.ID)
	req.Header.Set("X-Protocol-Version", "1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Read response to measure backend processing time
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	elapsed := time.Since(start).Milliseconds()
	c.metrics.ingestionTime.Observe(float64(elapsed))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return len(body), fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return len(body), nil
}

// Print final report
func printReport(metrics *LoadTestMetrics, config LoadTestConfig) {
	elapsed := time.Since(metrics.startTime)
	success := atomic.LoadInt64(&metrics.successCount)
	errors := atomic.LoadInt64(&metrics.errorCount)
	total := success + errors

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	fmt.Println("\n" + "═"*80)
	fmt.Println("                           LOAD TEST REPORT")
	fmt.Println("═"*80)
	fmt.Println()

	fmt.Println("TEST CONFIGURATION")
	fmt.Printf("  Collectors:         %d\n", config.NumCollectors)
	fmt.Printf("  Protocol:           %s\n", config.Protocol)
	fmt.Printf("  Duration:           %v\n", config.Duration)
	fmt.Printf("  Collection Interval: %v\n", config.Interval)
	fmt.Printf("  Metrics/Collector:  %d\n", config.MetricsPerCollector)
	fmt.Println()

	fmt.Println("RESULTS SUMMARY")
	fmt.Printf("  Total Collections:  %d\n", total)
	fmt.Printf("  Successful:         %d (%.2f%%)\n", success, successRate)
	fmt.Printf("  Errors:             %d (%.2f%%)\n", errors, 100-successRate)
	fmt.Printf("  Actual Duration:    %v\n", elapsed)
	fmt.Println()

	fmt.Println("PERFORMANCE METRICS")
	metricsPerHour := float64(config.NumCollectors*config.MetricsPerCollector) * (60.0 / config.Interval.Minutes())
	fmt.Printf("  Metrics/Hour:       %.0f\n", metricsPerHour)
	fmt.Printf("  Throughput:         %.2f metrics/sec\n", float64(success*int64(config.MetricsPerCollector))/elapsed.Seconds())
	fmt.Println()

	fmt.Println("═"*80)
	fmt.Println()
}
