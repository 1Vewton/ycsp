package api

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// clientAddressResolver 从请求里取出“发起登录的学生地址”，用于登录失败限流。
//
// 默认只信直连地址：不读任何转发头，客户端伪造 X-Forwarded-For 也不会生效。
// 部署在 nginx 等反向代理后面时，把代理自己的地址 / CIDR 配进 YCSP_TRUSTED_PROXIES，
// 才会从 X-Forwarded-For 里解析真实客户端（见 backend/README.md《反向代理部署》）。
type clientAddressResolver struct {
	trustedProxies []netip.Prefix
}

func newClientAddressResolver(trustedProxies []netip.Prefix) clientAddressResolver {
	return clientAddressResolver{trustedProxies: trustedProxies}
}

// ParseTrustedProxies 解析逗号分隔的 IP / CIDR 列表，例如 "127.0.0.1/32,10.0.0.0/8"。
// 空字符串表示不信任任何代理，保持“只按直连地址”的默认行为。
func ParseTrustedProxies(raw string) ([]netip.Prefix, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	prefixes := make([]netip.Prefix, 0)
	for _, entry := range strings.Split(trimmed, ",") {
		candidate := strings.TrimSpace(entry)
		if candidate == "" {
			continue
		}
		prefix, err := parseTrustedProxy(candidate)
		if err != nil {
			return nil, fmt.Errorf("YCSP_TRUSTED_PROXIES: %w", err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func parseTrustedProxy(entry string) (netip.Prefix, error) {
	if strings.Contains(entry, "/") {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return netip.Prefix{}, fmt.Errorf("invalid CIDR %q", entry)
		}
		prefix = normalizePrefix(prefix)
		if prefix.Bits() == 0 {
			return netip.Prefix{}, fmt.Errorf(
				"refusing %q: trusting every address lets clients forge X-Forwarded-For", entry,
			)
		}
		return prefix, nil
	}

	address, err := netip.ParseAddr(entry)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("invalid IP %q", entry)
	}
	address = address.Unmap()
	return netip.PrefixFrom(address, address.BitLen()), nil
}

// normalizePrefix 把 IPv4 映射到 IPv6 的写法（::ffff:a.b.c.d/n）折回普通 IPv4 前缀。
func normalizePrefix(prefix netip.Prefix) netip.Prefix {
	if prefix.Addr().Is4In6() && prefix.Bits() >= 96 {
		return netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
	}
	return prefix.Masked()
}

// Address 返回用于限流的客户端地址（不带端口）。
//
// 只有直连地址属于可信代理时才看 X-Forwarded-For；并从头尾顺序的右侧开始，
// 取第一个“不受信”的地址：左侧内容可能是客户端自己伪造的，右侧才是最近一跳代理追加的。
func (resolver clientAddressResolver) Address(request *http.Request) string {
	direct := directAddress(request.RemoteAddr)
	directIP, ok := parseAddress(direct)
	if !ok {
		return direct
	}
	directIP = directIP.Unmap()
	if !resolver.trusts(directIP) {
		return directIP.String()
	}

	hops := strings.Split(strings.Join(request.Header.Values("X-Forwarded-For"), ","), ",")
	for index := len(hops) - 1; index >= 0; index-- {
		hop, ok := parseAddress(strings.TrimSpace(hops[index]))
		if !ok {
			// 头部格式不对：退回直连地址，至少不会把伪造值当成学生地址。
			return directIP.String()
		}
		hop = hop.Unmap()
		if !resolver.trusts(hop) {
			return hop.String()
		}
	}
	// 只有受信代理、没有客户端地址：退回直连地址。
	return directIP.String()
}

func (resolver clientAddressResolver) trusts(address netip.Addr) bool {
	for _, prefix := range resolver.trustedProxies {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

// directAddress 取直连地址（不带端口）；RemoteAddr 不带端口时原样返回。
func directAddress(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// parseAddress 解析单个地址；同时接受代理偶尔带上的端口写法。
func parseAddress(raw string) (netip.Addr, bool) {
	if raw == "" {
		return netip.Addr{}, false
	}
	if address, err := netip.ParseAddr(raw); err == nil {
		return address, true
	}
	if addressPort, err := netip.ParseAddrPort(raw); err == nil {
		return addressPort.Addr(), true
	}
	return netip.Addr{}, false
}
