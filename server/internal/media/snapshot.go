package media

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

const snapshotTimeout = 10 * time.Second

// captureSnapshot runs ffmpeg against an RTSP URL and returns a single JPEG frame.
func captureSnapshot(ctx context.Context, rtspURL string) ([]byte, error) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found on PATH")
	}

	ctx, cancel := context.WithTimeout(ctx, snapshotTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-y",
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-frames:v", "1",
		"-f", "image2",
		"pipe:1",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := bytes.TrimSpace(stderr.Bytes())
		if len(msg) == 0 {
			msg = []byte(err.Error())
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("ffmpeg snapshot timed out")
		}
		return nil, fmt.Errorf("ffmpeg snapshot: %s", msg)
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("ffmpeg produced empty snapshot")
	}
	return stdout.Bytes(), nil
}
