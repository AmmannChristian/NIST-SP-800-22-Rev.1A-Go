package nist

import (
	"fmt"
)

// TestResult represents the outcome of a single NIST SP 800-22 statistical test.
type TestResult struct {
	// Name is the snake_case identifier of the test (e.g. "frequency_monobit").
	Name string
	// PValue is the computed p-value; values below Alpha indicate non-randomness.
	PValue float64
	// Passed indicates whether the p-value meets or exceeds the Alpha threshold.
	Passed bool
	// Proportion is 1.0 if the test passed, 0.0 otherwise (single-sequence mode).
	Proportion float64
	// Warning contains a diagnostic message when the test could not run normally,
	// such as insufficient input bits. It is empty when no issues occurred.
	Warning string
}

const (
	// MinBits is the minimum required bits for the full 15-test suite (Universal test).
	MinBits = 387840
	// MaxBits is a safety cap to avoid unbounded allocations.
	MaxBits = 10000000
)

// RunAllTests executes all 15 NIST SP 800-22 Rev 1a statistical tests against
// the provided bitstream and returns one TestResult per test. The bitstream
// length must be between MinBits and MaxBits (in units of 8 bits). An error is
// returned if the bitstream length is outside these bounds.
func RunAllTests(bitstream []byte) ([]TestResult, error) {
	numBits := len(bitstream) * 8
	if numBits < MinBits {
		return nil, fmt.Errorf("insufficient bits: got %d, need at least %d", numBits, MinBits)
	}
	if numBits > MaxBits {
		return nil, fmt.Errorf("too many bits: got %d, maximum %d", numBits, MaxBits)
	}

	results := make([]TestResult, 0, 15)

	appendResult := func(name string, p float64, passed bool, warning string) {
		r := TestResult{
			Name:       name,
			PValue:     p,
			Passed:     passed,
			Proportion: 0,
			Warning:    warning,
		}
		if passed {
			r.Proportion = 1.0
		}
		results = append(results, r)
	}

	// 1. Frequency (Monobit)
	p, pass := FrequencyTest(bitstream)
	appendResult("frequency_monobit", p, pass, "")

	// 2. Block Frequency (M = 128)
	p, pass = BlockFrequencyTest(bitstream, 128)
	warn := ""
	if !pass && p == 0 {
		warn = "insufficient bits for block size"
	}
	appendResult("block_frequency", p, pass, warn)

	// 3. Cumulative Sums
	p, pass = CumulativeSumsTest(bitstream)
	appendResult("cumulative_sums", p, pass, "")

	// 4. Runs
	p, pass = RunsTest(bitstream)
	warn = ""
	if p == 0 && !pass {
		warn = "Pi estimator criteria not met"
	}
	appendResult("runs", p, pass, warn)

	// 5. Longest Run of Ones
	p, pass = LongestRunOfOnesTest(bitstream)
	warn = ""
	if p == 0 && !pass {
		warn = "insufficient bits for test"
	}
	appendResult("longest_run", p, pass, warn)

	// 6. Binary Matrix Rank
	p, pass = BinaryMatrixRankTest(bitstream)
	warn = ""
	if p == 0 && !pass {
		warn = "insufficient bits for 32x32 matrices"
	}
	appendResult("binary_matrix_rank", p, pass, warn)

	// 7. Discrete Fourier Transform
	p, pass = DiscreteFourierTransformTest(bitstream)
	appendResult("discrete_fourier_transform", p, pass, "")

	// 8. Non-overlapping Template (m = 9)
	p, pass = NonOverlappingTemplateTest(bitstream, 9)
	warn = ""
	if p == 0 && !pass {
		warn = "only m=9 supported or insufficient bits"
	}
	appendResult("non_overlapping_template", p, pass, warn)

	// 9. Overlapping Template (m = 9)
	p, pass = OverlappingTemplateTest(bitstream, 9)
	appendResult("overlapping_template", p, pass, "")

	// 10. Universal Statistical
	p, pass = UniversalStatisticalTest(bitstream)
	warn = ""
	if p == 0 && !pass {
		warn = "insufficient bits or invalid parameters"
	}
	appendResult("universal_statistical", p, pass, warn)

	// 11. Approximate Entropy (m = 10)
	p, pass = ApproximateEntropyTest(bitstream, 10)
	appendResult("approximate_entropy", p, pass, "")

	// 12. Random Excursions
	p, pass = RandomExcursionsTest(bitstream)
	warn = ""
	if p == 0 && !pass {
		warn = "insufficient cycles (J < 500) or fail"
	}
	appendResult("random_excursions", p, pass, warn)

	// 13. Random Excursions Variant
	p, pass = RandomExcursionsVariantTest(bitstream)
	warn = ""
	if p == 0 && !pass {
		warn = "insufficient cycles (J < 500) or fail"
	}
	appendResult("random_excursions_variant", p, pass, warn)

	// 14. Serial (m = 16)
	p, pass = SerialTest(bitstream, 16)
	appendResult("serial", p, pass, "")

	// 15. Linear Complexity (M = 500)
	p, pass = LinearComplexityTest(bitstream, 500)
	appendResult("linear_complexity", p, pass, "")

	return results, nil
}
