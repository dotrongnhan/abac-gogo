package pep

import (
	"context"
	"fmt"
	"sync"
	"time"

	"abac_go_example/evaluator"
	"abac_go_example/models"
)

// PolicyEnforcementPoint (PEP) is the main enforcement engine
type PolicyEnforcementPoint struct {
	pdp            *evaluator.PolicyDecisionPoint
	auditLogger    AuditLogger
	cache          *DecisionCache
	config         *PEPConfig
	metrics        *PEPMetrics
	rateLimiter    *RateLimiter
	mu             sync.RWMutex
	circuitBreaker *CircuitBreaker
}

// PEPConfig holds configuration for the PEP
type PEPConfig struct {
	// Performance settings
	CacheEnabled      bool          `json:"cache_enabled"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	CacheMaxSize      int           `json:"cache_max_size"`
	EvaluationTimeout time.Duration `json:"evaluation_timeout"`

	// Security settings
	FailSafeMode     bool `json:"fail_safe_mode"`    // Default to DENY on errors
	StrictValidation bool `json:"strict_validation"` // Strict input validation
	AuditEnabled     bool `json:"audit_enabled"`     // Enable audit logging

	// Rate limiting
	RateLimitEnabled  bool `json:"rate_limit_enabled"`
	RequestsPerSecond int  `json:"requests_per_second"`
	BurstSize         int  `json:"burst_size"`

	// Circuit breaker
	CircuitBreakerEnabled bool          `json:"circuit_breaker_enabled"`
	FailureThreshold      int           `json:"failure_threshold"`
	RecoveryTimeout       time.Duration `json:"recovery_timeout"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
}

// DefaultPEPConfig returns default configuration
func DefaultPEPConfig() *PEPConfig {
	return &PEPConfig{
		CacheEnabled:          true,
		CacheTTL:              time.Minute * 5,
		CacheMaxSize:          10000,
		EvaluationTimeout:     time.Millisecond * 100,
		FailSafeMode:          true,
		StrictValidation:      true,
		AuditEnabled:          true,
		RateLimitEnabled:      true,
		RequestsPerSecond:     1000,
		BurstSize:             100,
		CircuitBreakerEnabled: true,
		FailureThreshold:      10,
		RecoveryTimeout:       time.Second * 30,
		MaxConcurrentRequests: 100,
	}
}

// NewPolicyEnforcementPoint creates a new PEP instance
func NewPolicyEnforcementPoint(pdp *evaluator.PolicyDecisionPoint, auditLogger AuditLogger, config *PEPConfig) *PolicyEnforcementPoint {
	if config == nil {
		config = DefaultPEPConfig()
	}

	pep := &PolicyEnforcementPoint{
		pdp:         pdp,
		auditLogger: auditLogger,
		config:      config,
		metrics:     NewPEPMetrics(),
	}

	// Initialize cache if enabled
	if config.CacheEnabled {
		pep.cache = NewDecisionCache(config.CacheMaxSize, config.CacheTTL)
	}

	// Initialize rate limiter if enabled
	if config.RateLimitEnabled {
		pep.rateLimiter = NewRateLimiter(config.RequestsPerSecond, config.BurstSize)
	}

	// Initialize circuit breaker if enabled
	if config.CircuitBreakerEnabled {
		pep.circuitBreaker = NewCircuitBreaker(config.FailureThreshold, config.RecoveryTimeout, config.MaxConcurrentRequests)
	}

	return pep
}

