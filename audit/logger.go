package audit

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"abac_go_example/models"
)

// AuditLogger handles audit logging for policy evaluations
type AuditLogger struct {
	logFile *os.File
	logger  *log.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logFilePath string) (*AuditLogger, error) {
	var logFile *os.File
	var err error

	if logFilePath != "" {
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	} else {
		logFile = os.Stdout
	}

	logger := log.New(logFile, "", 0) // No default timestamp, we'll add our own

	return &AuditLogger{
		logFile: logFile,
		logger:  logger,
	}, nil
}

// LogEvaluation logs a policy evaluation result
func (a *AuditLogger) LogEvaluation(request *models.EvaluationRequest, decision *models.Decision, context *models.EvaluationContext) error {
	auditContext := map[string]interface{}{
		"matched_policies": decision.MatchedPolicies,
		"reason":           decision.Reason,
	}

	// Safely add environment context
	if context.Environment != nil {
		if sourceIP, ok := context.Environment["source_ip"]; ok {
			auditContext["source_ip"] = sourceIP
		}
		if userAgent, ok := context.Environment["user_agent"]; ok {
			auditContext["user_agent"] = userAgent
		}
		if timestamp, ok := context.Environment["timestamp"]; ok {
			auditContext["timestamp"] = timestamp
		}
	}

	// Safely add subject context
	if context.Subject != nil {
		auditContext["subject_type"] = context.Subject.SubjectType
	}

	// Safely add resource context
	if context.Resource != nil {
		auditContext["resource_type"] = context.Resource.ResourceType
	}

	// Safely add action context
	if context.Action != nil {
		auditContext["action_category"] = context.Action.ActionCategory
	}

	// Get subject ID from Subject interface
	subjectID := ""
	if request.Subject != nil {
		subjectID = request.Subject.GetID()
	}

	auditEntry := models.AuditLog{
		RequestID:    request.RequestID,
		SubjectID:    subjectID,
		ResourceID:   request.ResourceID,
		ActionID:     request.Action,
		Decision:     decision.Result,
		EvaluationMs: decision.EvaluationTimeMs,
		CreatedAt:    time.Now(),
		Context:      auditContext,
	}

	return a.logEntry(auditEntry)
}

// LogAccessAttempt logs an access attempt with additional context
func (a *AuditLogger) LogAccessAttempt(request *models.EvaluationRequest, decision *models.Decision, additionalContext map[string]interface{}) error {
	// Get subject ID from Subject interface
	subjectID := ""
	if request.Subject != nil {
		subjectID = request.Subject.GetID()
	}

	auditEntry := models.AuditLog{
		RequestID:    request.RequestID,
		SubjectID:    subjectID,
		ResourceID:   request.ResourceID,
		ActionID:     request.Action,
		Decision:     decision.Result,
		EvaluationMs: decision.EvaluationTimeMs,
		CreatedAt:    time.Now(),
		Context:      make(map[string]interface{}),
	}

	// Copy request context
	for k, v := range request.Context {
		auditEntry.Context[k] = v
	}

	// Add decision context
	auditEntry.Context["matched_policies"] = decision.MatchedPolicies
	auditEntry.Context["reason"] = decision.Reason

	// Add additional context
	for k, v := range additionalContext {
		auditEntry.Context[k] = v
	}

	return a.logEntry(auditEntry)
}

// LogSecurityEvent logs security-related events
func (a *AuditLogger) LogSecurityEvent(eventType string, subjectID string, details map[string]interface{}) error {
	auditEntry := models.AuditLog{
		RequestID: fmt.Sprintf("security-%d", time.Now().UnixNano()),
		SubjectID: subjectID,
		Decision:  eventType,
		CreatedAt: time.Now(),
		Context: map[string]interface{}{
			"event_type": eventType,
			"details":    details,
		},
	}

	return a.logEntry(auditEntry)
}

