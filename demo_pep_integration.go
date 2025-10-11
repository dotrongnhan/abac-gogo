package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/pep"
	"abac_go_example/storage"
)

// Demo: PEP Integration Examples
// NOTE: This main function has been moved to main.go as runPEPIntegrationDemo()
/*
func main() {
	fmt.Println("🚀 ABAC PEP Integration Demo")
	fmt.Println("============================")

	// Run different integration examples
	runBasicPEPDemo()
	runHTTPMiddlewareDemo()
	runServiceIntegrationDemo()
}
*/

// Basic PEP usage demonstration
func runBasicPEPDemo() {
	fmt.Println("\n📋 1. Basic PEP Usage Demo")
	fmt.Println("--------------------------")

	// Initialize components
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)
	auditLogger, _ := pep.NewSimpleAuditLogger("demo_audit.log")

	// Create simple PEP
	pepInstance := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

	// Test different scenarios
	testScenarios := []struct {
		name       string
		subjectID  string
		resourceID string
		action     string
		expected   string
	}{
		{
			name:       "Engineering user reading API",
			subjectID:  "sub-001", // John Doe - Engineering
			resourceID: "/api/v1/users",
			action:     "read",
			expected:   "permit",
		},
		{
			name:       "Probation user writing data",
			subjectID:  "sub-004", // Bob Wilson - On probation
			resourceID: "/api/v1/users",
			action:     "write",
			expected:   "deny",
		},
		{
			name:       "Finance user accessing financial data",
			subjectID:  "sub-002", // Alice Smith - Finance
			resourceID: "DOC-2024-Q1-FINANCE",
			action:     "read",
			expected:   "permit",
		},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("\n🧪 Testing: %s\n", scenario.name)

		request := &models.EvaluationRequest{
			RequestID:  fmt.Sprintf("demo_%d", time.Now().UnixNano()),
			SubjectID:  scenario.subjectID,
			ResourceID: scenario.resourceID,
			Action:     scenario.action,
			Context: map[string]interface{}{
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"demo":      true,
			},
		}

		result, err := pepInstance.EnforceRequest(context.Background(), request)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		status := "✅"
		if result.Decision.Result != scenario.expected {
			status = "❌"
		}

		fmt.Printf("%s Decision: %s (Expected: %s)\n", status, result.Decision.Result, scenario.expected)
		fmt.Printf("   Reason: %s\n", result.Reason)
		fmt.Printf("   Evaluation Time: %dms\n", result.EvaluationTimeMs)
		fmt.Printf("   Matched Policies: %v\n", result.Decision.MatchedPolicies)
	}

	// Show metrics
	metrics := pepInstance.GetMetrics()
	fmt.Printf("\n📊 PEP Metrics:\n")
	fmt.Printf("   Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("   Permit Decisions: %d\n", metrics.PermitDecisions)
	fmt.Printf("   Deny Decisions: %d\n", metrics.DenyDecisions)
	fmt.Printf("   Validation Errors: %d\n", metrics.ValidationErrors)
}

// HTTP Middleware integration demonstration
func runHTTPMiddlewareDemo() {
	fmt.Println("\n🌐 2. HTTP Middleware Demo")
	fmt.Println("--------------------------")

	// Setup PEP
	mockStorage, _ := storage.NewMockStorage(".")
	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)
	auditLogger, _ := pep.NewSimpleAuditLogger("demo_middleware_audit.log")
	pepInstance := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

	// Create middleware
	middleware := pep.NewHTTPMiddleware(pepInstance, nil)

	// Setup routes
	mux := http.NewServeMux()

	// Protected routes
	mux.Handle("/api/users", middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get enforcement result from context
		if result, ok := pep.GetEnforcementResult(r); ok {
			w.Header().Set("X-PEP-Decision", result.Decision.Result)
			w.Header().Set("X-PEP-Evaluation-Time", fmt.Sprintf("%dms", result.EvaluationTimeMs))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Users endpoint accessed successfully", "users": ["user1", "user2"]}`))
	})))

	mux.Handle("/api/admin", middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Admin panel accessed", "admin": true}`))
	})))

	// Unprotected routes
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
	})

	fmt.Println("🚀 HTTP Server with PEP middleware started on :8080")
	fmt.Println("📝 Test endpoints:")
	fmt.Println("   GET /health (unprotected)")
	fmt.Println("   GET /api/users (protected - add 'X-Subject-ID: sub-001' header)")
	fmt.Println("   GET /api/admin (protected - add 'X-Subject-ID: sub-002' header)")
	fmt.Println("   POST /api/users (protected - will be denied for probation users)")
	fmt.Println("\n💡 Example curl commands:")
	fmt.Println("   curl http://localhost:8080/health")
	fmt.Println("   curl -H 'X-Subject-ID: sub-001' http://localhost:8080/api/users")
	fmt.Println("   curl -H 'X-Subject-ID: sub-004' -X POST http://localhost:8080/api/users")

	// Start server in a goroutine for demo
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a bit to let server start
	time.Sleep(time.Second)
	fmt.Println("✅ Server started successfully!")
}

