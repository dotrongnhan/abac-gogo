package constants

// Core context key constants
const (
	ContextKeyRequestUserID     = "request:UserId"
	ContextKeyRequestAction     = "request:Action"
	ContextKeyRequestResourceID = "request:ResourceId"
	ContextKeyRequestTime       = "request:Time"
)

// Context key prefixes
const (
	ContextKeyUserPrefix        = "user:"
	ContextKeyResourcePrefix    = "resource:"
	ContextKeyEnvironmentPrefix = "environment:"
	ContextKeyRequestPrefix     = "request:"
)

// Enhanced context keys for improved features
const (
	ContextKeyClientIP  = "environment:client_ip"
	ContextKeyUserAgent = "environment:user_agent"
	ContextKeyCountry   = "environment:country"
	ContextKeyRegion    = "environment:region"
	ContextKeyTimeOfDay = "environment:time_of_day"
	ContextKeyDayOfWeek = "environment:day_of_week"
)

// Attribute resolver context keys
const (
	// Input context keys
	ContextKeyTimestamp     = "timestamp"
	ContextKeySourceIP      = "source_ip"
	ContextKeyClientIPShort = "client_ip"

	// Computed environment attributes
	ContextKeyTimeOfDayShort  = "time_of_day"
	ContextKeyDayOfWeekShort  = "day_of_week"
	ContextKeyHour            = "hour"
	ContextKeyIsBusinessHours = "is_business_hours"
	ContextKeyIsInternalIP    = "is_internal_ip"
	ContextKeyIPSubnet        = "ip_subnet"

	// Dynamic subject attributes
	ContextKeyYearsOfService = "years_of_service"
	ContextKeyCurrentHour    = "current_hour"
	ContextKeyCurrentDay     = "current_day"

	// Subject attribute keys
	ContextKeyHireDate       = "hire_date"
	ContextKeyDepartment     = "department"
	ContextKeyRole           = "role"
	ContextKeyClearanceLevel = "clearance_level"
)
