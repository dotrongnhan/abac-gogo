package core

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"abac_go_example/attributes"
	"abac_go_example/constants"
	"abac_go_example/evaluator/conditions"
	"abac_go_example/evaluator/matchers"
	"abac_go_example/models"
	"abac_go_example/storage"
)

// PolicyDecisionPointInterface defines the interface for policy evaluation
type PolicyDecisionPointInterface interface {
	Evaluate(request *models.EvaluationRequest) (*models.Decision, error)
}

// PolicyDecisionPoint (PDP) is the main evaluation engine
type PolicyDecisionPoint struct {
	storage                    storage.Storage
	attributeResolver          *attributes.AttributeResolver
	actionMatcher              *matchers.ActionMatcher
	resourceMatcher            *matchers.ResourceMatcher
	enhancedConditionEvaluator *conditions.EnhancedConditionEvaluator
}

// NewPolicyDecisionPoint creates a new PDP instance and returns the interface
func NewPolicyDecisionPoint(storage storage.Storage) PolicyDecisionPointInterface {
	return &PolicyDecisionPoint{
		storage:                    storage,
		attributeResolver:          attributes.NewAttributeResolver(storage),
		actionMatcher:              matchers.NewActionMatcher(),
		resourceMatcher:            matchers.NewResourceMatcher(),
		enhancedConditionEvaluator: conditions.NewEnhancedConditionEvaluator(),
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

	// Step 2: Get applicable policies with pre-filtering
	allPolicies, err := pdp.storage.GetPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	// Step 3: Build enhanced evaluation context with time-based and environmental attributes
	evalContext := pdp.BuildEnhancedEvaluationContext(request, context)

	// Step 4: Evaluate all policies with Deny-Override algorithm
	decision := pdp.evaluateNewPolicies(allPolicies, evalContext)

	// Step 5: Calculate evaluation time
	evaluationTime := int(time.Since(startTime).Milliseconds())
	decision.EvaluationTimeMs = evaluationTime

	return decision, nil
}

// BuildEnhancedEvaluationContext builds enhanced context map with structured attributes
func (pdp *PolicyDecisionPoint) BuildEnhancedEvaluationContext(request *models.EvaluationRequest, context *models.EvaluationContext) map[string]interface{} {
	evalContext := make(map[string]interface{}, 50)

	// Request context
	evalContext[constants.ContextKeyRequestUserID] = request.SubjectID
	evalContext[constants.ContextKeyRequestAction] = request.Action
	evalContext[constants.ContextKeyRequestResourceID] = request.ResourceID
	evalContext[constants.ContextKeyRequestTime] = context.Timestamp.Format(time.RFC3339)

	// Enhanced time-based attributes
	pdp.addTimeBasedAttributes(evalContext, request, context)

	// Enhanced environmental context
	pdp.addEnvironmentalContext(evalContext, request)

	// Structured subject attributes
	pdp.addStructuredSubjectAttributes(evalContext, context)

	// Structured resource attributes
	pdp.addStructuredResourceAttributes(evalContext, context)

	// Add custom context from request
	for key, value := range request.Context {
		evalContext[constants.ContextKeyRequestPrefix+key] = value
	}

	// Legacy environment attributes for backward compatibility
	for key, value := range context.Environment {
		evalContext[constants.ContextKeyEnvironmentPrefix+key] = value
	}

	return evalContext
}

// addTimeBasedAttributes adds time-based attributes (improvement #4)
func (pdp *PolicyDecisionPoint) addTimeBasedAttributes(evalContext map[string]interface{}, request *models.EvaluationRequest, context *models.EvaluationContext) {
	var timestamp time.Time

	// Use provided timestamp or current time
	if request.Timestamp != nil {
		timestamp = *request.Timestamp
	} else {
		timestamp = context.Timestamp
	}

	// Add time of day (HH:MM format)
	timeOfDay := timestamp.Format("15:04")
	evalContext[constants.ContextKeyTimeOfDay] = timeOfDay

	// Add day of week
	dayOfWeek := timestamp.Weekday().String()
	evalContext[constants.ContextKeyDayOfWeek] = dayOfWeek

	// Add additional time attributes
	evalContext[constants.ContextKeyEnvironmentPrefix+"hour"] = timestamp.Hour()
	evalContext[constants.ContextKeyEnvironmentPrefix+"minute"] = timestamp.Minute()
	evalContext[constants.ContextKeyEnvironmentPrefix+"is_weekend"] = timestamp.Weekday() == time.Saturday || timestamp.Weekday() == time.Sunday
	evalContext[constants.ContextKeyEnvironmentPrefix+"is_business_hours"] = timestamp.Hour() >= 9 && timestamp.Hour() < 17
}

// addEnvironmentalContext adds environmental context (improvement #5)
func (pdp *PolicyDecisionPoint) addEnvironmentalContext(evalContext map[string]interface{}, request *models.EvaluationRequest) {
	if request.Environment == nil {
		return
	}

	env := request.Environment

	// Client IP and related attributes
	if env.ClientIP != "" {
		evalContext[constants.ContextKeyClientIP] = env.ClientIP
		evalContext[constants.ContextKeyEnvironmentPrefix+"is_internal_ip"] = pdp.isInternalIP(env.ClientIP)
		evalContext[constants.ContextKeyEnvironmentPrefix+"ip_class"] = pdp.getIPClass(env.ClientIP)
	}

	// User Agent and device detection
	if env.UserAgent != "" {
		evalContext[constants.ContextKeyUserAgent] = env.UserAgent
		evalContext[constants.ContextKeyEnvironmentPrefix+"is_mobile"] = pdp.isMobileUserAgent(env.UserAgent)
		evalContext[constants.ContextKeyEnvironmentPrefix+"browser"] = pdp.getBrowserFromUserAgent(env.UserAgent)
	}

	// Location attributes
	if env.Country != "" {
		evalContext[constants.ContextKeyCountry] = env.Country
	}
	if env.Region != "" {
		evalContext[constants.ContextKeyRegion] = env.Region
	}

	// Custom environment attributes
	for key, value := range env.Attributes {
		evalContext[constants.ContextKeyEnvironmentPrefix+key] = value
	}
}

// addStructuredSubjectAttributes adds structured subject attributes (improvement #6)
func (pdp *PolicyDecisionPoint) addStructuredSubjectAttributes(evalContext map[string]interface{}, context *models.EvaluationContext) {
	if context.Subject == nil {
		return
	}

	// Flat attributes for backward compatibility
	for key, value := range context.Subject.Attributes {
		evalContext[constants.ContextKeyUserPrefix+key] = value
	}
	evalContext[constants.ContextKeyUserPrefix+"SubjectType"] = context.Subject.SubjectType

	// Structured attributes for enhanced access
	userContext := map[string]interface{}{
		"subject_type": context.Subject.SubjectType,
		"attributes":   map[string]interface{}(context.Subject.Attributes),
	}
	evalContext["user"] = userContext
}

// addStructuredResourceAttributes adds structured resource attributes (improvement #6)
func (pdp *PolicyDecisionPoint) addStructuredResourceAttributes(evalContext map[string]interface{}, context *models.EvaluationContext) {
	if context.Resource == nil {
		return
	}

	// Flat attributes for backward compatibility
	for key, value := range context.Resource.Attributes {
		evalContext[constants.ContextKeyResourcePrefix+key] = value
	}
	evalContext[constants.ContextKeyResourcePrefix+"ResourceType"] = context.Resource.ResourceType
	evalContext[constants.ContextKeyResourcePrefix+"ResourceId"] = context.Resource.ResourceID

	// Structured attributes for enhanced access
	resourceContext := map[string]interface{}{
		"resource_type": context.Resource.ResourceType,
		"resource_id":   context.Resource.ResourceID,
		"attributes":    map[string]interface{}(context.Resource.Attributes),
	}
	evalContext["resource"] = resourceContext
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
				if strings.ToLower(statement.Effect) == constants.EffectDeny {
					return &models.Decision{
						Result:          constants.ResultDeny,
						MatchedPolicies: matchedPolicies,
						Reason:          fmt.Sprintf(constants.ReasonDeniedByStatement, statement.Sid),
					}
				}
			}
		}
	}

	// Step 3: If we have any Allow statements, return allow
	if len(matchedStatements) > 0 {
		return &models.Decision{
			Result:          constants.ResultPermit,
			MatchedPolicies: matchedPolicies,
			Reason:          fmt.Sprintf(constants.ReasonAllowedByStatements, strings.Join(matchedStatements, ", ")),
		}
	}

	// Step 4: Default deny (no matching policies)
	return &models.Decision{
		Result:          constants.ResultDeny,
		MatchedPolicies: []string{},
		Reason:          constants.ReasonImplicitDeny,
	}
}

