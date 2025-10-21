package storage

import "abac_go_example/models"

// Storage interface defines the contract for data access
type Storage interface {
	// Core ABAC entities
	GetSubject(id string) (*models.Subject, error)
	GetResource(id string) (*models.Resource, error)
	GetAction(name string) (*models.Action, error)
	GetPolicies() ([]*models.Policy, error)

	// Bulk operations
	GetAllSubjects() ([]*models.Subject, error)
	GetAllResources() ([]*models.Resource, error)
	GetAllActions() ([]*models.Action, error)

	// CRUD operations
	CreateSubject(subject *models.Subject) error
	CreateResource(resource *models.Resource) error
	CreateAction(action *models.Action) error
	CreatePolicy(policy *models.Policy) error

	UpdateSubject(subject *models.Subject) error
	UpdateResource(resource *models.Resource) error
	UpdateAction(action *models.Action) error
	UpdatePolicy(policy *models.Policy) error

	DeleteSubject(id string) error
	DeleteResource(id string) error
	DeleteAction(id string) error
	DeletePolicy(id string) error

	// Audit operations
	LogAudit(auditLog *models.AuditLog) error
	GetAuditLogs(limit, offset int) ([]*models.AuditLog, error)

	// Connection management
	Close() error
}
