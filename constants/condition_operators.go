package constants

// ConditionOperatorType represents the type of condition operator
type ConditionOperatorType string

// Condition operator constants for string operations
const (
	ConditionStringEquals    ConditionOperatorType = "StringEquals"
	ConditionStringNotEquals ConditionOperatorType = "StringNotEquals"
	ConditionStringLike      ConditionOperatorType = "StringLike"
)

// Condition operator constants for numeric operations
const (
	ConditionNumericLessThan          ConditionOperatorType = "NumericLessThan"
	ConditionNumericLessThanEquals    ConditionOperatorType = "NumericLessThanEquals"
	ConditionNumericGreaterThan       ConditionOperatorType = "NumericGreaterThan"
	ConditionNumericGreaterThanEquals ConditionOperatorType = "NumericGreaterThanEquals"
)

// Condition operator constants for boolean operations
const (
	ConditionBool ConditionOperatorType = "Bool"
)

// Condition operator constants for network operations
const (
	ConditionIpAddress ConditionOperatorType = "IpAddress"
)

// Condition operator constants for date/time operations
const (
	ConditionDateGreaterThan ConditionOperatorType = "DateGreaterThan"
	ConditionDateLessThan    ConditionOperatorType = "DateLessThan"
)

// Condition operator constants for logical operations
const (
	ConditionAnd ConditionOperatorType = "And"
	ConditionOr  ConditionOperatorType = "Or"
	ConditionNot ConditionOperatorType = "Not"
)

// AllConditionOperatorTypes returns all available condition operator types
func AllConditionOperatorTypes() []ConditionOperatorType {
	return []ConditionOperatorType{
		ConditionStringEquals,
		ConditionStringNotEquals,
		ConditionStringLike,
		ConditionNumericLessThan,
		ConditionNumericLessThanEquals,
		ConditionNumericGreaterThan,
		ConditionNumericGreaterThanEquals,
		ConditionBool,
		ConditionIpAddress,
		ConditionDateGreaterThan,
		ConditionDateLessThan,
		ConditionAnd,
		ConditionOr,
		ConditionNot,
	}
}

// String returns the string representation of the condition operator type
func (cot ConditionOperatorType) String() string {
	return string(cot)
}

// IsValid checks if the condition operator type is valid
func (cot ConditionOperatorType) IsValid() bool {
	for _, validOp := range AllConditionOperatorTypes() {
		if cot == validOp {
			return true
		}
	}
	return false
}

// GetOperatorCategory returns the category of the operator (string, numeric, boolean, etc.)
func (cot ConditionOperatorType) GetOperatorCategory() string {
	switch cot {
	case ConditionStringEquals, ConditionStringNotEquals, ConditionStringLike:
		return "string"
	case ConditionNumericLessThan, ConditionNumericLessThanEquals,
		ConditionNumericGreaterThan, ConditionNumericGreaterThanEquals:
		return "numeric"
	case ConditionBool:
		return "boolean"
	case ConditionIpAddress:
		return "network"
	case ConditionDateGreaterThan, ConditionDateLessThan:
		return "date"
	case ConditionAnd, ConditionOr, ConditionNot:
		return "logical"
	default:
		return "unknown"
	}
}
