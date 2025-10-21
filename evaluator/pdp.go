package evaluator

import (
	"fmt"
	"strings"
	"time"

	"abac_go_example/attributes"
	"abac_go_example/models"
	"abac_go_example/operators"
	"abac_go_example/storage"
)

// PolicyDecisionPoint (PDP) is the main evaluation engine
type PolicyDecisionPoint struct {
	storage            storage.Storage
	attributeResolver  *attributes.AttributeResolver
	operatorRegistry   *operators.OperatorRegistry
	actionMatcher      *ActionMatcher
	resourceMatcher    *ResourceMatcher
	conditionEvaluator *ConditionEvaluator
}

// NewPolicyDecisionPoint creates a new PDP instance
func NewPolicyDecisionPoint(storage storage.Storage) *PolicyDecisionPoint {
	return &PolicyDecisionPoint{
		storage:            storage,
		attributeResolver:  attributes.NewAttributeResolver(storage),
		operatorRegistry:   operators.NewOperatorRegistry(),
		actionMatcher:      NewActionMatcher(),
		resourceMatcher:    NewResourceMatcher(),
		conditionEvaluator: NewConditionEvaluator(),
	}
}

// Evaluate performs policy evaluation for a given request
func (pdp *PolicyDecisionPoint) Evaluate(request *models.EvaluationRequest) (*models.Decision, error) {
	startTime := time.Now()

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

	// Step 3: Filter applicable policies (LEGACY - using new method)
	// applicablePolicies := pdp.filterApplicablePolicies(allPolicies, context)

	// Step 4: Sort policies by priority (legacy - commented out for new format)
	// sort.Slice(applicablePolicies, func(i, j int) bool {
	//	return applicablePolicies[i].Priority < applicablePolicies[j].Priority
	// })

	// Step 5: Evaluate policies with short-circuit logic (LEGACY - using new method)
	// decision := pdp.evaluatePolicies(applicablePolicies, context)

	// Use new evaluation method
	evalContext := pdp.buildEvaluationContext(&models.EvaluationRequest{
		SubjectID:  context.Subject.ID,
		ResourceID: context.Resource.ResourceID,
		Action:     context.Action.ActionName,
	}, context)
	decision := pdp.evaluateNewPolicies(allPolicies, evalContext)

	// Step 6: Calculate evaluation time
	evaluationTime := int(time.Since(startTime).Milliseconds())
	decision.EvaluationTimeMs = evaluationTime

	return decision, nil
}

// EvaluateNew performs policy evaluation using the new JSON policy format
func (pdp *PolicyDecisionPoint) EvaluateNew(request *models.EvaluationRequest) (*models.Decision, error) {
	startTime := time.Now()

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

	// Step 3: Build evaluation context for new format
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
	evalContext["request:UserId"] = request.SubjectID
	evalContext["request:Action"] = request.Action
	evalContext["request:ResourceId"] = request.ResourceID
	evalContext["request:Time"] = context.Timestamp.Format(time.RFC3339)

	// Add custom context from request
	for key, value := range request.Context {
		evalContext["request:"+key] = value
	}

	// Subject attributes
	if context.Subject != nil {
		for key, value := range context.Subject.Attributes {
			evalContext["user:"+key] = value
		}
		evalContext["user:SubjectType"] = context.Subject.SubjectType
	}

	// Resource attributes
	if context.Resource != nil {
		for key, value := range context.Resource.Attributes {
			evalContext["resource:"+key] = value
		}
		evalContext["resource:ResourceType"] = context.Resource.ResourceType
		evalContext["resource:ResourceId"] = context.Resource.ResourceID
	}

	// Environment attributes
	for key, value := range context.Environment {
		evalContext["environment:"+key] = value
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
				if strings.ToLower(statement.Effect) == "deny" {
					return &models.Decision{
						Result:          "deny",
						MatchedPolicies: matchedPolicies,
						Reason:          fmt.Sprintf("Denied by statement: %s", statement.Sid),
					}
				}
			}
		}
	}

	// Step 3: If we have any Allow statements, return allow
	if len(matchedStatements) > 0 {
		return &models.Decision{
			Result:          "permit",
			MatchedPolicies: matchedPolicies,
			Reason:          fmt.Sprintf("Allowed by statements: %s", strings.Join(matchedStatements, ", ")),
		}
	}

	// Step 4: Default deny (no matching policies)
	return &models.Decision{
		Result:          "deny",
		MatchedPolicies: []string{},
		Reason:          "No matching policies found (implicit deny)",
	}
}

