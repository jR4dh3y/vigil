package gdrive

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/api/drive/v3"
)

// archivesFolderName is the Drive folder that holds archived segment files.
const archivesFolderName = "NVR Archives"

// EnsureArchivesFolder returns the Drive folder ID for NVR archives, creating it if needed.
// Uses the cached gdrive.folder_id setting when still valid.
func (s *Service) EnsureArchivesFolder(ctx context.Context) (string, error) {
	svc, err := s.driveService(ctx)
	if err != nil {
		return "", err
	}

	cached, err := s.getSetting(ctx, KeyFolderID)
	if err != nil {
		return "", err
	}
	if id := strings.TrimSpace(cached); id != "" {
		f, err := svc.Files.Get(id).Context(ctx).Fields("id", "trashed", "mimeType").Do()
		if err == nil && f != nil && !f.Trashed && f.MimeType == "application/vnd.google-apps.folder" {
			return f.Id, nil
		}
		if err != nil {
			slog.Info("gdrive cached folder invalid, will recreate", "folder_id", id, "err", err)
		}
	}

	// Look for an existing top-level archives folder (Drive query uses single quotes).
	q := "name = '" + archivesFolderName + "' and mimeType = 'application/vnd.google-apps.folder' and trashed = false and 'root' in parents"
	list, err := svc.Files.List().Context(ctx).Q(q).Spaces("drive").
		Fields("files(id, name)").PageSize(1).Do()
	if err != nil {
		return "", fmt.Errorf("list archives folder: %w", err)
	}
	if len(list.Files) > 0 && list.Files[0].Id != "" {
		id := list.Files[0].Id
		if err := s.setSetting(ctx, KeyFolderID, id); err != nil {
			return "", err
		}
		return id, nil
	}

	// Create the folder under Drive root.
	created, err := svc.Files.Create(&drive.File{
		Name:     archivesFolderName,
		MimeType: "application/vnd.google-apps.folder",
	}).Context(ctx).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("create archives folder: %w", err)
	}
	if created == nil || created.Id == "" {
		return "", fmt.Errorf("create archives folder: empty id")
	}
	if err := s.setSetting(ctx, KeyFolderID, created.Id); err != nil {
		return "", err
	}
	slog.Info("gdrive archives folder created", "folder_id", created.Id)
	return created.Id, nil
}
