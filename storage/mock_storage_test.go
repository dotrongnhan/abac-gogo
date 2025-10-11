package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"abac_go_example/models"
)

func TestMockStorageInitialization(t *testing.T) {
	// Create temporary directory for test data
	tempDir, err := ioutil.TempDir("", "abac_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test data files
	createTestDataFiles(t, tempDir)

	// Initialize mock storage
	storage, err := NewMockStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize mock storage: %v", err)
	}

	// Test that data was loaded correctly
	subjects, err := storage.GetAllSubjects()
	if err != nil {
		t.Fatalf("Failed to get subjects: %v", err)
	}
	if len(subjects) != 2 {
		t.Errorf("Expected 2 subjects, got %d", len(subjects))
	}

	resources, err := storage.GetAllResources()
	if err != nil {
		t.Fatalf("Failed to get resources: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}

	actions, err := storage.GetAllActions()
	if err != nil {
		t.Fatalf("Failed to get actions: %v", err)
	}
	if len(actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(actions))
	}

	policies, err := storage.GetPolicies()
	if err != nil {
		t.Fatalf("Failed to get policies: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(policies))
	}
}

func TestGetSubject(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Test getting existing subject
	subject, err := storage.GetSubject("sub-001")
	if err != nil {
		t.Fatalf("Failed to get subject: %v", err)
	}

	if subject.ID != "sub-001" {
		t.Errorf("Expected subject ID sub-001, got %s", subject.ID)
	}

	if subject.ExternalID != "john.doe@company.com" {
		t.Errorf("Expected external ID john.doe@company.com, got %s", subject.ExternalID)
	}

	if subject.SubjectType != "user" {
		t.Errorf("Expected subject type user, got %s", subject.SubjectType)
	}

	// Test getting non-existent subject
	_, err = storage.GetSubject("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent subject")
	}
}

func TestGetResource(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Test getting existing resource
	resource, err := storage.GetResource("res-001")
	if err != nil {
		t.Fatalf("Failed to get resource: %v", err)
	}

	if resource.ID != "res-001" {
		t.Errorf("Expected resource ID res-001, got %s", resource.ID)
	}

	if resource.ResourceType != "api_endpoint" {
		t.Errorf("Expected resource type api_endpoint, got %s", resource.ResourceType)
	}

	if resource.ResourceID != "/api/v1/users" {
		t.Errorf("Expected resource ID /api/v1/users, got %s", resource.ResourceID)
	}

	// Test getting non-existent resource
	_, err = storage.GetResource("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent resource")
	}
}

func TestGetAction(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Test getting existing action
	action, err := storage.GetAction("read")
	if err != nil {
		t.Fatalf("Failed to get action: %v", err)
	}

	if action.ActionName != "read" {
		t.Errorf("Expected action name read, got %s", action.ActionName)
	}

	if action.ActionCategory != "crud" {
		t.Errorf("Expected action category crud, got %s", action.ActionCategory)
	}

	// Test getting non-existent action
	_, err = storage.GetAction("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent action")
	}
}

func TestGetPolicies(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	policies, err := storage.GetPolicies()
	if err != nil {
		t.Fatalf("Failed to get policies: %v", err)
	}

	if len(policies) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(policies))
	}

	// Check first policy
	policy := policies[0]
	if policy.ID != "pol-001" {
		t.Errorf("Expected policy ID pol-001, got %s", policy.ID)
	}

	if policy.Effect != "permit" {
		t.Errorf("Expected policy effect permit, got %s", policy.Effect)
	}

	if !policy.Enabled {
		t.Error("Expected policy to be enabled")
	}

	if len(policy.Rules) == 0 {
		t.Error("Expected policy to have rules")
	}
}

func TestInvalidDataFiles(t *testing.T) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "abac_test_invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create invalid JSON file
	invalidJSON := `{"subjects": [{"id": "sub-001", "invalid_json": }]}`
	err = ioutil.WriteFile(filepath.Join(tempDir, "subjects.json"), []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	// Create other required files with valid JSON
	createMinimalTestFiles(t, tempDir)

	// Try to initialize storage with invalid JSON
	_, err = NewMockStorage(tempDir)
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}

func TestMissingDataFiles(t *testing.T) {
	// Create temporary directory with no files
	tempDir, err := ioutil.TempDir("", "abac_test_missing")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Try to initialize storage with missing files
	_, err = NewMockStorage(tempDir)
	if err == nil {
		t.Error("Expected error when data files are missing")
	}
}

// Helper functions

func setupTestStorage(t *testing.T) (string, *MockStorage) {
	tempDir, err := ioutil.TempDir("", "abac_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	createTestDataFiles(t, tempDir)

	storage, err := NewMockStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize mock storage: %v", err)
	}

	return tempDir, storage
}

