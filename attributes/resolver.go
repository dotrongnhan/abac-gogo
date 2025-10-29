package attributes

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"abac_go_example/constants"
	"abac_go_example/models"
	"abac_go_example/storage"
)

// AttributeResolver handles attribute resolution and context enrichment
type AttributeResolver struct {
	storage storage.Storage
}

// NewAttributeResolver creates a new attribute resolver
func NewAttributeResolver(storage storage.Storage) *AttributeResolver {
	return &AttributeResolver{
		storage: storage,
	}
}

// validateRequest validates the evaluation request
func (r *AttributeResolver) validateRequest(request *models.EvaluationRequest) error {
	if request == nil {
		return fmt.Errorf("evaluation request cannot be nil")
	}

	if request.SubjectID == "" {
		return fmt.Errorf("subject ID cannot be empty")
	}

	if request.ResourceID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}

	if request.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}

	return nil
}

// EnrichContext enriches the evaluation context with all necessary attributes
func (r *AttributeResolver) EnrichContext(request *models.EvaluationRequest) (*models.EvaluationContext, error) {
	if err := r.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get subject
	subject, err := r.storage.GetSubject(request.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve subject '%s': %w", request.SubjectID, err)
	}
	if subject == nil {
		return nil, fmt.Errorf("subject '%s' not found", request.SubjectID)
	}

	// Get resource
	resource, err := r.storage.GetResource(request.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource '%s': %w", request.ResourceID, err)
	}
	if resource == nil {
		return nil, fmt.Errorf("resource '%s' not found", request.ResourceID)
	}

	// Get action
	action, err := r.storage.GetAction(request.Action)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve action '%s': %w", request.Action, err)
	}
	if action == nil {
		return nil, fmt.Errorf("action '%s' not found", request.Action)
	}

	// Enrich environment context
	environment := r.enrichEnvironmentContext(request.Context)

	// Resolve dynamic attributes
	r.resolveDynamicAttributes(subject, environment)

	return &models.EvaluationContext{
		Subject:     subject,
		Resource:    resource,
		Action:      action,
		Environment: environment,
		Timestamp:   time.Now(),
	}, nil
}

// EnrichContextWithTimeout enriches context with timeout support
func (r *AttributeResolver) EnrichContextWithTimeout(ctx context.Context, request *models.EvaluationRequest) (*models.EvaluationContext, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return r.EnrichContext(request)
}

// enrichEnvironmentContext adds computed environment attributes
func (r *AttributeResolver) enrichEnvironmentContext(context map[string]interface{}) map[string]interface{} {
	enriched := make(map[string]interface{})

	// Copy existing context
	for k, v := range context {
		enriched[k] = v
	}

	// Add current timestamp if not present
	if _, exists := enriched[constants.ContextKeyTimestamp]; !exists {
		enriched[constants.ContextKeyTimestamp] = time.Now().Format(time.RFC3339)
	}

	// Extract time_of_day from timestamp
	if timestampStr, ok := enriched[constants.ContextKeyTimestamp].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			enriched[constants.ContextKeyTimeOfDayShort] = t.Format("15:04")
			enriched[constants.ContextKeyDayOfWeekShort] = strings.ToLower(t.Weekday().String())
			enriched[constants.ContextKeyHour] = t.Hour()
			enriched[constants.ContextKeyIsBusinessHours] = r.isBusinessHours(t)
		}
	}

	// Add derived IP attributes
	if sourceIP, ok := enriched[constants.ContextKeyClientIPShort].(string); ok {
		enriched[constants.ContextKeyIsInternalIP] = r.isInternalIP(sourceIP)
		enriched[constants.ContextKeyIPSubnet] = r.getIPSubnet(sourceIP)
	}

	// Also check for 'source_ip' key for backward compatibility
	if sourceIP, ok := enriched[constants.ContextKeySourceIP].(string); ok {
		enriched[constants.ContextKeyIsInternalIP] = r.isInternalIP(sourceIP)
		enriched[constants.ContextKeyIPSubnet] = r.getIPSubnet(sourceIP)
	}

	return enriched
}

