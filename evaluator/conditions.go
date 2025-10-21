package evaluator

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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

// evaluateOperator evaluates a specific operator's conditions
func (ce *ConditionEvaluator) evaluateOperator(operator string, operatorConditions interface{}, context map[string]interface{}) bool {
	conditionsMap, ok := operatorConditions.(map[string]interface{})
	if !ok {
		return false
	}

	switch operator {
	case "StringEquals":
		return ce.evaluateStringEquals(conditionsMap, context)
	case "StringNotEquals":
		return ce.evaluateStringNotEquals(conditionsMap, context)
	case "StringLike":
		return ce.evaluateStringLike(conditionsMap, context)
	case "NumericLessThan":
		return ce.evaluateNumericLessThan(conditionsMap, context)
	case "NumericLessThanEquals":
		return ce.evaluateNumericLessThanEquals(conditionsMap, context)
	case "NumericGreaterThan":
		return ce.evaluateNumericGreaterThan(conditionsMap, context)
	case "NumericGreaterThanEquals":
		return ce.evaluateNumericGreaterThanEquals(conditionsMap, context)
	case "Bool":
		return ce.evaluateBool(conditionsMap, context)
	case "IpAddress":
		return ce.evaluateIpAddress(conditionsMap, context)
	case "DateGreaterThan":
		return ce.evaluateDateGreaterThan(conditionsMap, context)
	case "DateLessThan":
		return ce.evaluateDateLessThan(conditionsMap, context)
	default:
		return false // Unknown operator
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
