// Command validate_nist_go_vs_c is a validation tool that compares p-values
// produced by the pure-Go NIST SP 800-22 implementation against reference
// p-values from the NIST C reference implementation. It reads a binary or
// ASCII-encoded bitstream, runs each test, and reports per-test absolute
// differences against a configurable tolerance.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmmannChristian/nist-sp800-22-rev1a/internal/nist"
)

// testResult holds the comparison outcome for a single statistical test,
// including reference and Go p-values, their absolute differences, and
// whether the comparison passed the tolerance threshold.
type testResult struct {
	Name     string    `json:"name"`
	Ref      []float64 `json:"ref"`
	GoVals   []float64 `json:"go"`
	Diffs    []float64 `json:"diff"`
	Skipped  bool      `json:"skipped"`
	Reason   string    `json:"reason,omitempty"`
	PassDiff bool      `json:"pass_diff"`
}

func main() {
	var (
		dataset    = flag.String("dataset", "", "Path to binary dataset (NIST data.pi, etc)")
		bits       = flag.Int("bits", 1000000, "Number of bits to test")
		resultsDir = flag.String("results", "", "Path to NIST experiments/AlgorithmTesting directory")
		outJSON    = flag.Bool("json", false, "Print JSON output instead of table")
		tolerance  = flag.Float64("tolerance", 1e-6, "Absolute tolerance for p-value comparison")
		encoding   = flag.String("encoding", "ascii", "Input encoding: binary or ascii")
	)
	flag.Parse()

	if *dataset == "" || *resultsDir == "" {
		log.Fatalf("dataset and results are required")
	}

	rawData, err := os.ReadFile(*dataset)
	if err != nil {
		log.Fatalf("read dataset: %v", err)
	}

	data, err := parseBitstream(rawData, *encoding, *bits)
	if err != nil {
		log.Fatalf("parse bitstream: %v", err)
	}

	tests := []struct {
		name         string
		file         string
		compute      func([]byte) ([]float64, error)
		transformRef func([]float64) ([]float64, error)
		skipReason   string
	}{
		{"Frequency", "Frequency/results.txt", single(func(b []byte) (float64, bool) { return nist.FrequencyTest(b) }), nil, ""},
		{"BlockFrequency", "BlockFrequency/results.txt", single(func(b []byte) (float64, bool) { return nist.BlockFrequencyTest(b, 128) }), nil, ""},
		{"CumulativeSums", "CumulativeSums/results.txt", func(b []byte) ([]float64, error) {
			p, _ := nist.CumulativeSumsTest(b)
			return []float64{p}, nil // NIST logs forward+reverse; we compare on min value
		}, func(vals []float64) ([]float64, error) {
			if len(vals) == 0 {
				return nil, fmt.Errorf("no ref values")
			}
			return []float64{minValue(vals)}, nil
		}, ""},
		{"Runs", "Runs/results.txt", single(func(b []byte) (float64, bool) { return nist.RunsTest(b) }), nil, ""},
		{"LongestRun", "LongestRun/results.txt", single(func(b []byte) (float64, bool) { return nist.LongestRunOfOnesTest(b) }), nil, ""},
		{"Rank", "Rank/results.txt", single(func(b []byte) (float64, bool) { return nist.BinaryMatrixRankTest(b) }), nil, ""},
		{"FFT", "FFT/results.txt", single(func(b []byte) (float64, bool) { return nist.DiscreteFourierTransformTest(b) }), nil, ""},
		{"OverlappingTemplate", "OverlappingTemplate/results.txt", single(func(b []byte) (float64, bool) { return nist.OverlappingTemplateTest(b, 9) }), nil, ""},
		{"ApproximateEntropy", "ApproximateEntropy/results.txt", single(func(b []byte) (float64, bool) { return nist.ApproximateEntropyTest(b, 10) }), nil, ""},
		{"Universal", "Universal/results.txt", single(func(b []byte) (float64, bool) { return nist.UniversalStatisticalTest(b) }), nil, ""},
		{"LinearComplexity", "LinearComplexity/results.txt", single(func(b []byte) (float64, bool) { return nist.LinearComplexityTest(b, 500) }), nil, ""},
		{"Serial", "Serial/results.txt", single(func(b []byte) (float64, bool) { return nist.SerialTest(b, 16) }), func(vals []float64) ([]float64, error) {
			if len(vals) == 0 {
				return nil, fmt.Errorf("no ref values")
			}
			return []float64{minValue(vals)}, nil
		}, ""},
		{"NonOverlappingTemplate", "NonOverlappingTemplate/results.txt", single(func(b []byte) (float64, bool) { return nist.NonOverlappingTemplateTest(b, 9) }), func(vals []float64) ([]float64, error) {
			if len(vals) == 0 {
				return nil, fmt.Errorf("no ref values")
			}
			return []float64{minValue(vals)}, nil
		}, ""},
		{"RandomExcursions", "RandomExcursions/results.txt", single(func(b []byte) (float64, bool) { return nist.RandomExcursionsTest(b) }), func(vals []float64) ([]float64, error) {
			if len(vals) == 0 {
				return nil, fmt.Errorf("no ref values")
			}
			return []float64{minValue(vals)}, nil
		}, ""},
		{"RandomExcursionsVariant", "RandomExcursionsVariant/results.txt", single(func(b []byte) (float64, bool) { return nist.RandomExcursionsVariantTest(b) }), func(vals []float64) ([]float64, error) {
			if len(vals) == 0 {
				return nil, fmt.Errorf("no ref values")
			}
			return []float64{minValue(vals)}, nil
		}, ""},
	}

	var results []testResult
	allPass := true

	for _, tt := range tests {
		res := testResult{Name: tt.name}
		if tt.compute == nil {
			res.Skipped = true
			res.Reason = tt.skipReason
			results = append(results, res)
			continue
		}

		refPath := filepath.Join(*resultsDir, tt.file)
		refVals, err := readRef(refPath)
		if err != nil {
			res.Skipped = true
			res.Reason = fmt.Sprintf("read ref: %v", err)
			results = append(results, res)
			continue
		}

		if tt.transformRef != nil {
			refVals, err = tt.transformRef(refVals)
			if err != nil {
				res.Skipped = true
				res.Reason = fmt.Sprintf("transform ref: %v", err)
				results = append(results, res)
				continue
			}
		}

		goVals, err := tt.compute(data)
		if err != nil {
			res.Skipped = true
			res.Reason = fmt.Sprintf("compute go: %v", err)
			results = append(results, res)
			continue
		}

		res.Ref = refVals
		res.GoVals = goVals
		res.Diffs = make([]float64, min(len(refVals), len(goVals)))
		res.PassDiff = true
		for i := range res.Diffs {
			res.Diffs[i] = math.Abs(refVals[i] - goVals[i])
			if res.Diffs[i] > *tolerance || !validRange(refVals[i]) || !validRange(goVals[i]) {
				res.PassDiff = false
			}
		}
		allPass = allPass && res.PassDiff
		results = append(results, res)
	}

	if *outJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			log.Fatalf("encode json: %v", err)
		}
		return
	}

	fmt.Printf("Dataset: %s (%d bits) | Reference dir: %s | tolerance: %g\n", *dataset, *bits, *resultsDir, *tolerance)
	fmt.Println("Test                      Go p-value        Ref p-value       |diff|      Status")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, r := range results {
		if r.Skipped {
			fmt.Printf("%-24s %-16s %-16s %-12s SKIPPED (%s)\n", r.Name, "-", "-", "-", r.Reason)
			continue
		}
		goStr := fmtVal(r.GoVals)
		refStr := fmtVal(r.Ref)
		diffStr := fmtVal(r.Diffs)
		status := "OK"
		if !r.PassDiff {
			status = "MISMATCH"
		}
		fmt.Printf("%-24s %-16s %-16s %-12s %s\n", r.Name, goStr, refStr, diffStr, status)
	}
	if !allPass {
		os.Exit(1)
	}
}