// resolveDynamicAttributes computes dynamic subject attributes
func (r *AttributeResolver) resolveDynamicAttributes(subject *models.Subject, environment map[string]interface{}) {
	if subject.Attributes == nil {
		subject.Attributes = make(map[string]interface{})
	}

	// Calculate years_of_service if hire_date is available
	if hireDateStr, ok := subject.Attributes[constants.ContextKeyHireDate].(string); ok {
		if hireDate, err := time.Parse("2006-01-02", hireDateStr); err == nil {
			years := time.Since(hireDate).Hours() / (24 * 365.25)
			subject.Attributes[constants.ContextKeyYearsOfService] = int(years)
		}
	}

	// Add computed attributes based on current time
	now := time.Now()
	subject.Attributes[constants.ContextKeyCurrentHour] = now.Hour()
	subject.Attributes[constants.ContextKeyCurrentDay] = strings.ToLower(now.Weekday().String())
}

// GetAttributeValue retrieves a nested attribute value using dot notation
func (r *AttributeResolver) GetAttributeValue(target interface{}, path string) interface{} {
	if target == nil {
		return nil
	}

	parts := strings.Split(path, ".")
	current := target

	for _, part := range parts {
		current = r.getFieldValue(current, part)
		if current == nil {
			return nil
		}
	}

	return current
}

// getFieldValue gets a field value from an object
func (r *AttributeResolver) getFieldValue(obj interface{}, field string) interface{} {
	if obj == nil {
		return nil
	}

	// Handle map access
	if m, ok := obj.(map[string]interface{}); ok {
		return m[field]
	}

	// Handle JSONMap custom type (convert to map[string]interface{})
	if jsonMap, ok := obj.(models.JSONMap); ok {
		return map[string]interface{}(jsonMap)[field]
	}

	// Handle struct access using reflection
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	// Try to find field by name (case-insensitive)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		// Check JSON tag
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName == field {
				fieldName = fieldType.Name
			}
		}

		if strings.EqualFold(fieldName, field) || strings.EqualFold(fieldType.Name, field) {
			fieldValue := v.Field(i)
			if fieldValue.CanInterface() {
				return fieldValue.Interface()
			}
		}
	}

	return nil
}

// ResolveHierarchy resolves hierarchical resource paths
func (r *AttributeResolver) ResolveHierarchy(resourcePath string) []string {
	if resourcePath == "" {
		return []string{}
	}

	parts := strings.Split(strings.Trim(resourcePath, "/"), "/")
	hierarchy := make([]string, 0, len(parts))

	current := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		current += "/" + part
		hierarchy = append(hierarchy, current)
	}

	// Add wildcard patterns
	wildcardHierarchy := make([]string, 0, len(hierarchy)*2)
	for _, path := range hierarchy {
		wildcardHierarchy = append(wildcardHierarchy, path)
		wildcardHierarchy = append(wildcardHierarchy, path+"/*")
	}

	return wildcardHierarchy
}

// Helper functions

func (r *AttributeResolver) isBusinessHours(t time.Time) bool {
	hour := t.Hour()
	weekday := t.Weekday()

	// Use constants from business rules
	return weekday >= constants.BusinessDayStart &&
		weekday <= constants.BusinessDayEnd &&
		hour >= constants.BusinessHoursStart &&
		hour < constants.BusinessHoursEnd
}

func (r *AttributeResolver) isInternalIP(ip string) bool {
	// Handle localhost string
	if ip == "localhost" {
		return true
	}

	// Parse IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check against private IP ranges using CIDR
	for _, cidr := range constants.PrivateIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func (r *AttributeResolver) getIPSubnet(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) >= 3 {
		return strings.Join(parts[:3], ".") + ".0/24"
	}
	return ip
}

// MatchResourcePattern checks if a resource matches a pattern (supports wildcards)
func (r *AttributeResolver) MatchResourcePattern(pattern, resource string) bool {
	if pattern == "*" {
		return true
	}

	if pattern == resource {
		return true
	}

	// Use Go's built-in pattern matching for better accuracy
	if matched, err := filepath.Match(pattern, resource); err == nil && matched {
		return true
	}

	// Fallback to simple wildcard matching for complex patterns
	if strings.Contains(pattern, "*") {
		return r.simpleWildcardMatch(pattern, resource)
	}

	return false
}

func (r *AttributeResolver) simpleWildcardMatch(pattern, str string) bool {
	// Simple wildcard matching implementation
	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	// Split by * and check each part
	parts := strings.Split(pattern, "*")
	if len(parts) == 0 {
		return true
	}

	// Check if string starts with first part
	if parts[0] != "" && !strings.HasPrefix(str, parts[0]) {
		return false
	}

	// Check if string ends with last part
	if parts[len(parts)-1] != "" && !strings.HasSuffix(str, parts[len(parts)-1]) {
		return false
	}

	// For more complex patterns, we'd need proper regex
	// This is a simplified implementation
	return true
}
