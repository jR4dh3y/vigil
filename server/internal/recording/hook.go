package recording

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// flexFloat unmarshals JSON numbers or numeric strings (MediaMTX env expansion
// injects strings into our curl JSON body).
type flexFloat float64

func (f *flexFloat) UnmarshalJSON(b []byte) error {
	b = bytesTrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = flexFloat(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = flexFloat(v)
	return nil
}

func (f flexFloat) Float64() float64 { return float64(f) }

// flexInt64 unmarshals JSON numbers or numeric strings into int64.
type flexInt64 int64

func (n *flexInt64) UnmarshalJSON(b []byte) error {
	b = bytesTrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*n = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*n = 0
			return nil
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Accept float-looking strings for size.
			f, err2 := strconv.ParseFloat(s, 64)
			if err2 != nil {
				return err
			}
			*n = flexInt64(int64(f))
			return nil
		}
		*n = flexInt64(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*n = flexInt64(int64(v))
	return nil
}

func (n flexInt64) Int64() int64 { return int64(n) }

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

// segmentCompletePayload is a flexible body for MediaMTX runOnRecordSegmentComplete.
// MediaMTX can POST JSON or form fields; field names vary by version/config.
type segmentCompletePayload struct {
	Path         string    `json:"path"`
	PathName     string    `json:"path_name"`
	PathName2    string    `json:"pathName"`
	FilePath     string    `json:"file_path"`
	FilePath2    string    `json:"filePath"`
	Segment      string    `json:"segment_path"`
	Segment2     string    `json:"segmentPath"`
	DurationMS   flexFloat `json:"duration_ms"`
	DurationMS2  flexFloat `json:"durationMs"`
	Duration     flexFloat `json:"duration"`
	DurationSec  flexFloat `json:"duration_sec"`
	DurationSec2 flexFloat `json:"durationSec"`
	StartTime    string    `json:"start_time"`
	StartTime2   string    `json:"startTime"`
	Start        string    `json:"start"`
	Codec        string    `json:"codec"`
	SizeBytes    flexInt64 `json:"size_bytes"`
	SizeBytes2   flexInt64 `json:"sizeBytes"`
	Size         flexInt64 `json:"size"`
}

// SegmentCompleteHandler returns POST /internal/mediamtx/segment-complete.
// Accepts JSON or form; best-effort field mapping; indexes the segment.
func (s *Service) SegmentCompleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		payload, err := decodeSegmentComplete(r)
		if err != nil {
			slog.Warn("segment-complete decode", "err", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		pathName := firstNonEmpty(payload.Path, payload.PathName, payload.PathName2)
		filePath := firstNonEmpty(payload.FilePath, payload.FilePath2, payload.Segment, payload.Segment2)
		if pathName == "" && filePath == "" {
			slog.Warn("segment-complete missing path and file_path")
			http.Error(w, "path or file_path required", http.StatusBadRequest)
			return
		}

		cameraID, ok := CameraIDFromPathName(pathName)
		if !ok && filePath != "" {
			// Try to infer path name from file path components.
			cameraID, ok = cameraIDFromFilePath(filePath)
		}
		if !ok {
			slog.Warn("segment-complete cannot map path to camera", "path", pathName, "file", filePath)
			http.Error(w, "unknown path", http.StatusBadRequest)
			return
		}

		durationSec := payload.DurationSec.Float64()
		if durationSec == 0 {
			durationSec = payload.DurationSec2.Float64()
		}
		if durationSec == 0 && payload.DurationMS.Float64() > 0 {
			durationSec = payload.DurationMS.Float64() / 1000
		}
		if durationSec == 0 && payload.DurationMS2.Float64() > 0 {
			durationSec = payload.DurationMS2.Float64() / 1000
		}
		if durationSec == 0 && payload.Duration.Float64() > 0 {
			// Heuristic: values > 1000 are likely ms.
			d := payload.Duration.Float64()
			if d > 1000 {
				durationSec = d / 1000
			} else {
				durationSec = d
			}
		}
		if durationSec == 0 {
			durationSec = 60 // MediaMTX default segment length
		}

		startedAt := time.Now().UTC().Add(-time.Duration(durationSec * float64(time.Second)))
		if st := firstNonEmpty(payload.StartTime, payload.StartTime2, payload.Start); st != "" {
			if t, err := parseTime(st); err == nil {
				startedAt = t
			}
		} else if filePath != "" {
			if t, ok := startedAtFromFileName(filePath); ok {
				startedAt = t
			}
		}

		sizeBytes := payload.SizeBytes.Int64()
		if sizeBytes == 0 {
			sizeBytes = payload.SizeBytes2.Int64()
		}
		if sizeBytes == 0 {
			sizeBytes = payload.Size.Int64()
		}
		if sizeBytes == 0 && filePath != "" {
			if fi, err := os.Stat(filePath); err == nil {
				sizeBytes = fi.Size()
			}
		}

		storePath := filePath
		if storePath == "" {
			// Synthesize preferred layout when only path name is known.
			storePath = filepath.Join(
				cameraID,
				startedAt.UTC().Format("2006-01-02"),
				startedAt.UTC().Format("15-04-05")+".mp4",
			)
		}

		seg, err := s.IndexSegment(r.Context(), cameraID, storePath, startedAt, durationSec, sizeBytes, payload.Codec)
		if err != nil {
			slog.Error("segment-complete index", "err", err, "camera_id", cameraID, "file", filePath)
			http.Error(w, "index failed", http.StatusInternalServerError)
			return
		}

		slog.Info("indexed recording segment",
			"id", seg.ID,
			"camera_id", cameraID,
			"path", seg.Path,
			"duration_sec", seg.DurationSec,
			"size_bytes", seg.SizeBytes,
		)
		w.WriteHeader(http.StatusNoContent)
	}
}

func decodeSegmentComplete(r *http.Request) (segmentCompletePayload, error) {
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	var p segmentCompletePayload

	if strings.Contains(ct, "application/json") || ct == "" {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			return p, err
		}
		if len(strings.TrimSpace(string(body))) == 0 {
			// Fall through to empty payload; query params may still help.
		} else if err := json.Unmarshal(body, &p); err != nil {
			// Try form-encoded body as fallback.
			if strings.Contains(ct, "form") || lookLikeForm(body) {
				return parseFormPayload(string(body)), nil
			}
			return p, err
		}
		// Overlay query params for partial JSON.
		overlayQuery(&p, r)
		return p, nil
	}

	if strings.Contains(ct, "form") {
		if err := r.ParseForm(); err != nil {
			return p, err
		}
		p = formToPayload(r.Form)
		overlayQuery(&p, r)
		return p, nil
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return p, err
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &p); err == nil {
			overlayQuery(&p, r)
			return p, nil
		}
		if lookLikeForm(body) {
			p = parseFormPayload(string(body))
			overlayQuery(&p, r)
			return p, nil
		}
	}
	overlayQuery(&p, r)
	return p, nil
}

