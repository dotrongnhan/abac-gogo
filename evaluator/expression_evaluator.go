package evaluator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"abac_go_example/models"
)

// OperatorFunc represents a function that evaluates an operator
type OperatorFunc func(left, right interface{}) bool

// ExpressionEvaluator handles complex boolean expression evaluation
type ExpressionEvaluator struct {
	operators map[string]OperatorFunc
}

// NewExpressionEvaluator creates a new expression evaluator with default operators
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{
		operators: map[string]OperatorFunc{
			"eq":       equals,
			"ne":       notEquals,
			"gt":       greaterThan,
			"gte":      greaterThanEqual,
			"lt":       lessThan,
			"lte":      lessThanEqual,
			"in":       inArray,
			"nin":      notInArray,
			"contains": containsValue,
			"regex":    regexMatch,
			"and":      andOperator,
			"or":       orOperator,
			"not":      notOperator,
			"exists":   existsOperator,
		},
	}
}

// RegisterOperator adds a custom operator
func (ee *ExpressionEvaluator) RegisterOperator(name string, fn OperatorFunc) {
	ee.operators[name] = fn
}

// EvaluateExpression evaluates a boolean expression against attributes
func (ee *ExpressionEvaluator) EvaluateExpression(expr *models.BooleanExpression, attributes map[string]interface{}) bool {
	if expr == nil {
		return true
	}

	switch expr.Type {
	case "simple":
		return ee.evaluateSimpleCondition(expr.Condition, attributes)
	case "compound":
		return ee.evaluateCompoundExpression(expr, attributes)
	default:
		return false
	}
}

// evaluateSimpleCondition evaluates a simple condition
func (ee *ExpressionEvaluator) evaluateSimpleCondition(condition *models.SimpleCondition, attributes map[string]interface{}) bool {
	if condition == nil {
		return true
	}

	// Get the actual value from attributes
	actualValue := ee.getNestedValue(condition.AttributePath, attributes)

	// Get the operator function
	operatorFn, exists := ee.operators[condition.Operator]
	if !exists {
		return false
	}

	// Evaluate the condition
	return operatorFn(actualValue, condition.Value)
}

// evaluateCompoundExpression evaluates a compound expression with logical operators
func (ee *ExpressionEvaluator) evaluateCompoundExpression(expr *models.BooleanExpression, attributes map[string]interface{}) bool {
	switch expr.Operator {
	case "and":
		if expr.Left == nil || expr.Right == nil {
			return false
		}
		return ee.EvaluateExpression(expr.Left, attributes) && ee.EvaluateExpression(expr.Right, attributes)

	case "or":
		if expr.Left == nil || expr.Right == nil {
			return false
		}
		return ee.EvaluateExpression(expr.Left, attributes) || ee.EvaluateExpression(expr.Right, attributes)

	case "not":
		if expr.Left == nil {
			return false
		}
		return !ee.EvaluateExpression(expr.Left, attributes)

	default:
		return false
	}
}

// getNestedValue retrieves a value from nested attributes using dot notation
func (ee *ExpressionEvaluator) getNestedValue(path string, attributes map[string]interface{}) interface{} {
	keys := strings.Split(path, ".")
	current := attributes

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key, return the value
			return current[key]
		}

		// Navigate deeper
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// Operator implementations

func equals(left, right interface{}) bool {
	return reflect.DeepEqual(left, right)
}

func notEquals(left, right interface{}) bool {
	return !reflect.DeepEqual(left, right)
}

func greaterThan(left, right interface{}) bool {
	leftNum := toFloat64(left)
	rightNum := toFloat64(right)
	return leftNum > rightNum
}

func greaterThanEqual(left, right interface{}) bool {
	leftNum := toFloat64(left)
	rightNum := toFloat64(right)
	return leftNum >= rightNum
}

func lessThan(left, right interface{}) bool {
	leftNum := toFloat64(left)
	rightNum := toFloat64(right)
	return leftNum < rightNum
}

func lessThanEqual(left, right interface{}) bool {
	leftNum := toFloat64(left)
	rightNum := toFloat64(right)
	return leftNum <= rightNum
}

func inArray(left, right interface{}) bool {
	rightSlice := toInterfaceSlice(right)
	if rightSlice == nil {
		return false
	}

	for _, item := range rightSlice {
		if reflect.DeepEqual(left, item) {
			return true
		}
	}
	return false
}

func notInArray(left, right interface{}) bool {
	return !inArray(left, right)
}

func containsValue(left, right interface{}) bool {
	leftSlice := toInterfaceSlice(left)
	if leftSlice == nil {
		// If left is not a slice, check if it's a string containing right
		leftStr := toString(left)
		rightStr := toString(right)
		return strings.Contains(leftStr, rightStr)
	}

	for _, item := range leftSlice {
		if reflect.DeepEqual(item, right) {
			return true
		}
	}
	return false
}

func regexMatch(left, right interface{}) bool {
	leftStr := toString(left)
	rightStr := toString(right)

	if rightStr == "" {
		return false
	}

	matched, err := regexp.MatchString(rightStr, leftStr)
	if err != nil {
		return false
	}
	return matched
}

func andOperator(left, right interface{}) bool {
	leftBool := toBool(left)
	rightBool := toBool(right)
	return leftBool && rightBool
}

func orOperator(left, right interface{}) bool {
	leftBool := toBool(left)
	rightBool := toBool(right)
	return leftBool || rightBool
}

func notOperator(left, right interface{}) bool {
	leftBool := toBool(left)
	return !leftBool
}

func existsOperator(left, right interface{}) bool {
	return left != nil
}

// Helper functions

func toFloat64(value interface{}) float64 {
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
	case int32:
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

func toBool(value interface{}) bool {
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
	case int64:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

func toInterfaceSlice(value interface{}) []interface{} {
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

// EvaluateComplexExpression evaluates a complex expression with multiple conditions
// Example: (user.department == "Engineering" AND user.level >= 5) OR
//
//	(user.role == "Admin" AND time.hour >= 9 AND time.hour <= 17)
func (ee *ExpressionEvaluator) EvaluateComplexExpression(conditions []map[string]interface{}, attributes map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	// Evaluate each condition group with OR logic between groups
	for _, conditionGroup := range conditions {
		if ee.evaluateConditionGroup(conditionGroup, attributes) {
			return true
		}
	}

	return false
}

// evaluateConditionGroup evaluates a group of conditions with AND logic
func (ee *ExpressionEvaluator) evaluateConditionGroup(conditionGroup map[string]interface{}, attributes map[string]interface{}) bool {
	for operator, operatorConditions := range conditionGroup {
		if !ee.evaluateOperatorConditions(operator, operatorConditions, attributes) {
			return false
		}
	}
	return true
}

// evaluateOperatorConditions evaluates conditions for a specific operator
func (ee *ExpressionEvaluator) evaluateOperatorConditions(operator string, operatorConditions interface{}, attributes map[string]interface{}) bool {
	conditionsMap, ok := operatorConditions.(map[string]interface{})
	if !ok {
		return false
	}

	operatorFn, exists := ee.operators[operator]
	if !exists {
		return false
	}

	// All conditions for this operator must pass
	for attributePath, expectedValue := range conditionsMap {
		actualValue := ee.getNestedValue(attributePath, attributes)
		if !operatorFn(actualValue, expectedValue) {
			return false
		}
	}

	return true
}
