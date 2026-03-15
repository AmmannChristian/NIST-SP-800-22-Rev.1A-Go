#!/usr/bin/env bash
# Phase 0: SP 800-22 Feasibility Benchmark
#
# Runs the full 15-test suite at 10M, 30M, 50M, and 100M bits to measure
# wall-clock runtime and peak RSS per test. This is the mandatory go/no-go
# gate for single-sequence mode.
#
# Go/no-go criteria:
#   30M bits: < 5 min, < 500 MB peak RSS
#   100M bits: < 10 min, < 2 GB peak RSS
#
# Usage:
#   cd nist-800-22-test-suite && ./scripts/benchmark-feasibility.sh

set -euo pipefail

cd "$(dirname "$0")/.."

echo "=== Phase 0: SP 800-22 Feasibility Benchmark ==="
echo "Date: $(date -Iseconds)"
echo "Go version: $(go version)"
echo ""

echo "--- Full suite benchmarks at multiple sizes (1 iteration each) ---"
echo "Running BenchmarkRunAllTests_Sizes with -benchtime=1x..."
echo ""

go test -bench='BenchmarkRunAllTests_Sizes' \
  -benchtime=1x \
  -benchmem \
  -timeout=30m \
  -v \
  ./internal/nist/

echo ""
echo "--- Per-test breakdown at 30M bits ---"
echo "Running BenchmarkIndividualTests_30M with -benchtime=1x..."
echo ""

go test -bench='BenchmarkIndividualTests_30M' \
  -benchtime=1x \
  -benchmem \
  -timeout=30m \
  -v \
  ./internal/nist/

echo ""
echo "=== Benchmark complete ==="
echo ""
echo "Review the results above against go/no-go criteria:"
echo "  30M bits: < 5 min wall-clock, < 500 MB peak RSS"
echo "  100M bits: < 10 min wall-clock, < 2 GB peak RSS"
