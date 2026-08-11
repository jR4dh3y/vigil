package camera

import "testing"

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
