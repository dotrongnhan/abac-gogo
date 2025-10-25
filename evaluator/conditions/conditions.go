package conditions

// ComplexCondition represents a complex condition with logical operators
// This is kept for backward compatibility with existing policy formats
type ComplexCondition struct {
	Type       string             `json:"type"`                 // "simple", "logical"
	Operator   string             `json:"operator,omitempty"`   // For simple: StringEquals, etc. For logical: And, Or, Not
	Key        string             `json:"key,omitempty"`        // For simple conditions: attribute path
	Value      interface{}        `json:"value,omitempty"`      // For simple conditions: expected value
	Left       *ComplexCondition  `json:"left,omitempty"`       // For logical conditions: left operand
	Right      *ComplexCondition  `json:"right,omitempty"`      // For logical conditions: right operand
	Operand    *ComplexCondition  `json:"operand,omitempty"`    // For NOT operator: single operand
	Conditions []ComplexCondition `json:"conditions,omitempty"` // For array of conditions (alternative format)
}
