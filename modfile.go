/*
Offset  Bytes  Description
------  -----  -----------

	0     20    Songname. Remember to put trailing null bytes at the end...

Information for sample 1-31:

Offset  Bytes  Description
------  -----  -----------

	20     22    Samplename for sample 1. Pad with null bytes.
	42      2    Samplelength for sample 1. Stored as number of words.
	             Multiply by two to get real sample length in bytes.
	44      1    Lower four bits are the finetune value, stored as a signed
	             four bit number. The upper four bits are not used, and
	             should be set to zero.
	             Value:  Finetune:
	               0        0
	               1       +1
	               2       +2
	               3       +3
	               4       +4
	               5       +5
	               6       +6
	               7       +7
	               8       -8
	               9       -7
	               A       -6
	               B       -5
	               C       -4
	               D       -3
	               E       -2
	               F       -1

	45      1    Volume for sample 1. Range is $00-$40, or 0-64 decimal.
	46      2    Repeat point for sample 1. Stored as number of words offset
	             from start of sample. Multiply by two to get offset in bytes.
	48      2    Repeat Length for sample 1. Stored as number of words in
	             loop. Multiply by two to get replen in bytes.

Information for the next 30 samples starts here. It's just like the info for
sample 1.

Offset  Bytes  Description
------  -----  -----------

	 50     30    Sample 2...
	 80     30    Sample 3...
	  .
	  .
	  .
	890     30    Sample 30...
	920     30    Sample 31...

Offset  Bytes  Description
------  -----  -----------

	950      1    Songlength. Range is 1-128.
	951      1    Well... this little byte here is set to 127, so that old
	              trackers will search through all patterns when loading.
	              Noisetracker uses this byte for restart, but we don't.
	952    128    Song positions 0-127. Each hold a number from 0-63 that
	              tells the tracker what pattern to play at that position.

1080      4    The four letters "M.K." - This is something Mahoney & Kaktus

	inserted when they increased the number of samples from
	15 to 31. If it's not there, the module/song uses 15 samples
	or the text has been removed to make the module harder to
	rip. Startrekker puts "FLT4" or "FLT8" there instead.

Offset  Bytes  Description
------  -----  -----------
1084    1024   Data for pattern 00.

	.
	.
	.

xxxx  Number of patterns stored is equal to the highest patternnumber

	in the song position table (at offset 952-1079).

Each note is stored as 4 bytes, and all four notes at each position in
the pattern are stored after each other.

00 -  chan1  chan2  chan3  chan4
01 -  chan1  chan2  chan3  chan4
02 -  chan1  chan2  chan3  chan4
etc.

Info for each note:

	_____byte 1_____   byte2_    _____byte 3_____   byte4_

/                \ /      \  /                \ /      \
0000          0000-00000000  0000          0000-00000000

Upper four    12 bits for    Lower four    Effect command.
bits of sam-  note period.   bits of sam-
ple number.                  ple number.

Periodtable for Tuning 0, Normal

	C-1 to B-1 : 856,808,762,720,678,640,604,570,538,508,480,453
	C-2 to B-2 : 428,404,381,360,339,320,302,285,269,254,240,226
	C-3 to B-3 : 214,202,190,180,170,160,151,143,135,127,120,113

To determine what note to show, scan through the table until you find
the same period as the one stored in byte 1-2. Use the index to look
up in a notenames table.

This is the data stored in a normal song. A packed song starts with the
four letters "PACK", and then comes the packed data.

In a module, all the samples are stored right after the patterndata.
To determine where a sample starts and stops, you use the sampleinfo
structures in the beginning of the file (from offset 20). Take a look
at the mt_init routine in the playroutine, and you'll see just how it
is done.
*/
package modfile

import "strings"

const (
	lengthSongName    uint16 = 20
	lengthSampleChunk uint16 = 30
	lengthSampleName  uint16 = 22
	lengthNote        uint16 = 4
	lengthMagicFormat uint8  = 4

	// key offset markers
	offsetModuleNameStart        uint32 = 0
	offsetSampleMetaDataStart    uint32 = 20
	offsetNumberOfPatternsInSong uint32 = 950
	offsetSongEndJump            uint32 = 951
	offsetPatternTableStart      uint32 = 952
	offsetModuleFormatMagic      uint32 = 1080
	offsetPatternDataStart       uint32 = 1084

	rowsPerPattern      uint8 = 64
	maxNumberOfPatterns uint8 = 128
)

