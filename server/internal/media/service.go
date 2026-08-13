package media

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nvr/nvr/server/internal/camera"
)

// Domain errors for the media plane.
var (
	ErrCameraDisabled = errors.New("camera disabled")
	ErrNoLiveSource   = errors.New("no live stream source")
	ErrSnapshotFailed = errors.New("snapshot failed")
)

// ArchivedPlaybackTokenTTL allows a browser to make follow-up range requests
// while seeking through a Drive-backed one-minute segment.
const ArchivedPlaybackTokenTTL = 15 * time.Minute

const defaultRetentionDays = 7

// LiveStream is the live playback bundle returned to API clients.
type LiveStream struct {
	CameraID  string
	WHEPURL   string
	HLSURL    string
	Token     string
	ExpiresAt time.Time
}

// PlaybackStream is a short-lived playback session for recorded video.
type PlaybackStream struct {
	CameraID    string
	PlaybackURL string
	Token       string
	ExpiresAt   time.Time
}

// Config holds browser-facing MediaMTX bases, control API URL, and recording root.
type Config struct {
	APIURL           string // Control API, e.g. http://127.0.0.1:9997
	WebRTCURL        string // WHEP base, e.g. http://127.0.0.1:8889
	HLSURL           string // HLS base, e.g. http://127.0.0.1:8888
	PlaybackURL      string // Optional MediaMTX playback server base, e.g. http://127.0.0.1:9996
	RecordingsDir    string // Root directory for recorded segments
	RecordingEnabled bool   // Continuous recording when true and RecordingsDir is set
	RetentionDays    int    // MediaMTX fallback local retention when archive cleanup is unavailable
}

// CameraReader is the subset of camera.Service used by media.
type CameraReader interface {
	Get(ctx context.Context, id string) (camera.Camera, error)
	LiveSourceURL(ctx context.Context, id string) (string, error)
}

// Service orchestrates MediaMTX paths, stream tokens, and FFmpeg snapshots.
type Service struct {
	mu               sync.RWMutex
	cfg              Config
	recordingEnabled bool
	cams             CameraReader
	mtx              *MediaMTXClient
	tokens           *TokenStore
}

// NewService constructs a media Service.
// Callers should set RecordingEnabled explicitly (main defaults it to true when a dir is configured).
func NewService(cfg Config, cams CameraReader) *Service {
	dir := strings.TrimSpace(cfg.RecordingsDir)
	enabled := cfg.RecordingEnabled && dir != ""
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = defaultRetentionDays
	}
	return &Service{
		cfg: Config{
			APIURL:           strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/"),
			WebRTCURL:        strings.TrimRight(strings.TrimSpace(cfg.WebRTCURL), "/"),
			HLSURL:           strings.TrimRight(strings.TrimSpace(cfg.HLSURL), "/"),
			PlaybackURL:      strings.TrimRight(strings.TrimSpace(cfg.PlaybackURL), "/"),
			RecordingsDir:    dir,
			RecordingEnabled: cfg.RecordingEnabled,
			RetentionDays:    cfg.RetentionDays,
		},
		recordingEnabled: enabled,
		cams:             cams,
		mtx:              NewMediaMTXClient(cfg.APIURL),
		tokens:           NewTokenStore(),
	}
}

// RetentionDays returns the fallback local retention configured for MediaMTX.
func (s *Service) RetentionDays() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.RetentionDays
}

// SetRetentionDays updates the fallback local retention used when camera paths
// are next applied. Callers should reapply the paths after changing it.
func (s *Service) SetRetentionDays(days int) {
	if days <= 0 {
		days = defaultRetentionDays
	}
	s.mu.Lock()
	s.cfg.RetentionDays = days
	s.mu.Unlock()
}

// TokenStore exposes the in-memory token store (auth hook / tests).
func (s *Service) TokenStore() *TokenStore {
	return s.tokens
}

// RecordingsDir returns the current recordings root.
func (s *Service) RecordingsDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.RecordingsDir
}

// RecordingEnabled reports whether continuous recording is on.
func (s *Service) RecordingEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.recordingEnabled && s.cfg.RecordingsDir != ""
}

// SetRecordingConfig updates the recordings directory and enable flag.
// When enabled is true, dir must be non-empty; the directory is created if needed.
func (s *Service) SetRecordingConfig(dir string, enabled bool) error {
	dir = strings.TrimSpace(dir)
	if enabled && dir == "" {
		return fmt.Errorf("recordings directory is required when recording is enabled")
	}
	if dir != "" {
		clean := filepath.Clean(dir)
		if clean == "." || clean == "" {
			return fmt.Errorf("invalid recordings directory")
		}
		if err := os.MkdirAll(clean, 0o755); err != nil {
			return fmt.Errorf("create recordings directory: %w", err)
		}
		dir = clean
	}

	s.mu.Lock()
	s.cfg.RecordingsDir = dir
	s.recordingEnabled = enabled && dir != ""
	s.mu.Unlock()
	return nil
}

