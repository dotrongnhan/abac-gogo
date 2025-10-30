package models

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrMissingAuthentication is returned when no authentication is provided
	ErrMissingAuthentication = errors.New("missing authentication credentials")
	// ErrInvalidAuthenticationType is returned when authentication type is not recognized
	ErrInvalidAuthenticationType = errors.New("invalid authentication type")
	// ErrUserDataNotFound is returned when user data cannot be found
	ErrUserDataNotFound = errors.New("user data not found")
)

const (
	headerSubjectID     = "X-Subject-ID"
	headerUserID        = "X-User-ID"
	headerServiceToken  = "X-Service-Token"
	headerAPIKey        = "X-API-Key"
	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "
)

// SubjectFactory creates Subject instances from various authentication sources
type SubjectFactory struct {
	userLoader    UserLoader
	serviceLoader ServiceLoader
}

// UserLoader defines the interface for loading user data
type UserLoader interface {
	LoadUser(userID string) (*User, *UserProfile, []Role, error)
}

// ServiceLoader defines the interface for loading service data
type ServiceLoader interface {
	LoadService(serviceID string) (*ServiceSubject, error)
}

// NewSubjectFactory creates a new SubjectFactory instance
func NewSubjectFactory(userLoader UserLoader, serviceLoader ServiceLoader) *SubjectFactory {
	return &SubjectFactory{
		userLoader:    userLoader,
		serviceLoader: serviceLoader,
	}
}

// CreateFromRequest creates a Subject from an HTTP request
// It detects the authentication type and delegates to appropriate creation method
func (sf *SubjectFactory) CreateFromRequest(r *http.Request) (SubjectInterface, error) {
	// Priority 1: X-User-ID header (modern user authentication)
	if userID := r.Header.Get(headerUserID); userID != "" {
		return sf.CreateFromUserID(userID)
	}

	// Priority 2: X-Subject-ID header (legacy subject ID, backward compatibility)
	if subjectID := r.Header.Get(headerSubjectID); subjectID != "" {
		return sf.CreateFromSubjectID(subjectID)
	}

	// Priority 3: Authorization Bearer token (JWT)
	if authHeader := r.Header.Get(headerAuthorization); authHeader != "" {
		if strings.HasPrefix(authHeader, bearerPrefix) {
			token := strings.TrimPrefix(authHeader, bearerPrefix)
			return sf.CreateFromJWT(token)
		}
	}

	// Priority 4: X-Service-Token header (service-to-service)
	if serviceToken := r.Header.Get(headerServiceToken); serviceToken != "" {
		return sf.CreateFromServiceToken(serviceToken)
	}

	// Priority 5: X-API-Key header (API key authentication)
	if apiKey := r.Header.Get(headerAPIKey); apiKey != "" {
		return sf.CreateFromAPIKey(apiKey)
	}

	return nil, ErrMissingAuthentication
}

// CreateFromUserID creates a UserSubject from a user ID
func (sf *SubjectFactory) CreateFromUserID(userID string) (SubjectInterface, error) {
	if sf.userLoader == nil {
		return nil, fmt.Errorf("user loader not configured")
	}

	user, profile, roles, err := sf.userLoader.LoadUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	if user == nil {
		return nil, ErrUserDataNotFound
	}

	return NewUserSubject(user, profile, roles), nil
}

// CreateFromSubjectID creates a Subject from a legacy subject ID
// This method provides backward compatibility with the old Subject model
func (sf *SubjectFactory) CreateFromSubjectID(subjectID string) (SubjectInterface, error) {
	// Try to load as user first
	if sf.userLoader != nil {
		user, profile, roles, err := sf.userLoader.LoadUser(subjectID)
		if err == nil && user != nil {
			return NewUserSubject(user, profile, roles), nil
		}
	}

	// If not found as user, try to load as service
	if sf.serviceLoader != nil {
		service, err := sf.serviceLoader.LoadService(subjectID)
		if err == nil && service != nil {
			return service, nil
		}
	}

	return nil, fmt.Errorf("subject not found: %s", subjectID)
}

// CreateFromUser creates a UserSubject from a User entity
func (sf *SubjectFactory) CreateFromUser(user *User, profile *UserProfile, roles []Role) SubjectInterface {
	return NewUserSubject(user, profile, roles)
}

// CreateFromJWT creates a Subject from a JWT token
// This is a placeholder - in production, you would parse and validate the JWT
func (sf *SubjectFactory) CreateFromJWT(token string) (SubjectInterface, error) {
	// TODO: Implement JWT parsing and validation
	// For now, this is a placeholder that returns an error
	return nil, fmt.Errorf("JWT authentication not yet implemented")
}

// CreateFromServiceToken creates a ServiceSubject from a service token
// This is a placeholder - in production, you would validate the token
func (sf *SubjectFactory) CreateFromServiceToken(token string) (SubjectInterface, error) {
	// TODO: Implement service token validation
	// For now, this is a placeholder
	return nil, fmt.Errorf("service token authentication not yet implemented")
}

// CreateFromAPIKey creates a Subject from an API key
// This is a placeholder - in production, you would validate the API key
func (sf *SubjectFactory) CreateFromAPIKey(apiKey string) (SubjectInterface, error) {
	// TODO: Implement API key validation
	// For now, this is a placeholder
	return nil, fmt.Errorf("API key authentication not yet implemented")
}

// CreateFromClaims creates a Subject from JWT claims
// This helper method processes already-validated JWT claims
func (sf *SubjectFactory) CreateFromClaims(claims map[string]interface{}) (SubjectInterface, error) {
	// Check for user_id in claims
	if userID, ok := claims["user_id"].(string); ok && userID != "" {
		return sf.CreateFromUserID(userID)
	}

	// Check for sub (standard JWT subject claim)
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sf.CreateFromUserID(sub)
	}

	// Check for service claims
	if serviceID, ok := claims["service_id"].(string); ok && serviceID != "" {
		if sf.serviceLoader != nil {
			return sf.serviceLoader.LoadService(serviceID)
		}
	}

	return nil, fmt.Errorf("cannot determine subject from claims")
}

// DetectAuthenticationType detects the type of authentication from an HTTP request
func DetectAuthenticationType(r *http.Request) string {
	if r.Header.Get(headerUserID) != "" {
		return "user_id"
	}
	if r.Header.Get(headerSubjectID) != "" {
		return "subject_id"
	}
	if strings.HasPrefix(r.Header.Get(headerAuthorization), bearerPrefix) {
		return "jwt"
	}
	if r.Header.Get(headerServiceToken) != "" {
		return "service_token"
	}
	if r.Header.Get(headerAPIKey) != "" {
		return "api_key"
	}
	return "none"
}