// http://coppershade.org/articles/More!/Topics/Protracker_File_Format/

type Mod struct {
	Name    string
	Samples []*Sample
	// Songlength represents the number of patterns which are played in the entire song.
	// This is NOT the number of patterns in the file
	Songlength uint8
	// SequenceTable is an index that defines the sequencer table. This stores the Pattern number
	// at the index in which it is played
	SequenceTable    []uint8
	numberOfPatterns uint8 // the number of (unique) patterns that make up this song
	Format           ModuleFormat
	Patterns         []*Pattern
}

type Pattern struct {
	rowChannel [][]*Note // row count and then channel
}

// NoteTable is the index of notes per frequency table below
var NoteTable = [12]string{"B-", "A#", "A-", "G#", "G-", "F#", "F-", "E-", "D#", "D-", "C#", "C-"}

// FrequencyTable maps  frequencies for notes/octaves
var FrequencyTable = []int{
	// B, A#, A, G#, G, F#, F, E, D#, D, C#, C
	57, 60, 64, 67, 71, 76, 80, 85, 90, 95, 101, 107,
	113, 120, 127, 135, 143, 151, 160, 170, 180, 190, 202, 214,
	226, 240, 254, 269, 285, 302, 320, 339, 360, 381, 404, 428,
	453, 480, 508, 538, 570, 604, 640, 678, 720, 762, 808, 856,
	907, 961, 1017, 1077, 1141, 1209, 1281, 1357, 1440, 1525, 1616, 1712,
}

// patternSizeBytes calculates the size of the pattern for the module type. For example
// For a four-channel file there are (4 bytes * 4 channels * 64 lines) =1024 bytes of information per pattern
func (m *Mod) patternDataSizeBytes() uint32 {
	return uint32(m.Format.Channels) * uint32(rowsPerPattern) * uint32(lengthNote) * uint32(m.numberOfPatterns)
}

func Read(buffer []byte) (*Mod, error) {
	var err error

	pt := new(Mod)
	pt.Format, err = DetectModFileFormat(buffer)
	if err != nil {
		return nil, err
	}

	pt.Name = string(buffer[offsetModuleNameStart:lengthSongName])
	pt.Name = strings.Replace(pt.Name, "\x00", "", -1)

	offset := offsetSampleMetaDataStart
	pt.Samples = make([]*Sample, pt.Format.Samples)
	for i := uint8(0); i < pt.Format.Samples; i++ {
		pt.Samples[i] = NewSample(buffer[offset : offset+uint32(lengthSampleChunk)])
		offset += uint32(lengthSampleChunk)
	}

	pt.Songlength = buffer[offsetNumberOfPatternsInSong]
	pt.numberOfPatterns = 0
	pt.SequenceTable = make([]uint8, maxNumberOfPatterns)
	// Song positions 0-127. Each hold a number from 0-63 that tells the tracker what pattern to play at that position
	for i := uint8(0); i < maxNumberOfPatterns; i++ {
		pt.SequenceTable[i] = uint8(buffer[offsetPatternTableStart+uint32(i)])
		// Number of patterns stored is equal to the highest patternnumber in the song position table (at offset 952-1079)
		if pt.SequenceTable[i] > pt.numberOfPatterns {
			pt.numberOfPatterns = pt.SequenceTable[i]
		}
	}

	offset = offsetPatternDataStart
	pt.Patterns = make([]*Pattern, pt.numberOfPatterns)
	for i := uint8(0); i < pt.numberOfPatterns; i++ {
		pattern := new(Pattern)
		pattern.rowChannel = make([][]*Note, rowsPerPattern)
		for row := uint8(0); row < rowsPerPattern; row++ {
			pattern.rowChannel[row] = make([]*Note, pt.Format.Channels)
			for channel := uint8(0); channel < pt.Format.Channels; channel++ {
				pattern.rowChannel[row][channel] = NewNote(buffer[offset : offset+uint32(lengthNote)])
				offset += uint32(lengthNote)
			}
		}
		pt.Patterns[i] = pattern
	}

	offset = offsetPatternDataStart + pt.patternDataSizeBytes() // sample data falls after the pattern data

	for i := uint8(0); i < pt.Format.Samples; i++ {
		pt.Samples[i].AudioData = buffer[offset : offset+uint32(pt.Samples[i].SampleLength)]
		offset += uint32(pt.Samples[i].SampleLength)
	}
	return pt, nil
}