// evaluateStatement evaluates a single policy statement against the given context.
// It performs three main checks: action matching, resource matching, and condition evaluation.
// Returns true if all checks pass, false otherwise.
func (pdp *PolicyDecisionPoint) evaluateStatement(statement models.PolicyStatement, context map[string]interface{}) bool {
	// Validate input parameters
	if !pdp.isValidEvaluationContext(context) {
		log.Printf("Error: Invalid evaluation context provided")
		return false
	}

	// Early return pattern for better readability
	if !pdp.isActionMatched(statement.Action, context) {
		return false
	}

	if !pdp.isResourceMatched(statement, context) {
		return false
	}

	return pdp.areConditionsSatisfied(statement.Condition, context)
}

// isValidEvaluationContext validates that the evaluation context contains required keys
// and is properly structured for policy evaluation.
func (pdp *PolicyDecisionPoint) isValidEvaluationContext(context map[string]interface{}) bool {
	if context == nil {
		return false
	}

	// Check for essential context keys (Action and ResourceId are mandatory)
	essentialKeys := []string{
		constants.ContextKeyRequestAction,
		constants.ContextKeyRequestResourceID,
	}

	for _, key := range essentialKeys {
		if _, exists := context[key]; !exists {
			log.Printf("Warning: Missing essential context key: %s", key)
			return false
		}
	}

	// UserId is recommended but not strictly required for some evaluation scenarios
	if _, exists := context[constants.ContextKeyRequestUserID]; !exists {
		log.Printf("Info: UserId not provided in context - some policies may not evaluate correctly")
	}

	// Validate context size to prevent DoS attacks
	if len(context) > constants.MaxConditionKeys {
		log.Printf("Warning: Context size exceeds maximum allowed keys: %d > %d", len(context), constants.MaxConditionKeys)
		return false
	}

	return true
}

