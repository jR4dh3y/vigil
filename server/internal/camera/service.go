package camera

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/store"
)

// ErrNotFound is returned when a camera does not exist.
var ErrNotFound = errors.New("camera not found")

// Service provides camera CRUD and probing.
type Service struct {
	db         *sql.DB
	q          *store.Queries
	secretsKey string
	driver     Driver
}

// NewService constructs a camera Service backed by db.
func NewService(db *sql.DB, secretsKey string) *Service {
	return &Service{
		db:         db,
		q:          store.New(db),
		secretsKey: secretsKey,
		driver:     NewGenericRTSPDriver(),
	}
}

// SetDriver overrides the default probe driver (tests).
func (s *Service) SetDriver(d Driver) {
	s.driver = d
}

// List returns all cameras with stream profiles.
func (s *Service) List(ctx context.Context) ([]Camera, error) {
	rows, err := s.q.ListCameras(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cameras: %w", err)
	}
	if len(rows) == 0 {
		return []Camera{}, nil
	}

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	profiles, err := s.q.ListStreamProfilesByCameraIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list stream profiles: %w", err)
	}
	byCam := groupProfiles(profiles)

	out := make([]Camera, 0, len(rows))
	for _, r := range rows {
		c, err := toDomain(r, byCam[r.ID])
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Get returns a single camera by id.
func (s *Service) Get(ctx context.Context, id string) (Camera, error) {
	row, err := s.q.GetCamera(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Camera{}, ErrNotFound
		}
		return Camera{}, fmt.Errorf("get camera: %w", err)
	}
	profiles, err := s.q.ListStreamProfilesByCamera(ctx, id)
	if err != nil {
		return Camera{}, fmt.Errorf("list stream profiles: %w", err)
	}
	return toDomain(row, profiles)
}

