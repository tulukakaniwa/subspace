package util

import (
	"net"
	"testing" // テストで使える関数・構造体が用意されているパッケージをimport
)

func TestCalcDefaultGatewayV6(t *testing.T) {
	{
		cidr := "1234:1234:1234:1234::/64"
		gw := "1234:1234:1234:1234::1"
		ip, err := CalcDefaultGateway(cidr)
		if err != nil {
			t.Error(err)
		}
		if !ip.Equal(net.ParseIP(gw)) {
			t.Errorf("Default gateway of %s must be %s, but got %s", cidr, gw, ip.String())
		}
	}
	{
		cidr := "1234:1234:1234:1234::/128"
		ip, err := CalcDefaultGateway(cidr)
		if err == nil {
			t.Errorf("There should not be default GW for %s, but got %s", cidr, ip.String())
		}
	}
}

func TestCalcDefaultGatewayV4(t *testing.T) {
	{
		cidr := "192.168.128.0/18"
		gw := "192.168.128.1"
		ip, err := CalcDefaultGateway(cidr)
		if err != nil {
			t.Error(err)
		}
		if !ip.Equal(net.ParseIP(gw)) {
			t.Errorf("Default gateway of %s must be %s, but got %s", cidr, gw, ip.String())
		}
	}
	{
		cidr := "192.168.128.0/32"
		ip, err := CalcDefaultGateway(cidr)
		if err == nil {
			t.Errorf("There should not be default GW for %s, but got %s", cidr, ip.String())
		}
	}
}

func TestGenerateIPAddr(t *testing.T) {
	_, v4Net, _ := net.ParseCIDR("10.10.0.0/16")
	_, v6Net, _ := net.ParseCIDR("fd80::/112")
	{
		ipv4, ipv6, err := GenerateIPAddr(v4Net, v6Net, 100)
		if err != nil {
			t.Error("Failed to generate IP: ", err)
		}
		if expected := "10.10.0.100"; ipv4.String() != expected {
			t.Errorf("Failed to generate IPv4: %s(expected) != %s(actual)", ipv4.String(), expected)
		}
		if expected := "fd80::100"; ipv6.String() != expected {
			t.Errorf("Failed to generate IPv6: %s(expected) != %s(actual)", ipv6.String(), expected)
		}
	}
	_, ipv6, err := GenerateIPAddr(v4Net, v6Net, 256)
	if err == nil {
		t.Errorf("%s contain only 255 valid v6 address, but got: %s", v6Net.String(), ipv6.String())
	}
}
