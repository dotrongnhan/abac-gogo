package evaluator

import (
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"abac_go_example/models"
)

// TimeProvider interface for time operations (allows mocking in tests)
type TimeProvider interface {
	Now(timezone string) (time.Time, error)
}

// LocationProvider interface for location operations
type LocationProvider interface {
	GetLocationFromIP(ip string) (*models.LocationInfo, error)
	CalculateDistance(lat1, lon1, lat2, lon2 float64) float64
}

// RealTimeProvider implements TimeProvider using real system time
type RealTimeProvider struct{}

func (rtp *RealTimeProvider) Now(timezone string) (time.Time, error) {
	if timezone == "" {
		return time.Now(), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", timezone)
	}

	return time.Now().In(loc), nil
}

// RealLocationProvider implements LocationProvider with basic functionality
type RealLocationProvider struct{}

func (rlp *RealLocationProvider) GetLocationFromIP(ip string) (*models.LocationInfo, error) {
	// This is a placeholder implementation
	// In a real system, you would integrate with a GeoIP service
	return &models.LocationInfo{
		Country: "Unknown",
		Region:  "Unknown",
		City:    "Unknown",
	}, nil
}

func (rlp *RealLocationProvider) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Haversine formula to calculate distance between two points on Earth
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// EnhancedConditionEvaluator handles advanced condition evaluation
type EnhancedConditionEvaluator struct {
	timeProvider     TimeProvider
	locationProvider LocationProvider
}

// NewEnhancedConditionEvaluator creates a new enhanced condition evaluator
func NewEnhancedConditionEvaluator() *EnhancedConditionEvaluator {
	return &EnhancedConditionEvaluator{
		timeProvider:     &RealTimeProvider{},
		locationProvider: &RealLocationProvider{},
	}
}

// NewEnhancedConditionEvaluatorWithProviders creates evaluator with custom providers
func NewEnhancedConditionEvaluatorWithProviders(timeProvider TimeProvider, locationProvider LocationProvider) *EnhancedConditionEvaluator {
	return &EnhancedConditionEvaluator{
		timeProvider:     timeProvider,
		locationProvider: locationProvider,
	}
}

// EvaluateTimeWindow evaluates time-based access control
func (ece *EnhancedConditionEvaluator) EvaluateTimeWindow(tw models.TimeWindow, env *models.Environment) bool {
	now, err := ece.timeProvider.Now(tw.Timezone)
	if err != nil {
		return false
	}

	// Check day of week
	currentDay := strings.ToLower(now.Weekday().String())
	if !contains(tw.DaysOfWeek, currentDay) {
		return false
	}

	// Check excluded dates
	dateStr := now.Format("2006-01-02")
	if contains(tw.ExcludeDates, dateStr) {
		return false
	}

	// Check time range
	return ece.isTimeInRange(now, tw.StartTime, tw.EndTime)
}

// EvaluateLocation evaluates location-based access control
func (ece *EnhancedConditionEvaluator) EvaluateLocation(loc *models.LocationCondition, env *models.Environment) bool {
	if loc == nil {
		return true
	}

	// IP-based location check
	if len(loc.IPRanges) > 0 {
		clientIP := env.GetClientIP()
		if !ece.isIPInRanges(clientIP, loc.IPRanges) {
			return false
		}
	}

	// Country-based check
	if len(loc.AllowedCountries) > 0 && env.Location != nil {
		if !contains(loc.AllowedCountries, env.Location.Country) {
			return false
		}
	}

	// Region-based check
	if len(loc.AllowedRegions) > 0 && env.Location != nil {
		if !contains(loc.AllowedRegions, env.Location.Region) {
			return false
		}
	}

	// Geographic fencing
	if loc.GeoFencing != nil {
		userLat := env.GetLatitude()
		userLng := env.GetLongitude()

		if userLat == 0 && userLng == 0 {
			return false // No location data available
		}

		distance := ece.locationProvider.CalculateDistance(
			userLat, userLng,
			loc.GeoFencing.Latitude,
			loc.GeoFencing.Longitude,
		)

		if distance > loc.GeoFencing.Radius {
			return false
		}
	}

	return true
}

