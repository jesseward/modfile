package player

import (
	"encoding/binary"
	"io"

	"github.com/ebitengine/oto/v3"
)

// AudioPlayer defines the interface for writing audio data to a destination.
type AudioPlayer interface {
	io.WriteCloser
	GetSampleRate() int
}

// --- OtoPlayer ---

type OtoPlayer struct {
	player     *oto.Player
	ctx        *oto.Context
	pr         *io.PipeReader
	pw         *io.PipeWriter
	sampleRate int
}

func NewOtoPlayer(opts PlayerOptions) (*OtoPlayer, error) {
	c, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   opts.SampleRate,
		ChannelCount: opts.NumChannels,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, err
	}
	<-ready

	pr, pw := io.Pipe()

	player := c.NewPlayer(pr)

	return &OtoPlayer{
		player:     player,
		ctx:        c,
		pr:         pr,
		pw:         pw,
		sampleRate: opts.SampleRate,
	}, nil
}

func (o *OtoPlayer) Write(data []byte) (int, error) {
	return o.pw.Write(data)
}

func (o *OtoPlayer) Close() error {
	o.player.Close()
	o.pw.Close()
	return o.pr.Close()
}

func (o *OtoPlayer) GetSampleRate() int {
	return o.sampleRate
}

// --- StreamPlayer ---

type StreamPlayer struct {
	writer io.WriteCloser
	opts   PlayerOptions
}

func NewStreamPlayer(writer io.WriteCloser, opts PlayerOptions) *StreamPlayer {
	return &StreamPlayer{writer: writer, opts: opts}
}

func (s *StreamPlayer) Write(data []byte) (int, error) {
	return s.writer.Write(data)
}

func (s *StreamPlayer) Close() error {
	return s.writer.Close()
}

func (s *StreamPlayer) GetSampleRate() int {
	return s.opts.SampleRate
}

// --- WavPlayer ---

type WavPlayer struct {
	writer    io.WriteCloser
	opts      PlayerOptions
	dataSize  uint32
	chunkSize uint32
}

func NewWavPlayer(writer io.WriteCloser, opts PlayerOptions) *WavPlayer {
	return &WavPlayer{writer: writer, opts: opts}
}

func (w *WavPlayer) Write(data []byte) (int, error) {
	if w.dataSize == 0 {
		w.writeWavHeader()
	}
	n, err := w.writer.Write(data)
	w.dataSize += uint32(n)
	return n, err
}

func (w *WavPlayer) Close() error {
	w.chunkSize = 36 + w.dataSize
	// It's a bit of a hack to seek back to the beginning of the file to write the final chunk and data sizes
	if seeker, ok := w.writer.(io.Seeker); ok {
		seeker.Seek(4, io.SeekStart)
		binary.Write(w.writer, binary.LittleEndian, w.chunkSize)
		seeker.Seek(40, io.SeekStart)
		binary.Write(w.writer, binary.LittleEndian, w.dataSize)
	}
	return w.writer.Close()
}

func (w *WavPlayer) GetSampleRate() int {
	return w.opts.SampleRate
}

func (w *WavPlayer) writeWavHeader() {
	// RIFF header
	w.writer.Write([]byte("RIFF"))
	binary.Write(w.writer, binary.LittleEndian, uint32(0)) // Placeholder for chunk size
	w.writer.Write([]byte("WAVE"))

	// "fmt " sub-chunk
	w.writer.Write([]byte("fmt "))
	binary.Write(w.writer, binary.LittleEndian, uint32(16))                                           // Sub-chunk size
	binary.Write(w.writer, binary.LittleEndian, uint16(1))                                            // Audio format (PCM)
	binary.Write(w.writer, binary.LittleEndian, uint16(w.opts.NumChannels))                           // Num channels
	binary.Write(w.writer, binary.LittleEndian, uint32(w.opts.SampleRate))                            // Sample rate
	binary.Write(w.writer, binary.LittleEndian, uint32(w.opts.SampleRate*w.opts.NumChannels*w.opts.BitDepth)) // Byte rate
	binary.Write(w.writer, binary.LittleEndian, uint16(w.opts.NumChannels*w.opts.BitDepth))            // Block align
	binary.Write(w.writer, binary.LittleEndian, uint16(w.opts.BitDepth*8))                             // Bits per sample

	// "data" sub-chunk
	w.writer.Write([]byte("data"))
	binary.Write(w.writer, binary.LittleEndian, uint32(0)) // Placeholder for data size
}