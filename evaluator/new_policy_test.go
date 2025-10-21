package evaluator

import (
	"encoding/json"
	"testing"

	"abac_go_example/models"
)

func TestActionMatcher(t *testing.T) {
	matcher := NewActionMatcher()

	tests := []struct {
		pattern  string
		action   string
		expected bool
	}{
		// Exact matches
		{"document-service:file:read", "document-service:file:read", true},
		{"document-service:file:read", "document-service:file:write", false},

		// Wildcard matches
		{"document-service:file:*", "document-service:file:read", true},
		{"document-service:file:*", "document-service:file:write", true},
		{"document-service:*:read", "document-service:file:read", true},
		{"document-service:*:read", "document-service:folder:read", true},
		{"*:*:read", "document-service:file:read", true},
		{"*:*:read", "payment-service:transaction:read", true},
		{"*", "any:action:here", true},

		// Non-matches
		{"document-service:file:*", "payment-service:file:read", false},
		{"document-service:*:read", "document-service:file:write", false},
		{"*:*:read", "document-service:file:write", false},
	}

	for _, test := range tests {
		result := matcher.Match(test.pattern, test.action)
		if result != test.expected {
			t.Errorf("ActionMatcher.Match(%q, %q) = %v, want %v",
				test.pattern, test.action, result, test.expected)
		}
	}
}

func TestResourceMatcher(t *testing.T) {
	matcher := NewResourceMatcher()

	context := map[string]interface{}{
		"request:UserId":  "user-123",
		"user:Department": "engineering",
	}

	tests := []struct {
		pattern  string
		resource string
		expected bool
	}{
		// Exact matches
		{"api:documents:doc-123", "api:documents:doc-123", true},
		{"api:documents:doc-123", "api:documents:doc-456", false},

		// Wildcard matches
		{"api:documents:*", "api:documents:doc-123", true},
		{"api:*:doc-123", "api:documents:doc-123", true},
		{"*:*:*", "api:documents:doc-123", true},
		{"*", "anything", true},

		// Variable substitution
		{"api:documents:owner:${request:UserId}/*", "api:documents:owner:user-123/file-1", true},
		{"api:documents:dept:${user:Department}/*", "api:documents:dept:engineering/file-1", true},

		// Prefix wildcards
		{"api:documents:admin-*", "api:documents:admin-123", true},
		{"api:documents:admin-*", "api:documents:user-123", false},
	}

	for _, test := range tests {
		result := matcher.Match(test.pattern, test.resource, context)
		if result != test.expected {
			t.Errorf("ResourceMatcher.Match(%q, %q) = %v, want %v",
				test.pattern, test.resource, result, test.expected)
		}
	}
}

