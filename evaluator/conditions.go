package evaluator

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// ConditionOperatorType represents the type of condition operator
type ConditionOperatorType string

// Enum constants for condition operators
const (
	// String operators
	ConditionStringEquals    ConditionOperatorType = "StringEquals"
	ConditionStringNotEquals ConditionOperatorType = "StringNotEquals"
	ConditionStringLike      ConditionOperatorType = "StringLike"

	// Numeric operators
	ConditionNumericLessThan          ConditionOperatorType = "NumericLessThan"
	ConditionNumericLessThanEquals    ConditionOperatorType = "NumericLessThanEquals"
	ConditionNumericGreaterThan       ConditionOperatorType = "NumericGreaterThan"
	ConditionNumericGreaterThanEquals ConditionOperatorType = "NumericGreaterThanEquals"

	// Boolean operator
	ConditionBool ConditionOperatorType = "Bool"

	// Network operator
	ConditionIpAddress ConditionOperatorType = "IpAddress"

	// Date operators
	ConditionDateGreaterThan ConditionOperatorType = "DateGreaterThan"
	ConditionDateLessThan    ConditionOperatorType = "DateLessThan"

	// Logical operators for complex conditions
	ConditionAnd ConditionOperatorType = "And"
	ConditionOr  ConditionOperatorType = "Or"
	ConditionNot ConditionOperatorType = "Not"
)

// AllConditionOperatorTypes returns all available condition operator types
func AllConditionOperatorTypes() []ConditionOperatorType {
	return []ConditionOperatorType{
		ConditionStringEquals,
		ConditionStringNotEquals,
		ConditionStringLike,
		ConditionNumericLessThan,
		ConditionNumericLessThanEquals,
		ConditionNumericGreaterThan,
		ConditionNumericGreaterThanEquals,
		ConditionBool,
		ConditionIpAddress,
		ConditionDateGreaterThan,
		ConditionDateLessThan,
		ConditionAnd,
		ConditionOr,
		ConditionNot,
	}
}

// String returns the string representation of the condition operator type
func (cot ConditionOperatorType) String() string {
	return string(cot)
}

// IsValid checks if the condition operator type is valid
func (cot ConditionOperatorType) IsValid() bool {
	for _, validOp := range AllConditionOperatorTypes() {
		if cot == validOp {
			return true
		}
	}
	return false
}

// GetOperatorCategory returns the category of the operator (string, numeric, boolean, etc.)
func (cot ConditionOperatorType) GetOperatorCategory() string {
	switch cot {
	case ConditionStringEquals, ConditionStringNotEquals, ConditionStringLike:
		return "string"
	case ConditionNumericLessThan, ConditionNumericLessThanEquals,
		ConditionNumericGreaterThan, ConditionNumericGreaterThanEquals:
		return "numeric"
	case ConditionBool:
		return "boolean"
	case ConditionIpAddress:
		return "network"
	case ConditionDateGreaterThan, ConditionDateLessThan:
		return "date"
	case ConditionAnd, ConditionOr, ConditionNot:
		return "logical"
	default:
		return "unknown"
	}
}

// ComplexCondition represents a complex condition with logical operators
type ComplexCondition struct {
	Type       string                `json:"type"`                 // "simple", "logical"
	Operator   ConditionOperatorType `json:"operator,omitempty"`   // For simple: StringEquals, etc. For logical: And, Or, Not
	Key        string                `json:"key,omitempty"`        // For simple conditions: attribute path
	Value      interface{}           `json:"value,omitempty"`      // For simple conditions: expected value
	Left       *ComplexCondition     `json:"left,omitempty"`       // For logical conditions: left operand
	Right      *ComplexCondition     `json:"right,omitempty"`      // For logical conditions: right operand
	Operand    *ComplexCondition     `json:"operand,omitempty"`    // For NOT operator: single operand
	Conditions []ComplexCondition    `json:"conditions,omitempty"` // For array of conditions (alternative format)
}

// ConditionEvaluator handles condition evaluation
type ConditionEvaluator struct{}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{}
}

// Evaluate evaluates all conditions in a condition block
// All conditions must pass (AND logic)
func (ce *ConditionEvaluator) Evaluate(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true // No conditions means always pass
	}

	for operator, operatorConditions := range conditions {
		if !ce.evaluateOperator(operator, operatorConditions, context) {
			return false
		}
	}
	return true
}

