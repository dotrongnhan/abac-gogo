package attributes

import (
	"context"
	"testing"
	"time"

	"abac_go_example/constants"
	"abac_go_example/models"
)

// Mock storage for testing
type mockStorage struct {
	subjects  map[string]*models.Subject
	resources map[string]*models.Resource
	actions   map[string]*models.Action
}

func (m *mockStorage) GetSubject(id string) (*models.Subject, error) {
	if subject, exists := m.subjects[id]; exists {
		return subject, nil
	}
	return nil, nil
}

func (m *mockStorage) GetResource(id string) (*models.Resource, error) {
	if resource, exists := m.resources[id]; exists {
		return resource, nil
	}
	return nil, nil
}

func (m *mockStorage) GetAction(name string) (*models.Action, error) {
	if action, exists := m.actions[name]; exists {
		return action, nil
	}
	return nil, nil
}

func (m *mockStorage) GetPolicies() ([]*models.Policy, error) {
	return []*models.Policy{}, nil
}

func (m *mockStorage) GetAllSubjects() ([]*models.Subject, error) {
	return []*models.Subject{}, nil
}

func (m *mockStorage) GetAllResources() ([]*models.Resource, error) {
	return []*models.Resource{}, nil
}

func (m *mockStorage) GetAllActions() ([]*models.Action, error) {
	return []*models.Action{}, nil
}

// Additional methods to implement Storage interface
func (m *mockStorage) CreateSubject(subject *models.Subject) error    { return nil }
func (m *mockStorage) CreateResource(resource *models.Resource) error { return nil }
func (m *mockStorage) CreateAction(action *models.Action) error       { return nil }
func (m *mockStorage) CreatePolicy(policy *models.Policy) error       { return nil }
func (m *mockStorage) UpdateSubject(subject *models.Subject) error    { return nil }
func (m *mockStorage) UpdateResource(resource *models.Resource) error { return nil }
func (m *mockStorage) UpdateAction(action *models.Action) error       { return nil }
func (m *mockStorage) UpdatePolicy(policy *models.Policy) error       { return nil }
func (m *mockStorage) DeleteSubject(id string) error                  { return nil }
func (m *mockStorage) DeleteResource(id string) error                 { return nil }
func (m *mockStorage) DeleteAction(id string) error                   { return nil }
func (m *mockStorage) DeletePolicy(id string) error                   { return nil }
func (m *mockStorage) LogAudit(auditLog *models.AuditLog) error       { return nil }
func (m *mockStorage) GetAuditLogs(limit, offset int) ([]*models.AuditLog, error) {
	return []*models.AuditLog{}, nil
}
func (m *mockStorage) Close() error { return nil }

func createMockStorage() *mockStorage {
	return &mockStorage{
		subjects: map[string]*models.Subject{
			"sub-001": {
				ID:          "sub-001",
				ExternalID:  "john.doe@company.com",
				SubjectType: "user",
				Attributes: map[string]interface{}{
					"department":                 "engineering",
					"role":                       []string{"senior_developer", "code_reviewer"},
					"clearance_level":            3,
					constants.ContextKeyHireDate: "2019-01-15",
				},
			},
		},
		resources: map[string]*models.Resource{
			"res-001": {
				ID:           "res-001",
				ResourceType: "api_endpoint",
				ResourceID:   "/api/v1/users",
				Path:         "api.v1.users",
				Attributes: map[string]interface{}{
					"data_classification": "internal",
					"requires_auth":       true,
				},
			},
		},
		actions: map[string]*models.Action{
			"read": {
				ID:             "act-001",
				ActionName:     "read",
				ActionCategory: "crud",
				Description:    "Read/View resource",
			},
		},
	}
}