func TestConditionEvaluator(t *testing.T) {
	evaluator := NewConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"department": "engineering",
			"level":      5,
			"mfa":        true,
		},
		"resource": map[string]interface{}{
			"sensitivity": "confidential",
			"owner":       "user-123",
		},
		"request": map[string]interface{}{
			"sourceIp": "10.0.1.50",
			"time":     "2024-10-21T10:00:00Z",
		},
		"transaction": map[string]interface{}{
			"amount": 500000,
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "StringEquals - match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.department": "engineering",
				},
			},
			expected: true,
		},
		{
			name: "StringEquals - no match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.department": "finance",
				},
			},
			expected: false,
		},
		{
			name: "NumericLessThan - match",
			conditions: map[string]interface{}{
				"NumericLessThan": map[string]interface{}{
					"transaction.amount": 1000000,
				},
			},
			expected: true,
		},
		{
			name: "NumericLessThan - no match",
			conditions: map[string]interface{}{
				"NumericLessThan": map[string]interface{}{
					"transaction.amount": 100000,
				},
			},
			expected: false,
		},
		{
			name: "Bool - match",
			conditions: map[string]interface{}{
				"Bool": map[string]interface{}{
					"user.mfa": true,
				},
			},
			expected: true,
		},
		{
			name: "Multiple conditions (AND) - all match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.department": "engineering",
				},
				"Bool": map[string]interface{}{
					"user.mfa": true,
				},
			},
			expected: true,
		},
		{
			name: "Multiple conditions (AND) - one fails",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.department": "engineering",
				},
				"Bool": map[string]interface{}{
					"user.mfa": false,
				},
			},
			expected: false,
		},
		{
			name: "IpAddress - match",
			conditions: map[string]interface{}{
				"IpAddress": map[string]interface{}{
					"request.sourceIp": []interface{}{"10.0.0.0/8", "192.168.1.0/24"},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.Evaluate(test.conditions, context)
			if result != test.expected {
				t.Errorf("ConditionEvaluator.Evaluate() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestJSONActionResource(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected models.JSONActionResource
	}{
		{
			name:  "Single string",
			input: `"document-service:file:read"`,
			expected: models.JSONActionResource{
				Single:  "document-service:file:read",
				IsArray: false,
			},
		},
		{
			name:  "Array of strings",
			input: `["document-service:file:read", "document-service:file:write"]`,
			expected: models.JSONActionResource{
				Multiple: []string{"document-service:file:read", "document-service:file:write"},
				IsArray:  true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result models.JSONActionResource
			err := json.Unmarshal([]byte(test.input), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.IsArray != test.expected.IsArray {
				t.Errorf("IsArray = %v, want %v", result.IsArray, test.expected.IsArray)
			}

			if !result.IsArray && result.Single != test.expected.Single {
				t.Errorf("Single = %v, want %v", result.Single, test.expected.Single)
			}

			if result.IsArray {
				if len(result.Multiple) != len(test.expected.Multiple) {
					t.Errorf("Multiple length = %v, want %v", len(result.Multiple), len(test.expected.Multiple))
				}
				for i, v := range result.Multiple {
					if v != test.expected.Multiple[i] {
						t.Errorf("Multiple[%d] = %v, want %v", i, v, test.expected.Multiple[i])
					}
				}
			}
		})
	}
}

func TestPolicyStatement(t *testing.T) {
	// Test parsing a complete policy statement
	statementJSON := `{
		"Sid": "OwnDocumentsFullAccess",
		"Effect": "Allow",
		"Action": "document-service:file:*",
		"Resource": "api:documents:owner:${request:UserId}/*",
		"Condition": {
			"StringEquals": {
				"user.department": "engineering"
			}
		}
	}`

	var statement models.PolicyStatement
	err := json.Unmarshal([]byte(statementJSON), &statement)
	if err != nil {
		t.Fatalf("Failed to unmarshal statement: %v", err)
	}

	if statement.Sid != "OwnDocumentsFullAccess" {
		t.Errorf("Sid = %v, want %v", statement.Sid, "OwnDocumentsFullAccess")
	}

	if statement.Effect != "Allow" {
		t.Errorf("Effect = %v, want %v", statement.Effect, "Allow")
	}

	if statement.Action.Single != "document-service:file:*" {
		t.Errorf("Action.Single = %v, want %v", statement.Action.Single, "document-service:file:*")
	}

	if statement.Resource.Single != "api:documents:owner:${request:UserId}/*" {
		t.Errorf("Resource.Single = %v, want %v", statement.Resource.Single, "api:documents:owner:${request:UserId}/*")
	}

	if len(statement.Condition) == 0 {
		t.Error("Condition should not be empty")
	}
}

func TestDenyOverrideAlgorithm(t *testing.T) {
	// Create mock policies for testing
	policies := []*models.Policy{
		{
			ID:      "pol-allow",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "AllowRead",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "document-service:file:read", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:documents:*", IsArray: false},
				},
			},
		},
		{
			ID:      "pol-deny",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "DenyConfidential",
					Effect:   "Deny",
					Action:   models.JSONActionResource{Single: "document-service:file:*", IsArray: false},
					Resource: models.JSONActionResource{Single: "*", IsArray: false},
					Condition: map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"resource.sensitivity": "confidential",
						},
					},
				},
			},
		},
	}

	pdp := &PolicyDecisionPoint{
		actionMatcher:      NewActionMatcher(),
		resourceMatcher:    NewResourceMatcher(),
		conditionEvaluator: NewConditionEvaluator(),
	}

	tests := []struct {
		name     string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "Allow case - no deny conditions",
			context: map[string]interface{}{
				"request:Action":     "document-service:file:read",
				"request:ResourceId": "api:documents:doc-123",
				"resource": map[string]interface{}{
					"sensitivity": "public",
				},
			},
			expected: "permit",
		},
		{
			name: "Deny case - deny condition matches",
			context: map[string]interface{}{
				"request:Action":     "document-service:file:read",
				"request:ResourceId": "api:documents:doc-123",
				"resource": map[string]interface{}{
					"sensitivity": "confidential",
				},
			},
			expected: "deny",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision := pdp.evaluateNewPolicies(policies, test.context)
			if decision.Result != test.expected {
				t.Errorf("Decision.Result = %v, want %v", decision.Result, test.expected)
			}
		})
	}
}
