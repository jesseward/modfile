package xm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/jesseward/impulse/pkg/module"
)

// Module represents an entire XM module.
type Module struct {
	Header      Header
	Patterns    []*Pattern
	Instruments []*Instrument
}

// Header represents the XM file header.
type Header struct {
	IDText          string
	ModuleName      string
	TrackerName     string
	Version         uint16
	HeaderSize      uint32
	SongLength      uint16
	RestartPosition uint16
	NumChannels     uint16
	NumPatterns     uint16
	NumInstruments  uint16
	Flags           uint16
	DefaultTempo    uint16
	DefaultBPM      uint16
	patternOrder    []byte
}

// Pattern represents a single pattern.
type Pattern struct {
	HeaderLength   uint32
	PackingType    byte
	NumRows        uint16
	PackedDataSize uint16
	Notes          [][]Note
}

// Note represents a single note in a pattern.
type Note struct {
	Note        byte
	Instrument  byte
	Volume      byte
	EffectType  byte
	EffectParam byte
}

// EnvelopePoint represents a single point in an envelope.
type EnvelopePoint struct {
	Frame uint16
	Value uint16
}

// Instrument represents an XM instrument.
type Instrument struct {
	HeaderSize            uint32
	Name                  string
	Type                  byte
	NumSamples            uint16
	SampleHeaderSize      uint32
	SampleKeymap          [96]byte
	VolumeEnvelopePoints  [12]EnvelopePoint
	PanningEnvelopePoints [12]EnvelopePoint
	NumVolumePoints       byte
	NumPanningPoints      byte
	VolumeSustainPoint    byte
	VolumeLoopStartPoint  byte
	VolumeLoopEndPoint    byte
	PanningSustainPoint   byte
	PanningLoopStartPoint byte
	PanningLoopEndPoint   byte
	VolumeType            byte
	PanningType           byte
	VibratoType           byte
	VibratoSweep          byte
	VibratoDepth          byte
	VibratoRate           byte
	VolumeFadeout         uint16
	Reserved              uint16
	Samples               []*Sample
}

// Sample represents a single sample.
type Sample struct {
	length       uint32
	loopStart    uint32
	loopLength   uint32
	volume       byte
	finetune     int8
	Type         byte
	panning      byte
	relativeNote int8
	Reserved     byte
	name         string
	data         []int16 // Using int16 to accommodate both 8-bit and 16-bit samples
	flags        byte
}

// Read reads an XM module from the given reader.
func Read(r io.Reader) (*Module, error) {
	mod := &Module{}

	// Parse Header
	if err := mod.Header.parse(r); err != nil {
		return nil, fmt.Errorf("failed to parse XM header: %w", err)
	}

	// Parse Patterns
	mod.Patterns = make([]*Pattern, mod.Header.NumPatterns)
	for i := range mod.Patterns {
		p := &Pattern{}
		if err := p.parse(r, int(mod.Header.NumChannels)); err != nil {
			return nil, fmt.Errorf("failed to parse pattern %d: %w", i, err)
		}
		mod.Patterns[i] = p
	}

	// Parse Instruments
	mod.Instruments = make([]*Instrument, mod.Header.NumInstruments)
	for i := range mod.Instruments {
		inst := &Instrument{}
		if err := inst.parse(r); err != nil {
			return nil, fmt.Errorf("failed to parse instrument %d: %w", i, err)
		}
		mod.Instruments[i] = inst
	}

	return mod, nil
}

