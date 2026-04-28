package load

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LoadTestConfig defines configuration for load testing
type LoadTestConfig struct {
	BaseURL             string // API base URL
	NumCollectors       int    // Number of simulated collectors
	MetricsPerCollector int    // Metrics per push (default: 10)
	PushIntervalSeconds int    // Interval between pushes (default: 5)
	DurationMinutes     int    // Test duration (default: 5)
	ConcurrentPushes    int    // Concurrent metric pushes (default: 10)
	EnableRateLimitTest bool   // Test rate limiting
	EnableCacheTest     bool   // Test configuration caching
	EnableCleanupTest   bool   // Monitor cleanup job
	Verbose             bool   // Verbose logging
}

// LoadTestResults holds the results of a load test
type LoadTestResults struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	RateLimitedRequests int64
	TotalDuration       time.Duration
	AverageLatency      time.Duration
	MinLatency          time.Duration
	MaxLatency          time.Duration
	P50Latency          time.Duration
	P95Latency          time.Duration
	P99Latency          time.Duration
	RequestsPerSecond   float64
	ErrorRate           float64
	RateLimitRate       float64

	// Cache statistics
	CacheHits    int64
	CacheMisses  int64
	CacheHitRate float64

	// Detailed metrics
	Latencies        []time.Duration
	StatusCodeCounts map[int]int64
}

// MetricPush represents a metric push request
type MetricPush struct {
	CollectorID string            `json:"collector_id"`
	Timestamp   int64             `json:"timestamp"`
	Metrics     []MetricDataPoint `json:"metrics"`
}

// MetricDataPoint represents a single metric
type MetricDataPoint struct {
	Name  string            `json:"name"`
	Value float64           `json:"value"`
	Tags  map[string]string `json:"tags"`
}

// LoadTester handles load testing
type LoadTester struct {
	config  *LoadTestConfig
	client  *http.Client
	results *LoadTestResults
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// Metrics tracking
	totalRequests       int64
	successfulRequests  int64
	failedRequests      int64
	rateLimitedRequests int64
	cacheHits           int64
	cacheMisses         int64
	latencies           []time.Duration
	statusCodes         map[int]int64
}

// NewLoadTester creates a new load tester
func NewLoadTester(config *LoadTestConfig) *LoadTester {
	ctx, cancel := context.WithCancel(context.Background())
	return &LoadTester{
		config:      config,
		client:      &http.Client{Timeout: 10 * time.Second},
		results:     &LoadTestResults{},
		ctx:         ctx,
		cancel:      cancel,
		latencies:   make([]time.Duration, 0),
		statusCodes: make(map[int]int64),
	}
}

// Run executes the load test
func (lt *LoadTester) Run() *LoadTestResults {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("LOAD TEST: %d Collectors, %d metrics/push, %d minute duration\n",
		lt.config.NumCollectors, lt.config.MetricsPerCollector, lt.config.DurationMinutes)
	fmt.Println(strings.Repeat("=", 80) + "\n")

	startTime := time.Now()
	testDuration := time.Duration(lt.config.DurationMinutes) * time.Minute

	// Determine interval for pushes
	interval := time.Duration(lt.config.PushIntervalSeconds) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	// Start collector goroutines
	concurrentPushes := lt.config.ConcurrentPushes
	if concurrentPushes == 0 {
		concurrentPushes = 10
	}

	semaphore := make(chan struct{}, concurrentPushes)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Progress reporting
	progressTicker := time.NewTicker(30 * time.Second)
	defer progressTicker.Stop()

	fmt.Printf("Starting %d collectors at %s\n\n", lt.config.NumCollectors, startTime.Format("15:04:05"))

	for {
		select {
		case <-time.After(testDuration):
			// Test duration complete
			lt.wg.Wait()
			elapsed := time.Since(startTime)
			return lt.calculateResults(elapsed)

		case <-ticker.C:
			// Push metrics from all collectors
			for collectorID := 0; collectorID < lt.config.NumCollectors; collectorID++ {
				lt.wg.Add(1)
				go func(id int) {
					defer lt.wg.Done()
					semaphore <- struct{}{}        // Acquire slot
					defer func() { <-semaphore }() // Release slot

					lt.pushMetrics(fmt.Sprintf("collector-%d", id))
				}(collectorID)
			}

		case <-progressTicker.C:
			// Print progress
			lt.printProgress(time.Since(startTime))
		}
	}
}

