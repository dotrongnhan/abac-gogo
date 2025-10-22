package evaluator

import (
	"fmt"
	"strings"
	"time"

	"abac_go_example/attributes"
	"abac_go_example/models"
	"abac_go_example/storage"
)

// Constants for policy effects
const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// Constants for decision results
const (
	ResultPermit = "permit"
	ResultDeny   = "deny"
)

// Constants for decision reasons
const (
	ReasonDeniedByStatement   = "Denied by statement: %s"
	ReasonAllowedByStatements = "Allowed by statements: %s"
	ReasonImplicitDeny        = "No matching policies found (implicit deny)"
)

// Constants for context keys
const (
	ContextKeyRequestUserID     = "request:UserId"
	ContextKeyRequestAction     = "request:Action"
	ContextKeyRequestResourceID = "request:ResourceId"
	ContextKeyRequestTime       = "request:Time"
	ContextKeyUserPrefix        = "user:"
	ContextKeyResourcePrefix    = "resource:"
	ContextKeyEnvironmentPrefix = "environment:"
	ContextKeyRequestPrefix     = "request:"
)

// PolicyDecisionPointInterface defines the interface for policy evaluation
type PolicyDecisionPointInterface interface {
	Evaluate(request *models.EvaluationRequest) (*models.Decision, error)
}

// PolicyDecisionPoint (PDP) is the main evaluation engine
type PolicyDecisionPoint struct {
	storage            storage.Storage
	attributeResolver  *attributes.AttributeResolver
	actionMatcher      *ActionMatcher
	resourceMatcher    *ResourceMatcher
	conditionEvaluator *ConditionEvaluator
}

// NewPolicyDecisionPoint creates a new PDP instance and returns the interface
func NewPolicyDecisionPoint(storage storage.Storage) PolicyDecisionPointInterface {
	return &PolicyDecisionPoint{
		storage:            storage,
		attributeResolver:  attributes.NewAttributeResolver(storage),
		actionMatcher:      NewActionMatcher(),
		resourceMatcher:    NewResourceMatcher(),
		conditionEvaluator: NewConditionEvaluator(),
	}
}

// Evaluate performs optimized policy evaluation for a given request
// This unified method combines the best practices from both legacy approaches
func (pdp *PolicyDecisionPoint) Evaluate(request *models.EvaluationRequest) (*models.Decision, error) {
	startTime := time.Now()

	// Input validation
	if request == nil {
		return nil, fmt.Errorf("evaluation request cannot be nil")
	}
	if request.SubjectID == "" || request.ResourceID == "" || request.Action == "" {
		return nil, fmt.Errorf("invalid request: missing required fields (SubjectID, ResourceID, Action)")
	}

	// Step 1: Enrich context with all necessary attributes
	context, err := pdp.attributeResolver.EnrichContext(request)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich context: %w", err)
	}

	// Step 2: Get all policies
	allPolicies, err := pdp.storage.GetPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	// Step 3: Build evaluation context using original request (preserves all context)
	// This approach maintains data integrity from EvaluateNew while adding validation
	evalContext := pdp.buildEvaluationContext(request, context)

	// Step 4: Evaluate policies with Deny-Override algorithm
	decision := pdp.evaluateNewPolicies(allPolicies, evalContext)

	// Step 5: Calculate evaluation time
	evaluationTime := int(time.Since(startTime).Milliseconds())
	decision.EvaluationTimeMs = evaluationTime

	return decision, nil
}

// buildEvaluationContext builds context map for new policy format
func (pdp *PolicyDecisionPoint) buildEvaluationContext(request *models.EvaluationRequest, context *models.EvaluationContext) map[string]interface{} {
	evalContext := make(map[string]interface{})

	// Request context
	evalContext[ContextKeyRequestUserID] = request.SubjectID
	evalContext[ContextKeyRequestAction] = request.Action
	evalContext[ContextKeyRequestResourceID] = request.ResourceID
	evalContext[ContextKeyRequestTime] = context.Timestamp.Format(time.RFC3339)

	// Add custom context from request
	for key, value := range request.Context {
		evalContext[ContextKeyRequestPrefix+key] = value
	}

	// Subject attributes
	if context.Subject != nil {
		for key, value := range context.Subject.Attributes {
			evalContext[ContextKeyUserPrefix+key] = value
		}
		evalContext[ContextKeyUserPrefix+"SubjectType"] = context.Subject.SubjectType
	}

	// Resource attributes
	if context.Resource != nil {
		for key, value := range context.Resource.Attributes {
			evalContext[ContextKeyResourcePrefix+key] = value
		}
		evalContext[ContextKeyResourcePrefix+"ResourceType"] = context.Resource.ResourceType
		evalContext[ContextKeyResourcePrefix+"ResourceId"] = context.Resource.ResourceID
	}

	// Environment attributes
	for key, value := range context.Environment {
		evalContext[ContextKeyEnvironmentPrefix+key] = value
	}

	return evalContext
}

