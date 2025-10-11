package pep

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

// ExampleWebServer demonstrates PEP integration with HTTP server
func ExampleWebServer() {
	// Initialize storage (use your preferred storage)
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize storage: %v", err))
	}

	// Initialize PDP
	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	// Initialize audit logger
	auditLogger, err := NewSimpleAuditLogger("./audit.log")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize audit logger: %v", err))
	}

	// Create simplified PEP config (without advanced features)
	config := &PEPConfig{
		CacheEnabled:      false, // Disabled for now
		FailSafeMode:      true,
		StrictValidation:  true,
		AuditEnabled:      true,
		RateLimitEnabled:  false, // Disabled for now
		EvaluationTimeout: time.Millisecond * 100,
	}

	// Initialize PEP
	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)

	// Create HTTP middleware
	middleware := NewHTTPMiddleware(pep, nil)

	// Setup routes with PEP protection
	mux := http.NewServeMux()

	// Protected routes
	mux.Handle("/api/users", middleware.Handler(http.HandlerFunc(handleUsers)))
	mux.Handle("/api/admin", middleware.Handler(http.HandlerFunc(handleAdmin)))

	// Unprotected routes
	mux.HandleFunc("/health", handleHealth)

	fmt.Println("Starting server on :8080 with PEP protection...")
	http.ListenAndServe(":8080", mux)
}

// ExampleServiceIntegration demonstrates PEP integration with business services
func ExampleServiceIntegration() {
	// Initialize components
	mockStorage, _ := storage.NewMockStorage(".")
	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)
	auditLogger, _ := NewSimpleAuditLogger("./audit.log")

	config := &PEPConfig{
		CacheEnabled:     false,
		FailSafeMode:     true,
		AuditEnabled:     true,
		RateLimitEnabled: false,
	}

	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)

	// Create secure service
	userService := NewSecureUserService(pep)

	// Example usage
	ctx := context.Background()

	// Try to get user (should check permissions)
	user, err := userService.GetUser(ctx, "sub-001", "user-123")
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
	} else {
		fmt.Printf("Got user: %v\n", user)
	}

	// Try to update user (should check permissions)
	err = userService.UpdateUser(ctx, "sub-001", "user-123", map[string]interface{}{
		"name": "Updated Name",
	})
	if err != nil {
		fmt.Printf("Failed to update user: %v\n", err)
	} else {
		fmt.Println("User updated successfully")
	}
}

// ExampleDatabaseIntegration demonstrates PEP integration with database operations
func ExampleDatabaseIntegration() {
	// Initialize components
	mockStorage, _ := storage.NewMockStorage(".")
	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)
	auditLogger, _ := NewSimpleAuditLogger("./audit.log")

	config := &PEPConfig{
		CacheEnabled:     false,
		FailSafeMode:     true,
		AuditEnabled:     true,
		RateLimitEnabled: false,
	}

	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, config)

	// Create database service with PEP protection
	dbService := NewSecureDBService(pep)

	ctx := context.Background()

	// Example database operations with access control
	err := dbService.QueryUsers(ctx, "sub-001")
	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
	}

	err = dbService.InsertUser(ctx, "sub-001", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	if err != nil {
		fmt.Printf("Insert failed: %v\n", err)
	}
}

// HTTP handlers
func handleUsers(w http.ResponseWriter, r *http.Request) {
	// Get enforcement result from context
	if result, ok := GetEnforcementResult(r); ok {
		w.Header().Set("X-PEP-Decision", result.Decision.Result)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users": ["user1", "user2"]}`))
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"admin": "panel"}`))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// SecureUserService demonstrates service-level PEP integration
type SecureUserService struct {
	pep *SimplePolicyEnforcementPoint
}

func NewSecureUserService(pep *SimplePolicyEnforcementPoint) *SecureUserService {
	return &SecureUserService{pep: pep}
}

func (s *SecureUserService) GetUser(ctx context.Context, subjectID, userID string) (interface{}, error) {
	// Create evaluation request
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("get_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: fmt.Sprintf("user:%s", userID),
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "get_user",
		},
	}

	// Check access
	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return nil, fmt.Errorf("access denied: %s", result.Reason)
	}

	// Business logic
	return map[string]interface{}{
		"id":   userID,
		"name": "John Doe",
		"role": "user",
	}, nil
}

func (s *SecureUserService) UpdateUser(ctx context.Context, subjectID, userID string, data map[string]interface{}) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("update_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: fmt.Sprintf("user:%s", userID),
		Action:     "update",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "update_user",
			"data":      data,
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Business logic
	fmt.Printf("Updating user %s with data: %v\n", userID, data)
	return nil
}

// SecureDBService demonstrates database-level PEP integration
type SecureDBService struct {
	pep *SimplePolicyEnforcementPoint
}

func NewSecureDBService(pep *SimplePolicyEnforcementPoint) *SecureDBService {
	return &SecureDBService{pep: pep}
}

func (s *SecureDBService) QueryUsers(ctx context.Context, subjectID string) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("query_users_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: "db.users",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "query",
			"table":     "users",
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Database operation
	fmt.Println("Executing: SELECT * FROM users")
	return nil
}

func (s *SecureDBService) InsertUser(ctx context.Context, subjectID string, userData map[string]interface{}) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("insert_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: "db.users",
		Action:     "create",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "insert",
			"table":     "users",
			"data":      userData,
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Database operation
	fmt.Printf("Executing: INSERT INTO users VALUES %v\n", userData)
	return nil
}
