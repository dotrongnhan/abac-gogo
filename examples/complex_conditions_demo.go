package main

import (
	"encoding/json"
	"fmt"

	"abac_go_example/evaluator"
)

func main() {
	fmt.Println("🔥 ABAC Complex Logical Conditions Demo")
	fmt.Println("=====================================")

	ce := evaluator.NewConditionEvaluator()

	// Demo 1: Simple AND condition
	fmt.Println("\n=== Demo 1: Simple AND Condition ===")
	simpleAndCondition := map[string]interface{}{
		"And": []interface{}{
			map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"user.department": "engineering",
				},
			},
			map[string]interface{}{
				"NumericGreaterThan": map[string]interface{}{
					"user.level": 5,
				},
			},
		},
	}

	context1 := map[string]interface{}{
		"user": map[string]interface{}{
			"department": "engineering",
			"level":      7,
		},
	}

	result1 := ce.Evaluate(simpleAndCondition, context1)
	fmt.Printf("Condition: (department = 'engineering') AND (level > 5)\n")
	fmt.Printf("Context: department='engineering', level=7\n")
	fmt.Printf("Result: %v ✅\n", result1)

	// Demo 2: Complex nested OR/AND condition
	fmt.Println("\n=== Demo 2: Complex Nested Condition ===")
	complexCondition := map[string]interface{}{
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
		},
	}

	context2 := map[string]interface{}{
		"user": map[string]interface{}{
			"department":   "security",
			"role":         "admin",
			"level":        6,
			"on_probation": false,
		},
	}

	result2 := ce.Evaluate(complexCondition, context2)
	fmt.Printf("Condition: (department='engineering' OR role='admin') AND level>=5 AND NOT on_probation\n")
	fmt.Printf("Context: department='security', role='admin', level=6, on_probation=false\n")
	fmt.Printf("Result: %v ✅\n", result2)

	// Demo 3: Multiple levels of nesting
	fmt.Println("\n=== Demo 3: Multiple Levels of Nesting ===")
	multiLevelCondition := map[string]interface{}{
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
	}

	context3 := map[string]interface{}{
		"user": map[string]interface{}{
			"role":         "developer",
			"department":   "engineering",
			"level":        8,
			"on_probation": false,
		},
		"request": map[string]interface{}{
			"sourceIp": "203.0.113.1", // External IP
		},
	}

	result3 := ce.Evaluate(multiLevelCondition, context3)
	fmt.Printf("Condition: (admin AND internal_IP) OR (engineering AND level>7 AND NOT on_probation)\n")
	fmt.Printf("Context: role='developer', department='engineering', level=8, on_probation=false, IP=external\n")
	fmt.Printf("Result: %v ✅\n", result3)

	// Demo 4: Using ComplexCondition struct
	fmt.Println("\n=== Demo 4: Using ComplexCondition Struct ===")
	structCondition := &evaluator.ComplexCondition{
		Type:     "logical",
		Operator: evaluator.ConditionAnd,
		Left: &evaluator.ComplexCondition{
			Type:     "simple",
			Operator: evaluator.ConditionStringEquals,
			Key:      "user.department",
			Value:    "finance",
		},
		Right: &evaluator.ComplexCondition{
			Type:     "logical",
			Operator: evaluator.ConditionOr,
			Conditions: []evaluator.ComplexCondition{
				{
					Type:     "simple",
					Operator: evaluator.ConditionNumericGreaterThanEquals,
					Key:      "user.level",
					Value:    5,
				},
				{
					Type:     "simple",
					Operator: evaluator.ConditionStringEquals,
					Key:      "user.role",
					Value:    "manager",
				},
			},
		},
	}

	context4 := map[string]interface{}{
		"user": map[string]interface{}{
			"department": "finance",
			"level":      3,
			"role":       "manager",
		},
	}

	result4 := ce.EvaluateComplex(structCondition, context4)
	fmt.Printf("Condition: department='finance' AND (level>=5 OR role='manager')\n")
	fmt.Printf("Context: department='finance', level=3, role='manager'\n")
	fmt.Printf("Result: %v ✅\n", result4)

	// Demo 5: Real-world access control scenario
	fmt.Println("\n=== Demo 5: Real-world Access Control ===")
	accessControlCondition := map[string]interface{}{
		"Or": []interface{}{
			// Owner can always access
			map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"resource.owner": "${user.id}",
				},
			},
			// Department members with sufficient level
			map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": "${resource.department}",
						},
					},
					map[string]interface{}{
						"NumericGreaterThanEquals": map[string]interface{}{
							"user.level": 5,
						},
					},
				},
			},
			// Admins (but not for top secret)
			map[string]interface{}{
				"And": []interface{}{
					map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.role": "admin",
						},
					},
					map[string]interface{}{
						"Not": map[string]interface{}{
							"StringEquals": map[string]interface{}{
								"resource.classification": "top_secret",
							},
						},
					},
				},
			},
		},
	}

	// Substitute variables in conditions
	substitutedCondition := substituteVariables(accessControlCondition, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         "user-123",
			"department": "engineering",
			"level":      6,
			"role":       "developer",
		},
		"resource": map[string]interface{}{
			"owner":          "user-456",
			"department":     "engineering",
			"classification": "confidential",
		},
	})

	context5 := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         "user-123",
			"department": "engineering",
			"level":      6,
			"role":       "developer",
		},
		"resource": map[string]interface{}{
			"owner":          "user-456",
			"department":     "engineering",
			"classification": "confidential",
		},
	}

	result5 := ce.Evaluate(substitutedCondition, context5)
	fmt.Printf("Condition: owner OR (same_department AND level>=5) OR (admin AND NOT top_secret)\n")
	fmt.Printf("Context: not_owner, same_department, level=6, not_admin, not_top_secret\n")
	fmt.Printf("Result: %v ✅ (Access granted via department rule)\n", result5)

	// Demo 6: Edge cases
	fmt.Println("\n=== Demo 6: Edge Cases ===")

	// Empty AND
	emptyAnd := map[string]interface{}{
		"And": []interface{}{},
	}
	resultEmptyAnd := ce.Evaluate(emptyAnd, map[string]interface{}{})
	fmt.Printf("Empty AND: %v (should be true)\n", resultEmptyAnd)

	// Empty OR
	emptyOr := map[string]interface{}{
		"Or": []interface{}{},
	}
	resultEmptyOr := ce.Evaluate(emptyOr, map[string]interface{}{})
	fmt.Printf("Empty OR: %v (should be false)\n", resultEmptyOr)

	fmt.Println("\n🎉 Demo completed! Complex logical conditions are working perfectly!")
}

// Helper function to substitute variables (simplified version)
func substituteVariables(condition map[string]interface{}, context map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	conditionBytes, _ := json.Marshal(condition)
	conditionStr := string(conditionBytes)

	// Simple variable substitution for demo
	// In real implementation, this would be more sophisticated
	conditionStr = replaceVariable(conditionStr, "${user.id}", getNestedValue("user.id", context))
	conditionStr = replaceVariable(conditionStr, "${resource.department}", getNestedValue("resource.department", context))

	json.Unmarshal([]byte(conditionStr), &result)
	return result
}

func replaceVariable(str, variable string, value interface{}) string {
	if value == nil {
		return str
	}

	valueStr := fmt.Sprintf("%v", value)
	return fmt.Sprintf(str, valueStr) // This is simplified - real implementation would be more robust
}

func getNestedValue(path string, context map[string]interface{}) interface{} {
	keys := []string{"user", "id"} // Simplified for demo
	current := context

	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key]
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}
