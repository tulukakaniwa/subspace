package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
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
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse CIDR: %v\n", err)
		os.Exit(-1)
	}
	if ones, bits := network.Mask.Size(); bits-ones == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Given CIDR (%s) doet not represent a network.\n", cidr)
		os.Exit(-1)
	}
	netIP := network.IP[:]
	netIP[len(netIP)-1] |= 1
	fmt.Printf("%s\n", netIP.String())
}
