# netkit

[![CI](https://github.com/yinebebt/netkit/actions/workflows/ci.yml/badge.svg)](https://github.com/yinebebt/netkit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/yinebebt/netkit.svg)](https://pkg.go.dev/github.com/yinebebt/netkit)

Toolkit for IP address manipulation, CIDR calculations, and general networking utilities.

## Features

- [x] IP address information lookup (`info`)
- [x] Forward DNS lookup (`resolve`) for domain -> IP
- [x] Reverse DNS lookup (`reverse`) for IP -> hostname (PTR)
- [x] Optional JSON output for machine-friendly automation
- [x] CIDR calculator (`cidr`) for network, netmask, broadcast, host range and size (IPv4/IPv6)
- [ ] Utilities for ports, protocols, and common network operations

## Installation

```shell
go install github.com/yinebebt/netkit@latest
```

## Usage

```bash
Usage:
  netkit [command]

Available Commands:
  cidr        Calculate the address range of a CIDR block
  help        Help about any command
  info        Show information about an IP address
  resolve     Resolve domain name to IP addresses (A/AAAA)
  reverse     Resolve IP address to hostnames (PTR / reverse DNS)

Flags:
  -h, --help   help for netkit

Use "netkit [command] --help" for more information about a command.
```