package storage

import (
	"os"
	"testing"

	"abac_go_example/models"
)

// TestDatabaseConfig returns a test database configuration
func TestDatabaseConfig() *DatabaseConfig {
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

// NewTestStorage creates a PostgreSQL storage instance for testing
func NewTestStorage(t *testing.T) *PostgreSQLStorage {
	// Skip test if no database available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database tests (SKIP_DB_TESTS=true)")
	}

	config := TestDatabaseConfig()
	storage, err := NewPostgreSQLStorage(config)
	if err != nil {
		t.Skipf("Failed to create test storage (database not available): %v", err)
	}

	// Clean up tables for fresh test
	cleanupTestTables(t, storage)

	return storage
}

// cleanupTestTables cleans up test data
func cleanupTestTables(t *testing.T, storage *PostgreSQLStorage) {
	// Delete all test data
	storage.db.Exec("DELETE FROM audit_logs")
	storage.db.Exec("DELETE FROM policies")
	storage.db.Exec("DELETE FROM actions")
	storage.db.Exec("DELETE FROM resources")
	storage.db.Exec("DELETE FROM subjects")
}

// SeedTestData seeds the database with test data
func SeedTestData(t *testing.T, storage *PostgreSQLStorage) {
	// Create test subjects
	subjects := []*models.Subject{
		{
			ID:          "sub-001",
			ExternalID:  "john.doe@company.com",
			SubjectType: "user",
			Metadata: models.JSONMap{
				"full_name":   "John Doe",
				"email":       "john.doe@company.com",
				"employee_id": "EMP-12345",
			},
			Attributes: models.JSONMap{
				"department":       "engineering",
				"role":             []string{"senior_developer", "code_reviewer"},
				"clearance_level":  3,
				"location":         "VN-HCM",
				"team":             "platform",
				"years_of_service": 5,
			},
		},
		{
			ID:          "sub-002",
			ExternalID:  "alice.smith@company.com",
			SubjectType: "user",
			Metadata: models.JSONMap{
				"full_name":   "Alice Smith",
				"email":       "alice.smith@company.com",
				"employee_id": "EMP-23456",
			},
			Attributes: models.JSONMap{
				"department":       "finance",
				"role":             []string{"accountant", "report_viewer"},
				"clearance_level":  2,
				"location":         "VN-HN",
				"team":             "accounts",
				"years_of_service": 3,
			},
		},
		{
			ID:          "sub-004",
			ExternalID:  "bob.wilson@company.com",
			SubjectType: "user",
			Metadata: models.JSONMap{
				"full_name":   "Bob Wilson",
				"email":       "bob.wilson@company.com",
				"employee_id": "EMP-34567",
			},
			Attributes: models.JSONMap{
				"department":       "engineering",
				"role":             []string{"junior_developer"},
				"clearance_level":  1,
				"location":         "VN-DN",
				"team":             "frontend",
				"years_of_service": 1,
				"on_probation":     true,
			},
		},
	}

	for _, subject := range subjects {
		if err := storage.CreateSubject(subject); err != nil {
			t.Fatalf("Failed to create test subject: %v", err)
		}
	}

	// Create test resources
	resources := []*models.Resource{
		{
			ID:           "res-001",
			ResourceType: "api_endpoint",
			ResourceID:   "/api/v1/users",
			Path:         "api.v1.users",
			Metadata: models.JSONMap{
				"description": "User management API",
				"version":     "v1",
			},
			Attributes: models.JSONMap{
				"data_classification": "internal",
				"methods":             []string{"GET", "POST", "PUT", "DELETE"},
				"rate_limit":          1000,
				"requires_auth":       true,
				"pii_data":            true,
			},
		},
		{
			ID:           "res-001-financial",
			ResourceType: "api_endpoint",
			ResourceID:   "/api/v1/financial",
			Path:         "api.v1.financial",
			Metadata: models.JSONMap{
				"description": "Financial data API",
				"version":     "v1",
			},
			Attributes: models.JSONMap{
				"data_classification": "highly_confidential",
				"methods":             []string{"GET"},
				"rate_limit":          100,
				"requires_auth":       true,
				"department":          "finance",
			},
		},
	}

	for _, resource := range resources {
		if err := storage.CreateResource(resource); err != nil {
			t.Fatalf("Failed to create test resource: %v", err)
		}
	}

	// Create test actions
	actions := []*models.Action{
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
		{
			ID:             "act-006",
			ActionName:     "admin",
			ActionCategory: "system",
			Description:    "Administrative access",
			IsSystem:       true,
		},
	}

	for _, action := range actions {
		if err := storage.CreateAction(action); err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}
	}

	// Create test policies using new format
	policies := []*models.Policy{
		{
			ID:          "pol-001",
			PolicyName:  "Engineering Read Access",
			Description: "Allow engineering team to read technical resources",
			Effect:      "permit",
			Version:     "2024-10-21",
			Enabled:     true,
			Statement: models.JSONStatements{
				models.PolicyStatement{
					Sid:    "EngineeringReadAccess",
					Effect: "Allow",
					Action: models.JSONActionResource{
						Single: "document-service:file:read",
					},
					Resource: models.JSONActionResource{
						Single: "api:documents:dept:engineering/*",
					},
					Condition: models.JSONMap{
						"StringEquals": map[string]interface{}{
							"user.department": "engineering",
						},
					},
				},
			},
		},
		{
			ID:          "pol-004",
			PolicyName:  "Deny Probation Write",
			Description: "Deny write access for employees on probation",
			Effect:      "deny",
			Version:     "2024-10-21",
			Enabled:     true,
			Statement: models.JSONStatements{
				models.PolicyStatement{
					Sid:    "DenyProbationWrite",
					Effect: "Deny",
					Action: models.JSONActionResource{
						Multiple: []string{"document-service:file:write", "document-service:file:delete"},
					},
					Resource: models.JSONActionResource{
						Single: "*",
					},
					Condition: models.JSONMap{
						"Bool": map[string]interface{}{
							"user.on_probation": true,
						},
					},
				},
			},
		},
	}

	for _, policy := range policies {
		if err := storage.CreatePolicy(policy); err != nil {
			t.Fatalf("Failed to create test policy: %v", err)
		}
	}
}

// CleanupTestStorage cleans up and closes test storage
func CleanupTestStorage(t *testing.T, storage *PostgreSQLStorage) {
	if storage != nil {
		cleanupTestTables(t, storage)
		storage.Close()
	}
}
