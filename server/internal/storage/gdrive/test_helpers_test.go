package gdrive

import (
	"os"
)

func writeTestRecording(path string) error {
	return os.WriteFile(path, []byte("recording"), 0o600)
}
