package gdrive

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/googleapi"
)

var (
	ErrArchiveNotFound     = errors.New("Drive archive not found")
	ErrRangeNotSatisfiable = errors.New("Drive archive byte range not satisfiable")
)

// Download opens an archived Drive file for streaming. A single HTTP byte
// range is forwarded to Drive so browser seeking does not download the whole
// recording for every request. The caller owns and must close the response.
func (s *Service) Download(ctx context.Context, fileID, byteRange string) (*http.Response, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, fmt.Errorf("drive file id is required")
	}
	callRange := strings.TrimSpace(byteRange)
	if callRange != "" && (!strings.HasPrefix(callRange, "bytes=") || strings.Contains(callRange, ",") || len(callRange) > 128) {
		return nil, fmt.Errorf("invalid byte range")
	}

	svc, err := s.driveService(ctx)
	if err != nil {
		return nil, err
	}
	call := svc.Files.Get(fileID).Context(ctx)
	if callRange != "" {
		call.Header().Set("Range", callRange)
	}
	resp, err := call.Download()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) {
			switch apiErr.Code {
			case http.StatusNotFound:
				return nil, fmt.Errorf("%w: %v", ErrArchiveNotFound, err)
			case http.StatusRequestedRangeNotSatisfiable:
				return nil, fmt.Errorf("%w: %v", ErrRangeNotSatisfiable, err)
			}
		}
		return nil, fmt.Errorf("download Drive archive: %w", err)
	}
	return resp, nil
}
