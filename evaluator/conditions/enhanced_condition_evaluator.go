package conditions

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"abac_go_example/evaluator/path"
	"abac_go_example/operators"
)

// EnhancedConditionEvaluator provides advanced condition evaluation capabilities
type EnhancedConditionEvaluator struct {
	// Cache for compiled regex patterns
	regexCache map[string]*regexp.Regexp
	// Path resolver for attribute access
	pathResolver path.PathResolver
	// Network utilities for IP and user agent processing
	networkUtils *operators.NetworkUtils
}

// NewEnhancedConditionEvaluator creates a new enhanced condition evaluator
func NewEnhancedConditionEvaluator() *EnhancedConditionEvaluator {
	return &EnhancedConditionEvaluator{
		regexCache:   make(map[string]*regexp.Regexp),
		pathResolver: path.NewCompositePathResolver(),
		networkUtils: operators.NewNetworkUtils(),
	}
}

// EvaluateConditions evaluates conditions with enhanced operators and complex expressions
func (ece *EnhancedConditionEvaluator) EvaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	// Evaluate each condition operator
	for operator, operatorConditions := range conditions {
		if !ece.evaluateOperator(operator, operatorConditions, context) {
			return false
		}
	}

	return true
}

// evaluateOperator evaluates a specific condition operator
func (ece *EnhancedConditionEvaluator) evaluateOperator(operator string, operatorConditions interface{}, context map[string]interface{}) bool {
	switch strings.ToLower(operator) {
	// String operators
	case "stringequals":
		return ece.evaluateStringEquals(operatorConditions, context)
	case "stringnotequals":
		return ece.evaluateStringNotEquals(operatorConditions, context)
	case "stringlike":
		return ece.evaluateStringLike(operatorConditions, context)
	case "stringcontains":
		return ece.evaluateStringContains(operatorConditions, context)
	case "stringstartswith":
		return ece.evaluateStringStartsWith(operatorConditions, context)
	case "stringendswith":
		return ece.evaluateStringEndsWith(operatorConditions, context)
	case "stringregex":
		return ece.evaluateStringRegex(operatorConditions, context)

	// Numeric operators
	case "numericequals":
		return ece.evaluateNumericEquals(operatorConditions, context)
	case "numericnotequals":
		return ece.evaluateNumericNotEquals(operatorConditions, context)
	case "numericlessthan":
		return ece.evaluateNumericLessThan(operatorConditions, context)
	case "numericlessthanequals":
		return ece.evaluateNumericLessThanEquals(operatorConditions, context)
	case "numericgreaterthan":
		return ece.evaluateNumericGreaterThan(operatorConditions, context)
	case "numericgreaterthanequals":
		return ece.evaluateNumericGreaterThanEquals(operatorConditions, context)
	case "numericbetween":
		return ece.evaluateNumericBetween(operatorConditions, context)

	// Date/Time operators (enhanced)
	case "datelessthan", "timelessthan":
		return ece.evaluateTimeLessThan(operatorConditions, context)
	case "datelessthanequals", "timelessthanequals":
		return ece.evaluateTimeLessThanEquals(operatorConditions, context)
	case "dategreaterthan", "timegreaterthan":
		return ece.evaluateTimeGreaterThan(operatorConditions, context)
	case "dategreaterthanequals", "timegreaterthanequals":
		return ece.evaluateTimeGreaterThanEquals(operatorConditions, context)
	case "datebetween", "timebetween":
		return ece.evaluateTimeBetween(operatorConditions, context)
	case "dayofweek":
		return ece.evaluateDayOfWeek(operatorConditions, context)
	case "timeofday":
		return ece.evaluateTimeOfDay(operatorConditions, context)
	case "isbusinesshours":
		return ece.evaluateIsBusinessHours(operatorConditions, context)

	// Array operators
	case "arraycontains":
		return ece.evaluateArrayContains(operatorConditions, context)
	case "arraynotcontains":
		return ece.evaluateArrayNotContains(operatorConditions, context)
	case "arraysize":
		return ece.evaluateArraySize(operatorConditions, context)

	// Network operators (enhanced)
	case "ipinrange":
		return ece.evaluateIPInRange(operatorConditions, context)
	case "ipnotinrange":
		return ece.evaluateIPNotInRange(operatorConditions, context)
	case "isinternalip":
		return ece.evaluateIsInternalIP(operatorConditions, context)

	// Boolean operators
	case "bool", "boolean":
		return ece.evaluateBoolean(operatorConditions, context)

	// Complex operators
	case "and":
		return ece.evaluateAnd(operatorConditions, context)
	case "or":
		return ece.evaluateOr(operatorConditions, context)
	case "not":
		return ece.evaluateNot(operatorConditions, context)

	default:
		// Fallback to legacy evaluation for unknown operators
		return true
	}
}

