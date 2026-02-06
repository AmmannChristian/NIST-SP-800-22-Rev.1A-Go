package nist

import "math"

// CumulativeSumsTest implements the NIST Cumulative Sums (Cusum) test.
// It returns the minimum p-value across forward and reverse runs and whether it passes at Alpha.
func CumulativeSumsTest(bitstream []byte) (float64, bool) {
	bits := expandBits(bitstream)
	n := len(bits)
	if n == 0 {
		return 0, false
	}

	pForward := cumulativeSums(bits, false)
	pReverse := cumulativeSums(bits, true)
	pValue := math.Min(pForward, pReverse)

	return pValue, pValue >= Alpha
}

// cumulativeSums computes the p-value for a single pass (forward or reverse) of
// the Cumulative Sums test. It constructs a partial-sum random walk from the
// mapped bit sequence, determines the maximum absolute excursion, and evaluates
// the corresponding p-value via the normal distribution.
func cumulativeSums(bits []uint8, reverse bool) float64 {
	n := len(bits)
	var sup, inf, sum float64

	if reverse {
		for i := n - 1; i >= 0; i-- {
			if bits[i] == 1 {
				sum++
			} else {
				sum--
			}
			if sum > sup {
				sup = sum
			}
			if sum < inf {
				inf = sum
			}
		}
	} else {
		for i := 0; i < n; i++ {
			if bits[i] == 1 {
				sum++
			} else {
				sum--
			}
			if sum > sup {
				sup = sum
			}
			if sum < inf {
				inf = sum
			}
		}
	}

	z := sup
	if math.Abs(inf) > z {
		z = math.Abs(inf)
	}

	start := int((-float64(n)/z + 1) / 4)
	finish := int((float64(n)/z - 1) / 4)

	var sum1 float64
	for k := start; k <= finish; k++ {
		term1 := normal((4*float64(k)+1)*z/math.Sqrt(float64(n))) - normal((4*float64(k)-1)*z/math.Sqrt(float64(n)))
		sum1 += term1
	}
	sum1 = 1.0 - sum1
	for k := start; k <= finish; k++ {
		term2 := normal((4*float64(k)+3)*z/math.Sqrt(float64(n))) - normal((4*float64(k)+1)*z/math.Sqrt(float64(n)))
		sum1 += term2
	}

	return sum1
}
