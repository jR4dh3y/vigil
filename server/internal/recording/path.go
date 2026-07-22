package recording

import (
	"strings"

	"github.com/google/uuid"
)

// CameraIDFromPathName maps a MediaMTX path name (cam_{uuid-without-dashes})
// back to a canonical UUID string. Also accepts a raw UUID.
func CameraIDFromPathName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "/")
	if i := strings.IndexByte(name, '/'); i >= 0 {
		name = name[:i]
	}
	if name == "" {
		return "", false
	}

	// Already a UUID?
	if id, err := uuid.Parse(name); err == nil {
		return id.String(), true
	}

	// Strip cam_ prefix used by media.PathName.
	raw := name
	if strings.HasPrefix(strings.ToLower(raw), "cam_") {
		raw = raw[4:]
	}
	raw = strings.ReplaceAll(raw, "-", "")
	if len(raw) != 32 {
		return "", false
	}
	// Validate hex and rehydrate 8-4-4-4-12.
	for _, c := range raw {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return "", false
		}
	}
	hydrated := raw[0:8] + "-" + raw[8:12] + "-" + raw[12:16] + "-" + raw[16:20] + "-" + raw[20:32]
	id, err := uuid.Parse(hydrated)
	if err != nil {
		return "", false
	}
	return id.String(), true
}
