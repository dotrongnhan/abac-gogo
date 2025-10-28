package conditions

import "abac_go_example/evaluator/path"

// LogicalConditionEvaluator handles logical operations (AND, OR, NOT)
type LogicalConditionEvaluator struct {
	*BaseEvaluator
	mainEvaluator ConditionEvaluator // Reference to main evaluator for recursive evaluation
}

// NewLogicalEvaluator creates a new logical evaluator
func NewLogicalEvaluator(pathResolver path.PathResolver) *LogicalConditionEvaluator {
	return &LogicalConditionEvaluator{
		BaseEvaluator: NewBaseEvaluator(pathResolver),
	}
}

// SetMainEvaluator sets the main evaluator for recursive calls
func (le *LogicalConditionEvaluator) SetMainEvaluator(evaluator ConditionEvaluator) {
	le.mainEvaluator = evaluator
}

// Evaluate delegates to the appropriate logical evaluation method
func (le *LogicalConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	// This is a generic method - specific operations should use dedicated methods
	return le.EvaluateAnd(conditions, context)
}

// EvaluateAnd evaluates AND logic - all conditions must be true
func (le *LogicalConditionEvaluator) EvaluateAnd(conditions interface{}, context map[string]interface{}) bool {
	condArray, ok := conditions.([]interface{})
	if !ok {
		return false
	}

	for _, condition := range condArray {
		if condMap, ok := condition.(map[string]interface{}); ok {
			if !le.evaluateConditionMap(condMap, context) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// EvaluateOr evaluates OR logic - at least one condition must be true
func (le *LogicalConditionEvaluator) EvaluateOr(conditions interface{}, context map[string]interface{}) bool {
	condArray, ok := conditions.([]interface{})
	if !ok {
		return false
	}

	for _, condition := range condArray {
		if condMap, ok := condition.(map[string]interface{}); ok {
			if le.evaluateConditionMap(condMap, context) {
				return true
			}
		}
	}

	return false
}

// EvaluateNot evaluates NOT logic - condition must be false
func (le *LogicalConditionEvaluator) EvaluateNot(conditions interface{}, context map[string]interface{}) bool {
	if condMap, ok := conditions.(map[string]interface{}); ok {
		return !le.evaluateConditionMap(condMap, context)
	}
	return false
}

// evaluateConditionMap evaluates a condition map using the main evaluator
func (le *LogicalConditionEvaluator) evaluateConditionMap(condMap map[string]interface{}, context map[string]interface{}) bool {
	if le.mainEvaluator != nil {
		return le.mainEvaluator.Evaluate(condMap, context)
	}

	// Fallback - should not happen in normal operation
	return false
}
