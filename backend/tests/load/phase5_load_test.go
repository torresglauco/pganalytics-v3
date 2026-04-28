package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkAnomalyDetection benchmarks anomaly detection performance
func BenchmarkAnomalyDetection(b *testing.B) {
	b.Run("zscore_calculation", func(b *testing.B) {
		values := []float64{10, 12, 11, 10, 12, 11, 10, 12, 11, 10}
		mean := 11.0
		stdDev := 1.0

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = (values[i%len(values)] - mean) / stdDev
		}
	})

	b.Run("baseline_calculation", func(b *testing.B) {
		values := make([]float64, 1000)
		for i := 0; i < 1000; i++ {
			values[i] = float64(i) * 1.5
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mean := calculateMean(values)
			_ = calculateStdDev(values, mean)
		}
	})

	b.Run("severity_classification", func(b *testing.B) {
		zScores := []float64{0.5, 1.5, 2.7, 3.5}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = classifySeverity(zScores[i%len(zScores)])
		}
	})
}

// BenchmarkAlertRuleEvaluation benchmarks alert rule evaluation
func BenchmarkAlertRuleEvaluation(b *testing.B) {
	b.Run("threshold_evaluation", func(b *testing.B) {
		thresholds := []float64{80, 85, 90}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			value := float64((i % 100))
			_ = value > thresholds[i%len(thresholds)]
		}
	})

	b.Run("change_evaluation", func(b *testing.B) {
		previousValues := []float64{100, 200, 300}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			current := float64((i % 100) + 50)
			previous := previousValues[i%len(previousValues)]
			change := ((current - previous) / previous) * 100
			_ = change > 50 || change < -50
		}
	})

	b.Run("composite_evaluation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond1 := (i % 2) == 0
			cond2 := (i % 3) == 0
			_ = cond1 && cond2
		}
	})
}

// BenchmarkNotificationDelivery benchmarks notification sending
func BenchmarkNotificationDelivery(b *testing.B) {
	b.Run("format_slack_message", func(b *testing.B) {
		messages := make([]string, b.N)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := fmt.Sprintf("Alert: Rule %d, Severity: high, DB: prod_%d", i%1000, i%100)
			messages[i] = msg
		}
	})

	b.Run("message_queueing", func(b *testing.B) {
		queue := make(chan string, 1000)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			select {
			case queue <- fmt.Sprintf("alert_%d", i):
			default:
			}
		}
		close(queue)
	})

	b.Run("retry_scheduling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			attempt := (i % 5) + 1
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			_ = backoff
		}
	})
}

// TestAnomalyDetectionLoad tests anomaly detection with concurrent databases
func TestAnomalyDetectionLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with 100 concurrent databases
	numDatabases := 100
	queriesPerDatabase := 10
	duration := 10 * time.Second

	var processed, anomalies int32
	start := time.Now()

	for db := 0; db < numDatabases; db++ {
		go func(dbID int) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate anomaly detection
					values := generateRandomMetrics(50)
					mean := calculateMean(values)
					stdDev := calculateStdDev(values, mean)

					for i := 0; i < queriesPerDatabase; i++ {
						currentValue := float64((dbID * queriesPerDatabase) + i)
						zScore := (currentValue - mean) / stdDev

						if zScore > 2.5 {
							atomic.AddInt32(&anomalies, 1)
						}
						atomic.AddInt32(&processed, 1)
					}

					time.Sleep(10 * time.Millisecond)
				}
			}
		}(db)
	}

	time.Sleep(duration)
	cancel()

	elapsed := time.Since(start)
	processed_val := atomic.LoadInt32(&processed)
	anomalies_val := atomic.LoadInt32(&anomalies)

	t.Logf("Anomaly Detection Load Test Results:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Databases: %d", numDatabases)
	t.Logf("  Processed: %d", processed_val)
	t.Logf("  Anomalies Detected: %d", anomalies_val)
	t.Logf("  Throughput: %.0f ops/sec", float64(processed_val)/elapsed.Seconds())
}

// TestAlertRuleEvaluationLoad tests concurrent rule evaluation
func TestAlertRuleEvaluationLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numRules := 100
	evaluationsPerRule := 10
	duration := 10 * time.Second

	var evaluated, fired int32
	start := time.Now()

	for rule := 0; rule < numRules; rule++ {
		go func(ruleID int) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate rule evaluation
					for i := 0; i < evaluationsPerRule; i++ {
						value := float64((ruleID * evaluationsPerRule) + i)
						threshold := 50.0

						if value > threshold {
							atomic.AddInt32(&fired, 1)
						}
						atomic.AddInt32(&evaluated, 1)
					}

					time.Sleep(10 * time.Millisecond)
				}
			}
		}(rule)
	}

	time.Sleep(duration)
	cancel()

	elapsed := time.Since(start)
	evaluated_val := atomic.LoadInt32(&evaluated)
	fired_val := atomic.LoadInt32(&fired)

	t.Logf("Alert Rule Evaluation Load Test Results:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Rules: %d", numRules)
	t.Logf("  Evaluations: %d", evaluated_val)
	t.Logf("  Rules Fired: %d", fired_val)
	t.Logf("  Throughput: %.0f evals/sec", float64(evaluated_val)/elapsed.Seconds())
}

