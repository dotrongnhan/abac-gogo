package evaluator

import (
	"fmt"
	"testing"

	"abac_go_example/models"
)

func TestPolicyFilter_FilterApplicablePolicies(t *testing.T) {
	filter := NewPolicyFilter()

	// Create test policies
	policies := []*models.Policy{
		{
			ID:      "pol-001",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "ReadDocuments",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "document-service:file:read", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:documents:*", IsArray: false},
				},
			},
		},
		{
			ID:      "pol-002",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "WriteReports",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "report-service:*:write", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:reports:*", IsArray: false},
				},
			},
		},
		{
			ID:      "pol-003",
			Enabled: false, // Disabled policy
			Statement: []models.PolicyStatement{
				{
					Sid:      "DisabledPolicy",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "*", IsArray: false},
					Resource: models.JSONActionResource{Single: "*", IsArray: false},
				},
			},
		},
		{
			ID:      "pol-004",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "AdminAccess",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "admin-service:*:*", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:admin:*", IsArray: false},
				},
			},
		},
	}

	tests := []struct {
		name              string
		request           *models.EvaluationRequest
		expectedCount     int
		expectedPolicyIDs []string
	}{
		{
			name: "Document read request - should match pol-001",
			request: &models.EvaluationRequest{
				RequestID:  "req-001",
				SubjectID:  "user-123",
				ResourceID: "api:documents:doc-456",
				Action:     "document-service:file:read",
			},
			expectedCount:     1,
			expectedPolicyIDs: []string{"pol-001"},
		},
		{
			name: "Report write request - should match pol-002",
			request: &models.EvaluationRequest{
				RequestID:  "req-002",
				SubjectID:  "user-123",
				ResourceID: "api:reports:monthly",
				Action:     "report-service:monthly:write",
			},
			expectedCount:     1,
			expectedPolicyIDs: []string{"pol-002"},
		},
		{
			name: "Admin request - should match pol-004",
			request: &models.EvaluationRequest{
				RequestID:  "req-003",
				SubjectID:  "admin-123",
				ResourceID: "api:admin:users",
				Action:     "admin-service:users:delete",
			},
			expectedCount:     1,
			expectedPolicyIDs: []string{"pol-004"},
		},
		{
			name: "No matching request - should return empty",
			request: &models.EvaluationRequest{
				RequestID:  "req-004",
				SubjectID:  "user-123",
				ResourceID: "api:payments:transaction-123",
				Action:     "payment-service:transaction:process",
			},
			expectedCount:     0,
			expectedPolicyIDs: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := filter.FilterApplicablePolicies(policies, test.request)

			if len(result) != test.expectedCount {
				t.Errorf("Expected %d policies, got %d", test.expectedCount, len(result))
			}

			// Check if expected policies are in result
			resultIDs := make(map[string]bool)
			for _, policy := range result {
				resultIDs[policy.ID] = true
			}

			for _, expectedID := range test.expectedPolicyIDs {
				if !resultIDs[expectedID] {
					t.Errorf("Expected policy %s not found in results", expectedID)
				}
			}
		})
	}
}

func TestPolicyFilter_FastPatternMatch(t *testing.T) {
	filter := NewPolicyFilter()

	tests := []struct {
		name     string
		pattern  string
		value    string
		expected bool
	}{
		// Universal wildcard
		{
			name:     "Universal wildcard",
			pattern:  "*",
			value:    "anything",
			expected: true,
		},
		// Exact matches
		{
			name:     "Exact match - success",
			pattern:  "document-service:file:read",
			value:    "document-service:file:read",
			expected: true,
		},
		{
			name:     "Exact match - failure",
			pattern:  "document-service:file:read",
			value:    "document-service:file:write",
			expected: false,
		},
		// Prefix wildcards
		{
			name:     "Prefix wildcard - match",
			pattern:  "document-service:*",
			value:    "document-service:file:read",
			expected: true,
		},
		{
			name:     "Prefix wildcard - no match",
			pattern:  "document-service:*",
			value:    "report-service:file:read",
			expected: false,
		},
		// Suffix wildcards
		{
			name:     "Suffix wildcard - match",
			pattern:  "*:read",
			value:    "document-service:file:read",
			expected: true,
		},
		{
			name:     "Suffix wildcard - no match",
			pattern:  "*:read",
			value:    "document-service:file:write",
			expected: false,
		},
		// Contains wildcards
		{
			name:     "Contains wildcard - match",
			pattern:  "*file*",
			value:    "document-service:file:read",
			expected: true,
		},
		{
			name:     "Contains wildcard - no match",
			pattern:  "*admin*",
			value:    "document-service:file:read",
			expected: false,
		},
		// Complex patterns
		{
			name:     "Complex pattern - match",
			pattern:  "document-*:file:*",
			value:    "document-service:file:read",
			expected: true,
		},
		{
			name:     "Complex pattern - no match",
			pattern:  "document-*:admin:*",
			value:    "document-service:file:read",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := filter.fastPatternMatch(test.pattern, test.value)
			if result != test.expected {
				t.Errorf("fastPatternMatch(%q, %q) = %v, want %v",
					test.pattern, test.value, result, test.expected)
			}
		})
	}
}

