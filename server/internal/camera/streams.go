package camera

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	onvifroot "github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
	onviftypes "github.com/use-go/onvif/xsd/onvif"
)

const onvifRequestTimeout = 8 * time.Second

type getProfilesEnvelope struct {
	Body struct {
		Response media.GetProfilesResponse `xml:"GetProfilesResponse"`
	} `xml:"Body"`
}

type getStreamURIEnvelope struct {
	Body struct {
		Response media.GetStreamUriResponse `xml:"GetStreamUriResponse"`
	} `xml:"Body"`
}

// DiscoverStreams authenticates to an ONVIF device and asks it for RTSP stream
// URLs. The credentials are used only for this request and are never stored.
func (s *Service) DiscoverStreams(ctx context.Context, in StreamDiscoveryInput) (StreamDiscoveryResult, error) {
	host, err := onvifDeviceHost(in.XAddr)
	if err != nil {
		return StreamDiscoveryResult{}, err
	}

	device, err := onvifroot.NewDevice(onvifroot.DeviceParams{
		Xaddr:    host,
		Username: strings.TrimSpace(in.Username),
		Password: in.Password,
		HttpClient: &http.Client{
			Timeout: onvifRequestTimeout,
		},
	})
	if err != nil {
		return StreamDiscoveryResult{}, fmt.Errorf("connect to ONVIF device: %w", err)
	}

	profiles, err := getONVIFProfiles(device)
	if err != nil {
		return StreamDiscoveryResult{}, err
	}

	streams := make([]string, 0, len(profiles))
	seen := make(map[string]struct{})
	for _, profile := range profiles {
		stream, err := getONVIFStreamURI(device, profile.Token)
		if err != nil {
			continue
		}
		stream = strings.TrimSpace(stream)
		if stream == "" {
			continue
		}
		stream = stripRTSPCredentials(stream)
		if _, exists := seen[stream]; exists {
			continue
		}
		seen[stream] = struct{}{}
		streams = append(streams, stream)
	}
	if len(streams) == 0 {
		return StreamDiscoveryResult{}, fmt.Errorf("ONVIF device did not return an RTSP stream")
	}

	result := StreamDiscoveryResult{
		LiveRTSPURL:   streams[0],
		RecordRTSPURL: streams[0],
	}
	if len(streams) > 1 {
		result.RecordRTSPURL = streams[1]
	}
	return result, nil
}

func stripRTSPCredentials(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	parsed.User = nil
	return parsed.String()
}

func getONVIFProfiles(device *onvifroot.Device) ([]onviftypes.Profile, error) {
	response, err := device.CallMethod(media.GetProfiles{})
	if err != nil {
		return nil, fmt.Errorf("request ONVIF profiles: %w", err)
	}

	var envelope getProfilesEnvelope
	if err := decodeONVIFResponse(response, &envelope); err != nil {
		return nil, fmt.Errorf("decode ONVIF profiles: %w", err)
	}
	if len(envelope.Body.Response.Profiles) == 0 {
		return nil, fmt.Errorf("ONVIF device returned no media profiles")
	}
	return envelope.Body.Response.Profiles, nil
}

func getONVIFStreamURI(device *onvifroot.Device, token onviftypes.ReferenceToken) (string, error) {
	response, err := device.CallMethod(media.GetStreamUri{
		StreamSetup: onviftypes.StreamSetup{
			Stream: onviftypes.StreamType("RTP-Unicast"),
			Transport: onviftypes.Transport{
				Protocol: onviftypes.TransportProtocol("RTSP"),
			},
		},
		ProfileToken: token,
	})
	if err != nil {
		return "", fmt.Errorf("request ONVIF stream URI: %w", err)
	}

	var envelope getStreamURIEnvelope
	if err := decodeONVIFResponse(response, &envelope); err != nil {
		return "", fmt.Errorf("decode ONVIF stream URI: %w", err)
	}
	return string(envelope.Body.Response.MediaUri.Uri), nil
}

func decodeONVIFResponse(response *http.Response, target any) error {
	if response == nil {
		return fmt.Errorf("empty ONVIF response")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return fmt.Errorf("ONVIF returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return xml.NewDecoder(response.Body).Decode(target)
}

func onvifDeviceHost(rawXAddr string) (string, error) {
	xaddr := strings.TrimSpace(rawXAddr)
	if xaddr == "" {
		return "", fmt.Errorf("ONVIF device address is required")
	}
	parsed, err := url.Parse(xaddr)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("http://" + xaddr)
		if err != nil {
			return "", fmt.Errorf("parse ONVIF device host: %w", err)
		}
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("ONVIF device address has no host")
	}
	return parsed.Host, nil
}
