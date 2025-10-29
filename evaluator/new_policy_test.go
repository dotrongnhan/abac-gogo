package evaluator

import (
	"encoding/json"
	"testing"

	"abac_go_example/evaluator/matchers"
	"abac_go_example/models"
)

func TestActionMatcher(t *testing.T) {
	matcher := matchers.NewActionMatcher()

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
	matcher := matchers.NewResourceMatcher()

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
		{"api:documents:owner-${request:UserId}", "api:documents:owner-user-123", true},
		{"api:departments:${user:Department}/api:documents:*", "api:departments:engineering/api:documents:file-1", true},

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

// Legacy ConditionEvaluator tests removed - now using EnhancedConditionEvaluator exclusively

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
				Single: "document-service:file:read",
			},
		},
		{
			name:  "Array of strings",
			input: `["document-service:file:read", "document-service:file:write"]`,
			expected: models.JSONActionResource{
				Multiple: []string{"document-service:file:read", "document-service:file:write"},
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

			if result.Single != test.expected.Single {
				t.Errorf("Single = %v, want %v", result.Single, test.expected.Single)
			}

			if len(result.Multiple) != len(test.expected.Multiple) {
				t.Errorf("Multiple length = %v, want %v", len(result.Multiple), len(test.expected.Multiple))
			}
			for i, v := range result.Multiple {
				if i < len(test.expected.Multiple) && v != test.expected.Multiple[i] {
					t.Errorf("Multiple[%d] = %v, want %v", i, v, test.expected.Multiple[i])
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

// TestDenyOverrideAlgorithm removed - testing internal implementation details
// Integration tests should be in the core package instead

func TestNotResourceExclusion(t *testing.T) {
	matcher := matchers.NewResourceMatcher()

	context := map[string]interface{}{
		"request:ResourceId": "api:admin:admin-123",
	}

	// Test NotResource exclusion
	statement := models.PolicyStatement{
		Sid:      "GlobalWithExclusion",
		Effect:   "Allow",
		Action:   models.JSONActionResource{Single: "*:*:read"},
		Resource: models.JSONActionResource{Single: "api:*:*"},
		NotResource: models.JSONActionResource{
			Multiple: []string{"api:admin:*", "api:system:*"},
		},
	}

	// This should be excluded by NotResource
	resourceMatches := false
	resourceValues := statement.Resource.GetValues()
	for _, resourcePattern := range resourceValues {
		if matcher.Match(resourcePattern, context["request:ResourceId"].(string), context) {
			resourceMatches = true
			break
		}
	}

	if !resourceMatches {
		t.Error("Resource should match the main pattern")
	}

	// Check NotResource exclusions
	excluded := false
	notResourceValues := statement.NotResource.GetValues()
	for _, notResourcePattern := range notResourceValues {
		if matcher.Match(notResourcePattern, context["request:ResourceId"].(string), context) {
			excluded = true
			break
		}
	}

	if !excluded {
		t.Error("Resource should be excluded by NotResource pattern")
	}
}

// TestResourceFormatValidation removed - testing internal implementation details
// Validation tests should be in the matchers package instead
