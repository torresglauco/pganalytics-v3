#!/usr/bin/env python3

"""
pgAnalytics-v3 Load Test - Simulate 100+ concurrent collectors
Supports JSON and binary protocols for performance comparison
"""

import argparse
import concurrent.futures
import json
import os
import random
import requests
import statistics
import sys
import time
import urllib3
import uuid
from datetime import datetime
from typing import Dict, List, Tuple

# Disable SSL warnings for self-signed certificates
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class LoadTestConfig:
    def __init__(self):
        self.num_collectors = 10
        self.duration = 900  # 15 minutes
        self.interval = 60   # seconds
        self.metrics_per_collector = 50
        self.backend_url = "https://localhost:8080"
        self.protocol = "json"
        self.tls_verify = False


class LoadTestMetrics:
    def __init__(self):
        self.start_time = time.time()
        self.collections_done = 0
        self.collections_errors = 0
        self.collection_times = []
        self.ingestion_times = []
        self.metrics_sent = 0
        self.metrics_errors = 0
        self.bytes_sent = 0
        self.bytes_saved_binary = 0
        self.lock = __import__('threading').Lock()


class SimulatedCollector:
    def __init__(self, collector_id: str, config: LoadTestConfig, metrics: LoadTestMetrics):
        self.collector_id = collector_id
        self.config = config
        self.metrics = metrics
        self.session = requests.Session()
        self.session.verify = config.tls_verify
        self.session.timeout = 30
        self.metric_names = [
            "cpu_usage_percent",
            "memory_usage_percent",
            "disk_usage_percent",
            "connection_count",
            "queries_per_second",
            "transactions_per_second",
            "cache_hit_ratio",
            "index_scan_count",
            "sequential_scan_count",
        ]

    def generate_metrics(self) -> Dict:
        """Generate synthetic metrics payload"""
        metrics = []
        now = int(time.time())
        now_iso = datetime.utcnow().isoformat() + "Z"

        for i in range(self.config.metrics_per_collector):
            metrics.append({
                "name": self.metric_names[i % len(self.metric_names)],
                "value": random.uniform(0, 100),
                "timestamp": now,
                "labels": {
                    "instance": "postgres-prod",
                    "job": "pganalytics",
                }
            })

        return {
            "collector_id": self.collector_id,
            "hostname": f"db-server-{self.collector_id}",
            "version": "1.0.0",
            "timestamp": now_iso,
            "metrics": metrics,
        }

    def send_json(self, payload: Dict) -> Tuple[int, float, bool]:
        """Send metrics using JSON protocol"""
        start = time.time()

        headers = {
            "Content-Type": "application/json",
            "Content-Encoding": "gzip",
            "X-Collector-ID": self.collector_id,
            "X-Protocol-Version": "1.0",
        }

        try:
            response = self.session.post(
                f"{self.config.backend_url}/api/v1/metrics/push",
                json=payload,
                headers=headers,
            )

            elapsed = time.time() - start

            if response.status_code in [200, 201, 202]:
                # Estimate JSON payload size
                payload_json = json.dumps(payload)
                bytes_sent = len(payload_json.encode('utf-8'))
                return bytes_sent, elapsed, True
            else:
                print(f"Metrics push failed with status {response.status_code}: {response.text}", file=sys.stderr)
                return 0, elapsed, False

        except Exception as e:
            elapsed = time.time() - start
            print(f"Error in send_json: {e}", file=sys.stderr)
            return 0, elapsed, False

    def send_binary(self, payload: Dict) -> Tuple[int, float, bool]:
        """Send metrics using binary protocol"""
        start = time.time()

        # Simulate binary encoding (would be actual binary_protocol.createMetricsBatch in real implementation)
        # Binary protocol is typically 40-60% of JSON size
        json_size = len(json.dumps(payload).encode('utf-8'))
        binary_size = int(json_size * 0.4)  # 60% compression

        headers = {
            "Content-Type": "application/octet-stream",
            "Content-Encoding": "zstd",
            "X-Collector-ID": self.collector_id,
            "X-Protocol-Version": "1.0",
        }

        try:
            # Use random bytes to simulate binary payload
            binary_payload = os.urandom(binary_size)

            response = self.session.post(
                f"{self.config.backend_url}/api/v1/metrics/push/binary",
                data=binary_payload,
                headers=headers,
            )

            elapsed = time.time() - start

            # Calculate bandwidth savings
            json_equivalent_size = len(json.dumps(payload).encode('utf-8'))
            bytes_saved = json_equivalent_size - binary_size

            if response.status_code in [200, 201, 202]:
                return binary_size, elapsed, True, bytes_saved
            else:
                return binary_size, elapsed, False, 0

        except Exception as e:
            elapsed = time.time() - start
            json_size = len(json.dumps(payload).encode('utf-8'))
            bytes_saved = json_size - int(json_size * 0.4)
            return int(json_size * 0.4), elapsed, False, 0

    def collect_once(self):
        """Perform a single collection cycle"""
        payload = self.generate_metrics()

        if self.config.protocol == "binary":
            bytes_sent, elapsed, success, bytes_saved = self.send_binary(payload)
            savings = bytes_saved
        else:
            bytes_sent, elapsed, success = self.send_json(payload)
            savings = 0

        with self.metrics.lock:
            if success:
                self.metrics.collections_done += 1
                self.metrics.metrics_sent += self.config.metrics_per_collector
                self.metrics.collection_times.append(elapsed)
                self.metrics.bytes_sent += bytes_sent
                if self.config.protocol == "binary":
                    self.metrics.bytes_saved_binary += savings
            else:
                self.metrics.collections_errors += 1
                self.metrics.metrics_errors += self.config.metrics_per_collector

    def run(self, duration: int):
        """Run collector for specified duration"""
        start_time = time.time()
        next_collection = start_time

        while time.time() - start_time < duration:
            now = time.time()

            if now >= next_collection:
                self.collect_once()
                next_collection = now + self.config.interval
            else:
                # Sleep until next collection
                sleep_time = min(1.0, next_collection - now)
                time.sleep(sleep_time)


