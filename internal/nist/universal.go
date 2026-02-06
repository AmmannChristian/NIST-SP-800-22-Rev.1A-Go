package nist

import "math"

// UniversalStatisticalTest implements Maurer's Universal Statistical test.
// It returns the p-value and whether it passes at Alpha.
func UniversalStatisticalTest(bitstream []byte) (float64, bool) {
	bits := expandBits(bitstream)
	n := len(bits)

	L := 5
	switch {
	case n >= 1059061760:
		L = 16
	case n >= 496435200:
		L = 15
	case n >= 231669760:
		L = 14
	case n >= 107560960:
		L = 13
	case n >= 49643520:
		L = 12
	case n >= 22753280:
		L = 11
	case n >= 10342400:
		L = 10
	case n >= 4654080:
		L = 9
	case n >= 2068480:
		L = 8
	case n >= 904960:
		L = 7
	case n >= 387840:
		L = 6
	default:
		L = 5
	}

	Q := 10 * (1 << L)
	K := n/L - Q
	if L < 6 || L > 16 || Q < 10*(1<<L) || K <= 0 {
		return 0, false
	}

	// Tabulated expected values of the test statistic, indexed by block length L (valid for L=6..16).
	expected := [...]float64{0, 0, 0, 0, 0, 0, 5.2177052, 6.1962507, 7.1836656, 8.1764248, 9.1723243, 10.170032, 11.168765, 12.16807, 13.167693, 14.167488, 15.167379}
	// Tabulated variance values of the test statistic, indexed by block length L (valid for L=6..16).
	variance := [...]float64{0, 0, 0, 0, 0, 0, 2.954, 3.125, 3.238, 3.311, 3.356, 3.384, 3.401, 3.41, 3.416, 3.419, 3.421}

	T := make([]int, 1<<L)
	for i := 0; i < len(T); i++ {
		T[i] = 0
	}

	for i := 1; i <= Q; i++ {
		decRep := 0
		for j := 0; j < L; j++ {
			decRep = decRep*2 + int(bits[(i-1)*L+j])
		}
		T[decRep] = i
	}

	sum := 0.0
	for i := Q + 1; i <= Q+K; i++ {
		decRep := 0
		for j := 0; j < L; j++ {
			decRep = decRep*2 + int(bits[(i-1)*L+j])
		}
		sum += math.Log(float64(i-T[decRep])) / math.Log(2)
		T[decRep] = i
	}

	phi := sum / float64(K)
	sigma := (0.7 - 0.8/float64(L) + (4+32/float64(L))*math.Pow(float64(K), -3/float64(L))/15) * math.Sqrt(variance[L]/float64(K))
	arg := math.Abs(phi-expected[L]) / (math.Sqrt2 * sigma)
	pValue := math.Erfc(arg)

	return pValue, pValue >= Alpha
}