func TestPolicyFilter_NotResourceExclusion(t *testing.T) {
	filter := NewPolicyFilter()

	tests := []struct {
		name              string
		notResourceSpec   models.JSONActionResource
		requestedResource string
		expectedExcluded  bool
	}{
		{
			name: "No NotResource - not excluded",
			notResourceSpec: models.JSONActionResource{
				Single:  "",
				IsArray: false,
			},
			requestedResource: "api:admin:users",
			expectedExcluded:  false,
		},
		{
			name: "Single NotResource - excluded",
			notResourceSpec: models.JSONActionResource{
				Single:  "api:admin:*",
				IsArray: false,
			},
			requestedResource: "api:admin:users",
			expectedExcluded:  true,
		},
		{
			name: "Single NotResource - not excluded",
			notResourceSpec: models.JSONActionResource{
				Single:  "api:admin:*",
				IsArray: false,
			},
			requestedResource: "api:documents:file-123",
			expectedExcluded:  false,
		},
		{
			name: "Multiple NotResource - excluded by first",
			notResourceSpec: models.JSONActionResource{
				Multiple: []string{"api:admin:*", "api:system:*"},
				IsArray:  true,
			},
			requestedResource: "api:admin:users",
			expectedExcluded:  true,
		},
		{
			name: "Multiple NotResource - excluded by second",
			notResourceSpec: models.JSONActionResource{
				Multiple: []string{"api:admin:*", "api:system:*"},
				IsArray:  true,
			},
			requestedResource: "api:system:config",
			expectedExcluded:  true,
		},
		{
			name: "Multiple NotResource - not excluded",
			notResourceSpec: models.JSONActionResource{
				Multiple: []string{"api:admin:*", "api:system:*"},
				IsArray:  true,
			},
			requestedResource: "api:documents:file-123",
			expectedExcluded:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := filter.isExcludedByNotResource(test.notResourceSpec, test.requestedResource)
			if result != test.expectedExcluded {
				t.Errorf("isExcludedByNotResource() = %v, want %v", result, test.expectedExcluded)
			}
		})
	}
}

func TestPolicyFilter_PatternCache(t *testing.T) {
	filter := NewPolicyFilter()

	pattern := "document-service:*"
	value := "document-service:file:read"

	// First call - should cache result
	result1 := filter.fastPatternMatch(pattern, value)
	if !result1 {
		t.Error("First call should return true")
	}

	// Verify cache entry exists
	if len(filter.patternCache) == 0 {
		t.Error("Cache should have entries after first call")
	}

	// Second call - should use cached result
	result2 := filter.fastPatternMatch(pattern, value)
	if !result2 {
		t.Error("Second call should return true")
	}

	// Verify cache was used (same result)
	if result1 != result2 {
		t.Error("Cached result should be consistent")
	}
}

func TestPolicyFilter_AdvancedFiltering(t *testing.T) {
	filter := NewPolicyFilter()

	// Create policies with different characteristics
	policies := []*models.Policy{
		{
			ID:      "user-policy",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "UserAccess",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "user-service:*:*", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:users:*", IsArray: false},
					Condition: map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user:subjecttype": "employee",
						},
					},
				},
			},
		},
		{
			ID:      "document-policy",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "DocumentAccess",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "document-service:*:*", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:documents:*", IsArray: false},
				},
			},
		},
		{
			ID:      "admin-policy",
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      "AdminAccess",
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: "admin-service:*:*", IsArray: false},
					Resource: models.JSONActionResource{Single: "api:admin:*", IsArray: false},
				},
			},
		},
	}

	tests := []struct {
		name          string
		filterFunc    func([]*models.Policy) []*models.Policy
		expectedCount int
		expectedIDs   []string
	}{
		{
			name: "Filter by subject type",
			filterFunc: func(policies []*models.Policy) []*models.Policy {
				return filter.FilterBySubjectType(policies, "employee")
			},
			expectedCount: 1,
			expectedIDs:   []string{"user-policy"},
		},
		{
			name: "Filter by resource type",
			filterFunc: func(policies []*models.Policy) []*models.Policy {
				return filter.FilterByResourceType(policies, "document")
			},
			expectedCount: 1,
			expectedIDs:   []string{"document-policy"},
		},
		{
			name: "Filter by action category",
			filterFunc: func(policies []*models.Policy) []*models.Policy {
				return filter.FilterByActionCategory(policies, "admin")
			},
			expectedCount: 1,
			expectedIDs:   []string{"admin-policy"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.filterFunc(policies)

			if len(result) != test.expectedCount {
				t.Errorf("Expected %d policies, got %d", test.expectedCount, len(result))
			}

			resultIDs := make(map[string]bool)
			for _, policy := range result {
				resultIDs[policy.ID] = true
			}

			for _, expectedID := range test.expectedIDs {
				if !resultIDs[expectedID] {
					t.Errorf("Expected policy %s not found in results", expectedID)
				}
			}
		})
	}
}

