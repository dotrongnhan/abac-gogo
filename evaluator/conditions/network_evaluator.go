package conditions

import (
	"net"

	"abac_go_example/evaluator/path"
	"abac_go_example/operators"
)

// NetworkConditionEvaluator handles all network-based condition evaluations
type NetworkConditionEvaluator struct {
	*BaseEvaluator
	networkUtils *operators.NetworkUtils
}

// NewNetworkEvaluator creates a new network evaluator
func NewNetworkEvaluator(pathResolver path.PathResolver, networkUtils *operators.NetworkUtils) *NetworkConditionEvaluator {
	return &NetworkConditionEvaluator{
		BaseEvaluator: NewBaseEvaluator(pathResolver),
		networkUtils:  networkUtils,
	}
}

// Evaluate delegates to the appropriate network evaluation method
func (ne *NetworkConditionEvaluator) Evaluate(conditions interface{}, context map[string]interface{}) bool {
	// This is a generic method - specific operations should use dedicated methods
	return ne.EvaluateIPInRange(conditions, context)
}

// EvaluateIPInRange checks if IP is within specified ranges
func (ne *NetworkConditionEvaluator) EvaluateIPInRange(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		ipStr := ne.ToString(evalCtx.ActualValue)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return false
		}

		rangeList := ne.convertToRangeList(evalCtx.ExpectedValue)
		return ne.isIPInRanges(ip, rangeList)
	})
}

// EvaluateIPNotInRange checks if IP is not within specified ranges
func (ne *NetworkConditionEvaluator) EvaluateIPNotInRange(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		ipStr := ne.ToString(evalCtx.ActualValue)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return false
		}

		rangeList := ne.convertToRangeList(evalCtx.ExpectedValue)
		return !ne.isIPInRanges(ip, rangeList)
	})
}

// EvaluateIsInternalIP checks if IP is internal/private
func (ne *NetworkConditionEvaluator) EvaluateIsInternalIP(conditions interface{}, context map[string]interface{}) bool {
	return ne.EvaluateWithConditionMap(conditions, context, func(evalCtx EvaluationContext) bool {
		expectedBool := ne.ToBool(evalCtx.ExpectedValue)

		var isInternal bool
		if boolValue, ok := evalCtx.ActualValue.(bool); ok {
			// If the value is already a boolean, use it directly
			isInternal = boolValue
		} else {
			// Try to parse as IP and check if internal
			ipStr := ne.ToString(evalCtx.ActualValue)
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return false
			}
			isInternal = ne.networkUtils.IsInternalIPAddress(ip)
		}

		return isInternal == expectedBool
	})
}

// convertToRangeList converts ranges value to string slice
func (ne *NetworkConditionEvaluator) convertToRangeList(ranges interface{}) []string {
	var rangeList []string

	if rangeArray, ok := ranges.([]interface{}); ok {
		for _, r := range rangeArray {
			rangeList = append(rangeList, ne.ToString(r))
		}
	} else {
		rangeList = []string{ne.ToString(ranges)}
	}

	return rangeList
}

// isIPInRanges checks if IP is within any of the provided CIDR ranges
func (ne *NetworkConditionEvaluator) isIPInRanges(ip net.IP, ranges []string) bool {
	for _, rangeStr := range ranges {
		_, cidr, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
