package nist

import (
	"math"
	"math/cmplx"

	"gonum.org/v1/gonum/dsp/fourier"
)

// DiscreteFourierTransformTest implements the NIST Spectral (FFT) test.
// It returns the p-value and whether it passes at Alpha.
func DiscreteFourierTransformTest(bitstream []byte) (float64, bool) {
	n := len(bitstream) * 8
	if n == 0 {
		return 0, false
	}

	series := make([]float64, n)
	for i := 0; i < n; i++ {
		if bitAt(bitstream, i) == 1 {
			series[i] = 1
		} else {
			series[i] = -1
		}
	}

	fft := fourier.NewFFT(n)
	coeffs := fft.Coefficients(nil, series)

	magnitudes := make([]float64, n/2)
	for i := 0; i < n/2; i++ {
		magnitudes[i] = cmplx.Abs(coeffs[i])
	}

	// 2.995732274 equals -ln(0.05), the 95% threshold for spectral peaks.
	upperBound := math.Sqrt(2.995732274 * float64(n))
	count := 0
	for _, m := range magnitudes {
		if m < upperBound {
			count++
		}
	}

	d := (float64(count) - 0.95*float64(n)/2.0) / math.Sqrt(float64(n)/4.0*0.95*0.05)
	pValue := math.Erfc(math.Abs(d) / math.Sqrt2)

	return pValue, pValue >= Alpha
}
