package service

import (
	"context"
	"net"
	"strings"
)

// SSRF defense helpers:
//   - validateEndpoint blocks http/loopback/private/cloud-metadata URLs at admin submission time
//   - safeDialContext re-validates the real IP at socket level to prevent DNS rebinding

// Blocked cloud metadata hostnames (compared in lowercase).
var monitorBlockedHostnames = map[string]struct{}{
	"localhost":                  {},
	"localhost.localdomain":      {},
	"metadata":                   {},
	"metadata.google.internal":   {},
	"metadata.goog":              {},
	"instance-data":              {},
	"instance-data.ec2.internal": {},
}

// Blocked CIDR ranges covering all private/reserved address spaces.
// Parsed once at init; production path only calls Contains.
var monitorBlockedCIDRs = mustParseCIDRs([]string{
	"127.0.0.0/8",    // IPv4 loopback
	"10.0.0.0/8",     // RFC1918
	"172.16.0.0/12",  // RFC1918
	"192.168.0.0/16", // RFC1918
	"169.254.0.0/16", // link-local (including cloud metadata 169.254.169.254)
	"100.64.0.0/10",  // CGNAT
	"0.0.0.0/8",      // "this network"
	"::1/128",        // IPv6 loopback
	"fc00::/7",       // IPv6 ULA
	"fe80::/10",      // IPv6 link-local
	"::/128",         // IPv6 unspecified
})

// Shared dialer aligned with net/http defaults.
var monitorDialer = &net.Dialer{
	Timeout:   monitorDialTimeout,
	KeepAlive: monitorDialKeepAlive,
}

// mustParseCIDRs parses CIDR strings at package init time; panics on invalid input.
func mustParseCIDRs(rawCIDRs []string) []*net.IPNet {
	networks := make([]*net.IPNet, 0, len(rawCIDRs))
	for idx := 0; idx < len(rawCIDRs); idx++ {
		_, subnet, parseErr := net.ParseCIDR(rawCIDRs[idx])
		if parseErr != nil {
			panic("channel_monitor_ssrf: bad CIDR " + rawCIDRs[idx] + ": " + parseErr.Error())
		}
		networks = append(networks, subnet)
	}
	return networks
}

// isBlockedHostname returns true if the hostname matches the deny list.
func isBlockedHostname(host string) bool {
	if host == "" {
		return true
	}
	_, found := monitorBlockedHostnames[strings.ToLower(host)]
	return found
}

// isPrivateIP returns true if the IP falls within any blocked range.
func isPrivateIP(addr net.IP) bool {
	if addr == nil {
		return true
	}
	if addr.IsUnspecified() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() {
		return true
	}
	for idx := 0; idx < len(monitorBlockedCIDRs); idx++ {
		if monitorBlockedCIDRs[idx].Contains(addr) {
			return true
		}
	}
	return false
}

// isPrivateOrLoopbackHost resolves all A/AAAA records for the given hostname
// and returns true if any resolved IP falls within a private/loopback range.
// IP literals are checked directly.
func isPrivateOrLoopbackHost(reqCtx context.Context, host string) (bool, error) {
	if isBlockedHostname(host) {
		return true, nil
	}
	// Handle IP literal directly.
	if literal := net.ParseIP(host); literal != nil {
		return isPrivateIP(literal), nil
	}
	resolved, lookupErr := net.DefaultResolver.LookupIPAddr(reqCtx, host)
	if lookupErr != nil {
		return false, lookupErr
	}
	if len(resolved) == 0 {
		return true, nil
	}
	for idx := 0; idx < len(resolved); idx++ {
		if isPrivateIP(resolved[idx].IP) {
			return true, nil
		}
	}
	return false, nil
}

// safeDialContext re-validates the target IP at dial time to defend against DNS rebinding.
// It resolves the hostname, checks each IP against the blocklist, and attempts connection
// only to public addresses.
func safeDialContext(reqCtx context.Context, network, addr string) (net.Conn, error) {
	host, port, splitErr := net.SplitHostPort(addr)
	if splitErr != nil {
		return nil, splitErr
	}
	// Fast path for IP literals.
	if literal := net.ParseIP(host); literal != nil {
		if isPrivateIP(literal) {
			return nil, &net.AddrError{Err: "blocked by SSRF policy", Addr: addr}
		}
		return monitorDialer.DialContext(reqCtx, network, addr)
	}
	if isBlockedHostname(host) {
		return nil, &net.AddrError{Err: "blocked by SSRF policy", Addr: addr}
	}
	resolved, lookupErr := net.DefaultResolver.LookupIPAddr(reqCtx, host)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if len(resolved) == 0 {
		return nil, &net.AddrError{Err: "no addresses for host", Addr: host}
	}
	var dialErr error
	for idx := 0; idx < len(resolved); idx++ {
		resolvedIP := resolved[idx].IP
		if isPrivateIP(resolvedIP) {
			dialErr = &net.AddrError{Err: "blocked by SSRF policy", Addr: resolvedIP.String()}
			continue
		}
		conn, connErr := monitorDialer.DialContext(reqCtx, network, net.JoinHostPort(resolvedIP.String(), port))
		if connErr == nil {
			return conn, nil
		}
		dialErr = connErr
	}
	if dialErr == nil {
		dialErr = &net.AddrError{Err: "no usable addresses", Addr: host}
	}
	return nil, dialErr
}
