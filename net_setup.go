package main

import (
	"fmt"
	"net"
)

var (
	wireguardConfig struct {
		gatewayIPv4 net.IP
		gatewayIPv6 net.IP
		networkIPv4 *net.IPNet
		networkIPv6 *net.IPNet
	}
)

func initWireguardConfig() error {
	var err error
	wireguardConfig.gatewayIPv4, wireguardConfig.networkIPv4, err = calcDefaultGateway(networkIPv4)
	if err != nil {
		return err
	}
	wireguardConfig.gatewayIPv6, wireguardConfig.networkIPv6, err = calcDefaultGateway(networkIPv6)
	if err != nil {
		return err
	}
	return nil
}

func configureWireguard() (string, error) {

	type value struct {
		Nameserver string
		NetworkIPv4 string
		IPv6NatEnabled bool
		NetworkIPv6 string
		GatewayIPv4 string
		GatewayIPv6 string
		GatewayIPv4WithCIDR string
		GatewayIPv6WithCIDR string
		DnsmasqEnabled bool
	}
	maskLenIPv4, _ := wireguardConfig.networkIPv4.Mask.Size()
	maskLenIPv6, _ := wireguardConfig.networkIPv6.Mask.Size()
	v := value{
		Nameserver:          nameserver,
		NetworkIPv4:         networkIPv4,
		IPv6NatEnabled:      ipv6NatEnabled,
		NetworkIPv6:         networkIPv6,
		GatewayIPv4:         wireguardConfig.gatewayIPv4.String(),
		GatewayIPv6:         wireguardConfig.gatewayIPv6.String(),
		GatewayIPv4WithCIDR: fmt.Sprintf("%s/%d", wireguardConfig.gatewayIPv4.String(), maskLenIPv4),
		GatewayIPv6WithCIDR: fmt.Sprintf("%s/%d", wireguardConfig.gatewayIPv6.String(), maskLenIPv6),
		DnsmasqEnabled: enableDnsmasq,
	}
	return bash(`
# Set DNS server
echo "nameserver {{.Nameserver}}" > /etc/resolv.conf

# Setup NAT

## IPv4
if ! /sbin/iptables -t nat --check POSTROUTING -s "{{.NetworkIPv4}}" -j MASQUERADE; then
  /sbin/iptables -t nat --append POSTROUTING -s "{{.NetworkIPv4}}" -j MASQUERADE
fi
if ! /sbin/iptables --check FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT; then
  /sbin/iptables --append FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT
fi
if ! /sbin/iptables --check FORWARD -s "{{.NetworkIPv4}}" -j ACCEPT; then
  /sbin/iptables --append FORWARD -s "{{.NetworkIPv4}}" -j ACCEPT
fi

{{if .IPv6NatEnabled}}
## IPv6
if ! /sbin/ip6tables -t nat --check POSTROUTING -s "{{.NetworkIPv6}}" -j MASQUERADE; then
  /sbin/ip6tables -t nat --append POSTROUTING -s "{{.NetworkIPv6}}" -j MASQUERADE
fi
if ! /sbin/ip6tables --check FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT; then
  /sbin/ip6tables --append FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT
fi
if ! /sbin/ip6tables --check FORWARD -s "{{.NetworkIPv6}}" -j ACCEPT; then
  /sbin/ip6tables --append FORWARD -s "{{.NetworkIPv6}}" -j ACCEPT
fi
{{end}}

# ipv4 - DNS Leak Protection
if ! /sbin/iptables -t nat --check OUTPUT -s "{{.NetworkIPv4}}" -p udp --dport 53 -j DNAT --to "{{.GwAddressIPv4}}:53"; then
  /sbin/iptables -t nat --append OUTPUT -s "{{.NetworkIPv4}}" -p udp --dport 53 -j DNAT --to "{{.GwAddressIPv4}}:53"
fi
if ! /sbin/iptables -t nat --check OUTPUT -s "{{.NetworkIPv4}}" -p tcp --dport 53 -j DNAT --to "{{.GwAddressIPv4}}:53"; then
  /sbin/iptables -t nat --append OUTPUT -s "{{.NetworkIPv4}}" -p tcp --dport 53 -j DNAT --to "{{.GwAddressIPv4}}:53"
fi

# ipv6 - DNS Leak Protection
if ! /sbin/ip6tables --wait -t nat --check OUTPUT -s "{{.NetworkIPv6}}" -p udp --dport 53 -j DNAT --to "{{.GwAddressIPv6}}"; then
  /sbin/ip6tables --wait -t nat --append OUTPUT -s "{{.NetworkIPv6}}" -p udp --dport 53 -j DNAT --to "{{.GwAddressIPv6}}"
fi
if ! /sbin/ip6tables --wait -t nat --check OUTPUT -s "{{.NetworkIPv6}}" -p tcp --dport 53 -j DNAT --to "{{.GwAddressIPv6}}"; then
  /sbin/ip6tables --wait -t nat --append OUTPUT -s "{{.NetworkIPv6}}" -p tcp --dport 53 -j DNAT --to "{{.GwAddressIPv6}}"
fi

# Create wireguard device
if ip link show wg0 2>/dev/null; then
  ip link del wg0
fi
ip link add wg0 type wireguard
ip addr add "{{.GatewayIPv4WithCIDR}}" dev wg0
ip addr add "{{.GatewayIPv6WithCIDR}}" dev wg0
wg setconf wg0 /data/wireguard/server.conf
ip link set wg0 up

{{if .DnsmasqEnabled}}
# Reload dnsmasq
sed -i -e 's/listen-address=.\+$/listen-address=127.0.0.1,{{.GatewayIPv4}},{{.GatewayIPv6}}/g' /etc/dnsmasq.conf
sv restart dnsmasq
{{end}}
`, &v)
}