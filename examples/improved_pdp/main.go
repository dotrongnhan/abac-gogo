package main

import (
	"fmt"
	"log"
	"time"

	"abac_go_example/evaluator/core"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func main() {
	fmt.Println("🚀 Improved PDP Example - Demonstrating Enhanced Features")

	// Initialize storage
	dbConfig := &storage.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "password",
		DatabaseName: "abac_db",
		SSLMode:      "disable",
		TimeZone:     "UTC",
	}

	storage, err := storage.NewPostgreSQLStorage(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Create improved PDP
	pdp := core.NewPolicyDecisionPoint(storage)

	// Example 1: Time-based attributes (Improvement #4)
	fmt.Println("\n=== Example 1: Time-based Attributes ===")
	timeBasedExampleImproved(pdp)

	// Example 2: Environmental context (Improvement #5)
	fmt.Println("\n=== Example 2: Environmental Context ===")
	environmentalContextExample(pdp)

	// Example 3: Structured attributes (Improvement #6)
	fmt.Println("\n=== Example 3: Structured Attributes ===")
	structuredAttributesExample(pdp)

	// Example 4: Enhanced condition evaluation (Improvement #7)
	fmt.Println("\n=== Example 4: Enhanced Condition Evaluation ===")
	enhancedConditionExample(pdp)

	// Example 5: Policy filtering performance (Improvement #8)
	fmt.Println("\n=== Example 5: Policy Filtering Performance ===")
	policyFilteringExample(pdp)

	// Example 6: Pre-filtering optimization (Improvement #12)
	fmt.Println("\n=== Example 6: Pre-filtering Optimization ===")
	preFilteringExample(pdp)
}

func timeBasedExampleImproved(pdp core.PolicyDecisionPointInterface) {
	// Create request with time-based attributes
	now := time.Now()
	request := &models.EvaluationRequest{
		RequestID:  "time-001",
		Subject:    models.NewMockUserSubject("user123", "user123"),
		ResourceID: "/api/reports",
		Action:     "read",
		Timestamp:  &now, // Enhanced: explicit timestamp
		Environment: &models.EnvironmentInfo{
			TimeOfDay: now.Format("15:04"),
			DayOfWeek: now.Weekday().String(),
		},
		Context: map[string]interface{}{
			"session_id": "sess_123",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Time-based Decision: %s\n", decision.Result)
	fmt.Printf("Reason: %s\n", decision.Reason)
	fmt.Printf("Evaluation Time: %dms\n", decision.EvaluationTimeMs)

	// The enhanced PDP now automatically adds:
	// - environment:time_of_day
	// - environment:day_of_week
	// - environment:hour
	// - environment:minute
	// - environment:is_weekend
	// - environment:is_business_hours
}

func environmentalContextExample(pdp core.PolicyDecisionPointInterface) {
	// Create request with rich environmental context
	request := &models.EvaluationRequest{
		RequestID:  "env-001",
		Subject:    models.NewMockUserSubject("user456", "user456"),
		ResourceID: "/api/financial/reports",
		Action:     "read",
		Environment: &models.EnvironmentInfo{
			ClientIP:  "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			Country:   "Vietnam",
			Region:    "Ho Chi Minh City",
			Attributes: map[string]interface{}{
				"device_type":   "desktop",
				"connection":    "wifi",
				"screen_size":   "1920x1080",
				"vpn_connected": false,
			},
		},
		Context: map[string]interface{}{
			"session_id":   "sess_456",
			"mfa_verified": true,
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Environmental Decision: %s\n", decision.Result)
	fmt.Printf("Reason: %s\n", decision.Reason)

	// The enhanced PDP now automatically adds:
	// - environment:client_ip
	// - environment:user_agent
	// - environment:country
	// - environment:region
	// - environment:is_internal_ip
	// - environment:ip_class
	// - environment:is_mobile
	// - environment:browser
}

func structuredAttributesExample(pdp core.PolicyDecisionPointInterface) {
	// Create request that will benefit from structured attributes
	request := &models.EvaluationRequest{
		RequestID:  "struct-001",
		Subject:    models.NewMockUserSubject("user789", "user789"),
		ResourceID: "/documents/confidential/project-alpha.pdf",
		Action:     "read",
		Context: map[string]interface{}{
			"department": "Engineering",
			"clearance":  "confidential",
			"project":    "alpha",
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Structured Attributes Decision: %s\n", decision.Result)
	fmt.Printf("Reason: %s\n", decision.Reason)

	// The enhanced PDP now provides both:
	// Flat access: user:department, resource:classification
	// Structured access: user.department, resource.classification
	// This allows policies to use dot notation for nested attributes
}

func enhancedConditionExample(pdp core.PolicyDecisionPointInterface) {
	// This example shows how enhanced conditions work
	// The actual policy would be stored in database with enhanced operators

	fmt.Println("Enhanced condition operators now supported:")
	fmt.Println("- String: StringContains, StringStartsWith, StringEndsWith, StringRegex")
	fmt.Println("- Numeric: NumericBetween, NumericNotEquals")
	fmt.Println("- Time: TimeOfDay, DayOfWeek, IsBusinessHours, TimeBetween")
	fmt.Println("- Network: IPInRange, IPNotInRange, IsInternalIP")
	fmt.Println("- Array: ArrayContains, ArraySize")
	fmt.Println("- Complex: And, Or, Not operators for boolean logic")

	// Example request that would use enhanced conditions
	request := &models.EvaluationRequest{
		RequestID:  "enhanced-001",
		Subject:    models.NewMockUserSubject("user999", "user999"),
		ResourceID: "/api/admin/users",
		Action:     "write",
		Environment: &models.EnvironmentInfo{
			ClientIP:  "10.0.1.50",
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
			TimeOfDay: "14:30",
			DayOfWeek: "Wednesday",
		},
		Context: map[string]interface{}{
			"user_level": 8,
			"roles":      []string{"admin", "developer"},
		},
	}

	decision, err := pdp.Evaluate(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Enhanced Condition Decision: %s\n", decision.Result)
	fmt.Printf("Reason: %s\n", decision.Reason)
}

func policyFilteringExample(pdp core.PolicyDecisionPointInterface) {
	fmt.Println("Policy Filtering Performance Improvements:")
	fmt.Println("- Smart pre-filtering reduces policies to evaluate")
	fmt.Println("- Pattern matching cache for repeated evaluations")
	fmt.Println("- Fast wildcard matching for common patterns")
	fmt.Println("- Subject/Resource/Action type filtering")

	// Simulate multiple requests to show filtering benefits
	requests := []*models.EvaluationRequest{
		{
			RequestID:  "filter-001",
			Subject:    models.NewMockUserSubject("user1", "user1"),
			ResourceID: "/api/users",
			Action:     "read",
		},
		{
			RequestID:  "filter-002",
			Subject:    models.NewMockUserSubject("user2", "user2"),
			ResourceID: "/api/reports",
			Action:     "read",
		},
		{
			RequestID:  "filter-003",
			Subject:    models.NewMockUserSubject("user3", "user3"),
			ResourceID: "/api/admin",
			Action:     "write",
		},
	}

	totalTime := 0
	for i, request := range requests {
		start := time.Now()
		decision, err := pdp.Evaluate(request)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("Request %d Error: %v\n", i+1, err)
			continue
		}

		fmt.Printf("Request %d: %s (%v)\n", i+1, decision.Result, elapsed)
		totalTime += decision.EvaluationTimeMs
	}

	fmt.Printf("Total evaluation time: %dms\n", totalTime)
	fmt.Printf("Average per request: %dms\n", totalTime/len(requests))
}

func preFilteringExample(pdp core.PolicyDecisionPointInterface) {
	fmt.Println("Pre-filtering Optimization Benefits:")
	fmt.Println("- Reduces O(n) policy evaluation to O(k) where k << n")
	fmt.Println("- Fast action/resource pattern matching")
	fmt.Println("- NotResource exclusion pre-filtering")
	fmt.Println("- Disabled policy skipping")

	// Example showing how pre-filtering works
	request := &models.EvaluationRequest{
		RequestID:  "prefilter-001",
		Subject:    models.NewMockUserSubject("service-account-123", "service-account-123"),
		ResourceID: "/api/payments/process",
		Action:     "execute",
		Environment: &models.EnvironmentInfo{
			ClientIP:  "10.0.2.100",
			UserAgent: "ServiceClient/1.0",
		},
		Context: map[string]interface{}{
			"service_type": "payment_processor",
			"api_version":  "v2",
		},
	}

	start := time.Now()
	decision, err := pdp.Evaluate(request)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Pre-filtered Decision: %s\n", decision.Result)
	fmt.Printf("Reason: %s\n", decision.Reason)
	fmt.Printf("Evaluation Time: %dms (with pre-filtering)\n", decision.EvaluationTimeMs)
	fmt.Printf("Total Time: %v\n", elapsed)

	fmt.Println("\nPre-filtering steps performed:")
	fmt.Println("1. Skip disabled policies")
	fmt.Println("2. Quick action pattern matching")
	fmt.Println("3. Quick resource pattern matching")
	fmt.Println("4. NotResource exclusion check")
	fmt.Println("5. Only evaluate remaining candidate policies")
}

// Helper function to demonstrate policy creation with enhanced features
func createEnhancedPolicy() *models.Policy {
	return &models.Policy{
		ID:          "enhanced-policy-001",
		PolicyName:  "Enhanced Business Hours Access",
		Description: "Demonstrates enhanced PDP features",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "BusinessHoursWithLocationCheck",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Multiple: []string{"read", "list"},
				},
				Resource: models.JSONActionResource{
					Single: "/api/reports/*",
				},
				Condition: map[string]interface{}{
					// Enhanced time-based conditions
					"IsBusinessHours": map[string]interface{}{
						"environment:is_business_hours": true,
					},
					"DayOfWeek": map[string]interface{}{
						"environment:day_of_week": []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
					},
					// Enhanced network conditions
					"IsInternalIP": map[string]interface{}{
						"environment:is_internal_ip": true,
					},
					// Enhanced string conditions
					"StringContains": map[string]interface{}{
						"user:department": "Engineering",
					},
					// Enhanced numeric conditions
					"NumericGreaterThanEquals": map[string]interface{}{
						"user:level": 5,
					},
				},
			},
		},
	}
}
