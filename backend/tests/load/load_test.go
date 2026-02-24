package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	"github.com/torresglauco/pganalytics-v3/backend/tests/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// LoadTestResult contains results from a load test
type LoadTestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailureRequests int64
	Duration        time.Duration
	RequestsPerSec  float64
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
}

// TestMLClientLoadPredictions tests ML client under load
func TestMLClientLoadPredictions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	// Load test parameters
	numGoroutines := 10
	requestsPerGoroutine := 100
	totalRequests := numGoroutines * requestsPerGoroutine

	// Results tracking
	var successCount int64
	var failureCount int64
	var totalLatency int64
	latencies := make([]time.Duration, 0, totalRequests)
	latenciesMu := sync.Mutex{}

	// Run load test
	ctx := context.Background()
	start := time.Now()
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < requestsPerGoroutine; i++ {
				queryHash := int64(4000 + (goroutineID*requestsPerGoroutine + i)%100)
				req := &ml.PredictionRequest{
					QueryHash: queryHash,
					Features: map[string]interface{}{
						"mean_execution_time_ms": 125.5,
						"calls_per_minute":       100.0,
					},
				}

				reqStart := time.Now()
				_, err := client.PredictQueryExecution(ctx, req)
				latency := time.Since(reqStart)

				latenciesMu.Lock()
				latencies = append(latencies, latency)
				latenciesMu.Unlock()

				atomic.AddInt64(&totalLatency, latency.Nanoseconds())

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Calculate statistics
	result := LoadTestResult{
		TotalRequests:   int64(totalRequests),
		SuccessRequests: atomic.LoadInt64(&successCount),
		FailureRequests: atomic.LoadInt64(&failureCount),
		Duration:        elapsed,
		RequestsPerSec:  float64(totalRequests) / elapsed.Seconds(),
		AvgLatency:      time.Duration(atomic.LoadInt64(&totalLatency) / int64(totalRequests)),
	}

	// Calculate percentiles
	result.MinLatency, result.MaxLatency = calculatePercentiles(latencies, &result)

	// Log results
	t.Logf("=== Load Test Results: ML Client Predictions ===")
	t.Logf("Total Requests:    %d", result.TotalRequests)
	t.Logf("Successful:        %d (%.2f%%)", result.SuccessRequests, float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	t.Logf("Failed:            %d (%.2f%%)", result.FailureRequests, float64(result.FailureRequests)/float64(result.TotalRequests)*100)
	t.Logf("Duration:          %v", result.Duration)
	t.Logf("Requests/Sec:      %.2f", result.RequestsPerSec)
	t.Logf("Avg Latency:       %v", result.AvgLatency)
	t.Logf("Min Latency:       %v", result.MinLatency)
	t.Logf("Max Latency:       %v", result.MaxLatency)
	t.Logf("P50 Latency:       %v", result.P50Latency)
	t.Logf("P95 Latency:       %v", result.P95Latency)
	t.Logf("P99 Latency:       %v", result.P99Latency)

	// Assertions
	if result.SuccessRequests == 0 {
		t.Fatalf("Expected successful requests, got 0")
	}

	// Success rate should be >90%
	successRate := float64(result.SuccessRequests) / float64(result.TotalRequests)
	if successRate < 0.90 {
		t.Errorf("Expected success rate >90%%, got %.2f%%", successRate*100)
	}

	// Average latency should be reasonable (<1s)
	if result.AvgLatency > time.Second {
		t.Errorf("Average latency too high: %v", result.AvgLatency)
	}
}

// TestCircuitBreakerLoadBehavior tests circuit breaker under load
func TestCircuitBreakerLoadBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	logger := zaptest.NewLogger(t)
	cb := ml.NewCircuitBreaker(logger)

	// Load test parameters
	numGoroutines := 20
	operationsPerGoroutine := 500
	totalOperations := numGoroutines * operationsPerGoroutine

	// Results tracking
	var operationCount int64

	// Run load test
	start := time.Now()
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < operationsPerGoroutine; i++ {
				// Simulate mixed operations
				switch i % 4 {
				case 0:
					_ = cb.IsOpen()
				case 1:
					cb.RecordSuccess()
				case 2:
					cb.RecordFailure()
				case 3:
					_ = cb.State()
				}
				atomic.AddInt64(&operationCount, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	count := atomic.LoadInt64(&operationCount)
	opsPerSec := float64(count) / elapsed.Seconds()

	t.Logf("=== Load Test Results: Circuit Breaker ===")
	t.Logf("Total Operations:  %d", count)
	t.Logf("Duration:          %v", elapsed)
	t.Logf("Ops/Sec:           %.2f", opsPerSec)

	// Verify all operations completed
	if count != int64(totalOperations) {
		t.Errorf("Expected %d operations, got %d", totalOperations, count)
	}

	// Verify circuit breaker is still functional
	if !cb.IsOpen() {
		t.Errorf("Circuit breaker should be accepting requests")
	}
}

// TestConcurrentTrainingRequests tests multiple concurrent training requests
func TestConcurrentTrainingRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	numGoroutines := 5
	trainingRequestsPerGoroutine := 20
	totalRequests := numGoroutines * trainingRequestsPerGoroutine

	var successCount int64
	var failureCount int64

	ctx := context.Background()
	start := time.Now()
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < trainingRequestsPerGoroutine; i++ {
				req := &ml.TrainingRequest{
					DatabaseURL:  "postgresql://localhost/db",
					LookbackDays: 90,
					ModelType:    "random_forest",
					JobID:        fmt.Sprintf("job-%d-%d", goroutineID, i),
				}

				_, err := client.TrainPerformanceModel(ctx, req)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	success := atomic.LoadInt64(&successCount)
	failure := atomic.LoadInt64(&failureCount)

	t.Logf("=== Concurrent Training Requests ===")
	t.Logf("Total Requests:    %d", totalRequests)
	t.Logf("Successful:        %d", success)
	t.Logf("Failed:            %d", failure)
	t.Logf("Duration:          %v", elapsed)
	t.Logf("Requests/Sec:      %.2f", float64(totalRequests)/elapsed.Seconds())

	if success == 0 {
		t.Fatalf("Expected successful training requests")
	}
}

// TestHighContention tests circuit breaker with high contention
func TestHighContention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	logger := zaptest.NewLogger(t)
	cb := ml.NewCircuitBreaker(logger)

	numGoroutines := 50
	operationsPerGoroutine := 200
	failurePattern := 30 // Fail every 30th operation

	var successCount int64
	var failureCount int64

	start := time.Now()
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < operationsPerGoroutine; i++ {
				// Simulate failures periodically
				if (goroutineID*operationsPerGoroutine + i) % failurePattern == 0 {
					cb.RecordFailure()
					atomic.AddInt64(&failureCount, 1)
				} else {
					cb.RecordSuccess()
					atomic.AddInt64(&successCount, 1)
				}

				// Also check state
				_ = cb.IsOpen()
				_ = cb.State()
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	success := atomic.LoadInt64(&successCount)
	failure := atomic.LoadInt64(&failureCount)
	totalOps := success + failure

	t.Logf("=== High Contention Test ===")
	t.Logf("Total Operations:  %d", totalOps)
	t.Logf("Successful:        %d", success)
	t.Logf("Failed:            %d", failure)
	t.Logf("Duration:          %v", elapsed)
	t.Logf("Ops/Sec:           %.2f", float64(totalOps)/elapsed.Seconds())
	t.Logf("Circuit State:     %s", cb.State())
}

// TestSustainedLoad tests sustained load over longer duration
func TestSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	// Run load for 5 seconds
	duration := 5 * time.Second
	numGoroutines := 10

	var requestCount int64
	var successCount int64
	var failureCount int64

	ctx := context.Background()
	done := make(chan struct{})

	go func() {
		time.Sleep(duration)
		close(done)
	}()

	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-done:
					return
				default:
					req := &ml.PredictionRequest{
						QueryHash: int64(4000 + (requestCount % 100)),
						Features:  make(map[string]interface{}),
					}

					_, err := client.PredictQueryExecution(ctx, req)
					atomic.AddInt64(&requestCount, 1)

					if err != nil {
						atomic.AddInt64(&failureCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}
		}()
	}

	start := time.Now()
	wg.Wait()
	elapsed := time.Since(start)

	requests := atomic.LoadInt64(&requestCount)
	successes := atomic.LoadInt64(&successCount)
	failures := atomic.LoadInt64(&failureCount)

	t.Logf("=== Sustained Load Test (5 seconds) ===")
	t.Logf("Total Requests:    %d", requests)
	t.Logf("Successful:        %d (%.2f%%)", successes, float64(successes)/float64(requests)*100)
	t.Logf("Failed:            %d (%.2f%%)", failures, float64(failures)/float64(requests)*100)
	t.Logf("Actual Duration:   %v", elapsed)
	t.Logf("Requests/Sec:      %.2f", float64(requests)/elapsed.Seconds())

	if successes == 0 {
		t.Fatalf("Expected successful requests during sustained load")
	}
}

// Helper function to calculate percentiles
func calculatePercentiles(latencies []time.Duration, result *LoadTestResult) (time.Duration, time.Duration) {
	if len(latencies) == 0 {
		return 0, 0
	}

	// Simple sorting (not optimized for large datasets)
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Bubble sort for simplicity (acceptable for benchmarking)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	min := sorted[0]
	max := sorted[len(sorted)-1]

	// Calculate percentiles
	p50Index := len(sorted) / 2
	p95Index := (len(sorted) * 95) / 100
	p99Index := (len(sorted) * 99) / 100

	result.P50Latency = sorted[p50Index]
	result.P95Latency = sorted[p95Index]
	result.P99Latency = sorted[p99Index]

	return min, max
}
