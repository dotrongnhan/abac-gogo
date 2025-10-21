package storage

import (
	"testing"
	"time"

	"abac_go_example/models"
)

// Test database configuration for testing
func getTestDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:         getEnv("TEST_DB_HOST", "localhost"),
		Port:         getEnvAsInt("TEST_DB_PORT", 5432),
		User:         getEnv("TEST_DB_USER", "postgres"),
		Password:     getEnv("TEST_DB_PASSWORD", "postgres"),
		DatabaseName: getEnv("TEST_DB_NAME", "abac_test"),
		SSLMode:      getEnv("TEST_DB_SSL_MODE", "disable"),
		TimeZone:     getEnv("TEST_DB_TIMEZONE", "UTC"),
	}
}

func setupTestDatabase(t *testing.T) *PostgreSQLStorage {
	config := getTestDatabaseConfig()

	// Try to connect to test database
	storage, err := NewPostgreSQLStorage(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL tests - database not available: %v", err)
		return nil
	}

	// Clean up existing test data
	cleanupTestData(t, storage)

	return storage
}

func cleanupTestData(t *testing.T, storage *PostgreSQLStorage) {
	// Delete test data in reverse dependency order
	storage.db.Exec("DELETE FROM audit_logs WHERE request_id LIKE 'test-%'")
	storage.db.Exec("DELETE FROM policies WHERE id LIKE 'test-%'")
	storage.db.Exec("DELETE FROM actions WHERE id LIKE 'test-%'")
	storage.db.Exec("DELETE FROM resources WHERE id LIKE 'test-%'")
	storage.db.Exec("DELETE FROM subjects WHERE id LIKE 'test-%'")
}

func TestPostgreSQLStorageSubjects(t *testing.T) {
	storage := setupTestDatabase(t)
	if storage == nil {
		return
	}
	defer storage.Close()

	// Test data
	testSubject := &models.Subject{
		ID:          "test-subject-001",
		ExternalID:  "ext-test-001",
		SubjectType: "user",
		Metadata: models.JSONMap{
			"department": "engineering",
			"location":   "office",
		},
		Attributes: models.JSONMap{
			"clearance_level": 2,
			"role":            "developer",
		},
	}

	// Test Create
	err := storage.CreateSubject(testSubject)
	if err != nil {
		t.Fatalf("Failed to create subject: %v", err)
	}

	// Test Get
	retrievedSubject, err := storage.GetSubject("test-subject-001")
	if err != nil {
		t.Fatalf("Failed to get subject: %v", err)
	}

	if retrievedSubject.ID != testSubject.ID {
		t.Errorf("Expected ID %s, got %s", testSubject.ID, retrievedSubject.ID)
	}
	if retrievedSubject.SubjectType != testSubject.SubjectType {
		t.Errorf("Expected SubjectType %s, got %s", testSubject.SubjectType, retrievedSubject.SubjectType)
	}

	// Test metadata and attributes
	if retrievedSubject.Metadata["department"] != "engineering" {
		t.Errorf("Expected department 'engineering', got %v", retrievedSubject.Metadata["department"])
	}
	if retrievedSubject.Attributes["role"] != "developer" {
		t.Errorf("Expected role 'developer', got %v", retrievedSubject.Attributes["role"])
	}

	// Test Update
	testSubject.Attributes["role"] = "senior_developer"
	err = storage.UpdateSubject(testSubject)
	if err != nil {
		t.Fatalf("Failed to update subject: %v", err)
	}

	updatedSubject, err := storage.GetSubject("test-subject-001")
	if err != nil {
		t.Fatalf("Failed to get updated subject: %v", err)
	}
	if updatedSubject.Attributes["role"] != "senior_developer" {
		t.Errorf("Expected updated role 'senior_developer', got %v", updatedSubject.Attributes["role"])
	}

	// Test GetAllSubjects
	subjects, err := storage.GetAllSubjects()
	if err != nil {
		t.Fatalf("Failed to get all subjects: %v", err)
	}

	found := false
	for _, subject := range subjects {
		if subject.ID == "test-subject-001" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Test subject not found in GetAllSubjects result")
	}

	// Test Delete
	err = storage.DeleteSubject("test-subject-001")
	if err != nil {
		t.Fatalf("Failed to delete subject: %v", err)
	}

	// Verify deletion
	_, err = storage.GetSubject("test-subject-001")
	if err == nil {
		t.Error("Expected error when getting deleted subject")
	}
}