// EvaluateComplex evaluates a complex condition with logical operators
func (ce *ConditionEvaluator) EvaluateComplex(condition *ComplexCondition, context map[string]interface{}) bool {
	if condition == nil {
		return true
	}

	switch condition.Type {
	case "simple":
		return ce.evaluateSimpleCondition(condition, context)
	case "logical":
		return ce.evaluateLogicalCondition(condition, context)
	default:
		// Try to infer type based on operator
		if condition.Operator == ConditionAnd || condition.Operator == ConditionOr || condition.Operator == ConditionNot {
			return ce.evaluateLogicalCondition(condition, context)
		}
		return ce.evaluateSimpleCondition(condition, context)
	}
}

// evaluateOperator evaluates a specific operator's conditions
func (ce *ConditionEvaluator) evaluateOperator(operator string, operatorConditions interface{}, context map[string]interface{}) bool {
	switch ConditionOperatorType(operator) {
	case ConditionAnd:
		return ce.evaluateAndOperator(operatorConditions, context)
	case ConditionOr:
		return ce.evaluateOrOperator(operatorConditions, context)
	case ConditionNot:
		return ce.evaluateNotOperator(operatorConditions, context)
	default:
		// Handle traditional operators
		conditionsMap, ok := operatorConditions.(map[string]interface{})
		if !ok {
			return false
		}
		return ce.evaluateTraditionalOperator(ConditionOperatorType(operator), conditionsMap, context)
	}
}

// evaluateTraditionalOperator evaluates traditional (non-logical) operators
func (ce *ConditionEvaluator) evaluateTraditionalOperator(operator ConditionOperatorType, conditionsMap map[string]interface{}, context map[string]interface{}) bool {
	switch operator {
	case ConditionStringEquals:
		return ce.evaluateStringEquals(conditionsMap, context)
	case ConditionStringNotEquals:
		return ce.evaluateStringNotEquals(conditionsMap, context)
	case ConditionStringLike:
		return ce.evaluateStringLike(conditionsMap, context)
	case ConditionNumericLessThan:
		return ce.evaluateNumericLessThan(conditionsMap, context)
	case ConditionNumericLessThanEquals:
		return ce.evaluateNumericLessThanEquals(conditionsMap, context)
	case ConditionNumericGreaterThan:
		return ce.evaluateNumericGreaterThan(conditionsMap, context)
	case ConditionNumericGreaterThanEquals:
		return ce.evaluateNumericGreaterThanEquals(conditionsMap, context)
	case ConditionBool:
		return ce.evaluateBool(conditionsMap, context)
	case ConditionIpAddress:
		return ce.evaluateIpAddress(conditionsMap, context)
	case ConditionDateGreaterThan:
		return ce.evaluateDateGreaterThan(conditionsMap, context)
	case ConditionDateLessThan:
		return ce.evaluateDateLessThan(conditionsMap, context)
	default:
		return false // Unknown operator
	}
}

// Logical operators

