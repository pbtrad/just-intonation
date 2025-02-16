package audio

import (
	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate = 44100
	bufferSize = 4096
)

type Capturer struct {
	stream *portaudio.Stream
	buffer []float32
}

func NewCapturer() *Capturer {
	return &Capturer{
		buffer: make([]float32, bufferSize),
	}
}

func (c *Capturer) StartCapture() error {
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(c.buffer), c.buffer)
	if err != nil {
		return err
	}
	c.stream = stream
	return stream.Start()
}

func (c *Capturer) GetBuffer() []float32 {
	return c.buffer
}

func (c *Capturer) Close() error {
	if c.stream != nil {
		return c.stream.Close()
	}
	return nil
}

func (c *Capturer) Stop() error {
	if c.stream != nil {
		return c.stream.Stop()
	}
	return nil
}
