package conditions

import (
	"strings"

	"abac_go_example/constants"
	"abac_go_example/evaluator/path"
	"abac_go_example/operators"
)

// EnhancedConditionEvaluator provides advanced condition evaluation capabilities
type EnhancedConditionEvaluator struct {
	// Specialized evaluators
	stringEvaluator  StringEvaluator
	numericEvaluator NumericEvaluator
	timeEvaluator    TimeEvaluator
	arrayEvaluator   ArrayEvaluator
	networkEvaluator NetworkEvaluator
	logicalEvaluator LogicalEvaluator
}

// NewEnhancedConditionEvaluator creates a new enhanced condition evaluator
func NewEnhancedConditionEvaluator() *EnhancedConditionEvaluator {
	pathResolver := path.NewCompositePathResolver()
	networkUtils := operators.NewNetworkUtils()

	logicalEvaluator := NewLogicalEvaluator(pathResolver)

	ece := &EnhancedConditionEvaluator{
		stringEvaluator:  NewStringEvaluator(pathResolver),
		numericEvaluator: NewNumericEvaluator(pathResolver),
		timeEvaluator:    NewTimeEvaluator(pathResolver, networkUtils),
		arrayEvaluator:   NewArrayEvaluator(pathResolver),
		networkEvaluator: NewNetworkEvaluator(pathResolver, networkUtils),
		logicalEvaluator: logicalEvaluator,
	}

	// Set circular reference for logical evaluator
	logicalEvaluator.SetMainEvaluator(ece)

	return ece
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

// Evaluate implements ConditionEvaluator interface
func (ece *EnhancedConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	if condMap, ok := conditions.(map[string]interface{}); ok {
		return ece.EvaluateConditions(condMap, context)
	}
	return false
}

// evaluateOperator evaluates a specific condition operator using specialized evaluators
func (ece *EnhancedConditionEvaluator) evaluateOperator(operator string, operatorConditions interface{}, context map[string]interface{}) bool {
	switch strings.ToLower(operator) {
	// String operators
	case constants.OpStringEquals:
		return ece.stringEvaluator.EvaluateEquals(operatorConditions, context)
	case constants.OpStringNotEquals:
		return ece.stringEvaluator.EvaluateNotEquals(operatorConditions, context)
	case constants.OpStringLike:
		return ece.stringEvaluator.EvaluateLike(operatorConditions, context)
	case constants.OpStringContains:
		return ece.stringEvaluator.EvaluateContains(operatorConditions, context)
	case constants.OpStringStartsWith:
		return ece.stringEvaluator.EvaluateStartsWith(operatorConditions, context)
	case constants.OpStringEndsWith:
		return ece.stringEvaluator.EvaluateEndsWith(operatorConditions, context)
	case constants.OpStringRegex:
		return ece.stringEvaluator.EvaluateRegex(operatorConditions, context)

	// Numeric operators
	case constants.OpNumericEquals:
		return ece.numericEvaluator.EvaluateEquals(operatorConditions, context)
	case constants.OpNumericNotEquals:
		return ece.numericEvaluator.EvaluateNotEquals(operatorConditions, context)
	case constants.OpNumericLessThan:
		return ece.numericEvaluator.EvaluateLessThan(operatorConditions, context)
	case constants.OpNumericLessThanEquals:
		return ece.numericEvaluator.EvaluateLessThanEquals(operatorConditions, context)
	case constants.OpNumericGreaterThan:
		return ece.numericEvaluator.EvaluateGreaterThan(operatorConditions, context)
	case constants.OpNumericGreaterThanEquals:
		return ece.numericEvaluator.EvaluateGreaterThanEquals(operatorConditions, context)
	case constants.OpNumericBetween:
		return ece.numericEvaluator.EvaluateBetween(operatorConditions, context)

	// Date/Time operators (enhanced)
	case constants.OpDateLessThan, constants.OpTimeLessThan:
		return ece.timeEvaluator.EvaluateLessThan(operatorConditions, context)
	case constants.OpDateLessThanEquals, constants.OpTimeLessThanEquals:
		return ece.timeEvaluator.EvaluateLessThanEquals(operatorConditions, context)
	case constants.OpDateGreaterThan, constants.OpTimeGreaterThan:
		return ece.timeEvaluator.EvaluateGreaterThan(operatorConditions, context)
	case constants.OpDateGreaterThanEquals, constants.OpTimeGreaterThanEquals:
		return ece.timeEvaluator.EvaluateGreaterThanEquals(operatorConditions, context)
	case constants.OpDateBetween, constants.OpTimeBetween:
		return ece.timeEvaluator.EvaluateBetween(operatorConditions, context)
	case constants.OpDayOfWeek:
		return ece.timeEvaluator.EvaluateDayOfWeek(operatorConditions, context)
	case constants.OpTimeOfDay:
		return ece.timeEvaluator.EvaluateTimeOfDay(operatorConditions, context)
	case constants.OpIsBusinessHours:
		return ece.timeEvaluator.EvaluateIsBusinessHours(operatorConditions, context)

	// Array operators
	case constants.OpArrayContains:
		return ece.arrayEvaluator.EvaluateContains(operatorConditions, context)
	case constants.OpArrayNotContains:
		return ece.arrayEvaluator.EvaluateNotContains(operatorConditions, context)
	case constants.OpArraySize:
		return ece.arrayEvaluator.EvaluateSize(operatorConditions, context)

	// Network operators (enhanced)
	case constants.OpIPInRange:
		return ece.networkEvaluator.EvaluateIPInRange(operatorConditions, context)
	case constants.OpIPNotInRange:
		return ece.networkEvaluator.EvaluateIPNotInRange(operatorConditions, context)
	case constants.OpIsInternalIP:
		return ece.networkEvaluator.EvaluateIsInternalIP(operatorConditions, context)

	// Boolean operators
	case constants.OpBool, constants.OpBoolean:
		return ece.evaluateBoolean(operatorConditions, context)

	// Complex operators
	case constants.OpAnd:
		return ece.logicalEvaluator.EvaluateAnd(operatorConditions, context)
	case constants.OpOr:
		return ece.logicalEvaluator.EvaluateOr(operatorConditions, context)
	case constants.OpNot:
		return ece.logicalEvaluator.EvaluateNot(operatorConditions, context)

	default:
		// Fallback to legacy evaluation for unknown operators
		return true
	}
}

// evaluateBoolean handles boolean condition evaluation
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

// Helper methods for backward compatibility

func (ece *EnhancedConditionEvaluator) getValueFromContext(attributePath string, context map[string]interface{}) interface{} {
	// Delegate to string evaluator's base evaluator
	return ece.stringEvaluator.(*StringConditionEvaluator).GetValueFromContext(attributePath, context)
}

func (ece *EnhancedConditionEvaluator) toBool(value interface{}) bool {
	// Delegate to string evaluator's base evaluator
	return ece.stringEvaluator.(*StringConditionEvaluator).ToBool(value)
}
