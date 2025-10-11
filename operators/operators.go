package operators

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Operator interface defines the contract for rule operators
type Operator interface {
	Evaluate(actual, expected interface{}) bool
}

// OperatorRegistry holds all available operators
type OperatorRegistry struct {
	operators map[string]Operator
}

// NewOperatorRegistry creates a new operator registry with default operators
func NewOperatorRegistry() *OperatorRegistry {
	registry := &OperatorRegistry{
		operators: make(map[string]Operator),
	}

	// Register default operators
	registry.Register("eq", &EqualOperator{})
	registry.Register("neq", &NotEqualOperator{})
	registry.Register("in", &InOperator{})
	registry.Register("nin", &NotInOperator{})
	registry.Register("contains", &ContainsOperator{})
	registry.Register("regex", &RegexOperator{})
	registry.Register("gt", &GreaterThanOperator{})
	registry.Register("gte", &GreaterThanEqualOperator{})
	registry.Register("lt", &LessThanOperator{})
	registry.Register("lte", &LessThanEqualOperator{})
	registry.Register("between", &BetweenOperator{})
	registry.Register("exists", &ExistsOperator{})

	return registry
}

// Register adds an operator to the registry
func (r *OperatorRegistry) Register(name string, operator Operator) {
	r.operators[name] = operator
}

// Get retrieves an operator by name
func (r *OperatorRegistry) Get(name string) (Operator, error) {
	operator, exists := r.operators[name]
	if !exists {
		return nil, fmt.Errorf("operator not found: %s", name)
	}
	return operator, nil
}

// EqualOperator implements exact equality comparison
type EqualOperator struct{}

func (o *EqualOperator) Evaluate(actual, expected interface{}) bool {
	return reflect.DeepEqual(actual, expected)
}

// NotEqualOperator implements inequality comparison
type NotEqualOperator struct{}

func (o *NotEqualOperator) Evaluate(actual, expected interface{}) bool {
	return !reflect.DeepEqual(actual, expected)
}

// InOperator checks if actual value is in expected array
type InOperator struct{}

func (o *InOperator) Evaluate(actual, expected interface{}) bool {
	expectedSlice := toSlice(expected)
	if expectedSlice == nil {
		return false
	}

	for _, item := range expectedSlice {
		if reflect.DeepEqual(actual, item) {
			return true
		}
	}
	return false
}

// NotInOperator checks if actual value is not in expected array
type NotInOperator struct{}

func (o *NotInOperator) Evaluate(actual, expected interface{}) bool {
	inOp := &InOperator{}
	return !inOp.Evaluate(actual, expected)
}

// ContainsOperator checks if actual array contains expected value
type ContainsOperator struct{}

func (o *ContainsOperator) Evaluate(actual, expected interface{}) bool {
	actualSlice := toSlice(actual)
	if actualSlice == nil {
		return false
	}

	for _, item := range actualSlice {
		if reflect.DeepEqual(item, expected) {
			return true
		}
	}
	return false
}

// RegexOperator performs regular expression matching
type RegexOperator struct{}

func (o *RegexOperator) Evaluate(actual, expected interface{}) bool {
	actualStr := toString(actual)
	expectedStr := toString(expected)

	if expectedStr == "" {
		return false
	}

	matched, err := regexp.MatchString(expectedStr, actualStr)
	if err != nil {
		return false
	}
	return matched
}

// GreaterThanOperator performs > comparison
type GreaterThanOperator struct{}

func (o *GreaterThanOperator) Evaluate(actual, expected interface{}) bool {
	return compareNumbers(actual, expected) > 0
}

// GreaterThanEqualOperator performs >= comparison
type GreaterThanEqualOperator struct{}

func (o *GreaterThanEqualOperator) Evaluate(actual, expected interface{}) bool {
	return compareNumbers(actual, expected) >= 0
}

// LessThanOperator performs < comparison
type LessThanOperator struct{}

func (o *LessThanOperator) Evaluate(actual, expected interface{}) bool {
	return compareNumbers(actual, expected) < 0
}

// LessThanEqualOperator performs <= comparison
type LessThanEqualOperator struct{}

func (o *LessThanEqualOperator) Evaluate(actual, expected interface{}) bool {
	return compareNumbers(actual, expected) <= 0
}

// BetweenOperator checks if value is between two bounds (inclusive)
type BetweenOperator struct{}

func (o *BetweenOperator) Evaluate(actual, expected interface{}) bool {
	expectedSlice := toSlice(expected)
	if expectedSlice == nil || len(expectedSlice) != 2 {
		return false
	}

	// For time-based comparisons (like time_of_day)
	if actualStr := toString(actual); actualStr != "" {
		if strings.Contains(actualStr, ":") {
			return isTimeBetween(actualStr, toString(expectedSlice[0]), toString(expectedSlice[1]))
		}
	}

	// For numeric comparisons
	actualNum := toFloat64(actual)
	lowerNum := toFloat64(expectedSlice[0])
	upperNum := toFloat64(expectedSlice[1])

	return actualNum >= lowerNum && actualNum <= upperNum
}

// ExistsOperator checks if a value exists (is not nil)
type ExistsOperator struct{}

func (o *ExistsOperator) Evaluate(actual, expected interface{}) bool {
	return actual != nil
}

// Helper functions

func toSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil
	}

	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	return result
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toFloat64(value interface{}) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func compareNumbers(actual, expected interface{}) int {
	actualNum := toFloat64(actual)
	expectedNum := toFloat64(expected)

	if actualNum > expectedNum {
		return 1
	} else if actualNum < expectedNum {
		return -1
	}
	return 0
}

func isTimeBetween(timeStr, startStr, endStr string) bool {
	// Parse time strings in HH:MM format
	parseTime := func(s string) (int, int, error) {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid time format")
		}
		hour, err1 := strconv.Atoi(parts[0])
		minute, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return 0, 0, fmt.Errorf("invalid time format")
		}
		return hour, minute, nil
	}

	timeHour, timeMin, err1 := parseTime(timeStr)
	startHour, startMin, err2 := parseTime(startStr)
	endHour, endMin, err3 := parseTime(endStr)

	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}

	timeMinutes := timeHour*60 + timeMin
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	return timeMinutes >= startMinutes && timeMinutes <= endMinutes
}
