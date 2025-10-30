package storage

import (
	"fmt"
	"time"

	"abac_go_example/models"
)

// MockStorage implements Storage interface for testing
type MockStorage struct {
	subjects     map[string]*models.Subject
	resources    map[string]*models.Resource
	actions      map[string]*models.Action
	policies     map[string]*models.Policy
	auditLogs    []*models.AuditLog
	users        map[string]*models.User
	userProfiles map[string]models.UserProfile // Store value, not pointer
	roles        map[string]*models.Role
	userRoles    map[string][]string // userID -> []roleIDs
}

// NewMockStorage creates a new mock storage instance
func NewMockStorage() *MockStorage {
	return &MockStorage{
		subjects:     make(map[string]*models.Subject),
		resources:    make(map[string]*models.Resource),
		actions:      make(map[string]*models.Action),
		policies:     make(map[string]*models.Policy),
		auditLogs:    make([]*models.AuditLog, 0),
		users:        make(map[string]*models.User),
		userProfiles: make(map[string]models.UserProfile),
		roles:        make(map[string]*models.Role),
		userRoles:    make(map[string][]string),
	}
}

// SetPolicies sets the policies for testing
func (m *MockStorage) SetPolicies(policies []*models.Policy) {
	m.policies = make(map[string]*models.Policy)
	for _, policy := range policies {
		m.policies[policy.ID] = policy
	}
}

// Subject operations
func (m *MockStorage) CreateSubject(subject *models.Subject) error {
	if subject.ID == "" {
		return fmt.Errorf("subject ID cannot be empty")
	}
	subject.CreatedAt = time.Now()
	subject.UpdatedAt = time.Now()
	m.subjects[subject.ID] = subject
	return nil
}

func (m *MockStorage) GetSubject(id string) (*models.Subject, error) {
	subject, exists := m.subjects[id]
	if !exists {
		return nil, fmt.Errorf("subject not found: %s", id)
	}
	return subject, nil
}

func (m *MockStorage) UpdateSubject(subject *models.Subject) error {
	if _, exists := m.subjects[subject.ID]; !exists {
		return fmt.Errorf("subject not found: %s", subject.ID)
	}
	subject.UpdatedAt = time.Now()
	m.subjects[subject.ID] = subject
	return nil
}

func (m *MockStorage) DeleteSubject(id string) error {
	if _, exists := m.subjects[id]; !exists {
		return fmt.Errorf("subject not found: %s", id)
	}
	delete(m.subjects, id)
	return nil
}

func (m *MockStorage) ListSubjects() ([]*models.Subject, error) {
	subjects := make([]*models.Subject, 0, len(m.subjects))
	for _, subject := range m.subjects {
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

func (m *MockStorage) GetAllSubjects() ([]*models.Subject, error) {
	return m.ListSubjects()
}

// Resource operations
func (m *MockStorage) CreateResource(resource *models.Resource) error {
	if resource.ID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}
	m.resources[resource.ID] = resource
	return nil
}

func (m *MockStorage) GetResource(id string) (*models.Resource, error) {
	resource, exists := m.resources[id]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", id)
	}
	return resource, nil
}

func (m *MockStorage) UpdateResource(resource *models.Resource) error {
	if _, exists := m.resources[resource.ID]; !exists {
		return fmt.Errorf("resource not found: %s", resource.ID)
	}
	m.resources[resource.ID] = resource
	return nil
}

func (m *MockStorage) DeleteResource(id string) error {
	if _, exists := m.resources[id]; !exists {
		return fmt.Errorf("resource not found: %s", id)
	}
	delete(m.resources, id)
	return nil
}

func (m *MockStorage) ListResources() ([]*models.Resource, error) {
	resources := make([]*models.Resource, 0, len(m.resources))
	for _, resource := range m.resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

func (m *MockStorage) GetAllResources() ([]*models.Resource, error) {
	return m.ListResources()
}

// Action operations
func (m *MockStorage) CreateAction(action *models.Action) error {
	if action.ID == "" {
		return fmt.Errorf("action ID cannot be empty")
	}
	m.actions[action.ID] = action
	return nil
}

func (m *MockStorage) GetAction(name string) (*models.Action, error) {
	// Search by action name instead of ID
	for _, action := range m.actions {
		if action.ActionName == name {
			return action, nil
		}
	}
	return nil, fmt.Errorf("action not found: %s", name)
}

func (m *MockStorage) GetActionByID(id string) (*models.Action, error) {
	action, exists := m.actions[id]
	if !exists {
		return nil, fmt.Errorf("action not found: %s", id)
	}
	return action, nil
}

func (m *MockStorage) UpdateAction(action *models.Action) error {
	if _, exists := m.actions[action.ID]; !exists {
		return fmt.Errorf("action not found: %s", action.ID)
	}
	m.actions[action.ID] = action
	return nil
}

func (m *MockStorage) DeleteAction(id string) error {
	if _, exists := m.actions[id]; !exists {
		return fmt.Errorf("action not found: %s", id)
	}
	delete(m.actions, id)
	return nil
}

func (m *MockStorage) ListActions() ([]*models.Action, error) {
	actions := make([]*models.Action, 0, len(m.actions))
	for _, action := range m.actions {
		actions = append(actions, action)
	}
	return actions, nil
}

func (m *MockStorage) GetAllActions() ([]*models.Action, error) {
	return m.ListActions()
}

// Policy operations
func (m *MockStorage) CreatePolicy(policy *models.Policy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	m.policies[policy.ID] = policy
	return nil
}

func (m *MockStorage) GetPolicy(id string) (*models.Policy, error) {
	policy, exists := m.policies[id]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", id)
	}
	return policy, nil
}

