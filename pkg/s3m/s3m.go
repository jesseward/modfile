package s3m

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/jesseward/impulse/pkg/module"
)

// Header represents the main header of an S3M file.
type Header struct {
	SongName          [28]byte
	Marker1A          byte
	FileType          byte
	Reserved1         [2]byte
	OrderCount        uint16
	InstrumentCount   uint16
	PatternCount      uint16
	Flags             uint16
	TrackerVersion    uint16
	SampleType        uint16
	Signature         [4]byte
	GlobalVolume      byte
	InitialSpeed      byte
	InitialTempo      byte
	MasterVolume      byte
	UltraClickRemoval byte
	DefaultPan        byte
	Reserved2         [8]byte
	Special           uint16
	ChannelSettings   [32]byte
}

// instrumentHeader represents the fixed-size portion of the S3M instrument data in the file.
type instrumentHeader struct {
	Type        byte
	DOSFilename [12]byte
	MemSeg      [3]byte
	Length      uint32
	LoopBegin   uint32
	LoopEnd     uint32
	Volume      byte
	Reserved    byte
	Pack        byte
	Flags       byte
	C2Spd       uint32
	Reserved2   [12]byte
	SampleName  [28]byte
	Signature   [4]byte
}

// Instrument represents an S3M instrument.
// See docs/s3mplayer.txt section 2.9
type Instrument struct {
	Type        byte
	DOSFilename [12]byte
	MemSeg      [3]byte // Pointer to sample data. High byte, then low word.
	length      uint32
	LoopBegin   uint32
	loopEnd     uint32
	volume      byte
	Reserved    byte
	Pack        byte
	flags       byte
	C2Spd       uint32
	Reserved2   [12]byte
	SampleName  [28]byte
	Signature   [4]byte // SCRS
	data        []int16
	Signed      bool
}

// GetName returns the name of the instrument.
func (inst *Instrument) Name() string {
	return strings.TrimRight(string(inst.SampleName[:]), "\x00")
}

// GetLength returns the length of the sample data.
func (inst *Instrument) Length() uint32 {
	return inst.length
}

// GetVolume returns the default volume of the sample.
func (inst *Instrument) Volume() uint8 {
	return inst.volume
}

// GetLoopStart returns the starting position of the sample loop.
func (inst *Instrument) LoopStart() uint32 {
	return inst.LoopBegin
}

// GetLoopLength returns the length of the sample loop.
func (inst *Instrument) LoopLength() uint32 {
	if inst.flags&1 == 0 { // Check loop flag
		return 0
	}
	return inst.loopEnd - inst.LoopBegin
}

func (inst *Instrument) Finetune() uint32 {
	return inst.C2Spd
}

func (inst *Instrument) Flags() byte {
	return inst.flags
}

func (inst *Instrument) IsPingPong() bool {
	return false
}

func (inst *Instrument) RelativeNote() int8 {
	return 0
}

func (inst *Instrument) Panning() byte {
	// S3M does not have per-sample panning.
	// Panning is set via effects or the default pan positions.
	// Returning a default center value.
	return 128
}

func (inst *Instrument) LoopEnd() uint32 {
	return inst.loopEnd
}

// GetData returns the raw sample data as 16-bit signed integers.
func (inst *Instrument) Data() []int16 {
	return inst.data
}

