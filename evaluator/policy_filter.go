package evaluator

import (
	"fmt"
	"strings"

	"abac_go_example/models"
)

// PolicyFilter provides policy filtering capabilities for performance optimization
type PolicyFilter struct {
	// Cache for pattern matching results
	patternCache map[string]bool
}

// NewPolicyFilter creates a new policy filter
func NewPolicyFilter() *PolicyFilter {
	return &PolicyFilter{
		patternCache: make(map[string]bool),
	}
}

// FilterApplicablePolicies performs pre-filtering to reduce the number of policies to evaluate
// This implements improvements #8 and #12: policy filtering and pre-filtering
func (pf *PolicyFilter) FilterApplicablePolicies(allPolicies []*models.Policy, request *models.EvaluationRequest) []*models.Policy {
	var applicablePolicies []*models.Policy

	for _, policy := range allPolicies {
		// Skip disabled policies
		if !policy.Enabled {
			continue
		}

		// Quick pre-filtering based on action and resource patterns
		if pf.isPolicyPotentiallyApplicable(policy, request) {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	return applicablePolicies
}

// isPolicyPotentiallyApplicable performs fast pre-filtering checks
func (pf *PolicyFilter) isPolicyPotentiallyApplicable(policy *models.Policy, request *models.EvaluationRequest) bool {
	// Check if any statement in the policy might apply
	for _, statement := range policy.Statement {
		if pf.isStatementPotentiallyApplicable(statement, request) {
			return true
		}
	}
	return false
}

// isStatementPotentiallyApplicable performs fast checks without expensive condition evaluation
func (pf *PolicyFilter) isStatementPotentiallyApplicable(statement models.PolicyStatement, request *models.EvaluationRequest) bool {
	// Quick action matching
	if !pf.quickActionMatch(statement.Action, request.Action) {
		return false
	}

	// Quick resource matching
	if !pf.quickResourceMatch(statement.Resource, request.ResourceID) {
		return false
	}

	// Quick NotResource exclusion check
	if pf.isExcludedByNotResource(statement.NotResource, request.ResourceID) {
		return false
	}

	return true
}

// quickActionMatch performs fast action pattern matching
func (pf *PolicyFilter) quickActionMatch(actionSpec models.JSONActionResource, requestedAction string) bool {
	actionValues := actionSpec.GetValues()

	for _, actionPattern := range actionValues {
		if pf.fastPatternMatch(actionPattern, requestedAction) {
			return true
		}
	}

	return false
}

// quickResourceMatch performs fast resource pattern matching
func (pf *PolicyFilter) quickResourceMatch(resourceSpec models.JSONActionResource, requestedResource string) bool {
	resourceValues := resourceSpec.GetValues()

	for _, resourcePattern := range resourceValues {
		if pf.fastPatternMatch(resourcePattern, requestedResource) {
			return true
		}
	}

	return false
}

// isExcludedByNotResource checks if resource is excluded by NotResource patterns
func (pf *PolicyFilter) isExcludedByNotResource(notResourceSpec models.JSONActionResource, requestedResource string) bool {
	// If no NotResource specified, not excluded
	if !notResourceSpec.IsArray && notResourceSpec.Single == "" {
		return false
	}

	notResourceValues := notResourceSpec.GetValues()

	for _, notResourcePattern := range notResourceValues {
		if pf.fastPatternMatch(notResourcePattern, requestedResource) {
			return true // Excluded
		}
	}

	return false
}

// fastPatternMatch performs optimized pattern matching for pre-filtering
func (pf *PolicyFilter) fastPatternMatch(pattern, value string) bool {
	// Use cache for repeated patterns
	cacheKey := pattern + "|" + value
	if result, exists := pf.patternCache[cacheKey]; exists {
		return result
	}

	result := pf.performPatternMatch(pattern, value)

	// Cache result (with size limit to prevent memory bloat)
	if len(pf.patternCache) < 1000 {
		pf.patternCache[cacheKey] = result
	}

	return result
}

// performPatternMatch performs the actual pattern matching logic
func (pf *PolicyFilter) performPatternMatch(pattern, value string) bool {
	// Universal wildcard
	if pattern == "*" {
		return true
	}

	// Exact match (most common case)
	if pattern == value {
		return true
	}

	// No wildcards - exact match only
	if !strings.Contains(pattern, "*") {
		return false
	}

	// Optimized wildcard matching for common patterns
	return pf.matchWildcardPattern(pattern, value)
}

// matchWildcardPattern handles wildcard pattern matching efficiently
func (pf *PolicyFilter) matchWildcardPattern(pattern, value string) bool {
	// Handle common wildcard patterns efficiently

	// Prefix wildcard: prefix*
	if strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(value, prefix)
	}

	// Suffix wildcard: *suffix
	if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(value, suffix)
	}

	// Contains wildcard: *middle*
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") &&
		!strings.Contains(pattern[1:len(pattern)-1], "*") {
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(value, middle)
	}

	// Complex patterns - use more sophisticated matching
	return pf.matchComplexPattern(pattern, value)
}

