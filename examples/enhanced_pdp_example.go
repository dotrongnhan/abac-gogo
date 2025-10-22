package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func main() {
	// Initialize storage (using PostgreSQL storage)
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

	// Create enhanced PDP with custom configuration
	config := &evaluator.PDPConfig{
		MaxEvaluationTime: 3 * time.Second,
		EnableAudit:       true,
	}

	enhancedPDP := evaluator.NewEnhancedPDP(storage, config)

	// Example 1: Basic policy evaluation
	fmt.Println("=== Example 1: Basic Policy Evaluation ===")
	basicExample(enhancedPDP)

	// Example 2: Time-based access control
	fmt.Println("\n=== Example 2: Time-based Access Control ===")
	timeBasedExample(enhancedPDP)

	// Example 3: Location-based access control
	fmt.Println("\n=== Example 3: Location-based Access Control ===")
	locationBasedExample(enhancedPDP)

	// Example 4: Complex boolean expressions
	fmt.Println("\n=== Example 4: Complex Boolean Expressions ===")
	complexExpressionExample(enhancedPDP)

	// Example 5: Policy validation
	fmt.Println("\n=== Example 5: Policy Validation ===")
	policyValidationExample(enhancedPDP)

	// Example 6: Health check
	fmt.Println("\n=== Example 6: Health Check ===")
	healthCheckExample(enhancedPDP)
}

func basicExample(pdp *evaluator.EnhancedPDP) {
	ctx := context.Background()

	// Create a decision request
	request := &models.DecisionRequest{
		Subject: &models.Subject{
			ID:          "user123",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      5,
				"role":       "Developer",
			},
		},
		Resource: &models.Resource{
			ID:           "resource456",
			ResourceType: "document",
			ResourceID:   "/documents/sensitive/project-alpha.pdf",
			Attributes: map[string]interface{}{
				"classification": "confidential",
				"project":        "alpha",
			},
		},
		Action: &models.Action{
			ID:             "action789",
			ActionName:     "read",
			ActionCategory: "data-access",
		},
		Environment: &models.Environment{
			Timestamp: time.Now(),
			ClientIP:  "192.168.1.100",
			UserAgent: "Mozilla/5.0...",
			Location: &models.LocationInfo{
				Country: "Vietnam",
				Region:  "Ho Chi Minh City",
			},
		},
		Context: map[string]interface{}{
			"session_id":   "sess_abc123",
			"mfa_verified": true,
		},
		RequestID: "req_001",
	}

	// Evaluate the request
	response, err := pdp.Evaluate(ctx, request)
	if err != nil {
		fmt.Printf("Error evaluating request: %v\n", err)
		return
	}

	fmt.Printf("Decision: %s\n", response.Decision)
	fmt.Printf("Reason: %s\n", response.Reason)
	fmt.Printf("Policies: %v\n", response.Policies)
	fmt.Printf("Duration: %v\n", response.Duration)
}

func timeBasedExample(pdp *evaluator.EnhancedPDP) {
	// Create a policy with time windows
	policy := &models.Policy{
		ID:          "time-policy-001",
		PolicyName:  "Business Hours Access",
		Description: "Allow access only during business hours",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "BusinessHoursOnly",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single:  "read",
					IsArray: false,
				},
				Resource: models.JSONActionResource{
					Single:  "*",
					IsArray: false,
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:department": "Engineering",
					},
					"DateGreaterThan": map[string]interface{}{
						"environment:time_of_day": "09:00",
					},
					"DateLessThan": map[string]interface{}{
						"environment:time_of_day": "17:00",
					},
				},
			},
		},
	}

	// Validate the policy
	err := pdp.ValidatePolicy(policy)
	if err != nil {
		fmt.Printf("Policy validation failed: %v\n", err)
		return
	}

	fmt.Printf("Time-based policy validated successfully\n")
}

