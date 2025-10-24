package evaluator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"abac_go_example/models"
	"abac_go_example/storage"
)

// AuditLogger interface for audit logging
type AuditLogger interface {
	Info(message string, data map[string]interface{})
}

// SimpleAuditLogger is a basic implementation of AuditLogger
type SimpleAuditLogger struct{}

func NewSimpleAuditLogger() *SimpleAuditLogger {
	return &SimpleAuditLogger{}
}

func (sal *SimpleAuditLogger) Info(message string, data map[string]interface{}) {
	// Simple implementation - in production, this would integrate with proper logging
	fmt.Printf("[AUDIT] %s: %+v\n", message, data)
}

// PDPConfig represents configuration for the enhanced PDP
type PDPConfig struct {
	MaxEvaluationTime time.Duration `json:"max_evaluation_time"`
	EnableAudit       bool          `json:"enable_audit"`
}

// DefaultPDPConfig returns default configuration
func DefaultPDPConfig() *PDPConfig {
	return &PDPConfig{
		MaxEvaluationTime: 5 * time.Second,
		EnableAudit:       true,
	}
}

// EnhancedPDP implements the enhanced Policy Decision Point
type EnhancedPDP struct {
	// Core components
	storage storage.Storage

	// Evaluation engines
	conditionEvaluator       *EnhancedConditionEvaluator
	expressionEvaluator      *ExpressionEvaluator
	policyValidator          *PolicyValidator
	legacyConditionEvaluator *ConditionEvaluator

	// Infrastructure
	auditor AuditLogger

	// Configuration
	config *PDPConfig
}

// NewEnhancedPDP creates a new enhanced PDP instance
func NewEnhancedPDP(storage storage.Storage, config *PDPConfig) *EnhancedPDP {
	if config == nil {
		config = DefaultPDPConfig()
	}

	var auditor AuditLogger
	if config.EnableAudit {
		auditor = NewSimpleAuditLogger()
	}

	return &EnhancedPDP{
		storage:                  storage,
		conditionEvaluator:       NewEnhancedConditionEvaluator(),
		expressionEvaluator:      NewExpressionEvaluator(),
		policyValidator:          NewPolicyValidator(),
		legacyConditionEvaluator: NewConditionEvaluator(),
		auditor:                  auditor,
		config:                   config,
	}
}

// Evaluate performs enhanced policy evaluation with context support
func (pdp *EnhancedPDP) Evaluate(ctx context.Context, req *models.DecisionRequest) (*models.DecisionResponse, error) {
	start := time.Now()

	// Set evaluation timeout
	ctx, cancel := context.WithTimeout(ctx, pdp.config.MaxEvaluationTime)
	defer cancel()

	// Validate request
	if err := pdp.validateRequest(req); err != nil {
		return pdp.createErrorResponse(models.DecisionIndeterminate, err.Error(), req.RequestID), nil
	}

	// Get applicable policies
	policies, err := pdp.GetApplicablePolicies(ctx, req)
	if err != nil {
		return pdp.createErrorResponse(models.DecisionIndeterminate, err.Error(), req.RequestID), nil
	}

	// Evaluate policies with priority (DENY > PERMIT > NOT_APPLICABLE)
	response := pdp.evaluatePoliciesWithPriority(ctx, policies, req)
	response.Duration = time.Since(start)
	response.EvaluatedAt = time.Now()
	response.RequestID = req.RequestID

	// Audit decision
	if pdp.config.EnableAudit && pdp.auditor != nil {
		pdp.auditDecision(ctx, req, response)
	}

	return response, nil
}

// ValidatePolicy validates a policy against schema and business rules
func (pdp *EnhancedPDP) ValidatePolicy(policy *models.Policy) error {
	return pdp.policyValidator.ValidatePolicy(policy)
}

