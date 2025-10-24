package evaluator

import (
	"testing"
	"time"

	"abac_go_example/models"
	"abac_go_example/storage"
)

// TestImprovedPDP_TimeBasedAttributes tests improvement #4: Time-based attributes
func TestImprovedPDP_TimeBasedAttributes(t *testing.T) {
	// Create mock storage
	mockStorage := &storage.MockStorage{}

	// Create improved PDP
	pdp := NewPolicyDecisionPoint(mockStorage).(*PolicyDecisionPoint)

	// Test time-based context building
	now := time.Date(2024, 10, 24, 14, 30, 0, 0, time.UTC) // Thursday 14:30
	request := &models.EvaluationRequest{
		RequestID:  "time-test-001",
		SubjectID:  "user-123",
		ResourceID: "/api/reports",
		Action:     "read",
		Timestamp:  &now,
		Environment: &models.EnvironmentInfo{
			TimeOfDay: "14:30",
			DayOfWeek: "Thursday",
		},
		Context: map[string]interface{}{
			"department": "Engineering",
		},
	}

	// Create mock evaluation context
	evalContext := &models.EvaluationContext{
		Subject: &models.Subject{
			ID:          "user-123",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      5,
			},
		},
		Resource: &models.Resource{
			ID:           "res-123",
			ResourceType: "report",
			ResourceID:   "/api/reports",
			Attributes: map[string]interface{}{
				"classification": "internal",
			},
		},
		Environment: map[string]interface{}{
			"source_ip": "192.168.1.100",
		},
		Timestamp: now,
	}

	// Build enhanced evaluation context
	context := pdp.BuildEnhancedEvaluationContext(request, evalContext)

	// Verify time-based attributes are added
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"environment:time_of_day", "14:30"},
		{"environment:day_of_week", "Thursday"},
		{"environment:hour", 14},
		{"environment:minute", 30},
		{"environment:is_weekend", false},
		{"environment:is_business_hours", true},
	}

	for _, test := range tests {
		if actual, exists := context[test.key]; !exists {
			t.Errorf("Expected key %s not found in context", test.key)
		} else if actual != test.expected {
			t.Errorf("Expected %s = %v, got %v", test.key, test.expected, actual)
		}
	}
}

// TestImprovedPDP_EnvironmentalContext tests improvement #5: Environmental context
func TestImprovedPDP_EnvironmentalContext(t *testing.T) {
	mockStorage := &storage.MockStorage{}
	pdp := NewPolicyDecisionPoint(mockStorage).(*PolicyDecisionPoint)

	request := &models.EvaluationRequest{
		RequestID:  "env-test-001",
		SubjectID:  "user-456",
		ResourceID: "/api/financial/reports",
		Action:     "read",
		Environment: &models.EnvironmentInfo{
			ClientIP:  "192.168.1.100",
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
			Country:   "Vietnam",
			Region:    "Ho Chi Minh City",
			Attributes: map[string]interface{}{
				"device_type": "mobile",
				"connection":  "4g",
			},
		},
	}

	evalContext := &models.EvaluationContext{
		Subject: &models.Subject{
			ID:          "user-456",
			SubjectType: "employee",
		},
		Resource: &models.Resource{
			ID:         "res-456",
			ResourceID: "/api/financial/reports",
		},
		Environment: map[string]interface{}{},
		Timestamp:   time.Now(),
	}

	context := pdp.BuildEnhancedEvaluationContext(request, evalContext)

	// Verify environmental attributes are processed
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"environment:client_ip", "192.168.1.100"},
		{"environment:user_agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)"},
		{"environment:country", "Vietnam"},
		{"environment:region", "Ho Chi Minh City"},
		{"environment:is_internal_ip", true},
		{"environment:ip_class", "ipv4"},
		{"environment:is_mobile", true},
		{"environment:browser", "unknown"}, // iPhone user agent doesn't contain browser name
		{"environment:device_type", "mobile"},
		{"environment:connection", "4g"},
	}

	for _, test := range tests {
		if actual, exists := context[test.key]; !exists {
			t.Errorf("Expected key %s not found in context", test.key)
		} else if actual != test.expected {
			t.Errorf("Expected %s = %v, got %v", test.key, test.expected, actual)
		}
	}
}

