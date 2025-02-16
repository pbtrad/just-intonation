package audio

import (
	"math"
	"math/cmplx"

	"github.com/mjibson/go-dsp/fft"
)

const (
	noiseThreshold = 0.1
	minFreq        = 20.0
	maxFreq        = 4000.0
)

type Analyzer struct {
	capturer  *Capturer
	frequency float64
}

func NewAnalyzer(capturer *Capturer) *Analyzer {
	return &Analyzer{
		capturer: capturer,
	}
}

func (a *Analyzer) GetCurrentFrequency() float64 {
	buffer := a.capturer.GetBuffer()

	// Convert buffer to complex numbers for FFT
	complexData := make([]complex128, len(buffer))
	for i, sample := range buffer {
		complexData[i] = complex(float64(sample), 0)
	}

	// Hanning window
	for i := range complexData {
		multiplier := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(complexData))))
		complexData[i] *= complex(multiplier, 0)
	}

	spectrum := fft.FFT(complexData)

	// Find peak frequency with quadratic interpolation
	maxMagnitude := 0.0
	maxIndex := 0
	for i := 1; i < len(spectrum)/2-1; i++ {
		magnitude := cmplx.Abs(spectrum[i])
		if magnitude > maxMagnitude && magnitude > noiseThreshold {
			prevMag := cmplx.Abs(spectrum[i-1])
			nextMag := cmplx.Abs(spectrum[i+1])

			if magnitude > prevMag && magnitude > nextMag {
				maxMagnitude = magnitude
				maxIndex = i
			}
		}
	}

	// If no clear peak found
	if maxMagnitude < noiseThreshold {
		return 0.0
	}

	// Quadratic interpolation for better frequency precision
	alpha := cmplx.Abs(spectrum[maxIndex-1])
	beta := cmplx.Abs(spectrum[maxIndex])
	gamma := cmplx.Abs(spectrum[maxIndex+1])

	p := 0.5 * (alpha - gamma) / (alpha - 2*beta + gamma)
	interpolatedIndex := float64(maxIndex) + p

	// Convert index to freq
	frequency := interpolatedIndex * float64(sampleRate) / float64(len(buffer))

	// Validate freq
	if frequency < minFreq || frequency > maxFreq {
		return 0.0
	}

	// Moving average to smooth results
	a.frequency = (a.frequency*0.7 + frequency*0.3) // Smoothing factor

	return a.frequency
}

func (a *Analyzer) getFrequencyFromZeroCrossings() float64 {
	buffer := a.capturer.GetBuffer()
	crossings := 0
	for i := 1; i < len(buffer); i++ {
		if (buffer[i-1] < 0 && buffer[i] >= 0) ||
			(buffer[i-1] >= 0 && buffer[i] < 0) {
			crossings++
		}
	}

	return float64(crossings) * float64(sampleRate) / (2 * float64(len(buffer)))
}
