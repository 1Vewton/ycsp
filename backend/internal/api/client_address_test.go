package api

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func mustTrustedPrefixes(t *testing.T, raw string) []netip.Prefix {
	t.Helper()

	prefixes, err := ParseTrustedProxies(raw)
	if err != nil {
		t.Fatalf("ParseTrustedProxies(%q): %v", raw, err)
	}
	return prefixes
}

func addressRequest(remoteAddr string, forwardedFor string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/sessions", nil)
	request.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}
	return request
}

func TestClientAddressDefaultsToDirectPeer(t *testing.T) {
	resolver := newClientAddressResolver(nil)

	got := resolver.Address(addressRequest("198.51.100.9:4321", "203.0.113.7"))
	if got != "198.51.100.9" {
		t.Errorf("address = %q, want direct peer 198.51.100.9", got)
	}
}

func TestClientAddressUsesForwardedForFromTrustedProxy(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32"))

	got := resolver.Address(addressRequest("127.0.0.1:54321", "203.0.113.7"))
	if got != "203.0.113.7" {
		t.Errorf("address = %q, want forwarded client 203.0.113.7", got)
	}
}

func TestClientAddressIgnoresClientSuppliedHeadOfForwardedFor(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32"))

	// 左侧 10.9.9.9 可以是客户端自己塞的；右侧才是 nginx 追加的真实地址。
	got := resolver.Address(addressRequest("127.0.0.1:54321", "10.9.9.9, 203.0.113.7"))
	if got != "203.0.113.7" {
		t.Errorf("address = %q, want rightmost untrusted 203.0.113.7", got)
	}
}

func TestClientAddressWalksPastTrustedProxyChain(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32,10.0.0.0/8"))

	got := resolver.Address(addressRequest("127.0.0.1:54321", "203.0.113.7, 10.1.2.3"))
	if got != "203.0.113.7" {
		t.Errorf("address = %q, want 203.0.113.7 after skipping trusted 10.1.2.3", got)
	}
}

func TestClientAddressIgnoresForwardedForWhenPeerIsUntrusted(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32"))

	got := resolver.Address(addressRequest("198.51.100.9:4321", "203.0.113.7"))
	if got != "198.51.100.9" {
		t.Errorf("address = %q, want direct peer (spoofed header ignored)", got)
	}
}

func TestClientAddressAcceptsPortsAndIPv6(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "::1/128"))

	got := resolver.Address(addressRequest("[::1]:54321", "2001:db8::99"))
	if got != "2001:db8::99" {
		t.Errorf("address = %q, want 2001:db8::99", got)
	}

	got = resolver.Address(addressRequest("[::1]:54321", "203.0.113.7:4711"))
	if got != "203.0.113.7" {
		t.Errorf("address = %q, want 203.0.113.7 (port stripped)", got)
	}
}

func TestClientAddressUnmapsIPv4MappedIPv6(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32"))

	got := resolver.Address(addressRequest("127.0.0.1:54321", "::ffff:203.0.113.7"))
	if got != "203.0.113.7" {
		t.Errorf("address = %q, want unmapped 203.0.113.7", got)
	}
}

func TestClientAddressFallsBackWhenForwardedForIsMalformed(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32"))

	// 右侧（最近一跳）不是合法地址：退回直连地址，绝不把可疑字符串当学生地址。
	got := resolver.Address(addressRequest("127.0.0.1:54321", "203.0.113.7, not-an-ip"))
	if got != "127.0.0.1" {
		t.Errorf("address = %q, want direct proxy 127.0.0.1", got)
	}

	// 左侧坏了但右侧合法：右侧先被读到，仍然能拿到真实客户端。
	got = resolver.Address(addressRequest("127.0.0.1:54321", "not-an-ip, 203.0.113.7"))
	if got != "203.0.113.7" {
		t.Errorf("address = %q, want 203.0.113.7 from the valid rightmost hop", got)
	}
}

func TestClientAddressFallsBackWhenEveryHopIsTrusted(t *testing.T) {
	resolver := newClientAddressResolver(mustTrustedPrefixes(t, "127.0.0.1/32,10.0.0.0/8"))

	got := resolver.Address(addressRequest("127.0.0.1:54321", "10.1.2.3"))
	if got != "127.0.0.1" {
		t.Errorf("address = %q, want direct proxy when no untrusted hop is present", got)
	}
}

func TestParseTrustedProxies(t *testing.T) {
	prefixes, err := ParseTrustedProxies(" 127.0.0.1 , 10.0.0.0/8 ")
	if err != nil {
		t.Fatalf("ParseTrustedProxies: %v", err)
	}
	if len(prefixes) != 2 {
		t.Fatalf("prefix count = %d, want 2", len(prefixes))
	}
	if !prefixes[0].Contains(netip.MustParseAddr("127.0.0.1")) {
		t.Errorf("first prefix = %v, want 127.0.0.1/32", prefixes[0])
	}
	if !prefixes[1].Contains(netip.MustParseAddr("10.4.5.6")) {
		t.Errorf("second prefix = %v, want 10.0.0.0/8", prefixes[1])
	}

	mappedPrefixes := mustTrustedPrefixes(t, "::ffff:127.0.0.1/128")
	if len(mappedPrefixes) != 1 || !mappedPrefixes[0].Contains(netip.MustParseAddr("127.0.0.1")) {
		t.Errorf("IPv4-mapped prefix = %v, want an IPv4 127.0.0.1/32", mappedPrefixes)
	}
}

func TestParseTrustedProxiesRejectsUnsafeOrInvalidEntries(t *testing.T) {
	for _, raw := range []string{"0.0.0.0/0", "::/0", "not-an-ip", "10.0.0.0/33"} {
		if _, err := ParseTrustedProxies(raw); err == nil {
			t.Errorf("ParseTrustedProxies(%q) accepted an unsafe/invalid entry", raw)
		}
	}
}
