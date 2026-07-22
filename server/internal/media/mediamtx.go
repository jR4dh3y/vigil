package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MediaMTXClient talks to the MediaMTX Control API (v3).
type MediaMTXClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMediaMTXClient constructs a client for the given Control API base URL
// (e.g. http://127.0.0.1:9997). Trailing slashes are stripped.
func NewMediaMTXClient(baseURL string) *MediaMTXClient {
	return &MediaMTXClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PathRecordOptions configures continuous recording for a MediaMTX path.
type PathRecordOptions struct {
	// Enabled turns on native recording for the path.
	Enabled bool
	// RecordPath is the MediaMTX recordPath template (e.g. /rec/%path/%Y-%m-%d/%H-%M-%S-%f).
	RecordPath string
	// SegmentDuration is e.g. "1m" (MediaMTX duration string). Empty defaults to "1m".
	SegmentDuration string
	// Format is e.g. "fmp4". Empty defaults to "fmp4".
	Format string
}

// pathConf is the JSON body for path add/replace (subset of MediaMTX PathConf).
type pathConf struct {
	Source                     string `json:"source"`
	SourceOnDemand             bool   `json:"sourceOnDemand"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout,omitempty"`
	// rtspTransport: automatic|udp|multicast|tcp — TCP is most reliable for DVRs/NATs.
	RTSPTransport         string `json:"rtspTransport,omitempty"`
	Record                bool   `json:"record"`
	RecordPath            string `json:"recordPath,omitempty"`
	RecordFormat          string `json:"recordFormat,omitempty"`
	RecordSegmentDuration string `json:"recordSegmentDuration,omitempty"`
}

// UpsertPath creates or replaces a MediaMTX path that pulls sourceRTSP on demand.
// When rec.Enabled is true, continuous fMP4 recording is configured under rec.RecordPath.
// Connection failures return a clear error; they do not panic.
func (c *MediaMTXClient) UpsertPath(ctx context.Context, name, sourceRTSP string, rec PathRecordOptions) error {
	if c.baseURL == "" {
		return fmt.Errorf("mediamtx api url not configured")
	}
	body := pathConf{
		Source:                     sourceRTSP,
		SourceOnDemand:             true,
		SourceOnDemandStartTimeout: "20s",
		RTSPTransport:              "tcp",
		Record:                     rec.Enabled,
	}
	if rec.Enabled {
		body.RecordPath = rec.RecordPath
		body.RecordFormat = rec.Format
		if body.RecordFormat == "" {
			body.RecordFormat = "fmp4"
		}
		body.RecordSegmentDuration = rec.SegmentDuration
		if body.RecordSegmentDuration == "" {
			body.RecordSegmentDuration = "1m"
		}
	}
	// Prefer replace (idempotent). If path is missing, add it.
	status, errBody, err := c.doJSON(ctx, http.MethodPost, "/v3/config/paths/replace/"+url.PathEscape(name), body)
	if err != nil {
		return fmt.Errorf("mediamtx replace path %q: %w", name, err)
	}
	if status >= 200 && status < 300 {
		return nil
	}
	// Not found / unknown path → add
	if status == http.StatusNotFound || status == http.StatusBadRequest {
		status2, errBody2, err2 := c.doJSON(ctx, http.MethodPost, "/v3/config/paths/add/"+url.PathEscape(name), body)
		if err2 != nil {
			return fmt.Errorf("mediamtx add path %q: %w", name, err2)
		}
		if status2 >= 200 && status2 < 300 {
			return nil
		}
		// Already exists → replace again (race)
		if status2 == http.StatusBadRequest {
			status3, errBody3, err3 := c.doJSON(ctx, http.MethodPost, "/v3/config/paths/replace/"+url.PathEscape(name), body)
			if err3 != nil {
				return fmt.Errorf("mediamtx replace path %q (retry): %w", name, err3)
			}
			if status3 >= 200 && status3 < 300 {
				return nil
			}
			return fmt.Errorf("mediamtx replace path %q: status %d: %s", name, status3, errBody3)
		}
		return fmt.Errorf("mediamtx add path %q: status %d: %s", name, status2, errBody2)
	}
	return fmt.Errorf("mediamtx replace path %q: status %d: %s", name, status, errBody)
}

// DeletePath removes a path configuration from MediaMTX.
// Missing paths are treated as success.
func (c *MediaMTXClient) DeletePath(ctx context.Context, name string) error {
	if c.baseURL == "" {
		return fmt.Errorf("mediamtx api url not configured")
	}
	status, errBody, err := c.doJSON(ctx, http.MethodDelete, "/v3/config/paths/delete/"+url.PathEscape(name), nil)
	if err != nil {
		return fmt.Errorf("mediamtx delete path %q: %w", name, err)
	}
	if status >= 200 && status < 300 || status == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("mediamtx delete path %q: status %d: %s", name, status, errBody)
}

func (c *MediaMTXClient) doJSON(ctx context.Context, method, path string, body any) (status int, respBody string, err error) {
	var rdr io.Reader
	if body != nil {
		b, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return 0, "", marshalErr
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return 0, "", err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, strings.TrimSpace(string(raw)), nil
}
