package evaluator

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"abac_go_example/models"
)

// PolicyValidator validates policies against schema and business rules
type PolicyValidator struct {
	timeZones        map[string]bool
	allowedEffects   map[string]bool
	allowedOperators map[string]bool
}

// NewPolicyValidator creates a new policy validator
func NewPolicyValidator() *PolicyValidator {
	return &PolicyValidator{
		timeZones: map[string]bool{
			"UTC":              true,
			"Asia/Ho_Chi_Minh": true,
			"America/New_York": true,
			"Europe/London":    true,
			"Asia/Tokyo":       true,
			// Add more timezones as needed
		},
		allowedEffects: map[string]bool{
			"Allow": true,
			"Deny":  true,
		},
		allowedOperators: map[string]bool{
			"eq":       true,
			"ne":       true,
			"gt":       true,
			"gte":      true,
			"lt":       true,
			"lte":      true,
			"in":       true,
			"nin":      true,
			"contains": true,
			"regex":    true,
			"exists":   true,
		},
	}
}

// ValidationError represents a policy validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", ve.Field, ve.Message)
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// ValidatePolicy validates a policy against schema and business rules
func (pv *PolicyValidator) ValidatePolicy(policy *models.Policy) error {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Basic field validation
	pv.validateBasicFields(policy, result)

	// Statement validation
	pv.validateStatements(policy.Statement, result)

	if !result.Valid {
		return fmt.Errorf("policy validation failed: %v", result.Errors)
	}

	return nil
}

// ValidatePolicyRule validates a policy rule (legacy format)
func (pv *PolicyValidator) ValidatePolicyRule(rule *models.PolicyRule) error {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Validate basic rule fields
	pv.validateRuleFields(rule, result)

	// Validate time windows
	pv.validateTimeWindows(rule.TimeWindows, result)

	// Validate location conditions
	pv.validateLocationCondition(rule.Location, result)

	if !result.Valid {
		return fmt.Errorf("policy rule validation failed: %v", result.Errors)
	}

	return nil
}

// validateBasicFields validates basic policy fields
func (pv *PolicyValidator) validateBasicFields(policy *models.Policy, result *ValidationResult) {
	if policy.ID == "" {
		pv.addError(result, "id", "policy ID is required", policy.ID)
	}

	if policy.PolicyName == "" {
		pv.addError(result, "policy_name", "policy name is required", policy.PolicyName)
	}

	if policy.Version == "" {
		pv.addError(result, "version", "policy version is required", policy.Version)
	}

	if len(policy.Statement) == 0 {
		pv.addError(result, "statement", "at least one statement is required", len(policy.Statement))
	}
}

// validateStatements validates policy statements
func (pv *PolicyValidator) validateStatements(statements []models.PolicyStatement, result *ValidationResult) {
	for i, stmt := range statements {
		fieldPrefix := fmt.Sprintf("statement[%d]", i)

		// Validate effect
		if !pv.allowedEffects[stmt.Effect] {
			pv.addError(result, fieldPrefix+".effect", "invalid effect, must be 'Allow' or 'Deny'", stmt.Effect)
		}

		// Validate action
		pv.validateActionResource(stmt.Action, fieldPrefix+".action", result)

		// Validate resource
		pv.validateActionResource(stmt.Resource, fieldPrefix+".resource", result)

		// Validate conditions
		pv.validateConditions(stmt.Condition, fieldPrefix+".condition", result)
	}
}

// validateActionResource validates action or resource fields
func (pv *PolicyValidator) validateActionResource(ar models.JSONActionResource, fieldName string, result *ValidationResult) {
	if ar.IsArray {
		if len(ar.Multiple) == 0 {
			pv.addError(result, fieldName, "array cannot be empty", ar.Multiple)
		}
		for i, value := range ar.Multiple {
			if value == "" {
				pv.addError(result, fmt.Sprintf("%s[%d]", fieldName, i), "value cannot be empty", value)
			}
		}
	} else {
		if ar.Single == "" {
			pv.addError(result, fieldName, "value cannot be empty", ar.Single)
		}
	}
}

