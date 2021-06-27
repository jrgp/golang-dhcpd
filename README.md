## What

This is a tiny in-progress POC IPv4-only dhcp server written in Golang

## Why

Other popular dhcpd servers such as isc-dhcpd and dnsmasq are written in C and occasionally
have CVEs.

I wanted to make a small and safe Go solution, as well as to learn the DHCP protocol.

## Quickstart

1. Clone repo
2. As this is unfinished, Hardcode IP ranges into main.go and connection.go's sendBroadcast
3. go build
4. Run with permissions needed to listen on port 67 (eg run as root or use linux capabilities), as follows
5. Run a separate VM on the same bridge/vlan as a dhcp client. This is tested against udhcpc

```
joe@vm-devuan:~/dev/mydhcpd$ sudo ./mygodhcpd
2021/06/26 22:56:55 DHCPDISCOVER from 52:54:0:7a:a6:6e
2021/06/26 22:56:55 Got a new lease for 52:54:0:7a:a6:6e: 172.17.0.100
2021/06/26 22:56:55 Sending DHCPOFFER with 172.17.0.100 to 52:54:0:7a:a6:6e
2021/06/26 22:56:55 DHCPREQUEST from 52:54:0:7a:a6:6e
2021/06/26 22:56:55 Sending DHCPACK with 172.17.0.100 to 52:54:0:7a:a6:6e
```

## Goals

- Be small
- Be fast
- Be configurable
- Don't require anything outside of the Go standard library (except maybe testify)

## Status

Verified to work with Alpine's `udhcpc` client.

## Implemented

- Bare minimum wire protocol for DHCPDISCOVER, DHCPOFFER, DHCPREQUEST, and DHCPACK to work
- IP Pools

## TODO

- On disk configuration format, rather than hard coding config in main.go
- Persist saved leases to disk, to survive restarts
- Support listening on specific interfaces
- Support pools per interface, and more than one pool
- Tests
