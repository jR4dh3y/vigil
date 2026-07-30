package gdrive

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

const recordingIDProperty = "nvr_recording_id"

// Upload creates a file under the NVR Archives folder with Name=objectKey
// (path-like). A stable recordingID makes retries idempotent: an earlier upload
// is reused if the process failed before marking the database row archived.
// contentType defaults to video/mp4. Returns the Drive file ID.
func (s *Service) Upload(ctx context.Context, recordingID, objectKey string, r io.Reader, size int64, contentType string) (fileID string, err error) {
	recordingID = strings.TrimSpace(recordingID)
	if recordingID == "" {
		return "", fmt.Errorf("recording id is required")
	}
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}
	if r == nil {
		return "", fmt.Errorf("reader is required")
	}
	if contentType == "" {
		contentType = "video/mp4"
	}

	folderID, err := s.EnsureArchivesFolder(ctx)
	if err != nil {
		return "", err
	}
	svc, err := s.driveService(ctx)
	if err != nil {
		return "", err
	}

	existingID, err := findUploadedRecording(ctx, svc, folderID, recordingID)
	if err != nil {
		return "", err
	}
	if existingID != "" {
		return existingID, nil
	}

	meta := &drive.File{
		Name:          objectKey,
		Parents:       []string{folderID},
		AppProperties: map[string]string{recordingIDProperty: recordingID},
	}
	opts := []googleapi.MediaOption{
		googleapi.ContentType(contentType),
		googleapi.ChunkSize(googleapi.DefaultUploadChunkSize),
	}
	// size is reserved for future ContentLength hints; Media reads from r.
	_ = size

	created, err := svc.Files.Create(meta).
		Context(ctx).
		Media(r, opts...).
		Fields("id").
		Do()
	if err != nil {
		return "", fmt.Errorf("drive upload %q: %w", objectKey, err)
	}
	if created == nil || created.Id == "" {
		return "", fmt.Errorf("drive upload %q: empty file id", objectKey)
	}
	return created.Id, nil
}

func findUploadedRecording(ctx context.Context, svc *drive.Service, folderID, recordingID string) (string, error) {
	q := fmt.Sprintf(
		"'%s' in parents and trashed = false and appProperties has { key='%s' and value='%s' }",
		escapeDriveQuery(folderID),
		recordingIDProperty,
		escapeDriveQuery(recordingID),
	)
	list, err := svc.Files.List().
		Context(ctx).
		Q(q).
		Spaces("drive").
		Fields("files(id)").
		PageSize(1).
		Do()
	if err != nil {
		return "", fmt.Errorf("find existing Drive archive: %w", err)
	}
	if len(list.Files) == 0 {
		return "", nil
	}
	return list.Files[0].Id, nil
}

func escapeDriveQuery(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `'`, `\'`)
}
