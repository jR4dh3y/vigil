package gdrive

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// driveService builds an authenticated Drive API client using the stored OAuth tokens.
func (s *Service) driveService(ctx context.Context) (*drive.Service, error) {
	ts, err := s.tokenSource(ctx)
	if err != nil {
		return nil, err
	}
	client := oauth2.NewClient(ctx, ts)
	options := []option.ClientOption{option.WithHTTPClient(client)}
	if s.cfg.APIEndpoint != "" {
		options = append(options, option.WithEndpoint(s.cfg.APIEndpoint))
	}
	svc, err := drive.NewService(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("drive client: %w", err)
	}
	return svc, nil
}