func (h *Header) parse(r io.Reader) error {
	var id [17]byte
	if _, err := io.ReadFull(r, id[:]); err != nil {
		return fmt.Errorf("reading ID text: %w", err)
	}
	h.IDText = string(id[:])

	var modName [20]byte
	if _, err := io.ReadFull(r, modName[:]); err != nil {
		return fmt.Errorf("reading module name: %w", err)
	}
	h.ModuleName = strings.TrimRight(string(modName[:]), "\x00")

	var magic [1]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return fmt.Errorf("reading magic byte: %w", err)
	}

	var trackerName [20]byte
	if _, err := io.ReadFull(r, trackerName[:]); err != nil {
		return fmt.Errorf("reading tracker name: %w", err)
	}
	h.TrackerName = strings.TrimRight(string(trackerName[:]), "\x00")

	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
		return fmt.Errorf("reading version: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &h.HeaderSize); err != nil {
		return fmt.Errorf("reading header size: %w", err)
	}

	headerRestBytes := make([]byte, h.HeaderSize-4)
	if _, err := io.ReadFull(r, headerRestBytes); err != nil {
		return fmt.Errorf("reading rest of header: %w", err)
	}
	restReader := bytes.NewReader(headerRestBytes)

	binary.Read(restReader, binary.LittleEndian, &h.SongLength)
	binary.Read(restReader, binary.LittleEndian, &h.RestartPosition)
	binary.Read(restReader, binary.LittleEndian, &h.NumChannels)
	binary.Read(restReader, binary.LittleEndian, &h.NumPatterns)
	binary.Read(restReader, binary.LittleEndian, &h.NumInstruments)
	binary.Read(restReader, binary.LittleEndian, &h.Flags)
	binary.Read(restReader, binary.LittleEndian, &h.DefaultTempo)
	binary.Read(restReader, binary.LittleEndian, &h.DefaultBPM)

	h.patternOrder = make([]byte, 256)
	if _, err := io.ReadFull(restReader, h.patternOrder); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// some trackers don't write the full 256 bytes if the song is shorter
			// This is fine.
		} else {
			return fmt.Errorf("reading pattern order: %w", err)
		}
	}

	return nil
}

func (p *Pattern) parse(r io.Reader, numChannels int) error {
	var header struct {
		HeaderLength   uint32
		PackingType    byte
		NumRows        uint16
		PackedDataSize uint16
	}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("reading pattern header: %w", err)
	}
	p.HeaderLength = header.HeaderLength
	p.PackingType = header.PackingType
	p.NumRows = header.NumRows
	p.PackedDataSize = header.PackedDataSize

	if p.HeaderLength > 9 {
		if _, err := io.CopyN(io.Discard, r, int64(p.HeaderLength-9)); err != nil {
			return fmt.Errorf("skipping rest of pattern header: %w", err)
		}
	}

	if p.PackedDataSize == 0 {
		p.Notes = make([][]Note, p.NumRows)
		for i := range p.Notes {
			p.Notes[i] = make([]Note, numChannels)
		}
		return nil
	}

	packedData := make([]byte, p.PackedDataSize)
	if _, err := io.ReadFull(r, packedData); err != nil {
		return fmt.Errorf("reading packed pattern data: %w", err)
	}

	p.Notes = make([][]Note, p.NumRows)
	for i := range p.Notes {
		p.Notes[i] = make([]Note, numChannels)
	}

	row, ch := 0, 0
	i := 0
	for i < len(packedData) {
		b := packedData[i]
		i++
		var note Note
		if b&0x80 != 0 { // Packed
			if b&0x01 != 0 {
				note.Note = packedData[i]
				i++
			}
			if b&0x02 != 0 {
				note.Instrument = packedData[i]
				i++
			}
			if b&0x04 != 0 {
				note.Volume = packedData[i]
				i++
			}
			if b&0x08 != 0 {
				note.EffectType = packedData[i]
				i++
			}
			if b&0x10 != 0 {
				note.EffectParam = packedData[i]
				i++
			}
		} else { // Unpacked
			note.Note = b
			note.Instrument = packedData[i]
			note.Volume = packedData[i+1]
			note.EffectType = packedData[i+2]
			note.EffectParam = packedData[i+3]
			i += 4
		}
		if row < int(p.NumRows) && ch < numChannels {
			p.Notes[row][ch] = note
		}
		ch++
		if ch >= numChannels {
			ch = 0
			row++
		}
		if row >= int(p.NumRows) {
			break
		}
	}

	return nil
}

