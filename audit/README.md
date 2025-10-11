# Audit Package - Compliance & Security Logging

## 📋 Tổng Quan

Package `audit` cung cấp **Comprehensive Audit System** cho hệ thống ABAC. Component này chịu trách nhiệm logging tất cả policy evaluations, security events, và compliance activities để đảm bảo transparency, accountability và regulatory compliance.

## 🎯 Trách Nhiệm Chính

1. **Evaluation Logging**: Log tất cả policy evaluation decisions
2. **Security Event Tracking**: Monitor và log security-related events
3. **Compliance Reporting**: Generate compliance reports cho auditors
4. **Performance Monitoring**: Track system performance metrics
5. **Forensic Analysis**: Support incident investigation
6. **Regulatory Compliance**: Meet audit requirements (SOX, GDPR, etc.)

## 📁 Cấu Trúc Files

```
audit/
├── logger.go          # AuditLogger implementation
└── logger_test.go     # Unit tests cho audit system
```

## 🏗️ Core Architecture

### AuditLogger Struct

```go
type AuditLogger struct {
    logFile *os.File      // Log file handle
    logger  *log.Logger   // Structured logger
}
```

**Design Characteristics:**
- **Structured Logging**: JSON format cho machine readability
- **File-Based**: Persistent storage trong log files
- **Thread-Safe**: Safe cho concurrent access
- **Configurable**: Support different output destinations

## 🔄 Audit Logging Flow

### 1. Logger Initialization

```go
func NewAuditLogger(logFilePath string) (*AuditLogger, error) {
    var logFile *os.File
    var err error
    
    if logFilePath != "" {
        logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return nil, fmt.Errorf("failed to open log file: %w", err)
        }
    } else {
        logFile = os.Stdout  // Default to stdout
    }
    
    logger := log.New(logFile, "", 0) // No default timestamp, we add our own
    
    return &AuditLogger{
        logFile: logFile,
        logger:  logger,
    }, nil
}
```

**Configuration Options:**
- **File Path**: Specify log file location
- **Stdout**: Log to console for development
- **Append Mode**: Preserve existing logs
- **Permissions**: Secure file permissions (0666)

### 2. Audit Log Structure

```go
type AuditLog struct {
    ID           int64                  `json:"id"`            // Unique log entry ID
    RequestID    string                 `json:"request_id"`    // Link to evaluation request
    SubjectID    string                 `json:"subject_id"`    // Who performed action
    ResourceID   string                 `json:"resource_id"`   // What resource was accessed
    ActionID     string                 `json:"action_id"`     // What action was attempted
    Decision     string                 `json:"decision"`      // PERMIT/DENY/NOT_APPLICABLE
    EvaluationMs int                    `json:"evaluation_ms"` // Performance metric
    Context      map[string]interface{} `json:"context"`       // Full context data
    CreatedAt    time.Time              `json:"created_at"`    // Timestamp
}
```

**Field Purposes:**
- **ID**: Unique identifier cho log correlation
- **RequestID**: Link multiple log entries to same request
- **Decision**: Core access control outcome
- **EvaluationMs**: Performance monitoring
- **Context**: Rich contextual information
- **CreatedAt**: Precise timestamp cho chronological ordering

## 📝 Logging Methods

### 1. Policy Evaluation Logging

```go
func (a *AuditLogger) LogEvaluation(request *models.EvaluationRequest, decision *models.Decision, context *models.EvaluationContext) error
```

**Implementation:**
```go
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
    
    // Add entity types
    if context.Subject != nil {
        auditContext["subject_type"] = context.Subject.SubjectType
    }
    if context.Resource != nil {
        auditContext["resource_type"] = context.Resource.ResourceType
    }
    if context.Action != nil {
        auditContext["action_category"] = context.Action.ActionCategory
    }
    
    auditEntry := models.AuditLog{
        RequestID:    request.RequestID,
        SubjectID:    request.SubjectID,
        ResourceID:   request.ResourceID,
        ActionID:     request.Action,
        Decision:     decision.Result,
        EvaluationMs: decision.EvaluationTimeMs,
        CreatedAt:    time.Now(),
        Context:      auditContext,
    }
    
    return a.logEntry(auditEntry)
}
```

