package operators

import (
	"testing"
)

func TestEqualOperator(t *testing.T) {
	op := &EqualOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{"engineering", "engineering", true},
		{"engineering", "finance", false},
		{123, 123, true},
		{123, 456, false},
		{true, true, true},
		{true, false, false},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a", "b"}, []string{"b", "a"}, false},
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("EqualOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestInOperator(t *testing.T) {
	op := &InOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{"engineering", []string{"engineering", "finance"}, true},
		{"marketing", []string{"engineering", "finance"}, false},
		{123, []int{123, 456, 789}, true},
		{999, []int{123, 456, 789}, false},
		{"test", "not_an_array", false},
		{"test", nil, false},
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("InOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestContainsOperator(t *testing.T) {
	op := &ContainsOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{[]string{"senior_developer", "code_reviewer"}, "senior_developer", true},
		{[]string{"junior_developer"}, "senior_developer", false},
		{[]int{1, 2, 3}, 2, true},
		{[]int{1, 2, 3}, 4, false},
		{"not_an_array", "test", false},
		{nil, "test", false},
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("ContainsOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestRegexOperator(t *testing.T) {
	op := &RegexOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{"10.0.1.100", "^10\\.", true},
		{"192.168.1.100", "^10\\.", false},
		{"john.doe@company.com", ".*@company\\.com$", true},
		{"john.doe@external.com", ".*@company\\.com$", false},
		{"test123", "\\d+", true},
		{"testABC", "\\d+", false},
		{123, "\\d+", true}, // Should convert to string
		{"", "^$", true},    // Empty string matches empty pattern
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("RegexOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestGreaterThanOperator(t *testing.T) {
	op := &GreaterThanOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{5, 3, true},
		{3, 5, false},
		{5, 5, false},
		{5.5, 3.2, true},
		{3.2, 5.5, false},
		{"5", "3", true}, // Should convert strings to numbers
		{"3", "5", false},
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("GreaterThanOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestBetweenOperator(t *testing.T) {
	op := &BetweenOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{5, []int{3, 7}, true},
		{2, []int{3, 7}, false},
		{8, []int{3, 7}, false},
		{5.5, []float64{3.0, 7.0}, true},
		{"10:30", []string{"08:00", "18:00"}, true},  // Time between
		{"07:30", []string{"08:00", "18:00"}, false}, // Time before
		{"19:30", []string{"08:00", "18:00"}, false}, // Time after
		{5, []int{3}, false},                         // Invalid range (not 2 elements)
		{5, "not_an_array", false},                   // Invalid expected value
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("BetweenOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestExistsOperator(t *testing.T) {
	op := &ExistsOperator{}

	testCases := []struct {
		actual   interface{}
		expected interface{}
		result   bool
	}{
		{"test", nil, true},
		{123, nil, true},
		{false, nil, true}, // false is still a value
		{nil, nil, false},
	}

	for _, tc := range testCases {
		result := op.Evaluate(tc.actual, tc.expected)
		if result != tc.result {
			t.Errorf("ExistsOperator.Evaluate(%v, %v) = %v, expected %v",
				tc.actual, tc.expected, result, tc.result)
		}
	}
}

func TestOperatorRegistry(t *testing.T) {
	registry := NewOperatorRegistry()

	// Test that all expected operators are registered
	expectedOperators := []string{
		"eq", "neq", "in", "nin", "contains", "regex",
		"gt", "gte", "lt", "lte", "between", "exists",
	}

	for _, opName := range expectedOperators {
		op, err := registry.Get(opName)
		if err != nil {
			t.Errorf("Expected operator %s to be registered, got error: %v", opName, err)
		}
		if op == nil {
			t.Errorf("Expected operator %s to be non-nil", opName)
		}
	}

	// Test getting non-existent operator
	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent operator")
	}

	// Test registering custom operator
	customOp := &EqualOperator{}
	registry.Register("custom", customOp)

	retrievedOp, err := registry.Get("custom")
	if err != nil {
		t.Errorf("Failed to retrieve custom operator: %v", err)
	}
	if retrievedOp != customOp {
		t.Error("Retrieved operator should be the same instance as registered")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test toSlice
	testSlice := toSlice([]string{"a", "b", "c"})
	if len(testSlice) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(testSlice))
	}

	nonSlice := toSlice("not a slice")
	if nonSlice != nil {
		t.Error("Expected nil for non-slice input")
	}

	// Test toString
	if toString("test") != "test" {
		t.Error("String conversion failed for string input")
	}

	if toString(123) != "123" {
		t.Error("String conversion failed for int input")
	}

	if toString(nil) != "" {
		t.Error("String conversion failed for nil input")
	}

	// Test toFloat64
	if toFloat64(123) != 123.0 {
		t.Error("Float conversion failed for int input")
	}

	if toFloat64("123.45") != 123.45 {
		t.Error("Float conversion failed for string input")
	}

	if toFloat64("invalid") != 0 {
		t.Error("Float conversion should return 0 for invalid input")
	}

	if toFloat64(nil) != 0 {
		t.Error("Float conversion should return 0 for nil input")
	}
}

func TestTimeBetween(t *testing.T) {
	testCases := []struct {
		timeStr  string
		startStr string
		endStr   string
		result   bool
	}{
		{"10:30", "08:00", "18:00", true},
		{"07:30", "08:00", "18:00", false},
		{"19:30", "08:00", "18:00", false},
		{"08:00", "08:00", "18:00", true},    // Boundary case
		{"18:00", "08:00", "18:00", true},    // Boundary case
		{"invalid", "08:00", "18:00", false}, // Invalid time format
		{"10:30", "invalid", "18:00", false}, // Invalid start format
		{"10:30", "08:00", "invalid", false}, // Invalid end format
	}

	for _, tc := range testCases {
		result := isTimeBetween(tc.timeStr, tc.startStr, tc.endStr)
		if result != tc.result {
			t.Errorf("isTimeBetween(%s, %s, %s) = %v, expected %v",
				tc.timeStr, tc.startStr, tc.endStr, result, tc.result)
		}
	}
}
