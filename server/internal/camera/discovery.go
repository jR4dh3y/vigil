package camera

import (
	"context"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wsdiscovery "github.com/use-go/onvif/ws-discovery"
)

const (
	onvifDiscoveryType      = "dn:NetworkVideoTransmitter"
	onvifDiscoveryNamespace = "http://www.onvif.org/ver10/network/wsdl"
	onvifDiscoveryTimeout   = 4 * time.Second
	dahuaRTSPPort           = "554"
	rtspPortProbeTimeout    = 250 * time.Millisecond
	rtspChannelProbeTimeout = 4 * time.Second
	rtspDiscoveryMaxHosts   = 1024
	dahuaDiscoveryChannels  = 32
	rtspHostWorkers         = 32
	rtspDiscoveryWorkers    = 8
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

// Discover returns cameras visible on the NVR host's local network that accept
// the supplied credentials. It checks ONVIF first, then falls back to the
// Dahua RTSP channel URL used by many NVRs. Credentials are used only for the
// scan and are never persisted.
func (s *Service) Discover(ctx context.Context, in DiscoveryInput) ([]DiscoveredCamera, error) {
	candidates, err := discoverONVIF(ctx)
	if err != nil {
		return nil, err
	}
	authenticated := authenticateDiscoveredCameras(ctx, candidates, in, s.DiscoverStreams)

	hosts, err := localIPv4Hosts()
	if err != nil {
		return nil, err
	}
	openHosts := findOpenRTSPHosts(ctx, hosts)
	rtspCameras := discoverDahuaRTSP(ctx, in, openHosts, s.driver.Probe)
	if len(authenticated) == 0 {
		return rtspCameras, nil
	}
	return mergeDiscoveredCameras(authenticated, rtspCameras), nil
}

type streamDiscoverer func(context.Context, StreamDiscoveryInput) (StreamDiscoveryResult, error)

type discoveryAuthenticationResult struct {
	camera        DiscoveredCamera
	authenticated bool
}

func authenticateDiscoveredCameras(
	ctx context.Context,
	candidates []DiscoveredCamera,
	in DiscoveryInput,
	discoverStreams streamDiscoverer,
) []DiscoveredCamera {
	if len(candidates) == 0 {
		return []DiscoveredCamera{}
	}

	results := make(chan discoveryAuthenticationResult, len(candidates))
	for _, candidate := range candidates {
		go func(discovered DiscoveredCamera) {
			streams, err := discoverStreams(ctx, StreamDiscoveryInput{
				XAddr:    discovered.XAddr,
				Username: in.Username,
				Password: in.Password,
			})
			if err != nil {
				slog.Warn("ONVIF discovery authentication failed", "host", discovered.Host, "err", err)
			} else {
				discovered.LiveRTSPURL = strings.TrimSpace(streams.LiveRTSPURL)
				discovered.RecordRTSPURL = strings.TrimSpace(streams.RecordRTSPURL)
				if discovered.RecordRTSPURL == "" {
					discovered.RecordRTSPURL = discovered.LiveRTSPURL
				}
			}
			select {
			case results <- discoveryAuthenticationResult{
				camera:        discovered,
				authenticated: err == nil,
			}:
			case <-ctx.Done():
			}
		}(candidate)
	}

	authenticated := make([]DiscoveredCamera, 0, len(candidates))
	for range candidates {
		select {
		case result := <-results:
			if result.authenticated {
				authenticated = append(authenticated, result.camera)
			}
		case <-ctx.Done():
			return authenticated
		}
	}
	sort.Slice(authenticated, func(i, j int) bool {
		if authenticated[i].Name == authenticated[j].Name {
			return authenticated[i].Host < authenticated[j].Host
		}
		return authenticated[i].Name < authenticated[j].Name
	})
	return authenticated
}

type rtspProber func(context.Context, string, string, string) (ProbeResult, error)

// localIPv4Hosts returns usable addresses in the local IPv4 networks. Broad
// networks are skipped so a Docker bridge or enterprise /16 cannot trigger an
// unexpectedly large scan.
func localIPv4Hosts() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}

	seen := make(map[string]struct{})
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			network, ok := localIPv4Network(address)
			if !ok {
				continue
			}
			ones, bits := network.Mask.Size()
			if bits != 32 || ones < 16 {
				continue
			}
			hostCount := uint64(1) << uint(bits-ones)
			if hostCount <= 2 || hostCount-2 > rtspDiscoveryMaxHosts {
				continue
			}

			base := binary.BigEndian.Uint32(network.IP.To4())
			for offset := uint32(1); uint64(offset) < hostCount-1; offset++ {
				host := make(net.IP, net.IPv4len)
				binary.BigEndian.PutUint32(host, base+offset)
				seen[host.String()] = struct{}{}
			}
		}
	}

	hosts := make([]string, 0, len(seen))
	for host := range seen {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts, nil
}