func (m *MockStorage) UpdatePolicy(policy *models.Policy) error {
	if _, exists := m.policies[policy.ID]; !exists {
		return fmt.Errorf("policy not found: %s", policy.ID)
	}
	policy.UpdatedAt = time.Now()
	m.policies[policy.ID] = policy
	return nil
}

func (m *MockStorage) DeletePolicy(id string) error {
	if _, exists := m.policies[id]; !exists {
		return fmt.Errorf("policy not found: %s", id)
	}
	delete(m.policies, id)
	return nil
}

func (m *MockStorage) GetPolicies() ([]*models.Policy, error) {
	policies := make([]*models.Policy, 0, len(m.policies))
	for _, policy := range m.policies {
		policies = append(policies, policy)
	}
	return policies, nil
}

func (m *MockStorage) ListPolicies() ([]*models.Policy, error) {
	return m.GetPolicies()
}

// Audit operations
func (m *MockStorage) CreateAuditLog(auditLog *models.AuditLog) error {
	if auditLog.RequestID == "" {
		return fmt.Errorf("audit log request ID cannot be empty")
	}
	auditLog.ID = int64(len(m.auditLogs) + 1)
	auditLog.CreatedAt = time.Now()
	m.auditLogs = append(m.auditLogs, auditLog)
	return nil
}

func (m *MockStorage) LogAudit(auditLog *models.AuditLog) error {
	return m.CreateAuditLog(auditLog)
}

func (m *MockStorage) GetAuditLogs(limit, offset int) ([]*models.AuditLog, error) {
	if offset >= len(m.auditLogs) {
		return []*models.AuditLog{}, nil
	}

	end := offset + limit
	if end > len(m.auditLogs) {
		end = len(m.auditLogs)
	}

	return m.auditLogs[offset:end], nil
}

// Health check
func (m *MockStorage) HealthCheck() error {
	return nil
}

// Close
func (m *MockStorage) Close() error {
	return nil
}

// Additional helper methods for testing

// Clear clears all data
func (m *MockStorage) Clear() {
	m.subjects = make(map[string]*models.Subject)
	m.resources = make(map[string]*models.Resource)
	m.actions = make(map[string]*models.Action)
	m.policies = make(map[string]*models.Policy)
	m.auditLogs = make([]*models.AuditLog, 0)
}

// SeedTestData seeds mock storage with test data
func (m *MockStorage) SeedTestData() {
	// Create test subjects
	subjects := []*models.Subject{
		{
			ID:          "user-123",
			ExternalID:  "john.doe@company.com",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department":   "Engineering",
				"level":        5,
				"clearance":    "confidential",
				"mfa_verified": true,
			},
		},
		{
			ID:          "user-456",
			ExternalID:  "jane.smith@company.com",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department":   "Finance",
				"level":        6,
				"clearance":    "internal",
				"mfa_verified": true,
			},
		},
		{
			ID:          "admin-001",
			ExternalID:  "admin@company.com",
			SubjectType: "admin",
			Attributes: map[string]interface{}{
				"department":   "IT",
				"level":        9,
				"clearance":    "top_secret",
				"mfa_verified": true,
			},
		},
	}

	for _, subject := range subjects {
		m.CreateSubject(subject)
	}

	// Create test resources
	resources := []*models.Resource{
		{
			ID:           "res-123",
			ResourceType: "document",
			ResourceID:   "/documents/project-alpha.pdf",
			Attributes: map[string]interface{}{
				"classification": "confidential",
				"project":        "alpha",
				"owner":          "engineering-team",
			},
		},
		{
			ID:           "res-456",
			ResourceType: "api",
			ResourceID:   "/api/financial/reports",
			Attributes: map[string]interface{}{
				"classification": "internal",
				"department":     "finance",
			},
		},
	}

	for _, resource := range resources {
		m.CreateResource(resource)
	}

	// Create test actions
	actions := []*models.Action{
		{
			ID:             "action-read",
			ActionName:     "read",
			ActionCategory: "data-access",
		},
		{
			ID:             "action-write",
			ActionName:     "write",
			ActionCategory: "data-modification",
		},
		{
			ID:             "action-admin",
			ActionName:     "admin",
			ActionCategory: "system-administration",
		},
	}

	for _, action := range actions {
		m.CreateAction(action)
	}
}

