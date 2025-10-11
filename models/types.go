package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap is a custom type for handling map[string]interface{} in GORM
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for GORM
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for GORM
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// JSONStringSlice is a custom type for handling []string in GORM
type JSONStringSlice []string

// Value implements the driver.Valuer interface for GORM
func (j JSONStringSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for GORM
func (j *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONStringSlice", value)
	}

	return json.Unmarshal(bytes, j)
}

// JSONPolicyRules is a custom type for handling []PolicyRule in GORM
type JSONPolicyRules []PolicyRule

// Value implements the driver.Valuer interface for GORM
func (j JSONPolicyRules) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for GORM
func (j *JSONPolicyRules) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONPolicyRules", value)
	}

	return json.Unmarshal(bytes, j)
}

// Subject represents a user, service, or application
type Subject struct {
	ID          string    `json:"id" gorm:"primaryKey;size:255"`
	ExternalID  string    `json:"external_id" gorm:"size:255;index"`
	SubjectType string    `json:"subject_type" gorm:"size:100;not null;index"`
	Metadata    JSONMap   `json:"metadata" gorm:"type:jsonb"`
	Attributes  JSONMap   `json:"attributes" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Subject
func (Subject) TableName() string {
	return "subjects"
}

// Resource represents an API, document, or data object
type Resource struct {
	ID           string    `json:"id" gorm:"primaryKey;size:255"`
	ResourceType string    `json:"resource_type" gorm:"size:100;not null;index"`
	ResourceID   string    `json:"resource_id" gorm:"size:255;index"`
	Path         string    `json:"path" gorm:"size:500"`
	ParentID     string    `json:"parent_id,omitempty" gorm:"size:255;index"`
	Metadata     JSONMap   `json:"metadata" gorm:"type:jsonb"`
	Attributes   JSONMap   `json:"attributes" gorm:"type:jsonb"`
	CreatedAt    time.Time `json:"created_at,omitempty" gorm:"autoCreateTime"`
}

// TableName specifies the table name for Resource
func (Resource) TableName() string {
	return "resources"
}

// Action represents an operation that can be performed
type Action struct {
	ID             string `json:"id" gorm:"primaryKey;size:255"`
	ActionName     string `json:"action_name" gorm:"size:100;not null;uniqueIndex"`
	ActionCategory string `json:"action_category" gorm:"size:100;index"`
	Description    string `json:"description" gorm:"type:text"`
	IsSystem       bool   `json:"is_system" gorm:"default:false;index"`
}

// TableName specifies the table name for Action
func (Action) TableName() string {
	return "actions"
}

// Policy represents an access control policy
type Policy struct {
	ID               string          `json:"id" gorm:"primaryKey;size:255"`
	PolicyName       string          `json:"policy_name" gorm:"size:255;not null;uniqueIndex"`
	Description      string          `json:"description" gorm:"type:text"`
	Effect           string          `json:"effect" gorm:"size:10;not null;index"` // "permit" or "deny"
	Priority         int             `json:"priority" gorm:"not null;index"`
	Enabled          bool            `json:"enabled" gorm:"default:true;index"`
	Version          int             `json:"version" gorm:"default:1"`
	Conditions       JSONMap         `json:"conditions" gorm:"type:jsonb"`
	Rules            JSONPolicyRules `json:"rules" gorm:"type:jsonb"`
	Actions          JSONStringSlice `json:"actions" gorm:"type:jsonb"`
	ResourcePatterns JSONStringSlice `json:"resource_patterns" gorm:"type:jsonb"`
	CreatedAt        time.Time       `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Policy
func (Policy) TableName() string {
	return "policies"
}

// PolicyRule represents a single rule within a policy
type PolicyRule struct {
	ID            string      `json:"id,omitempty"`
	TargetType    string      `json:"target_type"`    // "subject", "resource", "action", "environment"
	AttributePath string      `json:"attribute_path"` // e.g., "attributes.department"
	Operator      string      `json:"operator"`       // "eq", "in", "contains", "regex", etc.
	ExpectedValue interface{} `json:"expected_value"`
	IsNegative    bool        `json:"is_negative,omitempty"`
	RuleOrder     int         `json:"rule_order,omitempty"`
}

// EvaluationRequest represents a request for policy evaluation
type EvaluationRequest struct {
	RequestID  string                 `json:"request_id"`
	SubjectID  string                 `json:"subject_id"`
	ResourceID string                 `json:"resource_id"`
	Action     string                 `json:"action"`
	Context    map[string]interface{} `json:"context"`
}

// EvaluationContext contains all the context needed for evaluation
type EvaluationContext struct {
	Subject     *Subject
	Resource    *Resource
	Action      *Action
	Environment map[string]interface{}
	Timestamp   time.Time
}

// Decision represents the result of a policy evaluation
type Decision struct {
	Result           string   `json:"result"` // "permit", "deny", "not_applicable"
	MatchedPolicies  []string `json:"matched_policies"`
	EvaluationTimeMs int      `json:"evaluation_time_ms"`
	Reason           string   `json:"reason,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID    string    `json:"request_id" gorm:"size:255;not null;index"`
	SubjectID    string    `json:"subject_id" gorm:"size:255;not null;index"`
	ResourceID   string    `json:"resource_id" gorm:"size:255;not null;index"`
	ActionID     string    `json:"action_id" gorm:"size:255;not null;index"`
	Decision     string    `json:"decision" gorm:"size:20;not null;index"`
	EvaluationMs int       `json:"evaluation_ms" gorm:"not null"`
	Context      JSONMap   `json:"context" gorm:"type:jsonb"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}
