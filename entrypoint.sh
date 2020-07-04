#!/usr/bin/env sh
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Require environment variables.
if [ -z "${SUBSPACE_HTTP_HOST-}" ]; then
  echo "Environment variable SUBSPACE_HTTP_HOST required. Exiting."
  exit 1
fi
# Optional environment variables.
if [ -z "${SUBSPACE_BACKLINK-}" ]; then
  export SUBSPACE_BACKLINK="/"
fi

if [ -z "${SUBSPACE_NETWORK_IPV4-}" ]; then
  export SUBSPACE_NETWORK_IPV4="10.99.97.0/24"
fi

if [ -z "${SUBSPACE_NETWORK_IPV4-}" ]; then
  export SUBSPACE_NETWORK_IPV6="fd00::10:97:0/64"
fi

if [ -z "${SUBSPACE_NAMESERVER-}" ]; then
  export SUBSPACE_NAMESERVER="1.1.1.1"
fi

if [ -z "${SUBSPACE_LETSENCRYPT-}" ]; then
  export SUBSPACE_LETSENCRYPT="true"
fi

if [ -z "${SUBSPACE_HTTP_ADDR-}" ]; then
  export SUBSPACE_HTTP_ADDR=":80"
fi

if [ -z "${SUBSPACE_LISTENPORT-}" ]; then
  export SUBSPACE_LISTENPORT="51820"
fi

if [ -z "${SUBSPACE_HTTP_INSECURE-}" ]; then
  export SUBSPACE_HTTP_INSECURE="false"
fi

if [ -z "${SUBSPACE_THEME-}" ]; then
  export SUBSPACE_THEME="green"
fi

export DEBIAN_FRONTEND="noninteractive"

if [ -z "${SUBSPACE_IPV6_NAT_ENABLED-}" ]; then
  export SUBSPACE_IPV6_NAT_ENABLED=1
fi

#
# WireGuard (${SUBSPACE_NETWORK_IPV4})
#
if ! test -d /data/wireguard; then
  mkdir /data/wireguard
  cd /data/wireguard

  mkdir clients
  touch clients/null.conf # So you can cat *.conf safely
  mkdir peers
  touch peers/null.conf # So you can cat *.conf safely

  # Generate public/private server keys.
  wg genkey | tee server.private | wg pubkey > server.public
fi

cat <<WGSERVER >/data/wireguard/server.conf
[Interface]
PrivateKey = $(cat /data/wireguard/server.private)
ListenPort = ${SUBSPACE_LISTENPORT}

WGSERVER
cat /data/wireguard/peers/*.conf >>/data/wireguard/server.conf

# dnsmasq service
if ! test -d /etc/service/dnsmasq; then
  cat <<DNSMASQ >/etc/dnsmasq.conf
    # Only listen on necessary addresses.
    listen-address=127.0.0.1,${SUBSPACE_IPV4_GW},${SUBSPACE_IPV6_GW}

    # Never forward plain names (without a dot or domain part)
    domain-needed

    # Never forward addresses in the non-routed address spaces.
    bogus-priv
DNSMASQ

  mkdir -p /etc/service/dnsmasq
  cat <<RUNIT >/etc/service/dnsmasq/run
#!/bin/sh
exec /usr/sbin/dnsmasq --no-daemon
RUNIT
  chmod +x /etc/service/dnsmasq/run

  # dnsmasq service log
  mkdir -p /etc/service/dnsmasq/log/main
  cat <<RUNIT >/etc/service/dnsmasq/log/run
#!/bin/sh
exec svlogd -tt ./main
RUNIT
  chmod +x /etc/service/dnsmasq/log/run
fi

# subspace service
if ! test -d /etc/service/subspace; then
  mkdir /etc/service/subspace
  cat <<RUNIT >/etc/service/subspace/run
#!/bin/sh
source /etc/envvars
exec /usr/bin/subspace \
    "--http-host=${SUBSPACE_HTTP_HOST}" \
    "--http-addr=${SUBSPACE_HTTP_ADDR}" \
    "--http-insecure=${SUBSPACE_HTTP_INSECURE}" \
    "--backlink=${SUBSPACE_BACKLINK}" \
    "--letsencrypt=${SUBSPACE_LETSENCRYPT}" \
    "--theme=${SUBSPACE_THEME}"
    "--setup-network=true"
    "--nameserver=${SUBSPACE_NAMESERVER}"
    "--listen-port=${SUBSPACE_LISTENPORT}"
    "--network-ipv4=${SUBSPACE_NETWORK_IPV4}"
    "--network-ipv6=${SUBSPACE_NETWORK_IPV6}"
    "--enable-ipv6-nat=${SUBSPACE_IPV6_NAT_ENABLED}"
RUNIT
  chmod +x /etc/service/subspace/run

  # subspace service log
  mkdir /etc/service/subspace/log
  mkdir /etc/service/subspace/log/main
  cat <<RUNIT >/etc/service/subspace/log/run
#!/bin/sh
exec svlogd -tt ./main
RUNIT
  chmod +x /etc/service/subspace/log/run
fi

exec "$@"
