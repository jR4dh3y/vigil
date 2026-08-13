package media

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpsertPathSourceDemandFollowsRecording(t *testing.T) {
	tests := []struct {
		name         string
		record       PathRecordOptions
		onDemand     bool
		startTimeout string
	}{
		{
			name: "continuous recording keeps source connected",
			record: PathRecordOptions{
				Enabled:     true,
				RecordPath:  "/recordings/%path/%Y-%m-%d/%H-%M-%S-%f",
				DeleteAfter: "1d",
			},
			onDemand: false,
		},
		{
			name:         "non-recording source remains on demand",
			record:       PathRecordOptions{Enabled: false},
			onDemand:     true,
			startTimeout: "20s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got pathConf
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewMediaMTXClient(server.URL)
			if err := client.UpsertPath(context.Background(), "cam_test", "rtsp://camera/live", tt.record); err != nil {
				t.Fatalf("UpsertPath: %v", err)
			}

			if got.SourceOnDemand != tt.onDemand {
				t.Fatalf("SourceOnDemand = %v, want %v", got.SourceOnDemand, tt.onDemand)
			}
			if got.SourceOnDemandStartTimeout != tt.startTimeout {
				t.Fatalf("SourceOnDemandStartTimeout = %q, want %q", got.SourceOnDemandStartTimeout, tt.startTimeout)
			}
			if tt.record.Enabled && got.RecordDeleteAfter != tt.record.DeleteAfter {
				t.Fatalf("RecordDeleteAfter = %q, want %q", got.RecordDeleteAfter, tt.record.DeleteAfter)
			}
		})
	}
}
