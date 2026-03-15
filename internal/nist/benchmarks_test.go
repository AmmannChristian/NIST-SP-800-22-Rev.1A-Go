package nist

import (
	"crypto/rand"
	"testing"
)

// BenchmarkFrequencyTest benchmarks the Frequency (Monobit) test
func BenchmarkFrequencyTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FrequencyTest(bits)
	}
}

// BenchmarkBlockFrequencyTest benchmarks the Block Frequency test
func BenchmarkBlockFrequencyTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BlockFrequencyTest(bits, 128)
	}
}

// BenchmarkCumulativeSumsTest benchmarks the Cumulative Sums test
func BenchmarkCumulativeSumsTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CumulativeSumsTest(bits)
	}
}

// BenchmarkRunsTest benchmarks the Runs test
func BenchmarkRunsTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunsTest(bits)
	}
}

// BenchmarkLongestRunOfOnesTest benchmarks the Longest Run of Ones test
func BenchmarkLongestRunOfOnesTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LongestRunOfOnesTest(bits)
	}
}

// BenchmarkBinaryMatrixRankTest benchmarks the Binary Matrix Rank test
func BenchmarkBinaryMatrixRankTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BinaryMatrixRankTest(bits)
	}
}

// BenchmarkDiscreteFourierTransformTest benchmarks the DFT test
func BenchmarkDiscreteFourierTransformTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DiscreteFourierTransformTest(bits)
	}
}

// BenchmarkNonOverlappingTemplateTest benchmarks the Non-overlapping Template test
func BenchmarkNonOverlappingTemplateTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NonOverlappingTemplateTest(bits, 9)
	}
}

// BenchmarkOverlappingTemplateTest benchmarks the Overlapping Template test
func BenchmarkOverlappingTemplateTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OverlappingTemplateTest(bits, 9)
	}
}

// BenchmarkUniversalStatisticalTest benchmarks the Universal Statistical test
func BenchmarkUniversalStatisticalTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UniversalStatisticalTest(bits)
	}
}

// BenchmarkApproximateEntropyTest benchmarks the Approximate Entropy test
func BenchmarkApproximateEntropyTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApproximateEntropyTest(bits, 10)
	}
}

// BenchmarkRandomExcursionsTest benchmarks the Random Excursions test
func BenchmarkRandomExcursionsTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RandomExcursionsTest(bits)
	}
}

// BenchmarkRandomExcursionsVariantTest benchmarks the Random Excursions Variant test
func BenchmarkRandomExcursionsVariantTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RandomExcursionsVariantTest(bits)
	}
}

// BenchmarkSerialTest benchmarks the Serial test
func BenchmarkSerialTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SerialTest(bits, 16)
	}
}

// BenchmarkLinearComplexityTest benchmarks the Linear Complexity test
func BenchmarkLinearComplexityTest(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LinearComplexityTest(bits, 500)
	}
}

// BenchmarkRunAllTests benchmarks the complete NIST test suite
func BenchmarkRunAllTests(b *testing.B) {
	bits := make([]byte, 125000) // 1M bits
	rand.Read(bits)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunAllTests(bits)
	}
}

// BenchmarkRunAllTests_Sizes benchmarks with different input sizes
func BenchmarkRunAllTests_Sizes(b *testing.B) {
	sizes := []struct {
		name string
		bits int
	}{
		{"387840_bits", 48480},  // Minimum required
		{"1M_bits", 125000},     // 1 million bits
		{"5M_bits", 625000},     // 5 million bits
		{"10M_bits", 1250000},   // Previous maximum
		{"30M_bits", 3750000},   // Typical 1-hour window
		{"50M_bits", 6250000},   // Stress test
		{"100M_bits", 12500000}, // New maximum
	}

	for _, size := range sizes {
		bits := make([]byte, size.bits)
		rand.Read(bits)

		b.Run(size.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				RunAllTests(bits)
			}
		})
	}
}

// BenchmarkIndividualTests_30M benchmarks each test individually at 30M bits (typical 1-hour window)
// to identify which tests scale worst at this size.
func BenchmarkIndividualTests_30M(b *testing.B) {
	bits := make([]byte, 3750000) // 30M bits
	rand.Read(bits)

	tests := []struct {
		name string
		fn   func()
	}{
		{"FrequencyTest", func() { FrequencyTest(bits) }},
		{"BlockFrequencyTest", func() { BlockFrequencyTest(bits, 128) }},
		{"CumulativeSumsTest", func() { CumulativeSumsTest(bits) }},
		{"RunsTest", func() { RunsTest(bits) }},
		{"LongestRunOfOnesTest", func() { LongestRunOfOnesTest(bits) }},
		{"BinaryMatrixRankTest", func() { BinaryMatrixRankTest(bits) }},
		{"DiscreteFourierTransformTest", func() { DiscreteFourierTransformTest(bits) }},
		{"NonOverlappingTemplateTest", func() { NonOverlappingTemplateTest(bits, 9) }},
		{"OverlappingTemplateTest", func() { OverlappingTemplateTest(bits, 9) }},
		{"UniversalStatisticalTest", func() { UniversalStatisticalTest(bits) }},
		{"ApproximateEntropyTest", func() { ApproximateEntropyTest(bits, 10) }},
		{"RandomExcursionsTest", func() { RandomExcursionsTest(bits) }},
		{"RandomExcursionsVariantTest", func() { RandomExcursionsVariantTest(bits) }},
		{"SerialTest", func() { SerialTest(bits, 16) }},
		{"LinearComplexityTest", func() { LinearComplexityTest(bits, 500) }},
	}

	for _, t := range tests {
		fn := t.fn
		b.Run(t.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fn()
			}
		})
	}
}
