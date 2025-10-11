package main

import (
	"testing"
	"time"

	"abac_go_example/audit"
	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func setupIntegrationTest(t *testing.T) (*storage.PostgreSQLStorage, *evaluator.PolicyDecisionPoint, *audit.AuditLogger) {
	// Try PostgreSQL first, fallback to mock storage if not available
	config := &storage.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		DatabaseName: "abac_test",
		SSLMode:      "disable",
		TimeZone:     "UTC",
	}

	pgStorage, err := storage.NewPostgreSQLStorage(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL integration tests - database not available: %v", err)
		return nil, nil, nil
	}

	// Clean up test data
	pgStorage.DeleteSubject("integration-sub-001")
	pgStorage.DeleteSubject("integration-sub-002")
	pgStorage.DeleteResource("integration-res-001")
	pgStorage.DeleteAction("integration-action-read")
	pgStorage.DeletePolicy("integration-policy-001")

	// Seed test data
	seedIntegrationTestData(t, pgStorage)

	auditLogger, err := audit.NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to initialize audit logger: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(pgStorage)

	return pgStorage, pdp, auditLogger
}

func seedIntegrationTestData(t *testing.T, storage *storage.PostgreSQLStorage) {
	// Create test subject
	testSubject := &models.Subject{
		ID:          "integration-sub-001",
		ExternalID:  "john.doe@company.com",
		SubjectType: "user",
		Metadata: models.JSONMap{
			"name":       "John Doe",
			"department": "engineering",
		},
		Attributes: models.JSONMap{
			"role":            "senior_developer",
			"clearance_level": 2,
			"team":            "backend",
		},
	}
	if err := storage.CreateSubject(testSubject); err != nil {
		t.Fatalf("Failed to create test subject: %v", err)
	}

	// Create probation user for negative tests
	probationSubject := &models.Subject{
		ID:          "integration-sub-002",
		ExternalID:  "bob.wilson@company.com",
		SubjectType: "user",
		Metadata: models.JSONMap{
			"name":       "Bob Wilson",
			"department": "engineering",
		},
		Attributes: models.JSONMap{
			"role":            "junior_developer",
			"clearance_level": 1,
			"status":          "probation",
		},
	}
	if err := storage.CreateSubject(probationSubject); err != nil {
		t.Fatalf("Failed to create probation subject: %v", err)
	}

	// Create test resource
	testResource := &models.Resource{
		ID:           "integration-res-001",
		ResourceType: "api_endpoint",
		ResourceID:   "/api/v1/users",
		Path:         "/api/v1/users",
		Metadata: models.JSONMap{
			"classification": "internal",
			"owner":          "engineering",
		},
		Attributes: models.JSONMap{
			"sensitivity":        "medium",
			"required_clearance": 2,
		},
	}
	if err := storage.CreateResource(testResource); err != nil {
		t.Fatalf("Failed to create test resource: %v", err)
	}

	// Create test action
	testAction := &models.Action{
		ID:             "integration-action-read",
		ActionName:     "read",
		ActionCategory: "data_access",
		Description:    "Read data access",
		IsSystem:       false,
	}
	if err := storage.CreateAction(testAction); err != nil {
		t.Fatalf("Failed to create test action: %v", err)
	}

	// Create test policy
	testPolicy := &models.Policy{
		ID:          "integration-policy-001",
		PolicyName:  "Engineering Read Access",
		Description: "Allow engineering team to read internal resources",
		Effect:      "permit",
		Priority:    100,
		Enabled:     true,
		Version:     1,
		Conditions: models.JSONMap{
			"department": "engineering",
		},
		Rules: models.JSONPolicyRules{
			{
				ID:            "rule-dept",
				TargetType:    "subject",
				AttributePath: "attributes.role",
				Operator:      "in",
				ExpectedValue: []string{"senior_developer", "lead_developer"},
				IsNegative:    false,
				RuleOrder:     1,
			},
			{
				ID:            "rule-clearance",
				TargetType:    "subject",
				AttributePath: "attributes.clearance_level",
				Operator:      "gte",
				ExpectedValue: 2,
				IsNegative:    false,
				RuleOrder:     2,
			},
			{
				ID:            "rule-not-probation",
				TargetType:    "subject",
				AttributePath: "attributes.status",
				Operator:      "neq",
				ExpectedValue: "probation",
				IsNegative:    false,
				RuleOrder:     3,
			},
		},
		Actions:          models.JSONStringSlice{"read"},
		ResourcePatterns: models.JSONStringSlice{"/api/v1/*"},
	}
	if err := storage.CreatePolicy(testPolicy); err != nil {
		t.Fatalf("Failed to create test policy: %v", err)
	}
}