// isActionMatched checks if the requested action matches the statement's action specification.
func (pdp *PolicyDecisionPoint) isActionMatched(actionSpec models.JSONActionResource, context map[string]interface{}) bool {
	requestedAction, ok := context[constants.ContextKeyRequestAction].(string)
	if !ok {
		log.Printf("Warning: Missing or invalid action in context: %v", context[constants.ContextKeyRequestAction])
		return false
	}

	if requestedAction == "" {
		log.Printf("Warning: Empty action provided in evaluation context")
		return false
	}

	actionValues := actionSpec.GetValues()
	if len(actionValues) == 0 {
		log.Printf("Warning: No action patterns specified in policy statement")
		return false
	}

	for _, actionPattern := range actionValues {
		if actionPattern == "" {
			log.Printf("Warning: Empty action pattern found in policy statement")
			continue
		}
		if pdp.actionMatcher.Match(actionPattern, requestedAction) {
			return true
		}
	}
	return false
}

// isResourceMatched checks if the requested resource matches the statement's resource specification
// and does not match any NotResource exclusion patterns.
func (pdp *PolicyDecisionPoint) isResourceMatched(statement models.PolicyStatement, context map[string]interface{}) bool {
	requestedResource, ok := context[constants.ContextKeyRequestResourceID].(string)
	if !ok {
		log.Printf("Warning: Missing or invalid resource ID in context: %v", context[constants.ContextKeyRequestResourceID])
		return false
	}

	if requestedResource == "" {
		log.Printf("Warning: Empty resource ID provided in evaluation context")
		return false
	}

	// Check positive resource matching
	if !pdp.matchesResourcePatterns(statement.Resource, requestedResource, context) {
		return false
	}

	// Check NotResource exclusions
	return !pdp.matchesNotResourcePatterns(statement.NotResource, requestedResource, context)
}

