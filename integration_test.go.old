package main

import (
	"testing"
	"time"

	"abac_go_example/audit"
	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func TestFullSystemIntegration(t *testing.T) {
	// Initialize the full system with PostgreSQL storage
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)

	// Seed test data
	storage.SeedTestData(t, testStorage)

	auditLogger, err := audit.NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to initialize audit logger: %v", err)
	}
	defer auditLogger.Close()

	pdp := evaluator.NewPolicyDecisionPoint(testStorage)

	// Test scenarios from the JSON data
	testScenarios := []struct {
		name             string
		request          *models.EvaluationRequest
		expectedDecision string
	}{
		{
			name: "Engineering Read Access",
			request: &models.EvaluationRequest{
				RequestID:  "integration-001",
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp":   "2024-01-15T14:00:00Z",
					"source_ip":   "10.0.1.50",
					"time_of_day": "14:00",
				},
			},
			expectedDecision: "permit",
		},
		{
			name: "Probation Write Denial",
			request: &models.EvaluationRequest{
				RequestID:  "integration-002",
				SubjectID:  "sub-004",
				ResourceID: "res-002",
				Action:     "write",
				Context: map[string]interface{}{
					"timestamp": "2024-01-15T15:00:00Z",
					"source_ip": "10.0.2.100",
				},
			},
			expectedDecision: "deny",
		},
		{
			name: "Finance Confidential Access",
			request: &models.EvaluationRequest{
				RequestID:  "integration-003",
				SubjectID:  "sub-002",
				ResourceID: "res-003",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": "2024-01-15T09:00:00Z",
					"source_ip": "192.168.1.100",
				},
			},
			expectedDecision: "permit",
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Perform evaluation
			decision, err := pdp.Evaluate(scenario.request)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			// Check decision
			if decision.Result != scenario.expectedDecision {
				t.Errorf("Expected decision %s, got %s", scenario.expectedDecision, decision.Result)
			}

			// Log the evaluation
			subject, _ := testStorage.GetSubject(scenario.request.SubjectID)
			resource, _ := testStorage.GetResource(scenario.request.ResourceID)
			action, _ := testStorage.GetAction(scenario.request.Action)

			auditContext := &models.EvaluationContext{
				Subject:     subject,
				Resource:    resource,
				Action:      action,
				Environment: scenario.request.Context,
				Timestamp:   time.Now(),
			}

			err = auditLogger.LogEvaluation(scenario.request, decision, auditContext)
			if err != nil {
				t.Errorf("Failed to log evaluation: %v", err)
			}

			// Verify performance
			if decision.EvaluationTimeMs > 100 {
				t.Errorf("Evaluation took too long: %dms", decision.EvaluationTimeMs)
			}

			t.Logf("✅ %s: %s (%dms)", scenario.name, decision.Result, decision.EvaluationTimeMs)
		})
	}
}

func TestSecurityScenarios(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)

	auditLogger, err := audit.NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to initialize audit logger: %v", err)
	}
	defer auditLogger.Close()

	pdp := evaluator.NewPolicyDecisionPoint(testStorage)

	securityTests := []struct {
		name        string
		request     *models.EvaluationRequest
		expectDeny  bool
		description string
	}{
		{
			name: "After Hours Access",
			request: &models.EvaluationRequest{
				RequestID:  "security-001",
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "write",
				Context: map[string]interface{}{
					"timestamp":   "2024-01-15T22:00:00Z",
					"time_of_day": "22:00",
					"source_ip":   "10.0.1.50",
				},
			},
			expectDeny:  true,
			description: "Senior developer trying to write after business hours",
		},
		{
			name: "External IP Access",
			request: &models.EvaluationRequest{
				RequestID:  "security-002",
				SubjectID:  "sub-003",
				ResourceID: "res-001",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": "2024-01-15T14:00:00Z",
					"source_ip": "203.0.113.1",
				},
			},
			expectDeny:  true,
			description: "Service trying to access from external IP",
		},
		{
			name: "Privilege Escalation",
			request: &models.EvaluationRequest{
				RequestID:  "security-003",
				SubjectID:  "sub-004",
				ResourceID: "res-003",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": "2024-01-15T14:00:00Z",
					"source_ip": "10.0.1.200",
				},
			},
			expectDeny:  true,
			description: "Junior developer trying to access financial documents",
		},
	}

	for _, test := range securityTests {
		t.Run(test.name, func(t *testing.T) {
			decision, err := pdp.Evaluate(test.request)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			if test.expectDeny && decision.Result != "deny" && decision.Result != "not_applicable" {
				t.Errorf("Expected deny/not_applicable for security test, got %s", decision.Result)
			}

			// Log security event if access was blocked
			if decision.Result == "deny" || decision.Result == "not_applicable" {
				auditLogger.LogSecurityEvent("security_test_blocked", test.request.SubjectID, map[string]interface{}{
					"test_name":   test.name,
					"description": test.description,
					"decision":    decision.Result,
					"reason":      decision.Reason,
				})
			}

			t.Logf("🔒 %s: %s - %s", test.name, decision.Result, test.description)
		})
	}
}