func createTestDataFiles(t *testing.T, dir string) {
	// Create subjects.json
	subjects := map[string]interface{}{
		"subjects": []models.Subject{
			{
				ID:          "sub-001",
				ExternalID:  "john.doe@company.com",
				SubjectType: "user",
				Metadata: map[string]interface{}{
					"full_name": "John Doe",
					"email":     "john.doe@company.com",
				},
				Attributes: map[string]interface{}{
					"department": "engineering",
					"role":       []string{"senior_developer"},
				},
			},
			{
				ID:          "sub-002",
				ExternalID:  "alice.smith@company.com",
				SubjectType: "user",
				Metadata: map[string]interface{}{
					"full_name": "Alice Smith",
					"email":     "alice.smith@company.com",
				},
				Attributes: map[string]interface{}{
					"department": "finance",
					"role":       []string{"accountant"},
				},
			},
		},
	}
	writeJSONFile(t, filepath.Join(dir, "subjects.json"), subjects)

	// Create resources.json
	resources := map[string]interface{}{
		"resources": []models.Resource{
			{
				ID:           "res-001",
				ResourceType: "api_endpoint",
				ResourceID:   "/api/v1/users",
				Path:         "api.v1.users",
				Attributes: map[string]interface{}{
					"data_classification": "internal",
				},
			},
			{
				ID:           "res-002",
				ResourceType: "database",
				ResourceID:   "prod-db-customers",
				Path:         "database.production.customers",
				Attributes: map[string]interface{}{
					"data_classification": "confidential",
				},
			},
		},
	}
	writeJSONFile(t, filepath.Join(dir, "resources.json"), resources)

	// Create actions.json
	actions := map[string]interface{}{
		"actions": []models.Action{
			{
				ID:             "act-001",
				ActionName:     "read",
				ActionCategory: "crud",
				Description:    "Read/View resource",
				IsSystem:       false,
			},
			{
				ID:             "act-002",
				ActionName:     "write",
				ActionCategory: "crud",
				Description:    "Create/Update resource",
				IsSystem:       false,
			},
		},
	}
	writeJSONFile(t, filepath.Join(dir, "actions.json"), actions)

	// Create policies.json
	policies := map[string]interface{}{
		"policies": []models.Policy{
			{
				ID:          "pol-001",
				PolicyName:  "Engineering Read Access",
				Description: "Allow engineering team to read technical resources",
				Effect:      "permit",
				Priority:    100,
				Enabled:     true,
				Version:     1,
				Rules: []models.PolicyRule{
					{
						TargetType:    "subject",
						AttributePath: "attributes.department",
						Operator:      "eq",
						ExpectedValue: "engineering",
					},
				},
				Actions:          []string{"read"},
				ResourcePatterns: []string{"/api/v1/*"},
			},
			{
				ID:          "pol-002",
				PolicyName:  "Finance Access",
				Description: "Allow finance team access",
				Effect:      "permit",
				Priority:    200,
				Enabled:     true,
				Version:     1,
				Rules: []models.PolicyRule{
					{
						TargetType:    "subject",
						AttributePath: "attributes.department",
						Operator:      "eq",
						ExpectedValue: "finance",
					},
				},
				Actions:          []string{"read", "write"},
				ResourcePatterns: []string{"/finance/*"},
			},
		},
	}
	writeJSONFile(t, filepath.Join(dir, "policies.json"), policies)
}

func createMinimalTestFiles(t *testing.T, dir string) {
	// Create minimal valid files for other tests
	writeJSONFile(t, filepath.Join(dir, "resources.json"), map[string]interface{}{"resources": []interface{}{}})
	writeJSONFile(t, filepath.Join(dir, "actions.json"), map[string]interface{}{"actions": []interface{}{}})
	writeJSONFile(t, filepath.Join(dir, "policies.json"), map[string]interface{}{"policies": []interface{}{}})
}

func writeJSONFile(t *testing.T, filename string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", filename, err)
	}
}

// Benchmark tests

func BenchmarkGetSubject(b *testing.B) {
	tempDir, storage := setupBenchmarkStorage(b)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetSubject("sub-001")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetResource(b *testing.B) {
	tempDir, storage := setupBenchmarkStorage(b)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetResource("res-001")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAction(b *testing.B) {
	tempDir, storage := setupBenchmarkStorage(b)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetAction("read")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAllSubjects(b *testing.B) {
	tempDir, storage := setupBenchmarkStorage(b)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetAllSubjects()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func setupBenchmarkStorage(b *testing.B) (string, *MockStorage) {
	tempDir, err := ioutil.TempDir("", "abac_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	createTestDataFiles(&testing.T{}, tempDir)

	storage, err := NewMockStorage(tempDir)
	if err != nil {
		b.Fatalf("Failed to initialize mock storage: %v", err)
	}

	return tempDir, storage
}