// String evaluation methods

func (ece *EnhancedConditionEvaluator) evaluateStringEquals(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedValue := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualStr := ece.toString(actualValue)
		expectedStr := ece.toString(expectedValue)

		if actualStr != expectedStr {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateStringNotEquals(conditions interface{}, context map[string]interface{}) bool {
	return !ece.evaluateStringEquals(conditions, context)
}

func (ece *EnhancedConditionEvaluator) evaluateStringLike(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, pattern := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualStr := ece.toString(actualValue)
		patternStr := ece.toString(pattern)

		// Convert SQL LIKE pattern to regex
		regexPattern := strings.ReplaceAll(patternStr, "%", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "_", ".")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, actualStr)
		if err != nil || !matched {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateStringContains(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, substring := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualStr := ece.toString(actualValue)
		substringStr := ece.toString(substring)

		if !strings.Contains(actualStr, substringStr) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateStringStartsWith(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, prefix := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualStr := ece.toString(actualValue)
		prefixStr := ece.toString(prefix)

		if !strings.HasPrefix(actualStr, prefixStr) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateStringEndsWith(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, suffix := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualStr := ece.toString(actualValue)
		suffixStr := ece.toString(suffix)

		if !strings.HasSuffix(actualStr, suffixStr) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateStringRegex(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, pattern := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualStr := ece.toString(actualValue)
		patternStr := ece.toString(pattern)

		// Use cached regex if available
		regex, exists := ece.regexCache[patternStr]
		if !exists {
			var err error
			regex, err = regexp.Compile(patternStr)
			if err != nil {
				return false
			}
			ece.regexCache[patternStr] = regex
		}

		if !regex.MatchString(actualStr) {
			return false
		}
	}

	return true
}

// Numeric evaluation methods

func (ece *EnhancedConditionEvaluator) evaluateNumericEquals(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedValue := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualNum := ece.toFloat64(actualValue)
		expectedNum := ece.toFloat64(expectedValue)

		if actualNum != expectedNum {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateNumericNotEquals(conditions interface{}, context map[string]interface{}) bool {
	return !ece.evaluateNumericEquals(conditions, context)
}

func (ece *EnhancedConditionEvaluator) evaluateNumericLessThan(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualNum := ece.toFloat64(actualValue)
		thresholdNum := ece.toFloat64(threshold)

		if actualNum >= thresholdNum {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateNumericLessThanEquals(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualNum := ece.toFloat64(actualValue)
		thresholdNum := ece.toFloat64(threshold)

		if actualNum > thresholdNum {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateNumericGreaterThan(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualNum := ece.toFloat64(actualValue)
		thresholdNum := ece.toFloat64(threshold)

		if actualNum <= thresholdNum {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateNumericGreaterThanEquals(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualNum := ece.toFloat64(actualValue)
		thresholdNum := ece.toFloat64(threshold)

		if actualNum < thresholdNum {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateNumericBetween(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, rangeValue := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualNum := ece.toFloat64(actualValue)

		// Range can be array [min, max] or map {"min": x, "max": y}
		if rangeArray, ok := rangeValue.([]interface{}); ok && len(rangeArray) == 2 {
			min := ece.toFloat64(rangeArray[0])
			max := ece.toFloat64(rangeArray[1])
			if actualNum < min || actualNum > max {
				return false
			}
		} else if rangeMap, ok := rangeValue.(map[string]interface{}); ok {
			min := ece.toFloat64(rangeMap["min"])
			max := ece.toFloat64(rangeMap["max"])
			if actualNum < min || actualNum > max {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// Enhanced Time evaluation methods

func (ece *EnhancedConditionEvaluator) evaluateTimeLessThan(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualTime := ece.parseTime(actualValue)
		thresholdTime := ece.parseTime(threshold)

		if actualTime.After(thresholdTime) || actualTime.Equal(thresholdTime) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateTimeLessThanEquals(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualTime := ece.parseTime(actualValue)
		thresholdTime := ece.parseTime(threshold)

		if actualTime.After(thresholdTime) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateTimeGreaterThan(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualTime := ece.parseTime(actualValue)
		thresholdTime := ece.parseTime(threshold)

		if actualTime.Before(thresholdTime) || actualTime.Equal(thresholdTime) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateTimeGreaterThanEquals(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, threshold := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualTime := ece.parseTime(actualValue)
		thresholdTime := ece.parseTime(threshold)

		if actualTime.Before(thresholdTime) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateTimeBetween(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, rangeValue := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualTime := ece.parseTime(actualValue)

		if rangeArray, ok := rangeValue.([]interface{}); ok && len(rangeArray) == 2 {
			startTime := ece.parseTime(rangeArray[0])
			endTime := ece.parseTime(rangeArray[1])
			if actualTime.Before(startTime) || actualTime.After(endTime) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateDayOfWeek(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedDays := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualDay := ece.toString(actualValue)

		// expectedDays can be string or array of strings
		if dayArray, ok := expectedDays.([]interface{}); ok {
			found := false
			for _, day := range dayArray {
				if strings.EqualFold(actualDay, ece.toString(day)) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		} else {
			expectedDay := ece.toString(expectedDays)
			if !strings.EqualFold(actualDay, expectedDay) {
				return false
			}
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateTimeOfDay(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedTime := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualTimeStr := ece.toString(actualValue)
		expectedTimeStr := ece.toString(expectedTime)

		// Parse time in HH:MM format
		actualTime, err1 := time.Parse("15:04", actualTimeStr)
		expectedTime, err2 := time.Parse("15:04", expectedTimeStr)

		if err1 != nil || err2 != nil {
			return false
		}

		if !actualTime.Equal(expectedTime) {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateIsBusinessHours(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expected := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		expectedBool := ece.toBool(expected)

		// Check if current time is business hours using constants
		var isBusinessHours bool
		if timeValue, ok := actualValue.(time.Time); ok {
			hour := timeValue.Hour()
			weekday := int(timeValue.Weekday())
			isBusinessHours = ece.networkUtils.IsBusinessHours(hour, weekday)
		} else if boolValue, ok := actualValue.(bool); ok {
			// If the value is already a boolean, use it directly
			isBusinessHours = boolValue
		} else {
			// Try to calculate from context
			hourValue := ece.getValueFromContext("environment.hour", context)
			dayValue := ece.getValueFromContext("environment.day_of_week", context)

			hour := int(ece.toFloat64(hourValue))
			dayStr := ece.toString(dayValue)

			// Convert day string to weekday number for consistency
			var weekday int
			switch strings.ToLower(dayStr) {
			case "sunday":
				weekday = 0
			case "monday":
				weekday = 1
			case "tuesday":
				weekday = 2
			case "wednesday":
				weekday = 3
			case "thursday":
				weekday = 4
			case "friday":
				weekday = 5
			case "saturday":
				weekday = 6
			default:
				weekday = -1 // Invalid day
			}
			isBusinessHours = ece.networkUtils.IsBusinessHours(hour, weekday)
		}

		if isBusinessHours != expectedBool {
			return false
		}
	}

	return true
}

// Network evaluation methods

func (ece *EnhancedConditionEvaluator) evaluateIPInRange(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, ranges := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		ipStr := ece.toString(actualValue)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return false
		}

		// ranges can be string or array of strings
		var rangeList []string
		if rangeArray, ok := ranges.([]interface{}); ok {
			for _, r := range rangeArray {
				rangeList = append(rangeList, ece.toString(r))
			}
		} else {
			rangeList = []string{ece.toString(ranges)}
		}

		inRange := false
		for _, rangeStr := range rangeList {
			_, cidr, err := net.ParseCIDR(rangeStr)
			if err != nil {
				continue
			}
			if cidr.Contains(ip) {
				inRange = true
				break
			}
		}

		if !inRange {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateIPNotInRange(conditions interface{}, context map[string]interface{}) bool {
	return !ece.evaluateIPInRange(conditions, context)
}

func (ece *EnhancedConditionEvaluator) evaluateIsInternalIP(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expected := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		expectedBool := ece.toBool(expected)

		var isInternal bool
		if boolValue, ok := actualValue.(bool); ok {
			// If the value is already a boolean, use it directly
			isInternal = boolValue
		} else {
			// Try to parse as IP and check if internal
			ipStr := ece.toString(actualValue)
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return false
			}
			isInternal = ece.networkUtils.IsInternalIPAddress(ip)
		}

		if isInternal != expectedBool {
			return false
		}
	}

	return true
}

// Array evaluation methods

func (ece *EnhancedConditionEvaluator) evaluateArrayContains(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedValue := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)

		// Convert to array if needed
		var actualArray []interface{}
		if arr, ok := actualValue.([]interface{}); ok {
			actualArray = arr
		} else {
			// Single value treated as array of one
			actualArray = []interface{}{actualValue}
		}

		expectedStr := ece.toString(expectedValue)
		found := false
		for _, item := range actualArray {
			if ece.toString(item) == expectedStr {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateArrayNotContains(conditions interface{}, context map[string]interface{}) bool {
	return !ece.evaluateArrayContains(conditions, context)
}

func (ece *EnhancedConditionEvaluator) evaluateArraySize(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, sizeCondition := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)

		var actualSize int
		if arr, ok := actualValue.([]interface{}); ok {
			actualSize = len(arr)
		} else {
			actualSize = 1 // Single value
		}

		// sizeCondition can be number or map with operators
		if sizeMap, ok := sizeCondition.(map[string]interface{}); ok {
			for op, value := range sizeMap {
				expectedSize := int(ece.toFloat64(value))
				switch strings.ToLower(op) {
				case "eq", "equals":
					if actualSize != expectedSize {
						return false
					}
				case "gt", "greaterthan":
					if actualSize <= expectedSize {
						return false
					}
				case "gte", "greaterthanequals":
					if actualSize < expectedSize {
						return false
					}
				case "lt", "lessthan":
					if actualSize >= expectedSize {
						return false
					}
				case "lte", "lessthanequals":
					if actualSize > expectedSize {
						return false
					}
				}
			}
		} else {
			expectedSize := int(ece.toFloat64(sizeCondition))
			if actualSize != expectedSize {
				return false
			}
		}
	}

	return true
}

// Boolean evaluation

func (ece *EnhancedConditionEvaluator) evaluateBoolean(conditions interface{}, context map[string]interface{}) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedValue := range condMap {
		actualValue := ece.getValueFromContext(attributePath, context)
		actualBool := ece.toBool(actualValue)
		expectedBool := ece.toBool(expectedValue)

		if actualBool != expectedBool {
			return false
		}
	}

	return true
}

// Complex logical operators

func (ece *EnhancedConditionEvaluator) evaluateAnd(conditions interface{}, context map[string]interface{}) bool {
	condArray, ok := conditions.([]interface{})
	if !ok {
		return false
	}

	for _, condition := range condArray {
		if condMap, ok := condition.(map[string]interface{}); ok {
			if !ece.EvaluateConditions(condMap, context) {
				return false
			}
		}
	}

	return true
}

func (ece *EnhancedConditionEvaluator) evaluateOr(conditions interface{}, context map[string]interface{}) bool {
	condArray, ok := conditions.([]interface{})
	if !ok {
		return false
	}

	for _, condition := range condArray {
		if condMap, ok := condition.(map[string]interface{}); ok {
			if ece.EvaluateConditions(condMap, context) {
				return true
			}
		}
	}

	return false
}

func (ece *EnhancedConditionEvaluator) evaluateNot(conditions interface{}, context map[string]interface{}) bool {
	if condMap, ok := conditions.(map[string]interface{}); ok {
		return !ece.EvaluateConditions(condMap, context)
	}
	return false
}

// Helper methods

func (ece *EnhancedConditionEvaluator) getValueFromContext(attributePath string, context map[string]interface{}) interface{} {
	// Use the composite path resolver to handle all resolution strategies
	value, _ := ece.pathResolver.Resolve(attributePath, context)
	return value
}

func (ece *EnhancedConditionEvaluator) toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func (ece *EnhancedConditionEvaluator) toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func (ece *EnhancedConditionEvaluator) toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true" || v == "1"
	case int:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

func (ece *EnhancedConditionEvaluator) parseTime(value interface{}) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		// Try multiple time formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"15:04", // Time of day
			"2006-01-02",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// isInternalIP method removed - now using networkUtils.IsInternalIPAddress
