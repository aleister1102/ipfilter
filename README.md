# ipfilter

A command-line tool that filters out private IPv4/IPv6 addresses and private CIDR ranges from input lists, keeping only public IPs.

> **Deprecated:** `ipinfo grepip --exclude-reserved` can already perform the same filtering as `ipfilter`. For example:
> ```
> [some command that produces ips] | ipinfo grepip --exclude-reserved
> ```
> Consider switching to `ipinfo/cli` instead of relying on this repository.

**Key features:**
- Removes private IPv4 ranges (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, etc.)
- Filters IPv6 ULA and link-local ranges
- Expands public CIDRs with ≤128 addresses into individual IPs
- Preserves larger CIDR ranges in their original notation

## Quick Start

### Build

```bash
make build
```

Produces `bin/ipfilter` for your platform.

For cross-platform release binaries (macOS and Linux, amd64/arm64):

```bash
make release
```

Artifacts go to `dist/` with platform suffixes (e.g., `dist/ipfilter-darwin-arm64`).

## Usage

```
ipfilter [-i input_file] [-o output_file]
```

**Options:**
- `-i` — Input file with newline-delimited IPs/CIDRs (default: stdin)
- `-o` — Output file for filtered results (default: stdout)
- `-update` — Self-update `ipfilter` to the latest stable release from `aleister1102/ipfilter`.
> Set `GITHUB_TOKEN` if you hit GitHub rate limits during self-update.

## Examples

**From stdin to stdout:**
```bash
cat ips.txt | ./ipfilter
```

**From file to file:**
```bash
./ipfilter -i ips.txt -o filtered.txt
```

## How It Works

**Private IPv4 Ranges**
Removes addresses from RFC 1918 private ranges:
- `10.0.0.0/8` — Class A private network
- `172.16.0.0/12` — Class B private network
- `192.168.0.0/16` — Class C private network

Also filters loopback (`127.0.0.0/8`) and link-local (`169.254.0.0/16`) addresses.

**IPv6 Private Ranges**
Filters out Unique Local Addresses (ULA, `fc00::/7`) and link-local addresses (`fe80::/10`), which are IPv6 equivalents of private ranges.

**CIDR Expansion**
Public CIDR ranges with 128 addresses or fewer are expanded into individual IP addresses. For example, `203.0.113.0/30` (4 IPs) becomes:
```
203.0.113.0
203.0.113.1
203.0.113.2
203.0.113.3
```

**Large CIDR Preservation**
CIDRs larger than 128 addresses stay in their original notation to keep output manageable. For example, `203.0.113.0/24` (256 IPs) remains as-is rather than expanding to 256 lines.
