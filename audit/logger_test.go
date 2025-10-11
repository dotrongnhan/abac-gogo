package audit

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"abac_go_example/models"
)

func TestAuditLoggerCreation(t *testing.T) {
	// Test with temporary file
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	if logger.logFile == nil {
		t.Error("Logger should have a log file")
	}

	if logger.logger == nil {
		t.Error("Logger should have an internal logger")
	}
}

func TestAuditLoggerWithStdout(t *testing.T) {
	// Test with stdout (empty filename)
	logger, err := NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to create audit logger with stdout: %v", err)
	}
	defer logger.Close()

	if logger.logFile != os.Stdout {
		t.Error("Logger should use stdout when no filename provided")
	}
}

func TestLogEvaluation(t *testing.T) {
	// Create temporary log file
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	// Create test data
	request := &models.EvaluationRequest{
		RequestID:  "test-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"source_ip": "10.0.1.100",
			"timestamp": "2024-01-15T14:00:00Z",
		},
	}

	decision := &models.Decision{
		Result:           "permit",
		MatchedPolicies:  []string{"pol-001", "pol-002"},
		EvaluationTimeMs: 5,
		Reason:           "Access granted by matching permit policies",
	}

	context := &models.EvaluationContext{
		Subject: &models.Subject{
			ID:          "sub-001",
			SubjectType: "user",
		},
		Resource: &models.Resource{
			ID:           "res-001",
			ResourceType: "api_endpoint",
		},
		Action: &models.Action{
			ID:             "act-001",
			ActionName:     "read",
			ActionCategory: "crud",
		},
		Environment: map[string]interface{}{
			"source_ip": "10.0.1.100",
			"timestamp": "2024-01-15T14:00:00Z",
		},
		Timestamp: time.Now(),
	}

	// Log the evaluation
	err = logger.LogEvaluation(request, decision, context)
	if err != nil {
		t.Fatalf("Failed to log evaluation: %v", err)
	}

	// Read the log file and verify content
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse the JSON log entry
	var logEntry models.AuditLog
	err = json.Unmarshal(content, &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify log entry content
	if logEntry.RequestID != request.RequestID {
		t.Errorf("Expected RequestID %s, got %s", request.RequestID, logEntry.RequestID)
	}

	if logEntry.SubjectID != request.SubjectID {
		t.Errorf("Expected SubjectID %s, got %s", request.SubjectID, logEntry.SubjectID)
	}

	if logEntry.Decision != decision.Result {
		t.Errorf("Expected Decision %s, got %s", decision.Result, logEntry.Decision)
	}

	if logEntry.EvaluationMs != decision.EvaluationTimeMs {
		t.Errorf("Expected EvaluationMs %d, got %d", decision.EvaluationTimeMs, logEntry.EvaluationMs)
	}

	// Verify context data
	if matchedPolicies, exists := logEntry.Context["matched_policies"]; !exists {
		t.Error("Log entry should contain matched_policies in context")
	} else {
		policies := matchedPolicies.([]interface{})
		if len(policies) != len(decision.MatchedPolicies) {
			t.Errorf("Expected %d matched policies, got %d", len(decision.MatchedPolicies), len(policies))
		}
	}

	if reason, exists := logEntry.Context["reason"]; !exists {
		t.Error("Log entry should contain reason in context")
	} else if reason != decision.Reason {
		t.Errorf("Expected reason %s, got %s", decision.Reason, reason)
	}
}

func TestLogAccessAttempt(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	request := &models.EvaluationRequest{
		RequestID:  "access-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"source_ip":  "10.0.1.100",
			"user_agent": "Mozilla/5.0",
		},
	}

	decision := &models.Decision{
		Result:           "deny",
		MatchedPolicies:  []string{"pol-004"},
		EvaluationTimeMs: 3,
		Reason:           "Access denied due to probation",
	}

	additionalContext := map[string]interface{}{
		"security_event": "probation_violation",
		"severity":       "medium",
	}

	err = logger.LogAccessAttempt(request, decision, additionalContext)
	if err != nil {
		t.Fatalf("Failed to log access attempt: %v", err)
	}

	// Verify the log was written
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var logEntry models.AuditLog
	err = json.Unmarshal(content, &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify additional context was included
	if securityEvent, exists := logEntry.Context["security_event"]; !exists {
		t.Error("Log entry should contain security_event")
	} else if securityEvent != "probation_violation" {
		t.Errorf("Expected security_event probation_violation, got %s", securityEvent)
	}

	if severity, exists := logEntry.Context["severity"]; !exists {
		t.Error("Log entry should contain severity")
	} else if severity != "medium" {
		t.Errorf("Expected severity medium, got %s", severity)
	}
}