// evaluateStatement evaluates a single policy statement
func (pdp *PolicyDecisionPoint) evaluateStatement(statement models.PolicyStatement, context map[string]interface{}) bool {
	// Step 1: Check action matching
	if !pdp.matchAction(statement.Action, context) {
		return false
	}

	// Step 2: Check resource matching
	if !pdp.matchResource(statement.Resource, context) {
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
	requestedAction, ok := context["request:Action"].(string)
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
func (pdp *PolicyDecisionPoint) matchResource(resourceSpec models.JSONActionResource, context map[string]interface{}) bool {
	requestedResource, ok := context["request:ResourceId"].(string)
	if !ok {
		return false
	}

	resourceValues := resourceSpec.GetValues()
	for _, resourcePattern := range resourceValues {
		if pdp.resourceMatcher.Match(resourcePattern, requestedResource, context) {
			return true
		}
	}
	return false
}

// filterApplicablePolicies filters policies that might apply to the request (LEGACY - COMMENTED OUT)
/*
func (pdp *PolicyDecisionPoint) filterApplicablePolicies(policies []*models.Policy, context *models.EvaluationContext) []*models.Policy {
	var applicable []*models.Policy

	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		// Check if action matches
		if !pdp.actionMatches(policy.Actions, context.Action.ActionName) {
			continue
		}

		// Check if resource pattern matches
		if !pdp.resourcePatternMatches(policy.ResourcePatterns, context.Resource) {
			continue
		}

		applicable = append(applicable, policy)
	}

	return applicable
}
*/

// LEGACY METHODS - COMMENTED OUT FOR NEW FORMAT
/*
// actionMatches checks if the requested action matches policy actions
func (pdp *PolicyDecisionPoint) actionMatches(policyActions []string, requestedAction string) bool {
	if len(policyActions) == 0 {
		return true // No action restriction
	}

	for _, action := range policyActions {
		if action == "*" || action == requestedAction {
			return true
		}
	}
	return false
}

// resourcePatternMatches checks if the resource matches any of the policy patterns
func (pdp *PolicyDecisionPoint) resourcePatternMatches(patterns []string, resource *models.Resource) bool {
	if len(patterns) == 0 {
		return true // No resource restriction
	}

	resourcePath := resource.ResourceID
	if resource.Path != "" {
		resourcePath = resource.Path
	}

	for _, pattern := range patterns {
		if pdp.attributeResolver.MatchResourcePattern(pattern, resourcePath) {
			return true
		}
		// Also check against resource ID
		if pdp.attributeResolver.MatchResourcePattern(pattern, resource.ResourceID) {
			return true
		}
	}
	return false
}

// evaluatePolicies evaluates all applicable policies and returns final decision
func (pdp *PolicyDecisionPoint) evaluatePolicies(policies []*models.Policy, context *models.EvaluationContext) *models.Decision {
	var matchedPolicies []string
	var permitFound bool

	for _, policy := range policies {
		// Evaluate all rules in the policy
		if pdp.evaluatePolicy(policy, context) {
			matchedPolicies = append(matchedPolicies, policy.ID)

			if policy.Effect == "deny" {
				// DENY overrides everything - short circuit
				return &models.Decision{
					Result:          "deny",
					MatchedPolicies: matchedPolicies,
					Reason:          fmt.Sprintf("Denied by policy: %s", policy.PolicyName),
				}
			} else if policy.Effect == "permit" {
				permitFound = true
			}
		}
	}

	// Final decision logic
	if permitFound {
		return &models.Decision{
			Result:          "permit",
			MatchedPolicies: matchedPolicies,
			Reason:          "Access granted by matching permit policies",
		}
	}

	return &models.Decision{
		Result:          "not_applicable",
		MatchedPolicies: matchedPolicies,
		Reason:          "No applicable policies found",
	}
}

// evaluatePolicy evaluates a single policy against the context
func (pdp *PolicyDecisionPoint) evaluatePolicy(policy *models.Policy, context *models.EvaluationContext) bool {
	// All rules must match (AND logic)
	for _, rule := range policy.Rules {
		if !pdp.evaluateRule(rule, context) {
			return false
		}
	}
	return true
}

// evaluateRule evaluates a single rule against the context
func (pdp *PolicyDecisionPoint) evaluateRule(rule models.PolicyRule, context *models.EvaluationContext) bool {
	// Get the actual value based on target type
	var actualValue interface{}

	switch rule.TargetType {
	case "subject":
		actualValue = pdp.attributeResolver.GetAttributeValue(context.Subject, rule.AttributePath)
	case "resource":
		actualValue = pdp.attributeResolver.GetAttributeValue(context.Resource, rule.AttributePath)
	case "action":
		actualValue = pdp.attributeResolver.GetAttributeValue(context.Action, rule.AttributePath)
	case "environment":
		actualValue = pdp.attributeResolver.GetAttributeValue(context.Environment, rule.AttributePath)
	default:
		return false
	}

	// Get the operator
	operator, err := pdp.operatorRegistry.Get(rule.Operator)
	if err != nil {
		return false
	}

	// Evaluate the rule
	result := operator.Evaluate(actualValue, rule.ExpectedValue)

	// Apply negation if needed
	if rule.IsNegative {
		result = !result
	}

	return result
}
*/

// LEGACY METHODS - COMMENTED OUT FOR NEW FORMAT
/*
// BatchEvaluate evaluates multiple requests in parallel
func (pdp *PolicyDecisionPoint) BatchEvaluate(requests []*models.EvaluationRequest) ([]*models.Decision, error) {
	decisions := make([]*models.Decision, len(requests))
	errors := make([]error, len(requests))

	// Simple sequential evaluation (could be parallelized with goroutines)
	for i, request := range requests {
		decision, err := pdp.Evaluate(request)
		decisions[i] = decision
		errors[i] = err
	}

	// Check if any errors occurred
	for _, err := range errors {
		if err != nil {
			return decisions, err
		}
	}

	return decisions, nil
}

// GetApplicablePolicies returns policies that would apply to a request (for debugging)
func (pdp *PolicyDecisionPoint) GetApplicablePolicies(request *models.EvaluationRequest) ([]*models.Policy, error) {
	context, err := pdp.attributeResolver.EnrichContext(request)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich context: %w", err)
	}

	allPolicies, err := pdp.storage.GetPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	return pdp.filterApplicablePolicies(allPolicies, context), nil
}
*/

/*
// ExplainDecision provides detailed explanation of how a decision was reached
func (pdp *PolicyDecisionPoint) ExplainDecision(request *models.EvaluationRequest) (map[string]interface{}, error) {
	context, err := pdp.attributeResolver.EnrichContext(request)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich context: %w", err)
	}

	allPolicies, err := pdp.storage.GetPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	applicablePolicies := pdp.filterApplicablePolicies(allPolicies, context)

	explanation := map[string]interface{}{
		"request":             request,
		"context":             context,
		"total_policies":      len(allPolicies),
		"applicable_policies": len(applicablePolicies),
		"policy_evaluations":  []map[string]interface{}{},
	}

	// Evaluate each policy and record results
	policyEvaluations := []map[string]interface{}{}
	for _, policy := range applicablePolicies {
		policyResult := map[string]interface{}{
			"policy_id":   policy.ID,
			"policy_name": policy.PolicyName,
			"effect":      policy.Effect,
			"priority":    policy.Priority,
			"matched":     false,
			"rules":       []map[string]interface{}{},
		}

		allRulesMatch := true
		ruleResults := []map[string]interface{}{}

		for _, rule := range policy.Rules {
			var actualValue interface{}
			switch rule.TargetType {
			case "subject":
				actualValue = pdp.attributeResolver.GetAttributeValue(context.Subject, rule.AttributePath)
			case "resource":
				actualValue = pdp.attributeResolver.GetAttributeValue(context.Resource, rule.AttributePath)
			case "action":
				actualValue = pdp.attributeResolver.GetAttributeValue(context.Action, rule.AttributePath)
			case "environment":
				actualValue = pdp.attributeResolver.GetAttributeValue(context.Environment, rule.AttributePath)
			}

			operator, _ := pdp.operatorRegistry.Get(rule.Operator)
			ruleMatch := false
			if operator != nil {
				ruleMatch = operator.Evaluate(actualValue, rule.ExpectedValue)
				if rule.IsNegative {
					ruleMatch = !ruleMatch
				}
			}

			if !ruleMatch {
				allRulesMatch = false
			}

			ruleResults = append(ruleResults, map[string]interface{}{
				"target_type":    rule.TargetType,
				"attribute_path": rule.AttributePath,
				"operator":       rule.Operator,
				"expected_value": rule.ExpectedValue,
				"actual_value":   actualValue,
				"matched":        ruleMatch,
				"is_negative":    rule.IsNegative,
			})
		}

		policyResult["matched"] = allRulesMatch
		policyResult["rules"] = ruleResults
		policyEvaluations = append(policyEvaluations, policyResult)
	}

	explanation["policy_evaluations"] = policyEvaluations
	return explanation, nil
}
*/
