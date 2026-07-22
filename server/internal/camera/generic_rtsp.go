package camera

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const probeTimeout = 8 * time.Second

// GenericRTSPDriver probes RTSP streams via ffprobe when available on PATH.
type GenericRTSPDriver struct{}

// NewGenericRTSPDriver returns the default RTSP driver.
func NewGenericRTSPDriver() *GenericRTSPDriver {
	return &GenericRTSPDriver{}
}

// Probe runs ffprobe against rtspURL. If username/password are provided and the
// URL has no userinfo, they are injected as rtsp://user:pass@host/...
func (d *GenericRTSPDriver) Probe(ctx context.Context, rtspURL, user, pass string) (ProbeResult, error) {
	url := injectRTSPCredentials(rtspURL, user, pass)

	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return ProbeResult{
			Reachable: false,
			Error:     "ffprobe not found on PATH",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	// -rtsp_transport tcp is more reliable through NATs/firewalls for probes.
	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "error",
		"-rtsp_transport", "tcp",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,width,height",
		"-of", "json",
		url,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		if ctx.Err() != nil {
			msg = "ffprobe timed out"
		}
		return ProbeResult{
			Reachable: false,
			Error:     msg,
		}, nil
	}

	var payload struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		return ProbeResult{
			Reachable: false,
			Error:     fmt.Sprintf("parse ffprobe output: %v", err),
		}, nil
	}
	if len(payload.Streams) == 0 {
		return ProbeResult{
			Reachable: false,
			Error:     "no video stream found",
		}, nil
	}

	s := payload.Streams[0]
	return ProbeResult{
		Reachable: true,
		Codec:     s.CodecName,
		Width:     s.Width,
		Height:    s.Height,
		H265:      isH265(s.CodecName),
	}, nil
}

func isH265(codec string) bool {
	c := strings.ToLower(strings.TrimSpace(codec))
	return c == "hevc" || c == "h265" || strings.Contains(c, "h.265") || strings.Contains(c, "h265")
}

// injectRTSPCredentials adds userinfo to an RTSP URL when missing.
func injectRTSPCredentials(rtspURL, user, pass string) string {
	if user == "" && pass == "" {
		return rtspURL
	}
	// Already has userinfo.
	if strings.Contains(rtspURL, "@") {
		return rtspURL
	}
	const prefix = "rtsp://"
	if !strings.HasPrefix(strings.ToLower(rtspURL), prefix) {
		return rtspURL
	}
	rest := rtspURL[len(prefix):]
	// Escape is intentionally minimal for phase 2; special chars in passwords may need encoding later.
	return prefix + user + ":" + pass + "@" + rest
}