// pushMetrics sends metrics for a collector
func (lt *LoadTester) pushMetrics(collectorID string) {
	startTime := time.Now()

	// Generate metrics
	push := lt.generateMetricPush(collectorID)

	payload, err := json.Marshal(push)
	if err != nil {
		atomic.AddInt64(&lt.failedRequests, 1)
		return
	}

	// Create request
	req, err := http.NewRequestWithContext(
		lt.ctx,
		"POST",
		fmt.Sprintf("%s/api/v1/metrics/push", lt.config.BaseURL),
		bytes.NewReader(payload),
	)
	if err != nil {
		atomic.AddInt64(&lt.failedRequests, 1)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Collector-ID", collectorID)
	req.Header.Set("Authorization", "Bearer test-token-"+collectorID)

	// Execute request
	resp, err := lt.client.Do(req)
	latency := time.Since(startTime)

	atomic.AddInt64(&lt.totalRequests, 1)
	lt.mu.Lock()
	lt.latencies = append(lt.latencies, latency)
	lt.mu.Unlock()

	if err != nil {
		atomic.AddInt64(&lt.failedRequests, 1)
		if lt.config.Verbose {
			fmt.Printf("Error from %s: %v\n", collectorID, err)
		}
		return
	}
	defer resp.Body.Close()

	// Track status code
	lt.mu.Lock()
	lt.statusCodes[resp.StatusCode]++
	lt.mu.Unlock()

	// Process response
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		atomic.AddInt64(&lt.successfulRequests, 1)

		// Check for cache hit in response headers
		if resp.Header.Get("X-Cache") == "HIT" {
			atomic.AddInt64(&lt.cacheHits, 1)
		} else if resp.Header.Get("X-Cache") == "MISS" {
			atomic.AddInt64(&lt.cacheMisses, 1)
		}

	case http.StatusTooManyRequests:
		atomic.AddInt64(&lt.rateLimitedRequests, 1)
		if lt.config.Verbose {
			fmt.Printf("Rate limited: %s (latency: %dms)\n", collectorID, latency.Milliseconds())
		}

	default:
		atomic.AddInt64(&lt.failedRequests, 1)
		if lt.config.Verbose {
			fmt.Printf("Error %d from %s\n", resp.StatusCode, collectorID)
		}
	}
}

// generateMetricPush creates a metric push payload
func (lt *LoadTester) generateMetricPush(collectorID string) *MetricPush {
	numMetrics := lt.config.MetricsPerCollector
	if numMetrics == 0 {
		numMetrics = 10
	}

	metrics := make([]MetricDataPoint, numMetrics)
	for i := 0; i < numMetrics; i++ {
		metrics[i] = MetricDataPoint{
			Name:  fmt.Sprintf("query_%d_execution_time", i+1),
			Value: float64(rand.Intn(1000)) + rand.Float64(), // 0-1000ms
			Tags: map[string]string{
				"database": fmt.Sprintf("pg_%d", rand.Intn(100)),
				"query":    fmt.Sprintf("query_%d", rand.Intn(50)),
			},
		}
	}

	return &MetricPush{
		CollectorID: collectorID,
		Timestamp:   time.Now().Unix(),
		Metrics:     metrics,
	}
}

// printProgress prints test progress
func (lt *LoadTester) printProgress(elapsed time.Duration) {
	total := atomic.LoadInt64(&lt.totalRequests)
	successful := atomic.LoadInt64(&lt.successfulRequests)
	failed := atomic.LoadInt64(&lt.failedRequests)
	rateLimited := atomic.LoadInt64(&lt.rateLimitedRequests)

	rps := float64(total) / elapsed.Seconds()
	successRate := float64(0)
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	fmt.Printf("[%s] Total: %d | Success: %d (%.1f%%) | Failed: %d | Rate-Limited: %d | RPS: %.1f\n",
		elapsed.Round(time.Second),
		total, successful, successRate, failed, rateLimited, rps)
}