// validateConditions validates policy conditions
func (pv *PolicyValidator) validateConditions(conditions map[string]interface{}, fieldPrefix string, result *ValidationResult) {
	for operator, operatorConditions := range conditions {
		if !pv.isValidConditionOperator(operator) {
			pv.addError(result, fieldPrefix+"."+operator, "unknown condition operator", operator)
			continue
		}

		// Validate operator-specific conditions
		pv.validateOperatorConditions(operator, operatorConditions, fieldPrefix+"."+operator, result)
	}
}

// validateOperatorConditions validates conditions for a specific operator
func (pv *PolicyValidator) validateOperatorConditions(operator string, conditions interface{}, fieldPrefix string, result *ValidationResult) {
	conditionsMap, ok := conditions.(map[string]interface{})
	if !ok {
		pv.addError(result, fieldPrefix, "conditions must be an object", conditions)
		return
	}

	for key, value := range conditionsMap {
		fieldName := fieldPrefix + "." + key

		// Validate based on operator type
		switch operator {
		case "StringEquals", "StringNotEquals", "StringLike":
			if _, ok := value.(string); !ok {
				pv.addError(result, fieldName, "value must be a string for string operators", value)
			}
		case "NumericLessThan", "NumericLessThanEquals", "NumericGreaterThan", "NumericGreaterThanEquals":
			if !pv.isNumeric(value) {
				pv.addError(result, fieldName, "value must be numeric for numeric operators", value)
			}
		case "Bool":
			if _, ok := value.(bool); !ok {
				pv.addError(result, fieldName, "value must be boolean for Bool operator", value)
			}
		case "IpAddress":
			if !pv.isValidIPOrCIDR(value) {
				pv.addError(result, fieldName, "value must be valid IP address or CIDR", value)
			}
		case "DateGreaterThan", "DateLessThan":
			if !pv.isValidDateString(value) {
				pv.addError(result, fieldName, "value must be valid date string", value)
			}
		}
	}
}

// validateRuleFields validates basic rule fields
func (pv *PolicyValidator) validateRuleFields(rule *models.PolicyRule, result *ValidationResult) {
	if rule.TargetType == "" {
		pv.addError(result, "target_type", "target type is required", rule.TargetType)
	}

	validTargetTypes := []string{"subject", "resource", "action", "environment"}
	if !pv.contains(validTargetTypes, rule.TargetType) {
		pv.addError(result, "target_type", "invalid target type", rule.TargetType)
	}

	if rule.AttributePath == "" {
		pv.addError(result, "attribute_path", "attribute path is required", rule.AttributePath)
	}

	if rule.Operator == "" {
		pv.addError(result, "operator", "operator is required", rule.Operator)
	}

	if !pv.allowedOperators[rule.Operator] {
		pv.addError(result, "operator", "invalid operator", rule.Operator)
	}
}

// validateTimeWindows validates time window configurations
func (pv *PolicyValidator) validateTimeWindows(timeWindows []models.TimeWindow, result *ValidationResult) {
	for i, tw := range timeWindows {
		fieldPrefix := fmt.Sprintf("time_windows[%d]", i)

		// Validate time format
		if !pv.isValidTimeFormat(tw.StartTime) {
			pv.addError(result, fieldPrefix+".start_time", "invalid time format, expected HH:MM", tw.StartTime)
		}

		if !pv.isValidTimeFormat(tw.EndTime) {
			pv.addError(result, fieldPrefix+".end_time", "invalid time format, expected HH:MM", tw.EndTime)
		}

		// Validate days of week
		if len(tw.DaysOfWeek) == 0 {
			pv.addError(result, fieldPrefix+".days_of_week", "at least one day of week is required", tw.DaysOfWeek)
		}

		validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
		for j, day := range tw.DaysOfWeek {
			if !pv.contains(validDays, strings.ToLower(day)) {
				pv.addError(result, fmt.Sprintf("%s.days_of_week[%d]", fieldPrefix, j), "invalid day of week", day)
			}
		}

		// Validate timezone
		if tw.Timezone != "" && !pv.timeZones[tw.Timezone] {
			pv.addError(result, fieldPrefix+".timezone", "invalid or unsupported timezone", tw.Timezone)
		}

		// Validate excluded dates
		for j, date := range tw.ExcludeDates {
			if !pv.isValidDateFormat(date) {
				pv.addError(result, fmt.Sprintf("%s.exclude_dates[%d]", fieldPrefix, j), "invalid date format, expected YYYY-MM-DD", date)
			}
		}
	}
}

