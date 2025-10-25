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
