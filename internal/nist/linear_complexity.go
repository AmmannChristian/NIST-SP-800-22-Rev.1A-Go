package nist

import (
	"math"

	"gonum.org/v1/gonum/mathext"
)

// LinearComplexityTest implements the NIST Linear Complexity test.
// It returns the p-value and whether it passes at Alpha.
func LinearComplexityTest(bitstream []byte, M int) (float64, bool) {
	bits := expandBits(bitstream)
	n := len(bits)

	N := n / M
	if N == 0 {
		return 0, false
	}

	K := 6
	pi := []float64{0.01047, 0.03125, 0.125, 0.5, 0.25, 0.0625, 0.020833}
	nu := make([]float64, K+1)

	for ii := 0; ii < N; ii++ {
		// Berlekamp-Massey algorithm: computes the linear complexity L of the
		// current M-bit block. C is the current connection polynomial, B is the
		// previous connection polynomial before the last length change, and T is
		// a temporary copy used during updates.
		C := make([]uint8, M)
		B := make([]uint8, M)
		T := make([]uint8, M)
		P := make([]uint8, M)
		L := 0
		m := -1
		d := uint8(0)
		C[0] = 1
		B[0] = 1

		N_ := 0
		for N_ < M {
			d = bits[ii*M+N_]
			for i := 1; i <= L; i++ {
				d ^= C[i] & bits[ii*M+N_-i]
			}
			if d == 1 {
				copy(T, C)
				for j := 0; j < M; j++ {
					if B[j] == 1 {
						P[j+N_-m] = 1
					}
				}
				for i := 0; i < M; i++ {
					C[i] ^= P[i]
					P[i] = 0
				}
				if L <= N_/2 {
					L = N_ + 1 - L
					m = N_
					copy(B, T)
				}
			}
			N_++
		}

		sign := 1.0
		if (M+1)%2 == 0 {
			sign = -1.0
		}
		mean := float64(M)/2.0 + (9.0+sign)/36.0 - (float64(M)/3.0+2.0/9.0)/math.Pow(2, float64(M))
		if M%2 == 0 {
			sign = 1.0
		} else {
			sign = -1.0
		}
		Tval := sign*(float64(L)-mean) + 2.0/9.0

		switch {
		case Tval <= -2.5:
			nu[0]++
		case Tval <= -1.5:
			nu[1]++
		case Tval <= -0.5:
			nu[2]++
		case Tval <= 0.5:
			nu[3]++
		case Tval <= 1.5:
			nu[4]++
		case Tval <= 2.5:
			nu[5]++
		default:
			nu[6]++
		}
	}

	chi2 := 0.0
	for i := 0; i < K+1; i++ {
		expected := float64(N) * pi[i]
		diff := nu[i] - expected
		chi2 += diff * diff / expected
	}

	pValue := mathext.GammaIncRegComp(float64(K)/2.0, chi2/2.0)
	return pValue, pValue >= Alpha
}