func (i *Instrument) parse(r io.Reader) error {
	var headerSize uint32
	if err := binary.Read(r, binary.LittleEndian, &headerSize); err != nil {
		return fmt.Errorf("reading instrument header size: %w", err)
	}
	i.HeaderSize = headerSize

	if headerSize == 0 {
		return nil
	}

	headerBytes := make([]byte, headerSize-4)
	if _, err := io.ReadFull(r, headerBytes); err != nil {
		return fmt.Errorf("reading instrument header data: %w", err)
	}
	hr := bytes.NewReader(headerBytes)

	var nameBytes [22]byte
	binary.Read(hr, binary.LittleEndian, &nameBytes)
	i.Name = strings.TrimRight(string(nameBytes[:]), "\x00")

	binary.Read(hr, binary.LittleEndian, &i.Type)
	binary.Read(hr, binary.LittleEndian, &i.NumSamples)

	if i.NumSamples > 0 {
		binary.Read(hr, binary.LittleEndian, &i.SampleHeaderSize)
		binary.Read(hr, binary.LittleEndian, &i.SampleKeymap)

		var volPoints, panPoints [48]byte
		binary.Read(hr, binary.LittleEndian, &volPoints)
		binary.Read(hr, binary.LittleEndian, &panPoints)

		volReader := bytes.NewReader(volPoints[:])
		for j := 0; j < 12; j++ {
			binary.Read(volReader, binary.LittleEndian, &i.VolumeEnvelopePoints[j])
		}
		panReader := bytes.NewReader(panPoints[:])
		for j := 0; j < 12; j++ {
			binary.Read(panReader, binary.LittleEndian, &i.PanningEnvelopePoints[j])
		}

		binary.Read(hr, binary.LittleEndian, &i.NumVolumePoints)
		binary.Read(hr, binary.LittleEndian, &i.NumPanningPoints)
		binary.Read(hr, binary.LittleEndian, &i.VolumeSustainPoint)
		binary.Read(hr, binary.LittleEndian, &i.VolumeLoopStartPoint)
		binary.Read(hr, binary.LittleEndian, &i.VolumeLoopEndPoint)
		binary.Read(hr, binary.LittleEndian, &i.PanningSustainPoint)
		binary.Read(hr, binary.LittleEndian, &i.PanningLoopStartPoint)
		binary.Read(hr, binary.LittleEndian, &i.PanningLoopEndPoint)
		binary.Read(hr, binary.LittleEndian, &i.VolumeType)
		binary.Read(hr, binary.LittleEndian, &i.PanningType)
		binary.Read(hr, binary.LittleEndian, &i.VibratoType)
		binary.Read(hr, binary.LittleEndian, &i.VibratoSweep)
		binary.Read(hr, binary.LittleEndian, &i.VibratoDepth)
		binary.Read(hr, binary.LittleEndian, &i.VibratoRate)
		binary.Read(hr, binary.LittleEndian, &i.VolumeFadeout)
		binary.Read(hr, binary.LittleEndian, &i.Reserved)
	}

	i.Samples = make([]*Sample, i.NumSamples)
	for j := 0; j < int(i.NumSamples); j++ {
		s := &Sample{}
		if err := s.parseHeader(r); err != nil {
			return fmt.Errorf("parsing sample header for instrument %s, sample %d: %w", i.Name, j, err)
		}
		i.Samples[j] = s
	}

	for j := 0; j < int(i.NumSamples); j++ {
		if err := i.Samples[j].parseData(r); err != nil {
			return fmt.Errorf("parsing sample data for instrument %s, sample %d: %w", i.Name, j, err)
		}
	}

	return nil
}

func (s *Sample) parseHeader(r io.Reader) error {
	var header struct {
		Length       uint32
		LoopStart    uint32
		LoopLength   uint32
		Volume       byte
		Finetune     int8
		Type         byte
		Panning      byte
		RelativeNote int8
		Reserved     byte
		Name         [22]byte
	}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return err
	}
	s.length = header.Length
	s.loopStart = header.LoopStart
	s.loopLength = header.LoopLength
	s.volume = header.Volume
	s.finetune = header.Finetune
	s.Type = header.Type
	s.panning = header.Panning
	s.relativeNote = header.RelativeNote
	s.Reserved = header.Reserved
	s.name = strings.TrimRight(string(header.Name[:]), "\x00")
	s.flags = s.Type
	return nil
}

func (s *Sample) parseData(r io.Reader) error {
	if s.length == 0 {
		return nil
	}
	is16bit := s.Type&0x10 != 0
	sampleLen := s.length
	if is16bit {
		sampleLen /= 2
	}

	rawData := make([]byte, s.length)
	if _, err := io.ReadFull(r, rawData); err != nil {
		return fmt.Errorf("reading sample data: %w", err)
	}

	s.data = make([]int16, sampleLen)

	if is16bit {
		var old int16
		for i := uint32(0); i < sampleLen; i++ {
			val := int16(binary.LittleEndian.Uint16(rawData[i*2:]))
			new := val + old
			s.data[i] = new
			old = new
		}
	} else {
		var old int8
		for i := uint32(0); i < sampleLen; i++ {
			val := int8(rawData[i])
			new := val + old
			s.data[i] = int16(new) << 8
			old = new
		}
	}
	return nil
}

