package evaluator

import (
	"fmt"
	"sort"
	"time"

	"abac_go_example/attributes"
	"abac_go_example/models"
	"abac_go_example/operators"
	"abac_go_example/storage"
)

// PolicyDecisionPoint (PDP) is the main evaluation engine
type PolicyDecisionPoint struct {
	storage           storage.Storage
	attributeResolver *attributes.AttributeResolver
	operatorRegistry  *operators.OperatorRegistry
}

// NewPolicyDecisionPoint creates a new PDP instance
func NewPolicyDecisionPoint(storage storage.Storage) *PolicyDecisionPoint {
	return &PolicyDecisionPoint{
		storage:           storage,
		attributeResolver: attributes.NewAttributeResolver(storage),
		operatorRegistry:  operators.NewOperatorRegistry(),
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

	// Step 3: Filter applicable policies
	applicablePolicies := pdp.filterApplicablePolicies(allPolicies, context)

	// Step 4: Sort policies by priority (ascending - lower number = higher priority)
	sort.Slice(applicablePolicies, func(i, j int) bool {
		return applicablePolicies[i].Priority < applicablePolicies[j].Priority
	})

	// Step 5: Evaluate policies with short-circuit logic
	decision := pdp.evaluatePolicies(applicablePolicies, context)

	// Step 6: Calculate evaluation time
	evaluationTime := int(time.Since(startTime).Milliseconds())
	decision.EvaluationTimeMs = evaluationTime

	return decision, nil
}

// filterApplicablePolicies filters policies that might apply to the request
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