func TestPostgreSQLStorageResources(t *testing.T) {
	storage := setupTestDatabase(t)
	if storage == nil {
		return
	}
	defer storage.Close()

	testResource := &models.Resource{
		ID:           "test-resource-001",
		ResourceType: "api_endpoint",
		ResourceID:   "/api/v1/test",
		Path:         "/api/v1/test",
		Metadata: models.JSONMap{
			"classification": "internal",
			"owner":          "engineering",
		},
		Attributes: models.JSONMap{
			"sensitivity": "medium",
			"data_type":   "user_data",
		},
	}

	// Test Create
	err := storage.CreateResource(testResource)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	// Test Get
	retrievedResource, err := storage.GetResource("test-resource-001")
	if err != nil {
		t.Fatalf("Failed to get resource: %v", err)
	}

	if retrievedResource.ResourceType != testResource.ResourceType {
		t.Errorf("Expected ResourceType %s, got %s", testResource.ResourceType, retrievedResource.ResourceType)
	}

	// Test GetAllResources
	resources, err := storage.GetAllResources()
	if err != nil {
		t.Fatalf("Failed to get all resources: %v", err)
	}

	found := false
	for _, resource := range resources {
		if resource.ID == "test-resource-001" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Test resource not found in GetAllResources result")
	}

	// Clean up
	storage.DeleteResource("test-resource-001")
}

func TestPostgreSQLStorageActions(t *testing.T) {
	storage := setupTestDatabase(t)
	if storage == nil {
		return
	}
	defer storage.Close()

	testAction := &models.Action{
		ID:             "test-action-001",
		ActionName:     "test_read",
		ActionCategory: "data_access",
		Description:    "Test read operation",
		IsSystem:       false,
	}

	// Test Create
	err := storage.CreateAction(testAction)
	if err != nil {
		t.Fatalf("Failed to create action: %v", err)
	}

	// Test Get by name
	retrievedAction, err := storage.GetAction("test_read")
	if err != nil {
		t.Fatalf("Failed to get action: %v", err)
	}

	if retrievedAction.ActionCategory != testAction.ActionCategory {
		t.Errorf("Expected ActionCategory %s, got %s", testAction.ActionCategory, retrievedAction.ActionCategory)
	}

	// Test GetAllActions
	actions, err := storage.GetAllActions()
	if err != nil {
		t.Fatalf("Failed to get all actions: %v", err)
	}

	found := false
	for _, action := range actions {
		if action.ID == "test-action-001" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Test action not found in GetAllActions result")
	}

	// Clean up
	storage.DeleteAction("test-action-001")
}

