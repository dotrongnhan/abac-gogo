package constants

// Policy effect constants
const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// Decision result constants
const (
	ResultPermit = "permit"
	ResultDeny   = "deny"
)

// Decision reason templates
const (
	ReasonDeniedByStatement   = "Denied by statement: %s"
	ReasonAllowedByStatements = "Allowed by statements: %s"
	ReasonImplicitDeny        = "No matching policies found (implicit deny)"
)

// Validation and performance constants
const (
	MaxConditionDepth      = 10   // Maximum depth for nested conditions
	MaxConditionKeys       = 100  // Maximum number of condition keys
	MaxEvaluationTimeMs    = 5000 // Maximum evaluation time in milliseconds
	MinRequiredContextKeys = 3    // Minimum required context keys (action, resource, subject)
)
