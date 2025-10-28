package constants

import "time"

// Business hours configuration
const (
	BusinessHoursStart = 9  // 9 AM
	BusinessHoursEnd   = 17 // 5 PM (17:00)
)

// Business days configuration
const (
	BusinessDayStart = time.Monday
	BusinessDayEnd   = time.Friday
)

// Private IP ranges for internal network detection
var PrivateIPRanges = []string{
	"10.0.0.0/8",     // Class A private network
	"172.16.0.0/12",  // Class B private network
	"192.168.0.0/16", // Class C private network
	"127.0.0.0/8",    // Loopback addresses
}

// Context map sizing constants
const (
	DefaultContextMapSize = 50  // Default size for evaluation context maps
	MaxContextMapSize     = 200 // Maximum allowed context map size
)

// Mobile user agent detection patterns
var MobileUserAgentPatterns = []string{
	"(?i)mobile",
	"(?i)android",
	"(?i)iphone",
	"(?i)ipad",
	"(?i)blackberry",
	"(?i)windows phone",
}

// Browser detection patterns
var BrowserPatterns = map[string]string{
	"(?i)chrome":  "chrome",
	"(?i)firefox": "firefox",
	"(?i)safari":  "safari",
	"(?i)edge":    "edge",
	"(?i)opera":   "opera",
}
