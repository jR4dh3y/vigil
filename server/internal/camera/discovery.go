package camera

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"

	wsdiscovery "github.com/use-go/onvif/ws-discovery"
)

const (
	onvifDiscoveryType      = "dn:NetworkVideoTransmitter"
	onvifDiscoveryNamespace = "http://www.onvif.org/ver10/network/wsdl"
	onvifDiscoveryTimeout   = 4 * time.Second
)

type probeEnvelope struct {
	Body struct {
		ProbeMatches struct {
			Matches []probeMatch `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}

type probeMatch struct {
	EndpointReference struct {
		Address string `xml:"Address"`
	} `xml:"EndpointReference"`
	XAddrs string `xml:"XAddrs"`
	Scopes string `xml:"Scopes"`
}

// Discover returns ONVIF cameras visible on the NVR host's local network.
// Discovery does not contact cameras with credentials; credentials are only
// requested later when the user configures the selected camera.
func (s *Service) Discover(ctx context.Context) ([]DiscoveredCamera, error) {
	return discoverONVIF(ctx)
}

func discoverONVIF(ctx context.Context) ([]DiscoveredCamera, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}

	var names []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagMulticast == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		names = append(names, iface.Name)
	}
	if len(names) == 0 {
		return []DiscoveredCamera{}, nil
	}
	sort.Strings(names)

	scanCtx, cancel := context.WithTimeout(ctx, onvifDiscoveryTimeout)
	defer cancel()

	responses := make(chan []string, len(names))
	for _, name := range names {
		go func(interfaceName string) {
			messages, err := wsdiscovery.SendProbe(
				interfaceName,
				nil,
				[]string{onvifDiscoveryType},
				map[string]string{"dn": onvifDiscoveryNamespace},
			)
			if err != nil {
				slog.Warn("ONVIF discovery probe failed", "interface", interfaceName, "err", err)
			}
			select {
			case responses <- messages:
			case <-scanCtx.Done():
			}
		}(name)
	}

	var messages []string
	remaining := len(names)
	for remaining > 0 {
		select {
		case batch := <-responses:
			messages = append(messages, batch...)
			remaining--
		case <-scanCtx.Done():
			remaining = 0
		}
	}

	return parseProbeResponses(messages), nil
}

func parseProbeResponses(messages []string) []DiscoveredCamera {
	byID := make(map[string]DiscoveredCamera)
	for _, message := range messages {
		var envelope probeEnvelope
		if err := xml.Unmarshal([]byte(message), &envelope); err != nil {
			continue
		}
		for _, match := range envelope.Body.ProbeMatches.Matches {
			for _, rawXAddr := range strings.Fields(match.XAddrs) {
				xaddr := strings.TrimSpace(rawXAddr)
				host := discoveryHost(xaddr)
				if host == "" {
					continue
				}

				id := strings.TrimSpace(match.EndpointReference.Address)
				if id == "" {
					id = strings.ToLower(host)
				}
				if _, exists := byID[id]; exists {
					continue
				}

				name := discoveryScope(match.Scopes, "name")
				if name == "" {
					name = discoveryScope(match.Scopes, "hardware")
				}
				if name == "" {
					name = host
				}
				byID[id] = DiscoveredCamera{ID: id, Name: name, Host: host, XAddr: xaddr}
				break
			}
		}
	}

	out := make([]DiscoveredCamera, 0, len(byID))
	for _, discovered := range byID {
		out = append(out, discovered)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].Host < out[j].Host
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func discoveryHost(rawXAddr string) string {
	xaddr := strings.TrimSpace(rawXAddr)
	parsed, err := url.Parse(xaddr)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return parsed.Host
}

func discoveryScope(scopes, scopeName string) string {
	marker := "onvif://www.onvif.org/" + scopeName + "/"
	start := strings.Index(scopes, marker)
	if start < 0 {
		return ""
	}
	value := scopes[start+len(marker):]
	if end := strings.Index(value, " onvif://"); end >= 0 {
		value = value[:end]
	}
	value = strings.TrimSpace(value)
	if decoded, err := url.PathUnescape(value); err == nil {
		value = decoded
	}
	return value
}
