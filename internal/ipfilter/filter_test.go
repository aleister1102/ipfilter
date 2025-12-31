package ipfilter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestProcessorProcessIPs(t *testing.T) {
	t.Parallel()

	proc := processor{maxExpand: DefaultMaxCIDRAddresses}
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "public IPv4",
			input: "203.0.113.10",
			want:  []string{"203.0.113.10"},
		},
		{
			name:  "private IPv4",
			input: "192.168.0.10",
		},
		{
			name:  "this network 0.0.0.0",
			input: "0.0.0.0",
		},
		{
			name:  "this network address in range",
			input: "0.0.0.5",
		},
		{
			name:  "public IPv6",
			input: "2001:db8::1",
			want:  []string{"2001:db8::1"},
		},
		{
			name:  "private IPv6",
			input: "fc00::1",
		},
		{
			name:    "invalid address",
			input:   "not-an-ip",
			wantErr: true,
		},
		{
			name:  "IPv4 with port",
			input: "203.0.113.10:443",
			want:  []string{"203.0.113.10"},
		},
		{
			name:  "IPv4 with port 80",
			input: "203.0.113.10:80",
			want:  []string{"203.0.113.10"},
		},
		{
			name:  "IPv4 with port 8443",
			input: "203.0.113.10:8443",
			want:  []string{"203.0.113.10"},
		},
		{
			name:  "private IPv4 with port",
			input: "192.168.0.10:443",
		},
		{
			name:  "IPv6 with port brackets",
			input: "[2001:db8::1]:443",
			want:  []string{"2001:db8::1"},
		},
		{
			name:  "IPv6 with brackets no port",
			input: "[2001:db8::1]",
			want:  []string{"2001:db8::1"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := proc.process(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slicesEqual(got, tt.want) {
				t.Fatalf("unexpected output: got %v want %v", got, tt.want)
			}
		})
	}
}

func TestProcessorProcessCIDRs(t *testing.T) {
	t.Parallel()
	proc := processor{maxExpand: DefaultMaxCIDRAddresses}

	tests := []struct {
		name string
		cidr string
		want []string
	}{
		{
			name: "private IPv4 CIDR",
			cidr: "10.0.0.0/24",
		},
		{
			name: "public IPv4 CIDR small",
			cidr: "203.0.113.0/30",
			want: []string{
				"203.0.113.0",
				"203.0.113.1",
				"203.0.113.2",
				"203.0.113.3",
			},
		},
		{
			name: "public IPv6 CIDR small",
			cidr: "2001:db8::/126",
			want: []string{
				"2001:db8::",
				"2001:db8::1",
				"2001:db8::2",
				"2001:db8::3",
			},
		},
		{
			name: "public IPv4 CIDR large limit",
			cidr: "203.0.113.0/24",
			want: []string{"203.0.113.0/24"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := proc.process(tt.cidr)
			if err != nil {
				t.Fatalf("process() error = %v", err)
			}
			if !slicesEqual(got, tt.want) {
				t.Fatalf("unexpected output: got %v want %v", got, tt.want)
			}
		})
	}
}

func TestFilterStream(t *testing.T) {
	input := strings.Join([]string{
		"203.0.113.1",
		"192.168.1.10",
		"203.0.113.0/30",
		"",
	}, "\n")

	var buf bytes.Buffer

	err := Filter(strings.NewReader(input), &buf, Options{})
	if err != nil {
		t.Fatalf("Filter() error = %v", err)
	}

	want := strings.Join([]string{
		"203.0.113.1",
		"203.0.113.0",
		"203.0.113.1",
		"203.0.113.2",
		"203.0.113.3",
		"",
	}, "\n")

	if diff := compareLines(buf.String(), want); diff != "" {
		t.Fatalf("unexpected output:\n%s", diff)
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareLines(got, want string) string {
	if got == want {
		return ""
	}
	return fmt.Sprintf("got:\n%s\nwant:\n%s", got, want)
}
