package conditions

import "time"

// ConditionEvaluator defines the interface for all condition evaluators
type ConditionEvaluator interface {
	Evaluate(conditions interface{}, context map[string]interface{}) bool
}

// StringEvaluator handles string-based condition evaluations
type StringEvaluator interface {
	ConditionEvaluator
	EvaluateEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateNotEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateLike(conditions interface{}, context map[string]interface{}) bool
	EvaluateContains(conditions interface{}, context map[string]interface{}) bool
	EvaluateStartsWith(conditions interface{}, context map[string]interface{}) bool
	EvaluateEndsWith(conditions interface{}, context map[string]interface{}) bool
	EvaluateRegex(conditions interface{}, context map[string]interface{}) bool
}

// NumericEvaluator handles numeric-based condition evaluations
type NumericEvaluator interface {
	ConditionEvaluator
	EvaluateEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateNotEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateLessThan(conditions interface{}, context map[string]interface{}) bool
	EvaluateLessThanEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateGreaterThan(conditions interface{}, context map[string]interface{}) bool
	EvaluateGreaterThanEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateBetween(conditions interface{}, context map[string]interface{}) bool
}

// TimeEvaluator handles time-based condition evaluations
type TimeEvaluator interface {
	ConditionEvaluator
	EvaluateLessThan(conditions interface{}, context map[string]interface{}) bool
	EvaluateLessThanEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateGreaterThan(conditions interface{}, context map[string]interface{}) bool
	EvaluateGreaterThanEquals(conditions interface{}, context map[string]interface{}) bool
	EvaluateBetween(conditions interface{}, context map[string]interface{}) bool
	EvaluateDayOfWeek(conditions interface{}, context map[string]interface{}) bool
	EvaluateTimeOfDay(conditions interface{}, context map[string]interface{}) bool
	EvaluateIsBusinessHours(conditions interface{}, context map[string]interface{}) bool
}

// ArrayEvaluator handles array-based condition evaluations
type ArrayEvaluator interface {
	ConditionEvaluator
	EvaluateContains(conditions interface{}, context map[string]interface{}) bool
	EvaluateNotContains(conditions interface{}, context map[string]interface{}) bool
	EvaluateSize(conditions interface{}, context map[string]interface{}) bool
}

// NetworkEvaluator handles network-based condition evaluations
type NetworkEvaluator interface {
	ConditionEvaluator
	EvaluateIPInRange(conditions interface{}, context map[string]interface{}) bool
	EvaluateIPNotInRange(conditions interface{}, context map[string]interface{}) bool
	EvaluateIsInternalIP(conditions interface{}, context map[string]interface{}) bool
}

// LogicalEvaluator handles logical operations (AND, OR, NOT)
type LogicalEvaluator interface {
	ConditionEvaluator
	EvaluateAnd(conditions interface{}, context map[string]interface{}) bool
	EvaluateOr(conditions interface{}, context map[string]interface{}) bool
	EvaluateNot(conditions interface{}, context map[string]interface{}) bool
}

// ValueConverter provides common type conversion utilities
type ValueConverter interface {
	ToString(value interface{}) string
	ToFloat64(value interface{}) float64
	ToBool(value interface{}) bool
	ParseTime(value interface{}) time.Time
}

// ContextResolver provides attribute path resolution
type ContextResolver interface {
	GetValueFromContext(attributePath string, context map[string]interface{}) interface{}
}