**Sample Log Entry:**
```json
{
  "id": 1001,
  "request_id": "eval-001",
  "subject_id": "sub-001",
  "resource_id": "res-001",
  "action_id": "read",
  "decision": "permit",
  "evaluation_ms": 5,
  "context": {
    "matched_policies": ["pol-001", "pol-002"],
    "reason": "Access granted by matching permit policies",
    "source_ip": "10.0.1.50",
    "subject_type": "user",
    "resource_type": "api_endpoint",
    "action_category": "crud"
  },
  "created_at": "2024-01-15T14:30:00Z"
}
```

### 2. Security Event Logging

```go
func (a *AuditLogger) LogSecurityEvent(eventType string, subjectID string, details map[string]interface{}) error
```

**Security Event Types:**
- `probation_write_attempt`: User on probation attempting write
- `external_ip_access`: Access from external IP
- `privilege_escalation_attempt`: Unauthorized privilege escalation
- `after_hours_access`: Access outside business hours
- `failed_authentication`: Authentication failures
- `suspicious_activity`: Anomalous behavior patterns

**Implementation:**
```go
func (a *AuditLogger) LogSecurityEvent(eventType string, subjectID string, details map[string]interface{}) error {
    auditEntry := models.AuditLog{
        RequestID: fmt.Sprintf("security-%d", time.Now().UnixNano()),
        SubjectID: subjectID,
        Decision:  eventType,  // Reuse Decision field for event type
        CreatedAt: time.Now(),
        Context: map[string]interface{}{
            "event_type": eventType,
            "details":    details,
        },
    }
    
    return a.logEntry(auditEntry)
}
```

**Sample Security Event:**
```json
{
  "id": 1002,
  "request_id": "security-1705329000123456789",
  "subject_id": "sub-004",
  "decision": "probation_write_attempt",
  "context": {
    "event_type": "probation_write_attempt",
    "details": {
      "resource": "res-002",
      "action": "write",
      "blocked": true,
      "reason": "User on probation cannot perform write operations"
    }
  },
  "created_at": "2024-01-15T14:30:00Z"
}
```

### 3. Access Attempt Logging

```go
func (a *AuditLogger) LogAccessAttempt(request *models.EvaluationRequest, decision *models.Decision, additionalContext map[string]interface{}) error
```

**Use Cases:**
- Detailed access logging với custom context
- Integration với external systems
- Enhanced forensic information

**Implementation:**
```go
func (a *AuditLogger) LogAccessAttempt(request *models.EvaluationRequest, decision *models.Decision, additionalContext map[string]interface{}) error {
    auditEntry := models.AuditLog{
        RequestID:    request.RequestID,
        SubjectID:    request.SubjectID,
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
```

### 4. Policy Change Logging

```go
func (a *AuditLogger) LogPolicyChange(changeType string, policyID string, changedBy string, changes map[string]interface{}) error
```

**Change Types:**
- `policy_created`: New policy added
- `policy_updated`: Existing policy modified
- `policy_deleted`: Policy removed
- `policy_enabled`: Policy activated
- `policy_disabled`: Policy deactivated

**Implementation:**
```go
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
```

**Sample Policy Change:**
```json
{
  "request_id": "policy-change-1705329000123456789",
  "subject_id": "admin-001",
  "resource_id": "pol-001",
  "decision": "policy_updated",
  "context": {
    "change_type": "policy_updated",
    "changed_by": "admin-001",
    "changes": {
      "priority": {"old": 100, "new": 50},
      "enabled": {"old": false, "new": true}
    }
  },
  "created_at": "2024-01-15T14:30:00Z"
}
```

## 📊 Compliance & Reporting

### 1. Audit Statistics

```go
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
```

### 2. Compliance Report Structure

```go
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
```

### 3. Report Generation

```go
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
```

## 🔍 Audit Analysis Examples

### 1. Security Event Analysis