// EnforceRequest is the main enforcement method
func (pep *PolicyEnforcementPoint) EnforceRequest(ctx context.Context, request *models.EvaluationRequest) (*EnforcementResult, error) {
	startTime := time.Now()

	// Update metrics
	pep.metrics.IncrementTotalRequests()

	// Input validation
	if err := pep.validateRequest(request); err != nil {
		pep.metrics.IncrementValidationErrors()
		return pep.createDenyResult("Invalid request: "+err.Error(), startTime), nil
	}

	// Rate limiting check
	if pep.config.RateLimitEnabled && !pep.rateLimiter.Allow() {
		pep.metrics.IncrementRateLimitExceeded()
		return pep.createDenyResult("Rate limit exceeded", startTime), nil
	}

	// Circuit breaker check
	if pep.config.CircuitBreakerEnabled && !pep.circuitBreaker.Allow() {
		pep.metrics.IncrementCircuitBreakerOpen()
		return pep.createDenyResult("Circuit breaker open", startTime), nil
	}

	// Check cache first
	if pep.config.CacheEnabled {
		if cachedResult := pep.cache.Get(request); cachedResult != nil {
			pep.metrics.IncrementCacheHits()
			cachedResult.CacheHit = true
			cachedResult.EvaluationTimeMs = int(time.Since(startTime).Milliseconds())
			return cachedResult, nil
		}
		pep.metrics.IncrementCacheMisses()
	}

	// Create context with timeout
	evalCtx, cancel := context.WithTimeout(ctx, pep.config.EvaluationTimeout)
	defer cancel()

	// Perform policy evaluation
	decision, err := pep.evaluateWithTimeout(evalCtx, request)
	if err != nil {
		pep.metrics.IncrementEvaluationErrors()

		// Circuit breaker: record failure
		if pep.config.CircuitBreakerEnabled {
			pep.circuitBreaker.RecordFailure()
		}

		// Fail-safe mode: deny on error
		if pep.config.FailSafeMode {
			result := pep.createDenyResult("Evaluation error: "+err.Error(), startTime)
			pep.auditDecision(request, result)
			return result, nil
		}
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	// Circuit breaker: record success
	if pep.config.CircuitBreakerEnabled {
		pep.circuitBreaker.RecordSuccess()
	}

	// Create enforcement result
	result := &EnforcementResult{
		Decision:         decision,
		Allowed:          decision.Result == "permit",
		Reason:           decision.Reason,
		EvaluationTimeMs: int(time.Since(startTime).Milliseconds()),
		CacheHit:         false,
		Timestamp:        time.Now(),
	}

	// Cache the result if enabled
	if pep.config.CacheEnabled && decision.Result != "not_applicable" {
		pep.cache.Set(request, result)
	}

	// Update metrics based on decision
	switch decision.Result {
	case "permit":
		pep.metrics.IncrementPermitDecisions()
	case "deny":
		pep.metrics.IncrementDenyDecisions()
	case "not_applicable":
		pep.metrics.IncrementNotApplicableDecisions()
	}

	// Audit logging
	if pep.config.AuditEnabled {
		pep.auditDecision(request, result)
	}

	return result, nil
}

// EnforceRequestBatch processes multiple requests efficiently
func (pep *PolicyEnforcementPoint) EnforceRequestBatch(ctx context.Context, requests []*models.EvaluationRequest) ([]*EnforcementResult, error) {
	results := make([]*EnforcementResult, len(requests))
	errors := make([]error, len(requests))

	// Use worker pool for concurrent processing
	const maxWorkers = 10
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i, request := range requests {
		wg.Add(1)
		go func(idx int, req *models.EvaluationRequest) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			result, err := pep.EnforceRequest(ctx, req)
			results[idx] = result
			errors[idx] = err
		}(i, request)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// evaluateWithTimeout performs evaluation with timeout
func (pep *PolicyEnforcementPoint) evaluateWithTimeout(ctx context.Context, request *models.EvaluationRequest) (*models.Decision, error) {
	resultChan := make(chan *models.Decision, 1)
	errorChan := make(chan error, 1)

	go func() {
		decision, err := pep.pdp.Evaluate(request)
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
func (pep *PolicyEnforcementPoint) validateRequest(request *models.EvaluationRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.SubjectID == "" {
		return fmt.Errorf("subject_id is required")
	}

	if request.ResourceID == "" {
		return fmt.Errorf("resource_id is required")
	}

	if request.Action == "" {
		return fmt.Errorf("action is required")
	}

	// Additional strict validation if enabled
	if pep.config.StrictValidation {
		if len(request.SubjectID) > 255 {
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
func (pep *PolicyEnforcementPoint) createDenyResult(reason string, startTime time.Time) *EnforcementResult {
	return &EnforcementResult{
		Decision: &models.Decision{
			Result:           "deny",
			MatchedPolicies:  []string{},
			EvaluationTimeMs: int(time.Since(startTime).Milliseconds()),
			Reason:           reason,
		},
		Allowed:          false,
		Reason:           reason,
		EvaluationTimeMs: int(time.Since(startTime).Milliseconds()),
		CacheHit:         false,
		Timestamp:        time.Now(),
	}
}

// auditDecision logs the decision for audit purposes
func (pep *PolicyEnforcementPoint) auditDecision(request *models.EvaluationRequest, result *EnforcementResult) {
	if pep.auditLogger == nil {
		return
	}

	auditData := map[string]interface{}{
		"request_id":       request.RequestID,
		"subject_id":       request.SubjectID,
		"resource_id":      request.ResourceID,
		"action":           request.Action,
		"decision":         result.Decision.Result,
		"allowed":          result.Allowed,
		"evaluation_ms":    result.EvaluationTimeMs,
		"cache_hit":        result.CacheHit,
		"matched_policies": result.Decision.MatchedPolicies,
		"context":          request.Context,
	}

	pep.auditLogger.LogDecision(auditData)
}

// GetMetrics returns current PEP metrics
func (pep *PolicyEnforcementPoint) GetMetrics() *PEPMetrics {
	return pep.metrics
}

// GetConfig returns current PEP configuration
func (pep *PolicyEnforcementPoint) GetConfig() *PEPConfig {
	pep.mu.RLock()
	defer pep.mu.RUnlock()
	return pep.config
}

// UpdateConfig updates PEP configuration (thread-safe)
func (pep *PolicyEnforcementPoint) UpdateConfig(config *PEPConfig) {
	pep.mu.Lock()
	defer pep.mu.Unlock()
	pep.config = config
}

// Shutdown gracefully shuts down the PEP
func (pep *PolicyEnforcementPoint) Shutdown(ctx context.Context) error {
	// Stop accepting new requests
	if pep.circuitBreaker != nil {
		pep.circuitBreaker.ForceOpen()
	}

	// Wait for ongoing requests to complete or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 5):
		// Force shutdown after 5 seconds
	}

	return nil
}

// EnforcementResult represents the result of policy enforcement
type EnforcementResult struct {
	Decision         *models.Decision `json:"decision"`
	Allowed          bool             `json:"allowed"`
	Reason           string           `json:"reason"`
	EvaluationTimeMs int              `json:"evaluation_time_ms"`
	CacheHit         bool             `json:"cache_hit"`
	Timestamp        time.Time        `json:"timestamp"`
}
