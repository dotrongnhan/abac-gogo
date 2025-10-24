package evaluator

import (
	"testing"
	"time"
)

func TestEnhancedConditionEvaluator_StringOperators(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"department": "Engineering",
			"email":      "john.doe@company.com",
			"role":       "Senior Developer",
		},
		"resource": map[string]interface{}{
			"path":           "/documents/project-alpha/specs.pdf",
			"classification": "confidential",
		},
		"environment": map[string]interface{}{
			"user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
			"client_ip":  "192.168.1.100",
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
					"user.department": "Engineering",
				},
			},
			expected: true,
		},
		{
			name: "StringEquals - no match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.department": "Finance",
				},
			},
			expected: false,
		},
		{
			name: "StringContains - match",
			conditions: map[string]interface{}{
				"StringContains": map[string]interface{}{
					"user.email": "@company.com",
				},
			},
			expected: true,
		},
		{
			name: "StringContains - no match",
			conditions: map[string]interface{}{
				"StringContains": map[string]interface{}{
					"user.email": "@external.com",
				},
			},
			expected: false,
		},
		{
			name: "StringStartsWith - match",
			conditions: map[string]interface{}{
				"StringStartsWith": map[string]interface{}{
					"resource.path": "/documents",
				},
			},
			expected: true,
		},
		{
			name: "StringEndsWith - match",
			conditions: map[string]interface{}{
				"StringEndsWith": map[string]interface{}{
					"resource.path": ".pdf",
				},
			},
			expected: true,
		},
		{
			name: "StringRegex - match",
			conditions: map[string]interface{}{
				"StringRegex": map[string]interface{}{
					"user.email": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				},
			},
			expected: true,
		},
		{
			name: "StringLike - match with wildcards",
			conditions: map[string]interface{}{
				"StringLike": map[string]interface{}{
					"resource.path": "/documents/project-%/%.pdf",
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_NumericOperators(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"level":  8,
			"salary": 75000.50,
			"age":    32,
		},
		"transaction": map[string]interface{}{
			"amount": 250000,
			"fee":    12.50,
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "NumericEquals - match",
			conditions: map[string]interface{}{
				"NumericEquals": map[string]interface{}{
					"user.level": 8,
				},
			},
			expected: true,
		},
		{
			name: "NumericGreaterThan - match",
			conditions: map[string]interface{}{
				"NumericGreaterThan": map[string]interface{}{
					"user.age": 30,
				},
			},
			expected: true,
		},
		{
			name: "NumericLessThanEquals - match",
			conditions: map[string]interface{}{
				"NumericLessThanEquals": map[string]interface{}{
					"transaction.fee": 15.0,
				},
			},
			expected: true,
		},
		{
			name: "NumericBetween - match with array",
			conditions: map[string]interface{}{
				"NumericBetween": map[string]interface{}{
					"user.age": []interface{}{25, 40},
				},
			},
			expected: true,
		},
		{
			name: "NumericBetween - match with map",
			conditions: map[string]interface{}{
				"NumericBetween": map[string]interface{}{
					"user.salary": map[string]interface{}{
						"min": 70000,
						"max": 80000,
					},
				},
			},
			expected: true,
		},
		{
			name: "NumericBetween - no match",
			conditions: map[string]interface{}{
				"NumericBetween": map[string]interface{}{
					"user.age": []interface{}{40, 50},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_TimeOperators(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	now := time.Now()
	context := map[string]interface{}{
		"environment": map[string]interface{}{
			"time_of_day":       "14:30",
			"day_of_week":       "Wednesday",
			"hour":              14,
			"is_business_hours": true,
			"is_weekend":        false,
		},
		"request": map[string]interface{}{
			"timestamp": now.Format(time.RFC3339),
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "TimeOfDay - exact match",
			conditions: map[string]interface{}{
				"TimeOfDay": map[string]interface{}{
					"environment.time_of_day": "14:30",
				},
			},
			expected: true,
		},
		{
			name: "DayOfWeek - single day match",
			conditions: map[string]interface{}{
				"DayOfWeek": map[string]interface{}{
					"environment.day_of_week": "Wednesday",
				},
			},
			expected: true,
		},
		{
			name: "DayOfWeek - array match",
			conditions: map[string]interface{}{
				"DayOfWeek": map[string]interface{}{
					"environment.day_of_week": []interface{}{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
				},
			},
			expected: true,
		},
		{
			name: "IsBusinessHours - match",
			conditions: map[string]interface{}{
				"IsBusinessHours": map[string]interface{}{
					"environment.is_business_hours": true,
				},
			},
			expected: true,
		},
		{
			name: "TimeGreaterThan - match",
			conditions: map[string]interface{}{
				"TimeGreaterThan": map[string]interface{}{
					"environment.time_of_day": "09:00",
				},
			},
			expected: true,
		},
		{
			name: "TimeLessThan - match",
			conditions: map[string]interface{}{
				"TimeLessThan": map[string]interface{}{
					"environment.time_of_day": "17:00",
				},
			},
			expected: true,
		},
		{
			name: "TimeBetween - match",
			conditions: map[string]interface{}{
				"TimeBetween": map[string]interface{}{
					"environment.time_of_day": []interface{}{"09:00", "17:00"},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_NetworkOperators(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"environment": map[string]interface{}{
			"client_ip":      "192.168.1.100",
			"is_internal_ip": true,
			"ip_class":       "ipv4",
		},
		"request": map[string]interface{}{
			"source_ip": "10.0.1.50",
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "IPInRange - single range match",
			conditions: map[string]interface{}{
				"IPInRange": map[string]interface{}{
					"environment.client_ip": "192.168.1.0/24",
				},
			},
			expected: true,
		},
		{
			name: "IPInRange - multiple ranges match",
			conditions: map[string]interface{}{
				"IPInRange": map[string]interface{}{
					"request.source_ip": []interface{}{"10.0.0.0/8", "192.168.1.0/24"},
				},
			},
			expected: true,
		},
		{
			name: "IPNotInRange - not in range",
			conditions: map[string]interface{}{
				"IPNotInRange": map[string]interface{}{
					"environment.client_ip": "172.16.0.0/12",
				},
			},
			expected: true,
		},
		{
			name: "IsInternalIP - match",
			conditions: map[string]interface{}{
				"IsInternalIP": map[string]interface{}{
					"environment.is_internal_ip": true,
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_ArrayOperators(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"roles":       []interface{}{"developer", "code_reviewer", "team_lead"},
			"permissions": []interface{}{"read", "write", "execute"},
		},
		"resource": map[string]interface{}{
			"tags": []interface{}{"confidential", "project-alpha", "engineering"},
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "ArrayContains - match",
			conditions: map[string]interface{}{
				"ArrayContains": map[string]interface{}{
					"user.roles": "developer",
				},
			},
			expected: true,
		},
		{
			name: "ArrayContains - no match",
			conditions: map[string]interface{}{
				"ArrayContains": map[string]interface{}{
					"user.roles": "admin",
				},
			},
			expected: false,
		},
		{
			name: "ArrayNotContains - match",
			conditions: map[string]interface{}{
				"ArrayNotContains": map[string]interface{}{
					"user.roles": "admin",
				},
			},
			expected: true,
		},
		{
			name: "ArraySize - exact match",
			conditions: map[string]interface{}{
				"ArraySize": map[string]interface{}{
					"user.roles": 3,
				},
			},
			expected: true,
		},
		{
			name: "ArraySize - with operators",
			conditions: map[string]interface{}{
				"ArraySize": map[string]interface{}{
					"user.permissions": map[string]interface{}{
						"gte": 2,
					},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_ComplexLogic(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"department": "Engineering",
			"level":      8,
			"mfa":        true,
		},
		"environment": map[string]interface{}{
			"is_business_hours": true,
			"client_ip":         "192.168.1.100",
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "And - all conditions match",
			conditions: map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "Engineering",
						},
					},
					map[string]interface{}{
						"NumericGreaterThanEquals": map[string]interface{}{
							"user.level": 5,
						},
					},
					map[string]interface{}{
						"Bool": map[string]interface{}{
							"user.mfa": true,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "And - one condition fails",
			conditions: map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "Engineering",
						},
					},
					map[string]interface{}{
						"NumericGreaterThanEquals": map[string]interface{}{
							"user.level": 10, // This will fail
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Or - first condition matches",
			conditions: map[string]interface{}{
				"Or": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "Engineering",
						},
					},
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "Finance",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Not - negates condition",
			conditions: map[string]interface{}{
				"Not": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user.department": "Finance",
					},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_DotNotation(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"personal": map[string]interface{}{
					"name":  "John Doe",
					"email": "john.doe@company.com",
				},
				"work": map[string]interface{}{
					"department": "Engineering",
					"level":      8,
				},
			},
		},
		"resource": map[string]interface{}{
			"metadata": map[string]interface{}{
				"security": map[string]interface{}{
					"classification": "confidential",
					"owner":          "engineering-team",
				},
			},
		},
	}

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expected   bool
	}{
		{
			name: "Deep nested access - match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.profile.work.department": "Engineering",
				},
			},
			expected: true,
		},
		{
			name: "Deep nested numeric - match",
			conditions: map[string]interface{}{
				"NumericGreaterThan": map[string]interface{}{
					"user.profile.work.level": 5,
				},
			},
			expected: true,
		},
		{
			name: "Resource metadata access - match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"resource.metadata.security.classification": "confidential",
				},
			},
			expected: true,
		},
		{
			name: "Non-existent path - no match",
			conditions: map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.profile.nonexistent.field": "value",
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evaluator.EvaluateConditions(test.conditions, context)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEnhancedConditionEvaluator_PerformanceWithCache(t *testing.T) {
	evaluator := NewEnhancedConditionEvaluator()

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"email": "test@company.com",
		},
	}

	// Test regex caching
	conditions := map[string]interface{}{
		"StringRegex": map[string]interface{}{
			"user.email": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		},
	}

	// First evaluation - should compile and cache regex
	result1 := evaluator.EvaluateConditions(conditions, context)
	if !result1 {
		t.Error("First evaluation should succeed")
	}

	// Second evaluation - should use cached regex
	result2 := evaluator.EvaluateConditions(conditions, context)
	if !result2 {
		t.Error("Second evaluation should succeed")
	}

	// Verify cache has the pattern
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if _, exists := evaluator.regexCache[pattern]; !exists {
		t.Error("Regex pattern should be cached")
	}
}
