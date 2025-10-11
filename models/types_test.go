package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSubjectSerialization(t *testing.T) {
	subject := &Subject{
		ID:          "sub-001",
		ExternalID:  "john.doe@company.com",
		SubjectType: "user",
		Metadata: map[string]interface{}{
			"full_name": "John Doe",
			"email":     "john.doe@company.com",
		},
		Attributes: map[string]interface{}{
			"department": "engineering",
			"role":       []string{"senior_developer", "code_reviewer"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(subject)
	if err != nil {
		t.Fatalf("Failed to marshal subject: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Subject
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal subject: %v", err)
	}

	if unmarshaled.ID != subject.ID {
		t.Errorf("Expected ID %s, got %s", subject.ID, unmarshaled.ID)
	}

	if unmarshaled.SubjectType != subject.SubjectType {
		t.Errorf("Expected SubjectType %s, got %s", subject.SubjectType, unmarshaled.SubjectType)
	}
}

func TestPolicyValidation(t *testing.T) {
	policy := &Policy{
		ID:          "pol-001",
		PolicyName:  "Test Policy",
		Description: "Test policy description",
		Effect:      "permit",
		Priority:    100,
		Enabled:     true,
		Version:     1,
		Rules: []PolicyRule{
			{
				TargetType:    "subject",
				AttributePath: "attributes.department",
				Operator:      "eq",
				ExpectedValue: "engineering",
			},
		},
		Actions:          []string{"read"},
		ResourcePatterns: []string{"/api/v1/*"},
	}

	// Test valid effects
	validEffects := []string{"permit", "deny"}
	for _, effect := range validEffects {
		policy.Effect = effect
		if policy.Effect != effect {
			t.Errorf("Expected effect %s, got %s", effect, policy.Effect)
		}
	}

	// Test rule validation
	if len(policy.Rules) == 0 {
		t.Error("Policy should have at least one rule")
	}

	rule := policy.Rules[0]
	if rule.TargetType == "" {
		t.Error("Rule target type should not be empty")
	}

	if rule.AttributePath == "" {
		t.Error("Rule attribute path should not be empty")
	}

	if rule.Operator == "" {
		t.Error("Rule operator should not be empty")
	}
}

func TestEvaluationRequest(t *testing.T) {
	request := &EvaluationRequest{
		RequestID:  "test-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"source_ip": "10.0.1.100",
		},
	}

	if request.RequestID == "" {
		t.Error("Request ID should not be empty")
	}

	if request.SubjectID == "" {
		t.Error("Subject ID should not be empty")
	}

	if request.ResourceID == "" {
		t.Error("Resource ID should not be empty")
	}

	if request.Action == "" {
		t.Error("Action should not be empty")
	}

	if request.Context == nil {
		t.Error("Context should not be nil")
	}
}

func TestDecisionResults(t *testing.T) {
	decision := &Decision{
		Result:           "permit",
		MatchedPolicies:  []string{"pol-001", "pol-002"},
		EvaluationTimeMs: 5,
		Reason:           "Access granted by matching permit policies",
	}

	validResults := []string{"permit", "deny", "not_applicable"}
	found := false
	for _, validResult := range validResults {
		if decision.Result == validResult {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Invalid decision result: %s", decision.Result)
	}

	if len(decision.MatchedPolicies) == 0 {
		t.Error("Decision should have matched policies for permit result")
	}

	if decision.EvaluationTimeMs < 0 {
		t.Error("Evaluation time should not be negative")
	}
}

func TestAuditLogStructure(t *testing.T) {
	auditLog := &AuditLog{
		ID:           1001,
		RequestID:    "req-001",
		SubjectID:    "sub-001",
		ResourceID:   "res-001",
		ActionID:     "read",
		Decision:     "permit",
		EvaluationMs: 5,
		CreatedAt:    time.Now(),
		Context: map[string]interface{}{
			"matched_policies": []string{"pol-001"},
			"source_ip":        "10.0.1.100",
		},
	}

	// Test required fields
	if auditLog.RequestID == "" {
		t.Error("Audit log should have request ID")
	}

	if auditLog.SubjectID == "" {
		t.Error("Audit log should have subject ID")
	}

	if auditLog.Decision == "" {
		t.Error("Audit log should have decision")
	}

	if auditLog.CreatedAt.IsZero() {
		t.Error("Audit log should have creation timestamp")
	}

	// Test JSON serialization
	data, err := json.Marshal(auditLog)
	if err != nil {
		t.Fatalf("Failed to marshal audit log: %v", err)
	}

	var unmarshaled AuditLog
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal audit log: %v", err)
	}

	if unmarshaled.RequestID != auditLog.RequestID {
		t.Errorf("Expected RequestID %s, got %s", auditLog.RequestID, unmarshaled.RequestID)
	}
}
