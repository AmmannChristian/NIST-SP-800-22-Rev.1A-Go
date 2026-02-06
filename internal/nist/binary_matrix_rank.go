package nist

import "math"

// BinaryMatrixRankTest implements the NIST Binary Matrix Rank test (32x32).
// It returns the p-value and whether it passes at Alpha.
func BinaryMatrixRankTest(bitstream []byte) (float64, bool) {
	bits := expandBits(bitstream)
	n := len(bits)

	const (
		m = 32
		q = 32
	)

	N := n / (m * q)
	if N == 0 {
		return 0, false
	}

	p32 := binaryRankProbability(32)
	p31 := binaryRankProbability(31)
	p30 := 1 - (p32 + p31)

	var f32, f31 float64
	for k := 0; k < N; k++ {
		matrix := make([][]uint8, m)
		for i := range matrix {
			matrix[i] = make([]uint8, q)
			for j := 0; j < q; j++ {
				matrix[i][j] = bits[k*(m*q)+i*q+j]
			}
		}

		rank := computeRank(matrix)
		if rank == 32 {
			f32++
		} else if rank == 31 {
			f31++
		}
	}
	f30 := float64(N) - (f32 + f31)

	chiSquared := math.Pow(f32-float64(N)*p32, 2)/(float64(N)*p32) +
		math.Pow(f31-float64(N)*p31, 2)/(float64(N)*p31) +
		math.Pow(f30-float64(N)*p30, 2)/(float64(N)*p30)

	pValue := math.Exp(-chiSquared / 2.0)

	return pValue, pValue >= Alpha
}

// binaryRankProbability computes the theoretical probability that a random
// 32x32 binary matrix has rank exactly r, using the product formula over
// GF(2) rank distributions.
func binaryRankProbability(r int) float64 {
	product := 1.0
	for i := 0; i <= r-1; i++ {
		num := (1.0 - math.Pow(2, float64(i-32))) * (1.0 - math.Pow(2, float64(i-32)))
		den := 1.0 - math.Pow(2, float64(i-r))
		product *= num / den
	}
	return math.Pow(2, float64(r*(32+32-r)-32*32)) * product
}

// computeRank determines the rank of a binary matrix over GF(2) using
// forward and backward Gaussian elimination with row swapping.
func computeRank(matrix [][]uint8) int {
	M := len(matrix)
	Q := len(matrix[0])
	m := M
	if Q < m {
		m = Q
	}

	for i := 0; i < m-1; i++ {
		if matrix[i][i] == 1 {
			performRowOps(matrix, i, true)
		} else if findUnitAndSwap(matrix, i, true) {
			performRowOps(matrix, i, true)
		}
	}

	for i := m - 1; i > 0; i-- {
		if matrix[i][i] == 1 {
			performRowOps(matrix, i, false)
		} else if findUnitAndSwap(matrix, i, false) {
			performRowOps(matrix, i, false)
		}
	}

	rank := m
	for i := 0; i < M; i++ {
		allZero := true
		for j := 0; j < Q; j++ {
			if matrix[i][j] == 1 {
				allZero = false
				break
			}
		}
		if allZero {
			rank--
		}
	}

	return rank
}

// performRowOps applies row reduction over GF(2) using pivot row i. When
// forward is true, it eliminates entries below the pivot; otherwise it
// eliminates entries above.
func performRowOps(matrix [][]uint8, i int, forward bool) {
	M := len(matrix)
	Q := len(matrix[0])
	if forward {
		for j := i + 1; j < M; j++ {
			if matrix[j][i] == 1 {
				for k := i; k < Q; k++ {
					matrix[j][k] ^= matrix[i][k]
				}
			}
		}
	} else {
		for j := i - 1; j >= 0; j-- {
			if matrix[j][i] == 1 {
				for k := 0; k < Q; k++ {
					matrix[j][k] ^= matrix[i][k]
				}
			}
		}
	}
}

// findUnitAndSwap searches for a row with a 1 in column i and swaps it with
// row i to establish a pivot. It searches downward when forward is true and
// upward otherwise. It returns true if a suitable row was found and swapped.
func findUnitAndSwap(matrix [][]uint8, i int, forward bool) bool {
	M := len(matrix)

	if forward {
		for idx := i + 1; idx < M; idx++ {
			if matrix[idx][i] == 1 {
				matrix[i], matrix[idx] = matrix[idx], matrix[i]
				return true
			}
		}
	} else {
		for idx := i - 1; idx >= 0; idx-- {
			if matrix[idx][i] == 1 {
				matrix[i], matrix[idx] = matrix[idx], matrix[i]
				return true
			}
		}
	}

	return false
}