func TestDataConsistency(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)

	// Test that all referenced entities exist
	policies, err := testStorage.GetPolicies()
	if err != nil {
		t.Fatalf("Failed to get policies: %v", err)
	}

	for _, policy := range policies {
		// Check that all actions in policy exist
		for _, actionName := range policy.Actions {
			_, err := testStorage.GetAction(actionName)
			if err != nil {
				t.Errorf("Policy %s references non-existent action: %s", policy.ID, actionName)
			}
		}

		// Validate policy structure
		if policy.ID == "" {
			t.Error("Policy should have an ID")
		}

		if policy.Effect != "permit" && policy.Effect != "deny" {
			t.Errorf("Policy %s has invalid effect: %s", policy.ID, policy.Effect)
		}

		if len(policy.Rules) == 0 {
			t.Errorf("Policy %s should have at least one rule", policy.ID)
		}

		// Validate rules
		for _, rule := range policy.Rules {
			validTargets := []string{"subject", "resource", "action", "environment"}
			found := false
			for _, validTarget := range validTargets {
				if rule.TargetType == validTarget {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Policy %s has rule with invalid target type: %s", policy.ID, rule.TargetType)
			}

			if rule.AttributePath == "" {
				t.Errorf("Policy %s has rule with empty attribute path", policy.ID)
			}

			if rule.Operator == "" {
				t.Errorf("Policy %s has rule with empty operator", policy.ID)
			}
		}
	}

	// Test subjects
	subjects, err := testStorage.GetAllSubjects()
	if err != nil {
		t.Fatalf("Failed to get subjects: %v", err)
	}

	for _, subject := range subjects {
		if subject.ID == "" {
			t.Error("Subject should have an ID")
		}

		if subject.SubjectType == "" {
			t.Error("Subject should have a type")
		}

		validTypes := []string{"user", "service", "application", "device"}
		found := false
		for _, validType := range validTypes {
			if subject.SubjectType == validType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Subject %s has invalid type: %s", subject.ID, subject.SubjectType)
		}
	}

	// Test resources
	resources, err := testStorage.GetAllResources()
	if err != nil {
		t.Fatalf("Failed to get resources: %v", err)
	}

	for _, resource := range resources {
		if resource.ID == "" {
			t.Error("Resource should have an ID")
		}

		if resource.ResourceType == "" {
			t.Error("Resource should have a type")
		}

		if resource.ResourceID == "" {
			t.Error("Resource should have a resource ID")
		}
	}

	t.Logf("✅ Data consistency check passed: %d policies, %d subjects, %d resources",
		len(policies), len(subjects), len(resources))
}

func TestConcurrentEvaluations(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)

	pdp := evaluator.NewPolicyDecisionPoint(testStorage)

	// Create multiple evaluation requests
	requests := make([]*models.EvaluationRequest, 100)
	for i := 0; i < 100; i++ {
		requests[i] = &models.EvaluationRequest{
			RequestID:  "concurrent-" + string(rune(i)),
			SubjectID:  "sub-001",
			ResourceID: "res-001",
			Action:     "read",
			Context: map[string]interface{}{
				"timestamp":   "2024-01-15T14:00:00Z",
				"time_of_day": "14:00",
				"source_ip":   "10.0.1.50",
			},
		}
	}

	// Run concurrent evaluations
	start := time.Now()
	decisions, err := pdp.BatchEvaluate(requests)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Batch evaluation failed: %v", err)
	}

	if len(decisions) != len(requests) {
		t.Errorf("Expected %d decisions, got %d", len(requests), len(decisions))
	}

	// Verify all decisions are consistent
	for i, decision := range decisions {
		if decision.Result != "permit" {
			t.Errorf("Request %d: Expected permit, got %s", i, decision.Result)
		}
	}

	avgTime := elapsed / time.Duration(len(requests))
	t.Logf("⚡ Processed %d concurrent evaluations in %v (avg: %v per request)",
		len(requests), elapsed, avgTime)

	// Performance check
	if avgTime > 1*time.Millisecond {
		t.Errorf("Average evaluation time too slow: %v", avgTime)
	}
}
