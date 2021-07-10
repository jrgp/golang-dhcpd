package main

//
// DHCP Op types
//
const (
	BOOT_REQUEST = 1
	BOOT_REPLY   = 2
)

//
// DHCP Message types
//
const (
	DHCPDISCOVER = 1 // Implemented
	DHCPOFFER    = 2 // Implemented
	DHCPREQUEST  = 3 // Implemented
	DHCPDECLINE  = 4
	DHCPACK      = 5 // Implemented
	DHCPNAK      = 6 // Implemented
	DHCPRELEASE  = 7 // Implemented
	DHCPINFORM   = 8
)

var opNames = map[byte]string{
	DHCPOFFER: "DHCPOFFER",
	DHCPACK:   "DHCPACK",
	DHCPNAK:   "DHCPNAK",
}

//
// DHCP Option numbers. List partly taken from udhcpc
//

const (
	OPTION_PADDING       = 0
	OPTION_SUBNET        = 1
	OPTION_TIME_OFFSET   = 2
	OPTION_ROUTER        = 3
	OPTION_TIME_SERVER   = 4
	OPTION_NAME_SERVER   = 5
	OPTION_DNS_SERVER    = 6
	OPTION_LOG_SERVER    = 7
	OPTION_COOKIE_SERVER = 8
	OPTION_LPR_SERVER    = 9
	OPTION_HOST_NAME     = 12
	OPTION_BOOT_SIZE     = 13
	OPTION_DOMAIN_NAME   = 15
	OPTION_SWAP_SERVER   = 16
	OPTION_ROOT_PATH     = 17
	OPTION_IP_TTL        = 23
	OPTION_MTU           = 26
	OPTION_BROADCAST     = 28
	OPTION_NTP_SERVER    = 42
	OPTION_WINS_SERVER   = 44
	OPTION_REQUESTED_IP  = 50
	OPTION_LEASE_TIME    = 51
	OPTION_OPTION_OVER   = 52
	OPTION_MESSAGE_TYPE  = 53
	OPTION_SERVER_ID     = 54
	OPTION_PARAM_REQ     = 55
	OPTION_MESSAGE       = 56
	OPTION_MAX_SIZE      = 57
	OPTION_T1            = 58
	OPTION_T2            = 59
	OPTION_VENDOR        = 60
	OPTION_CLIENT_ID     = 61
	OPTION_SENTINEL      = 255
)