// evaluateAndOperator evaluates AND logical operator
func (ce *ConditionEvaluator) evaluateAndOperator(operatorConditions interface{}, context map[string]interface{}) bool {
	// Handle array of conditions
	if conditionsArray, ok := operatorConditions.([]interface{}); ok {
		for _, condition := range conditionsArray {
			if conditionMap, ok := condition.(map[string]interface{}); ok {
				if !ce.Evaluate(conditionMap, context) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}

	// Handle map of conditions (traditional format)
	if conditionsMap, ok := operatorConditions.(map[string]interface{}); ok {
		return ce.Evaluate(conditionsMap, context)
	}

	return false
}

// evaluateOrOperator evaluates OR logical operator
func (ce *ConditionEvaluator) evaluateOrOperator(operatorConditions interface{}, context map[string]interface{}) bool {
	// Handle array of conditions
	if conditionsArray, ok := operatorConditions.([]interface{}); ok {
		for _, condition := range conditionsArray {
			if conditionMap, ok := condition.(map[string]interface{}); ok {
				if ce.Evaluate(conditionMap, context) {
					return true
				}
			}
		}
		return false
	}

	// Handle map of conditions (at least one must pass)
	if conditionsMap, ok := operatorConditions.(map[string]interface{}); ok {
		for operator, operatorConditions := range conditionsMap {
			if ce.evaluateOperator(operator, operatorConditions, context) {
				return true
			}
		}
		return false
	}

	return false
}

// evaluateNotOperator evaluates NOT logical operator
func (ce *ConditionEvaluator) evaluateNotOperator(operatorConditions interface{}, context map[string]interface{}) bool {
	// Handle single condition map
	if conditionMap, ok := operatorConditions.(map[string]interface{}); ok {
		return !ce.Evaluate(conditionMap, context)
	}

	// Handle array with single condition
	if conditionsArray, ok := operatorConditions.([]interface{}); ok {
		if len(conditionsArray) == 1 {
			if conditionMap, ok := conditionsArray[0].(map[string]interface{}); ok {
				return !ce.Evaluate(conditionMap, context)
			}
		}
	}

	return false
}

// evaluateSimpleCondition evaluates a simple condition from ComplexCondition struct
func (ce *ConditionEvaluator) evaluateSimpleCondition(condition *ComplexCondition, context map[string]interface{}) bool {
	if condition.Key == "" || condition.Operator == "" {
		return false
	}

	// Create a temporary conditions map for the traditional evaluator
	conditionsMap := map[string]interface{}{
		condition.Key: condition.Value,
	}

	return ce.evaluateTraditionalOperator(condition.Operator, conditionsMap, context)
}

// evaluateLogicalCondition evaluates a logical condition from ComplexCondition struct
func (ce *ConditionEvaluator) evaluateLogicalCondition(condition *ComplexCondition, context map[string]interface{}) bool {
	switch condition.Operator {
	case ConditionAnd:
		// Handle array format
		if len(condition.Conditions) > 0 {
			for _, subCondition := range condition.Conditions {
				if !ce.EvaluateComplex(&subCondition, context) {
					return false
				}
			}
			return true
		}
		// Handle left/right format
		if condition.Left != nil && condition.Right != nil {
			return ce.EvaluateComplex(condition.Left, context) && ce.EvaluateComplex(condition.Right, context)
		}
		return false

	case ConditionOr:
		// Handle array format
		if len(condition.Conditions) > 0 {
			for _, subCondition := range condition.Conditions {
				if ce.EvaluateComplex(&subCondition, context) {
					return true
				}
			}
			return false
		}
		// Handle left/right format
		if condition.Left != nil && condition.Right != nil {
			return ce.EvaluateComplex(condition.Left, context) || ce.EvaluateComplex(condition.Right, context)
		}
		return false

	case ConditionNot:
		// Handle single operand
		if condition.Operand != nil {
			return !ce.EvaluateComplex(condition.Operand, context)
		}
		// Handle left operand (alternative format)
		if condition.Left != nil {
			return !ce.EvaluateComplex(condition.Left, context)
		}
		return false

	default:
		return false
	}
}

// String operators
func (ce *ConditionEvaluator) evaluateStringEquals(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualStr := ce.toString(actualValue)
		expectedStr := ce.toString(expectedValue)

		if actualStr != expectedStr {
			return false
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateStringNotEquals(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualStr := ce.toString(actualValue)
		expectedStr := ce.toString(expectedValue)

		if actualStr == expectedStr {
			return false
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateStringLike(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualStr := ce.toString(actualValue)
		pattern := ce.toString(expectedValue)

		// Convert wildcard pattern to simple matching
		if strings.Contains(pattern, "*") {
			if !ce.matchWildcard(pattern, actualStr) {
				return false
			}
		} else {
			if actualStr != pattern {
				return false
			}
		}
	}
	return true
}

// Numeric operators
func (ce *ConditionEvaluator) evaluateNumericLessThan(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualNum := ce.toFloat64(actualValue)
		expectedNum := ce.toFloat64(expectedValue)

		if actualNum >= expectedNum {
			return false
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateNumericLessThanEquals(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualNum := ce.toFloat64(actualValue)
		expectedNum := ce.toFloat64(expectedValue)

		if actualNum > expectedNum {
			return false
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateNumericGreaterThan(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualNum := ce.toFloat64(actualValue)
		expectedNum := ce.toFloat64(expectedValue)

		if actualNum <= expectedNum {
			return false
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateNumericGreaterThanEquals(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualNum := ce.toFloat64(actualValue)
		expectedNum := ce.toFloat64(expectedValue)

		if actualNum < expectedNum {
			return false
		}
	}
	return true
}

// Boolean operator
func (ce *ConditionEvaluator) evaluateBool(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualBool := ce.toBool(actualValue)
		expectedBool := ce.toBool(expectedValue)

		if actualBool != expectedBool {
			return false
		}
	}
	return true
}

// IP Address operator
func (ce *ConditionEvaluator) evaluateIpAddress(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualIP := ce.toString(actualValue)

		// Handle array of CIDR blocks
		if expectedArray, ok := expectedValue.([]interface{}); ok {
			matched := false
			for _, cidr := range expectedArray {
				cidrStr := ce.toString(cidr)
				if ce.ipInCIDR(actualIP, cidrStr) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		} else {
			// Single CIDR block
			cidrStr := ce.toString(expectedValue)
			if !ce.ipInCIDR(actualIP, cidrStr) {
				return false
			}
		}
	}
	return true
}

// Date operators
func (ce *ConditionEvaluator) evaluateDateGreaterThan(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualTime := ce.toTime(actualValue)
		expectedTime := ce.toTime(expectedValue)

		if actualTime.Before(expectedTime) || actualTime.Equal(expectedTime) {
			return false
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateDateLessThan(conditions map[string]interface{}, context map[string]interface{}) bool {
	for contextKey, expectedValue := range conditions {
		actualValue := ce.getContextValue(contextKey, context)
		actualTime := ce.toTime(actualValue)
		expectedTime := ce.toTime(expectedValue)

		if actualTime.After(expectedTime) || actualTime.Equal(expectedTime) {
			return false
		}
	}
	return true
}

// Helper functions
func (ce *ConditionEvaluator) getContextValue(key string, context map[string]interface{}) interface{} {
	// Support nested keys like "user.department"
	keys := strings.Split(key, ".")
	current := context

	for i, k := range keys {
		if i == len(keys)-1 {
			// Last key, return the value
			return current[k]
		}

		// Navigate deeper
		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

func (ce *ConditionEvaluator) toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func (ce *ConditionEvaluator) toFloat64(value interface{}) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func (ce *ConditionEvaluator) toBool(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	case int:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

func (ce *ConditionEvaluator) toTime(value interface{}) time.Time {
	if value == nil {
		return time.Time{}
	}

	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		// Try different time formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"15:04:05", // Time of day
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func (ce *ConditionEvaluator) matchWildcard(pattern, value string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// *middle* - contains
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(value, middle)
	} else if strings.HasPrefix(pattern, "*") {
		// *suffix - ends with
		suffix := pattern[1:]
		return strings.HasSuffix(value, suffix)
	} else if strings.HasSuffix(pattern, "*") {
		// prefix* - starts with
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(value, prefix)
	}

	return pattern == value
}

func (ce *ConditionEvaluator) ipInCIDR(ip, cidr string) bool {
	// Parse IP
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Maybe it's a single IP
		if singleIP := net.ParseIP(cidr); singleIP != nil {
			return ipAddr.Equal(singleIP)
		}
		return false
	}

	return ipNet.Contains(ipAddr)
}

// substituteVariables replaces ${...} variables in conditions
func (ce *ConditionEvaluator) SubstituteVariables(conditions map[string]interface{}, context map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range conditions {
		result[key] = ce.substituteValue(value, context)
	}

	return result
}

func (ce *ConditionEvaluator) substituteValue(value interface{}, context map[string]interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return ce.substituteString(v, context)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = ce.substituteValue(val, context)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = ce.substituteValue(val, context)
		}
		return result
	default:
		return value
	}
}

func (ce *ConditionEvaluator) substituteString(str string, context map[string]interface{}) string {
	if !strings.Contains(str, "${") {
		return str
	}

	result := str
	start := 0

	for {
		start = strings.Index(result[start:], "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}

		varName := result[start+2 : start+end]
		if value := ce.getContextValue(varName, context); value != nil {
			strValue := ce.toString(value)
			result = result[:start] + strValue + result[start+end+1:]
			start += len(strValue)
		} else {
			start += end + 1
		}
	}

	return result
}