// evaluateNewPolicies evaluates policies using the new format with Deny-Override
func (pdp *PolicyDecisionPoint) evaluateNewPolicies(policies []*models.Policy, context map[string]interface{}) *models.Decision {
	var matchedPolicies []string
	var matchedStatements []string

	// Step 1: Collect all matching statements
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		for _, statement := range policy.Statement {
			if pdp.evaluateStatement(statement, context) {
				matchedPolicies = append(matchedPolicies, policy.ID)
				if statement.Sid != "" {
					matchedStatements = append(matchedStatements, statement.Sid)
				}

				// Step 2: Apply Deny-Override - if any statement denies, return deny immediately
				if strings.ToLower(statement.Effect) == EffectDeny {
					return &models.Decision{
						Result:          ResultDeny,
						MatchedPolicies: matchedPolicies,
						Reason:          fmt.Sprintf(ReasonDeniedByStatement, statement.Sid),
					}
				}
			}
		}
	}

	// Step 3: If we have any Allow statements, return allow
	if len(matchedStatements) > 0 {
		return &models.Decision{
			Result:          ResultPermit,
			MatchedPolicies: matchedPolicies,
			Reason:          fmt.Sprintf(ReasonAllowedByStatements, strings.Join(matchedStatements, ", ")),
		}
	}

	// Step 4: Default deny (no matching policies)
	return &models.Decision{
		Result:          ResultDeny,
		MatchedPolicies: []string{},
		Reason:          ReasonImplicitDeny,
	}
}

// evaluateStatement evaluates a single policy statement
func (pdp *PolicyDecisionPoint) evaluateStatement(statement models.PolicyStatement, context map[string]interface{}) bool {
	// Step 1: Check action matching
	if !pdp.matchAction(statement.Action, context) {
		return false
	}

	// Step 2: Check resource matching (including NotResource exclusions)
	if !pdp.matchResource(statement, context) {
		return false
	}

	// Step 3: Check conditions
	if len(statement.Condition) > 0 {
		// Substitute variables in conditions
		expandedConditions := pdp.conditionEvaluator.SubstituteVariables(statement.Condition, context)
		if !pdp.conditionEvaluator.Evaluate(expandedConditions, context) {
			return false
		}
	}

	return true
}

// matchAction checks if the requested action matches statement action(s)
func (pdp *PolicyDecisionPoint) matchAction(actionSpec models.JSONActionResource, context map[string]interface{}) bool {
	requestedAction, ok := context[ContextKeyRequestAction].(string)
	if !ok {
		return false
	}

	actionValues := actionSpec.GetValues()
	for _, actionPattern := range actionValues {
		if pdp.actionMatcher.Match(actionPattern, requestedAction) {
			return true
		}
	}
	return false
}

// matchResource checks if the requested resource matches statement resource(s)
// and does not match NotResource patterns (if specified)
func (pdp *PolicyDecisionPoint) matchResource(statement models.PolicyStatement, context map[string]interface{}) bool {
	requestedResource, ok := context[ContextKeyRequestResourceID].(string)
	if !ok {
		return false
	}

	// Check if resource matches Resource patterns
	resourceMatches := false
	resourceValues := statement.Resource.GetValues()
	for _, resourcePattern := range resourceValues {
		if pdp.resourceMatcher.Match(resourcePattern, requestedResource, context) {
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
			if pdp.resourceMatcher.Match(notResourcePattern, requestedResource, context) {
				return false // Excluded by NotResource
			}
		}
	}

	return true
}