// single adapts a function that returns a single p-value into the multi-value
// signature expected by the test table, discarding the pass/fail boolean.
func single(f func([]byte) (float64, bool)) func([]byte) ([]float64, error) {
	return func(b []byte) ([]float64, error) {
		p, _ := f(b)
		return []float64{p}, nil
	}
}

func readRef(path string) ([]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var vals []float64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var v float64
		if _, err := fmt.Sscanf(line, "%f", &v); err != nil {
			return nil, fmt.Errorf("parse %q: %w", line, err)
		}
		vals = append(vals, v)
	}
	return vals, scanner.Err()
}

func fmtVal(vals []float64) string {
	if len(vals) == 0 {
		return "-"
	}
	if len(vals) == 1 {
		return fmt.Sprintf("%.6f", vals[0])
	}
	return fmt.Sprintf("%.6f…", vals[0])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func validRange(v float64) bool {
	return !math.IsNaN(v) && v >= 0 && v <= 1
}

func minValue(vals []float64) float64 {
	min := math.Inf(1)
	for _, v := range vals {
		if v < min {
			min = v
		}
	}
	return min
}

func parseBitstream(data []byte, encoding string, bits int) ([]byte, error) {
	if encoding == "binary" {
		requiredBytes := (bits + 7) / 8
		if len(data) < requiredBytes {
			return nil, fmt.Errorf("dataset too small: need %d bytes, have %d", requiredBytes, len(data))
		}
		return data[:requiredBytes], nil
	}

	if encoding == "ascii" {
		var packed []byte
		var currentByte byte
		var bitCount int
		var totalBits int

		for _, b := range data {
			if totalBits >= bits {
				break
			}
			if b == '0' || b == '1' {
				val := byte(b - '0')
				// Pack into currentByte, MSB first
				// bitCount 0 -> shift 7
				// bitCount 1 -> shift 6
				currentByte |= val << (7 - bitCount)
				bitCount++
				totalBits++

				if bitCount == 8 {
					packed = append(packed, currentByte)
					currentByte = 0
					bitCount = 0
				}
			}
			// Ignore whitespace and other characters
		}

		if totalBits < bits {
			return nil, fmt.Errorf("dataset too small: found %d bits, need %d", totalBits, bits)
		}

		// Append last partial byte if needed
		if bitCount > 0 {
			packed = append(packed, currentByte)
		}

		return packed, nil
	}

	return nil, fmt.Errorf("unknown encoding: %s", encoding)
}