// Service integration demonstration
func runServiceIntegrationDemo() {
	fmt.Println("\n🔧 3. Service Integration Demo")
	fmt.Println("------------------------------")

	// Setup PEP
	mockStorage, _ := storage.NewMockStorage(".")
	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)
	auditLogger, _ := pep.NewSimpleAuditLogger("demo_service_audit.log")
	pepInstance := pep.NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

	// Create secure services
	userService := NewDemoUserService(pepInstance)
	dbService := NewDemoDBService(pepInstance)

	ctx := context.Background()

	fmt.Println("\n🧪 Testing User Service:")

	// Test user service operations
	testUserOperations := []struct {
		operation string
		subjectID string
		userID    string
		action    func() error
	}{
		{
			operation: "Get User (Engineering user)",
			subjectID: "sub-001",
			userID:    "user-123",
			action: func() error {
				_, err := userService.GetUser(ctx, "sub-001", "user-123")
				return err
			},
		},
		{
			operation: "Update User (Engineering user)",
			subjectID: "sub-001",
			userID:    "user-123",
			action: func() error {
				return userService.UpdateUser(ctx, "sub-001", "user-123", map[string]interface{}{
					"name": "Updated Name",
				})
			},
		},
		{
			operation: "Delete User (Probation user - should be denied)",
			subjectID: "sub-004",
			userID:    "user-123",
			action: func() error {
				return userService.DeleteUser(ctx, "sub-004", "user-123")
			},
		},
	}

	for _, test := range testUserOperations {
		fmt.Printf("\n🔍 %s\n", test.operation)
		err := test.action()
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Success\n")
		}
	}

	fmt.Println("\n🧪 Testing Database Service:")

	// Test database service operations
	testDBOperations := []struct {
		operation string
		subjectID string
		action    func() error
	}{
		{
			operation: "Query Users (Engineering user)",
			subjectID: "sub-001",
			action: func() error {
				return dbService.QueryUsers(ctx, "sub-001")
			},
		},
		{
			operation: "Insert User (Finance user)",
			subjectID: "sub-002",
			action: func() error {
				return dbService.InsertUser(ctx, "sub-002", map[string]interface{}{
					"name":  "New User",
					"email": "newuser@example.com",
				})
			},
		},
		{
			operation: "Delete User (Probation user - should be denied)",
			subjectID: "sub-004",
			action: func() error {
				return dbService.DeleteUser(ctx, "sub-004", "user-456")
			},
		},
	}

	for _, test := range testDBOperations {
		fmt.Printf("\n🔍 %s\n", test.operation)
		err := test.action()
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Success\n")
		}
	}

	fmt.Println("\n🎉 Demo completed! Check audit logs for detailed access records.")
}

// Demo User Service with PEP integration
type DemoUserService struct {
	pep *pep.SimplePolicyEnforcementPoint
}

func NewDemoUserService(pepInstance *pep.SimplePolicyEnforcementPoint) *DemoUserService {
	return &DemoUserService{pep: pepInstance}
}

func (s *DemoUserService) GetUser(ctx context.Context, subjectID, userID string) (interface{}, error) {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("get_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: fmt.Sprintf("user:%s", userID),
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "get_user",
			"service":   "user_service",
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return nil, fmt.Errorf("access denied: %s", result.Reason)
	}

	// Simulate business logic
	return map[string]interface{}{
		"id":   userID,
		"name": "Demo User",
		"role": "user",
	}, nil
}

func (s *DemoUserService) UpdateUser(ctx context.Context, subjectID, userID string, data map[string]interface{}) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("update_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: fmt.Sprintf("user:%s", userID),
		Action:     "update",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "update_user",
			"service":   "user_service",
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

	// Simulate business logic
	fmt.Printf("   📝 Updating user %s with data: %v\n", userID, data)
	return nil
}

func (s *DemoUserService) DeleteUser(ctx context.Context, subjectID, userID string) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("delete_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: fmt.Sprintf("user:%s", userID),
		Action:     "delete",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "delete_user",
			"service":   "user_service",
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Simulate business logic
	fmt.Printf("   🗑️ Deleting user %s\n", userID)
	return nil
}

// Demo Database Service with PEP integration
type DemoDBService struct {
	pep *pep.SimplePolicyEnforcementPoint
}

func NewDemoDBService(pepInstance *pep.SimplePolicyEnforcementPoint) *DemoDBService {
	return &DemoDBService{pep: pepInstance}
}

func (s *DemoDBService) QueryUsers(ctx context.Context, subjectID string) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("query_users_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: "db.users",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "query",
			"table":     "users",
			"service":   "db_service",
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Simulate database operation
	fmt.Println("   📊 Executing: SELECT * FROM users")
	return nil
}

func (s *DemoDBService) InsertUser(ctx context.Context, subjectID string, userData map[string]interface{}) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("insert_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: "db.users",
		Action:     "create",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "insert",
			"table":     "users",
			"service":   "db_service",
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

	// Simulate database operation
	fmt.Printf("   📝 Executing: INSERT INTO users VALUES %v\n", userData)
	return nil
}

func (s *DemoDBService) DeleteUser(ctx context.Context, subjectID, userID string) error {
	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("delete_user_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: "db.users",
		Action:     "delete",
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "delete",
			"table":     "users",
			"service":   "db_service",
			"user_id":   userID,
		},
	}

	result, err := s.pep.EnforceRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("access control check failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied: %s", result.Reason)
	}

	// Simulate database operation
	fmt.Printf("   🗑️ Executing: DELETE FROM users WHERE id = %s\n", userID)
	return nil
}
