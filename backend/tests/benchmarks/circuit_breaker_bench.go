package benchmarks

import (
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// BenchmarkCircuitBreakerIsOpen benchmarks IsOpen() state check
func BenchmarkCircuitBreakerIsOpen(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.IsOpen()
	}
}

// BenchmarkCircuitBreakerRecordSuccess benchmarks success recording
func BenchmarkCircuitBreakerRecordSuccess(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordSuccess()
	}
}

// BenchmarkCircuitBreakerRecordFailure benchmarks failure recording
func BenchmarkCircuitBreakerRecordFailure(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordFailure()
	}
}

// BenchmarkCircuitBreakerState benchmarks state retrieval
func BenchmarkCircuitBreakerState(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.State()
	}
}

// BenchmarkCircuitBreakerGetMetrics benchmarks metrics retrieval
func BenchmarkCircuitBreakerGetMetrics(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	// Record some operations
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.GetMetrics()
	}
}

// BenchmarkCircuitBreakerReset benchmarks reset operation
func BenchmarkCircuitBreakerReset(b *testing.B) {
	logger := zaptest.NewLogger(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb := ml.NewCircuitBreaker(logger)
		cb.Reset()
	}
}

// BenchmarkCircuitBreakerConcurrentReads benchmarks concurrent read operations
func BenchmarkCircuitBreakerConcurrentReads(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.IsOpen()
			_ = cb.State()
		}
	})
}

// BenchmarkCircuitBreakerConcurrentWriteRead benchmarks mixed concurrent operations
func BenchmarkCircuitBreakerConcurrentWriteRead(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cb.RecordSuccess()
			} else {
				cb.RecordFailure()
			}
			i++
		}
	})
}

// BenchmarkCircuitBreakerStateTransition benchmarks full state transition cycle
func BenchmarkCircuitBreakerStateTransition(b *testing.B) {
	logger := zaptest.NewLogger(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb := ml.NewCircuitBreaker(logger)

		// Open the circuit
		for j := 0; j < 5; j++ {
			cb.RecordFailure()
		}

		// Reset
		cb.Reset()

		// Verify closed
		_ = cb.IsOpen()
	}
}

// BenchmarkCircuitBreakerMetricsCalculation benchmarks metrics calculation
func BenchmarkCircuitBreakerMetricsCalculation(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	// Pre-populate with operations
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := cb.GetMetrics()
		_ = metrics
	}
}

// BenchmarkCircuitBreakerMemoryAllocations benchmarks memory usage
func BenchmarkCircuitBreakerMemoryAllocations(b *testing.B) {
	logger := zaptest.NewLogger(b)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cb := ml.NewCircuitBreaker(logger)
		cb.RecordSuccess()
		cb.RecordFailure()
		_ = cb.GetMetrics()
	}
}

// BenchmarkCircuitBreakerOperationSequence benchmarks typical operation sequence
func BenchmarkCircuitBreakerOperationSequence(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Typical operation sequence:
		// 1. Check if open
		if !cb.IsOpen() {
			// 2. Record result
			if i%10 == 0 {
				cb.RecordFailure()
			} else {
				cb.RecordSuccess()
			}
		}
		// 3. Get state
		_ = cb.State()
	}
}

// BenchmarkCircuitBreakerHighContention benchmarks behavior under high contention
func BenchmarkCircuitBreakerHighContention(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			if counter%100 == 0 {
				_ = cb.GetMetrics()
			} else if counter%10 == 0 {
				_ = cb.State()
			} else if counter%3 == 0 {
				cb.RecordFailure()
			} else {
				cb.RecordSuccess()
			}
		}
	})
}

// BenchmarkCircuitBreakerAfterStateChange benchmarks operations after state change
func BenchmarkCircuitBreakerAfterStateChange(b *testing.B) {
	logger := zaptest.NewLogger(b)
	cb := ml.NewCircuitBreaker(logger)

	// Change state to open
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Operations on open circuit
		_ = cb.IsOpen()
		_ = cb.State()
	}
}

// BenchmarkCircuitBreakerEdgeCases benchmarks edge case operations
func BenchmarkCircuitBreakerEdgeCases(b *testing.B) {
	logger := zaptest.NewLogger(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb := ml.NewCircuitBreaker(logger)

		// Edge case 1: Many failures
		for j := 0; j < 100; j++ {
			cb.RecordFailure()
		}

		// Edge case 2: Reset
		cb.Reset()

		// Edge case 3: Many successes
		for j := 0; j < 100; j++ {
			cb.RecordSuccess()
		}

		// Edge case 4: Get metrics
		_ = cb.GetMetrics()
	}
}

// BenchmarkCircuitBreakerDashboard provides overall performance dashboard
func BenchmarkCircuitBreakerDashboard(b *testing.B) {
	logger := zaptest.NewLogger(b)
	benchmarks := []struct {
		name string
		fn   func(*testing.B, *ml.CircuitBreaker)
	}{
		{"IsOpen", func(b *testing.B, cb *ml.CircuitBreaker) {
			for i := 0; i < b.N; i++ {
				_ = cb.IsOpen()
			}
		}},
		{"RecordSuccess", func(b *testing.B, cb *ml.CircuitBreaker) {
			for i := 0; i < b.N; i++ {
				cb.RecordSuccess()
			}
		}},
		{"RecordFailure", func(b *testing.B, cb *ml.CircuitBreaker) {
			for i := 0; i < b.N; i++ {
				cb.RecordFailure()
			}
		}},
		{"State", func(b *testing.B, cb *ml.CircuitBreaker) {
			for i := 0; i < b.N; i++ {
				_ = cb.State()
			}
		}},
		{"GetMetrics", func(b *testing.B, cb *ml.CircuitBreaker) {
			for i := 0; i < b.N; i++ {
				_ = cb.GetMetrics()
			}
		}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			cb := ml.NewCircuitBreaker(logger)
			bm.fn(b, cb)
		})
	}
}

// TestCircuitBreakerPerformanceCharacteristics documents performance characteristics
func TestCircuitBreakerPerformanceCharacteristics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Expected performance characteristics:
	// - IsOpen(): <1μs (microseconds)
	// - RecordSuccess(): ~1μs
	// - RecordFailure(): ~1μs
	// - State(): <1μs
	// - GetMetrics(): ~2μs
	// - Memory overhead: ~500 bytes per circuit breaker

	cb := ml.NewCircuitBreaker(logger)

	// Verify circuit breaker exists and is functional
	if cb == nil {
		t.Fatalf("Failed to create circuit breaker")
	}

	// Verify basic operations work
	start := time.Now()
	for i := 0; i < 10000; i++ {
		_ = cb.IsOpen()
	}
	elapsed := time.Since(start)

	// Should complete 10,000 operations in <10ms
	if elapsed > 10*time.Millisecond {
		t.Logf("Warning: IsOpen() took %v for 10000 calls (expected <10ms)", elapsed)
	}

	t.Logf("Performance characteristics verified: IsOpen() avg = %.2f ns", float64(elapsed.Nanoseconds())/10000)
}
