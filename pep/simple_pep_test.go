package pep

import (
	"context"
	"testing"
	"time"

	"abac_go_example/evaluator/core"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func TestSimplePolicyEnforcementPoint_EnforceRequest(t *testing.T) {
	// Setup test environment
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)

	pdp := core.NewPolicyDecisionPoint(testStorage)
	auditLogger := NewNoOpAuditLogger()

	config := &PEPConfig{
		FailSafeMode:      true,
		StrictValidation:  true,
		AuditEnabled:      false, // Disable for testing
		EvaluationTimeout: time.Millisecond * 100,
	}

	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)

	tests := []struct {
		name           string
		request        *models.EvaluationRequest
		expectedResult string
		expectError    bool
	}{
		{
			name: "Valid permit request",
			request: &models.EvaluationRequest{
				RequestID:  "test-001",
				SubjectID:  "sub-001", // Engineering user
				ResourceID: "res-001", // Use existing resource ID from mock data
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			expectedResult: "permit",
			expectError:    false,
		},
		{
			name: "Valid deny request",
			request: &models.EvaluationRequest{
				RequestID:  "test-002",
				SubjectID:  "sub-004", // User on probation
				ResourceID: "res-001", // Use existing resource ID
				Action:     "write",
				Context: map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			expectedResult: "deny",
			expectError:    false,
		},
		{
			name: "Invalid request - missing subject",
			request: &models.EvaluationRequest{
				RequestID:  "test-003",
				SubjectID:  "",
				ResourceID: "/api/v1/users",
				Action:     "read",
			},
			expectedResult: "deny",
			expectError:    false,
		},
		{
			name: "Invalid request - missing resource",
			request: &models.EvaluationRequest{
				RequestID:  "test-004",
				SubjectID:  "sub-001",
				ResourceID: "",
				Action:     "read",
			},
			expectedResult: "deny",
			expectError:    false,
		},
		{
			name: "Invalid request - missing action",
			request: &models.EvaluationRequest{
				RequestID:  "test-005",
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "",
			},
			expectedResult: "deny",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := pep.EnforceRequest(ctx, tt.request)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatalf("Result should not be nil")
			}

			if result.Decision != tt.expectedResult {
				t.Errorf("Expected result %s, got %s", tt.expectedResult, result.Decision)
			}

			// Check that allowed matches the decision
			expectedAllowed := tt.expectedResult == "permit"
			if result.Allowed != expectedAllowed {
				t.Errorf("Expected allowed %v, got %v", expectedAllowed, result.Allowed)
			}
		})
	}
}

func TestSimplePolicyEnforcementPoint_Metrics(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)
	pdp := core.NewPolicyDecisionPoint(testStorage)
	auditLogger := NewNoOpAuditLogger()

	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

	// Initial metrics should be zero
	metrics := pep.GetMetrics()
	if metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", metrics.TotalRequests)
	}

	// Make a request
	ctx := context.Background()
	request := &models.EvaluationRequest{
		RequestID:  "test-metrics",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
	}

	_, err := pep.EnforceRequest(ctx, request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check metrics updated
	metrics = pep.GetMetrics()
	if metrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", metrics.TotalRequests)
	}

	if metrics.PermitDecisions != 1 {
		t.Errorf("Expected 1 permit decision, got %d", metrics.PermitDecisions)
	}
}

func TestSimplePolicyEnforcementPoint_Validation(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)
	pdp := core.NewPolicyDecisionPoint(testStorage)
	auditLogger := NewNoOpAuditLogger()

	config := &PEPConfig{
		StrictValidation: true,
		FailSafeMode:     true,
	}

	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)

	tests := []struct {
		name    string
		request *models.EvaluationRequest
	}{
		{
			name:    "Nil request",
			request: nil,
		},
		{
			name: "Long subject ID",
			request: &models.EvaluationRequest{
				SubjectID:  string(make([]byte, 300)), // Too long
				ResourceID: "test",
				Action:     "read",
			},
		},
		{
			name: "Long resource ID",
			request: &models.EvaluationRequest{
				SubjectID:  "test",
				ResourceID: string(make([]byte, 300)), // Too long
				Action:     "read",
			},
		},
		{
			name: "Long action",
			request: &models.EvaluationRequest{
				SubjectID:  "test",
				ResourceID: "test",
				Action:     string(make([]byte, 150)), // Too long
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := pep.EnforceRequest(ctx, tt.request)

			// Should not return error (fail-safe mode)
			if err != nil {
				t.Errorf("Unexpected error in fail-safe mode: %v", err)
			}

			// Should deny invalid requests
			if result == nil || result.Decision != "deny" {
				t.Errorf("Expected deny result for invalid request")
			}
		})
	}
}

func TestSimplePolicyEnforcementPoint_Timeout(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)
	pdp := core.NewPolicyDecisionPoint(testStorage)
	auditLogger := NewNoOpAuditLogger()

	config := &PEPConfig{
		EvaluationTimeout: time.Nanosecond, // Very short timeout
		FailSafeMode:      true,
	}

	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)

	request := &models.EvaluationRequest{
		RequestID:  "test-timeout",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
	}

	ctx := context.Background()
	result, err := pep.EnforceRequest(ctx, request)

	// Should not return error (fail-safe mode)
	if err != nil {
		t.Errorf("Unexpected error in fail-safe mode: %v", err)
	}

	// Should deny on timeout
	if result == nil || result.Decision != "deny" {
		t.Errorf("Expected deny result for timeout")
	}
}

func BenchmarkSimplePolicyEnforcementPoint_EnforceRequest(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}
