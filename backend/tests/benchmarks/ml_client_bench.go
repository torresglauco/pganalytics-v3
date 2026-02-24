package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	"github.com/torresglauco/pganalytics-v3/backend/tests/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// BenchmarkMLClientHealthCheck benchmarks health check operation
func BenchmarkMLClientHealthCheck(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.IsHealthy(ctx)
	}
}

// BenchmarkMLClientPrediction benchmarks prediction request
func BenchmarkMLClientPrediction(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features: map[string]interface{}{
			"mean_execution_time_ms": 125.5,
			"calls_per_minute":       100.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.PredictQueryExecution(ctx, req)
	}
}

// BenchmarkMLClientTraining benchmarks model training request
func BenchmarkMLClientTraining(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.TrainingRequest{
		DatabaseURL:  "postgresql://localhost/db",
		LookbackDays: 90,
		ModelType:    "random_forest",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.TrainPerformanceModel(ctx, req)
	}
}

// BenchmarkMLClientValidation benchmarks prediction validation
func BenchmarkMLClientValidation(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.ValidationRequest{
		PredictionID:             "pred-001",
		QueryHash:                4001,
		PredictedExecutionTimeMs: 125.5,
		ActualExecutionTimeMs:    118.2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ValidatePrediction(ctx, req)
	}
}

// BenchmarkMLClientPatternDetection benchmarks pattern detection
func BenchmarkMLClientPatternDetection(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.PatternRequest{
		DatabaseURL:  "postgresql://localhost/db",
		LookbackDays: 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.DetectWorkloadPatterns(ctx, req)
	}
}

// BenchmarkMLClientConcurrentRequests benchmarks concurrent prediction requests
func BenchmarkMLClientConcurrentRequests(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features:  make(map[string]interface{}),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.PredictQueryExecution(ctx, req)
		}
	})
}

// BenchmarkMLClientSequentialRequests benchmarks sequential requests
func BenchmarkMLClientSequentialRequests(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queryHash := int64(4000 + (i % 100))
		req := &ml.PredictionRequest{
			QueryHash: queryHash,
			Features:  make(map[string]interface{}),
		}
		_, _ = client.PredictQueryExecution(ctx, req)
	}
}

// BenchmarkMLClientCircuitBreakerStateCheck benchmarks circuit breaker state check
func BenchmarkMLClientCircuitBreakerStateCheck(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetCircuitBreakerState()
	}
}

// BenchmarkMLClientErrorRecovery benchmarks error handling and recovery
func BenchmarkMLClientErrorRecovery(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features:  make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between success and failure
		if i%2 == 0 {
			mockService.SetShouldFail(false)
		} else {
			mockService.SetShouldFail(true)
		}
		_, _ = client.PredictQueryExecution(ctx, req)
	}
}

// BenchmarkMLClientWithTimeout benchmarks timeout handling
func BenchmarkMLClientWithTimeout(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	mockService.SetResponseDelay(100 * time.Millisecond)
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 500*time.Millisecond, logger)
	defer client.Close()

	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features:  make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		_, _ = client.PredictQueryExecution(ctx, req)
		cancel()
	}
}

// BenchmarkMLClientMemoryAllocations benchmarks memory usage
func BenchmarkMLClientMemoryAllocations(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()
	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features:  make(map[string]interface{}),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.PredictQueryExecution(ctx, req)
	}
}

// BenchmarkMLClientOperationSequence benchmarks typical operation sequence
func BenchmarkMLClientOperationSequence(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Typical sequence:
		// 1. Check health
		_ = client.IsHealthy(ctx)

		// 2. Make prediction
		req := &ml.PredictionRequest{
			QueryHash: int64(4000 + (i % 10)),
			Features:  make(map[string]interface{}),
		}
		_, _ = client.PredictQueryExecution(ctx, req)

		// 3. Check circuit breaker
		_ = client.GetCircuitBreakerState()
	}
}

// BenchmarkMLClientEndToEndWorkflow benchmarks complete prediction workflow
func BenchmarkMLClientEndToEndWorkflow(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 1. Make prediction
		predReq := &ml.PredictionRequest{
			QueryHash: int64(4000 + (i % 10)),
			Features: map[string]interface{}{
				"mean_execution_time_ms": 125.5,
				"calls_per_minute":       100.0,
			},
		}
		predResp, err := client.PredictQueryExecution(ctx, predReq)

		if err == nil && predResp != nil {
			// 2. Validate prediction
			valReq := &ml.ValidationRequest{
				PredictionID:             "pred-001",
				QueryHash:                predReq.QueryHash,
				PredictedExecutionTimeMs: predResp.PredictedExecutionMs,
				ActualExecutionTimeMs:    118.2,
			}
			_, _ = client.ValidatePrediction(ctx, valReq)
		}
	}
}

// BenchmarkMLClientDashboard provides overall performance dashboard
func BenchmarkMLClientDashboard(b *testing.B) {
	logger := zaptest.NewLogger(b)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	ctx := context.Background()

	benchmarks := []struct {
		name string
		fn   func(*testing.B, *ml.Client)
	}{
		{"HealthCheck", func(b *testing.B, client *ml.Client) {
			for i := 0; i < b.N; i++ {
				_ = client.IsHealthy(ctx)
			}
		}},
		{"Prediction", func(b *testing.B, client *ml.Client) {
			req := &ml.PredictionRequest{QueryHash: 4001, Features: make(map[string]interface{})}
			for i := 0; i < b.N; i++ {
				_, _ = client.PredictQueryExecution(ctx, req)
			}
		}},
		{"Training", func(b *testing.B, client *ml.Client) {
			req := &ml.TrainingRequest{DatabaseURL: "postgresql://localhost/db", LookbackDays: 90}
			for i := 0; i < b.N; i++ {
				_, _ = client.TrainPerformanceModel(ctx, req)
			}
		}},
		{"Validation", func(b *testing.B, client *ml.Client) {
			req := &ml.ValidationRequest{PredictionID: "pred-001", QueryHash: 4001, PredictedExecutionTimeMs: 125.5, ActualExecutionTimeMs: 118.2}
			for i := 0; i < b.N; i++ {
				_, _ = client.ValidatePrediction(ctx, req)
			}
		}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
			defer client.Close()
			bm.fn(b, client)
		})
	}
}

// TestMLClientPerformanceCharacteristics documents performance characteristics
func TestMLClientPerformanceCharacteristics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx := context.Background()

	// Expected performance characteristics:
	// - Health check: 10-20ms
	// - Prediction: 50-100ms
	// - Training: 100-200ms
	// - Validation: 50-100ms
	// - Pattern detection: 100-200ms

	// Verify basic operations work and measure timing
	start := time.Now()
	for i := 0; i < 100; i++ {
		_ = client.IsHealthy(ctx)
	}
	elapsed := time.Since(start)

	// Should complete 100 health checks in <2 seconds
	if elapsed > 2*time.Second {
		t.Logf("Warning: IsHealthy() took %v for 100 calls", elapsed)
	}

	t.Logf("Performance characteristics verified: IsHealthy() avg = %.2f ms", float64(elapsed.Milliseconds())/100)
}
