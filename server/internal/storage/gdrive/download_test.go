package gdrive

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadForwardsByteRange(t *testing.T) {
	var gotRange string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRange = r.Header.Get("Range")
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Range", "bytes 100-199/1000")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "recording-bytes")
	}))
	defer upstream.Close()

	service := newTestService(t)
	service.cfg.APIEndpoint = upstream.URL + "/"
	resp, err := service.Download(context.Background(), "drive-file-1", "bytes=100-199")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if gotRange != "bytes=100-199" {
		t.Fatalf("upstream Range = %q", gotRange)
	}
	if resp.StatusCode != http.StatusPartialContent || string(body) != "recording-bytes" {
		t.Fatalf("response status=%d body=%q", resp.StatusCode, body)
	}
}

func TestDownloadRejectsMultipleRanges(t *testing.T) {
	service := newTestService(t)
	if _, err := service.Download(context.Background(), "drive-file-1", "bytes=0-1,4-5"); err == nil {
		t.Fatal("expected multiple range rejection")
	}
}