// matchComplexPattern handles complex wildcard patterns
func (pf *PolicyFilter) matchComplexPattern(pattern, value string) bool {
	// Convert pattern to segments
	segments := strings.Split(pattern, "*")

	if len(segments) == 1 {
		// No wildcards
		return pattern == value
	}

	// Check first segment (must match start if not empty)
	if segments[0] != "" && !strings.HasPrefix(value, segments[0]) {
		return false
	}

	// Check last segment (must match end if not empty)
	lastIdx := len(segments) - 1
	if segments[lastIdx] != "" && !strings.HasSuffix(value, segments[lastIdx]) {
		return false
	}

	// Check middle segments
	searchStart := 0
	if segments[0] != "" {
		searchStart = len(segments[0])
	}

	searchEnd := len(value)
	if segments[lastIdx] != "" {
		searchEnd = len(value) - len(segments[lastIdx])
	}

	searchValue := value[searchStart:searchEnd]

	// Find all middle segments in order
	for i := 1; i < lastIdx; i++ {
		segment := segments[i]
		if segment == "" {
			continue
		}

		idx := strings.Index(searchValue, segment)
		if idx == -1 {
			return false
		}

		// Move search position past this segment
		searchValue = searchValue[idx+len(segment):]
	}

	return true
}

// GetFilteringStats returns statistics about filtering performance
func (pf *PolicyFilter) GetFilteringStats() map[string]interface{} {
	return map[string]interface{}{
		"cache_size":     len(pf.patternCache),
		"cache_enabled":  true,
		"filter_version": "1.0",
	}
}

// ClearCache clears the pattern matching cache
func (pf *PolicyFilter) ClearCache() {
	pf.patternCache = make(map[string]bool)
}

// Advanced filtering methods for specific use cases

// FilterBySubjectType filters policies that might apply to a specific subject type
func (pf *PolicyFilter) FilterBySubjectType(policies []*models.Policy, subjectType string) []*models.Policy {
	var filtered []*models.Policy

	for _, policy := range policies {
		if pf.policyMightApplyToSubjectType(policy, subjectType) {
			filtered = append(filtered, policy)
		}
	}

	return filtered
}

// FilterByResourceType filters policies that might apply to a specific resource type
func (pf *PolicyFilter) FilterByResourceType(policies []*models.Policy, resourceType string) []*models.Policy {
	var filtered []*models.Policy

	for _, policy := range policies {
		if pf.policyMightApplyToResourceType(policy, resourceType) {
			filtered = append(filtered, policy)
		}
	}

	return filtered
}

// FilterByActionCategory filters policies that might apply to a specific action category
func (pf *PolicyFilter) FilterByActionCategory(policies []*models.Policy, actionCategory string) []*models.Policy {
	var filtered []*models.Policy

	for _, policy := range policies {
		if pf.policyMightApplyToActionCategory(policy, actionCategory) {
			filtered = append(filtered, policy)
		}
	}

	return filtered
}

// Helper methods for advanced filtering

func (pf *PolicyFilter) policyMightApplyToSubjectType(policy *models.Policy, subjectType string) bool {
	// This is a heuristic - check if policy conditions mention the subject type
	for _, statement := range policy.Statement {
		if pf.conditionsMightApplyToSubjectType(statement.Condition, subjectType) {
			return true
		}
	}
	return true // Default to include if uncertain
}

func (pf *PolicyFilter) policyMightApplyToResourceType(policy *models.Policy, resourceType string) bool {
	// Check if any statement's resource patterns might match resources of this type
	for _, statement := range policy.Statement {
		resourceValues := statement.Resource.GetValues()
		for _, pattern := range resourceValues {
			// Simple heuristic: if pattern contains resource type or is wildcard
			if pattern == "*" || strings.Contains(strings.ToLower(pattern), strings.ToLower(resourceType)) {
				return true
			}
		}
	}
	return false
}

func (pf *PolicyFilter) policyMightApplyToActionCategory(policy *models.Policy, actionCategory string) bool {
	// Check if any statement's action patterns might match actions of this category
	for _, statement := range policy.Statement {
		actionValues := statement.Action.GetValues()
		for _, pattern := range actionValues {
			// Simple heuristic: if pattern contains action category or is wildcard
			if pattern == "*" || strings.Contains(strings.ToLower(pattern), strings.ToLower(actionCategory)) {
				return true
			}
		}
	}
	return false
}

func (pf *PolicyFilter) conditionsMightApplyToSubjectType(conditions map[string]interface{}, subjectType string) bool {
	// Check if conditions reference subject type
	for operator, operatorConditions := range conditions {
		if pf.operatorConditionsMightApplyToSubjectType(operator, operatorConditions, subjectType) {
			return true
		}
	}
	return false
}

func (pf *PolicyFilter) operatorConditionsMightApplyToSubjectType(operator string, operatorConditions interface{}, subjectType string) bool {
	// Simple heuristic: check if conditions mention subject type attributes
	condStr := strings.ToLower(pf.toString(operatorConditions))
	subjectTypeStr := strings.ToLower(subjectType)

	// Look for subject type references
	subjectTypePatterns := []string{
		"user:subjecttype",
		"subject:subjecttype",
		"user.subjecttype",
		"subject.subjecttype",
		subjectTypeStr,
	}

	for _, pattern := range subjectTypePatterns {
		if strings.Contains(condStr, pattern) {
			return true
		}
	}

	return false
}

func (pf *PolicyFilter) toString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Convert map to string representation
		var parts []string
		for k, val := range v {
			parts = append(parts, k+":"+pf.toString(val))
		}
		return strings.Join(parts, ",")
	case []interface{}:
		// Convert array to string representation
		var parts []string
		for _, val := range v {
			parts = append(parts, pf.toString(val))
		}
		return strings.Join(parts, ",")
	default:
		return strings.ToLower(strings.TrimSpace(pf.toStringDefault(v)))
	}
}

func (pf *PolicyFilter) toStringDefault(value interface{}) string {
	return strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", value)))
}
