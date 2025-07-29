package loader

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/jesseward/impulse/pkg/module"
	"github.com/jesseward/impulse/pkg/protracker"
	"github.com/jesseward/impulse/pkg/s3m"
	"github.com/jesseward/impulse/pkg/xm"
)

var (
	MagicMK   = []byte{'M', '.', 'K', '.'}
	MagicM4   = []byte{'M', '!', 'K', '!'}
	MagicFLT4 = []byte{'F', 'L', 'T', '4'}
	Magic4CHN = []byte{'4', 'C', 'H', 'N'}
	Magic6CHN = []byte{'6', 'C', 'H', 'N'}
	Magic8CHN = []byte{'8', 'C', 'H', 'N'}
	MagicFLT8 = []byte{'F', 'L', 'T', '8'}
	MagicSCRM = []byte{'S', 'C', 'R', 'M'}
	MagicXM   = []byte{'E', 'x', 't', 'e', 'n', 'd', 'e', 'd', ' ', 'M', 'o', 'd', 'u', 'l', 'e', ':', ' '}
)

// Load detects the file type of a music module and loads it.
func Load(file *os.File) (module.Module, error) {
	// Read the first 1084 bytes of the file, which should be enough to identify the file type.
	buffer := make([]byte, 1084)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Reset the file pointer to the beginning of the file.
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	// Check for XM magic at offset 0
	if len(buffer) >= 17 {
		if bytes.Equal(buffer[0:17], MagicXM) {
			return xm.Read(file)
		}
	}

	// Check for S3M magic number at offset 44
	if len(buffer) >= 48 {
		if bytes.Equal(buffer[44:48], MagicSCRM) {
			return s3m.Read(file)
		}
	}

	// Check for MOD magic number at offset 1080
	if len(buffer) >= 1084 {
		if bytes.Equal(buffer[1080:1084], MagicMK) ||
			bytes.Equal(buffer[1080:1084], MagicM4) ||
			bytes.Equal(buffer[1080:1084], MagicFLT4) ||
			bytes.Equal(buffer[1080:1084], Magic4CHN) {
			return protracker.Read(file)
		}
	}

	return nil, errors.New("unknown file type")
}
