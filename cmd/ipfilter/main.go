package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/aleister1102/selfupdate"
	"github.com/quan-m-le/ipctl/internal/ipfilter"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ipfilter", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("i", "", "input file (defaults to stdin)")
	outputPath := fs.String("o", "", "output file (defaults to stdout)")
	versionFlag := fs.Bool("version", false, "show version")
	updateFlag := fs.Bool("update", false, "self-update the ipfilter binary to the latest stable release")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [-i input] [-o output]\n", fs.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if *versionFlag {
		fmt.Fprintf(stdout, "ipfilter %s (commit: %s, built at: %s)\n", Version, Commit, Date)
		return 0
	}

	if *updateFlag {
		err := selfupdate.Update(context.Background(), selfupdate.Config{
			Owner:          "aleister1102",
			Repo:           "ipfilter",
			Binary:         "ipfilter",
			CurrentVersion: Version,
		}, stdout, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "ipfilter: self-update failed: %v\n", err)
			return 1
		}
		return 0
	}

	in := stdin
	if *inputPath != "" {
		file, err := os.Open(*inputPath)
		if err != nil {
			fmt.Fprintf(stderr, "ipfilter: %v\n", err)
			return 1
		}
		defer file.Close()
		in = file
	}

	out := stdout
	if *outputPath != "" {
		file, err := os.Create(*outputPath)
		if err != nil {
			fmt.Fprintf(stderr, "ipfilter: %v\n", err)
			return 1
		}
		defer file.Close()
		out = file
	}

	opts := ipfilter.Options{MaxCIDRAddresses: ipfilter.DefaultMaxCIDRAddresses}
	if err := ipfilter.Filter(in, out, opts); err != nil {
		fmt.Fprintf(stderr, "ipfilter: %v\n", err)
		return 1
	}

	return 0
}
