package camera

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestParseProbeResponses(t *testing.T) {
	messages := []string{
		`<?xml version="1.0"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><d:ProbeMatches xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"><d:ProbeMatch><d:EndpointReference><a:Address xmlns:a="http://www.w3.org/2005/08/addressing">urn:uuid:camera-1</a:Address></d:EndpointReference><d:XAddrs>http://192.168.1.50/onvif/device_service http://192.168.1.50/onvif/device</d:XAddrs><d:Scopes>onvif://www.onvif.org/name/Front%20Door onvif://www.onvif.org/hardware/Camera</d:Scopes></d:ProbeMatch></d:ProbeMatches></s:Body></s:Envelope>`,
		`<Envelope><Body><ProbeMatches><ProbeMatch><EndpointReference><Address>urn:uuid:camera-1</Address></EndpointReference><XAddrs>http://192.168.1.50/onvif/device_service</XAddrs><Scopes>onvif://www.onvif.org/name/Front%20Door</Scopes></ProbeMatch></ProbeMatches></Body></Envelope>`,
		`<Envelope><Body><ProbeMatches><ProbeMatch><EndpointReference><Address>urn:uuid:camera-2</Address></EndpointReference><XAddrs>http://192.168.1.20:8080/onvif/device_service</XAddrs><Scopes>onvif://www.onvif.org/hardware/OutdoorCam</Scopes></ProbeMatch></ProbeMatches></Body></Envelope>`,
	}

	got := parseProbeResponses(messages)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique cameras, got %d: %+v", len(got), got)
	}
	if got[0] != (DiscoveredCamera{
		ID:    "urn:uuid:camera-1",
		Name:  "Front Door",
		Host:  "192.168.1.50",
		XAddr: "http://192.168.1.50/onvif/device_service",
	}) {
		t.Fatalf("unexpected first camera: %+v", got[0])
	}
	if got[1] != (DiscoveredCamera{
		ID:    "urn:uuid:camera-2",
		Name:  "OutdoorCam",
		Host:  "192.168.1.20:8080",
		XAddr: "http://192.168.1.20:8080/onvif/device_service",
	}) {
		t.Fatalf("unexpected second camera: %+v", got[1])
	}

}

func TestParseProbeResponsesIgnoresMalformedMessages(t *testing.T) {
	got := parseProbeResponses([]string{"not xml", `<Envelope><Body><ProbeMatches><ProbeMatch><XAddrs>not-a-url</XAddrs></ProbeMatch></ProbeMatches></Body></Envelope>`})
	if len(got) != 0 {
		t.Fatalf("expected no cameras, got %+v", got)
	}
}

func TestAuthenticateDiscoveredCamerasUsesCredentialsAndFiltersFailures(t *testing.T) {
	candidates := []DiscoveredCamera{
		{ID: "camera-2", Name: "Back Door", Host: "192.168.1.60", XAddr: "http://192.168.1.60/onvif/device_service"},
		{ID: "camera-1", Name: "Front Door", Host: "192.168.1.50", XAddr: "http://192.168.1.50/onvif/device_service"},
	}
	credentials := DiscoveryInput{Username: "camera-admin", Password: "secret"}
	inputs := make(chan StreamDiscoveryInput, len(candidates))

	got := authenticateDiscoveredCameras(
		context.Background(),
		candidates,
		credentials,
		func(_ context.Context, in StreamDiscoveryInput) (StreamDiscoveryResult, error) {
			inputs <- in
			if in.XAddr == candidates[0].XAddr {
				return StreamDiscoveryResult{}, errors.New("invalid credentials")
			}
			return StreamDiscoveryResult{LiveRTSPURL: "rtsp://camera/live"}, nil
		},
	)

	expected := candidates[1]
	expected.LiveRTSPURL = "rtsp://camera/live"
	expected.RecordRTSPURL = "rtsp://camera/live"
	if len(got) != 1 || got[0] != expected {
		t.Fatalf("authenticated cameras: %+v", got)
	}
	for range candidates {
		in := <-inputs
		if in.Username != credentials.Username || in.Password != credentials.Password {
			t.Fatalf("credentials not forwarded: %+v", in)
		}
	}
}

func TestDiscoverDahuaRTSPBuildsCredentiallessChannelURLs(t *testing.T) {
	credentials := DiscoveryInput{Username: "admin", Password: "secret"}
	var calls []string
	var callsMu sync.Mutex
	got := discoverDahuaRTSP(
		context.Background(),
		credentials,
		[]string{"192.168.1.240"},
		func(_ context.Context, streamURL, username, password string) (ProbeResult, error) {
			callsMu.Lock()
			calls = append(calls, streamURL+"|"+username+"|"+password)
			callsMu.Unlock()
			if streamURL == "rtsp://192.168.1.240:554/cam/realmonitor?channel=1&subtype=0" ||
				streamURL == "rtsp://192.168.1.240:554/cam/realmonitor?channel=2&subtype=0" {
				return ProbeResult{Reachable: true}, nil
			}
			return ProbeResult{}, nil
		},
	)

	if len(got) != 2 {
		t.Fatalf("expected 2 discovered channels, got %d: %+v", len(got), got)
	}
	for _, camera := range got {
		if camera.LiveRTSPURL != camera.XAddr || camera.RecordRTSPURL != camera.XAddr {
			t.Fatalf("stream URLs not carried through: %+v", camera)
		}
		if strings.Contains(camera.LiveRTSPURL, "@") {
			t.Fatalf("credentials leaked into discovered URL: %q", camera.LiveRTSPURL)
		}
	}
	if len(calls) != dahuaDiscoveryChannels {
		t.Fatalf("expected one probe per channel, got %d", len(calls))
	}
}
