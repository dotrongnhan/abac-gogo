package attributes

import (
	"fmt"
	"reflect"
	"strings"
	"time"

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

// EnrichContext enriches the evaluation context with all necessary attributes
func (r *AttributeResolver) EnrichContext(request *models.EvaluationRequest) (*models.EvaluationContext, error) {
	// Get subject
	subject, err := r.storage.GetSubject(request.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subject: %w", err)
	}

	// Get resource
	resource, err := r.storage.GetResource(request.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	// Get action
	action, err := r.storage.GetAction(request.Action)
	if err != nil {
		return nil, fmt.Errorf("failed to get action: %w", err)
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

// enrichEnvironmentContext adds computed environment attributes
func (r *AttributeResolver) enrichEnvironmentContext(context map[string]interface{}) map[string]interface{} {
	enriched := make(map[string]interface{})

	// Copy existing context
	for k, v := range context {
		enriched[k] = v
	}

	// Add current timestamp if not present
	if _, exists := enriched["timestamp"]; !exists {
		enriched["timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Extract time_of_day from timestamp
	if timestampStr, ok := enriched["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			enriched["time_of_day"] = t.Format("15:04")
			enriched["day_of_week"] = strings.ToLower(t.Weekday().String())
			enriched["hour"] = t.Hour()
			enriched["is_business_hours"] = r.isBusinessHours(t)
		}
	}

	// Add derived IP attributes
	if sourceIP, ok := enriched["client_ip"].(string); ok {
		enriched["is_internal_ip"] = r.isInternalIP(sourceIP)
		enriched["ip_subnet"] = r.getIPSubnet(sourceIP)
	}

	return enriched
}

// resolveDynamicAttributes computes dynamic subject attributes
func (r *AttributeResolver) resolveDynamicAttributes(subject *models.Subject, environment map[string]interface{}) {
	if subject.Attributes == nil {
		subject.Attributes = make(map[string]interface{})
	}

	// Calculate years_of_service if hire_date is available
	if hireDateStr, ok := subject.Attributes["hire_date"].(string); ok {
		if hireDate, err := time.Parse("2006-01-02", hireDateStr); err == nil {
			years := time.Since(hireDate).Hours() / (24 * 365.25)
			subject.Attributes["years_of_service"] = int(years)
		}
	}

	// Add computed attributes based on current time
	now := time.Now()
	subject.Attributes["current_hour"] = now.Hour()
	subject.Attributes["current_day"] = strings.ToLower(now.Weekday().String())
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

	// Business hours: 8 AM to 6 PM, Monday to Friday
	return weekday >= time.Monday && weekday <= time.Friday && hour >= 8 && hour < 18
}

func (r *AttributeResolver) isInternalIP(ip string) bool {
	// Check for private IP ranges
	return strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "172.16.") ||
		ip == "127.0.0.1" ||
		ip == "localhost"
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

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		// Convert wildcard pattern to regex
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		regexPattern = "^" + regexPattern + "$"

		// Simple regex matching (could be improved with proper regex library)
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
