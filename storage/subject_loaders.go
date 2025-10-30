package storage

import (
	"fmt"

	"abac_go_example/models"
)

// StorageUserLoader implements models.UserLoader using Storage interface
type StorageUserLoader struct {
	storage Storage
}

// NewStorageUserLoader creates a new StorageUserLoader
func NewStorageUserLoader(storage Storage) *StorageUserLoader {
	return &StorageUserLoader{
		storage: storage,
	}
}

// LoadUser loads a user with all relations from storage
func (sul *StorageUserLoader) LoadUser(userID string) (*models.User, *models.UserProfile, []models.Role, error) {
	user, err := sul.storage.GetUserWithRelations(userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load user: %w", err)
	}

	var profile *models.UserProfile
	var roles []models.Role

	if user.Profile != nil {
		profile = user.Profile
	}

	if len(user.Roles) > 0 {
		roles = user.Roles
	}

	return user, profile, roles, nil
}

// StorageServiceLoader implements models.ServiceLoader using Storage interface
type StorageServiceLoader struct {
	storage Storage
}

// NewStorageServiceLoader creates a new StorageServiceLoader
func NewStorageServiceLoader(storage Storage) *StorageServiceLoader {
	return &StorageServiceLoader{
		storage: storage,
	}
}

// LoadService loads a service subject from storage
// This is a placeholder - in production, you would have a services table
func (ssl *StorageServiceLoader) LoadService(serviceID string) (*models.ServiceSubject, error) {
	// For now, try to load from legacy subject table and convert to ServiceSubject
	subject, err := ssl.storage.GetSubject(serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load service: %w", err)
	}

	if subject == nil {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	// Convert legacy subject to ServiceSubject if it's a service type
	if subject.SubjectType == "service" {
		serviceSubject := models.NewServiceSubject(
			subject.ID,
			subject.ID,
			"default",
		)

		// Extract metadata if available
		if subject.Attributes != nil {
			if serviceName, ok := subject.Attributes["service_name"].(string); ok {
				serviceSubject.ServiceName = serviceName
			}
			if namespace, ok := subject.Attributes["namespace"].(string); ok {
				serviceSubject.Namespace = namespace
			}
			if scopes, ok := subject.Attributes["scopes"].([]interface{}); ok {
				for _, scope := range scopes {
					if scopeStr, ok := scope.(string); ok {
						serviceSubject.AddScope(scopeStr)
					}
				}
			}
		}

		serviceSubject.Status = subject.SubjectType
		serviceSubject.Metadata = map[string]interface{}(subject.Metadata)

		return serviceSubject, nil
	}

	return nil, fmt.Errorf("subject is not a service: %s", serviceID)
}
