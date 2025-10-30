package pep

import (
	"context"
	"fmt"
	"time"

	"abac_go_example/evaluator/core"
	"abac_go_example/models"
)

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogDecision(data map[string]interface{})
}

// SimplePolicyEnforcementPoint is a simplified version of PEP without advanced features
type SimplePolicyEnforcementPoint struct {
	pdp         core.PolicyDecisionPointInterface
	auditLogger AuditLogger
	config      *PEPConfig
	metrics     *SimplePEPMetrics
}

// SimplePEPMetrics holds basic metrics for the simple PEP
type SimplePEPMetrics struct {
	TotalRequests    int64 `json:"total_requests"`
	PermitDecisions  int64 `json:"permit_decisions"`
	DenyDecisions    int64 `json:"deny_decisions"`
	ValidationErrors int64 `json:"validation_errors"`
	EvaluationErrors int64 `json:"evaluation_errors"`
}

// NewSimplePolicyEnforcementPoint creates a new simplified PEP instance
func NewSimplePolicyEnforcementPoint(pdp core.PolicyDecisionPointInterface, auditLogger AuditLogger, config *PEPConfig) *SimplePolicyEnforcementPoint {
	if config == nil {
		config = &PEPConfig{
			FailSafeMode:      true,
			StrictValidation:  true,
			AuditEnabled:      true,
			EvaluationTimeout: time.Millisecond * 100,
		}
	}

	return &SimplePolicyEnforcementPoint{
		pdp:         pdp,
		auditLogger: auditLogger,
		config:      config,
		metrics:     &SimplePEPMetrics{},
	}
}

// EnforceRequest is the main enforcement method for simplified PEP
func (spep *SimplePolicyEnforcementPoint) EnforceRequest(ctx context.Context, request *models.EvaluationRequest) (*EnforcementResult, error) {
	startTime := time.Now()

	// Update metrics
	spep.metrics.TotalRequests++

	// Input validation
	if err := spep.validateRequest(request); err != nil {
		spep.metrics.ValidationErrors++
		return spep.createDenyResult("Invalid request: "+err.Error(), startTime), nil
	}

	// Create context with timeout
	evalCtx, cancel := context.WithTimeout(ctx, spep.config.EvaluationTimeout)
	defer cancel()

	// Perform policy evaluation
	decision, err := spep.evaluateWithTimeout(evalCtx, request)
	if err != nil {
		spep.metrics.EvaluationErrors++

		// Fail-safe mode: deny on error
		if spep.config.FailSafeMode {
			result := spep.createDenyResult("Evaluation error: "+err.Error(), startTime)
			spep.auditDecision(request, result)
			return result, nil
		}
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	// Create enforcement result
	result := &EnforcementResult{
		Decision:         decision.Result,
		Allowed:          decision.Result == "permit",
		Reason:           decision.Reason,
		MatchedPolicies:  decision.MatchedPolicies,
		EvaluationTimeMs: int(time.Since(startTime).Milliseconds()),
		CacheHit:         false,
		Timestamp:        time.Now(),
	}

	// Update metrics based on decision
	switch decision.Result {
	case "permit":
		spep.metrics.PermitDecisions++
	case "deny":
		spep.metrics.DenyDecisions++
	}

	// Audit logging
	if spep.config.AuditEnabled {
		spep.auditDecision(request, result)
	}

	return result, nil
}

// evaluateWithTimeout performs evaluation with timeout
func (spep *SimplePolicyEnforcementPoint) evaluateWithTimeout(ctx context.Context, request *models.EvaluationRequest) (*models.Decision, error) {
	resultChan := make(chan *models.Decision, 1)
	errorChan := make(chan error, 1)

	go func() {
		decision, err := spep.pdp.Evaluate(request)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- decision
	}()

	select {
	case decision := <-resultChan:
		return decision, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("evaluation timeout: %w", ctx.Err())
	}
}

// validateRequest validates the evaluation request
func (spep *SimplePolicyEnforcementPoint) validateRequest(request *models.EvaluationRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.Subject == nil {
		return fmt.Errorf("subject is required")
	}

	if request.Subject.GetID() == "" {
		return fmt.Errorf("subject ID is required")
	}

	if request.ResourceID == "" {
		return fmt.Errorf("resource_id is required")
	}

	if request.Action == "" {
		return fmt.Errorf("action is required")
	}

	// Additional strict validation if enabled
	if spep.config.StrictValidation {
		if len(request.Subject.GetID()) > 255 {
			return fmt.Errorf("subject_id too long (max 255 characters)")
		}
		if len(request.ResourceID) > 255 {
			return fmt.Errorf("resource_id too long (max 255 characters)")
		}
		if len(request.Action) > 100 {
			return fmt.Errorf("action too long (max 100 characters)")
		}
	}

	return nil
}

// createDenyResult creates a deny result for error cases
func (spep *SimplePolicyEnforcementPoint) createDenyResult(reason string, startTime time.Time) *EnforcementResult {
	return &EnforcementResult{
		Decision:         "deny",
		Allowed:          false,
		Reason:           reason,
		MatchedPolicies:  []string{},
		EvaluationTimeMs: int(time.Since(startTime).Milliseconds()),
		CacheHit:         false,
		Timestamp:        time.Now(),
	}
}

// auditDecision logs the decision for audit purposes
func (spep *SimplePolicyEnforcementPoint) auditDecision(request *models.EvaluationRequest, result *EnforcementResult) {
	if spep.auditLogger == nil {
		return
	}

	// Get subject ID from Subject interface
	subjectID := ""
	if request.Subject != nil {
		subjectID = request.Subject.GetID()
	}

	auditData := map[string]interface{}{
		"request_id":       request.RequestID,
		"subject_id":       subjectID,
		"resource_id":      request.ResourceID,
		"action":           request.Action,
		"decision":         result.Decision,
		"allowed":          result.Allowed,
		"evaluation_ms":    result.EvaluationTimeMs,
		"matched_policies": result.MatchedPolicies,
		"context":          request.Context,
	}

	spep.auditLogger.LogDecision(auditData)
}

// GetMetrics returns current simple PEP metrics
func (spep *SimplePolicyEnforcementPoint) GetMetrics() *SimplePEPMetrics {
	return spep.metrics
}

// GetConfig returns current PEP configuration
func (spep *SimplePolicyEnforcementPoint) GetConfig() *PEPConfig {
	return spep.config
}

// Reset resets metrics
func (spep *SimplePolicyEnforcementPoint) Reset() {
	spep.metrics = &SimplePEPMetrics{}
}