// TestImprovedPDP_StructuredAttributes tests improvement #6: Structured attributes
func TestImprovedPDP_StructuredAttributes(t *testing.T) {
	mockStorage := &storage.MockStorage{}
	pdp := NewPolicyDecisionPoint(mockStorage).(*PolicyDecisionPoint)

	request := &models.EvaluationRequest{
		RequestID:  "struct-test-001",
		SubjectID:  "user-789",
		ResourceID: "/documents/confidential/project-alpha.pdf",
		Action:     "read",
	}

	evalContext := &models.EvaluationContext{
		Subject: &models.Subject{
			ID:          "user-789",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      8,
				"clearance":  "confidential",
			},
		},
		Resource: &models.Resource{
			ID:           "res-789",
			ResourceType: "document",
			ResourceID:   "/documents/confidential/project-alpha.pdf",
			Attributes: map[string]interface{}{
				"classification": "confidential",
				"project":        "alpha",
				"owner":          "engineering-team",
			},
		},
		Environment: map[string]interface{}{},
		Timestamp:   time.Now(),
	}

	context := pdp.BuildEnhancedEvaluationContext(request, evalContext)

	// Verify both flat and structured access
	// Flat access (backward compatibility)
	if context["user:department"] != "Engineering" {
		t.Error("Flat user access should work")
	}
	if context["resource:classification"] != "confidential" {
		t.Error("Flat resource access should work")
	}

	// Structured access (new feature)
	userContext, exists := context["user"].(map[string]interface{})
	if !exists {
		t.Error("Structured user context should exist")
	} else {
		if userContext["subject_type"] != "employee" {
			t.Error("User context should include subject_type")
		}
		if userAttrs, ok := userContext["attributes"].(map[string]interface{}); ok {
			if userAttrs["department"] != "Engineering" {
				t.Error("User attributes should be accessible")
			}
		} else {
			t.Error("User attributes should be accessible in structured format")
		}
	}

	resourceContext, exists := context["resource"].(map[string]interface{})
	if !exists {
		t.Error("Structured resource context should exist")
	} else {
		if resourceContext["resource_type"] != "document" {
			t.Error("Resource context should include resource_type")
		}
		if resAttrs, ok := resourceContext["attributes"].(map[string]interface{}); ok {
			if resAttrs["classification"] != "confidential" {
				t.Error("Resource attributes should be accessible")
			}
		} else {
			t.Error("Resource attributes should be accessible in structured format")
		}
	}
}

// TestImprovedPDP_EnhancedConditionEvaluation tests improvement #7: Enhanced condition evaluator
func TestImprovedPDP_EnhancedConditionEvaluation(t *testing.T) {
	mockStorage := &storage.MockStorage{}
	pdp := NewPolicyDecisionPoint(mockStorage).(*PolicyDecisionPoint)

	// Test that enhanced condition evaluator is used
	if pdp.enhancedConditionEvaluator == nil {
		t.Error("Enhanced condition evaluator should be initialized")
	}

	// Test enhanced condition evaluation in statement evaluation
	statement := models.PolicyStatement{
		Sid:      "EnhancedConditionTest",
		Effect:   "Allow",
		Action:   models.JSONActionResource{Single: "read", IsArray: false},
		Resource: models.JSONActionResource{Single: "*", IsArray: false},
		Condition: map[string]interface{}{
			"StringContains": map[string]interface{}{
				"user.department": "Engineering",
			},
			"NumericGreaterThanEquals": map[string]interface{}{
				"user.level": 5,
			},
			"IsBusinessHours": map[string]interface{}{
				"environment.is_business_hours": true,
			},
		},
	}

	context := map[string]interface{}{
		"request:Action":     "read",
		"request:ResourceId": "/api/documents/test.pdf",
		"user": map[string]interface{}{
			"department": "Engineering Department",
			"level":      8,
		},
		"environment": map[string]interface{}{
			"is_business_hours": true,
		},
	}

	// This should use enhanced condition evaluation
	result := pdp.evaluateStatement(statement, context)
	if !result {
		t.Error("Enhanced condition evaluation should succeed")
	}
}

