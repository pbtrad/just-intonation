// https://pkg.go.dev/github.com/mjibson/go-dsp/fft for fft reference

package audio

import (
	"math"
	"math/cmplx"

	"github.com/gordonklaus/portaudio"
	"github.com/mjibson/go-dsp/fft"
)

const (
	sampleRate     = 44100
	bufferSize     = 4096
	noiseThreshold = 0.1  // Not sure might adjust this, need to test
	minFreq        = 20.0 // hz
	maxFreq        = 4000.0
)

type Analyzer struct {
	stream    *portaudio.Stream
	buffer    []float32
	frequency float64
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		buffer: make([]float32, bufferSize),
	}
}

func (a *Analyzer) StartCapture() error {
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(a.buffer), a.buffer)
	if err != nil {
		return err
	}
	a.stream = stream
	return stream.Start()
}

func (a *Analyzer) GetCurrentFrequency() float64 {
	// Convert buffer to complex numbers for FFT
	complexData := make([]complex128, len(a.buffer))
	for i, sample := range a.buffer {
		complexData[i] = complex(float64(sample), 0)
	}

	// Apply Hanning window
	for i := range complexData {
		multiplier := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(complexData))))
		complexData[i] *= complex(multiplier, 0)
	}

	// FFT
	spectrum := fft.FFT(complexData)

	// Find peak frequency with quadratic interpolation
	maxMagnitude := 0.0
	maxIndex := 0
	for i := 1; i < len(spectrum)/2-1; i++ {
		magnitude := cmplx.Abs(spectrum[i])
		if magnitude > maxMagnitude && magnitude > noiseThreshold {
			prevMag := cmplx.Abs(spectrum[i-1])
			nextMag := cmplx.Abs(spectrum[i+1])

			// Only update if it's a clear peak
			if magnitude > prevMag && magnitude > nextMag {
				maxMagnitude = magnitude
				maxIndex = i
			}
		}
	}

	if maxMagnitude < noiseThreshold {
		return 0.0
	}

	// Quadratic interpolation for better frequency precision
	alpha := cmplx.Abs(spectrum[maxIndex-1])
	beta := cmplx.Abs(spectrum[maxIndex])
	gamma := cmplx.Abs(spectrum[maxIndex+1])

	p := 0.5 * (alpha - gamma) / (alpha - 2*beta + gamma)
	interpolatedIndex := float64(maxIndex) + p

	// Index to freq
	frequency := interpolatedIndex * float64(sampleRate) / float64(len(a.buffer))

	// Validate freq range
	if frequency < minFreq || frequency > maxFreq {
		return 0.0
	}

	// Moving avergae
	a.frequency = (a.frequency*0.7 + frequency*0.3) // Smoothing factor

	return a.frequency
}

// Gets frequency from buffer using zero-crossing rate
func (a *Analyzer) getFrequencyFromZeroCrossings() float64 {
	crossings := 0
	for i := 1; i < len(a.buffer); i++ {
		if (a.buffer[i-1] < 0 && a.buffer[i] >= 0) ||
			(a.buffer[i-1] >= 0 && a.buffer[i] < 0) {
			crossings++
		}
	}

	// Zero-crossing rate to freq
	return float64(crossings) * float64(sampleRate) / (2 * float64(len(a.buffer)))
}

// Close audio stream
func (a *Analyzer) Close() error {
	if a.stream != nil {
		return a.stream.Close()
	}
	return nil
}

// Stop audio stream
func (a *Analyzer) Stop() error {
	if a.stream != nil {
		return a.stream.Stop()
	}
	return nil
}
