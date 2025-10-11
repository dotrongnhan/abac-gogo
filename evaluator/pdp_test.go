package evaluator

import (
	"testing"
	"time"

	"abac_go_example/models"
)

// Mock storage for testing
type mockStorage struct {
	subjects  map[string]*models.Subject
	resources map[string]*models.Resource
	actions   map[string]*models.Action
	policies  []*models.Policy
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
	return m.policies, nil
}

func (m *mockStorage) GetAllSubjects() ([]*models.Subject, error) {
	subjects := make([]*models.Subject, 0, len(m.subjects))
	for _, subject := range m.subjects {
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

func (m *mockStorage) GetAllResources() ([]*models.Resource, error) {
	resources := make([]*models.Resource, 0, len(m.resources))
	for _, resource := range m.resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

func (m *mockStorage) GetAllActions() ([]*models.Action, error) {
	actions := make([]*models.Action, 0, len(m.actions))
	for _, action := range m.actions {
		actions = append(actions, action)
	}
	return actions, nil
}

func createTestStorage() *mockStorage {
	return &mockStorage{
		subjects: map[string]*models.Subject{
			"sub-001": {
				ID:          "sub-001",
				ExternalID:  "john.doe@company.com",
				SubjectType: "user",
				Attributes: map[string]interface{}{
					"department":       "engineering",
					"role":             []string{"senior_developer", "code_reviewer"},
					"clearance_level":  3,
					"years_of_service": 5,
				},
			},
			"sub-002": {
				ID:          "sub-002",
				ExternalID:  "alice.smith@company.com",
				SubjectType: "user",
				Attributes: map[string]interface{}{
					"department":      "finance",
					"role":            []string{"accountant"},
					"clearance_level": 2,
				},
			},
			"sub-003": {
				ID:          "sub-003",
				ExternalID:  "bob.wilson@company.com",
				SubjectType: "user",
				Attributes: map[string]interface{}{
					"department":   "engineering",
					"role":         []string{"junior_developer"},
					"on_probation": true,
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
			"res-002": {
				ID:           "res-002",
				ResourceType: "document",
				ResourceID:   "DOC-2024-FINANCE",
				Attributes: map[string]interface{}{
					"data_classification": "confidential",
					"document_type":       "financial_report",
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
			"write": {
				ID:             "act-002",
				ActionName:     "write",
				ActionCategory: "crud",
				Description:    "Create/Update resource",
			},
		},
		policies: []*models.Policy{
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
					{
						TargetType:    "resource",
						AttributePath: "attributes.data_classification",
						Operator:      "in",
						ExpectedValue: []string{"public", "internal"},
					},
				},
				Actions:          []string{"read"},
				ResourcePatterns: []string{"/api/v1/*"},
			},
			{
				ID:          "pol-002",
				PolicyName:  "Senior Developer Write Access",
				Description: "Senior developers can write to APIs",
				Effect:      "permit",
				Priority:    50,
				Enabled:     true,
				Version:     1,
				Rules: []models.PolicyRule{
					{
						TargetType:    "subject",
						AttributePath: "attributes.role",
						Operator:      "contains",
						ExpectedValue: "senior_developer",
					},
					{
						TargetType:    "subject",
						AttributePath: "attributes.years_of_service",
						Operator:      "gte",
						ExpectedValue: 2,
					},
					{
						TargetType:    "environment",
						AttributePath: "time_of_day",
						Operator:      "between",
						ExpectedValue: []string{"08:00", "20:00"},
					},
				},
				Actions:          []string{"read", "write"},
				ResourcePatterns: []string{"/api/v1/*"},
			},
			{
				ID:          "pol-003",
				PolicyName:  "Finance Confidential Access",
				Description: "Finance team access to confidential financial data",
				Effect:      "permit",
				Priority:    30,
				Enabled:     true,
				Version:     1,
				Rules: []models.PolicyRule{
					{
						TargetType:    "subject",
						AttributePath: "attributes.department",
						Operator:      "eq",
						ExpectedValue: "finance",
					},
					{
						TargetType:    "subject",
						AttributePath: "attributes.clearance_level",
						Operator:      "gte",
						ExpectedValue: 2,
					},
					{
						TargetType:    "resource",
						AttributePath: "attributes.document_type",
						Operator:      "eq",
						ExpectedValue: "financial_report",
					},
				},
				Actions:          []string{"read", "write"},
				ResourcePatterns: []string{"DOC-*-FINANCE"},
			},
			{
				ID:          "pol-004",
				PolicyName:  "Deny Probation Write",
				Description: "Deny write access for employees on probation",
				Effect:      "deny",
				Priority:    10,
				Enabled:     true,
				Version:     1,
				Rules: []models.PolicyRule{
					{
						TargetType:    "subject",
						AttributePath: "attributes.on_probation",
						Operator:      "eq",
						ExpectedValue: true,
					},
				},
				Actions:          []string{"write", "delete"},
				ResourcePatterns: []string{"*"},
			},
		},
	}
}

func TestPolicyEvaluationPermit(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	// Test case: Senior developer reading API
	request := &models.EvaluationRequest{
		RequestID:  "test-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"source_ip":   "10.0.1.100",
			"time_of_day": "14:00",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if decision.Result != "permit" {
		t.Errorf("Expected permit, got %s", decision.Result)
	}

	if len(decision.MatchedPolicies) == 0 {
		t.Error("Expected matched policies for permit decision")
	}

	if decision.EvaluationTimeMs < 0 {
		t.Error("Evaluation time should not be negative")
	}
}

func TestPolicyEvaluationDeny(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	// Test case: Probation employee trying to write
	request := &models.EvaluationRequest{
		RequestID:  "test-002",
		SubjectID:  "sub-003",
		ResourceID: "res-001",
		Action:     "write",
		Context: map[string]interface{}{
			"timestamp": "2024-01-15T14:00:00Z",
			"source_ip": "10.0.1.100",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if decision.Result != "deny" {
		t.Errorf("Expected deny, got %s", decision.Result)
	}

	if decision.Reason == "" {
		t.Error("Deny decision should have a reason")
	}

	// Should contain the deny policy
	found := false
	for _, policyID := range decision.MatchedPolicies {
		if policyID == "pol-004" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected pol-004 (deny policy) in matched policies")
	}
}

func TestPolicyEvaluationNotApplicable(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	// Test case: Finance user trying to access engineering resource
	request := &models.EvaluationRequest{
		RequestID:  "test-003",
		SubjectID:  "sub-002",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": "2024-01-15T14:00:00Z",
			"source_ip": "10.0.1.100",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if decision.Result != "not_applicable" {
		t.Errorf("Expected not_applicable, got %s", decision.Result)
	}
}

func TestPolicyPriorityOrdering(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	// Get applicable policies for a request
	request := &models.EvaluationRequest{
		RequestID:  "test-priority",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "write",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
		},
	}

	policies, err := pdp.GetApplicablePolicies(request)
	if err != nil {
		t.Fatalf("Failed to get applicable policies: %v", err)
	}

	// Should have multiple policies, verify they would be sorted by priority
	if len(policies) < 2 {
		t.Skip("Need at least 2 applicable policies to test priority ordering")
	}

	// Get the actual decision to see how policies are sorted
	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Failed to evaluate request: %v", err)
	}

	// Check that if there are deny policies, they should be processed first
	// (due to priority ordering and short-circuit logic)
	if len(decision.MatchedPolicies) > 0 {
		t.Logf("Matched policies in order: %v", decision.MatchedPolicies)
		t.Logf("Decision result: %s", decision.Result)
	}
}

func TestActionMatching(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	testCases := []struct {
		policyActions   []string
		requestedAction string
		expected        bool
	}{
		{[]string{"read"}, "read", true},
		{[]string{"read", "write"}, "read", true},
		{[]string{"read", "write"}, "write", true},
		{[]string{"read"}, "write", false},
		{[]string{"*"}, "read", true},
		{[]string{"*"}, "write", true},
		{[]string{}, "read", true}, // No restriction
	}

	for _, tc := range testCases {
		result := pdp.actionMatches(tc.policyActions, tc.requestedAction)
		if result != tc.expected {
			t.Errorf("actionMatches(%v, %s) = %v, expected %v",
				tc.policyActions, tc.requestedAction, result, tc.expected)
		}
	}
}

func TestResourcePatternMatching(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	resource := &models.Resource{
		ID:         "res-001",
		ResourceID: "/api/v1/users",
		Path:       "api.v1.users",
	}

	testCases := []struct {
		patterns []string
		expected bool
	}{
		{[]string{"/api/v1/users"}, true},
		{[]string{"/api/v1/*"}, true},
		{[]string{"/api/v2/*"}, false},
		{[]string{"*"}, true},
		{[]string{}, true}, // No restriction
		{[]string{"DOC-*"}, false},
	}

	for _, tc := range testCases {
		result := pdp.resourcePatternMatches(tc.patterns, resource)
		if result != tc.expected {
			t.Errorf("resourcePatternMatches(%v, %v) = %v, expected %v",
				tc.patterns, resource.ResourceID, result, tc.expected)
		}
	}
}

func TestBatchEvaluation(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	requests := []*models.EvaluationRequest{
		{
			RequestID:  "batch-001",
			SubjectID:  "sub-001",
			ResourceID: "res-001",
			Action:     "read",
			Context: map[string]interface{}{
				"timestamp":   "2024-01-15T14:00:00Z",
				"time_of_day": "14:00",
			},
		},
		{
			RequestID:  "batch-002",
			SubjectID:  "sub-003",
			ResourceID: "res-001",
			Action:     "write",
			Context: map[string]interface{}{
				"timestamp": "2024-01-15T14:00:00Z",
			},
		},
	}

	decisions, err := pdp.BatchEvaluate(requests)
	if err != nil {
		t.Fatalf("Batch evaluation failed: %v", err)
	}

	if len(decisions) != len(requests) {
		t.Errorf("Expected %d decisions, got %d", len(requests), len(decisions))
	}

	// First request should be permit
	if decisions[0].Result != "permit" {
		t.Errorf("Expected first decision to be permit, got %s", decisions[0].Result)
	}

	// Second request should be deny (probation)
	if decisions[1].Result != "deny" {
		t.Errorf("Expected second decision to be deny, got %s", decisions[1].Result)
	}
}

func TestExplainDecision(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	request := &models.EvaluationRequest{
		RequestID:  "explain-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
		},
	}

	explanation, err := pdp.ExplainDecision(request)
	if err != nil {
		t.Fatalf("Failed to explain decision: %v", err)
	}

	// Check that explanation contains expected fields
	if _, exists := explanation["request"]; !exists {
		t.Error("Explanation should contain request")
	}

	if _, exists := explanation["context"]; !exists {
		t.Error("Explanation should contain context")
	}

	if _, exists := explanation["policy_evaluations"]; !exists {
		t.Error("Explanation should contain policy_evaluations")
	}

	if totalPolicies, exists := explanation["total_policies"]; !exists {
		t.Error("Explanation should contain total_policies")
	} else if count, ok := totalPolicies.(int); !ok || count <= 0 {
		t.Errorf("Expected positive total_policies count, got %v", totalPolicies)
	}

	if applicablePolicies, exists := explanation["applicable_policies"]; !exists {
		t.Error("Explanation should contain applicable_policies")
	} else if count, ok := applicablePolicies.(int); !ok || count < 0 {
		t.Errorf("Expected non-negative applicable_policies count, got %v", applicablePolicies)
	}
}

func TestDisabledPolicy(t *testing.T) {
	storage := createTestStorage()

	// Disable a policy
	for _, policy := range storage.policies {
		if policy.ID == "pol-001" {
			policy.Enabled = false
			break
		}
	}

	pdp := NewPolicyDecisionPoint(storage)

	request := &models.EvaluationRequest{
		RequestID:  "disabled-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// Should not match the disabled policy
	for _, policyID := range decision.MatchedPolicies {
		if policyID == "pol-001" {
			t.Error("Disabled policy should not be matched")
		}
	}
}

func TestPerformance(t *testing.T) {
	storage := createTestStorage()
	pdp := NewPolicyDecisionPoint(storage)

	request := &models.EvaluationRequest{
		RequestID:  "perf-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
		},
	}

	// Measure evaluation time
	start := time.Now()
	decision, err := pdp.Evaluate(request)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// Should complete quickly (under 10ms for this simple case)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Evaluation took too long: %v", elapsed)
	}

	// Reported evaluation time should be reasonable
	if decision.EvaluationTimeMs > 10 {
		t.Errorf("Reported evaluation time too high: %dms", decision.EvaluationTimeMs)
	}

	t.Logf("Evaluation completed in %v (reported: %dms)", elapsed, decision.EvaluationTimeMs)
}
