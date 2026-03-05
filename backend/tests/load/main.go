package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/torresglauco/pganalytics-v3/backend/tests/load"
)

func main() {
	// Parse command-line flags
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the API")
	numCollectors := flag.Int("collectors", 500, "Number of simulated collectors")
	metricsPerPush := flag.Int("metrics", 10, "Number of metrics per push")
	pushInterval := flag.Int("interval", 5, "Seconds between metric pushes")
	duration := flag.Int("duration", 5, "Test duration in minutes")
	concurrent := flag.Int("concurrent", 10, "Concurrent metric pushes")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	fmt.Println("\n╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    pgAnalytics Phase 4 - Load Test Suite                     ║")
	fmt.Println("║                   500+ Collectors Scalability Validation                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝\n")

	// Validate inputs
	if *baseURL == "" {
		log.Fatal("Error: --url is required")
	}
	if *numCollectors <= 0 {
		log.Fatal("Error: --collectors must be > 0")
	}
	if *duration <= 0 {
		log.Fatal("Error: --duration must be > 0")
	}

	// Create configuration
	config := &load.LoadTestConfig{
		BaseURL:             *baseURL,
		NumCollectors:       *numCollectors,
		MetricsPerCollector: *metricsPerPush,
		PushIntervalSeconds: *pushInterval,
		DurationMinutes:     *duration,
		ConcurrentPushes:    *concurrent,
		EnableRateLimitTest: true,
		EnableCacheTest:     true,
		Verbose:             *verbose,
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  API URL:           %s\n", config.BaseURL)
	fmt.Printf("  Collectors:        %d\n", config.NumCollectors)
	fmt.Printf("  Metrics/Push:      %d\n", config.MetricsPerCollector)
	fmt.Printf("  Push Interval:     %d seconds\n", config.PushIntervalSeconds)
	fmt.Printf("  Test Duration:     %d minutes\n", config.DurationMinutes)
	fmt.Printf("  Concurrent Pushes: %d\n", config.ConcurrentPushes)
	fmt.Printf("  Verbose:           %v\n\n", config.Verbose)

	fmt.Println("Press Enter to start the load test...")
	fmt.Scanln()

	// Create and run load tester
	tester := load.NewLoadTester(config)
	results := tester.Run()

	// Print results
	results.PrintResults()

	// Exit with appropriate code based on results
	if results.P95Latency.Milliseconds() <= 500 &&
		results.ErrorRate < 0.1 &&
		results.CacheHitRate > 75 {
		fmt.Println("✅ All success criteria passed!")
		os.Exit(0)
	} else {
		fmt.Println("❌ Some success criteria failed. Review results above.")
		os.Exit(1)
	}
}