func (inst *Instrument) AsciiWaveform(width, height int) string {
	if len(inst.data) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	bucketSize := float64(len(inst.data)) / float64(width)
	halfHeight := float64(height) / 2.0

	for i := 0; i < width; i++ {
		start := int(float64(i) * bucketSize)
		end := int(float64(i+1) * bucketSize)
		if end > len(inst.data) {
			end = len(inst.data)
		}
		if start >= end {
			continue
		}

		bucket := inst.data[start:end]
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

var noteTable = [12]string{"C-", "C#", "D-", "D#", "E-", "F-", "F#", "G-", "G#", "A-", "A#", "B-"}

func NoteToString(note byte) string {
	if note == 255 || note == 0 {
		return module.EmptyNote
	}
	if note == 254 {
		return "---"
	}
	octave := note >> 4
	noteVal := note & 0x0F
	return fmt.Sprintf("%s%d", noteTable[noteVal], octave)
}

// PatternEntry represents a single entry in a pattern for a single channel.
type PatternEntry struct {
	Note       byte
	Instrument byte
	Volume     byte
	Effect     module.Effect
}

// Pattern represents a pattern with 64 rows.
type Pattern [][]PatternEntry

// S3M represents a parsed S3M module.
type S3M struct {
	Header                 Header
	Orders                 []byte
	InstrumentParapointers []uint16
	PatternParapointers    []uint16
	DefaultPanPositions    []byte
	Instruments            []Instrument
	Patterns               []Pattern
	ChannelRemap           [32]int
	numChannels            int
	ActualPatternCount     int
	SignedSamples          bool
}

// Name returns the name of the module.
func (s *S3M) Name() string {
	return strings.TrimRight(string(s.Header.SongName[:]), "\x00")
}

func (s *S3M) Type() string {
	return "S3M"
}

// GetSongLength returns the length of the song in patterns.
func (s *S3M) SongLength() int {
	// This should be the count of orders, excluding markers.
	// For now, returning the raw order count.
	// A more accurate implementation would filter out markers 254 and 255.
	count := 0
	for _, order := range s.Orders {
		if order < 254 {
			count++
		}
	}
	return count
}

// NumChannels returns the number of channels in the module.
func (s *S3M) NumChannels() int {
	return s.numChannels
}

// NumPatterns returns the number of patterns in the module.
func (s *S3M) NumPatterns() int {
	return s.ActualPatternCount
}

func (s *S3M) NumRows(pattern int) int {
	return 64
}

// GetSamples returns a slice of the module's samples.
func (s *S3M) Samples() []module.Sample {
	samples := make([]module.Sample, len(s.Instruments))
	for i := range s.Instruments {
		samples[i] = &s.Instruments[i]
	}
	return samples
}

// PatternOrder returns the order of patterns to be played.
func (s *S3M) PatternOrder() []int {
	order := make([]int, 0, len(s.Orders))
	for _, o := range s.Orders {
		if o < 254 {
			order = append(order, int(o))
		}
	}
	return order
}

func (s *S3M) DefaultSpeed() int {
	return int(s.Header.InitialSpeed)
}

func (s *S3M) DefaultBPM() int {
	return int(s.Header.InitialTempo)
}

// Parse reads an S3M file from an io.Reader and returns a parsed S3M struct.
func Parse(r io.Reader) (*S3M, error) {
	var s3m S3M

	// The reader needs to be an io.ReadSeeker to jump around to read instruments and patterns.
	// If it's not, we'll read the whole thing into memory first.
	seeker, ok := r.(io.ReadSeeker)
	if !ok {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read S3M data: %w", err)
		}
		seeker = bytes.NewReader(data)
	}

	// Read the header
	if err := binary.Read(seeker, binary.LittleEndian, &s3m.Header); err != nil {
		return nil, fmt.Errorf("error reading S3M header: %w", err)
	}

	if s3m.Header.SampleType == 1 {
		s3m.SignedSamples = true
	}

	// Verify signature
	if string(s3m.Header.Signature[:]) != "SCRM" {
		// Not all files have this signature.
	}

	// Calculate number of channels and remap table
	s3m.numChannels = 0
	for i := 0; i < 32; i++ {
		s3m.ChannelRemap[i] = -1 // -1 means disabled
		if s3m.Header.ChannelSettings[i] < 16 {
			s3m.ChannelRemap[i] = s3m.numChannels
			s3m.numChannels++
		}
	}

	// Read orders
	s3m.Orders = make([]byte, s3m.Header.OrderCount)
	if _, err := io.ReadFull(seeker, s3m.Orders); err != nil {
		return nil, fmt.Errorf("error reading orders: %w", err)
	}

	// Find the highest pattern number from the order list to determine the actual number of patterns
	highestPattern := 0
	for _, order := range s3m.Orders {
		if order < 254 { // 254 and 255 are special markers
			if int(order) > highestPattern {
				highestPattern = int(order)
			}
		}
	}
	s3m.ActualPatternCount = highestPattern + 1

	// Read instrument parapointers
	s3m.InstrumentParapointers = make([]uint16, s3m.Header.InstrumentCount)
	if err := binary.Read(seeker, binary.LittleEndian, &s3m.InstrumentParapointers); err != nil {
		return nil, fmt.Errorf("error reading instrument parapointers: %w", err)
	}

	// Read pattern parapointers
	s3m.PatternParapointers = make([]uint16, s3m.Header.PatternCount)
	if err := binary.Read(seeker, binary.LittleEndian, &s3m.PatternParapointers); err != nil {
		return nil, fmt.Errorf("error reading pattern parapointers: %w", err)
	}

	// Read default panning positions if present
	if s3m.Header.DefaultPan == 252 {
		s3m.DefaultPanPositions = make([]byte, 32)
		if _, err := io.ReadFull(seeker, s3m.DefaultPanPositions); err != nil {
			return nil, fmt.Errorf("error reading default pan positions: %w", err)
		}
	}

	// Read instruments
	s3m.Instruments = make([]Instrument, s3m.Header.InstrumentCount)
	for i := 0; i < int(s3m.Header.InstrumentCount); i++ {
		if s3m.InstrumentParapointers[i] == 0 {
			continue
		}
		offset := int64(s3m.InstrumentParapointers[i]) * 16
		if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking to instrument %d: %w", i, err)
		}
		var header instrumentHeader
		if err := binary.Read(seeker, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading instrument %d: %w", i, err)
		}
		s3m.Instruments[i] = Instrument{
			Type:        header.Type,
			DOSFilename: header.DOSFilename,
			MemSeg:      header.MemSeg,
			length:      header.Length,
			LoopBegin:   header.LoopBegin,
			loopEnd:     header.LoopEnd,
			volume:      header.Volume,
			Reserved:    header.Reserved,
			Pack:        header.Pack,
			flags:       header.Flags,
			C2Spd:       header.C2Spd,
			Reserved2:   header.Reserved2,
			SampleName:  header.SampleName,
			Signature:   header.Signature,
			Signed:      s3m.SignedSamples,
		}
	}

	// Read patterns
	s3m.Patterns = make([]Pattern, s3m.Header.PatternCount)
	for i := 0; i < int(s3m.Header.PatternCount); i++ {
		if s3m.PatternParapointers[i] == 0 {
			s3m.Patterns[i] = make(Pattern, 64)
			for row := 0; row < 64; row++ {
				s3m.Patterns[i][row] = make([]PatternEntry, s3m.numChannels)
			}
			continue // Skip empty patterns
		}
		offset := int64(s3m.PatternParapointers[i]) * 16
		if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking to pattern %d: %w", i, err)
		}

		var packedLength uint16
		if err := binary.Read(seeker, binary.LittleEndian, &packedLength); err != nil {
			return nil, fmt.Errorf("reading pattern %d packed length: %w", i, err)
		}

		patternData := make([]byte, packedLength)
		if _, err := io.ReadFull(seeker, patternData); err != nil {
			return nil, fmt.Errorf("reading pattern %d data: %w", i, err)
		}

		patternReader := bytes.NewReader(patternData)

		s3m.Patterns[i] = make(Pattern, 64)
		for row := 0; row < 64; row++ {
			s3m.Patterns[i][row] = make([]PatternEntry, s3m.numChannels)
		}

		row := 0
		for row < 64 {
			what, err := patternReader.ReadByte()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("reading pattern %d row %d what: %w", i, row, err)
			}

			if what == 0 {
				row++
				continue
			}

			channel := int(what & 31)

			var entry PatternEntry
			entry.Volume = 255

			if what&32 != 0 {
				note, _ := patternReader.ReadByte()
				instrument, _ := patternReader.ReadByte()
				entry.Note = note
				entry.Instrument = instrument
			}
			if what&64 != 0 {
				volume, _ := patternReader.ReadByte()
				entry.Volume = volume
			}
			if what&128 != 0 {
				effect, _ := patternReader.ReadByte()
				effectParam, _ := patternReader.ReadByte()
				entry.Effect = module.Effect{
					Command: effect,
					X:       effectParam >> 4,
					Y:       effectParam & 0x0F,
				}
			}

			if s3m.ChannelRemap[channel] != -1 {
				remappedChannel := s3m.ChannelRemap[channel]
				s3m.Patterns[i][row][remappedChannel] = entry
			}
		}
	}

	// Read sample data
	for i, inst := range s3m.Instruments {
		if inst.Type != 1 { // Not a sample-based instrument
			continue
		}
		if inst.length == 0 {
			continue
		}

		// Memseg is a 24-bit pointer. High byte, then low word.
		memsegVal := (uint32(inst.MemSeg[0]) << 16) | uint32(binary.LittleEndian.Uint16(inst.MemSeg[1:3]))
		offset := int64(memsegVal) * 16

		if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking to sample data for instrument %d: %w", i, err)
		}

		data := make([]byte, inst.length)
		if _, err := io.ReadFull(seeker, data); err != nil {
			return nil, fmt.Errorf("reading sample data for instrument %d: %w", i, err)
		}

		is16Bit := inst.flags&4 != 0
		if is16Bit {
			s3m.Instruments[i].data = make([]int16, len(data)/2)
			for j := 0; j < len(data)/2; j++ {
				sampleValue := int32(binary.LittleEndian.Uint16(data[j*2:]))
				if !s3m.SignedSamples {
					sampleValue -= 32768
				}
				s3m.Instruments[i].data[j] = int16(sampleValue)
			}
		} else {
			s3m.Instruments[i].data = make([]int16, len(data))
			for j, v := range data {
				sampleValue := int16(v)
				if !s3m.SignedSamples {
					sampleValue -= 128
				}
				s3m.Instruments[i].data[j] = sampleValue << 8
			}
		}
	}

	return &s3m, nil
}

// PatternCell returns a generic representation of a pattern cell.
func (s *S3M) PatternCell(pattern, row, channel int) module.Cell {
	if pattern >= len(s.Patterns) || row >= 64 || channel >= s.numChannels {
		return module.Cell{}
	}
	p := s.Patterns[pattern]
	if row >= len(p) || channel >= len(p[row]) {
		return module.Cell{}
	}
	cell := p[row][channel]
	return module.Cell{
		HumanNote:   NoteToString(cell.Note),
		Note:        cell.Note,
		Instrument:  cell.Instrument,
		Volume:      cell.Volume,
		Effect:      cell.Effect.Command,
		EffectParam: cell.Effect.X<<4 | cell.Effect.Y,
	}
}

// Read parses an S3M file from an *os.File and returns a module.Module.
func Read(file *os.File) (module.Module, error) {
	return Parse(file)
}