func TestPostgreSQLStoragePolicies(t *testing.T) {
	storage := setupTestDatabase(t)
	if storage == nil {
		return
	}
	defer storage.Close()

	testPolicy := &models.Policy{
		ID:          "test-policy-001",
		PolicyName:  "Test Policy",
		Description: "Test policy for unit testing",
		Effect:      "permit",
		Priority:    100,
		Enabled:     true,
		Version:     1,
		Conditions: models.JSONMap{
			"time_range": "business_hours",
		},
		Rules: models.JSONPolicyRules{
			{
				ID:            "rule-001",
				TargetType:    "subject",
				AttributePath: "attributes.role",
				Operator:      "eq",
				ExpectedValue: "developer",
				IsNegative:    false,
				RuleOrder:     1,
			},
		},
		Actions:          models.JSONStringSlice{"read", "write"},
		ResourcePatterns: models.JSONStringSlice{"/api/v1/*"},
	}

	// Test Create
	err := storage.CreatePolicy(testPolicy)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	// Test GetPolicies
	policies, err := storage.GetPolicies()
	if err != nil {
		t.Fatalf("Failed to get policies: %v", err)
	}

	found := false
	for _, policy := range policies {
		if policy.ID == "test-policy-001" {
			found = true
			if policy.Effect != "permit" {
				t.Errorf("Expected Effect 'permit', got %s", policy.Effect)
			}
			if len(policy.Rules) != 1 {
				t.Errorf("Expected 1 rule, got %d", len(policy.Rules))
			}
			if len(policy.Actions) != 2 {
				t.Errorf("Expected 2 actions, got %d", len(policy.Actions))
			}
			break
		}
	}
	if !found {
		t.Error("Test policy not found in GetPolicies result")
	}

	// Clean up
	storage.DeletePolicy("test-policy-001")
}

func TestPostgreSQLStorageAuditLogs(t *testing.T) {
	storage := setupTestDatabase(t)
	if storage == nil {
		return
	}
	defer storage.Close()

	testAuditLog := &models.AuditLog{
		RequestID:    "test-request-001",
		SubjectID:    "test-subject-001",
		ResourceID:   "test-resource-001",
		ActionID:     "test-action-001",
		Decision:     "permit",
		EvaluationMs: 15,
		Context: models.JSONMap{
			"source_ip": "127.0.0.1",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	// Test Create
	err := storage.LogAudit(testAuditLog)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// Test GetAuditLogs
	auditLogs, err := storage.GetAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("Failed to get audit logs: %v", err)
	}

	found := false
	for _, log := range auditLogs {
		if log.RequestID == "test-request-001" {
			found = true
			if log.Decision != "permit" {
				t.Errorf("Expected Decision 'permit', got %s", log.Decision)
			}
			if log.EvaluationMs != 15 {
				t.Errorf("Expected EvaluationMs 15, got %d", log.EvaluationMs)
			}
			break
		}
	}
	if !found {
		t.Error("Test audit log not found in GetAuditLogs result")
	}
}

func TestPostgreSQLStorageConnectionHandling(t *testing.T) {
	storage := setupTestDatabase(t)
	if storage == nil {
		return
	}

	// Test that connection is working
	subjects, err := storage.GetAllSubjects()
	if err != nil {
		t.Fatalf("Failed to test connection: %v", err)
	}

	t.Logf("Connection test successful, found %d subjects", len(subjects))

	// Test Close
	err = storage.Close()
	if err != nil {
		t.Fatalf("Failed to close storage: %v", err)
	}
}

// Benchmark tests
func BenchmarkPostgreSQLStorageSubjectGet(b *testing.B) {
	storage := setupTestDatabase(&testing.T{})
	if storage == nil {
		b.Skip("Database not available")
		return
	}
	defer storage.Close()

	// Create test subject
	testSubject := &models.Subject{
		ID:          "bench-subject-001",
		ExternalID:  "bench-ext-001",
		SubjectType: "user",
		Metadata:    models.JSONMap{"test": "data"},
		Attributes:  models.JSONMap{"role": "developer"},
	}
	storage.CreateSubject(testSubject)
	defer storage.DeleteSubject("bench-subject-001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetSubject("bench-subject-001")
		if err != nil {
			b.Fatalf("Failed to get subject: %v", err)
		}
	}
}

func BenchmarkPostgreSQLStoragePolicyGet(b *testing.B) {
	storage := setupTestDatabase(&testing.T{})
	if storage == nil {
		b.Skip("Database not available")
		return
	}
	defer storage.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetPolicies()
		if err != nil {
			b.Fatalf("Failed to get policies: %v", err)
		}
	}
}
