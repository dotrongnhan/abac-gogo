package pep

import "time"

// PEPConfig holds basic configuration for SimplePEP
type PEPConfig struct {
	// Security settings
	FailSafeMode     bool `json:"fail_safe_mode"`    // Default to DENY on errors
	StrictValidation bool `json:"strict_validation"` // Strict input validation
	AuditEnabled     bool `json:"audit_enabled"`     // Enable audit logging

	// Performance settings
	EvaluationTimeout time.Duration `json:"evaluation_timeout"`
}

// DefaultPEPConfig returns default configuration for SimplePEP
func DefaultPEPConfig() *PEPConfig {
	return &PEPConfig{
		FailSafeMode:      true,
		StrictValidation:  true,
		AuditEnabled:      true,
		EvaluationTimeout: time.Millisecond * 100,
	}
}

// EnforcementResult represents the result of policy enforcement
type EnforcementResult struct {
	Allowed          bool                   `json:"allowed"`
	Decision         string                 `json:"decision"` // "permit" or "deny"
	Reason           string                 `json:"reason"`
	MatchedPolicies  []string               `json:"matched_policies,omitempty"`
	EvaluationTime   time.Duration          `json:"evaluation_time"`
	EvaluationTimeMs int                    `json:"evaluation_time_ms"`
	CacheHit         bool                   `json:"cache_hit"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