// recordOptionsForCamera builds MediaMTX recording settings under RecordingsDir.
// MediaMTX requires the literal "%path" in recordPath (expands to the path name, e.g. cam_<uuid>).
// Layout: {RecordingsDir}/%path/%Y-%m-%d/%H-%M-%S-%f (MediaMTX adds the extension).
func (s *Service) recordOptionsForCamera(cameraID string) PathRecordOptions {
	s.mu.RLock()
	dir := s.cfg.RecordingsDir
	enabled := s.recordingEnabled
	retentionDays := s.cfg.RetentionDays
	s.mu.RUnlock()

	if !enabled || dir == "" {
		return PathRecordOptions{Enabled: false}
	}
	_ = cameraID // path name already encodes camera id (cam_<uuid>); kept for call-site clarity
	root := filepath.ToSlash(dir)
	return PathRecordOptions{
		Enabled:         true,
		RecordPath:      root + "/%path/%Y-%m-%d/%H-%M-%S-%f",
		SegmentDuration: "1m",
		Format:          "fmp4",
		DeleteAfter:     strconv.Itoa(retentionDays) + "d",
	}
}

// ReapplyCameraPaths re-upserts MediaMTX paths so recording settings take effect.
func (s *Service) ReapplyCameraPaths(ctx context.Context, cameras []camera.Camera) {
	for _, cam := range cameras {
		if ctx.Err() != nil {
			return
		}
		if !cam.Enabled {
			continue
		}
		if err := s.EnsurePathForCamera(ctx, cam); err != nil {
			slog.Warn("reapply mediamtx path failed",
				"camera_id", cam.ID, "err", err)
		}
	}
}

// EnsurePathForCamera upserts a MediaMTX on-demand path for the camera's live RTSP
// with continuous recording enabled when RecordingsDir is configured.
// MediaMTX being down is logged and returned as error; callers may continue for live URL minting.
func (s *Service) EnsurePathForCamera(ctx context.Context, cam camera.Camera) error {
	source, err := s.cams.LiveSourceURL(ctx, cam.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoLiveSource, err)
	}
	name := PathName(cam.ID)
	rec := s.recordOptionsForCamera(cam.ID)
	if err := s.mtx.UpsertPath(ctx, name, source, rec); err != nil {
		return err
	}
	return nil
}

// DeletePath removes the MediaMTX path for a camera. Missing paths are success.
func (s *Service) DeletePath(ctx context.Context, cameraID string) error {
	return s.mtx.DeletePath(ctx, PathName(cameraID))
}

// IssueLive ensures the MediaMTX path (best-effort), mints a stream token, and returns live URLs.
// Camera must be enabled (ErrCameraDisabled). MediaMTX down only logs a warning.
func (s *Service) IssueLive(ctx context.Context, cameraID string) (LiveStream, error) {
	cam, err := s.cams.Get(ctx, cameraID)
	if err != nil {
		return LiveStream{}, err
	}
	if !cam.Enabled {
		return LiveStream{}, ErrCameraDisabled
	}

	path := PathName(cam.ID)
	if err := s.EnsurePathForCamera(ctx, cam); err != nil {
		// Still mint URLs so clients get a deterministic response when MTX is down.
		slog.Warn("mediamtx ensure path failed; returning live urls anyway",
			"camera_id", cam.ID, "path", path, "err", err)
	}

	token, expires, err := s.tokens.MintToken(cam.ID, path, DefaultTokenTTL)
	if err != nil {
		return LiveStream{}, fmt.Errorf("mint stream token: %w", err)
	}

	whep := joinURL(s.cfg.WebRTCURL, path+"/whep") + "?token=" + url.QueryEscape(token)
	hls := joinURL(s.cfg.HLSURL, path+"/index.m3u8") + "?token=" + url.QueryEscape(token)

	return LiveStream{
		CameraID:  cam.ID,
		WHEPURL:   whep,
		HLSURL:    hls,
		Token:     token,
		ExpiresAt: expires,
	}, nil
}

// IssuePlayback mints a stream token and returns a playback URL for recorded video.
// Prefers MediaMTX playback server when PlaybackURL is configured; otherwise tokenized HLS with start query.
func (s *Service) IssuePlayback(ctx context.Context, cameraID string, start time.Time, durationSec float64) (PlaybackStream, error) {
	cam, err := s.cams.Get(ctx, cameraID)
	if err != nil {
		return PlaybackStream{}, err
	}

	path := PathName(cam.ID)
	token, expires, err := s.tokens.MintToken(cam.ID, path, DefaultTokenTTL)
	if err != nil {
		return PlaybackStream{}, fmt.Errorf("mint stream token: %w", err)
	}

	if durationSec <= 0 {
		durationSec = 60
	}
	startUTC := start.UTC()
	startParam := startUTC.Format(time.RFC3339)

	var playback string
	if s.cfg.PlaybackURL != "" {
		// MediaMTX playback API style: /get?path=&start=&duration=
		q := url.Values{}
		q.Set("path", path)
		q.Set("start", startParam)
		q.Set("duration", formatDurationSec(durationSec))
		q.Set("format", "mp4")
		q.Set("token", token)
		playback = s.cfg.PlaybackURL + "/get?" + q.Encode()
	} else {
		// Placeholder: tokenized HLS with start/duration query hints for the player.
		q := url.Values{}
		q.Set("token", token)
		q.Set("start", startParam)
		q.Set("duration", formatDurationSec(durationSec))
		playback = joinURL(s.cfg.HLSURL, path+"/index.m3u8") + "?" + q.Encode()
	}

	return PlaybackStream{
		CameraID:    cam.ID,
		PlaybackURL: playback,
		Token:       token,
		ExpiresAt:   expires,
	}, nil
}