func lookLikeForm(body []byte) bool {
	s := string(body)
	return strings.Contains(s, "=") && !strings.HasPrefix(strings.TrimSpace(s), "{")
}

func parseFormPayload(raw string) segmentCompletePayload {
	vals := make(map[string][]string)
	for _, part := range strings.Split(raw, "&") {
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		// Minimal query unescape.
		k = strings.ReplaceAll(k, "+", " ")
		v = strings.ReplaceAll(v, "+", " ")
		vals[k] = append(vals[k], v)
	}
	return formToPayload(vals)
}

func formToPayload(vals map[string][]string) segmentCompletePayload {
	get := func(keys ...string) string {
		for _, k := range keys {
			if v := vals[k]; len(v) > 0 && v[0] != "" {
				return v[0]
			}
		}
		return ""
	}
	getFloat := func(keys ...string) float64 {
		s := get(keys...)
		if s == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	getInt := func(keys ...string) int64 {
		s := get(keys...)
		if s == "" {
			return 0
		}
		n, _ := strconv.ParseInt(s, 10, 64)
		return n
	}
	return segmentCompletePayload{
		Path:        get("path", "path_name", "pathName", "MTX_PATH"),
		FilePath:    get("file_path", "filePath", "segment_path", "segmentPath", "MTX_SEGMENT_PATH"),
		DurationMS:  flexFloat(getFloat("duration_ms", "durationMs")),
		DurationSec: flexFloat(getFloat("duration_sec", "durationSec")),
		Duration:    flexFloat(getFloat("duration")),
		StartTime:   get("start_time", "startTime", "start", "MTX_SEGMENT_START"),
		Codec:       get("codec"),
		SizeBytes:   flexInt64(getInt("size_bytes", "sizeBytes", "size")),
	}
}

func overlayQuery(p *segmentCompletePayload, r *http.Request) {
	q := r.URL.Query()
	if p.Path == "" {
		p.Path = firstNonEmpty(q.Get("path"), q.Get("path_name"), q.Get("pathName"))
	}
	if p.FilePath == "" {
		p.FilePath = firstNonEmpty(q.Get("file_path"), q.Get("filePath"), q.Get("segment_path"))
	}
	if p.StartTime == "" {
		p.StartTime = firstNonEmpty(q.Get("start_time"), q.Get("startTime"), q.Get("start"))
	}
	if p.Codec == "" {
		p.Codec = q.Get("codec")
	}
	if p.DurationMS == 0 {
		if v := q.Get("duration_ms"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.DurationMS = flexFloat(f)
			}
		}
	}
	if p.DurationSec == 0 {
		if v := q.Get("duration_sec"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.DurationSec = flexFloat(f)
			}
		}
	}
	if p.SizeBytes == 0 {
		if v := q.Get("size_bytes"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				p.SizeBytes = flexInt64(n)
			}
		}
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// cameraIDFromFilePath tries path components: .../{uuid}/... or .../cam_{hex}/...
func cameraIDFromFilePath(filePath string) (string, bool) {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for _, part := range parts {
		if id, ok := CameraIDFromPathName(part); ok {
			return id, true
		}
	}
	return "", false
}