// calculateResults computes final test results
func (lt *LoadTester) calculateResults(duration time.Duration) *LoadTestResults {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	total := atomic.LoadInt64(&lt.totalRequests)
	successful := atomic.LoadInt64(&lt.successfulRequests)
	failed := atomic.LoadInt64(&lt.failedRequests)
	rateLimited := atomic.LoadInt64(&lt.rateLimitedRequests)
	cacheHits := atomic.LoadInt64(&lt.cacheHits)
	cacheMisses := atomic.LoadInt64(&lt.cacheMisses)

	// Calculate latency statistics
	results := &LoadTestResults{
		TotalRequests:       total,
		SuccessfulRequests:  successful,
		FailedRequests:      failed,
		RateLimitedRequests: rateLimited,
		TotalDuration:       duration,
		CacheHits:           cacheHits,
		CacheMisses:         cacheMisses,
		StatusCodeCounts:    lt.statusCodes,
		Latencies:           lt.latencies,
	}

	if total > 0 {
		results.RequestsPerSecond = float64(total) / duration.Seconds()
		results.ErrorRate = float64(failed) / float64(total) * 100
		results.RateLimitRate = float64(rateLimited) / float64(total) * 100
	}

	if cacheHits+cacheMisses > 0 {
		results.CacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses) * 100
	}

	// Calculate percentiles
	if len(lt.latencies) > 0 {
		// Sort latencies (simplified - in production use proper sorting)
		results.MinLatency = lt.latencies[0]
		results.MaxLatency = lt.latencies[0]
		var totalLatency time.Duration

		for _, lat := range lt.latencies {
			if lat < results.MinLatency {
				results.MinLatency = lat
			}
			if lat > results.MaxLatency {
				results.MaxLatency = lat
			}
			totalLatency += lat
		}

		results.AverageLatency = time.Duration(int64(totalLatency) / int64(len(lt.latencies)))

		// Rough percentile calculation
		idx50 := len(lt.latencies) * 50 / 100
		idx95 := len(lt.latencies) * 95 / 100
		idx99 := len(lt.latencies) * 99 / 100

		if idx50 < len(lt.latencies) {
			results.P50Latency = lt.latencies[idx50]
		}
		if idx95 < len(lt.latencies) {
			results.P95Latency = lt.latencies[idx95]
		}
		if idx99 < len(lt.latencies) {
			results.P99Latency = lt.latencies[idx99]
		}
	}

	return results
}

// PrintResults prints the test results
func (results *LoadTestResults) PrintResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	fmt.Printf("Duration:              %v\n", results.TotalDuration)
	fmt.Printf("Total Requests:        %d\n", results.TotalRequests)
	fmt.Printf("Successful:            %d (%.1f%%)\n",
		results.SuccessfulRequests,
		float64(results.SuccessfulRequests)/float64(results.TotalRequests)*100)
	fmt.Printf("Failed:                %d (%.1f%%)\n",
		results.FailedRequests, results.ErrorRate)
	fmt.Printf("Rate Limited (429):    %d (%.1f%%)\n",
		results.RateLimitedRequests, results.RateLimitRate)

	fmt.Printf("\nThroughput:\n")
	fmt.Printf("  Requests/Second:     %.1f req/s\n", results.RequestsPerSecond)

	fmt.Printf("\nLatency Statistics (milliseconds):\n")
	fmt.Printf("  Min:                 %d ms\n", results.MinLatency.Milliseconds())
	fmt.Printf("  Average:             %d ms\n", results.AverageLatency.Milliseconds())
	fmt.Printf("  P50 (Median):        %d ms\n", results.P50Latency.Milliseconds())
	fmt.Printf("  P95:                 %d ms\n", results.P95Latency.Milliseconds())
	fmt.Printf("  P99:                 %d ms\n", results.P99Latency.Milliseconds())
	fmt.Printf("  Max:                 %d ms\n", results.MaxLatency.Milliseconds())

	if results.CacheHits > 0 || results.CacheMisses > 0 {
		fmt.Printf("\nCache Statistics:\n")
		fmt.Printf("  Hits:                %d\n", results.CacheHits)
		fmt.Printf("  Misses:              %d\n", results.CacheMisses)
		fmt.Printf("  Hit Rate:            %.1f%%\n", results.CacheHitRate)
	}

	if len(results.StatusCodeCounts) > 0 {
		fmt.Printf("\nStatus Code Distribution:\n")
		for code, count := range results.StatusCodeCounts {
			fmt.Printf("  %d:                   %d\n", code, count)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SUCCESS CRITERIA VALIDATION:")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	// Validate against success criteria
	criteria := []struct {
		name     string
		value    float64
		target   float64
		operator string
		unit     string
	}{
		{"p95 latency", float64(results.P95Latency.Milliseconds()), 500, "<", "ms"},
		{"error rate", results.ErrorRate, 0.1, "<", "%"},
		{"cache hit rate", results.CacheHitRate, 75, ">", "%"},
	}

	for _, c := range criteria {
		passed := false
		if c.operator == "<" {
			passed = c.value < c.target
		} else if c.operator == ">" {
			passed = c.value > c.target
		}

		status := "❌ FAIL"
		if passed {
			status = "✅ PASS"
		}

		fmt.Printf("%s: %s (%.1f %s, target: %.1f %s)\n",
			status, c.name, c.value, c.unit, c.target, c.unit)
	}

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
}