func TestPostgreSQLFullSystemIntegration(t *testing.T) {
	pgStorage, pdp, auditLogger := setupIntegrationTest(t)
	if pgStorage == nil {
		return // Test was skipped
	}
	defer pgStorage.Close()
	defer auditLogger.Close()

	// Test scenarios
	testScenarios := []struct {
		name             string
		request          *models.EvaluationRequest
		expectedDecision string
	}{
		{
			name: "Senior Developer Read Access - Should Permit",
			request: &models.EvaluationRequest{
				RequestID:  "pg-integration-001",
				SubjectID:  "integration-sub-001",
				ResourceID: "integration-res-001",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"source_ip": "10.0.1.50",
				},
			},
			expectedDecision: "permit",
		},
		{
			name: "Probation User Read Access - Should Deny",
			request: &models.EvaluationRequest{
				RequestID:  "pg-integration-002",
				SubjectID:  "integration-sub-002",
				ResourceID: "integration-res-001",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"source_ip": "10.0.1.60",
				},
			},
			expectedDecision: "deny",
		},
		{
			name: "Non-existent Subject - Should Error",
			request: &models.EvaluationRequest{
				RequestID:  "pg-integration-003",
				SubjectID:  "non-existent-subject",
				ResourceID: "integration-res-001",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"source_ip": "10.0.1.70",
				},
			},
			expectedDecision: "error", // This should result in an error
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			decision, err := pdp.Evaluate(scenario.request)

			if scenario.expectedDecision == "error" {
				if err == nil {
					t.Errorf("Expected error for scenario %s, but got decision: %s", scenario.name, decision.Result)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for scenario %s: %v", scenario.name, err)
			}

			if decision.Result != scenario.expectedDecision {
				t.Errorf("Scenario %s: expected decision %s, got %s",
					scenario.name, scenario.expectedDecision, decision.Result)
			}

			// Log the evaluation for audit
			subject, _ := pgStorage.GetSubject(scenario.request.SubjectID)
			resource, _ := pgStorage.GetResource(scenario.request.ResourceID)
			action, _ := pgStorage.GetAction(scenario.request.Action)

			auditContext := &models.EvaluationContext{
				Subject:     subject,
				Resource:    resource,
				Action:      action,
				Environment: scenario.request.Context,
				Timestamp:   time.Now(),
			}

			auditLogger.LogEvaluation(scenario.request, decision, auditContext)
		})
	}
}