// Create inserts a camera and optional stream profiles.
func (s *Service) Create(ctx context.Context, in CreateInput) (Camera, error) {
	name := strings.TrimSpace(in.Name)
	host := strings.TrimSpace(in.Host)
	if name == "" {
		return Camera{}, fmt.Errorf("name is required")
	}
	if host == "" {
		return Camera{}, fmt.Errorf("host is required")
	}

	driver := strings.TrimSpace(in.Driver)
	if driver == "" {
		driver = DefaultDriver
	}

	enc, err := EncryptCredential(s.secretsKey, in.Password)
	if err != nil {
		return Camera{}, fmt.Errorf("encrypt password: %w", err)
	}

	enabled := int64(0)
	if in.Enabled {
		enabled = 1
	}

	id := uuid.NewString()
	liveURL, recordURL := resolveStreamURLs(host, in.LiveRTSPURL, in.RecordRTSPURL)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Camera{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := s.q.WithTx(tx)
	if _, err := q.CreateCamera(ctx, store.CreateCameraParams{
		ID:          id,
		Name:        name,
		Driver:      driver,
		Host:        host,
		Username:    in.Username,
		PasswordEnc: enc,
		Enabled:     enabled,
		Status:      StatusUnknown,
	}); err != nil {
		return Camera{}, fmt.Errorf("create camera: %w", err)
	}

	if liveURL != "" {
		if _, err := q.UpsertStreamProfile(ctx, store.UpsertStreamProfileParams{
			ID:       uuid.NewString(),
			CameraID: id,
			Role:     RoleLive,
			RtspUrl:  liveURL,
		}); err != nil {
			return Camera{}, fmt.Errorf("create live profile: %w", err)
		}
	}
	if recordURL != "" {
		if _, err := q.UpsertStreamProfile(ctx, store.UpsertStreamProfileParams{
			ID:       uuid.NewString(),
			CameraID: id,
			Role:     RoleRecord,
			RtspUrl:  recordURL,
		}); err != nil {
			return Camera{}, fmt.Errorf("create record profile: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Camera{}, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, id)
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (Camera, error) {
	existing, err := s.q.GetCamera(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Camera{}, ErrNotFound
		}
		return Camera{}, fmt.Errorf("get camera: %w", err)
	}

	name := existing.Name
	driver := existing.Driver
	host := existing.Host
	username := existing.Username
	passwordEnc := existing.PasswordEnc
	enabled := existing.Enabled
	status := existing.Status

	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return Camera{}, fmt.Errorf("name cannot be empty")
		}
		name = n
	}
	if in.Driver != nil {
		d := strings.TrimSpace(*in.Driver)
		if d == "" {
			d = DefaultDriver
		}
		driver = d
	}
	if in.Host != nil {
		host = strings.TrimSpace(*in.Host)
	}
	if in.Username != nil {
		username = *in.Username
	}
	if in.Password != nil {
		enc, err := EncryptCredential(s.secretsKey, *in.Password)
		if err != nil {
			return Camera{}, fmt.Errorf("encrypt password: %w", err)
		}
		passwordEnc = enc
	}
	if in.Enabled != nil {
		if *in.Enabled {
			enabled = 1
		} else {
			enabled = 0
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Camera{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := s.q.WithTx(tx)
	if _, err := q.UpdateCamera(ctx, store.UpdateCameraParams{
		Name:        name,
		Driver:      driver,
		Host:        host,
		Username:    username,
		PasswordEnc: passwordEnc,
		Enabled:     enabled,
		Status:      status,
		ID:          id,
	}); err != nil {
		return Camera{}, fmt.Errorf("update camera: %w", err)
	}

	if in.LiveRTSPURL != nil {
		url := strings.TrimSpace(*in.LiveRTSPURL)
		if url != "" {
			if _, err := q.UpsertStreamProfile(ctx, store.UpsertStreamProfileParams{
				ID:       uuid.NewString(),
				CameraID: id,
				Role:     RoleLive,
				RtspUrl:  url,
			}); err != nil {
				return Camera{}, fmt.Errorf("upsert live profile: %w", err)
			}
		}
	}
	if in.RecordRTSPURL != nil {
		url := strings.TrimSpace(*in.RecordRTSPURL)
		if url != "" {
			if _, err := q.UpsertStreamProfile(ctx, store.UpsertStreamProfileParams{
				ID:       uuid.NewString(),
				CameraID: id,
				Role:     RoleRecord,
				RtspUrl:  url,
			}); err != nil {
				return Camera{}, fmt.Errorf("upsert record profile: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Camera{}, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, id)
}

// Delete removes a camera (stream profiles cascade).
func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.q.GetCamera(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get camera: %w", err)
	}
	if err := s.q.DeleteCamera(ctx, id); err != nil {
		return fmt.Errorf("delete camera: %w", err)
	}
	return nil
}

// SetStatus updates the camera online/offline/unknown status.
func (s *Service) SetStatus(ctx context.Context, id, status string) error {
	switch status {
	case StatusOnline, StatusOffline, StatusUnknown:
	default:
		return fmt.Errorf("invalid camera status %q", status)
	}
	_, err := s.q.GetCamera(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get camera: %w", err)
	}
	if err := s.q.UpdateCameraStatus(ctx, store.UpdateCameraStatusParams{
		Status: status,
		ID:     id,
	}); err != nil {
		return fmt.Errorf("update camera status: %w", err)
	}
	return nil
}

// Probe probes an RTSP URL using the configured driver.
func (s *Service) Probe(ctx context.Context, rtspURL, username, password string) (ProbeResult, error) {
	url := strings.TrimSpace(rtspURL)
	if url == "" {
		return ProbeResult{}, fmt.Errorf("rtspUrl is required")
	}
	return s.driver.Probe(ctx, url, username, password)
}

// LiveSourceURL returns the live-role RTSP URL with credentials injected for MediaMTX/FFmpeg.
// Prefers role=live; falls back to record if no live profile exists.
func (s *Service) LiveSourceURL(ctx context.Context, id string) (string, error) {
	row, err := s.q.GetCamera(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get camera: %w", err)
	}
	profiles, err := s.q.ListStreamProfilesByCamera(ctx, id)
	if err != nil {
		return "", fmt.Errorf("list stream profiles: %w", err)
	}

	var liveURL, anyURL string
	for _, p := range profiles {
		u := strings.TrimSpace(p.RtspUrl)
		if u == "" {
			continue
		}
		if anyURL == "" {
			anyURL = u
		}
		if p.Role == RoleLive {
			liveURL = u
			break
		}
	}
	url := liveURL
	if url == "" {
		url = anyURL
	}
	if url == "" {
		return "", fmt.Errorf("no RTSP stream profile for camera")
	}

	pass, err := DecryptCredential(s.secretsKey, row.PasswordEnc)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %w", err)
	}
	return injectRTSPCredentials(url, row.Username, pass), nil
}

// resolveStreamURLs decides live/record profile URLs from create input.
// Explicit URLs win; otherwise if host is an rtsp:// URL it is used for both.
func resolveStreamURLs(host, live, record string) (liveURL, recordURL string) {
	live = strings.TrimSpace(live)
	record = strings.TrimSpace(record)
	if live != "" || record != "" {
		return live, record
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(host)), "rtsp://") {
		u := strings.TrimSpace(host)
		return u, u
	}
	return "", ""
}

func groupProfiles(profiles []store.StreamProfile) map[string][]store.StreamProfile {
	m := make(map[string][]store.StreamProfile)
	for _, p := range profiles {
		m[p.CameraID] = append(m[p.CameraID], p)
	}
	return m
}

func toDomain(row store.Camera, profiles []store.StreamProfile) (Camera, error) {
	created, err := auth.ParseSQLiteTime(row.CreatedAt)
	if err != nil {
		// sqlite datetime may already be fine; fall back to now-ish zero with parse attempt
		created = time.Time{}
		if t, e2 := time.Parse(time.RFC3339, row.CreatedAt); e2 == nil {
			created = t
		}
	}
	updated, err := auth.ParseSQLiteTime(row.UpdatedAt)
	if err != nil {
		updated = created
		if t, e2 := time.Parse(time.RFC3339, row.UpdatedAt); e2 == nil {
			updated = t
		}
	}

	out := Camera{
		ID:             row.ID,
		Name:           row.Name,
		Driver:         row.Driver,
		Host:           row.Host,
		Username:       row.Username,
		Enabled:        row.Enabled != 0,
		Status:         row.Status,
		StreamProfiles: make([]StreamProfile, 0, len(profiles)),
		CreatedAt:      created,
		UpdatedAt:      updated,
	}
	for _, p := range profiles {
		out.StreamProfiles = append(out.StreamProfiles, toDomainProfile(p))
	}
	return out, nil
}

func toDomainProfile(p store.StreamProfile) StreamProfile {
	sp := StreamProfile{
		ID:      p.ID,
		Role:    p.Role,
		RTSPURL: p.RtspUrl,
	}
	if p.Codec.Valid {
		c := p.Codec.String
		sp.Codec = &c
	}
	if p.Width.Valid {
		w := int(p.Width.Int64)
		sp.Width = &w
	}
	if p.Height.Valid {
		h := int(p.Height.Int64)
		sp.Height = &h
	}
	return sp
}
