package conditions

import (
	"abac_go_example/constants"
	"abac_go_example/evaluator/path"
)

// NumericConditionEvaluator handles all numeric-based condition evaluations
type NumericConditionEvaluator struct {
	*BaseEvaluator
}

// NewNumericEvaluator creates a new numeric evaluator
func NewNumericEvaluator(pathResolver path.PathResolver) *NumericConditionEvaluator {
	return &NumericConditionEvaluator{
		BaseEvaluator: NewBaseEvaluator(pathResolver),
	}
}

// Evaluate delegates to the appropriate numeric evaluation method
func (ne *NumericConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	// This is a generic method - specific operations should use dedicated methods
	return ne.EvaluateEquals(conditions, context)
}

// EvaluateEquals checks if numeric values are equal
func (ne *NumericConditionEvaluator) EvaluateEquals(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)
		expectedNum := ne.ToFloat64(evalCtx.ExpectedValue)
		return actualNum == expectedNum
	})
}

// EvaluateNotEquals checks if numeric values are not equal
func (ne *NumericConditionEvaluator) EvaluateNotEquals(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)
		expectedNum := ne.ToFloat64(evalCtx.ExpectedValue)
		return actualNum != expectedNum
	})
}

// EvaluateLessThan checks if actual value is less than threshold
func (ne *NumericConditionEvaluator) EvaluateLessThan(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)
		thresholdNum := ne.ToFloat64(evalCtx.ExpectedValue)
		return actualNum < thresholdNum
	})
}

// EvaluateLessThanEquals checks if actual value is less than or equal to threshold
func (ne *NumericConditionEvaluator) EvaluateLessThanEquals(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)
		thresholdNum := ne.ToFloat64(evalCtx.ExpectedValue)
		return actualNum <= thresholdNum
	})
}

// EvaluateGreaterThan checks if actual value is greater than threshold
func (ne *NumericConditionEvaluator) EvaluateGreaterThan(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)
		thresholdNum := ne.ToFloat64(evalCtx.ExpectedValue)
		return actualNum > thresholdNum
	})
}

// EvaluateGreaterThanEquals checks if actual value is greater than or equal to threshold
func (ne *NumericConditionEvaluator) EvaluateGreaterThanEquals(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)
		thresholdNum := ne.ToFloat64(evalCtx.ExpectedValue)
		return actualNum >= thresholdNum
	})
}

// EvaluateBetween checks if value is within a numeric range
func (ne *NumericConditionEvaluator) EvaluateBetween(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualNum := ne.ToFloat64(evalCtx.ActualValue)

		// Range can be array [min, max] or map {constants.RangeKeyMin: x, constants.RangeKeyMax: y}
		if rangeArray, ok := evalCtx.ExpectedValue.([]interface{}); ok && len(rangeArray) == 2 {
			min := ne.ToFloat64(rangeArray[0])
			max := ne.ToFloat64(rangeArray[1])
			return actualNum >= min && actualNum <= max
		}

		if rangeMap, ok := evalCtx.ExpectedValue.(map[string]interface{}); ok {
			min := ne.ToFloat64(rangeMap[constants.RangeKeyMin])
			max := ne.ToFloat64(rangeMap[constants.RangeKeyMax])
			return actualNum >= min && actualNum <= max
		}

		return false
	})
}
