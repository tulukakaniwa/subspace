package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/subspacecommunity/subspace/util"
)

var help *bool = flag.Bool("h", false, "show help and exit")
var version string

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `Usage:
  %[1]s [-h]
  %[1]s CIDR
Flags:
`, filepath.Base(os.Args[0]))
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(os.Stderr, "Version: %s", version)
}

func main() {
	var err error
	flag.Parse()
	if *help {
		usage()
		return
	}
	if flag.NArg() != 1 {
		usage()
		os.Exit(-1)
	}
	cidr := flag.Arg(0)
	gw, err := util.CalcDefaultGateway(cidr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to calc default gw: %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("%s\n", gw.String())
}
