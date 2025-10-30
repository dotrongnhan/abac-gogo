package models

import (
	"strings"
)

const (
	defaultServiceAccessLevel = 10
)

// ServiceSubject implements SubjectInterface for service-to-service authentication
// This is used for API keys, service accounts, and application-level access
type ServiceSubject struct {
	ServiceID   string
	ServiceName string
	Namespace   string
	Scopes      []string
	Environment string
	Metadata    map[string]interface{}
	Status      string
}

// NewServiceSubject creates a new ServiceSubject instance
func NewServiceSubject(serviceID, serviceName, namespace string) *ServiceSubject {
	return &ServiceSubject{
		ServiceID:   serviceID,
		ServiceName: serviceName,
		Namespace:   namespace,
		Scopes:      []string{},
		Metadata:    make(map[string]interface{}),
		Status:      "active",
	}
}

// GetID returns the service's unique identifier
func (ss *ServiceSubject) GetID() string {
	return ss.ServiceID
}

// GetType returns the subject type as "service"
func (ss *ServiceSubject) GetType() SubjectType {
	return SubjectTypeService
}

// GetDisplayName returns the service name
func (ss *ServiceSubject) GetDisplayName() string {
	if ss.ServiceName != "" {
		return ss.ServiceName
	}
	return ss.ServiceID
}

// IsActive returns whether the service is currently active
func (ss *ServiceSubject) IsActive() bool {
	return strings.ToLower(ss.Status) == "active"
}

// GetAttributes returns all ABAC attributes as a flat map
func (ss *ServiceSubject) GetAttributes() map[string]interface{} {
	return ss.MapToAttributes()
}

// MapToAttributes implements AttributeMapper interface
// Converts service data into flat ABAC attributes
func (ss *ServiceSubject) MapToAttributes() map[string]interface{} {
	attributes := make(map[string]interface{}, maxAttributeMapSize)

	// Core service attributes
	attributes["service_id"] = ss.ServiceID
	attributes["service_name"] = ss.ServiceName
	attributes["subject_type"] = string(SubjectTypeService)
	attributes["status"] = ss.Status

	// Namespace for multi-tenancy
	if ss.Namespace != "" {
		attributes["namespace"] = ss.Namespace
	}

	// Environment (production, staging, development)
	if ss.Environment != "" {
		attributes["environment"] = ss.Environment
		attributes["is_production"] = strings.ToLower(ss.Environment) == "production"
	}

	// Scopes as permissions
	if len(ss.Scopes) > 0 {
		attributes["scopes"] = ss.Scopes
		attributes["scope_count"] = len(ss.Scopes)

		// Add individual scope flags for easier policy writing
		for _, scope := range ss.Scopes {
			attributes["has_scope_"+scope] = true
		}
	} else {
		attributes["scopes"] = []string{}
		attributes["scope_count"] = 0
	}

	// Default access level for services
	attributes["access_level"] = defaultServiceAccessLevel

	// Service-specific flags
	attributes["is_service"] = true
	attributes["is_user"] = false

	// Add custom metadata
	for key, value := range ss.Metadata {
		attributes["metadata_"+key] = value
	}

	return attributes
}

// HasScope checks if the service has a specific scope
func (ss *ServiceSubject) HasScope(scope string) bool {
	scopeLower := strings.ToLower(scope)
	for _, s := range ss.Scopes {
		if strings.ToLower(s) == scopeLower {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the service has any of the specified scopes
func (ss *ServiceSubject) HasAnyScope(scopes []string) bool {
	for _, scope := range scopes {
		if ss.HasScope(scope) {
			return true
		}
	}
	return false
}

// HasAllScopes checks if the service has all of the specified scopes
func (ss *ServiceSubject) HasAllScopes(scopes []string) bool {
	for _, scope := range scopes {
		if !ss.HasScope(scope) {
			return false
		}
	}
	return true
}

// AddScope adds a new scope to the service
func (ss *ServiceSubject) AddScope(scope string) {
	if !ss.HasScope(scope) {
		ss.Scopes = append(ss.Scopes, scope)
	}
}

// RemoveScope removes a scope from the service
func (ss *ServiceSubject) RemoveScope(scope string) {
	scopeLower := strings.ToLower(scope)
	newScopes := make([]string, 0, len(ss.Scopes))
	for _, s := range ss.Scopes {
		if strings.ToLower(s) != scopeLower {
			newScopes = append(newScopes, s)
		}
	}
	ss.Scopes = newScopes
}

// SetMetadata sets a metadata key-value pair
func (ss *ServiceSubject) SetMetadata(key string, value interface{}) {
	if ss.Metadata == nil {
		ss.Metadata = make(map[string]interface{})
	}
	ss.Metadata[key] = value
}

// GetMetadata retrieves a metadata value by key
func (ss *ServiceSubject) GetMetadata(key string) (interface{}, bool) {
	if ss.Metadata == nil {
		return nil, false
	}
	value, exists := ss.Metadata[key]
	return value, exists
}
