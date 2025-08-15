package hlnames

import (
	"fmt"
	"strings"
	"sync"

	hlbabyutil "vastly.page/hl.baby/util"
)

func ResolveHostAndURL(kv map[string]interface{}, path string) (host, url string, isRedirect bool, err error) {
	if redirect, ok := kv["REDIRECT"].(string); ok {
		isRedirect = true
		url = hlbabyutil.EnsureHTTPPrefix(redirect)
		return
	}

	if cid, ok := kv["CID"].(string); ok {
		host = hlbabyutil.GetGatewayURL(cid)
		url = hlbabyutil.MakeURL(host, 2) + path
		return
	}

	if cname, ok := kv["CNAME"].(string); ok {
		proto := hlbabyutil.IsHostReachable(cname)
		if proto > 0 {
			host = hlbabyutil.NormalizeHostname(cname)
			url = hlbabyutil.MakeURL(host, proto) + path
			return
		}
		err = fmt.Errorf("CNAME record not reachable")
		return
	}

	if mirror, ok := kv["MIRROR"].(string); ok {
		host = hlbabyutil.NormalizeHostname(mirror)
		proto := hlbabyutil.IsHostReachable(host)
		if proto > 0 {
			url = strings.TrimRight(mirror, "/") + path
			return
		}
		err = fmt.Errorf("MIRROR record not reachable")
		return
	}

	host, err = ResolveIPRecords(kv)
	if err == nil {
		url = hlbabyutil.MakeURL(host, 1) + path
	}
	return
}

func ResolveIPRecords(kv map[string]interface{}) (string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstRespondingHost string
	var found bool
	var wi int

	checkHost := func(ip string) {
		defer wg.Done()
		ip = strings.TrimSpace(ip)
		if hlbabyutil.IsHostReachable(ip) > 0 {
			mu.Lock()
			if !found {
				firstRespondingHost = ip
				found = true
				wg.Add(-wi)
			}
			mu.Unlock()
		}
	}

	if aaaa, ok := kv["AAAA"].(string); ok {
		for ip := range strings.SplitSeq(aaaa, ",") {
			wg.Add(1)
			wi++
			go checkHost(ip)
		}
	}

	if a, ok := kv["A"].(string); ok {
		for ip := range strings.SplitSeq(a, ",") {
			wg.Add(1)
			wi++
			go checkHost(ip)
		}
	}

	wg.Wait()
	if firstRespondingHost == "" {
		return "", fmt.Errorf("no responding records found")
	}
	return firstRespondingHost, nil
}