// GetApplicablePolicies returns policies that might apply to the request
func (pdp *EnhancedPDP) GetApplicablePolicies(ctx context.Context, req *models.DecisionRequest) ([]*models.Policy, error) {
	// Get all policies from storage
	allPolicies, err := pdp.storage.GetPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve policies: %w", err)
	}

	var applicablePolicies []*models.Policy

	// Filter policies based on basic criteria
	for _, policy := range allPolicies {
		if !policy.Enabled {
			continue
		}

		// Check if policy might apply (basic filtering)
		if pdp.isPolicyPotentiallyApplicable(policy, req) {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	return applicablePolicies, nil
}

// HealthCheck performs a health check on the PDP
func (pdp *EnhancedPDP) HealthCheck(ctx context.Context) error {
	// Check storage connectivity
	_, err := pdp.storage.GetPolicies()
	if err != nil {
		return fmt.Errorf("storage health check failed: %w", err)
	}

	return nil
}

// evaluatePoliciesWithPriority evaluates policies with proper ABAC priority handling
func (pdp *EnhancedPDP) evaluatePoliciesWithPriority(ctx context.Context, policies []*models.Policy, req *models.DecisionRequest) *models.DecisionResponse {
	var denyReasons []string
	var permitReasons []string
	var applicablePolicies []string

	// Build evaluation context
	evalContext := pdp.buildEvaluationContext(req)

	for _, policy := range policies {
		for _, statement := range policy.Statement {
			// Check if statement applies to this request
			if !pdp.isStatementApplicable(statement, evalContext) {
				continue
			}

			applicablePolicies = append(applicablePolicies, policy.ID)

			// Evaluate conditions
			conditionResult := pdp.evaluateStatementConditions(ctx, statement, req, evalContext)
			if !conditionResult {
				continue
			}

			// Statement matches - check effect
			switch strings.ToLower(statement.Effect) {
			case "deny":
				denyReasons = append(denyReasons, fmt.Sprintf("Denied by policy %s (statement %s)", policy.ID, statement.Sid))
				// DENY has highest priority - return immediately
				return &models.DecisionResponse{
					Decision: models.DecisionDeny,
					Reason:   strings.Join(denyReasons, "; "),
					Policies: applicablePolicies,
				}
			case "allow":
				permitReasons = append(permitReasons, fmt.Sprintf("Permitted by policy %s (statement %s)", policy.ID, statement.Sid))
			}
		}
	}

	// If we have PERMIT decisions and no DENY, return PERMIT
	if len(permitReasons) > 0 {
		return &models.DecisionResponse{
			Decision: models.DecisionPermit,
			Reason:   strings.Join(permitReasons, "; "),
			Policies: applicablePolicies,
		}
	}

	// No applicable policies found
	return &models.DecisionResponse{
		Decision: models.DecisionNotApplicable,
		Reason:   "No applicable policies found",
		Policies: applicablePolicies,
	}
}

// validateRequest validates the decision request
func (pdp *EnhancedPDP) validateRequest(req *models.DecisionRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Subject == nil {
		return fmt.Errorf("subject is required")
	}

	if req.Resource == nil {
		return fmt.Errorf("resource is required")
	}

	if req.Action == nil {
		return fmt.Errorf("action is required")
	}

	return nil
}

// buildEvaluationContext builds the evaluation context from the request
func (pdp *EnhancedPDP) buildEvaluationContext(req *models.DecisionRequest) map[string]interface{} {
	evalContext := make(map[string]interface{})

	// Add subject attributes
	if req.Subject != nil {
		evalContext["request:UserId"] = req.Subject.ID
		evalContext["user:SubjectType"] = req.Subject.SubjectType
		for key, value := range req.Subject.Attributes {
			evalContext["user:"+key] = value
		}
	}

	// Add resource attributes
	if req.Resource != nil {
		evalContext["request:ResourceId"] = req.Resource.ID
		evalContext["resource:ResourceType"] = req.Resource.ResourceType
		for key, value := range req.Resource.Attributes {
			evalContext["resource:"+key] = value
		}
	}

	// Add action attributes
	if req.Action != nil {
		evalContext["request:Action"] = req.Action.ActionName
		evalContext["action:ActionCategory"] = req.Action.ActionCategory
	}

	// Add environmental attributes
	if req.Environment != nil {
		evalContext["request:Time"] = req.Environment.Timestamp.Format(time.RFC3339)
		evalContext["environment:client_ip"] = req.Environment.ClientIP
		evalContext["environment:user_agent"] = req.Environment.UserAgent

		if req.Environment.Location != nil {
			evalContext["environment:country"] = req.Environment.Location.Country
			evalContext["environment:region"] = req.Environment.Location.Region
		}

		for key, value := range req.Environment.Attributes {
			evalContext["environment:"+key] = value
		}
	}

	// Add custom context
	for key, value := range req.Context {
		evalContext["request:"+key] = value
	}

	return evalContext
}

// isStatementApplicable checks if a statement applies to the request
func (pdp *EnhancedPDP) isStatementApplicable(statement models.PolicyStatement, context map[string]interface{}) bool {
	// Check action matching
	if !pdp.matchAction(statement.Action, context) {
		return false
	}

	// Check resource matching
	if !pdp.matchResource(statement, context) {
		return false
	}

	return true
}

// evaluateStatementConditions evaluates all conditions for a statement
func (pdp *EnhancedPDP) evaluateStatementConditions(ctx context.Context, statement models.PolicyStatement, req *models.DecisionRequest, evalContext map[string]interface{}) bool {
	if len(statement.Condition) == 0 {
		return true
	}

	// Substitute variables in conditions
	expandedConditions := pdp.legacyConditionEvaluator.SubstituteVariables(statement.Condition, evalContext)

	// Evaluate using enhanced condition evaluator with environmental context
	return pdp.conditionEvaluator.EvaluateConditions(expandedConditions, evalContext)
}

// matchAction checks if the requested action matches statement action(s)
func (pdp *EnhancedPDP) matchAction(actionSpec models.JSONActionResource, context map[string]interface{}) bool {
	requestedAction, ok := context["request:Action"].(string)
	if !ok {
		return false
	}

	actionValues := actionSpec.GetValues()
	for _, actionPattern := range actionValues {
		if pdp.matchPattern(actionPattern, requestedAction) {
			return true
		}
	}
	return false
}

// matchResource checks if the requested resource matches statement resource(s)
func (pdp *EnhancedPDP) matchResource(statement models.PolicyStatement, context map[string]interface{}) bool {
	requestedResource, ok := context["request:ResourceId"].(string)
	if !ok {
		return false
	}

	// Check if resource matches Resource patterns
	resourceMatches := false
	resourceValues := statement.Resource.GetValues()
	for _, resourcePattern := range resourceValues {
		if pdp.matchPattern(resourcePattern, requestedResource) {
			resourceMatches = true
			break
		}
	}

	if !resourceMatches {
		return false
	}

	// Check NotResource exclusions (if specified)
	if statement.NotResource.IsArray || statement.NotResource.Single != "" {
		notResourceValues := statement.NotResource.GetValues()
		for _, notResourcePattern := range notResourceValues {
			if pdp.matchPattern(notResourcePattern, requestedResource) {
				return false // Excluded by NotResource
			}
		}
	}

	return true
}

// matchPattern performs pattern matching with wildcard support
func (pdp *EnhancedPDP) matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		return pdp.matchWildcard(pattern, value)
	}

	return pattern == value
}