func locationBasedExample(pdp *evaluator.EnhancedPDP) {
	// Example of location-based condition evaluation
	conditionEvaluator := evaluator.NewEnhancedConditionEvaluator()

	// Define location condition
	locationCondition := &models.LocationCondition{
		AllowedCountries: []string{"Vietnam", "Singapore"},
		IPRanges:         []string{"192.168.1.0/24", "10.0.0.0/8"},
		GeoFencing: &models.GeoFenceCondition{
			Latitude:  10.8231, // Ho Chi Minh City
			Longitude: 106.6297,
			Radius:    50, // 50km radius
		},
	}

	// Test environment
	environment := &models.Environment{
		ClientIP: "192.168.1.100",
		Location: &models.LocationInfo{
			Country:   "Vietnam",
			Latitude:  10.8000,
			Longitude: 106.6500,
		},
	}

	// Evaluate location condition
	result := conditionEvaluator.EvaluateLocation(locationCondition, environment)
	fmt.Printf("Location-based access allowed: %t\n", result)
}

func complexExpressionExample(pdp *evaluator.EnhancedPDP) {
	// Example of complex boolean expression evaluation
	expressionEvaluator := evaluator.NewExpressionEvaluator()

	// Define a complex expression:
	// (user.department == "Engineering" AND user.level >= 5) OR
	// (user.role == "Admin")
	expression := &models.BooleanExpression{
		Type:     "compound",
		Operator: "or",
		Left: &models.BooleanExpression{
			Type:     "compound",
			Operator: "and",
			Left: &models.BooleanExpression{
				Type: "simple",
				Condition: &models.SimpleCondition{
					AttributePath: "user.department",
					Operator:      "eq",
					Value:         "Engineering",
				},
			},
			Right: &models.BooleanExpression{
				Type: "simple",
				Condition: &models.SimpleCondition{
					AttributePath: "user.level",
					Operator:      "gte",
					Value:         5,
				},
			},
		},
		Right: &models.BooleanExpression{
			Type: "simple",
			Condition: &models.SimpleCondition{
				AttributePath: "user.role",
				Operator:      "eq",
				Value:         "Admin",
			},
		},
	}

	// Test attributes
	attributes := map[string]interface{}{
		"user": map[string]interface{}{
			"department": "Engineering",
			"level":      6,
			"role":       "Developer",
		},
	}

	// Evaluate expression
	result := expressionEvaluator.EvaluateExpression(expression, attributes)
	fmt.Printf("Complex expression result: %t\n", result)
}

func policyValidationExample(pdp *evaluator.EnhancedPDP) {
	// Example of policy validation with various scenarios

	// Valid policy
	validPolicy := &models.Policy{
		ID:          "valid-policy-001",
		PolicyName:  "Valid Test Policy",
		Description: "A valid policy for testing",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "ValidStatement",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Multiple: []string{"read", "write"},
					IsArray:  true,
				},
				Resource: models.JSONActionResource{
					Single:  "/documents/*",
					IsArray: false,
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:department": "Engineering",
					},
				},
			},
		},
	}

	err := pdp.ValidatePolicy(validPolicy)
	if err != nil {
		fmt.Printf("Valid policy validation failed: %v\n", err)
	} else {
		fmt.Printf("Valid policy passed validation\n")
	}

	// Invalid policy (missing required fields)
	invalidPolicy := &models.Policy{
		ID: "invalid-policy-001",
		// Missing PolicyName and Version
		Statement: []models.PolicyStatement{},
	}

	err = pdp.ValidatePolicy(invalidPolicy)
	if err != nil {
		fmt.Printf("Invalid policy validation failed as expected: %v\n", err)
	} else {
		fmt.Printf("Invalid policy unexpectedly passed validation\n")
	}
}

func healthCheckExample(pdp *evaluator.EnhancedPDP) {
	ctx := context.Background()

	// Perform health check
	err := pdp.HealthCheck(ctx)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}
	fmt.Printf("Health check passed\n")
}