// validateLocationCondition validates location-based conditions
func (pv *PolicyValidator) validateLocationCondition(loc *models.LocationCondition, result *ValidationResult) {
	if loc == nil {
		return
	}

	// Validate IP ranges
	for i, ipRange := range loc.IPRanges {
		if !pv.isValidIPOrCIDR(ipRange) {
			pv.addError(result, fmt.Sprintf("location.ip_ranges[%d]", i), "invalid IP address or CIDR", ipRange)
		}
	}

	// Validate geo fencing
	if loc.GeoFencing != nil {
		if loc.GeoFencing.Latitude < -90 || loc.GeoFencing.Latitude > 90 {
			pv.addError(result, "location.geo_fencing.latitude", "latitude must be between -90 and 90", loc.GeoFencing.Latitude)
		}

		if loc.GeoFencing.Longitude < -180 || loc.GeoFencing.Longitude > 180 {
			pv.addError(result, "location.geo_fencing.longitude", "longitude must be between -180 and 180", loc.GeoFencing.Longitude)
		}

		if loc.GeoFencing.Radius <= 0 {
			pv.addError(result, "location.geo_fencing.radius", "radius must be positive", loc.GeoFencing.Radius)
		}
	}
}

// Helper validation methods

func (pv *PolicyValidator) isValidConditionOperator(operator string) bool {
	validOperators := []string{
		"StringEquals", "StringNotEquals", "StringLike",
		"NumericLessThan", "NumericLessThanEquals", "NumericGreaterThan", "NumericGreaterThanEquals",
		"Bool", "IpAddress", "DateGreaterThan", "DateLessThan",
	}
	return pv.contains(validOperators, operator)
}

func (pv *PolicyValidator) isNumeric(value interface{}) bool {
	switch v := value.(type) {
	case int, int32, int64, float32, float64:
		return true
	case string:
		_, err := parseFloat(v)
		return err == nil
	}
	return false
}

func (pv *PolicyValidator) isValidIPOrCIDR(value interface{}) bool {
	switch v := value.(type) {
	case string:
		// Check if it's a valid IP or CIDR
		if strings.Contains(v, "/") {
			// CIDR notation
			_, _, err := parseIPNet(v)
			return err == nil
		}

		// Single IP
		return parseIP(v) != nil
	default:
		return false
	}
}

func (pv *PolicyValidator) isValidDateString(value interface{}) bool {
	switch v := value.(type) {
	case string:
		// Try parsing as RFC3339
		_, err := time.Parse(time.RFC3339, v)
		return err == nil
	default:
		return false
	}
}

func (pv *PolicyValidator) isValidTimeFormat(timeStr string) bool {
	// Validate HH:MM format
	matched, _ := regexp.MatchString(`^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`, timeStr)
	return matched
}

func (pv *PolicyValidator) isValidDateFormat(dateStr string) bool {
	// Validate YYYY-MM-DD format
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func (pv *PolicyValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (pv *PolicyValidator) addError(result *ValidationResult, field, message string, value interface{}) {
	result.Valid = false
	result.Errors = append(result.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// Placeholder functions for network operations (would use net package in real implementation)
func parseFloat(s string) (float64, error) {
	// This would use strconv.ParseFloat in real implementation
	return 0, fmt.Errorf("not implemented")
}

func parseIP(s string) interface{} {
	// This would use net.ParseIP in real implementation
	return nil
}

func parseIPNet(s string) (interface{}, interface{}, error) {
	// This would use net.ParseCIDR in real implementation
	return nil, nil, fmt.Errorf("not implemented")
}
