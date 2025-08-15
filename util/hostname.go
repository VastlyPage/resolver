package hlbaby

import (
	"crypto/tls"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	protocolCache   = make(map[string]cachedProtocol)
	cacheMutex      sync.Mutex
	cacheExpiration = 60 * time.Second
)

type cachedProtocol struct {
	protocol string
	lastUsed time.Time
}

func init() {
	go cleanUpCache()
}

func cleanUpCache() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cacheMutex.Lock()
		for host, entry := range protocolCache {
			if time.Since(entry.lastUsed) > cacheExpiration {
				delete(protocolCache, host)
			}
		}
		cacheMutex.Unlock()
	}
}

func NormalizeHostname(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	if idx := strings.IndexAny(url, "/:@"); idx != -1 {
		url = url[idx+1:]
	}

	if strings.Contains(url, "[") && strings.Contains(url, "]") {
		url = strings.TrimSuffix(strings.Split(url, "]")[0], "[")
	}

	return url
}

func GetClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func GetSupportedProtocol(hostname string) string {
	cacheMutex.Lock()
	if entry, found := protocolCache[hostname]; found {
		cacheMutex.Unlock()
		return entry.protocol
	}
	cacheMutex.Unlock()

	client := GetClient()
	protocol := "http"
	if resp, err := client.Head("https://" + hostname); err == nil {
		defer resp.Body.Close()
		protocol = "https"
	}

	cacheMutex.Lock()
	protocolCache[hostname] = cachedProtocol{
		protocol: protocol,
		lastUsed: time.Now(),
	}
	cacheMutex.Unlock()

	return protocol
}

func MakeURL(hostname string, protoEnum uint) string {
	var proto string
	switch protoEnum {
	case 1:
		proto = "http"
	case 2:
		proto = "https"
	default:
		proto = GetSupportedProtocol(hostname)
	}

	if isIPv6(hostname) {
		return proto + "://[" + hostname + "]"
	}
	return proto + "://" + hostname
}

func IsHostReachable(hostname string) uint {
	cacheMutex.Lock()
	if entry, found := protocolCache[hostname]; found {
		cacheMutex.Unlock()
		if entry.protocol == "https" {
			return 2
		}
		return 1
	}
	cacheMutex.Unlock()

	client := GetClient()
	resp, err := client.Head(MakeURL(hostname, 0))
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if strings.HasPrefix(resp.Request.URL.String(), "https://") {
		return 2
	}
	return 1
}

func isIPv6(hostname string) bool {
	return strings.Contains(hostname, ":") && !strings.Contains(hostname, "]")
}