// EvaluateTimeWindows evaluates multiple time windows (OR logic)
func (ece *EnhancedConditionEvaluator) EvaluateTimeWindows(timeWindows []models.TimeWindow, env *models.Environment) bool {
	if len(timeWindows) == 0 {
		return true // No time restrictions
	}

	// At least one time window must match
	for _, tw := range timeWindows {
		if ece.EvaluateTimeWindow(tw, env) {
			return true
		}
	}

	return false
}

// isTimeInRange checks if current time is within the specified range
func (ece *EnhancedConditionEvaluator) isTimeInRange(now time.Time, startTime, endTime string) bool {
	currentTime := now.Format("15:04")

	// Parse time strings
	start, err1 := time.Parse("15:04", startTime)
	end, err2 := time.Parse("15:04", endTime)
	current, err3 := time.Parse("15:04", currentTime)

	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}

	// Convert to minutes for easier comparison
	startMinutes := start.Hour()*60 + start.Minute()
	endMinutes := end.Hour()*60 + end.Minute()
	currentMinutes := current.Hour()*60 + current.Minute()

	// Handle overnight ranges (e.g., 22:00 to 06:00)
	if startMinutes > endMinutes {
		return currentMinutes >= startMinutes || currentMinutes <= endMinutes
	}

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

// isIPInRanges checks if IP is in any of the allowed ranges
func (ece *EnhancedConditionEvaluator) isIPInRanges(ip string, ranges []string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	for _, cidr := range ranges {
		if ece.isIPInCIDR(ip, cidr) {
			return true
		}
	}

	return false
}

// isIPInCIDR checks if IP is in CIDR range
func (ece *EnhancedConditionEvaluator) isIPInCIDR(ip, cidr string) bool {
	// Parse IP
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Maybe it's a single IP
		if singleIP := net.ParseIP(cidr); singleIP != nil {
			return ipAddr.Equal(singleIP)
		}
		return false
	}

	return ipNet.Contains(ipAddr)
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// EvaluateEnvironmentalConditions evaluates all environmental conditions
func (ece *EnhancedConditionEvaluator) EvaluateEnvironmentalConditions(rule *models.PolicyRule, env *models.Environment) bool {
	// Evaluate time windows
	if !ece.EvaluateTimeWindows(rule.TimeWindows, env) {
		return false
	}

	// Evaluate location conditions
	if !ece.EvaluateLocation(rule.Location, env) {
		return false
	}

	return true
}

// EvaluateComplexCondition evaluates a complex condition with environmental context
func (ece *EnhancedConditionEvaluator) EvaluateComplexCondition(condition map[string]interface{}, env *models.Environment, context map[string]interface{}) bool {
	// Add environmental attributes to context
	enrichedContext := ece.enrichContextWithEnvironment(context, env)

	// Use the existing condition evaluator for basic conditions
	basicEvaluator := NewConditionEvaluator()
	return basicEvaluator.Evaluate(condition, enrichedContext)
}

// enrichContextWithEnvironment adds environmental data to evaluation context
func (ece *EnhancedConditionEvaluator) enrichContextWithEnvironment(context map[string]interface{}, env *models.Environment) map[string]interface{} {
	enriched := make(map[string]interface{})

	// Copy existing context
	for k, v := range context {
		enriched[k] = v
	}

	if env == nil {
		return enriched
	}

	// Add environmental attributes
	enriched["environment:timestamp"] = env.Timestamp.Format(time.RFC3339)
	enriched["environment:time_of_day"] = env.Timestamp.Format("15:04")
	enriched["environment:day_of_week"] = strings.ToLower(env.Timestamp.Weekday().String())
	enriched["environment:client_ip"] = env.ClientIP
	enriched["environment:user_agent"] = env.UserAgent

	if env.Location != nil {
		enriched["environment:country"] = env.Location.Country
		enriched["environment:region"] = env.Location.Region
		enriched["environment:city"] = env.Location.City
		enriched["environment:latitude"] = env.Location.Latitude
		enriched["environment:longitude"] = env.Location.Longitude
	}

	// Add custom environmental attributes
	for k, v := range env.Attributes {
		enriched["environment:"+k] = v
	}

	return enriched
}
