package conditions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"abac_go_example/constants"
	"abac_go_example/evaluator/path"
)

// BaseEvaluator provides common functionality for all evaluators
type BaseEvaluator struct {
	pathResolver path.PathResolver
}

// NewBaseEvaluator creates a new base evaluator
func NewBaseEvaluator(pathResolver path.PathResolver) *BaseEvaluator {
	return &BaseEvaluator{
		pathResolver: pathResolver,
	}
}

// EvaluationContext holds the context for a single evaluation
type EvaluationContext struct {
	AttributePath string
	ExpectedValue interface{}
	ActualValue   interface{}
}

// EvaluateWithConditionMap is a common pattern for evaluating condition maps
func (be *BaseEvaluator) EvaluateWithConditionMap(
	conditions interface{},
	context map[string]interface{},
	evaluator func(EvaluationContext) bool,
) bool {
	condMap, ok := conditions.(map[string]interface{})
	if !ok {
		return false
	}

	for attributePath, expectedValue := range condMap {
		actualValue := be.GetValueFromContext(attributePath, context)
		evalCtx := EvaluationContext{
			AttributePath: attributePath,
			ExpectedValue: expectedValue,
			ActualValue:   actualValue,
		}

		if !evaluator(evalCtx) {
			return false
		}
	}

	return true
}

// GetValueFromContext resolves attribute path from context
func (be *BaseEvaluator) GetValueFromContext(attributePath string, context map[string]interface{}) interface{} {
	value, _ := be.pathResolver.Resolve(attributePath, context)
	return value
}

// ToString converts any value to string
func (be *BaseEvaluator) ToString(value interface{}) string {
	if value == nil {
		return constants.DefaultEmptyString
	}
	return fmt.Sprintf("%v", value)
}

// ToFloat64 converts any value to float64
func (be *BaseEvaluator) ToFloat64(value interface{}) float64 {
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
	return constants.DefaultZeroFloat
}

// ToBool converts any value to bool
func (be *BaseEvaluator) ToBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == constants.BoolStringTrue || v == constants.BoolStringOne
	case int:
		return v != constants.DefaultZeroInt
	case float64:
		return v != constants.DefaultZeroFloat
	}
	return constants.DefaultFalse
}

// ParseTime converts any value to time.Time
func (be *BaseEvaluator) ParseTime(value interface{}) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		// Try multiple time formats
		for _, format := range constants.GetAllTimeFormats() {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
