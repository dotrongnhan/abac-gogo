package pep

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"abac_go_example/models"
)

// MethodInterceptor provides method-level access control
type MethodInterceptor struct {
	pep    *PolicyEnforcementPoint
	config *InterceptorConfig
}

// InterceptorConfig holds configuration for method interceptors
type InterceptorConfig struct {
	// Default values
	DefaultSubjectID string `json:"default_subject_id"`
	DefaultAction    string `json:"default_action"`

	// Error handling
	FailOnError    bool `json:"fail_on_error"`
	LogDeniedCalls bool `json:"log_denied_calls"`

	// Performance
	EnableCaching bool `json:"enable_caching"`
	TimeoutMs     int  `json:"timeout_ms"`
}

// DefaultInterceptorConfig returns default interceptor configuration
func DefaultInterceptorConfig() *InterceptorConfig {
	return &InterceptorConfig{
		DefaultSubjectID: "system",
		DefaultAction:    "execute",
		FailOnError:      true,
		LogDeniedCalls:   true,
		EnableCaching:    true,
		TimeoutMs:        100,
	}
}

// NewMethodInterceptor creates a new method interceptor
func NewMethodInterceptor(pep *PolicyEnforcementPoint, config *InterceptorConfig) *MethodInterceptor {
	if config == nil {
		config = DefaultInterceptorConfig()
	}

	return &MethodInterceptor{
		pep:    pep,
		config: config,
	}
}

// InterceptCall intercepts a method call and enforces access control
func (mi *MethodInterceptor) InterceptCall(ctx context.Context, subjectID, resourceID, action string, fn func() error) error {
	// Use defaults if not provided
	if subjectID == "" {
		subjectID = mi.config.DefaultSubjectID
	}
	if action == "" {
		action = mi.config.DefaultAction
	}
	if resourceID == "" {
		// Try to derive resource from function name
		resourceID = mi.getFunctionName(fn)
	}

	// Create evaluation request
	evalRequest := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("intercept_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: resourceID,
		Action:     action,
		Context: map[string]interface{}{
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"call_type":     "method_intercept",
			"function_name": mi.getFunctionName(fn),
		},
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(mi.config.TimeoutMs)*time.Millisecond)
	defer cancel()

	// Enforce policy
	result, err := mi.pep.EnforceRequest(timeoutCtx, evalRequest)
	if err != nil {
		if mi.config.FailOnError {
			return fmt.Errorf("access control check failed: %w", err)
		}
		// Log error but allow execution
		fmt.Printf("[PEP Interceptor] Error checking access for %s: %v\n", resourceID, err)
	}

	// Check if access is allowed
	if result != nil && !result.Allowed {
		if mi.config.LogDeniedCalls {
			fmt.Printf("[PEP Interceptor] Access denied for subject %s to resource %s (action: %s): %s\n",
				subjectID, resourceID, action, result.Reason)
		}
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Execute the function
	return fn()
}

// InterceptCallWithResult intercepts a method call that returns a result
func (mi *MethodInterceptor) InterceptCallWithResult(ctx context.Context, subjectID, resourceID, action string, fn func() (interface{}, error)) (interface{}, error) {
	err := mi.InterceptCall(ctx, subjectID, resourceID, action, func() error {
		_, err := fn()
		return err
	})

	if err != nil {
		return nil, err
	}

	return fn()
}

// getFunctionName extracts function name using reflection
func (mi *MethodInterceptor) getFunctionName(fn interface{}) string {
	if fn == nil {
		return "unknown"
	}

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return "not_a_function"
	}

	// Get function name from runtime
	fnPtr := fnValue.Pointer()
	fnInfo := runtime.FuncForPC(fnPtr)
	if fnInfo == nil {
		return "unknown_function"
	}

	fullName := fnInfo.Name()

	// Extract just the function name (remove package path)
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return fullName
}

// SecureFunction is a decorator that adds access control to functions
type SecureFunction struct {
	interceptor *MethodInterceptor
	subjectID   string
	resourceID  string
	action      string
}

// NewSecureFunction creates a new secure function decorator
func NewSecureFunction(interceptor *MethodInterceptor, subjectID, resourceID, action string) *SecureFunction {
	return &SecureFunction{
		interceptor: interceptor,
		subjectID:   subjectID,
		resourceID:  resourceID,
		action:      action,
	}
}

// Execute executes a function with access control
func (sf *SecureFunction) Execute(ctx context.Context, fn func() error) error {
	return sf.interceptor.InterceptCall(ctx, sf.subjectID, sf.resourceID, sf.action, fn)
}

// ExecuteWithResult executes a function with access control and returns result
func (sf *SecureFunction) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return sf.interceptor.InterceptCallWithResult(ctx, sf.subjectID, sf.resourceID, sf.action, fn)
}

// ServiceInterceptor provides service-level access control
type ServiceInterceptor struct {
	interceptor *MethodInterceptor
	serviceName string
}

// NewServiceInterceptor creates a new service interceptor
func NewServiceInterceptor(interceptor *MethodInterceptor, serviceName string) *ServiceInterceptor {
	return &ServiceInterceptor{
		interceptor: interceptor,
		serviceName: serviceName,
	}
}

