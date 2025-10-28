package conditions

import (
	"strings"
	"time"

	"abac_go_example/constants"
	"abac_go_example/evaluator/path"
	"abac_go_example/operators"
)

// TimeConditionEvaluator handles all time-based condition evaluations
type TimeConditionEvaluator struct {
	*BaseEvaluator
	networkUtils *operators.NetworkUtils
}

// NewTimeEvaluator creates a new time evaluator
func NewTimeEvaluator(pathResolver path.PathResolver, networkUtils *operators.NetworkUtils) *TimeConditionEvaluator {
	return &TimeConditionEvaluator{
		BaseEvaluator: NewBaseEvaluator(pathResolver),
		networkUtils:  networkUtils,
	}
}

// Evaluate delegates to the appropriate time evaluation method
func (te *TimeConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	// This is a generic method - specific operations should use dedicated methods
	return te.EvaluateLessThan(conditions, context)
}

// EvaluateLessThan checks if time is before threshold
func (te *TimeConditionEvaluator) EvaluateLessThan(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualTime := te.ParseTime(evalCtx.ActualValue)
		thresholdTime := te.ParseTime(evalCtx.ExpectedValue)
		return actualTime.Before(thresholdTime)
	})
}

// EvaluateLessThanEquals checks if time is before or equal to threshold
func (te *TimeConditionEvaluator) EvaluateLessThanEquals(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualTime := te.ParseTime(evalCtx.ActualValue)
		thresholdTime := te.ParseTime(evalCtx.ExpectedValue)
		return actualTime.Before(thresholdTime) || actualTime.Equal(thresholdTime)
	})
}

// EvaluateGreaterThan checks if time is after threshold
func (te *TimeConditionEvaluator) EvaluateGreaterThan(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualTime := te.ParseTime(evalCtx.ActualValue)
		thresholdTime := te.ParseTime(evalCtx.ExpectedValue)
		return actualTime.After(thresholdTime)
	})
}

// EvaluateGreaterThanEquals checks if time is after or equal to threshold
func (te *TimeConditionEvaluator) EvaluateGreaterThanEquals(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualTime := te.ParseTime(evalCtx.ActualValue)
		thresholdTime := te.ParseTime(evalCtx.ExpectedValue)
		return actualTime.After(thresholdTime) || actualTime.Equal(thresholdTime)
	})
}

// EvaluateBetween checks if time is within a time range
func (te *TimeConditionEvaluator) EvaluateBetween(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualTime := te.ParseTime(evalCtx.ActualValue)

		if rangeArray, ok := evalCtx.ExpectedValue.([]interface{}); ok && len(rangeArray) == 2 {
			startTime := te.ParseTime(rangeArray[0])
			endTime := te.ParseTime(rangeArray[1])
			return (actualTime.After(startTime) || actualTime.Equal(startTime)) &&
				(actualTime.Before(endTime) || actualTime.Equal(endTime))
		}

		return false
	})
}

// EvaluateDayOfWeek checks if current day matches expected day(s)
func (te *TimeConditionEvaluator) EvaluateDayOfWeek(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualDay := te.ToString(evalCtx.ActualValue)

		// expectedDays can be string or array of strings
		if dayArray, ok := evalCtx.ExpectedValue.([]interface{}); ok {
			for _, day := range dayArray {
				if strings.EqualFold(actualDay, te.ToString(day)) {
					return true
				}
			}
			return false
		}

		expectedDay := te.ToString(evalCtx.ExpectedValue)
		return strings.EqualFold(actualDay, expectedDay)
	})
}

// EvaluateTimeOfDay checks if current time matches expected time
func (te *TimeConditionEvaluator) EvaluateTimeOfDay(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		actualTimeStr := te.ToString(evalCtx.ActualValue)
		expectedTimeStr := te.ToString(evalCtx.ExpectedValue)

		// Parse time in HH:MM format
		actualTime, err1 := time.Parse(constants.TimeFormatHourMinute, actualTimeStr)
		expectedTime, err2 := time.Parse(constants.TimeFormatHourMinute, expectedTimeStr)

		if err1 != nil || err2 != nil {
			return false
		}

		return actualTime.Equal(expectedTime)
	})
}

// EvaluateIsBusinessHours checks if current time is within business hours
func (te *TimeConditionEvaluator) EvaluateIsBusinessHours(conditions interface{}, context map[string]interface{}) bool {
	return te.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		expectedBool := te.ToBool(evalCtx.ExpectedValue)

		// Check if current time is business hours
		var isBusinessHours bool
		if timeValue, ok := evalCtx.ActualValue.(time.Time); ok {
			hour := timeValue.Hour()
			weekday := int(timeValue.Weekday())
			isBusinessHours = te.networkUtils.IsBusinessHours(hour, weekday)
		} else if boolValue, ok := evalCtx.ActualValue.(bool); ok {
			// If the value is already a boolean, use it directly
			isBusinessHours = boolValue
		} else {
			// Try to calculate from context
			hourValue := te.GetValueFromContext(constants.ContextKeyEnvironmentHour, context)
			dayValue := te.GetValueFromContext(constants.ContextKeyEnvironmentDayOfWeek, context)

			hour := int(te.ToFloat64(hourValue))
			dayStr := te.ToString(dayValue)

			// Convert day string to weekday number for consistency
			weekday := constants.GetDayOfWeekNumber(strings.ToLower(dayStr))
			isBusinessHours = te.networkUtils.IsBusinessHours(hour, weekday)
		}

		return isBusinessHours == expectedBool
	})
}
