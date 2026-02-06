package nist

import "math"

// Alpha is the default significance level used by the NIST SP800-22 tests.
const Alpha = 0.01

// bitAt returns the bit (0 or 1) at position idx in big-endian bit order.
func bitAt(data []byte, idx int) uint8 {
	byteIdx := idx >> 3
	bitIdx := 7 - (idx & 7)
	return (data[byteIdx] >> bitIdx) & 1
}

// expandBits converts a byte slice to a slice of individual bits (0 or 1).
func expandBits(bitstream []byte) []uint8 {
	n := len(bitstream) * 8
	bits := make([]uint8, n)
	for i := 0; i < n; i++ {
		bits[i] = bitAt(bitstream, i)
	}
	return bits
}

// normal computes the normal (Gaussian) cumulative distribution function.
func normal(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}

// psi2 computes the psi-squared statistic for block length m, as used by the
// Serial and Approximate Entropy tests. It counts the frequency of each m-bit
// pattern in the circular bit sequence using a binary trie indexed by pattern
// value, then returns the normalized sum of squared counts.
func psi2(bits []uint8, m int) float64 {
	if m <= 0 {
		return 0
	}

	n := len(bits)
	powLen := (1 << (m + 1)) - 1
	P := make([]int, powLen)

	for i := 0; i < n; i++ {
		k := 1
		for j := 0; j < m; j++ {
			if bits[(i+j)%n] == 0 {
				k *= 2
			} else {
				k = 2*k + 1
			}
		}
		P[k-1]++
	}

	sum := 0.0
	start := (1 << m) - 1
	end := (1 << (m + 1)) - 1
	for i := start; i < end; i++ {
		sum += math.Pow(float64(P[i]), 2)
	}

	return sum*math.Pow(2, float64(m))/float64(n) - float64(n)
}
