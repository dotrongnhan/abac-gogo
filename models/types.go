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

// JSONStatements is a custom type for handling []PolicyStatement in GORM
type JSONStatements []PolicyStatement

// Value implements the driver.Valuer interface for GORM
func (j JSONStatements) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for GORM
func (j *JSONStatements) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan %T into JSONStatements", value)
	}

	return json.Unmarshal(bytes, j)
}

// JSONActionResource is a custom type for handling string or []string
type JSONActionResource struct {
	Single   string
	Multiple []string
}

// Value implements the driver.Valuer interface for GORM
func (j JSONActionResource) Value() (driver.Value, error) {
	if j.Single != "" {
		return json.Marshal(j.Single)
	}
	return json.Marshal(j.Multiple)
}

// Scan implements the sql.Scanner interface for GORM
func (j *JSONActionResource) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONActionResource", value)
	}

	// Try to unmarshal as array first
	var arr []string
	if err := json.Unmarshal(bytes, &arr); err == nil {
		j.Multiple = arr
		j.Single = ""
		return nil
	}

	// If that fails, try as single string
	var str string
	if err := json.Unmarshal(bytes, &str); err == nil {
		j.Single = str
		j.Multiple = nil
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into JSONActionResource", string(bytes))
}

// UnmarshalJSON implements custom JSON unmarshaling
func (j *JSONActionResource) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		j.Multiple = arr
		j.Single = ""
		return nil
	}

	// If that fails, try as single string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		j.Single = str
		j.Multiple = nil
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into JSONActionResource", string(data))
}

// MarshalJSON implements custom JSON marshaling
func (j JSONActionResource) MarshalJSON() ([]byte, error) {
	if j.Single != "" {
		return json.Marshal(j.Single)
	}
	return json.Marshal(j.Multiple)
}