func TestEnrichContext(t *testing.T) {
	storage := createMockStorage()
	resolver := NewAttributeResolver(storage)

	request := &models.EvaluationRequest{
		RequestID:  "test-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			constants.ContextKeyTimestamp: "2024-01-15T14:00:00Z",
			constants.ContextKeySourceIP:  "10.0.1.100",
		},
	}

	context, err := resolver.EnrichContext(request)
	if err != nil {
		t.Fatalf("Failed to enrich context: %v", err)
	}

	// Test that context was enriched properly
	if context.Subject == nil {
		t.Error("Context should have subject")
	}

	if context.Resource == nil {
		t.Error("Context should have resource")
	}

	if context.Action == nil {
		t.Error("Context should have action")
	}

	if context.Environment == nil {
		t.Error("Context should have environment")
	}

	// Test environment enrichment
	if _, exists := context.Environment[constants.ContextKeyTimeOfDayShort]; !exists {
		t.Error("Environment should have time_of_day")
	}

	if _, exists := context.Environment[constants.ContextKeyDayOfWeekShort]; !exists {
		t.Error("Environment should have day_of_week")
	}

	if _, exists := context.Environment[constants.ContextKeyIsBusinessHours]; !exists {
		t.Error("Environment should have is_business_hours")
	}

	if _, exists := context.Environment[constants.ContextKeyIsInternalIP]; !exists {
		t.Error("Environment should have is_internal_ip")
	}
}

func TestGetAttributeValue(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	// Test with map
	testMap := map[string]interface{}{
		"attributes": map[string]interface{}{
			"department": "engineering",
			"role":       []string{"senior_developer"},
		},
	}

	// Test simple path
	value := resolver.GetAttributeValue(testMap, "attributes.department")
	if value != "engineering" {
		t.Errorf("Expected 'engineering', got %v", value)
	}

	// Test array path
	value = resolver.GetAttributeValue(testMap, "attributes.role")
	if roles, ok := value.([]string); !ok || len(roles) != 1 || roles[0] != "senior_developer" {
		t.Errorf("Expected ['senior_developer'], got %v", value)
	}

	// Test non-existent path
	value = resolver.GetAttributeValue(testMap, "attributes.nonexistent")
	if value != nil {
		t.Errorf("Expected nil for non-existent path, got %v", value)
	}

	// Test with struct
	subject := &models.Subject{
		ID:          "sub-001",
		SubjectType: "user",
		Attributes: map[string]interface{}{
			"department": "engineering",
		},
	}

	value = resolver.GetAttributeValue(subject, "SubjectType")
	if value != "user" {
		t.Errorf("Expected 'user', got %v", value)
	}

	value = resolver.GetAttributeValue(subject, "attributes.department")
	if value != "engineering" {
		t.Errorf("Expected 'engineering', got %v", value)
	}
}

func TestMatchResourcePattern(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		pattern  string
		resource string
		expected bool
	}{
		{"*", "/api/v1/users", true},
		{"/api/v1/users", "/api/v1/users", true},
		{"/api/v1/*", "/api/v1/users", true},
		{"/api/v1/*", "/api/v1/posts", true},
		{"/api/v2/*", "/api/v1/users", false},
		{"DOC-*-FINANCE", "DOC-2024-FINANCE", true},
		{"DOC-*-FINANCE", "DOC-2024-HR", false},
	}

	for _, tc := range testCases {
		result := resolver.MatchResourcePattern(tc.pattern, tc.resource)
		if result != tc.expected {
			t.Errorf("MatchResourcePattern(%s, %s) = %v, expected %v",
				tc.pattern, tc.resource, result, tc.expected)
		}
	}
}

func TestResolveHierarchy(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	hierarchy := resolver.ResolveHierarchy("/api/v1/users/123")

	expectedPaths := []string{
		"/api", "/api/*",
		"/api/v1", "/api/v1/*",
		"/api/v1/users", "/api/v1/users/*",
		"/api/v1/users/123", "/api/v1/users/123/*",
	}

	if len(hierarchy) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(hierarchy))
	}

	for i, expected := range expectedPaths {
		if i < len(hierarchy) && hierarchy[i] != expected {
			t.Errorf("Expected path %s at index %d, got %s", expected, i, hierarchy[i])
		}
	}
}

func TestEnvironmentEnrichment(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			input: map[string]interface{}{
				constants.ContextKeyTimestamp: "2024-01-15T14:00:00Z",
				constants.ContextKeySourceIP:  "10.0.1.100",
			},
			expected: map[string]interface{}{
				constants.ContextKeyTimeOfDayShort:  "14:00",
				constants.ContextKeyDayOfWeekShort:  "monday",
				constants.ContextKeyHour:            14,
				constants.ContextKeyIsBusinessHours: true,
				constants.ContextKeyIsInternalIP:    true,
				constants.ContextKeyIPSubnet:        "10.0.1.0/24",
			},
		},
		{
			input: map[string]interface{}{
				constants.ContextKeyTimestamp: "2024-01-15T22:00:00Z",
				constants.ContextKeySourceIP:  "203.0.113.1",
			},
			expected: map[string]interface{}{
				constants.ContextKeyTimeOfDayShort:  "22:00",
				constants.ContextKeyIsBusinessHours: false,
				constants.ContextKeyIsInternalIP:    false,
			},
		},
	}

	for _, tc := range testCases {
		enriched := resolver.enrichEnvironmentContext(tc.input)

		for key, expectedValue := range tc.expected {
			if actualValue, exists := enriched[key]; !exists {
				t.Errorf("Expected key %s to exist in enriched context", key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected %s = %v, got %v", key, expectedValue, actualValue)
			}
		}
	}
}

