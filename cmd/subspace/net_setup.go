package main

import (
	"fmt"
	"net"

	"github.com/coreos/go-iptables/iptables"
)

var (
	wgConfig *wireguardConfig
)

type wireguardConfig struct {
	gatewayIPv4 net.IP
	gatewayIPv6 net.IP
	networkIPv4 *net.IPNet
	networkIPv6 *net.IPNet
}

func initWireguardConfig() error {
	gatewayIPv4, networkIPv4, err := calcDefaultGateway(networkIPv4)
	if err != nil {
		return err
	}
	gatewayIPv6, networkIPv6, err := calcDefaultGateway(networkIPv6)
	if err != nil {
		return err
	}
	wgConfig = &wireguardConfig{
		gatewayIPv4: gatewayIPv4,
		networkIPv4: networkIPv4,
		gatewayIPv6: gatewayIPv6,
		networkIPv6: networkIPv6,
	}
	return nil
}

func (conf *wireguardConfig) configureWireguard() (string, error) {
	var err error
	err = conf.configureNAT(iptables.ProtocolIPv4, networkIPv4, conf.gatewayIPv4.String())
	if err != nil {
		return "", err
	}
	if ipv6NatEnabled {
		err = conf.configureNAT(iptables.ProtocolIPv6, networkIPv6, conf.gatewayIPv6.String())
		if err != nil {
			return "", err
		}
	}
	maskLenIPv4, _ := conf.networkIPv4.Mask.Size()
	maskLenIPv6, _ := conf.networkIPv6.Mask.Size()
	return bash(`
# Set DNS server
echo "nameserver {{.Nameserver}}" > /etc/resolv.conf

# Create wireguard device
if ip link show wg0 2>/dev/null; then
  ip link delete wg0
fi
ip link add dev wg0 type wireguard
ip address add dev wg0 "{{.GatewayIPv4WithCIDR}}"
{{if .IPv6NatEnabled}}
ip address add dev wg0 "{{.GatewayIPv6WithCIDR}}"
{{end}}
wg setconf wg0 /data/wireguard/server.conf
ip link set up dev wg0

{{if .DnsmasqEnabled}}
# Reload dnsmasq
sed -i -e 's/listen-address=.\+$/listen-address=127.0.0.1,{{.GatewayIPv4}},{{.GatewayIPv6}}/g' /etc/dnsmasq.conf
sv restart dnsmasq
{{end}}
`, struct {
		Nameserver string
		IPv6NatEnabled bool
		GatewayIPv4 string
		GatewayIPv6 string
		GatewayIPv4WithCIDR string
		GatewayIPv6WithCIDR string
		DnsmasqEnabled bool
	}{
		Nameserver:          nameserver,
		IPv6NatEnabled:      ipv6NatEnabled,
		GatewayIPv4:         conf.gatewayIPv4.String(),
		GatewayIPv6:         conf.gatewayIPv6.String(),
		GatewayIPv4WithCIDR: fmt.Sprintf("%s/%d", conf.gatewayIPv4.String(), maskLenIPv4),
		GatewayIPv6WithCIDR: fmt.Sprintf("%s/%d", conf.gatewayIPv6.String(), maskLenIPv6),
		DnsmasqEnabled:      dnsmasqEnabled,
	})
}

func (conf *wireguardConfig) configureNAT(protocol iptables.Protocol, network, gateway string) error {
	iptable, err := iptables.NewWithProtocol(protocol)
	if err != nil {
		return err
	}
	if err = iptable.AppendUnique("nat", "POSTROUTING", "-s", network, "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err = iptable.AppendUnique("filter", "FORWARD", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err = iptable.AppendUnique("filter", "FORWARD", "-s", network, "-j", "ACCEPT"); err != nil {
		return err
	}
	// DNS Leak Protection
	if err = iptable.AppendUnique("nat", "OUTPUT", "-s", network, "-p", "udp", "--dport", "53", "-j", "DNAT", "--to", fmt.Sprintf("%s:53", gateway)); err != nil {
		return err
	}
	if err = iptable.AppendUnique("nat", "OUTPUT", "-s", network, "-p", "tcp", "--dport", "53", "-j", "DNAT", "--to", fmt.Sprintf("%s:53", gateway)); err != nil {
		return err
	}
	return nil
}

func (conf *wireguardConfig) generateIPAddr(id uint32) (net.IP, net.IP, error) {
	return generateIPAddr(conf.networkIPv4, conf.networkIPv6, id)
}
