package conditions

import (
	"strings"

	"abac_go_example/constants"
	"abac_go_example/evaluator/path"
)

// ArrayConditionEvaluator handles all array-based condition evaluations
type ArrayConditionEvaluator struct {
	*BaseEvaluator
}

// NewArrayEvaluator creates a new array evaluator
func NewArrayEvaluator(pathResolver path.PathResolver) *ArrayConditionEvaluator {
	return &ArrayConditionEvaluator{
		BaseEvaluator: NewBaseEvaluator(pathResolver),
	}
}

// Evaluate delegates to the appropriate array evaluation method
func (ae *ArrayConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	// This is a generic method - specific operations should use dedicated methods
	return ae.EvaluateContains(conditions, context)
}

// EvaluateContains checks if array contains expected value
func (ae *ArrayConditionEvaluator) EvaluateContains(conditions interface{}, context map[string]interface{}) bool {
	return ae.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		// Convert to array if needed
		actualArray := ae.convertToArray(evalCtx.ActualValue)
		expectedStr := ae.ToString(evalCtx.ExpectedValue)

		for _, item := range actualArray {
			if ae.ToString(item) == expectedStr {
				return true
			}
		}

		return false
	})
}

// EvaluateNotContains checks if array does not contain expected value
func (ae *ArrayConditionEvaluator) EvaluateNotContains(conditions interface{}, context map[string]interface{}) bool {
	return ae.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		// Convert to array if needed
		actualArray := ae.convertToArray(evalCtx.ActualValue)
		expectedStr := ae.ToString(evalCtx.ExpectedValue)

		for _, item := range actualArray {
			if ae.ToString(item) == expectedStr {
				return false
			}
		}

		return true
	})
}

// EvaluateSize checks if array size matches condition
func (ae *ArrayConditionEvaluator) EvaluateSize(conditions interface{}, context map[string]interface{}) bool {
	return ae.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualArray := ae.convertToArray(evalCtx.ActualValue)
		actualSize := len(actualArray)

		// sizeCondition can be number or map with operators
		if sizeMap, ok := evalCtx.ExpectedValue.(map[string]interface{}); ok {
			return ae.evaluateSizeWithOperators(actualSize, sizeMap)
		}

		// Simple equality check
		expectedSize := int(ae.ToFloat64(evalCtx.ExpectedValue))
		return actualSize == expectedSize
	})
}

// convertToArray converts value to array format
func (ae *ArrayConditionEvaluator) convertToArray(value interface{}) []interface{} {
	if arr, ok := value.([]interface{}); ok {
		return arr
	}
	// Single value treated as array of one
	return []interface{}{value}
}

// evaluateSizeWithOperators evaluates size with comparison operators
func (ae *ArrayConditionEvaluator) evaluateSizeWithOperators(actualSize int, sizeMap map[string]interface{}) bool {
	for op, value := range sizeMap {
		expectedSize := int(ae.ToFloat64(value))

		switch strings.ToLower(op) {
		case constants.SizeOpEquals, constants.SizeOpEqualsLong:
			if actualSize != expectedSize {
				return false
			}
		case constants.SizeOpGreaterThan, constants.SizeOpGreaterThanLong:
			if actualSize <= expectedSize {
				return false
			}
		case constants.SizeOpGreaterThanEquals, constants.SizeOpGreaterThanEqualsLong:
			if actualSize < expectedSize {
				return false
			}
		case constants.SizeOpLessThan, constants.SizeOpLessThanLong:
			if actualSize >= expectedSize {
				return false
			}
		case constants.SizeOpLessThanEquals, constants.SizeOpLessThanEqualsLong:
			if actualSize > expectedSize {
				return false
			}
		default:
			return false
		}
	}

	return true
}
