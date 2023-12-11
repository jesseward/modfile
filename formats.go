package modfile

import "fmt"

const (
	MagicTagMK   string = "M.K." // standard 4-channel, 64-pattern-max MOD.
	MagicTag6CHN string = "6CHN"
	MagicTag8CHN string = "8CHN"
	MagicTag2CHN string = "2CHN" // 2 channel fast tracker
)

var (
	ModFormat2Ch = ModuleFormat{
		Name:        MagicTag2CHN,
		Description: "2 Channel FastTracker",
		Channels:    2,
		Samples:     31,
	}
	ModFormatMK = ModuleFormat{
		Name:        MagicTagMK,
		Description: "Standard 4 channel, 64 pattern",
		Samples:     31,
		Channels:    4,
	}

	ModFormat6Chan = ModuleFormat{
		Name:        MagicTag6CHN,
		Description: "6 channel module",
		Samples:     31,
		Channels:    6,
	}

	ModFormat8Chan = ModuleFormat{
		Name:        MagicTag8CHN,
		Description: "8 channel module",
		Samples:     31,
		Channels:    6,
	}

	formatMap = map[string]ModuleFormat{
		MagicTag2CHN: ModFormat2Ch,
		MagicTag6CHN: ModFormat6Chan,
		MagicTag8CHN: ModFormat8Chan,
		MagicTagMK:   ModFormatMK,
	}
)

type ModuleFormat struct {
	Name        string // 4 byte format name. This is set to the magic value/string
	Description string
	Channels    uint8 // Number of channels for this mod format
	Samples     uint8 // max number of samples for the format
}

// Returns the version of the ProTracker file format. If the file is not recognized, a 0 is returned
func DetectModFileFormat(buffer []byte) (*ModuleFormat, error) {

	if len(buffer) < int(offsetModuleFormatMagic+uint32(lengthMagicFormat)) {
		return nil, fmt.Errorf("incorrect buffer size, buff is of size=%d, expected=%d", len(buffer), lengthMagicFormat)
	}

	if modFormat, ok := formatMap[string(buffer[offsetModuleFormatMagic:offsetModuleFormatMagic+4])]; ok {
		return &modFormat, nil
	}

	return nil, fmt.Errorf("unrecognized module file format, received buffer='%v'", buffer)
}
