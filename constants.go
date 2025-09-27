package main

// DHCP Op types
const (
	BOOT_REQUEST byte = 1
	BOOT_REPLY   byte = 2
)

// DHCP Message types
const (
	DHCPDISCOVER byte = 1 // Implemented
	DHCPOFFER    byte = 2 // Implemented
	DHCPREQUEST  byte = 3 // Implemented
	DHCPDECLINE  byte = 4
	DHCPACK      byte = 5 // Implemented
	DHCPNAK      byte = 6 // Implemented
	DHCPRELEASE  byte = 7 // Implemented
	DHCPINFORM   byte = 8
)

var messageNames = map[byte]string{
	DHCPOFFER: "DHCPOFFER",
	DHCPACK:   "DHCPACK",
	DHCPNAK:   "DHCPNAK",
}

//
// DHCP Option numbers. List partly taken from udhcpc
//

const (
	OPTION_PADDING       byte = 0
	OPTION_SUBNET        byte = 1
	OPTION_TIME_OFFSET   byte = 2
	OPTION_ROUTER        byte = 3
	OPTION_TIME_SERVER   byte = 4
	OPTION_NAME_SERVER   byte = 5
	OPTION_DNS_SERVER    byte = 6
	OPTION_LOG_SERVER    byte = 7
	OPTION_COOKIE_SERVER byte = 8
	OPTION_LPR_SERVER    byte = 9
	OPTION_HOST_NAME     byte = 12
	OPTION_BOOT_SIZE     byte = 13
	OPTION_DOMAIN_NAME   byte = 15
	OPTION_SWAP_SERVER   byte = 16
	OPTION_ROOT_PATH     byte = 17
	OPTION_IP_TTL        byte = 23
	OPTION_MTU           byte = 26
	OPTION_BROADCAST     byte = 28
	OPTION_NTP_SERVER    byte = 42
	OPTION_WINS_SERVER   byte = 44
	OPTION_REQUESTED_IP  byte = 50
	OPTION_LEASE_TIME    byte = 51
	OPTION_OPTION_OVER   byte = 52
	OPTION_MESSAGE_TYPE  byte = 53
	OPTION_SERVER_ID     byte = 54
	OPTION_PARAM_REQ     byte = 55
	OPTION_MESSAGE       byte = 56
	OPTION_MAX_SIZE      byte = 57
	OPTION_T1            byte = 58
	OPTION_T2            byte = 59
	OPTION_VENDOR        byte = 60
	OPTION_CLIENT_ID     byte = 61
	OPTION_DNS_SEARCH    byte = 119
	OPTION_STATIC_ROUTES byte = 121
	OPTION_SENTINEL      byte = 255
)

var nameToOption = map[string]byte{
	"padding":       OPTION_PADDING,
	"subnet":        OPTION_SUBNET,
	"time_offset":   OPTION_TIME_OFFSET,
	"router":        OPTION_ROUTER,
	"time_server":   OPTION_TIME_SERVER,
	"name_server":   OPTION_NAME_SERVER,
	"dns_server":    OPTION_DNS_SERVER,
	"log_server":    OPTION_LOG_SERVER,
	"cookie_server": OPTION_COOKIE_SERVER,
	"lpr_server":    OPTION_LPR_SERVER,
	"host_name":     OPTION_HOST_NAME,
	"boot_size":     OPTION_BOOT_SIZE,
	"domain_name":   OPTION_DOMAIN_NAME,
	"swap_server":   OPTION_SWAP_SERVER,
	"root_path":     OPTION_ROOT_PATH,
	"ip_ttl":        OPTION_IP_TTL,
	"mtu":           OPTION_MTU,
	"broadcast":     OPTION_BROADCAST,
	"ntp_server":    OPTION_NTP_SERVER,
	"wins_server":   OPTION_WINS_SERVER,
	"requested_ip":  OPTION_REQUESTED_IP,
	"dns_search":    OPTION_DNS_SEARCH,
	"lease_time":    OPTION_LEASE_TIME,
	"option_over":   OPTION_OPTION_OVER,
	"message_type":  OPTION_MESSAGE_TYPE,
	"server_id":     OPTION_SERVER_ID,
	"param_req":     OPTION_PARAM_REQ,
	"message":       OPTION_MESSAGE,
	"max_size":      OPTION_MAX_SIZE,
	"vendor":        OPTION_VENDOR,
	"client_id":     OPTION_CLIENT_ID,
	"static_routes": OPTION_STATIC_ROUTES,
	"sentinel":      OPTION_SENTINEL,
}

// optionNames is automatically generated from nameToOption
var optionNames = func() map[byte]string {
	result := make(map[byte]string)
	for name, option := range nameToOption {
		result[option] = name
	}
	return result
}()