**Probation User Write Attempts:**
```json
{
  "event_type": "probation_write_attempt",
  "subject_id": "sub-004",
  "resource": "res-002",
  "action": "write",
  "blocked": true,
  "timestamp": "2024-01-15T14:30:00Z"
}
```

**External IP Access:**
```json
{
  "event_type": "external_ip_access",
  "subject_id": "sub-003",
  "source_ip": "203.0.113.1",
  "resource": "res-001",
  "blocked": true,
  "timestamp": "2024-01-15T22:00:00Z"
}
```

### 2. Performance Analysis

**Evaluation Time Tracking:**
```json
{
  "request_id": "eval-001",
  "evaluation_ms": 5,
  "matched_policies": ["pol-001", "pol-002"],
  "policy_count": 2,
  "rule_count": 4
}
```

**Performance Metrics:**
- Average evaluation time: 3.2ms
- P95 evaluation time: 8.5ms
- P99 evaluation time: 15.2ms
- Slowest evaluations: Complex regex rules

### 3. Access Pattern Analysis

**User Activity Patterns:**
```json
{
  "subject_id": "sub-001",
  "daily_accesses": 45,
  "resources_accessed": ["res-001", "res-002", "res-004"],
  "actions_performed": ["read", "write"],
  "peak_hours": ["09:00", "14:00", "16:00"]
}
```

**Resource Usage:**
```json
{
  "resource_id": "res-001",
  "total_accesses": 1250,
  "unique_users": 25,
  "most_common_action": "read",
  "peak_usage_time": "14:00-15:00"
}
```

## 🔒 Security & Privacy

### 1. Data Sanitization

```go
func sanitizeAuditContext(context map[string]interface{}) {
    // Remove sensitive data
    sensitiveKeys := []string{"password", "token", "secret", "key"}
    
    for _, key := range sensitiveKeys {
        delete(context, key)
    }
    
    // Mask PII data
    if email, ok := context["email"].(string); ok {
        context["email"] = maskEmail(email)
    }
    
    if ip, ok := context["source_ip"].(string); ok {
        context["source_ip"] = maskIP(ip)
    }
}

func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***@***.***"
    }
    
    username := parts[0]
    if len(username) > 2 {
        username = username[:2] + "***"
    }
    
    return username + "@" + parts[1]
}
```

### 2. Access Control

```go
type SecureAuditLogger struct {
    logger AuditLogger
    acl    AccessControlList
}

func (s *SecureAuditLogger) GetAuditTrail(requesterID, subjectID string) ([]*models.AuditLog, error) {
    // Check permissions
    if !s.acl.CanViewAuditLogs(requesterID, subjectID) {
        return nil, fmt.Errorf("access denied: insufficient permissions")
    }
    
    return s.logger.GetAuditTrail(subjectID, "", 100)
}
```

### 3. Log Integrity

```go
func (a *AuditLogger) logEntryWithIntegrity(entry models.AuditLog) error {
    // Add integrity hash
    entryJSON, _ := json.Marshal(entry)
    hash := sha256.Sum256(entryJSON)
    entry.Context["integrity_hash"] = hex.EncodeToString(hash[:])
    
    return a.logEntry(entry)
}
```

## 📈 Performance Optimization

### 1. Asynchronous Logging

```go
type AsyncAuditLogger struct {
    logger   *AuditLogger
    logChan  chan models.AuditLog
    done     chan bool
}

func NewAsyncAuditLogger(logger *AuditLogger) *AsyncAuditLogger {
    async := &AsyncAuditLogger{
        logger:  logger,
        logChan: make(chan models.AuditLog, 1000), // Buffer 1000 entries
        done:    make(chan bool),
    }
    
    go async.processLogs()
    return async
}

func (a *AsyncAuditLogger) processLogs() {
    for {
        select {
        case entry := <-a.logChan:
            a.logger.logEntry(entry)
        case <-a.done:
            return
        }
    }
}

func (a *AsyncAuditLogger) LogEvaluation(request *models.EvaluationRequest, decision *models.Decision, context *models.EvaluationContext) error {
    // Create audit entry
    entry := createAuditEntry(request, decision, context)
    
    // Send to channel (non-blocking)
    select {
    case a.logChan <- entry:
        return nil
    default:
        return fmt.Errorf("audit log buffer full")
    }
}
```