func localIPv4Network(address net.Addr) (*net.IPNet, bool) {
	switch value := address.(type) {
	case *net.IPNet:
		ip := value.IP.To4()
		if ip == nil {
			return nil, false
		}
		return &net.IPNet{IP: ip, Mask: value.Mask}, true
	case *net.IPAddr:
		return nil, false
	default:
		return nil, false
	}
}

func findOpenRTSPHosts(ctx context.Context, hosts []string) []string {
	if len(hosts) == 0 {
		return []string{}
	}

	jobs := make(chan string, len(hosts))
	for _, host := range hosts {
		jobs <- host
	}
	close(jobs)

	results := make(chan string, len(hosts))
	workerCount := minInt(rtspHostWorkers, len(hosts))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for host := range jobs {
				if ctx.Err() != nil {
					return
				}
				if rtspPortOpen(ctx, host) {
					results <- host
				}
			}
		}()
	}
	workers.Wait()
	close(results)

	open := make([]string, 0)
	for host := range results {
		open = append(open, host)
	}
	sort.Strings(open)
	return open
}

func rtspPortOpen(ctx context.Context, host string) bool {
	probeCtx, cancel := context.WithTimeout(ctx, rtspPortProbeTimeout)
	defer cancel()
	dialer := net.Dialer{}
	connection, err := dialer.DialContext(probeCtx, "tcp", net.JoinHostPort(host, dahuaRTSPPort))
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func discoverDahuaRTSP(
	ctx context.Context,
	in DiscoveryInput,
	hosts []string,
	probe rtspProber,
) []DiscoveredCamera {
	found := make([]DiscoveredCamera, 0)
	for _, host := range hosts {
		jobs := make(chan int, dahuaDiscoveryChannels)
		results := make(chan DiscoveredCamera, dahuaDiscoveryChannels)
		for channel := 1; channel <= dahuaDiscoveryChannels; channel++ {
			jobs <- channel
		}
		close(jobs)

		workerCount := minInt(rtspDiscoveryWorkers, dahuaDiscoveryChannels)
		var workers sync.WaitGroup
		workers.Add(workerCount)
		for range workerCount {
			go func() {
				defer workers.Done()
				for channel := range jobs {
					if ctx.Err() != nil {
						return
					}
					streamURL := dahuaRTSPURL(host, channel)
					probeCtx, cancel := context.WithTimeout(ctx, rtspChannelProbeTimeout)
					result, err := probe(probeCtx, streamURL, in.Username, in.Password)
					cancel()
					if err != nil || !result.Reachable {
						continue
					}
					results <- DiscoveredCamera{
						ID:            fmt.Sprintf("%s-channel-%d", host, channel),
						Name:          fmt.Sprintf("%s · Channel %d", host, channel),
						Host:          net.JoinHostPort(host, dahuaRTSPPort),
						XAddr:         streamURL,
						LiveRTSPURL:   streamURL,
						RecordRTSPURL: streamURL,
					}
				}
			}()
		}
		workers.Wait()
		close(results)
		for camera := range results {
			found = append(found, camera)
		}
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].Host == found[j].Host {
			return found[i].ID < found[j].ID
		}
		return found[i].Host < found[j].Host
	})
	return found
}

func mergeDiscoveredCameras(primary, additional []DiscoveredCamera) []DiscoveredCamera {
	merged := make([]DiscoveredCamera, 0, len(primary)+len(additional))
	seen := make(map[string]struct{}, cap(merged))
	for _, cameras := range [][]DiscoveredCamera{primary, additional} {
		for _, camera := range cameras {
			key := camera.LiveRTSPURL
			if key == "" {
				key = camera.ID
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, camera)
		}
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Name == merged[j].Name {
			return merged[i].Host < merged[j].Host
		}
		return merged[i].Name < merged[j].Name
	})
	return merged
}

func dahuaRTSPURL(host string, channel int) string {
	query := url.Values{}
	query.Set("channel", strconv.Itoa(channel))
	query.Set("subtype", "0")
	return (&url.URL{
		Scheme:   "rtsp",
		Host:     net.JoinHostPort(host, dahuaRTSPPort),
		Path:     "/cam/realmonitor",
		RawQuery: query.Encode(),
	}).String()
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
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