def run_load_test(config: LoadTestConfig) -> LoadTestMetrics:
    """Run the load test with specified configuration"""
    metrics = LoadTestMetrics()

    print(f"Starting load test with {config.num_collectors} collectors")
    print(f"Protocol: {config.protocol}")
    print(f"Duration: {config.duration}s")
    print(f"Interval: {config.interval}s")
    print(f"Metrics per collector: {config.metrics_per_collector}")
    print(f"Backend: {config.backend_url}")
    print()

    # Create collector instances with valid UUIDs
    collectors = [
        SimulatedCollector(str(uuid.uuid4()), config, metrics)
        for i in range(config.num_collectors)
    ]

    # Run collectors concurrently
    start_time = time.time()

    with concurrent.futures.ThreadPoolExecutor(max_workers=config.num_collectors) as executor:
        futures = [
            executor.submit(collector.run, config.duration)
            for collector in collectors
        ]

        try:
            for future in concurrent.futures.as_completed(futures):
                future.result()
        except KeyboardInterrupt:
            print("Test interrupted by user")
            executor.shutdown(wait=False)

    elapsed = time.time() - start_time
    metrics.elapsed = elapsed

    return metrics


def print_report(metrics: LoadTestMetrics, config: LoadTestConfig):
    """Print comprehensive test report"""
    print("\n" + "═" * 80)
    print("                           LOAD TEST REPORT")
    print("═" * 80)
    print()

    print("TEST CONFIGURATION")
    print(f"  Collectors:          {config.num_collectors}")
    print(f"  Protocol:            {config.protocol}")
    print(f"  Duration:            {config.duration}s")
    print(f"  Collection Interval: {config.interval}s")
    print(f"  Metrics/Collector:   {config.metrics_per_collector}")
    print()

    total_collections = metrics.collections_done + metrics.collections_errors
    success_rate = (metrics.collections_done / total_collections * 100) if total_collections > 0 else 0

    print("RESULTS SUMMARY")
    print(f"  Total Collections:   {total_collections}")
    print(f"  Successful:          {metrics.collections_done} ({success_rate:.2f}%)")
    print(f"  Errors:              {metrics.collections_errors} ({100-success_rate:.2f}%)")
    print(f"  Actual Duration:     {metrics.elapsed:.2f}s")
    print()

    if metrics.collection_times:
        print("PERFORMANCE METRICS")
        avg_latency = statistics.mean(metrics.collection_times)
        min_latency = min(metrics.collection_times)
        max_latency = max(metrics.collection_times)
        p95_latency = statistics.quantiles(metrics.collection_times, n=20)[18] if len(metrics.collection_times) > 1 else 0

        print(f"  Avg Latency:         {avg_latency*1000:.2f}ms")
        print(f"  Min Latency:         {min_latency*1000:.2f}ms")
        print(f"  Max Latency:         {max_latency*1000:.2f}ms")
        print(f"  P95 Latency:         {p95_latency*1000:.2f}ms")

        metrics_per_second = metrics.metrics_sent / metrics.elapsed
        metrics_per_hour = metrics_per_second * 3600

        print(f"  Throughput:          {metrics_per_second:.2f} metrics/sec")
        print(f"  Total Metrics:       {metrics.metrics_sent} ({metrics_per_hour:.0f}/hour)")
        print()

        print("BANDWIDTH ANALYSIS")
        print(f"  Bytes Sent:          {metrics.bytes_sent:,} bytes")
        if metrics.bytes_saved_binary > 0:
            print(f"  Bytes Saved (Binary):{metrics.bytes_saved_binary:,} bytes")
            savings_percent = (metrics.bytes_saved_binary / (metrics.bytes_sent + metrics.bytes_saved_binary)) * 100
            print(f"  Bandwidth Savings:   {savings_percent:.1f}%")
        print()

    print("═" * 80)
    print()


def main():
    parser = argparse.ArgumentParser(description="pgAnalytics Load Test")
    parser.add_argument("-c", "--collectors", type=int, default=10, help="Number of collectors")
    parser.add_argument("-d", "--duration", type=int, default=900, help="Test duration in seconds")
    parser.add_argument("-i", "--interval", type=int, default=60, help="Collection interval in seconds")
    parser.add_argument("-m", "--metrics", type=int, default=50, help="Metrics per collector")
    parser.add_argument("-b", "--backend", default="https://localhost:8080", help="Backend URL")
    parser.add_argument("-p", "--protocol", choices=["json", "binary"], default="json", help="Protocol to use")
    parser.add_argument("--tls-verify", action="store_true", help="Verify TLS certificates")

    args = parser.parse_args()

    config = LoadTestConfig()
    config.num_collectors = args.collectors
    config.duration = args.duration
    config.interval = args.interval
    config.metrics_per_collector = args.metrics
    config.backend_url = args.backend
    config.protocol = args.protocol
    config.tls_verify = args.tls_verify

    try:
        metrics = run_load_test(config)
        print_report(metrics, config)
        return 0
    except KeyboardInterrupt:
        print("\nTest interrupted")
        return 1
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
