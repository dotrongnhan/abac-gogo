package evaluator

import (
	"testing"
)

func TestComplexConditionEvaluator(t *testing.T) {
	ce := NewConditionEvaluator()

	tests := []struct {
		name       string
		conditions map[string]interface{}
		context    map[string]interface{}
		expected   bool
	}{
		{
			name: "Simple AND - both conditions true",
			conditions: map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "engineering",
						},
					},
					map[string]interface{}{
						"NumericGreaterThan": map[string]interface{}{
							"user.level": 3,
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
					"level":      5,
				},
			},
			expected: true,
		},
		{
			name: "Simple AND - one condition false",
			conditions: map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "engineering",
						},
					},
					map[string]interface{}{
						"NumericGreaterThan": map[string]interface{}{
							"user.level": 10,
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
					"level":      5,
				},
			},
			expected: false,
		},
		{
			name: "Simple OR - one condition true",
			conditions: map[string]interface{}{
				"Or": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "hr",
						},
					},
					map[string]interface{}{
						"NumericGreaterThan": map[string]interface{}{
							"user.level": 3,
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
					"level":      5,
				},
			},
			expected: true,
		},
		{
			name: "Simple OR - both conditions false",
			conditions: map[string]interface{}{
				"Or": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "hr",
						},
					},
					map[string]interface{}{
						"NumericGreaterThan": map[string]interface{}{
							"user.level": 10,
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
					"level":      5,
				},
			},
			expected: false,
		},
		{
			name: "Simple NOT - condition true",
			conditions: map[string]interface{}{
				"Not": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user.department": "hr",
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
				},
			},
			expected: true,
		},
		{
			name: "Simple NOT - condition false",
			conditions: map[string]interface{}{
				"Not": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user.department": "engineering",
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
				},
			},
			expected: false,
		},
		{
			name: "Nested AND/OR - complex logic",
			conditions: map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"Or": []interface{}{
							map[string]interface{}{
								"StringEquals": map[string]interface{}{
									"user.department": "engineering",
								},
							},
							map[string]interface{}{
								"StringEquals": map[string]interface{}{
									"user.department": "security",
								},
							},
						},
					},
					map[string]interface{}{
						"NumericGreaterThan": map[string]interface{}{
							"user.level": 3,
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "security",
					"level":      5,
				},
			},
			expected: true,
		},
		{
			name: "Complex nested with NOT",
			conditions: map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "engineering",
						},
					},
					map[string]interface{}{
						"Not": map[string]interface{}{
							"Bool": map[string]interface{}{
								"user.on_probation": true,
							},
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department":   "engineering",
					"on_probation": false,
				},
			},
			expected: true,
		},
		{
			name: "Multiple levels of nesting",
			conditions: map[string]interface{}{
				"Or": []interface{}{
					map[string]interface{}{
						"And": []interface{}{
							map[string]interface{}{
								"StringEquals": map[string]interface{}{
									"user.role": "admin",
								},
							},
							map[string]interface{}{
								"IpAddress": map[string]interface{}{
									"request.sourceIp": []interface{}{"10.0.0.0/8", "192.168.0.0/16"},
								},
							},
						},
					},
					map[string]interface{}{
						"And": []interface{}{
							map[string]interface{}{
								"StringEquals": map[string]interface{}{
									"user.department": "engineering",
								},
							},
							map[string]interface{}{
								"NumericGreaterThan": map[string]interface{}{
									"user.level": 7,
								},
							},
							map[string]interface{}{
								"Not": map[string]interface{}{
									"Bool": map[string]interface{}{
										"user.on_probation": true,
									},
								},
							},
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"role":         "developer",
					"department":   "engineering",
					"level":        8,
					"on_probation": false,
				},
				"request": map[string]interface{}{
					"sourceIp": "203.0.113.1",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ce.Evaluate(tt.conditions, tt.context)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestComplexConditionStruct(t *testing.T) {
	ce := NewConditionEvaluator()

	tests := []struct {
		name      string
		condition *ComplexCondition
		context   map[string]interface{}
		expected  bool
	}{
		{
			name: "Simple condition struct",
			condition: &ComplexCondition{
				Type:     "simple",
				Operator: ConditionStringEquals,
				Key:      "user.department",
				Value:    "engineering",
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
				},
			},
			expected: true,
		},
		{
			name: "AND condition with left/right",
			condition: &ComplexCondition{
				Type:     "logical",
				Operator: ConditionAnd,
				Left: &ComplexCondition{
					Type:     "simple",
					Operator: ConditionStringEquals,
					Key:      "user.department",
					Value:    "engineering",
				},
				Right: &ComplexCondition{
					Type:     "simple",
					Operator: ConditionNumericGreaterThan,
					Key:      "user.level",
					Value:    3,
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
					"level":      5,
				},
			},
			expected: true,
		},
		{
			name: "OR condition with conditions array",
			condition: &ComplexCondition{
				Type:     "logical",
				Operator: ConditionOr,
				Conditions: []ComplexCondition{
					{
						Type:     "simple",
						Operator: ConditionStringEquals,
						Key:      "user.department",
						Value:    "hr",
					},
					{
						Type:     "simple",
						Operator: ConditionStringEquals,
						Key:      "user.role",
						Value:    "admin",
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
					"role":       "admin",
				},
			},
			expected: true,
		},
		{
			name: "NOT condition",
			condition: &ComplexCondition{
				Type:     "logical",
				Operator: ConditionNot,
				Operand: &ComplexCondition{
					Type:     "simple",
					Operator: ConditionBool,
					Key:      "user.on_probation",
					Value:    true,
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"on_probation": false,
				},
			},
			expected: true,
		},
		{
			name: "Deeply nested condition",
			condition: &ComplexCondition{
				Type:     "logical",
				Operator: ConditionAnd,
				Left: &ComplexCondition{
					Type:     "logical",
					Operator: ConditionOr,
					Left: &ComplexCondition{
						Type:     "simple",
						Operator: ConditionStringEquals,
						Key:      "user.department",
						Value:    "engineering",
					},
					Right: &ComplexCondition{
						Type:     "simple",
						Operator: ConditionStringEquals,
						Key:      "user.department",
						Value:    "security",
					},
				},
				Right: &ComplexCondition{
					Type:     "logical",
					Operator: ConditionNot,
					Operand: &ComplexCondition{
						Type:     "simple",
						Operator: ConditionBool,
						Key:      "user.on_probation",
						Value:    true,
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department":   "security",
					"on_probation": false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ce.EvaluateComplex(tt.condition, tt.context)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestComplexConditionEdgeCases(t *testing.T) {
	ce := NewConditionEvaluator()

	tests := []struct {
		name       string
		conditions map[string]interface{}
		context    map[string]interface{}
		expected   bool
	}{
		{
			name: "Empty AND array",
			conditions: map[string]interface{}{
				"And": []interface{}{},
			},
			context:  map[string]interface{}{},
			expected: true, // Empty AND should return true
		},
		{
			name: "Empty OR array",
			conditions: map[string]interface{}{
				"Or": []interface{}{},
			},
			context:  map[string]interface{}{},
			expected: false, // Empty OR should return false
		},
		{
			name: "Invalid condition in AND",
			conditions: map[string]interface{}{
				"And": []interface{}{
					"invalid_condition",
				},
			},
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name: "Mixed valid and invalid conditions in OR",
			conditions: map[string]interface{}{
				"Or": []interface{}{
					"invalid_condition",
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "engineering",
						},
					},
				},
			},
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"department": "engineering",
				},
			},
			expected: true, // Should pass because one valid condition is true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ce.Evaluate(tt.conditions, tt.context)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestComplexConditionWithAllOperators(t *testing.T) {
	ce := NewConditionEvaluator()

	// Test a comprehensive condition that uses all operator types
	conditions := map[string]interface{}{
		"And": []interface{}{
			map[string]interface{}{
				"Or": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "engineering",
						},
					},
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.role": "admin",
						},
					},
				},
			},
			map[string]interface{}{
				"NumericGreaterThanEquals": map[string]interface{}{
					"user.level": 5,
				},
			},
			map[string]interface{}{
				"Not": map[string]interface{}{
					"Bool": map[string]interface{}{
						"user.on_probation": true,
					},
				},
			},
			map[string]interface{}{
				"IpAddress": map[string]interface{}{
					"request.sourceIp": []interface{}{"10.0.0.0/8", "192.168.0.0/16"},
				},
			},
			map[string]interface{}{
				"StringLike": map[string]interface{}{
					"request.userAgent": "*Chrome*",
				},
			},
		},
	}

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"department":   "engineering",
			"role":         "developer",
			"level":        7,
			"on_probation": false,
		},
		"request": map[string]interface{}{
			"sourceIp":  "10.0.1.100",
			"userAgent": "Mozilla/5.0 Chrome/91.0",
		},
	}

	result := ce.Evaluate(conditions, context)
	if !result {
		t.Errorf("Expected true for comprehensive condition test, got false")
	}
}