// GetValues returns all values as a slice
func (j JSONActionResource) GetValues() []string {
	if j.Single != "" {
		return []string{j.Single}
	}
	return j.Multiple
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

// Policy represents an access control policy following the new JSON schema
type Policy struct {
	ID          string         `json:"id" gorm:"primaryKey;size:255"`
	PolicyName  string         `json:"policy_name" gorm:"size:255;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text"`
	Effect      string         `json:"effect,omitempty" gorm:"size:20;default:'permit'"`
	Version     string         `json:"version" gorm:"size:50;not null"`
	Statement   JSONStatements `json:"statement" gorm:"type:jsonb"`
	Enabled     bool           `json:"enabled" gorm:"default:true;index"`
	CreatedAt   time.Time      `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Policy
func (Policy) TableName() string {
	return "policies"
}

// PolicyRule represents a single rule within a policy (legacy format)
type PolicyRule struct {
	ID            string             `json:"id,omitempty"`
	TargetType    string             `json:"target_type"`    // "subject", "resource", "action", "environment"
	AttributePath string             `json:"attribute_path"` // e.g., "attributes.department"
	Operator      string             `json:"operator"`       // "eq", "in", "contains", "regex", etc.
	ExpectedValue interface{}        `json:"expected_value"`
	IsNegative    bool               `json:"is_negative,omitempty"`
	RuleOrder     int                `json:"rule_order,omitempty"`
	TimeWindows   []TimeWindow       `json:"time_windows,omitempty"`
	Location      *LocationCondition `json:"location,omitempty"`
}

// PolicyStatement represents a statement in the new policy format
type PolicyStatement struct {
	Sid         string             `json:"Sid,omitempty"`         // Statement ID for debugging
	Effect      string             `json:"Effect"`                // "Allow" or "Deny"
	Action      JSONActionResource `json:"Action"`                // string or []string
	Resource    JSONActionResource `json:"Resource"`              // string or []string
	NotResource JSONActionResource `json:"NotResource,omitempty"` // Exclusion patterns
	Condition   JSONMap            `json:"Condition,omitempty"`   // Runtime conditions
}

// PolicyDocument represents the complete policy document
type PolicyDocument struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
}

// EvaluationRequest represents a request for policy evaluation
type EvaluationRequest struct {
	RequestID  string                 `json:"request_id"`
	SubjectID  string                 `json:"subject_id"` // Deprecated: Use Subject instead
	ResourceID string                 `json:"resource_id"`
	Action     string                 `json:"action"`
	Context    map[string]interface{} `json:"context"`
	// Enhanced fields for improved PDP
	Environment *EnvironmentInfo `json:"environment,omitempty"`
	Timestamp   *time.Time       `json:"timestamp,omitempty"`
	// New Subject interface field (preferred over SubjectID)
	Subject SubjectInterface `json:"-"` // Not serialized to JSON
}

// EnvironmentInfo represents environmental context for basic PDP
type EnvironmentInfo struct {
	ClientIP   string                 `json:"client_ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Country    string                 `json:"country,omitempty"`
	Region     string                 `json:"region,omitempty"`
	TimeOfDay  string                 `json:"time_of_day,omitempty"` // "14:30"
	DayOfWeek  string                 `json:"day_of_week,omitempty"` // "Monday"
	Attributes map[string]interface{} `json:"attributes,omitempty"`
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

// Enhanced decision types for improved PDP
type DecisionType string

const (
	DecisionPermit        DecisionType = "PERMIT"
	DecisionDeny          DecisionType = "DENY"
	DecisionNotApplicable DecisionType = "NOT_APPLICABLE"
	DecisionIndeterminate DecisionType = "INDETERMINATE"
)

// DecisionRequest represents input to enhanced PDP
type DecisionRequest struct {
	Subject     *Subject               `json:"subject"`
	Resource    *Resource              `json:"resource"`
	Action      *Action                `json:"action"`
	Environment *Environment           `json:"environment"`
	Context     map[string]interface{} `json:"context"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// DecisionResponse represents enhanced PDP output
type DecisionResponse struct {
	Decision    DecisionType  `json:"decision"`
	Reason      string        `json:"reason"`
	Policies    []string      `json:"applicable_policies"`
	EvaluatedAt time.Time     `json:"evaluated_at"`
	Duration    time.Duration `json:"evaluation_duration"`
	RequestID   string        `json:"request_id,omitempty"`
}

// Environment represents environmental attributes for policy evaluation
type Environment struct {
	Timestamp  time.Time              `json:"timestamp"`
	ClientIP   string                 `json:"client_ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Location   *LocationInfo          `json:"location,omitempty"`
	TimeOfDay  string                 `json:"time_of_day,omitempty"`
	DayOfWeek  string                 `json:"day_of_week,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// LocationInfo represents geographical location information
type LocationInfo struct {
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// GetClientIP returns the client IP address
func (e *Environment) GetClientIP() string {
	return e.ClientIP
}

// GetLatitude returns the latitude
func (e *Environment) GetLatitude() float64 {
	if e.Location != nil {
		return e.Location.Latitude
	}
	return 0
}

// GetLongitude returns the longitude
func (e *Environment) GetLongitude() float64 {
	if e.Location != nil {
		return e.Location.Longitude
	}
	return 0
}

// TimeWindow represents time-based access control
type TimeWindow struct {
	StartTime    string   `json:"start_time"`    // "09:00"
	EndTime      string   `json:"end_time"`      // "17:00"
	DaysOfWeek   []string `json:"days_of_week"`  // ["monday", "tuesday"]
	Timezone     string   `json:"timezone"`      // "Asia/Ho_Chi_Minh"
	ExcludeDates []string `json:"exclude_dates"` // ["2025-12-25", "2025-01-01"]
}

// LocationCondition represents location-based access control
type LocationCondition struct {
	AllowedCountries []string           `json:"allowed_countries,omitempty"`
	AllowedRegions   []string           `json:"allowed_regions,omitempty"`
	IPRanges         []string           `json:"ip_ranges,omitempty"`
	GeoFencing       *GeoFenceCondition `json:"geo_fencing,omitempty"`
}

// GeoFenceCondition represents geographical fencing
type GeoFenceCondition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius_km"`
}

// BooleanExpression represents complex boolean expressions
type BooleanExpression struct {
	Type      string             `json:"type"`               // "simple" or "compound"
	Operator  string             `json:"operator,omitempty"` // "and", "or", "not"
	Condition *SimpleCondition   `json:"condition,omitempty"`
	Left      *BooleanExpression `json:"left,omitempty"`
	Right     *BooleanExpression `json:"right,omitempty"`
}

// SimpleCondition represents a simple condition
type SimpleCondition struct {
	AttributePath string      `json:"attribute_path"` // "user.department"
	Operator      string      `json:"operator"`       // "eq", "gt", "in"
	Value         interface{} `json:"value"`
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
