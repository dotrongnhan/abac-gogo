package storage

import (
	"errors"
	"fmt"

	"abac_go_example/models"

	"gorm.io/gorm"
)

// PostgreSQLStorage implements Storage interface using PostgreSQL with GORM
type PostgreSQLStorage struct {
	db             *gorm.DB
	userRepository *UserRepository
}

// NewPostgreSQLStorage creates a new PostgreSQL storage instance
func NewPostgreSQLStorage(config *DatabaseConfig) (*PostgreSQLStorage, error) {
	db, err := NewDatabaseConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	storage := &PostgreSQLStorage{
		db:             db,
		userRepository: NewUserRepository(db),
	}

	// Auto-migrate the schema
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return storage, nil
}

// migrate runs database migrations
func (s *PostgreSQLStorage) migrate() error {
	return s.db.AutoMigrate(
		// Legacy ABAC models
		&models.Subject{},
		&models.Resource{},
		&models.Action{},
		&models.Policy{},
		&models.AuditLog{},
		// User-based ABAC models
		&models.Company{},
		&models.Department{},
		&models.Position{},
		&models.Role{},
		&models.User{},
		&models.UserProfile{},
		&models.UserRole{},
		&models.UserAttributeHistory{},
	)
}

// GetSubject retrieves a subject by ID
func (s *PostgreSQLStorage) GetSubject(id string) (*models.Subject, error) {
	var subject models.Subject
	result := s.db.Where("id = ?", id).First(&subject)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subject not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get subject: %w", result.Error)
	}
	return &subject, nil
}

// GetResource retrieves a resource by ID
func (s *PostgreSQLStorage) GetResource(id string) (*models.Resource, error) {
	var resource models.Resource
	result := s.db.Where("id = ?", id).First(&resource)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("resource not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get resource: %w", result.Error)
	}
	return &resource, nil
}

// GetAction retrieves an action by name
func (s *PostgreSQLStorage) GetAction(name string) (*models.Action, error) {
	var action models.Action
	result := s.db.Where("action_name = ?", name).First(&action)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("action not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get action: %w", result.Error)
	}
	return &action, nil
}

// GetPolicies retrieves all policies
func (s *PostgreSQLStorage) GetPolicies() ([]*models.Policy, error) {
	var policies []*models.Policy
	result := s.db.Where("enabled = ?", true).Order("priority DESC").Find(&policies)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get policies: %w", result.Error)
	}
	return policies, nil
}

// GetAllSubjects retrieves all subjects
func (s *PostgreSQLStorage) GetAllSubjects() ([]*models.Subject, error) {
	var subjects []*models.Subject
	result := s.db.Find(&subjects)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all subjects: %w", result.Error)
	}
	return subjects, nil
}

// GetAllResources retrieves all resources
func (s *PostgreSQLStorage) GetAllResources() ([]*models.Resource, error) {
	var resources []*models.Resource
	result := s.db.Find(&resources)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all resources: %w", result.Error)
	}
	return resources, nil
}

// GetAllActions retrieves all actions
func (s *PostgreSQLStorage) GetAllActions() ([]*models.Action, error) {
	var actions []*models.Action
	result := s.db.Find(&actions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all actions: %w", result.Error)
	}
	return actions, nil
}

// CreateSubject creates a new subject
func (s *PostgreSQLStorage) CreateSubject(subject *models.Subject) error {
	result := s.db.Create(subject)
	if result.Error != nil {
		return fmt.Errorf("failed to create subject: %w", result.Error)
	}
	return nil
}

// CreateResource creates a new resource
func (s *PostgreSQLStorage) CreateResource(resource *models.Resource) error {
	result := s.db.Create(resource)
	if result.Error != nil {
		return fmt.Errorf("failed to create resource: %w", result.Error)
	}
	return nil
}

// CreateAction creates a new action
func (s *PostgreSQLStorage) CreateAction(action *models.Action) error {
	result := s.db.Create(action)
	if result.Error != nil {
		return fmt.Errorf("failed to create action: %w", result.Error)
	}
	return nil
}

// CreatePolicy creates a new policy
func (s *PostgreSQLStorage) CreatePolicy(policy *models.Policy) error {
	result := s.db.Create(policy)
	if result.Error != nil {
		return fmt.Errorf("failed to create policy: %w", result.Error)
	}
	return nil
}

// UpdateSubject updates an existing subject
func (s *PostgreSQLStorage) UpdateSubject(subject *models.Subject) error {
	result := s.db.Save(subject)
	if result.Error != nil {
		return fmt.Errorf("failed to update subject: %w", result.Error)
	}
	return nil
}

// UpdateResource updates an existing resource
func (s *PostgreSQLStorage) UpdateResource(resource *models.Resource) error {
	result := s.db.Save(resource)
	if result.Error != nil {
		return fmt.Errorf("failed to update resource: %w", result.Error)
	}
	return nil
}

// UpdateAction updates an existing action
func (s *PostgreSQLStorage) UpdateAction(action *models.Action) error {
	result := s.db.Save(action)
	if result.Error != nil {
		return fmt.Errorf("failed to update action: %w", result.Error)
	}
	return nil
}