func TestPolicyFilter_PerformanceOptimization(t *testing.T) {
	filter := NewPolicyFilter()

	// Create many policies to test performance
	var policies []*models.Policy
	for i := 0; i < 100; i++ {
		policy := &models.Policy{
			ID:      fmt.Sprintf("pol-%03d", i),
			Enabled: true,
			Statement: []models.PolicyStatement{
				{
					Sid:      fmt.Sprintf("Statement-%d", i),
					Effect:   "Allow",
					Action:   models.JSONActionResource{Single: fmt.Sprintf("service-%d:*:*", i%10), IsArray: false},
					Resource: models.JSONActionResource{Single: fmt.Sprintf("api:resource-%d:*", i%20), IsArray: false},
				},
			},
		}
		policies = append(policies, policy)
	}

	// Test request that should match only a few policies
	request := &models.EvaluationRequest{
		RequestID:  "perf-test-001",
		SubjectID:  "user-123",
		ResourceID: "api:resource-5:item-123",
		Action:     "service-5:item:read",
	}

	// Filter policies
	applicablePolicies := filter.FilterApplicablePolicies(policies, request)

	// Should significantly reduce the number of policies
	if len(applicablePolicies) >= len(policies)/2 {
		t.Errorf("Filtering should reduce policies significantly. Got %d out of %d",
			len(applicablePolicies), len(policies))
	}

	// Verify that some policies were filtered out
	if len(applicablePolicies) == 0 {
		t.Error("Should have at least some applicable policies")
	}

	// Test cache performance
	stats := filter.GetFilteringStats()
	if stats["cache_enabled"] != true {
		t.Error("Cache should be enabled")
	}

	// Clear cache and verify
	filter.ClearCache()
	statsAfterClear := filter.GetFilteringStats()
	if statsAfterClear["cache_size"].(int) != 0 {
		t.Error("Cache should be empty after clear")
	}
}

func TestPolicyFilter_EdgeCases(t *testing.T) {
	filter := NewPolicyFilter()

	tests := []struct {
		name     string
		policies []*models.Policy
		request  *models.EvaluationRequest
		expected int
	}{
		{
			name:     "Empty policies list",
			policies: []*models.Policy{},
			request: &models.EvaluationRequest{
				RequestID:  "req-001",
				SubjectID:  "user-123",
				ResourceID: "api:documents:doc-456",
				Action:     "document-service:file:read",
			},
			expected: 0,
		},
		{
			name: "All disabled policies",
			policies: []*models.Policy{
				{
					ID:      "pol-001",
					Enabled: false,
					Statement: []models.PolicyStatement{
						{
							Sid:      "DisabledStatement",
							Effect:   "Allow",
							Action:   models.JSONActionResource{Single: "*", IsArray: false},
							Resource: models.JSONActionResource{Single: "*", IsArray: false},
						},
					},
				},
			},
			request: &models.EvaluationRequest{
				RequestID:  "req-002",
				SubjectID:  "user-123",
				ResourceID: "api:documents:doc-456",
				Action:     "document-service:file:read",
			},
			expected: 0,
		},
		{
			name: "Policies with empty statements",
			policies: []*models.Policy{
				{
					ID:        "pol-empty",
					Enabled:   true,
					Statement: []models.PolicyStatement{},
				},
			},
			request: &models.EvaluationRequest{
				RequestID:  "req-003",
				SubjectID:  "user-123",
				ResourceID: "api:documents:doc-456",
				Action:     "document-service:file:read",
			},
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := filter.FilterApplicablePolicies(test.policies, test.request)
			if len(result) != test.expected {
				t.Errorf("Expected %d policies, got %d", test.expected, len(result))
			}
		})
	}
}
