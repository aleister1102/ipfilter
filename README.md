# ipfilter

Command-line helper that removes private IPv4 / IPv6 addresses and private CIDR
ranges from an input list. Public CIDRs with 128 addresses or fewer are expanded
into individual IPs; larger ranges are preserved in their original notation.

## Build

```bash
cd /Users/quan.m.le/Repos/ipctl
make build
```

This produces `bin/ipfilter` for the host platform. To cross-compile release
binaries for macOS and Linux (amd64 and arm64), run:

```bash
make release
```

Artifacts are written to `dist/` with platform suffixes (for example,
`dist/ipfilter-darwin-arm64`).

## Usage

```
ipfilter [-i input_file] [-o output_file]
```

- `-i` (optional): path to a file containing newline-delimited IPs / CIDRs. When
  omitted, stdin is used.
- `-o` (optional): path to write the filtered results. Defaults to stdout.

## Examples

Filter data from stdin and display results:

```bash
cat ips.txt | ./ipfilter
```

Read from a file and write to another file:

```bash
./ipfilter -i ips.txt -o filtered.txt
```

## Behavior

- Skips private IPv4 ranges (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`,
  etc.), loopback, link-local, and IPv6 ULA / link-local ranges.
- Drops any CIDR wholly contained within those private ranges.
- Expands public CIDRs up to 128 addresses; larger CIDRs remain as ranges to
  avoid enormous output.

