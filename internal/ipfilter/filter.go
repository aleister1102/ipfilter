package ipfilter

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"
)

// DefaultMaxCIDRAddresses controls how many addresses a CIDR may contain
// before the tool keeps the original range instead of expanding it.
const DefaultMaxCIDRAddresses = 128

// Options configure the filtering behavior.
type Options struct {
	// MaxCIDRAddresses caps how many IPs will be emitted for a single CIDR.
	// Values <= 0 fall back to DefaultMaxCIDRAddresses.
	MaxCIDRAddresses int
}

// Filter reads line-oriented IP / CIDR data from r, removes private entries,
// expands small public CIDRs, and writes the remaining addresses to w.
func Filter(r io.Reader, w io.Writer, opts Options) error {
	max := opts.MaxCIDRAddresses
	if max <= 0 {
		max = DefaultMaxCIDRAddresses
	}

	proc := processor{maxExpand: max}
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entries, err := proc.process(line)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if _, err := fmt.Fprintln(writer, entry); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return writer.Flush()
}

type processor struct {
	maxExpand int
}

func (p processor) process(entry string) ([]string, error) {
	// Strip port if present (e.g., "192.168.1.1:443" -> "192.168.1.1")
	entry = stripPort(entry)

	ip := net.ParseIP(entry)
	if ip != nil {
		return p.handleIP(ip)
	}

	_, network, err := net.ParseCIDR(entry)
	if err != nil {
		return nil, fmt.Errorf("invalid entry %q: %w", entry, err)
	}

	return p.handleCIDR(network)
}

// stripPort removes the port suffix from an IP address if present.
// Handles both IPv4 (192.168.1.1:443) and IPv6 ([::1]:443) formats.
func stripPort(entry string) string {
	// Handle IPv6 with port: [::1]:443
	if strings.HasPrefix(entry, "[") {
		if idx := strings.LastIndex(entry, "]:"); idx != -1 {
			return entry[1:idx]
		}
		// Just brackets without port: [::1]
		if strings.HasSuffix(entry, "]") {
			return strings.Trim(entry, "[]")
		}
		return entry
	}

	// Handle IPv4 with port: 192.168.1.1:443
	// Only strip if there's exactly one colon (to avoid breaking IPv6)
	if strings.Count(entry, ":") == 1 {
		if idx := strings.LastIndex(entry, ":"); idx != -1 {
			return entry[:idx]
		}
	}

	return entry
}

func (p processor) handleIP(ip net.IP) ([]string, error) {
	normalized := normalizeIP(ip)
	if isPrivateIP(normalized) {
		return nil, nil
	}
	return []string{formatIP(normalized)}, nil
}

func (p processor) handleCIDR(network *net.IPNet) ([]string, error) {
	if isPrivateCIDR(network) {
		return nil, nil
	}

	expanded, err := expandCIDR(network, p.maxExpand)
	if err != nil {
		return nil, err
	}
	return expanded, nil
}

func expandCIDR(network *net.IPNet, limit int) ([]string, error) {
	if limit <= 0 {
		limit = DefaultMaxCIDRAddresses
	}

	size := cidrSize(network)
	if size.Sign() == 0 {
		return nil, fmt.Errorf("unable to determine size of CIDR %q", network.String())
	}

	limitBig := big.NewInt(int64(limit))
	if size.Cmp(limitBig) == 1 {
		return []string{network.String()}, nil
	}

	count := int(size.Int64())
	result := make([]string, 0, count)

	start := normalizeIP(network.IP)
	current := make(net.IP, len(start))
	copy(current, start)

	for i := 0; i < count; i++ {
		result = append(result, formatIP(current))
		incrementIP(current)
	}

	return result, nil
}

func cidrSize(network *net.IPNet) *big.Int {
	ones, bits := network.Mask.Size()
	if ones == -1 || bits == 0 {
		return big.NewInt(0)
	}
	exp := uint(bits - ones)
	return new(big.Int).Lsh(big.NewInt(1), exp)
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func normalizeIP(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip.To16()
}

func formatIP(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.String()
}

var privateRanges = mustParseCIDRs(
	// IPv4 RFC1918
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	// IPv4 special-use
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	// IPv6 loopback + ULA + link-local
	"::1/128",
	"fc00::/7",
	"fe80::/10",
)

func isPrivateIP(ip net.IP) bool {
	for _, rng := range privateRanges {
		if rng.Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivateCIDR(network *net.IPNet) bool {
	for _, rng := range privateRanges {
		if cidrSubset(network, rng) {
			return true
		}
	}
	return false
}

func cidrSubset(child, parent *net.IPNet) bool {
	childOnes, childBits := child.Mask.Size()
	parentOnes, parentBits := parent.Mask.Size()

	if childBits != parentBits {
		return false
	}

	if childOnes < parentOnes {
		return false
	}

	return parent.Contains(child.IP)
}

func mustParseCIDRs(cidrs ...string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid CIDR %q: %v", cidr, err))
		}
		nets = append(nets, network)
	}
	return nets
}