// User-based ABAC methods

// GetUser retrieves a user by ID
func (m *MockStorage) GetUser(id string) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return user, nil
}

// GetUserWithRelations retrieves a user with all relations
func (m *MockStorage) GetUserWithRelations(id string) (*models.User, error) {
	user, err := m.GetUser(id)
	if err != nil {
		return nil, err
	}

	// Load profile
	if profile, exists := m.userProfiles[id]; exists {
		user.Profile = &profile
	}

	// Load roles
	if roleIDs, exists := m.userRoles[id]; exists {
		user.Roles = make([]models.Role, 0, len(roleIDs))
		for _, roleID := range roleIDs {
			if role, exists := m.roles[roleID]; exists {
				user.Roles = append(user.Roles, *role)
			}
		}
	}

	return user, nil
}

// GetUserProfile retrieves a user profile
func (m *MockStorage) GetUserProfile(userID string) (*models.UserProfile, error) {
	profile, exists := m.userProfiles[userID]
	if !exists {
		return nil, fmt.Errorf("user profile not found for user: %s", userID)
	}
	return &profile, nil
}

// GetUserRoles retrieves user roles
func (m *MockStorage) GetUserRoles(userID string) ([]models.Role, error) {
	roleIDs, exists := m.userRoles[userID]
	if !exists {
		return []models.Role{}, nil
	}

	roles := make([]models.Role, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if role, exists := m.roles[roleID]; exists {
			roles = append(roles, *role)
		}
	}
	return roles, nil
}

// GetUserAttributes builds ABAC attributes from user data
func (m *MockStorage) GetUserAttributes(userID string) (map[string]interface{}, error) {
	user, err := m.GetUserWithRelations(userID)
	if err != nil {
		return nil, err
	}

	var profile *models.UserProfile
	if user.Profile != nil {
		profile = user.Profile
	}

	userSubject := models.NewUserSubject(user, profile, user.Roles)
	if userSubject == nil {
		return nil, fmt.Errorf("failed to create user subject")
	}

	return userSubject.GetAttributes(), nil
}

// BuildSubjectFromUser creates a SubjectInterface from user ID
func (m *MockStorage) BuildSubjectFromUser(userID string) (models.SubjectInterface, error) {
	user, err := m.GetUserWithRelations(userID)
	if err != nil {
		return nil, err
	}

	var profile *models.UserProfile
	if user.Profile != nil {
		profile = user.Profile
	}

	return models.NewUserSubject(user, profile, user.Roles), nil
}

// GetAllUsers retrieves all users
func (m *MockStorage) GetAllUsers(status string, limit, offset int) ([]*models.User, error) {
	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		if status != "" && user.Status != status {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// CreateUser creates a new user
func (m *MockStorage) CreateUser(user *models.User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

// CreateUserProfile creates a new user profile
func (m *MockStorage) CreateUserProfile(profile *models.UserProfile) error {
	if profile.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	m.userProfiles[profile.UserID] = *profile
	return nil
}

// UpdateUser updates a user
func (m *MockStorage) UpdateUser(user *models.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

// UpdateUserProfile updates a user profile
func (m *MockStorage) UpdateUserProfile(profile *models.UserProfile) error {
	if _, exists := m.userProfiles[profile.UserID]; !exists {
		return fmt.Errorf("user profile not found for user: %s", profile.UserID)
	}
	profile.UpdatedAt = time.Now()
	m.userProfiles[profile.UserID] = *profile
	return nil
}

// DeleteUser deletes a user
func (m *MockStorage) DeleteUser(id string) error {
	if _, exists := m.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	delete(m.users, id)
	delete(m.userProfiles, id)
	delete(m.userRoles, id)
	return nil
}

// AssignRole assigns a role to a user
func (m *MockStorage) AssignRole(userID, roleID, assignedBy string) error {
	if _, exists := m.users[userID]; !exists {
		return fmt.Errorf("user not found: %s", userID)
	}
	if _, exists := m.roles[roleID]; !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	if m.userRoles[userID] == nil {
		m.userRoles[userID] = make([]string, 0)
	}

	// Check if role already assigned
	for _, existingRoleID := range m.userRoles[userID] {
		if existingRoleID == roleID {
			return nil // Already assigned
		}
	}

	m.userRoles[userID] = append(m.userRoles[userID], roleID)
	return nil
}

// RevokeRole revokes a role from a user
func (m *MockStorage) RevokeRole(userID, roleID string) error {
	if roleIDs, exists := m.userRoles[userID]; exists {
		newRoles := make([]string, 0, len(roleIDs))
		for _, rid := range roleIDs {
			if rid != roleID {
				newRoles = append(newRoles, rid)
			}
		}
		m.userRoles[userID] = newRoles
	}
	return nil
}

// GetRoleByCode retrieves a role by code
func (m *MockStorage) GetRoleByCode(code string) (*models.Role, error) {
	for _, role := range m.roles {
		if role.RoleCode == code {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found: %s", code)
}
