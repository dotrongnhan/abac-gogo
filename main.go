package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"abac_go_example/audit"
	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/pep"
	"abac_go_example/storage"
)

func main() {
	fmt.Println("🚀 ABAC System - Unified Demo & Management Tool")
	fmt.Println("===============================================")

	for {
		showMainMenu()
		choice := getUserInput("Select an option (1-5): ")

		switch choice {
		case "1":
			runPolicyEvaluationDemo()
		case "2":
			runPEPIntegrationDemo()
		case "3":
			runDatabaseMigrationAndSeeding()
		case "4":
			runInteractiveMode()
		case "5":
			fmt.Println("👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid option. Please try again.")
		}

		fmt.Println("\nPress Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func showMainMenu() {
	fmt.Println("\n📋 Main Menu")
	fmt.Println("============")
	fmt.Println("1. 🧪 Policy Evaluation Demo (Original main.go)")
	fmt.Println("2. 🔧 PEP Integration Demo (demo_pep_integration.go)")
	fmt.Println("3. 🗄️  Database Migration & Seeding (cmd/migrate/main.go)")
	fmt.Println("4. 🎮 Interactive Mode")
	fmt.Println("5. 🚪 Exit")
	fmt.Println()
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// ===== Policy Evaluation Demo (from main.go) =====
func runPolicyEvaluationDemo() {
	fmt.Println("\n🧪 Running Policy Evaluation Demo")
	fmt.Println("==================================")

	// Initialize PostgreSQL storage
	config := storage.DefaultDatabaseConfig()
	pgStorage, err := storage.NewPostgreSQLStorage(config)
	if err != nil {
		log.Printf("Failed to initialize PostgreSQL storage: %v", err)
		return
	}
	defer pgStorage.Close()

	fmt.Println("✅ PostgreSQL storage initialized successfully")

	// Initialize audit logger
	auditLogger, err := audit.NewAuditLogger("audit.log")
	if err != nil {
		log.Printf("Failed to initialize audit logger: %v", err)
		return
	}
	defer auditLogger.Close()

	// Initialize Policy Decision Point
	pdp := evaluator.NewPolicyDecisionPoint(pgStorage)

	// Load evaluation requests from JSON
	evaluationRequests, err := loadEvaluationRequests("evaluation_requests.json")
	if err != nil {
		log.Printf("Failed to load evaluation requests: %v", err)
		return
	}

	fmt.Printf("\n📋 Running %d evaluation scenarios...\n\n", len(evaluationRequests))

	// Process each evaluation request
	for i, request := range evaluationRequests {
		fmt.Printf("🔍 Scenario %d: %s\n", i+1, request.RequestID)
		fmt.Printf("   Subject: %s\n", request.SubjectID)
		fmt.Printf("   Resource: %s\n", request.ResourceID)
		fmt.Printf("   Action: %s\n", request.Action)

		// Perform evaluation
		decision, err := pdp.Evaluate(request)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}

		// Get full context for audit logging
		subject, _ := pgStorage.GetSubject(request.SubjectID)
		resource, _ := pgStorage.GetResource(request.ResourceID)
		action, _ := pgStorage.GetAction(request.Action)

		auditContext := &models.EvaluationContext{
			Subject:     subject,
			Resource:    resource,
			Action:      action,
			Environment: request.Context,
			Timestamp:   time.Now(),
		}

		// Log the evaluation
		auditLogger.LogEvaluation(request, decision, auditContext)

		// Display results
		fmt.Printf("   📊 Decision: %s\n", decision.Result)
		fmt.Printf("   ⏱️  Evaluation Time: %dms\n", decision.EvaluationTimeMs)
		fmt.Printf("   📝 Reason: %s\n", decision.Reason)

		if len(decision.MatchedPolicies) > 0 {
			fmt.Printf("   🎯 Matched Policies: %v\n", decision.MatchedPolicies)
		}

		// Check expected result if available
		if expectedDecision, ok := request.Context["expected_decision"].(string); ok {
			if decision.Result == expectedDecision {
				fmt.Printf("   ✅ Expected: %s (PASS)\n", expectedDecision)
			} else {
				fmt.Printf("   ❌ Expected: %s, Got: %s (FAIL)\n", expectedDecision, decision.Result)
			}
		}

		fmt.Println()
	}

	// Display system statistics
	fmt.Println("📈 System Statistics")
	fmt.Println("===================")

	subjects, _ := pgStorage.GetAllSubjects()
	resources, _ := pgStorage.GetAllResources()
	actions, _ := pgStorage.GetAllActions()
	policies, _ := pgStorage.GetPolicies()

	fmt.Printf("Subjects: %d\n", len(subjects))
	fmt.Printf("Resources: %d\n", len(resources))
	fmt.Printf("Actions: %d\n", len(actions))
	fmt.Printf("Policies: %d\n", len(policies))

	// Test detailed explanation for first request
	if len(evaluationRequests) > 0 {
		fmt.Println("\n🔬 Detailed Explanation for First Request")
		fmt.Println("=========================================")

		explanation, err := pdp.ExplainDecision(evaluationRequests[0])
		if err != nil {
			fmt.Printf("Error getting explanation: %v\n", err)
		} else {
			explanationJSON, _ := json.MarshalIndent(explanation, "", "  ")
			fmt.Println(string(explanationJSON))
		}
	}

	// Test batch evaluation
	fmt.Println("\n⚡ Batch Evaluation Test")
	fmt.Println("=======================")

	startTime := time.Now()
	decisions, err := pdp.BatchEvaluate(evaluationRequests)
	batchTime := time.Since(startTime)

	if err != nil {
		fmt.Printf("Batch evaluation error: %v\n", err)
	} else {
		fmt.Printf("Batch processed %d requests in %v\n", len(decisions), batchTime)
		fmt.Printf("Average time per request: %v\n", batchTime/time.Duration(len(decisions)))
	}

	// Test security scenarios
	fmt.Println("\n🔐 Security Scenarios")
	fmt.Println("====================")

	testSecurityScenarios(pdp, auditLogger)

	fmt.Println("\n✅ Policy Evaluation Demo Complete!")
}

// ===== PEP Integration Demo (from demo_pep_integration.go) =====
func runPEPIntegrationDemo() {
	fmt.Println("\n🔧 Running PEP Integration Demo")
	fmt.Println("===============================")

	// Run different integration examples
	runBasicPEPDemoLocal()
	runHTTPMiddlewareDemoLocal()
	runServiceIntegrationDemoLocal()
}

// Basic PEP usage demonstration (local implementation)
func runBasicPEPDemoLocal() {
	fmt.Println("\n📋 1. Basic PEP Usage Demo")
	fmt.Println("--------------------------")

	// Initialize components
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		log.Printf("Failed to initialize storage: %v", err)
		return
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

// HTTP Middleware integration demonstration (local implementation)
func runHTTPMiddlewareDemoLocal() {
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

// Service integration demonstration (local implementation)
func runServiceIntegrationDemoLocal() {
	fmt.Println("\n🔧 3. Service Integration Demo")
	fmt.Println("------------------------------")

	// Create secure services using the services from demo_pep_integration.go
	fmt.Println("⚠️  Note: Service integration demo requires DemoUserService and DemoDBService")
	fmt.Println("These are defined in demo_pep_integration.go")
	fmt.Println("For a complete demo, please run the PEP integration demo separately.")

	fmt.Println("\n🎉 Demo completed! Check audit logs for detailed access records.")
}

// ===== Database Migration & Seeding (from cmd/migrate/main.go) =====
func runDatabaseMigrationAndSeeding() {
	fmt.Println("\n🗄️ Running Database Migration & Seeding")
	fmt.Println("=======================================")

	// Initialize PostgreSQL storage
	config := storage.DefaultDatabaseConfig()
	pgStorage, err := storage.NewPostgreSQLStorage(config)
	if err != nil {
		log.Printf("Failed to initialize PostgreSQL storage: %v", err)
		return
	}
	defer pgStorage.Close()

	fmt.Println("✅ Database connection established and tables migrated")

	// Seed data from JSON files
	if err := seedData(pgStorage, "."); err != nil {
		log.Printf("Failed to seed data: %v", err)
		return
	}

	fmt.Println("✅ Data seeding completed successfully")
}

// ===== Interactive Mode =====
func runInteractiveMode() {
	fmt.Println("\n🎮 Interactive Mode")
	fmt.Println("==================")

	// Initialize components
	config := storage.DefaultDatabaseConfig()
	pgStorage, err := storage.NewPostgreSQLStorage(config)
	if err != nil {
		log.Printf("Failed to initialize PostgreSQL storage: %v", err)
		return
	}
	defer pgStorage.Close()

	pdp := evaluator.NewPolicyDecisionPoint(pgStorage)

	for {
		fmt.Println("\n📋 Interactive Options:")
		fmt.Println("1. 🔍 Evaluate Custom Request")
		fmt.Println("2. 📊 View System Statistics")
		fmt.Println("3. 🔍 List Subjects")
		fmt.Println("4. 🔍 List Resources")
		fmt.Println("5. 🔍 List Policies")
		fmt.Println("6. 🔙 Back to Main Menu")

		choice := getUserInput("Select option: ")

		switch choice {
		case "1":
			evaluateCustomRequest(pdp)
		case "2":
			showSystemStatistics(pgStorage)
		case "3":
			listSubjects(pgStorage)
		case "4":
			listResources(pgStorage)
		case "5":
			listPolicies(pgStorage)
		case "6":
			return
		default:
			fmt.Println("❌ Invalid option")
		}
	}
}

func evaluateCustomRequest(pdp *evaluator.PolicyDecisionPoint) {
	fmt.Println("\n🔍 Custom Request Evaluation")
	fmt.Println("============================")

	subjectID := getUserInput("Enter Subject ID: ")
	resourceID := getUserInput("Enter Resource ID: ")
	action := getUserInput("Enter Action: ")

	request := &models.EvaluationRequest{
		RequestID:  fmt.Sprintf("interactive_%d", time.Now().UnixNano()),
		SubjectID:  subjectID,
		ResourceID: resourceID,
		Action:     action,
		Context: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"mode":      "interactive",
		},
	}

	fmt.Println("\n⏳ Evaluating request...")
	decision, err := pdp.Evaluate(request)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("   Decision: %s\n", decision.Result)
	fmt.Printf("   Reason: %s\n", decision.Reason)
	fmt.Printf("   Evaluation Time: %dms\n", decision.EvaluationTimeMs)
	if len(decision.MatchedPolicies) > 0 {
		fmt.Printf("   Matched Policies: %v\n", decision.MatchedPolicies)
	}
}

func showSystemStatistics(pgStorage *storage.PostgreSQLStorage) {
	fmt.Println("\n📈 System Statistics")
	fmt.Println("===================")

	subjects, _ := pgStorage.GetAllSubjects()
	resources, _ := pgStorage.GetAllResources()
	actions, _ := pgStorage.GetAllActions()
	policies, _ := pgStorage.GetPolicies()

	fmt.Printf("📊 Total Subjects: %d\n", len(subjects))
	fmt.Printf("📊 Total Resources: %d\n", len(resources))
	fmt.Printf("📊 Total Actions: %d\n", len(actions))
	fmt.Printf("📊 Total Policies: %d\n", len(policies))
}

func listSubjects(pgStorage *storage.PostgreSQLStorage) {
	fmt.Println("\n👥 Subjects List")
	fmt.Println("===============")

	subjects, err := pgStorage.GetAllSubjects()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	for _, subject := range subjects {
		fmt.Printf("🆔 %s - %s (%s)\n", subject.ID, subject.ExternalID, subject.SubjectType)
		if len(subject.Attributes) > 0 {
			fmt.Printf("   Attributes: %v\n", subject.Attributes)
		}
	}
}

func listResources(pgStorage *storage.PostgreSQLStorage) {
	fmt.Println("\n📁 Resources List")
	fmt.Println("================")

	resources, err := pgStorage.GetAllResources()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	for _, resource := range resources {
		fmt.Printf("🆔 %s - %s (%s)\n", resource.ID, resource.ResourceID, resource.ResourceType)
		if resource.Path != "" {
			fmt.Printf("   Path: %s\n", resource.Path)
		}
	}
}

func listPolicies(pgStorage *storage.PostgreSQLStorage) {
	fmt.Println("\n📋 Policies List")
	fmt.Println("===============")

	policies, err := pgStorage.GetPolicies()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	for _, policy := range policies {
		status := "✅"
		if !policy.Enabled {
			status = "❌"
		}
		fmt.Printf("%s %s - %s (Priority: %d)\n", status, policy.ID, policy.PolicyName, policy.Priority)
		fmt.Printf("   Effect: %s | Version: %d\n", policy.Effect, policy.Version)
		fmt.Printf("   Description: %s\n", policy.Description)
	}
}

// ===== Helper Functions =====

func loadEvaluationRequests(filename string) ([]*models.EvaluationRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var requestsData struct {
		EvaluationRequests []*models.EvaluationRequest `json:"evaluation_requests"`
	}

	if err := json.Unmarshal(data, &requestsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Add expected decisions from the JSON data
	for _, req := range requestsData.EvaluationRequests {
		if req.Context == nil {
			req.Context = make(map[string]interface{})
		}

		// Look for expected_decision in the original JSON structure
		var originalData map[string]interface{}
		json.Unmarshal(data, &originalData)

		if evalRequests, ok := originalData["evaluation_requests"].([]interface{}); ok {
			for _, evalReq := range evalRequests {
				if reqMap, ok := evalReq.(map[string]interface{}); ok {
					if reqMap["request_id"] == req.RequestID {
						if expectedDecision, exists := reqMap["expected_decision"]; exists {
							req.Context["expected_decision"] = expectedDecision
						}
						if matchedPolicies, exists := reqMap["matched_policies"]; exists {
							req.Context["expected_matched_policies"] = matchedPolicies
						}
						break
					}
				}
			}
		}
	}

	return requestsData.EvaluationRequests, nil
}

func testSecurityScenarios(pdp *evaluator.PolicyDecisionPoint, auditLogger *audit.AuditLogger) {
	// Test 1: Probation user trying to write
	fmt.Println("1. Testing probation user write access...")
	probationRequest := &models.EvaluationRequest{
		RequestID:  "security-test-1",
		SubjectID:  "sub-004", // Bob Wilson (on probation)
		ResourceID: "res-002", // Production database
		Action:     "write",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"source_ip": "10.0.1.100",
		},
	}

	decision, err := pdp.Evaluate(probationRequest)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %s (Expected: deny)\n", decision.Result)
		fmt.Printf("   Reason: %s\n", decision.Reason)

		if decision.Result == "deny" {
			auditLogger.LogSecurityEvent("probation_write_attempt", "sub-004", map[string]interface{}{
				"resource": "res-002",
				"action":   "write",
				"blocked":  true,
			})
		}
	}

	// Test 2: After hours access attempt
	fmt.Println("\n2. Testing after-hours access...")
	afterHoursRequest := &models.EvaluationRequest{
		RequestID:  "security-test-2",
		SubjectID:  "sub-001", // John Doe (senior developer)
		ResourceID: "res-001", // API endpoint
		Action:     "write",
		Context: map[string]interface{}{
			"timestamp": "2024-01-15T22:00:00Z", // 10 PM
			"source_ip": "10.0.1.50",
		},
	}

	decision, err = pdp.Evaluate(afterHoursRequest)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", decision.Result)
		fmt.Printf("   Reason: %s\n", decision.Reason)
	}

	// Test 3: External IP access attempt
	fmt.Println("\n3. Testing external IP access...")
	externalIPRequest := &models.EvaluationRequest{
		RequestID:  "security-test-3",
		SubjectID:  "sub-003", // Payment service
		ResourceID: "res-001", // API endpoint
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"source_ip": "203.0.113.1", // External IP
		},
	}

	decision, err = pdp.Evaluate(externalIPRequest)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", decision.Result)
		fmt.Printf("   Reason: %s\n", decision.Reason)

		if decision.Result == "deny" {
			auditLogger.LogSecurityEvent("external_ip_access", "sub-003", map[string]interface{}{
				"source_ip": "203.0.113.1",
				"resource":  "res-001",
				"blocked":   true,
			})
		}
	}

	// Test 4: Privilege escalation attempt
	fmt.Println("\n4. Testing privilege escalation...")
	escalationRequest := &models.EvaluationRequest{
		RequestID:  "security-test-4",
		SubjectID:  "sub-004", // Junior developer
		ResourceID: "res-003", // Financial document
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"source_ip": "10.0.1.200",
		},
	}

	decision, err = pdp.Evaluate(escalationRequest)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", decision.Result)
		fmt.Printf("   Reason: %s\n", decision.Reason)

		if decision.Result == "deny" {
			auditLogger.LogSecurityEvent("privilege_escalation_attempt", "sub-004", map[string]interface{}{
				"attempted_resource": "res-003",
				"resource_type":      "financial_document",
				"user_clearance":     1,
				"required_clearance": 2,
			})
		}
	}
}

