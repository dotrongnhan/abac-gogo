package storage

import "abac_go_example/models"

// Storage interface defines the contract for data access
type Storage interface {
	// Core ABAC entities (legacy)
	GetSubject(id string) (*models.Subject, error)
	GetResource(id string) (*models.Resource, error)
	GetAction(name string) (*models.Action, error)
	GetPolicies() ([]*models.Policy, error)

	// User-based ABAC operations (new)
	GetUser(id string) (*models.User, error)
	GetUserWithRelations(id string) (*models.User, error)
	GetUserProfile(userID string) (*models.UserProfile, error)
	GetUserRoles(userID string) ([]models.Role, error)
	GetUserAttributes(userID string) (map[string]interface{}, error)
	BuildSubjectFromUser(userID string) (models.SubjectInterface, error)

	// Bulk operations
	GetAllSubjects() ([]*models.Subject, error)
	GetAllResources() ([]*models.Resource, error)
	GetAllActions() ([]*models.Action, error)
	GetAllUsers(status string, limit, offset int) ([]*models.User, error)

	// CRUD operations
	CreateSubject(subject *models.Subject) error
	CreateResource(resource *models.Resource) error
	CreateAction(action *models.Action) error
	CreatePolicy(policy *models.Policy) error
	CreateUser(user *models.User) error
	CreateUserProfile(profile *models.UserProfile) error

	UpdateSubject(subject *models.Subject) error
	UpdateResource(resource *models.Resource) error
	UpdateAction(action *models.Action) error
	UpdatePolicy(policy *models.Policy) error
	UpdateUser(user *models.User) error
	UpdateUserProfile(profile *models.UserProfile) error

	DeleteSubject(id string) error
	DeleteResource(id string) error
	DeleteAction(id string) error
	DeletePolicy(id string) error
	DeleteUser(id string) error

	// Role operations
	AssignRole(userID, roleID, assignedBy string) error
	RevokeRole(userID, roleID string) error
	GetRoleByCode(code string) (*models.Role, error)

	// Audit operations
	LogAudit(auditLog *models.AuditLog) error
	GetAuditLogs(limit, offset int) ([]*models.AuditLog, error)

	// Connection management
	Close() error
}