// matchWildcard performs wildcard pattern matching
func (pdp *EnhancedPDP) matchWildcard(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// *middle* - contains
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(value, middle)
	} else if strings.HasPrefix(pattern, "*") {
		// *suffix - ends with
		suffix := pattern[1:]
		return strings.HasSuffix(value, suffix)
	} else if strings.HasSuffix(pattern, "*") {
		// prefix* - starts with
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(value, prefix)
	}

	return pattern == value
}

// isPolicyPotentiallyApplicable performs basic filtering to determine if a policy might apply
func (pdp *EnhancedPDP) isPolicyPotentiallyApplicable(policy *models.Policy, req *models.DecisionRequest) bool {
	// This is a basic implementation - could be enhanced with more sophisticated filtering
	return true
}

// createErrorResponse creates an error response
func (pdp *EnhancedPDP) createErrorResponse(decision models.DecisionType, reason, requestID string) *models.DecisionResponse {
	return &models.DecisionResponse{
		Decision:    decision,
		Reason:      reason,
		Policies:    []string{},
		EvaluatedAt: time.Now(),
		RequestID:   requestID,
	}
}

// auditDecision logs the decision for audit purposes
func (pdp *EnhancedPDP) auditDecision(ctx context.Context, req *models.DecisionRequest, response *models.DecisionResponse) {
	if pdp.auditor == nil {
		return
	}

	auditData := map[string]interface{}{
		"request_id":      req.RequestID,
		"subject_id":      getSubjectID(req.Subject),
		"resource_id":     getResourceID(req.Resource),
		"action":          getActionNameForAudit(req.Action),
		"decision":        string(response.Decision),
		"reason":          response.Reason,
		"policies":        response.Policies,
		"evaluation_time": response.Duration.Milliseconds(),
		"timestamp":       response.EvaluatedAt,
	}

	pdp.auditor.Info("Policy decision made", auditData)
}

// Helper functions
func getSubjectID(subject *models.Subject) string {
	if subject == nil {
		return ""
	}
	return subject.ID
}

func getResourceID(resource *models.Resource) string {
	if resource == nil {
		return ""
	}
	return resource.ID
}

func getActionNameForAudit(action *models.Action) string {
	if action == nil {
		return ""
	}
	return action.ActionName
}