func TestLogSecurityEvent(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	eventType := "privilege_escalation_attempt"
	subjectID := "sub-003"
	details := map[string]interface{}{
		"attempted_resource": "res-002",
		"required_clearance": 3,
		"user_clearance":     1,
		"blocked":            true,
	}

	err = logger.LogSecurityEvent(eventType, subjectID, details)
	if err != nil {
		t.Fatalf("Failed to log security event: %v", err)
	}

	// Verify the log was written
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var logEntry models.AuditLog
	err = json.Unmarshal(content, &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if logEntry.SubjectID != subjectID {
		t.Errorf("Expected SubjectID %s, got %s", subjectID, logEntry.SubjectID)
	}

	if logEntry.Decision != eventType {
		t.Errorf("Expected Decision %s, got %s", eventType, logEntry.Decision)
	}

	if !strings.HasPrefix(logEntry.RequestID, "security-") {
		t.Error("Security event RequestID should start with 'security-'")
	}

	// Verify event details
	if contextEventType, exists := logEntry.Context["event_type"]; !exists {
		t.Error("Log entry should contain event_type")
	} else if contextEventType != eventType {
		t.Errorf("Expected event_type %s, got %s", eventType, contextEventType)
	}

	if contextDetails, exists := logEntry.Context["details"]; !exists {
		t.Error("Log entry should contain details")
	} else {
		detailsMap := contextDetails.(map[string]interface{})
		if blocked, exists := detailsMap["blocked"]; !exists {
			t.Error("Details should contain blocked field")
		} else if blocked != true {
			t.Errorf("Expected blocked true, got %v", blocked)
		}
	}
}

func TestLogPolicyChange(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	changeType := "policy_update"
	policyID := "pol-001"
	changedBy := "admin-001"
	changes := map[string]interface{}{
		"field":     "priority",
		"old_value": 100,
		"new_value": 50,
	}

	err = logger.LogPolicyChange(changeType, policyID, changedBy, changes)
	if err != nil {
		t.Fatalf("Failed to log policy change: %v", err)
	}

	// Verify the log was written
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var logEntry models.AuditLog
	err = json.Unmarshal(content, &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if logEntry.SubjectID != changedBy {
		t.Errorf("Expected SubjectID %s, got %s", changedBy, logEntry.SubjectID)
	}

	if logEntry.ResourceID != policyID {
		t.Errorf("Expected ResourceID %s, got %s", policyID, logEntry.ResourceID)
	}

	if logEntry.Decision != changeType {
		t.Errorf("Expected Decision %s, got %s", changeType, logEntry.Decision)
	}

	if !strings.HasPrefix(logEntry.RequestID, "policy-change-") {
		t.Error("Policy change RequestID should start with 'policy-change-'")
	}

	// Verify change details
	if contextChangeType, exists := logEntry.Context["change_type"]; !exists {
		t.Error("Log entry should contain change_type")
	} else if contextChangeType != changeType {
		t.Errorf("Expected change_type %s, got %s", changeType, contextChangeType)
	}

	if contextChanges, exists := logEntry.Context["changes"]; !exists {
		t.Error("Log entry should contain changes")
	} else {
		changesMap := contextChanges.(map[string]interface{})
		if field, exists := changesMap["field"]; !exists {
			t.Error("Changes should contain field")
		} else if field != "priority" {
			t.Errorf("Expected field priority, got %s", field)
		}
	}
}

func TestMultipleLogEntries(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	// Log multiple events
	for i := 0; i < 3; i++ {
		err = logger.LogSecurityEvent("test_event", "sub-001", map[string]interface{}{
			"iteration": i,
		})
		if err != nil {
			t.Fatalf("Failed to log security event %d: %v", i, err)
		}
	}

	// Read and verify all entries
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 log lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var logEntry models.AuditLog
		err = json.Unmarshal([]byte(line), &logEntry)
		if err != nil {
			t.Errorf("Failed to parse log entry %d: %v", i, err)
		}

		if logEntry.Decision != "test_event" {
			t.Errorf("Expected Decision test_event, got %s", logEntry.Decision)
		}
	}
}

func TestGetStats(t *testing.T) {
	logger, err := NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	since := time.Now().Add(-24 * time.Hour)
	stats, err := logger.GetStats(since)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// Verify stats structure (mock implementation returns empty stats)
	if stats == nil {
		t.Error("Stats should not be nil")
	}

	if stats.TopPolicies == nil {
		t.Error("TopPolicies should not be nil")
	}

	if stats.TopSubjects == nil {
		t.Error("TopSubjects should not be nil")
	}

	if stats.TopResources == nil {
		t.Error("TopResources should not be nil")
	}
}

func TestGenerateComplianceReport(t *testing.T) {
	logger, err := NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	since := time.Now().Add(-30 * 24 * time.Hour)
	until := time.Now()

	report, err := logger.GenerateComplianceReport(since, until)
	if err != nil {
		t.Fatalf("Failed to generate compliance report: %v", err)
	}

	if report == nil {
		t.Error("Report should not be nil")
	}

	if report.Period == "" {
		t.Error("Report should have a period")
	}

	if report.PolicyViolations == nil {
		t.Error("PolicyViolations should not be nil")
	}

	if report.UnusualActivity == nil {
		t.Error("UnusualActivity should not be nil")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Report should have GeneratedAt timestamp")
	}
}

func TestCloseLogger(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err := NewAuditLogger(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Close should not return error
	err = logger.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}

	// Test closing stdout logger (should not error)
	stdoutLogger, err := NewAuditLogger("")
	if err != nil {
		t.Fatalf("Failed to create stdout logger: %v", err)
	}

	err = stdoutLogger.Close()
	if err != nil {
		t.Errorf("Closing stdout logger should not return error: %v", err)
	}
}
