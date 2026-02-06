package nist

import (
	"math"

	"gonum.org/v1/gonum/mathext"
)

// LongestRunOfOnesTest implements the NIST Longest Run of Ones test.
// It returns the p-value and whether it passes at Alpha.
func LongestRunOfOnesTest(bitstream []byte) (float64, bool) {
	bits := expandBits(bitstream)
	n := len(bits)
	if n < 128 {
		return 0, false
	}

	// K, M, V, and pi are chosen based on the input length n, following the
	// NIST specification. K is the number of categories minus one, M is the
	// block size, V holds the longest-run bin boundaries, and pi contains the
	// theoretical probabilities for each bin under the null hypothesis.
	var K, M int
	var V []int
	var pi []float64

	switch {
	case n < 6272:
		K, M = 3, 8
		V = []int{1, 2, 3, 4}
		pi = []float64{0.21484375, 0.3671875, 0.23046875, 0.1875}
	case n < 750000:
		K, M = 5, 128
		V = []int{4, 5, 6, 7, 8, 9}
		pi = []float64{0.1174035788, 0.242955959, 0.249363483, 0.17517706, 0.102701071, 0.112398847}
	default:
		K, M = 6, 10000
		V = []int{10, 11, 12, 13, 14, 15, 16}
		pi = []float64{0.0882, 0.2092, 0.2483, 0.1933, 0.1208, 0.0675, 0.0727}
	}

	N := n / M
	if N == 0 {
		return 0, false
	}

	nu := make([]float64, K+1)
	for block := 0; block < N; block++ {
		longest := 0
		run := 0
		base := block * M
		for j := 0; j < M; j++ {
			if bits[base+j] == 1 {
				run++
				if run > longest {
					longest = run
				}
			} else {
				run = 0
			}
		}

		switch {
		case longest < V[0]:
			nu[0]++
		case longest > V[K]:
			nu[K]++
		default:
			for i := 0; i <= K; i++ {
				if longest == V[i] {
					nu[i]++
					break
				}
			}
		}
	}

	var chiSquared float64
	for i := 0; i <= K; i++ {
		chiSquared += math.Pow(nu[i]-float64(N)*pi[i], 2) / (float64(N) * pi[i])
	}

	pValue := mathext.GammaIncRegComp(float64(K)/2.0, chiSquared/2.0)

	return pValue, pValue >= Alpha
}
