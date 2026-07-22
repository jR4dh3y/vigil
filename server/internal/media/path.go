package media

import "strings"

// PathName returns the MediaMTX path name for a camera.
// Format: cam_{uuid without dashes}.
func PathName(cameraID string) string {
	id := strings.ReplaceAll(strings.TrimSpace(cameraID), "-", "")
	return "cam_" + id
}