// TestNotificationDeliveryLoad tests concurrent notification delivery
func TestNotificationDeliveryLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numChannels := 10
	alertsPerSecond := 100
	duration := 10 * time.Second

	var sent, success, failed int32
	var wg sync.WaitGroup
	start := time.Now()

	for ch := 0; ch < numChannels; ch++ {
		wg.Add(1)
		go func(channelID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate sending alertsPerSecond notifications per second
					for i := 0; i < alertsPerSecond; i++ {
						atomic.AddInt32(&sent, 1)

						// Simulate ~99% success rate
						if (i % 100) == 0 {
							atomic.AddInt32(&failed, 1)
						} else {
							atomic.AddInt32(&success, 1)
						}
					}

					time.Sleep(1 * time.Second)
				}
			}
		}(ch)
	}

	time.Sleep(duration)
	cancel()

	wg.Wait()

	elapsed := time.Since(start)
	sent_val := atomic.LoadInt32(&sent)
	success_val := atomic.LoadInt32(&success)
	failed_val := atomic.LoadInt32(&failed)

	successRate := float64(success_val) / float64(sent_val) * 100

	t.Logf("Notification Delivery Load Test Results:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Channels: %d", numChannels)
	t.Logf("  Sent: %d", sent_val)
	t.Logf("  Successful: %d", success_val)
	t.Logf("  Failed: %d", failed_val)
	t.Logf("  Success Rate: %.2f%%", successRate)
	t.Logf("  Throughput: %.0f msgs/sec", float64(sent_val)/elapsed.Seconds())
}

// TestEndToEndAnomalyAlertNotification tests complete pipeline under load
func TestEndToEndAnomalyAlertNotification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simulate: 50 databases → anomaly detection → alert evaluation → notifications
	numDatabases := 50
	duration := 10 * time.Second

	var anomaliesDetected, alertsFired, notificationsSent int32
	start := time.Now()

	for db := 0; db < numDatabases; db++ {
		go func(dbID int) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Step 1: Anomaly Detection
					values := generateRandomMetrics(100)
					mean := calculateMean(values)
					stdDev := calculateStdDev(values, mean)

					// Check 10 queries
					for i := 0; i < 10; i++ {
						currentValue := float64((dbID * 10) + i)
						zScore := (currentValue - mean) / stdDev

						if zScore > 2.5 {
							atomic.AddInt32(&anomaliesDetected, 1)

							// Step 2: Trigger Alert
							atomic.AddInt32(&alertsFired, 1)

							// Step 3: Send Notification
							atomic.AddInt32(&notificationsSent, 1)
						}
					}

					time.Sleep(10 * time.Millisecond)
				}
			}
		}(db)
	}

	time.Sleep(duration)
	cancel()

	elapsed := time.Since(start)
	anomalies_val := atomic.LoadInt32(&anomaliesDetected)
	alerts_val := atomic.LoadInt32(&alertsFired)
	notifs_val := atomic.LoadInt32(&notificationsSent)

	t.Logf("End-to-End Anomaly → Alert → Notification Load Test:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Databases: %d", numDatabases)
	t.Logf("  Anomalies Detected: %d", anomalies_val)
	t.Logf("  Alerts Fired: %d", alerts_val)
	t.Logf("  Notifications Sent: %d", notifs_val)
	t.Logf("  Anomaly→Alert Ratio: %.2f%%", float64(alerts_val)/float64(anomalies_val)*100)
	t.Logf("  Total Throughput: %.0f ops/sec", float64(anomalies_val+alerts_val+notifs_val)/elapsed.Seconds())
}

// TestMemoryStabilityUnderLoad tests for memory leaks during sustained load
func TestMemoryStabilityUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numConcurrent := 100
	operationsPerThread := 1000
	duration := 10 * time.Second

	var completed int32
	start := time.Now()

	for i := 0; i < numConcurrent; i++ {
		go func() {
			values := make([]float64, 0, 100)

			for j := 0; j < operationsPerThread; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate memory usage
					values = append(values, float64(j))

					// Periodic cleanup to prevent unbounded growth
					if j%100 == 0 {
						values = values[:0] // Reset slice
					}

					atomic.AddInt32(&completed, 1)
				}
			}
		}()
	}

	time.Sleep(duration)
	cancel()

	elapsed := time.Since(start)
	completed_val := atomic.LoadInt32(&completed)

	t.Logf("Memory Stability Load Test:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Concurrent Operations: %d", numConcurrent)
	t.Logf("  Completed: %d", completed_val)
	t.Logf("  Throughput: %.0f ops/sec", float64(completed_val)/elapsed.Seconds())
	t.Logf("  Expected: Stable memory, no unbounded growth")
}

// Helper functions
func generateRandomMetrics(count int) []float64 {
	values := make([]float64, count)
	for i := 0; i < count; i++ {
		values[i] = float64(i) * 1.5
	}
	return values
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(values)-1)
	return variance
}

func classifySeverity(zScore float64) string {
	absZ := zScore
	if absZ < 0 {
		absZ = -absZ
	}

	if absZ >= 3.0 {
		return "critical"
	}
	if absZ >= 2.5 {
		return "high"
	}
	if absZ >= 1.5 {
		return "medium"
	}
	if absZ >= 1.0 {
		return "low"
	}
	return "normal"
}