// LogPolicyChange logs policy configuration changes
func (a *AuditLogger) LogPolicyChange(changeType string, policyID string, changedBy string, changes map[string]interface{}) error {
	auditEntry := models.AuditLog{
		RequestID:  fmt.Sprintf("policy-change-%d", time.Now().UnixNano()),
		SubjectID:  changedBy,
		ResourceID: policyID,
		Decision:   changeType,
		CreatedAt:  time.Now(),
		Context: map[string]interface{}{
			"change_type": changeType,
			"changes":     changes,
			"changed_by":  changedBy,
		},
	}

	return a.logEntry(auditEntry)
}

// logEntry writes an audit entry to the log
func (a *AuditLogger) logEntry(entry models.AuditLog) error {
	// Convert to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Log the JSON entry
	a.logger.Println(string(jsonData))

	return nil
}

// GetAuditTrail retrieves audit trail for a specific subject or resource
func (a *AuditLogger) GetAuditTrail(subjectID, resourceID string, limit int) ([]*models.AuditLog, error) {
	// In a real implementation, this would query a database
	// For this mock implementation, we'll return empty results
	// as we're only writing to log files
	return []*models.AuditLog{}, nil
}

// Close closes the audit logger and any open files
func (a *AuditLogger) Close() error {
	if a.logFile != nil && a.logFile != os.Stdout {
		return a.logFile.Close()
	}
	return nil
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalEvaluations int            `json:"total_evaluations"`
	PermitCount      int            `json:"permit_count"`
	DenyCount        int            `json:"deny_count"`
	NotApplicable    int            `json:"not_applicable"`
	AvgEvaluationMs  float64        `json:"avg_evaluation_ms"`
	TopPolicies      map[string]int `json:"top_policies"`
	TopSubjects      map[string]int `json:"top_subjects"`
	TopResources     map[string]int `json:"top_resources"`
	SecurityEvents   int            `json:"security_events"`
	PolicyChanges    int            `json:"policy_changes"`
}

// GetStats returns audit statistics (mock implementation)
func (a *AuditLogger) GetStats(since time.Time) (*AuditStats, error) {
	// In a real implementation, this would analyze log data
	// For this mock, return empty stats
	return &AuditStats{
		TotalEvaluations: 0,
		PermitCount:      0,
		DenyCount:        0,
		NotApplicable:    0,
		AvgEvaluationMs:  0.0,
		TopPolicies:      make(map[string]int),
		TopSubjects:      make(map[string]int),
		TopResources:     make(map[string]int),
		SecurityEvents:   0,
		PolicyChanges:    0,
	}, nil
}

// ComplianceReport generates a compliance report
type ComplianceReport struct {
	Period           string            `json:"period"`
	TotalAccesses    int               `json:"total_accesses"`
	SuccessfulAccess int               `json:"successful_access"`
	DeniedAccess     int               `json:"denied_access"`
	PolicyViolations []PolicyViolation `json:"policy_violations"`
	UnusualActivity  []UnusualActivity `json:"unusual_activity"`
	GeneratedAt      time.Time         `json:"generated_at"`
}

type PolicyViolation struct {
	SubjectID   string    `json:"subject_id"`
	ResourceID  string    `json:"resource_id"`
	Action      string    `json:"action"`
	PolicyID    string    `json:"policy_id"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

type UnusualActivity struct {
	SubjectID    string                 `json:"subject_id"`
	ActivityType string                 `json:"activity_type"`
	Count        int                    `json:"count"`
	Details      map[string]interface{} `json:"details"`
	FirstSeen    time.Time              `json:"first_seen"`
	LastSeen     time.Time              `json:"last_seen"`
}

// GenerateComplianceReport generates a compliance report for a given period
func (a *AuditLogger) GenerateComplianceReport(since, until time.Time) (*ComplianceReport, error) {
	// Mock implementation - in reality would analyze audit logs
	report := &ComplianceReport{
		Period:           fmt.Sprintf("%s to %s", since.Format("2006-01-02"), until.Format("2006-01-02")),
		TotalAccesses:    0,
		SuccessfulAccess: 0,
		DeniedAccess:     0,
		PolicyViolations: []PolicyViolation{},
		UnusualActivity:  []UnusualActivity{},
		GeneratedAt:      time.Now(),
	}

	return report, nil
}