// UpdatePolicy updates an existing policy
func (s *PostgreSQLStorage) UpdatePolicy(policy *models.Policy) error {
	result := s.db.Save(policy)
	if result.Error != nil {
		return fmt.Errorf("failed to update policy: %w", result.Error)
	}
	return nil
}

// DeleteSubject deletes a subject by ID
func (s *PostgreSQLStorage) DeleteSubject(id string) error {
	result := s.db.Delete(&models.Subject{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete subject: %w", result.Error)
	}
	return nil
}

// DeleteResource deletes a resource by ID
func (s *PostgreSQLStorage) DeleteResource(id string) error {
	result := s.db.Delete(&models.Resource{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete resource: %w", result.Error)
	}
	return nil
}

// DeleteAction deletes an action by ID
func (s *PostgreSQLStorage) DeleteAction(id string) error {
	result := s.db.Delete(&models.Action{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete action: %w", result.Error)
	}
	return nil
}

// DeletePolicy deletes a policy by ID
func (s *PostgreSQLStorage) DeletePolicy(id string) error {
	result := s.db.Delete(&models.Policy{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete policy: %w", result.Error)
	}
	return nil
}

// LogAudit creates an audit log entry
func (s *PostgreSQLStorage) LogAudit(auditLog *models.AuditLog) error {
	result := s.db.Create(auditLog)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}
	return nil
}

// GetAuditLogs retrieves audit logs with pagination
func (s *PostgreSQLStorage) GetAuditLogs(limit, offset int) ([]*models.AuditLog, error) {
	var auditLogs []*models.AuditLog
	result := s.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&auditLogs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", result.Error)
	}
	return auditLogs, nil
}

// Close closes the database connection
func (s *PostgreSQLStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// User-based ABAC methods

// GetUser retrieves a user by ID
func (s *PostgreSQLStorage) GetUser(id string) (*models.User, error) {
	return s.userRepository.GetUserByID(id)
}

// GetUserWithRelations retrieves a user with all related data
func (s *PostgreSQLStorage) GetUserWithRelations(id string) (*models.User, error) {
	return s.userRepository.GetUserWithRelations(id)
}

// GetUserProfile retrieves the profile for a specific user
func (s *PostgreSQLStorage) GetUserProfile(userID string) (*models.UserProfile, error) {
	return s.userRepository.GetUserProfile(userID)
}

// GetUserRoles retrieves all active roles for a user
func (s *PostgreSQLStorage) GetUserRoles(userID string) ([]models.Role, error) {
	return s.userRepository.GetUserRoles(userID)
}

// GetUserAttributes builds ABAC attributes from user data
func (s *PostgreSQLStorage) GetUserAttributes(userID string) (map[string]interface{}, error) {
	return s.userRepository.GetUserAttributes(userID)
}

// BuildSubjectFromUser creates a SubjectInterface from a user ID
func (s *PostgreSQLStorage) BuildSubjectFromUser(userID string) (models.SubjectInterface, error) {
	user, err := s.userRepository.GetUserWithRelations(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user with relations: %w", err)
	}

	var profile *models.UserProfile
	var roles []models.Role

	if user.Profile != nil {
		profile = user.Profile
	}
	if len(user.Roles) > 0 {
		roles = user.Roles
	}

	userSubject := models.NewUserSubject(user, profile, roles)
	if userSubject == nil {
		return nil, fmt.Errorf("failed to create user subject")
	}

	return userSubject, nil
}

// GetAllUsers retrieves all users with optional filters
func (s *PostgreSQLStorage) GetAllUsers(status string, limit, offset int) ([]*models.User, error) {
	return s.userRepository.GetAllUsers(status, limit, offset)
}

// CreateUser creates a new user
func (s *PostgreSQLStorage) CreateUser(user *models.User) error {
	return s.userRepository.CreateUser(user)
}

// CreateUserProfile creates a new user profile
func (s *PostgreSQLStorage) CreateUserProfile(profile *models.UserProfile) error {
	return s.userRepository.CreateUserProfile(profile)
}

// UpdateUser updates an existing user
func (s *PostgreSQLStorage) UpdateUser(user *models.User) error {
	return s.userRepository.UpdateUser(user)
}

// UpdateUserProfile updates an existing user profile
func (s *PostgreSQLStorage) UpdateUserProfile(profile *models.UserProfile) error {
	return s.userRepository.UpdateUserProfile(profile)
}

// DeleteUser deletes a user by ID
func (s *PostgreSQLStorage) DeleteUser(id string) error {
	return s.userRepository.DeleteUser(id)
}

// AssignRole assigns a role to a user
func (s *PostgreSQLStorage) AssignRole(userID, roleID, assignedBy string) error {
	return s.userRepository.AssignRole(userID, roleID, assignedBy)
}

// RevokeRole revokes a role from a user
func (s *PostgreSQLStorage) RevokeRole(userID, roleID string) error {
	return s.userRepository.RevokeRole(userID, roleID)
}

// GetRoleByCode retrieves a role by its code
func (s *PostgreSQLStorage) GetRoleByCode(code string) (*models.Role, error) {
	return s.userRepository.GetRoleByCode(code)
}