func (m *Module) Name() string {
	return m.Header.ModuleName
}

func (m *Module) Type() string {
	return "FastTracker II Extended Module"
}

func (m *Module) SongLength() int {
	return int(m.Header.SongLength)
}

func (m *Module) NumChannels() int {
	return int(m.Header.NumChannels)
}

func (m *Module) NumPatterns() int {
	return int(m.Header.NumPatterns)
}

func (m *Module) NumRows(pattern int) int {
	if pattern >= len(m.Patterns) {
		return 0
	}
	return int(m.Patterns[pattern].NumRows)
}

func (m *Module) Samples() []module.Sample {
	var samples []module.Sample
	for _, inst := range m.Instruments {
		for _, s := range inst.Samples {
			samples = append(samples, s)
		}
	}
	return samples
}

func (m *Module) PatternOrder() []int {
	order := make([]int, m.Header.SongLength)
	for i := 0; i < int(m.Header.SongLength); i++ {
		order[i] = int(m.Header.patternOrder[i])
	}
	return order
}

func (m *Module) DefaultSpeed() int {
	return int(m.Header.DefaultTempo)
}

func (m *Module) DefaultBPM() int {
	return int(m.Header.DefaultBPM)
}

func (m *Module) PatternCell(pattern, row, channel int) module.Cell {
	if pattern >= len(m.Patterns) || row >= int(m.Patterns[pattern].NumRows) || channel >= int(m.Header.NumChannels) {
		return module.Cell{}
	}
	note := m.Patterns[pattern].Notes[row][channel]
	return module.Cell{
		Note:        note.Note,
		Instrument:  note.Instrument,
		Volume:      note.Volume,
		Effect:      note.EffectType,
		EffectParam: note.EffectParam,
	}
}

func (s *Sample) Name() string {
	return s.name
}

func (s *Sample) Length() uint32 {
	return s.length
}

func (s *Sample) Volume() uint8 {
	return s.volume
}

func (s *Sample) LoopStart() uint32 {
	return s.loopStart
}

func (s *Sample) LoopLength() uint32 {
	return s.loopLength
}

func (s *Sample) Data() []int16 {
	return s.data
}

func (s *Sample) Finetune() uint32 {
	return uint32(s.finetune)
}

func (s *Sample) Flags() byte {
	return s.flags
}

func (s *Sample) IsPingPong() bool {
	return s.flags&0x02 != 0
}

func (s *Sample) RelativeNote() int8 {
	return s.relativeNote
}

func (s *Sample) Panning() byte {
	return s.panning
}


func (s *Sample) LoopEnd() uint32 {
	return s.loopStart + s.loopLength
}

func (s *Sample) AsciiWaveform(width, height int) string {
	if len(s.data) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	bucketSize := float64(len(s.data)) / float64(width)
	halfHeight := float64(height) / 2.0

	for i := 0; i < width; i++ {
		start := int(float64(i) * bucketSize)
		end := int(float64(i+1) * bucketSize)
		if end > len(s.data) {
			end = len(s.data)
		}
		if start >= end {
			continue
		}

		bucket := s.data[start:end]
		var minVal, maxVal int16 = 0, 0
		for _, s := range bucket {
			if s < minVal {
				minVal = s
			}
			if s > maxVal {
				maxVal = s
			}
		}

		// Normalize and scale to the view height
		yMax := int(math.Round(float64(maxVal)/32767.0*halfHeight + halfHeight))
		yMin := int(math.Round(float64(minVal)/32767.0*halfHeight + halfHeight))

		// Clamp values to be within the grid
		if yMax >= height {
			yMax = height - 1
		}
		if yMin < 0 {
			yMin = 0
		}

		// Draw the vertical bar for the current bucket
		for y := yMin; y <= yMax; y++ {
			if y >= 0 && y < height {
				grid[y][i] = '█'
			}
		}
	}

	var builder strings.Builder
	for y := 0; y < height; y++ {
		builder.WriteString(string(grid[y]))
		builder.WriteRune('\n')
	}
	return builder.String()
}
