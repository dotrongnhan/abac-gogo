package operators

import (
	"net"
	"regexp"

	"abac_go_example/constants"
)

// NetworkUtils provides network-related utility functions
type NetworkUtils struct{}

// NewNetworkUtils creates a new NetworkUtils instance
func NewNetworkUtils() *NetworkUtils {
	return &NetworkUtils{}
}

// IsInternalIP checks if an IP address is internal/private
func (nu *NetworkUtils) IsInternalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	return nu.IsInternalIPAddress(ip)
}

// IsInternalIPAddress checks if a parsed IP address is internal/private
func (nu *NetworkUtils) IsInternalIPAddress(ip net.IP) bool {
	for _, rangeStr := range constants.PrivateIPRanges {
		_, cidr, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// GetIPClass returns the class of IP address (ipv4/ipv6/invalid)
func (nu *NetworkUtils) GetIPClass(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "invalid"
	}

	if ip.To4() != nil {
		return "ipv4"
	}
	return "ipv6"
}

// IsMobileUserAgent detects if user agent is from mobile device
func (nu *NetworkUtils) IsMobileUserAgent(userAgent string) bool {
	for _, pattern := range constants.MobileUserAgentPatterns {
		matched, _ := regexp.MatchString(pattern, userAgent)
		if matched {
			return true
		}
	}
	return false
}

// GetBrowserFromUserAgent extracts browser name from user agent
func (nu *NetworkUtils) GetBrowserFromUserAgent(userAgent string) string {
	for pattern, browser := range constants.BrowserPatterns {
		matched, _ := regexp.MatchString(pattern, userAgent)
		if matched {
			return browser
		}
	}
	return "unknown"
}

// IsBusinessHours checks if the given hour and weekday are within business hours
func (nu *NetworkUtils) IsBusinessHours(hour int, weekday int) bool {
	return hour >= constants.BusinessHoursStart &&
		hour < constants.BusinessHoursEnd &&
		weekday >= int(constants.BusinessDayStart) &&
		weekday <= int(constants.BusinessDayEnd)
}
