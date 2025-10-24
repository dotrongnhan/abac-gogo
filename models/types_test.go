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
		Version:     "2024-10-21",
		Enabled:     true,
		Statement: []PolicyStatement{
			{
				Sid:    "TestStatement",
				Effect: "Allow",
				Action: JSONActionResource{
					Single:  "document-service:file:read",
					IsArray: false,
				},
				Resource: JSONActionResource{
					Single:  "api:documents:*",
					IsArray: false,
				},
				Condition: JSONMap{
					"StringEquals": map[string]interface{}{
						"user.department": "engineering",
					},
				},
			},
		},
	}

	// Test valid effects
	validEffects := []string{"Allow", "Deny"}
	for _, effect := range validEffects {
		policy.Statement[0].Effect = effect
		if policy.Statement[0].Effect != effect {
			t.Errorf("Expected effect %s, got %s", effect, policy.Statement[0].Effect)
		}
	}

	// Test statement validation
	if len(policy.Statement) == 0 {
		t.Error("Policy should have at least one statement")
	}

	stmt := policy.Statement[0]
	if stmt.Effect == "" {
		t.Error("Statement effect should not be empty")
	}

	if stmt.Action.Single == "" && len(stmt.Action.Multiple) == 0 {
		t.Error("Statement should have an action")
	}

	if stmt.Resource.Single == "" && len(stmt.Resource.Multiple) == 0 {
		t.Error("Statement should have a resource")
	}
}

func TestEvaluationRequest(t *testing.T) {
	now := time.Now()
	request := &EvaluationRequest{
		RequestID:  "test-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Timestamp:  &now, // Enhanced: explicit timestamp
		Environment: &EnvironmentInfo{ // Enhanced: environmental context
			ClientIP:  "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Country:   "Vietnam",
			Region:    "Ho Chi Minh City",
			TimeOfDay: now.Format("15:04"),
			DayOfWeek: now.Weekday().String(),
			Attributes: map[string]interface{}{
				"device_type": "desktop",
				"connection":  "wifi",
			},
		},
		Context: map[string]interface{}{
			"timestamp": now.Format(time.RFC3339),
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

// Test for enhanced EnvironmentInfo struct
func TestEnvironmentInfo(t *testing.T) {
	now := time.Now()
	envInfo := &EnvironmentInfo{
		ClientIP:  "192.168.1.100",
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
		Country:   "Vietnam",
		Region:    "Ho Chi Minh City",
		TimeOfDay: now.Format("15:04"),
		DayOfWeek: now.Weekday().String(),
		Attributes: map[string]interface{}{
			"device_type":   "mobile",
			"connection":    "4g",
			"screen_size":   "375x812",
			"vpn_connected": false,
			"app_version":   "2.1.0",
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(envInfo)
	if err != nil {
		t.Fatalf("Failed to marshal EnvironmentInfo: %v", err)
	}

	var unmarshaled EnvironmentInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal EnvironmentInfo: %v", err)
	}

	// Validate fields
	if unmarshaled.ClientIP != envInfo.ClientIP {
		t.Errorf("Expected ClientIP %s, got %s", envInfo.ClientIP, unmarshaled.ClientIP)
	}

	if unmarshaled.UserAgent != envInfo.UserAgent {
		t.Errorf("Expected UserAgent %s, got %s", envInfo.UserAgent, unmarshaled.UserAgent)
	}

	if unmarshaled.Country != envInfo.Country {
		t.Errorf("Expected Country %s, got %s", envInfo.Country, unmarshaled.Country)
	}

	if unmarshaled.TimeOfDay != envInfo.TimeOfDay {
		t.Errorf("Expected TimeOfDay %s, got %s", envInfo.TimeOfDay, unmarshaled.TimeOfDay)
	}

	if unmarshaled.DayOfWeek != envInfo.DayOfWeek {
		t.Errorf("Expected DayOfWeek %s, got %s", envInfo.DayOfWeek, unmarshaled.DayOfWeek)
	}

	// Test attributes
	if len(unmarshaled.Attributes) != len(envInfo.Attributes) {
		t.Errorf("Expected %d attributes, got %d", len(envInfo.Attributes), len(unmarshaled.Attributes))
	}

	for key, expectedValue := range envInfo.Attributes {
		if actualValue, exists := unmarshaled.Attributes[key]; !exists {
			t.Errorf("Missing attribute %s", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected attribute %s = %v, got %v", key, expectedValue, actualValue)
		}
	}
}

// Test enhanced EvaluationRequest with new fields
func TestEnhancedEvaluationRequest(t *testing.T) {
	now := time.Now()
	request := &EvaluationRequest{
		RequestID:  "enhanced-test-001",
		SubjectID:  "user-123",
		ResourceID: "/api/documents/confidential/project-alpha.pdf",
		Action:     "read",
		Timestamp:  &now,
		Environment: &EnvironmentInfo{
			ClientIP:  "10.0.1.50",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			Country:   "Vietnam",
			Region:    "Ho Chi Minh City",
			TimeOfDay: "14:30",
			DayOfWeek: "Wednesday",
			Attributes: map[string]interface{}{
				"device_type":    "desktop",
				"browser":        "chrome",
				"is_mobile":      false,
				"connection":     "ethernet",
				"security_level": "high",
			},
		},
		Context: map[string]interface{}{
			"department":   "Engineering",
			"clearance":    "confidential",
			"project":      "alpha",
			"mfa_verified": true,
			"session_id":   "sess_abc123",
		},
	}

	// Test all required fields
	if request.RequestID == "" {
		t.Error("RequestID should not be empty")
	}

	if request.SubjectID == "" {
		t.Error("SubjectID should not be empty")
	}

	if request.ResourceID == "" {
		t.Error("ResourceID should not be empty")
	}

	if request.Action == "" {
		t.Error("Action should not be empty")
	}

	// Test enhanced fields
	if request.Timestamp == nil {
		t.Error("Timestamp should not be nil")
	}

	if request.Environment == nil {
		t.Error("Environment should not be nil")
	}

	// Test environmental context
	if request.Environment.ClientIP == "" {
		t.Error("ClientIP should not be empty")
	}

	if request.Environment.TimeOfDay == "" {
		t.Error("TimeOfDay should not be empty")
	}

	if request.Environment.DayOfWeek == "" {
		t.Error("DayOfWeek should not be empty")
	}

	// Test JSON serialization with enhanced fields
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal enhanced EvaluationRequest: %v", err)
	}

	var unmarshaled EvaluationRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal enhanced EvaluationRequest: %v", err)
	}

	// Validate unmarshaled data
	if unmarshaled.RequestID != request.RequestID {
		t.Errorf("Expected RequestID %s, got %s", request.RequestID, unmarshaled.RequestID)
	}

	if unmarshaled.Environment == nil {
		t.Error("Unmarshaled Environment should not be nil")
	}

	if unmarshaled.Environment != nil && unmarshaled.Environment.ClientIP != request.Environment.ClientIP {
		t.Errorf("Expected ClientIP %s, got %s", request.Environment.ClientIP, unmarshaled.Environment.ClientIP)
	}
}
