## What

This is a tiny IPv4-only dhcp server written in Golang.

## Why

Other popular dhcpd servers such as isc-dhcpd and dnsmasq are written in C and occasionally
have CVEs.

I wanted to make a small and safe Go solution, as well as to learn the [DHCP wire protocol](https://en.wikipedia.org/wiki/Dynamic_Host_Configuration_Protocol).

## Quickstart

1. Clone repo
2. go build
3. Configure conf.yaml
3. Run with permissions needed to listen on port 67 (eg run as root or use linux capabilities), as follows
4. Run a separate VM on the same bridge/vlan as a dhcp client

### Configuration

We use yaml. Multiple pools can be defined this way. DHCP traffic to interfaces not listed will be ignored.

```yaml
pools:
  - name: vm testing
    network: 172.17.0.0
    mask: 255.255.255.0
    start: 172.17.0.100
    end: 172.17.0.200
    leasetime: 60
    myip: 172.17.0.1
    routers: [ 172.17.0.1 ]
    dns: [ 1.1.1.1, 8.8.8.8 ]

    # Optional static IPs by mac address
    hosts:
      - ip: 172.17.0.5
        hw: 0:1c:42:b4:6e:1d

interfaces: [ eth1 ]
leasedir: /var/lib/godhcpd
```

### Example command output on VM acting as DHCP server

```
root@ubuntu1:~/dev/golang-dhcpd# go build
root@ubuntu1:~/dev/golang-dhcpd# ./mygodhcpd -conf conf.yaml
2021/07/05 21:36:58 Loaded pool vm testing on interface eth1
2021/07/05 21:37:18 DHCPREQUEST from 0:1c:42:b4:6e:1d for 172.17.0.100
2021/07/05 21:37:18 Unrecognized lease for 0:1c:42:b4:6e:1d
2021/07/05 21:37:18 Sending DHCPNAK to 0:1c:42:b4:6e:1d
2021/07/05 21:37:18 DHCPDISCOVER from 0:1c:42:b4:6e:1d (ubuntu2)
2021/07/05 21:37:18 Sending DHCPOFFER with 172.17.0.100 to 0:1c:42:b4:6e:1d
2021/07/05 21:37:18 DHCPREQUEST from 0:1c:42:b4:6e:1d for 172.17.0.100
2021/07/05 21:37:18 Sending DHCPACK with 172.17.0.100 to 0:1c:42:b4:6e:1d
```

### And in the other VM:

```
root@ubuntu2:~# dhclient eth1
root@ubuntu2:~# ip -4 a show eth1
3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    inet 172.17.0.100/24 brd 172.17.0.255 scope global dynamic eth1
       valid_lft 54sec preferred_lft 54sec
root@ubuntu2:~#
```

## Goals

- Be small
- Be fast
- Be configurable
- Don't require anything outside of the Go standard library (except maybe testify)

## Status

- Verified to work with Alpine's `udhcpc` client, Ubuntu's `dhclient` client, and Windows 10.
- Relay requests verified to work with isc-dhcp-relay.

## Implemented

- Bare minimum wire protocol for DHCPDISCOVER, DHCPOFFER, DHCPREQUEST, DHCPNAK, DHCPACK, and DHCPRELEASE to work
- Supports relayed requests
- Supports multiple IP Pools, sourced from configuration
- Supports hosts in config with hardcoded IPs, based on mac address

## TODO

- Support acting as a relay
- Support arbitrary options, including options scoped to specific hosts
- PXE with usage examples
- Example systemd unit, deb/rpm packages, etc
- More Tests