// InterceptMethod intercepts a service method call
func (si *ServiceInterceptor) InterceptMethod(ctx context.Context, subjectID, methodName, action string, fn func() error) error {
	resourceID := fmt.Sprintf("%s.%s", si.serviceName, methodName)
	return si.interceptor.InterceptCall(ctx, subjectID, resourceID, action, fn)
}

// InterceptMethodWithResult intercepts a service method call with result
func (si *ServiceInterceptor) InterceptMethodWithResult(ctx context.Context, subjectID, methodName, action string, fn func() (interface{}, error)) (interface{}, error) {
	resourceID := fmt.Sprintf("%s.%s", si.serviceName, methodName)
	return si.interceptor.InterceptCallWithResult(ctx, subjectID, resourceID, action, fn)
}

// DatabaseInterceptor provides database operation access control
type DatabaseInterceptor struct {
	interceptor *MethodInterceptor
}

// NewDatabaseInterceptor creates a new database interceptor
func NewDatabaseInterceptor(interceptor *MethodInterceptor) *DatabaseInterceptor {
	return &DatabaseInterceptor{
		interceptor: interceptor,
	}
}

// InterceptQuery intercepts database query operations
func (di *DatabaseInterceptor) InterceptQuery(ctx context.Context, subjectID, table string, fn func() error) error {
	resourceID := fmt.Sprintf("db.%s", table)
	return di.interceptor.InterceptCall(ctx, subjectID, resourceID, "read", fn)
}

// InterceptInsert intercepts database insert operations
func (di *DatabaseInterceptor) InterceptInsert(ctx context.Context, subjectID, table string, fn func() error) error {
	resourceID := fmt.Sprintf("db.%s", table)
	return di.interceptor.InterceptCall(ctx, subjectID, resourceID, "create", fn)
}

// InterceptUpdate intercepts database update operations
func (di *DatabaseInterceptor) InterceptUpdate(ctx context.Context, subjectID, table string, fn func() error) error {
	resourceID := fmt.Sprintf("db.%s", table)
	return di.interceptor.InterceptCall(ctx, subjectID, resourceID, "update", fn)
}

// InterceptDelete intercepts database delete operations
func (di *DatabaseInterceptor) InterceptDelete(ctx context.Context, subjectID, table string, fn func() error) error {
	resourceID := fmt.Sprintf("db.%s", table)
	return di.interceptor.InterceptCall(ctx, subjectID, resourceID, "delete", fn)
}

// APIInterceptor provides API endpoint access control
type APIInterceptor struct {
	interceptor *MethodInterceptor
}

// NewAPIInterceptor creates a new API interceptor
func NewAPIInterceptor(interceptor *MethodInterceptor) *APIInterceptor {
	return &APIInterceptor{
		interceptor: interceptor,
	}
}

// InterceptEndpoint intercepts API endpoint calls
func (ai *APIInterceptor) InterceptEndpoint(ctx context.Context, subjectID, endpoint, method string, fn func() error) error {
	action := strings.ToLower(method)
	if action == "get" {
		action = "read"
	} else if action == "post" {
		action = "create"
	} else if action == "put" || action == "patch" {
		action = "update"
	}

	resourceID := fmt.Sprintf("api%s", endpoint)
	return ai.interceptor.InterceptCall(ctx, subjectID, resourceID, action, fn)
}

// InterceptEndpointWithResult intercepts API endpoint calls with result
func (ai *APIInterceptor) InterceptEndpointWithResult(ctx context.Context, subjectID, endpoint, method string, fn func() (interface{}, error)) (interface{}, error) {
	action := strings.ToLower(method)
	if action == "get" {
		action = "read"
	} else if action == "post" {
		action = "create"
	} else if action == "put" || action == "patch" {
		action = "update"
	}

	resourceID := fmt.Sprintf("api%s", endpoint)
	return ai.interceptor.InterceptCallWithResult(ctx, subjectID, resourceID, action, fn)
}

// Example usage patterns

// SecureService demonstrates how to use interceptors in a service
type SecureService struct {
	interceptor *ServiceInterceptor
}

// NewSecureService creates a new secure service
func NewSecureService(pep *PolicyEnforcementPoint) *SecureService {
	methodInterceptor := NewMethodInterceptor(pep, nil)
	serviceInterceptor := NewServiceInterceptor(methodInterceptor, "UserService")

	return &SecureService{
		interceptor: serviceInterceptor,
	}
}

// GetUser demonstrates secured method call
func (s *SecureService) GetUser(ctx context.Context, subjectID, userID string) (interface{}, error) {
	return s.interceptor.InterceptMethodWithResult(ctx, subjectID, "GetUser", "read", func() (interface{}, error) {
		// Actual business logic here
		return map[string]string{"id": userID, "name": "John Doe"}, nil
	})
}

// UpdateUser demonstrates secured method call
func (s *SecureService) UpdateUser(ctx context.Context, subjectID, userID string, data map[string]interface{}) error {
	return s.interceptor.InterceptMethod(ctx, subjectID, "UpdateUser", "update", func() error {
		// Actual business logic here
		fmt.Printf("Updating user %s with data: %v\n", userID, data)
		return nil
	})
}

// DeleteUser demonstrates secured method call
func (s *SecureService) DeleteUser(ctx context.Context, subjectID, userID string) error {
	return s.interceptor.InterceptMethod(ctx, subjectID, "DeleteUser", "delete", func() error {
		// Actual business logic here
		fmt.Printf("Deleting user %s\n", userID)
		return nil
	})
}
