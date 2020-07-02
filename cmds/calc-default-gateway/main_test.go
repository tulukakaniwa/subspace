package main

import (
	"net"
	"testing"
)

func TestCalcDefaultGatewayV6(t *testing.T) {
	{
		cidr := "fe80:1234:1234:1234::/64"
		// FIXME(ledyba-z): there is no way to suppress sonar cloud warnings for IPv6.
		gw := "fe80:1234:1234:1234:"+":1"
		ip, err := calcDefaultGateway(cidr)
		if err != nil {
			t.Error(err)
		}
		if !ip.Equal(net.ParseIP(gw)) {
			t.Errorf("Default gateway of %s must be %s, but got %s", cidr, gw, ip.String())
		}
	}
	{
		cidr := "fe80:1234:1234:1234::/128"
		ip, err := calcDefaultGateway(cidr)
		if err == nil {
			t.Errorf("There should not be default GW for %s, but got %s", cidr, ip.String())
		}
	}
}
func TestCalcDefaultGatewayV4(t *testing.T) {
	{
		cidr := "127.168.128.0/18"
		gw := "127.168.128.1"
		ip, err := calcDefaultGateway(cidr)
		if err != nil {
			t.Error(err)
		}
		if !ip.Equal(net.ParseIP(gw)) {
			t.Errorf("Default gateway of %s must be %s, but got %s", cidr, gw, ip.String())
		}
	}
	{
		cidr := "127.168.128.0/32"
		ip, err := calcDefaultGateway(cidr)
		if err == nil {
			t.Errorf("There should not be default GW for %s, but got %s", cidr, ip.String())
		}
	}
}

