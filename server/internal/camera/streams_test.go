package camera

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiscoverStreamsUsesCredentialsAndReturnsRTSPURLs(t *testing.T) {
	var requests []string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		requests = append(requests, string(body))
		w.Header().Set("Content-Type", "application/soap+xml")
		switch {
		case strings.Contains(string(body), "GetCapabilities"):
			_, _ = io.WriteString(w, `<Envelope><Body><GetCapabilitiesResponse><Capabilities><Media><XAddr>`+server.URL+`/onvif/media</XAddr></Media></Capabilities></GetCapabilitiesResponse></Body></Envelope>`)
		case strings.Contains(string(body), "GetProfiles"):
			_, _ = io.WriteString(w, `<Envelope><Body><GetProfilesResponse><Profiles token="profile-main"><Name>Main</Name></Profiles></GetProfilesResponse></Body></Envelope>`)
		case strings.Contains(string(body), "GetStreamUri"):
			_, _ = io.WriteString(w, `<Envelope><Body><GetStreamUriResponse><MediaUri><Uri>rtsp://192.168.1.50:554/stream/main</Uri></MediaUri></GetStreamUriResponse></Body></Envelope>`)
		default:
			t.Fatalf("unexpected ONVIF request: %s", body)
		}
	}))
	defer server.Close()

	result, err := (&Service{}).DiscoverStreams(context.Background(), StreamDiscoveryInput{
		XAddr:    server.URL,
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("DiscoverStreams: %v", err)
	}
	if result.LiveRTSPURL != "rtsp://192.168.1.50:554/stream/main" {
		t.Fatalf("live URL: %q", result.LiveRTSPURL)
	}
	if result.RecordRTSPURL != result.LiveRTSPURL {
		t.Fatalf("record URL: %q", result.RecordRTSPURL)
	}
	if len(requests) != 3 {
		t.Fatalf("expected capabilities, profiles, and stream requests; got %d", len(requests))
	}
	if !strings.Contains(requests[0], "admin") {
		t.Fatal("expected username in authenticated ONVIF request")
	}
}

func TestOnvifDeviceHost(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		want string
	}{
		{name: "URL", in: "http://192.168.1.50:8080/onvif/device_service", want: "192.168.1.50:8080"},
		{name: "host", in: "192.168.1.50:8080", want: "192.168.1.50:8080"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := onvifDeviceHost(test.in)
			if err != nil {
				t.Fatalf("onvifDeviceHost: %v", err)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}
