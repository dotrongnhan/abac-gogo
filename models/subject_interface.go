package models

// SubjectType represents the type of subject in the ABAC system
type SubjectType string

const (
	// SubjectTypeUser represents a human user
	SubjectTypeUser SubjectType = "user"
	// SubjectTypeService represents a service account or application
	SubjectTypeService SubjectType = "service"
	// SubjectTypeAPIKey represents an API key authentication
	SubjectTypeAPIKey SubjectType = "api_key"
	// SubjectTypeLegacy represents legacy subject from subjects table (for backward compatibility)
	SubjectTypeLegacy SubjectType = "legacy"
)

// SubjectInterface defines the contract for all subject types in ABAC
// This abstraction allows different subject implementations (User, Service, APIKey)
// while maintaining a consistent interface for policy evaluation
type SubjectInterface interface {
	// GetID returns the unique identifier of the subject
	GetID() string

	// GetType returns the type of subject (user, service, api_key, etc.)
	GetType() SubjectType

	// GetAttributes returns all ABAC attributes as a flat map
	// This is the primary method used by the PDP for policy evaluation
	GetAttributes() map[string]interface{}

	// GetDisplayName returns a human-readable name for the subject
	GetDisplayName() string

	// IsActive returns whether the subject is currently active
	IsActive() bool
}

// AttributeMapper defines how subjects map their internal structure to ABAC attributes
type AttributeMapper interface {
	// MapToAttributes converts internal subject data to flat ABAC attributes
	MapToAttributes() map[string]interface{}
}
