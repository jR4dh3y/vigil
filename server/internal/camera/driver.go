package camera

import "context"

// ProbeResult is the outcome of probing an RTSP URL for reachability and stream metadata.
type ProbeResult struct {
	Reachable bool
	Codec     string
	Width     int
	Height    int
	Error     string
	H265      bool
}

// Driver abstracts camera-vendor behavior. Phase 2 only needs Probe.
type Driver interface {
	Probe(ctx context.Context, rtspURL, user, pass string) (ProbeResult, error)
}