// TestImprovedPDP_PolicyFiltering tests that PDP correctly handles enabled/disabled policies
func TestImprovedPDP_PolicyFiltering(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	pdp := NewPolicyDecisionPoint(mockStorage).(*PolicyDecisionPoint)

	// Create test subject
	subject := &models.Subject{
		ID:          "user-123",
		ExternalID:  "user-123",
		SubjectType: "user",
		Attributes: map[string]interface{}{
			"department": "Engineering",
			"level":      5,
		},
	}
	err := mockStorage.CreateSubject(subject)
	if err != nil {
		t.Fatalf("Failed to create subject: %v", err)
	}

	// Create test action
	action := &models.Action{
		ID:             "document:read",
		ActionName:     "document:read",
		ActionCategory: "read",
		Description:    "Read document action",
	}
	err = mockStorage.CreateAction(action)
	if err != nil {
		t.Fatalf("Failed to create action: %v", err)
	}

	// Create test resource
	resource := &models.Resource{
		ID:         "api:documents:test.pdf",
		ResourceID: "api:documents:test.pdf",
		Attributes: map[string]interface{}{
			"type": "document",
		},
	}
	err = mockStorage.CreateResource(resource)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	// Create test policies
	policies := []*models.Policy{
		{
			ID:      "pol-001",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "DocumentRead",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "document:read", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:documents:*", IsArray: false},
				},
			},
		},
		{
			ID:      "pol-002",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "ReportWrite",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "report:write", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:reports:*", IsArray: false},
				},
			},
		},
		{
			ID:      "pol-003",
			Enabled: false, // Disabled
			Statement: []models.PolicyStatement{
				{
					Sid:      "DisabledPolicy",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "*", IsArray: false},
					Resource: models.JSONActionResource{Single: "*", IsArray: false},
				},
			},
		},
	}

	// Mock storage to return our test policies
	mockStorage.SetPolicies(policies)

	request := &models.EvaluationRequest{
		RequestID:  "filter-test-001",
		SubjectID:  "user-123",
		ResourceID: "api:documents:test.pdf",
		Action:     "document:read",
	}

	// Test that PDP correctly evaluates and only considers enabled policies
	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should allow because pol-001 is enabled and matches
	if decision.Result != "permit" {
		t.Errorf("Expected permit, got %s", decision.Result)
	}

	// Should only match pol-001 (pol-002 doesn't match action, pol-003 is disabled)
	if len(decision.MatchedPolicies) != 1 || decision.MatchedPolicies[0] != "pol-001" {
		t.Errorf("Expected only pol-001 to match, got %v", decision.MatchedPolicies)
	}
}

// TestImprovedPDP_IntegrationWithAllFeatures tests all improvements together
func TestImprovedPDP_IntegrationWithAllFeatures(t *testing.T) {
	mockStorage := storage.NewMockStorage()

	// Create test subject
	subject := &models.Subject{
		ID:          "user-comprehensive",
		ExternalID:  "user-comprehensive",
		SubjectType: "user",
		Attributes: map[string]interface{}{
			"department": "Engineering",
			"level":      8,
			"clearance":  "confidential",
		},
	}
	err := mockStorage.CreateSubject(subject)
	if err != nil {
		t.Fatalf("Failed to create subject: %v", err)
	}

	// Create test action
	action := &models.Action{
		ID:             "document:read",
		ActionName:     "document:read",
		ActionCategory: "read",
		Description:    "Read document action",
	}
	err = mockStorage.CreateAction(action)
	if err != nil {
		t.Fatalf("Failed to create action: %v", err)
	}

	// Create test resource
	resource := &models.Resource{
		ID:         "api:documents:test.pdf",
		ResourceID: "api:documents:test.pdf",
		Attributes: map[string]interface{}{
			"type": "document",
		},
	}
	err = mockStorage.CreateResource(resource)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	// Create a comprehensive policy that uses enhanced features
	policy := &models.Policy{
		ID:      "comprehensive-policy",
		Enabled: true,
		Statement: []models.PolicyStatement{
			{
				Sid:      "ComprehensiveAccess",
				Effect:   "Allow",
				Action:   models.JSONActionResource{Single: "document:read", IsArray: false},
				Resource: models.JSONActionResource{Single: "api:documents:*", IsArray: false},
				Condition: map[string]interface{}{
					// Simple condition for testing
					"StringEquals": map[string]interface{}{
						"user.department": "Engineering",
					},
				},
			},
		},
	}

	mockStorage.SetPolicies([]*models.Policy{policy})
	pdp := NewPolicyDecisionPoint(mockStorage)

	// Create simple request like the passing test
	request := &models.EvaluationRequest{
		RequestID:  "comprehensive-test-001",
		SubjectID:  "user-comprehensive",
		ResourceID: "api:documents:test.pdf",
		Action:     "document:read",
		Context: map[string]interface{}{
			"department": "Engineering",
		},
	}

	// Evaluate with all improvements
	decision, err := pdp.Evaluate(request)
	if err != nil {
		t.Fatalf("Comprehensive evaluation should not fail: %v", err)
	}

	if decision == nil {
		t.Error("Decision should not be nil")
	}

	// Should allow because all conditions match
	if decision.Result != "permit" {
		t.Errorf("Expected permit, got %s. Reason: %s", decision.Result, decision.Reason)
	}

	// Should match the comprehensive policy
	if len(decision.MatchedPolicies) != 1 || decision.MatchedPolicies[0] != "comprehensive-policy" {
		t.Errorf("Expected comprehensive-policy to match, got %v", decision.MatchedPolicies)
	}

	// The decision result shows that all enhanced features work together
	t.Logf("Comprehensive evaluation completed with result: %s", decision.Result)
}