// startedAtFromFileName parses timestamps commonly used by MediaMTX recordPath.
// Examples: 15-04-05.mp4, 2024-01-02_15-04-05-000000.mp4, 15-04-05-123456.mp4
func startedAtFromFileName(filePath string) (time.Time, bool) {
	base := filepath.Base(filePath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	// Drop fractional suffix after last long numeric run if needed.
	candidates := []string{base}
	// If parent is a date directory, combine.
	dir := filepath.Base(filepath.Dir(filePath))
	if len(dir) == 10 && dir[4] == '-' && dir[7] == '-' {
		candidates = append(candidates, dir+"_"+base, dir+"T"+base)
	}

	layouts := []string{
		"2006-01-02_15-04-05",
		"2006-01-02_15-04-05-000000",
		"2006-01-02_15-04-05.000000",
		"2006-01-02T15-04-05",
		"2006-01-02T15:04:05",
		"15-04-05",
		"15-04-05-000000",
	}
	for _, c := range candidates {
		// Truncate trailing fractional -NNNNNN if present with date prefix.
		trimmed := c
		if i := strings.LastIndex(trimmed, "-"); i > 0 {
			tail := trimmed[i+1:]
			if len(tail) >= 3 && isAllDigits(tail) {
				// keep as full string first; also try without fraction
				candidates = append(candidates, trimmed[:i])
			}
		}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, trimmed, time.UTC); err == nil {
				// Time-only layouts use year 0; fill date from parent if needed.
				if t.Year() == 0 && len(dir) == 10 {
					if d, err := time.ParseInLocation("2006-01-02", dir, time.UTC); err == nil {
						t = time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
					}
				}
				if t.Year() > 0 {
					return t.UTC(), true
				}
			}
		}
	}
	return time.Time{}, false
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
