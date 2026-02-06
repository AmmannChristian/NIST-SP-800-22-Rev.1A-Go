package nist

import (
	"math"

	"gonum.org/v1/gonum/mathext"
)

// OverlappingTemplateTest implements the NIST Overlapping Template Matching test.
// It returns the p-value and whether it passes at Alpha.
func OverlappingTemplateTest(bitstream []byte, m int) (float64, bool) {
	bits := expandBits(bitstream)
	n := len(bits)
	if n < m {
		return 0, false
	}

	const K = 5
	M := 1032
	N := n / M
	if N == 0 {
		return 0, false
	}

	lambda := float64(M-m+1) / math.Pow(2, float64(m))
	eta := lambda / 2.0

	pi := make([]float64, K+1)
	sum := 0.0
	for i := 0; i < K; i++ {
		pi[i] = prHelper(i, eta)
		sum += pi[i]
	}
	pi[K] = 1 - sum

	nu := make([]int, K+1)
	for block := 0; block < N; block++ {
		wObs := 0
		for j := 0; j < M-m+1; j++ {
			match := true
			for k := 0; k < m; k++ {
				if bits[block*M+j+k] == 0 {
					match = false
					break
				}
			}
			if match {
				wObs++
			}
		}

		if wObs <= 4 {
			nu[wObs]++
		} else {
			nu[K]++
		}
	}

	chi2 := 0.0
	for i := 0; i < K+1; i++ {
		expected := float64(N) * pi[i]
		diff := float64(nu[i]) - expected
		chi2 += diff * diff / expected
	}

	pValue := mathext.GammaIncRegComp(float64(K)/2.0, chi2/2.0)
	return pValue, pValue >= Alpha
}

// prHelper computes the probability Pr(u) for the overlapping template
// matching test, where u is the observed match count and eta is half the
// expected number of matches per block. The calculation uses a recursive
// combinatorial formula involving log-gamma terms to avoid overflow.
func prHelper(u int, eta float64) float64 {
	if u == 0 {
		return math.Exp(-eta)
	}

	sum := 0.0
	for l := 1; l <= u; l++ {
		sum += math.Exp(-eta - float64(u)*math.Log(2) + float64(l)*math.Log(eta) - logGamma(float64(l)+1) +
			logGamma(float64(u)) - logGamma(float64(l)) - logGamma(float64(u-l)+1))
	}
	return sum
}
