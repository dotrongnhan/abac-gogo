package conditions

import (
	"regexp"
	"strings"

	"abac_go_example/evaluator/path"
)

// StringConditionEvaluator handles all string-based condition evaluations
type StringConditionEvaluator struct {
	*BaseEvaluator
	regexCache map[string]*regexp.Regexp
}

// NewStringEvaluator creates a new string evaluator
func NewStringEvaluator(pathResolver path.PathResolver) *StringConditionEvaluator {
	return &StringConditionEvaluator{
		BaseEvaluator: NewBaseEvaluator(pathResolver),
		regexCache:    make(map[string]*regexp.Regexp),
	}
}

// Evaluate delegates to the appropriate string evaluation method
func (se *StringConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	// This is a generic method - specific operations should use dedicated methods
	return se.EvaluateEquals(conditions, context)
}

// EvaluateEquals checks if string values are equal
func (se *StringConditionEvaluator) EvaluateEquals(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		expectedStr := se.ToString(evalCtx.ExpectedValue)
		return actualStr == expectedStr
	})
}

// EvaluateNotEquals checks if string values are not equal
func (se *StringConditionEvaluator) EvaluateNotEquals(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		expectedStr := se.ToString(evalCtx.ExpectedValue)
		return actualStr != expectedStr
	})
}

// EvaluateLike checks if string matches SQL LIKE pattern
func (se *StringConditionEvaluator) EvaluateLike(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		patternStr := se.ToString(evalCtx.ExpectedValue)

		// Convert SQL LIKE pattern to regex
		regexPattern := strings.ReplaceAll(patternStr, "%", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "_", ".")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, actualStr)
		return err == nil && matched
	})
}

// EvaluateContains checks if string contains substring
func (se *StringConditionEvaluator) EvaluateContains(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		substringStr := se.ToString(evalCtx.ExpectedValue)
		return strings.Contains(actualStr, substringStr)
	})
}

// EvaluateStartsWith checks if string starts with prefix
func (se *StringConditionEvaluator) EvaluateStartsWith(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		prefixStr := se.ToString(evalCtx.ExpectedValue)
		return strings.HasPrefix(actualStr, prefixStr)
	})
}

// EvaluateEndsWith checks if string ends with suffix
func (se *StringConditionEvaluator) EvaluateEndsWith(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		suffixStr := se.ToString(evalCtx.ExpectedValue)
		return strings.HasSuffix(actualStr, suffixStr)
	})
}

// EvaluateRegex checks if string matches regex pattern
func (se *StringConditionEvaluator) EvaluateRegex(conditions interface{}, context map[string]interface{}) bool {
	return se.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualStr := se.ToString(evalCtx.ActualValue)
		patternStr := se.ToString(evalCtx.ExpectedValue)

		// Use cached regex if available
		regex, exists := se.regexCache[patternStr]
		if !exists {
			var err error
			regex, err = regexp.Compile(patternStr)
			if err != nil {
				return false
			}
			se.regexCache[patternStr] = regex
		}

		return regex.MatchString(actualStr)
	})
}