// Note: PEP demo functions are defined in demo_pep_integration.go

// Database seeding functions
func seedData(storage *storage.PostgreSQLStorage, dataDir string) error {
	// Seed subjects
	if err := seedSubjects(storage, filepath.Join(dataDir, "subjects.json")); err != nil {
		return fmt.Errorf("failed to seed subjects: %w", err)
	}

	// Seed resources
	if err := seedResources(storage, filepath.Join(dataDir, "resources.json")); err != nil {
		return fmt.Errorf("failed to seed resources: %w", err)
	}

	// Seed actions
	if err := seedActions(storage, filepath.Join(dataDir, "actions.json")); err != nil {
		return fmt.Errorf("failed to seed actions: %w", err)
	}

	// Seed policies
	if err := seedPolicies(storage, filepath.Join(dataDir, "policies.json")); err != nil {
		return fmt.Errorf("failed to seed policies: %w", err)
	}

	return nil
}

func seedSubjects(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding subjects from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var subjectsData struct {
		Subjects []struct {
			ID          string                 `json:"id"`
			ExternalID  string                 `json:"external_id"`
			SubjectType string                 `json:"subject_type"`
			Metadata    map[string]interface{} `json:"metadata"`
			Attributes  map[string]interface{} `json:"attributes"`
		} `json:"subjects"`
	}

	if err := json.Unmarshal(data, &subjectsData); err != nil {
		return err
	}

	for _, subjectData := range subjectsData.Subjects {
		subject := &models.Subject{
			ID:          subjectData.ID,
			ExternalID:  subjectData.ExternalID,
			SubjectType: subjectData.SubjectType,
			Metadata:    models.JSONMap(subjectData.Metadata),
			Attributes:  models.JSONMap(subjectData.Attributes),
		}

		if err := storage.CreateSubject(subject); err != nil {
			// If subject already exists, update it
			if err := storage.UpdateSubject(subject); err != nil {
				return fmt.Errorf("failed to create/update subject %s: %w", subject.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d subjects\n", len(subjectsData.Subjects))
	return nil
}

func seedResources(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding resources from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var resourcesData struct {
		Resources []struct {
			ID           string                 `json:"id"`
			ResourceType string                 `json:"resource_type"`
			ResourceID   string                 `json:"resource_id"`
			Path         string                 `json:"path"`
			ParentID     string                 `json:"parent_id,omitempty"`
			Metadata     map[string]interface{} `json:"metadata"`
			Attributes   map[string]interface{} `json:"attributes"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(data, &resourcesData); err != nil {
		return err
	}

	for _, resourceData := range resourcesData.Resources {
		resource := &models.Resource{
			ID:           resourceData.ID,
			ResourceType: resourceData.ResourceType,
			ResourceID:   resourceData.ResourceID,
			Path:         resourceData.Path,
			ParentID:     resourceData.ParentID,
			Metadata:     models.JSONMap(resourceData.Metadata),
			Attributes:   models.JSONMap(resourceData.Attributes),
		}

		if err := storage.CreateResource(resource); err != nil {
			// If resource already exists, update it
			if err := storage.UpdateResource(resource); err != nil {
				return fmt.Errorf("failed to create/update resource %s: %w", resource.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d resources\n", len(resourcesData.Resources))
	return nil
}

func seedActions(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding actions from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var actionsData struct {
		Actions []struct {
			ID             string `json:"id"`
			ActionName     string `json:"action_name"`
			ActionCategory string `json:"action_category"`
			Description    string `json:"description"`
			IsSystem       bool   `json:"is_system"`
		} `json:"actions"`
	}

	if err := json.Unmarshal(data, &actionsData); err != nil {
		return err
	}

	for _, actionData := range actionsData.Actions {
		action := &models.Action{
			ID:             actionData.ID,
			ActionName:     actionData.ActionName,
			ActionCategory: actionData.ActionCategory,
			Description:    actionData.Description,
			IsSystem:       actionData.IsSystem,
		}

		if err := storage.CreateAction(action); err != nil {
			// If action already exists, update it
			if err := storage.UpdateAction(action); err != nil {
				return fmt.Errorf("failed to create/update action %s: %w", action.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d actions\n", len(actionsData.Actions))
	return nil
}

func seedPolicies(storage *storage.PostgreSQLStorage, filename string) error {
	fmt.Printf("📥 Seeding policies from %s...\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var policiesData struct {
		Policies []struct {
			ID               string                 `json:"id"`
			PolicyName       string                 `json:"policy_name"`
			Description      string                 `json:"description"`
			Effect           string                 `json:"effect"`
			Priority         int                    `json:"priority"`
			Enabled          bool                   `json:"enabled"`
			Version          int                    `json:"version"`
			Conditions       map[string]interface{} `json:"conditions"`
			Rules            []models.PolicyRule    `json:"rules"`
			Actions          []string               `json:"actions"`
			ResourcePatterns []string               `json:"resource_patterns"`
		} `json:"policies"`
	}

	if err := json.Unmarshal(data, &policiesData); err != nil {
		return err
	}

	for _, policyData := range policiesData.Policies {
		policy := &models.Policy{
			ID:               policyData.ID,
			PolicyName:       policyData.PolicyName,
			Description:      policyData.Description,
			Effect:           policyData.Effect,
			Priority:         policyData.Priority,
			Enabled:          policyData.Enabled,
			Version:          policyData.Version,
			Conditions:       models.JSONMap(policyData.Conditions),
			Rules:            models.JSONPolicyRules(policyData.Rules),
			Actions:          models.JSONStringSlice(policyData.Actions),
			ResourcePatterns: models.JSONStringSlice(policyData.ResourcePatterns),
		}

		if err := storage.CreatePolicy(policy); err != nil {
			// If policy already exists, update it
			if err := storage.UpdatePolicy(policy); err != nil {
				return fmt.Errorf("failed to create/update policy %s: %w", policy.ID, err)
			}
		}
	}

	fmt.Printf("✅ Seeded %d policies\n", len(policiesData.Policies))
	return nil
}

// Note: Demo service types are defined in demo_pep_integration.go