func TestPostgreSQLBatchEvaluation(t *testing.T) {
	pgStorage, pdp, auditLogger := setupIntegrationTest(t)
	if pgStorage == nil {
		return
	}
	defer pgStorage.Close()
	defer auditLogger.Close()

	// Create multiple evaluation requests
	requests := []*models.EvaluationRequest{
		{
			RequestID:  "batch-001",
			SubjectID:  "integration-sub-001",
			ResourceID: "integration-res-001",
			Action:     "read",
			Context:    map[string]interface{}{"timestamp": time.Now().Format(time.RFC3339)},
		},
		{
			RequestID:  "batch-002",
			SubjectID:  "integration-sub-002",
			ResourceID: "integration-res-001",
			Action:     "read",
			Context:    map[string]interface{}{"timestamp": time.Now().Format(time.RFC3339)},
		},
	}

	// Test batch evaluation
	startTime := time.Now()
	decisions, err := pdp.BatchEvaluate(requests)
	batchTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Batch evaluation failed: %v", err)
	}

	if len(decisions) != len(requests) {
		t.Errorf("Expected %d decisions, got %d", len(requests), len(decisions))
	}

	t.Logf("Batch evaluation of %d requests completed in %v", len(requests), batchTime)

	// Verify individual decisions
	if decisions[0].Result != "permit" {
		t.Errorf("Expected first decision to be 'permit', got %s", decisions[0].Result)
	}
	if decisions[1].Result != "deny" {
		t.Errorf("Expected second decision to be 'deny', got %s", decisions[1].Result)
	}
}

func TestPostgreSQLPolicyExplanation(t *testing.T) {
	pgStorage, pdp, auditLogger := setupIntegrationTest(t)
	if pgStorage == nil {
		return
	}
	defer pgStorage.Close()
	defer auditLogger.Close()

	request := &models.EvaluationRequest{
		RequestID:  "explanation-001",
		SubjectID:  "integration-sub-001",
		ResourceID: "integration-res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	explanation, err := pdp.ExplainDecision(request)
	if err != nil {
		t.Fatalf("Failed to get explanation: %v", err)
	}

	if explanation == nil {
		t.Fatal("Expected explanation but got nil")
	}

	t.Logf("Decision explanation received for request %s", request.RequestID)
}

func TestPostgreSQLAuditLogging(t *testing.T) {
	pgStorage, pdp, auditLogger := setupIntegrationTest(t)
	if pgStorage == nil {
		return
	}
	defer pgStorage.Close()
	defer auditLogger.Close()

	request := &models.EvaluationRequest{
		RequestID:  "audit-test-001",
		SubjectID:  "integration-sub-001",
		ResourceID: "integration-res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"source_ip": "10.0.1.100",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// Create audit log entry
	auditLog := &models.AuditLog{
		RequestID:    request.RequestID,
		SubjectID:    request.SubjectID,
		ResourceID:   request.ResourceID,
		ActionID:     request.Action,
		Decision:     decision.Result,
		EvaluationMs: decision.EvaluationTimeMs,
		Context:      models.JSONMap(request.Context),
	}

	err = pgStorage.LogAudit(auditLog)
	if err != nil {
		t.Fatalf("Failed to log audit: %v", err)
	}

	// Verify audit log was stored
	auditLogs, err := pgStorage.GetAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("Failed to get audit logs: %v", err)
	}

	found := false
	for _, log := range auditLogs {
		if log.RequestID == "audit-test-001" {
			found = true
			if log.Decision != decision.Result {
				t.Errorf("Expected audit decision %s, got %s", decision.Result, log.Decision)
			}
			break
		}
	}

	if !found {
		t.Error("Audit log not found in database")
	}
}

func BenchmarkPostgreSQLIntegrationEvaluation(b *testing.B) {
	config := &storage.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		DatabaseName: "abac_test",
		SSLMode:      "disable",
		TimeZone:     "UTC",
	}

	pgStorage, err := storage.NewPostgreSQLStorage(config)
	if err != nil {
		b.Skipf("Skipping PostgreSQL benchmark - database not available: %v", err)
		return
	}
	defer pgStorage.Close()

	pdp := evaluator.NewPolicyDecisionPoint(pgStorage)

	request := &models.EvaluationRequest{
		RequestID:  "benchmark-001",
		SubjectID:  "integration-sub-001",
		ResourceID: "integration-res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	// Ensure test data exists
	seedIntegrationTestData(&testing.T{}, pgStorage)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pdp.Evaluate(request)
		if err != nil {
			b.Fatalf("Evaluation failed: %v", err)
		}
	}
}