func TestDynamicAttributeResolution(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	subject := &models.Subject{
		ID: "sub-001",
		Attributes: map[string]interface{}{
			constants.ContextKeyHireDate: "2019-01-15",
		},
	}

	environment := map[string]interface{}{
		constants.ContextKeyTimestamp: time.Now().Format(time.RFC3339),
	}

	resolver.resolveDynamicAttributes(subject, environment)

	// Check that years_of_service was calculated
	if yearsOfService, exists := subject.Attributes[constants.ContextKeyYearsOfService]; !exists {
		t.Error("Expected years_of_service to be calculated")
	} else if years, ok := yearsOfService.(int); !ok || years < 5 {
		t.Errorf("Expected years_of_service to be at least 5, got %v", yearsOfService)
	}

	// Check that current time attributes were added
	if _, exists := subject.Attributes[constants.ContextKeyCurrentHour]; !exists {
		t.Error("Expected current_hour to be added")
	}

	if _, exists := subject.Attributes[constants.ContextKeyCurrentDay]; !exists {
		t.Error("Expected current_day to be added")
	}
}

func TestIsBusinessHours(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		time     time.Time
		expected bool
	}{
		{time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), true},  // Monday 10 AM
		{time.Date(2024, 1, 15, 7, 0, 0, 0, time.UTC), false},  // Monday 7 AM
		{time.Date(2024, 1, 15, 19, 0, 0, 0, time.UTC), false}, // Monday 7 PM
		{time.Date(2024, 1, 13, 10, 0, 0, 0, time.UTC), false}, // Saturday 10 AM
		{time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC), false}, // Sunday 10 AM
	}

	for _, tc := range testCases {
		result := resolver.isBusinessHours(tc.time)
		if result != tc.expected {
			t.Errorf("isBusinessHours(%v) = %v, expected %v", tc.time, result, tc.expected)
		}
	}
}

func TestIsInternalIP(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		ip       string
		expected bool
	}{
		{"10.0.1.100", true},
		{"192.168.1.100", true},
		{"172.16.1.100", true},
		{"127.0.0.1", true},
		{"localhost", true},
		{"203.0.113.1", false},
		{"8.8.8.8", false},
	}

	for _, tc := range testCases {
		result := resolver.isInternalIP(tc.ip)
		if result != tc.expected {
			t.Errorf("isInternalIP(%s) = %v, expected %v", tc.ip, result, tc.expected)
		}
	}
}

func TestGetIPSubnet(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		ip       string
		expected string
	}{
		{"10.0.1.100", "10.0.1.0/24"},
		{"192.168.1.50", "192.168.1.0/24"},
		{"172.16.0.1", "172.16.0.0/24"},
		{"invalid", "invalid"},
	}

	for _, tc := range testCases {
		result := resolver.getIPSubnet(tc.ip)
		if result != tc.expected {
			t.Errorf("getIPSubnet(%s) = %s, expected %s", tc.ip, result, tc.expected)
		}
	}
}

// Test input validation
func TestValidateRequest(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		name        string
		request     *models.EvaluationRequest
		expectError bool
	}{
		{
			name:        "nil request",
			request:     nil,
			expectError: true,
		},
		{
			name: "empty subject ID",
			request: &models.EvaluationRequest{
				SubjectID:  "",
				ResourceID: "res-001",
				Action:     "read",
			},
			expectError: true,
		},
		{
			name: "empty resource ID",
			request: &models.EvaluationRequest{
				SubjectID:  "sub-001",
				ResourceID: "",
				Action:     "read",
			},
			expectError: true,
		},
		{
			name: "empty action",
			request: &models.EvaluationRequest{
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "",
			},
			expectError: true,
		},
		{
			name: "valid request",
			request: &models.EvaluationRequest{
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "read",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := resolver.validateRequest(tc.request)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tc.name, err)
			}
		})
	}
}