// matchesResourcePatterns checks if the resource matches any of the specified patterns.
func (pdp *PolicyDecisionPoint) matchesResourcePatterns(resourceSpec models.JSONActionResource, requestedResource string, context map[string]interface{}) bool {
	resourceValues := resourceSpec.GetValues()
	for _, resourcePattern := range resourceValues {
		if pdp.resourceMatcher.Match(resourcePattern, requestedResource, context) {
			return true
		}
	}
	return false
}

// matchesNotResourcePatterns checks if the resource matches any NotResource exclusion patterns.
func (pdp *PolicyDecisionPoint) matchesNotResourcePatterns(notResourceSpec models.JSONActionResource, requestedResource string, context map[string]interface{}) bool {
	// If no NotResource patterns are specified, return false (no exclusions)
	if !notResourceSpec.IsArray && notResourceSpec.Single == "" {
		return false
	}

	notResourceValues := notResourceSpec.GetValues()
	for _, notResourcePattern := range notResourceValues {
		if pdp.resourceMatcher.Match(notResourcePattern, requestedResource, context) {
			return true // Resource is excluded
		}
	}
	return false
}

// areConditionsSatisfied evaluates all conditions in the statement.
// Returns true if no conditions are specified or all conditions pass.
func (pdp *PolicyDecisionPoint) areConditionsSatisfied(conditions map[string]interface{}, context map[string]interface{}) bool {
	// No conditions means always satisfied
	if len(conditions) == 0 {
		return true
	}

	// Validate context is not nil
	if context == nil {
		log.Printf("Error: Evaluation context is nil when evaluating conditions")
		return false
	}

	// Use enhanced condition evaluator exclusively
	result := pdp.enhancedConditionEvaluator.EvaluateConditions(conditions, context)
	if !result {
		log.Printf("Debug: Enhanced condition evaluation failed for conditions: %v", conditions)
	}
	return result
}

// Helper methods for environmental context processing

// isInternalIP checks if an IP address is internal/private
func (pdp *PolicyDecisionPoint) isInternalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, rangeStr := range privateRanges {
		_, cidr, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// getIPClass returns the class of IP address
func (pdp *PolicyDecisionPoint) getIPClass(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "invalid"
	}

	if ip.To4() != nil {
		return "ipv4"
	}
	return "ipv6"
}

// isMobileUserAgent detects if user agent is from mobile device
func (pdp *PolicyDecisionPoint) isMobileUserAgent(userAgent string) bool {
	mobilePatterns := []string{
		"(?i)mobile",
		"(?i)android",
		"(?i)iphone",
		"(?i)ipad",
		"(?i)blackberry",
		"(?i)windows phone",
	}

	for _, pattern := range mobilePatterns {
		matched, _ := regexp.MatchString(pattern, userAgent)
		if matched {
			return true
		}
	}

	return false
}

// getBrowserFromUserAgent extracts browser name from user agent
func (pdp *PolicyDecisionPoint) getBrowserFromUserAgent(userAgent string) string {
	browserPatterns := map[string]string{
		"(?i)chrome":  "chrome",
		"(?i)firefox": "firefox",
		"(?i)safari":  "safari",
		"(?i)edge":    "edge",
		"(?i)opera":   "opera",
	}

	for pattern, browser := range browserPatterns {
		matched, _ := regexp.MatchString(pattern, userAgent)
		if matched {
			return browser
		}
	}

	return "unknown"
}