### 2. Batch Logging

```go
type BatchAuditLogger struct {
    logger    *AuditLogger
    batch     []models.AuditLog
    batchSize int
    mutex     sync.Mutex
}

func (b *BatchAuditLogger) LogEvaluation(request *models.EvaluationRequest, decision *models.Decision, context *models.EvaluationContext) error {
    entry := createAuditEntry(request, decision, context)
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    b.batch = append(b.batch, entry)
    
    if len(b.batch) >= b.batchSize {
        return b.flushBatch()
    }
    
    return nil
}

func (b *BatchAuditLogger) flushBatch() error {
    for _, entry := range b.batch {
        if err := b.logger.logEntry(entry); err != nil {
            return err
        }
    }
    
    b.batch = b.batch[:0] // Clear batch
    return nil
}
```

## 🧪 Testing Strategies

### Unit Tests
```go
func TestAuditLogging(t *testing.T) {
    // Create temp log file
    tmpFile, err := ioutil.TempFile("", "audit_test_*.log")
    assert.NoError(t, err)
    defer os.Remove(tmpFile.Name())
    
    logger, err := audit.NewAuditLogger(tmpFile.Name())
    assert.NoError(t, err)
    defer logger.Close()
    
    // Test evaluation logging
    request := &models.EvaluationRequest{
        RequestID:  "test-001",
        SubjectID:  "sub-001",
        ResourceID: "res-001",
        Action:     "read",
    }
    
    decision := &models.Decision{
        Result:           "permit",
        MatchedPolicies:  []string{"pol-001"},
        EvaluationTimeMs: 5,
        Reason:           "Test permit",
    }
    
    context := &models.EvaluationContext{}
    
    err = logger.LogEvaluation(request, decision, context)
    assert.NoError(t, err)
    
    // Verify log content
    content, err := ioutil.ReadFile(tmpFile.Name())
    assert.NoError(t, err)
    assert.Contains(t, string(content), "test-001")
    assert.Contains(t, string(content), "permit")
}
```

### Integration Tests
```go
func TestAuditIntegration(t *testing.T) {
    // Test complete audit flow
    storage, _ := storage.NewMockStorage(".") // Uses value-based storage for efficiency
    logger, _ := audit.NewAuditLogger("")
    pdp := evaluator.NewPolicyDecisionPoint(storage)
    
    request := &models.EvaluationRequest{
        RequestID:  "integration-001",
        SubjectID:  "sub-001",
        ResourceID: "res-001",
        Action:     "read",
    }
    
    // Perform evaluation
    decision, err := pdp.Evaluate(request)
    assert.NoError(t, err)
    
    // Log evaluation
    context, _ := pdp.EnrichContext(request)
    err = logger.LogEvaluation(request, decision, context)
    assert.NoError(t, err)
}
```

## 📊 Monitoring & Alerting

### Key Metrics
- **Log Volume**: Entries per second/minute/hour
- **Error Rate**: Failed logging attempts
- **Disk Usage**: Log file size growth
- **Performance**: Logging latency
- **Security Events**: Count of security violations

### Alerting Rules
- **High Deny Rate**: > 20% deny decisions
- **Security Events**: Any privilege escalation attempts
- **Performance**: Evaluation time > 100ms
- **System Health**: Logging failures > 1%

## 🎯 Best Practices

1. **Comprehensive Logging**: Log all access decisions
2. **Structured Format**: Use JSON cho machine readability
3. **Performance**: Implement asynchronous logging
4. **Security**: Sanitize sensitive data
5. **Retention**: Implement log rotation và archival
6. **Monitoring**: Set up alerts cho anomalies
7. **Compliance**: Meet regulatory requirements
8. **Testing**: Test logging functionality thoroughly

Package `audit` cung cấp enterprise-grade audit capabilities, đảm bảo full visibility và compliance cho ABAC system.