// Test context timeout support
func TestEnrichContextWithTimeout(t *testing.T) {
	storage := createMockStorage()
	resolver := NewAttributeResolver(storage)

	request := &models.EvaluationRequest{
		RequestID:  "test-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			constants.ContextKeyTimestamp: "2024-01-15T14:00:00Z",
			constants.ContextKeySourceIP:  "10.0.1.100",
		},
	}

	// Test with valid context
	ctx := context.Background()
	enrichedContext, err := resolver.EnrichContextWithTimeout(ctx, request)
	if err != nil {
		t.Fatalf("Failed to enrich context with timeout: %v", err)
	}
	if enrichedContext == nil {
		t.Error("Context should not be nil")
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = resolver.EnrichContextWithTimeout(cancelledCtx, request)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

// Test improved IP detection with CIDR
func TestImprovedIPDetection(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		ip          string
		expected    bool
		description string
	}{
		{"10.0.1.100", true, "Class A private network"},
		{"192.168.1.100", true, "Class C private network"},
		{"172.16.1.100", true, "Class B private network (start)"},
		{"172.31.255.254", true, "Class B private network (end)"},
		{"127.0.0.1", true, "Loopback address"},
		{"127.255.255.254", true, "Loopback range"},
		{"localhost", true, "Localhost string"},
		{"203.0.113.1", false, "Public IP"},
		{"8.8.8.8", false, "Google DNS"},
		{"172.15.255.254", false, "Just outside private range"},
		{"172.32.0.1", false, "Just outside private range"},
		{"invalid-ip", false, "Invalid IP format"},
		{"", false, "Empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := resolver.isInternalIP(tc.ip)
			if result != tc.expected {
				t.Errorf("isInternalIP(%s) = %v, expected %v (%s)",
					tc.ip, result, tc.expected, tc.description)
			}
		})
	}
}

// Test enhanced pattern matching
func TestEnhancedPatternMatching(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	testCases := []struct {
		pattern     string
		resource    string
		expected    bool
		description string
	}{
		{"*", "/api/v1/users", true, "Universal wildcard"},
		{"/api/v1/users", "/api/v1/users", true, "Exact match"},
		{"/api/v1/*", "/api/v1/users", true, "Simple wildcard"},
		{"/api/*/users", "/api/v1/users", true, "Middle wildcard"},
		{"*.txt", "document.txt", true, "File extension pattern"},
		{"DOC-*-FINANCE", "DOC-2024-FINANCE", true, "Complex pattern"},
		{"/api/v2/*", "/api/v1/users", false, "No match"},
		{"*.pdf", "document.txt", false, "Wrong extension"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := resolver.MatchResourcePattern(tc.pattern, tc.resource)
			if result != tc.expected {
				t.Errorf("MatchResourcePattern(%s, %s) = %v, expected %v (%s)",
					tc.pattern, tc.resource, result, tc.expected, tc.description)
			}
		})
	}
}

// Test error handling in EnrichContext
func TestEnrichContextErrorHandling(t *testing.T) {
	resolver := NewAttributeResolver(createMockStorage())

	// Test with nil request
	_, err := resolver.EnrichContext(nil)
	if err == nil {
		t.Error("Expected error for nil request")
	}

	// Test with non-existent subject
	invalidRequest := &models.EvaluationRequest{
		SubjectID:  "non-existent",
		ResourceID: "res-001",
		Action:     "read",
	}
	_, err = resolver.EnrichContext(invalidRequest)
	if err == nil {
		t.Error("Expected error for non-existent subject")
	}

	// Test with non-existent resource
	invalidRequest2 := &models.EvaluationRequest{
		SubjectID:  "sub-001",
		ResourceID: "non-existent",
		Action:     "read",
	}
	_, err = resolver.EnrichContext(invalidRequest2)
	if err == nil {
		t.Error("Expected error for non-existent resource")
	}

	// Test with non-existent action
	invalidRequest3 := &models.EvaluationRequest{
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "non-existent",
	}
	_, err = resolver.EnrichContext(invalidRequest3)
	if err == nil {
		t.Error("Expected error for non-existent action")
	}
}