// IssueArchivedPlayback returns a tokenized, same-preview URL for a recording
// whose local file is no longer available. The API content handler validates
// the token before proxying byte ranges from the archive provider.
func (s *Service) IssueArchivedPlayback(ctx context.Context, cameraID, recordingID string) (PlaybackStream, error) {
	cam, err := s.cams.Get(ctx, cameraID)
	if err != nil {
		return PlaybackStream{}, err
	}
	recordingID = strings.TrimSpace(recordingID)
	if recordingID == "" {
		return PlaybackStream{}, fmt.Errorf("recording id is required")
	}

	tokenPath := archivedPlaybackTokenPath(recordingID)
	token, expires, err := s.tokens.MintToken(cam.ID, tokenPath, ArchivedPlaybackTokenTTL)
	if err != nil {
		return PlaybackStream{}, fmt.Errorf("mint archived playback token: %w", err)
	}
	q := url.Values{}
	q.Set("token", token)
	playback := "/api/v1/recordings/" + url.PathEscape(recordingID) + "/content?" + q.Encode()
	return PlaybackStream{
		CameraID:    cam.ID,
		PlaybackURL: playback,
		Token:       token,
		ExpiresAt:   expires,
	}, nil
}

// ValidateArchivedPlayback authorizes a Drive content request for one recording.
func (s *Service) ValidateArchivedPlayback(token, recordingID string) bool {
	return s.tokens.ValidateAndConsume(strings.TrimSpace(token), archivedPlaybackTokenPath(recordingID))
}

func archivedPlaybackTokenPath(recordingID string) string {
	return "recording:" + strings.TrimSpace(recordingID)
}

// Snapshot captures a JPEG frame from the camera's live RTSP via ffmpeg.
func (s *Service) Snapshot(ctx context.Context, cameraID string) ([]byte, error) {
	cam, err := s.cams.Get(ctx, cameraID)
	if err != nil {
		return nil, err
	}
	if !cam.Enabled {
		return nil, ErrCameraDisabled
	}
	source, err := s.cams.LiveSourceURL(ctx, cameraID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoLiveSource, err)
	}
	jpeg, err := captureSnapshot(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSnapshotFailed, err)
	}
	return jpeg, nil
}

// ValidateAuth validates a MediaMTX external-auth request.
// Accepts token from password, token field, or query string token=.
// Empty credentials return false (caller should 401 so clients send creds).
func (s *Service) ValidateAuth(req AuthRequest) bool {
	token := strings.TrimSpace(req.Password)
	if token == "" {
		token = strings.TrimSpace(req.Token)
	}
	if token == "" {
		token = tokenFromQuery(req.Query)
	}
	if token == "" {
		return false
	}

	path := strings.TrimSpace(req.Path)
	// MediaMTX may include leading slash or subpaths; normalize to path name only.
	path = strings.Trim(path, "/")
	if i := strings.IndexByte(path, '/'); i >= 0 {
		path = path[:i]
	}

	// Allow API/metrics/pprof without stream token (caller can gate actions if needed).
	// Validate read/publish/playback against stream tokens.
	switch req.Action {
	case "", "read", "publish", "playback":
		return s.tokens.ValidateAndConsume(token, path)
	default:
		// Other actions (api, metrics, pprof) are not gated by stream tokens here.
		return false
	}
}

func tokenFromQuery(q string) string {
	q = strings.TrimPrefix(strings.TrimSpace(q), "?")
	if q == "" {
		return ""
	}
	vals, err := url.ParseQuery(q)
	if err != nil {
		// Fallback: manual scan for token=
		for _, part := range strings.Split(q, "&") {
			if strings.HasPrefix(part, "token=") {
				v, _ := url.QueryUnescape(strings.TrimPrefix(part, "token="))
				return v
			}
		}
		return ""
	}
	return vals.Get("token")
}

func joinURL(base, path string) string {
	base = strings.TrimRight(base, "/")
	path = strings.TrimLeft(path, "/")
	if base == "" {
		return "/" + path
	}
	return base + "/" + path
}

func formatDurationSec(sec float64) string {
	// Prefer integer seconds when whole; otherwise fixed float.
	if sec == float64(int64(sec)) {
		return strconv.FormatInt(int64(sec), 10)
	}
	return strconv.FormatFloat(sec, 'f', 3, 64)
}
