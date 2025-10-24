package evaluator

import (
	"fmt"
	"testing"
	"time"

	"abac_go_example/models"
	"abac_go_example/storage"
)

// TestImprovedPDP_RealWorldScenarios tests realistic scenarios with all improvements
func TestImprovedPDP_RealWorldScenarios(t *testing.T) {
	// Setup mock storage with realistic policies
	mockStorage := storage.NewMockStorage()

	// Create realistic policies using enhanced features
	policies := []*models.Policy{
		{
			ID:          "business-hours-policy",
			PolicyName:  "Business Hours Document Access",
			Description: "Allow document access during business hours for employees",
			Version:     "1.0",
			Enabled:     true,
			Statement: []models.PolicyStatement{
				{
					Sid:    "BusinessHoursDocumentAccess",
					Effect: "Allow",
					Action: models.JSONActionResource{
						Multiple: []string{"document:read", "document:list"},
						IsArray:  true,
					},
					Resource: models.JSONActionResource{
						Single:  "api:documents:*",
						IsArray: false,
					},
					Condition: map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.subject_type": "employee",
						},
						"StringContains": map[string]interface{}{
							"user.department": "Engineering",
						},
						"IsBusinessHours": map[string]interface{}{
							"environment.is_business_hours": true,
						},
						"IPInRange": map[string]interface{}{
							"environment.client_ip": []interface{}{"192.168.1.0/24", "10.0.0.0/8"},
						},
					},
				},
			},
		},
		{
			ID:          "confidential-access-policy",
			PolicyName:  "Confidential Document Access",
			Description: "Allow confidential document access for senior staff",
			Version:     "1.0",
			Enabled:     true,
			Statement: []models.PolicyStatement{
				{
					Sid:    "ConfidentialDocumentAccess",
					Effect: "Allow",
					Action: models.JSONActionResource{
						Single:  "document:read",
						IsArray: false,
					},
					Resource: models.JSONActionResource{
						Single:  "api:documents:confidential/*",
						IsArray: false,
					},
					Condition: map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.clearance": "confidential",
						},
						"NumericGreaterThanEquals": map[string]interface{}{
							"user.level": 7,
						},
						"Bool": map[string]interface{}{
							"user.mfa_verified": true,
						},
					},
				},
			},
		},
		{
			ID:          "weekend-restriction-policy",
			PolicyName:  "Weekend Access Restriction",
			Description: "Deny sensitive operations during weekends",
			Version:     "1.0",
			Enabled:     true,
			Statement: []models.PolicyStatement{
				{
					Sid:    "WeekendRestriction",
					Effect: "Deny",
					Action: models.JSONActionResource{
						Multiple: []string{"document:write", "document:delete", "admin:*"},
						IsArray:  true,
					},
					Resource: models.JSONActionResource{
						Single:  "*",
						IsArray: false,
					},
					Condition: map[string]interface{}{
						"DayOfWeek": map[string]interface{}{
							"environment.day_of_week": []interface{}{"Saturday", "Sunday"},
						},
					},
				},
			},
		},
		{
			ID:          "mobile-device-policy",
			PolicyName:  "Mobile Device Restrictions",
			Description: "Restrict certain actions from mobile devices",
			Version:     "1.0",
			Enabled:     true,
			Statement: []models.PolicyStatement{
				{
					Sid:    "MobileDeviceRestriction",
					Effect: "Deny",
					Action: models.JSONActionResource{
						Multiple: []string{"admin:*", "finance:*"},
						IsArray:  true,
					},
					Resource: models.JSONActionResource{
						Single:  "*",
						IsArray: false,
					},
					Condition: map[string]interface{}{
						"Bool": map[string]interface{}{
							"environment.is_mobile": true,
						},
					},
				},
			},
		},
	}

	mockStorage.SetPolicies(policies)

	// Create test subjects that are referenced in the test scenarios
	subjects := []*models.Subject{
		{
			ID:          "emp-001",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department": "Engineering Team",
				"level":      5,
			},
		},
		{
			ID:          "senior-001",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department":   "Engineering",
				"level":        8,
				"clearance":    "confidential",
				"mfa_verified": true,
			},
		},
		{
			ID:          "emp-002",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department": "Finance",
				"level":      3,
			},
		},
		{
			ID:          "admin-001",
			SubjectType: "admin",
			Attributes: map[string]interface{}{
				"department": "IT",
				"level":      9,
				"role":       "system_admin",
			},
		},
		{
			ID:          "emp-003",
			SubjectType: "employee",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      4,
			},
		},
	}

	for _, subject := range subjects {
		err := mockStorage.CreateSubject(subject)
		if err != nil {
			t.Fatalf("Failed to create subject %s: %v", subject.ID, err)
		}
	}

	// Create test resources
	resources := []*models.Resource{
		{
			ID:         "api:documents:project-specs.pdf",
			ResourceID: "api:documents:project-specs.pdf",
			Attributes: map[string]interface{}{
				"type":           "document",
				"classification": "internal",
			},
		},
		{
			ID:         "api:documents:confidential/financial-report.pdf",
			ResourceID: "api:documents:confidential/financial-report.pdf",
			Attributes: map[string]interface{}{
				"type":           "document",
				"classification": "confidential",
			},
		},
		{
			ID:         "api:finance:budget-data",
			ResourceID: "api:finance:budget-data",
			Attributes: map[string]interface{}{
				"type":           "data",
				"classification": "sensitive",
			},
		},
		{
			ID:         "api:admin:user-management",
			ResourceID: "api:admin:user-management",
			Attributes: map[string]interface{}{
				"type":           "admin_panel",
				"classification": "restricted",
			},
		},
		{
			ID:         "api:documents:regular-file.pdf",
			ResourceID: "api:documents:regular-file.pdf",
			Attributes: map[string]interface{}{
				"type":           "document",
				"classification": "internal",
			},
		},
		{
			ID:         "api:documents:important-file.pdf",
			ResourceID: "api:documents:important-file.pdf",
			Attributes: map[string]interface{}{
				"type":           "document",
				"classification": "important",
			},
		},
	}

	for _, resource := range resources {
		err := mockStorage.CreateResource(resource)
		if err != nil {
			t.Fatalf("Failed to create resource %s: %v", resource.ID, err)
		}
	}

	// Create test actions
	actions := []*models.Action{
		{
			ID:             "document:read",
			ActionName:     "document:read",
			ActionCategory: "read",
		},
		{
			ID:             "document:list",
			ActionName:     "document:list",
			ActionCategory: "read",
		},
		{
			ID:             "document:write",
			ActionName:     "document:write",
			ActionCategory: "write",
		},
		{
			ID:             "document:delete",
			ActionName:     "document:delete",
			ActionCategory: "delete",
		},
		{
			ID:             "finance:read",
			ActionName:     "finance:read",
			ActionCategory: "read",
		},
		{
			ID:             "admin:user-management",
			ActionName:     "admin:user-management",
			ActionCategory: "admin",
		},
		{
			ID:             "admin:user:delete",
			ActionName:     "admin:user:delete",
			ActionCategory: "admin",
		},
	}

	for _, action := range actions {
		err := mockStorage.CreateAction(action)
		if err != nil {
			t.Fatalf("Failed to create action %s: %v", action.ID, err)
		}
	}

	pdp := NewPolicyDecisionPoint(mockStorage)

	// Test scenarios
	scenarios := []struct {
		name           string
		request        *models.EvaluationRequest
		expectedResult string
		description    string
	}{
		{
			name: "Business hours document access - should allow",
			request: &models.EvaluationRequest{
				RequestID:  "scenario-001",
				SubjectID:  "emp-001",
				ResourceID: "api:documents:project-specs.pdf",
				Action:     "document:read",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 14, 30, 0, 0, time.UTC)), // Thursday 14:30
				Environment: &models.EnvironmentInfo{
					ClientIP:  "192.168.1.100",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
					TimeOfDay: "14:30",
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"subject_type":                "employee",
					"department":                  "Engineering Team",
					"level":                       5,
					"is_business_hours":           true,
					"client_ip":                   "192.168.1.100",
				},
			},
			expectedResult: "permit",
			description:    "Employee accessing documents during business hours from office IP",
		},
		{
			name: "Confidential document access - should allow for senior staff",
			request: &models.EvaluationRequest{
				RequestID:  "scenario-002",
				SubjectID:  "senior-001",
				ResourceID: "api:documents:confidential/financial-report.pdf",
				Action:     "document:read",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 10, 0, 0, 0, time.UTC)),
				Environment: &models.EnvironmentInfo{
					ClientIP:  "10.0.1.50",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
					TimeOfDay: "10:00",
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"attributes": map[string]interface{}{
						"subject_type": "employee",
						"department":   "Engineering",
						"level":        8,
						"clearance":    "confidential",
						"mfa_verified": true,
					},
				},
			},
			expectedResult: "permit",
			description:    "Senior staff with confidential clearance accessing confidential documents",
		},
		{
			name: "Weekend restriction - should deny sensitive operations",
			request: &models.EvaluationRequest{
				RequestID:  "scenario-003",
				SubjectID:  "emp-002",
				ResourceID: "api:documents:important-file.pdf",
				Action:     "document:delete",
				Timestamp:  timePtr(time.Date(2024, 10, 26, 15, 0, 0, 0, time.UTC)), // Saturday 15:00
				Environment: &models.EnvironmentInfo{
					ClientIP:  "192.168.1.101",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
					TimeOfDay: "15:00",
					DayOfWeek: "Saturday",
				},
				Context: map[string]interface{}{
					"attributes": map[string]interface{}{
						"subject_type": "employee",
						"department":   "Engineering",
						"level":        6,
					},
				},
			},
			expectedResult: "deny",
			description:    "Delete operation blocked during weekend",
		},
		{
			name: "Mobile device restriction - should deny admin actions",
			request: &models.EvaluationRequest{
				RequestID:  "scenario-004",
				SubjectID:  "admin-001",
				ResourceID: "api:admin:user-management",
				Action:     "admin:user:delete",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 11, 0, 0, 0, time.UTC)),
				Environment: &models.EnvironmentInfo{
					ClientIP:  "192.168.1.102",
					UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
					TimeOfDay: "11:00",
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"attributes": map[string]interface{}{
						"subject_type": "employee",
						"department":   "IT",
						"level":        9,
						"role":         "admin",
					},
				},
			},
			expectedResult: "deny",
			description:    "Admin action blocked from mobile device",
		},
		{
			name: "After hours access - should deny",
			request: &models.EvaluationRequest{
				RequestID:  "scenario-005",
				SubjectID:  "emp-003",
				ResourceID: "api:documents:regular-file.pdf",
				Action:     "document:read",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 22, 0, 0, 0, time.UTC)), // Thursday 22:00
				Environment: &models.EnvironmentInfo{
					ClientIP:  "192.168.1.103",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
					TimeOfDay: "22:00",
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"attributes": map[string]interface{}{
						"subject_type": "employee",
						"department":   "Engineering",
						"level":        4,
					},
					"environment": map[string]interface{}{
						"is_business_hours": false,
						"client_ip":         "192.168.1.103",
					},
				},
			},
			expectedResult: "deny",
			description:    "Document access blocked after business hours",
		},
	}

	// Run scenarios
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			decision, err := pdp.Evaluate(scenario.request)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			if decision.Result != scenario.expectedResult {
				t.Errorf("Expected result %s, got %s. Reason: %s",
					scenario.expectedResult, decision.Result, decision.Reason)
			}

			// Verify evaluation time is reasonable
			if decision.EvaluationTimeMs <= 0 {
				t.Error("Evaluation time should be positive")
			}

			// Log results for analysis
			t.Logf("Scenario: %s", scenario.description)
			t.Logf("Result: %s (in %dms)", decision.Result, decision.EvaluationTimeMs)
			t.Logf("Reason: %s", decision.Reason)
			if len(decision.MatchedPolicies) > 0 {
				t.Logf("Matched policies: %v", decision.MatchedPolicies)
			}
		})
	}
}

// TestImprovedPDP_PerformanceComparison tests performance improvements
func TestImprovedPDP_PerformanceComparison(t *testing.T) {
	mockStorage := storage.NewMockStorage()

	// Create many policies to test performance
	var policies []*models.Policy
	for i := 0; i < 100; i++ {
		policy := &models.Policy{
			ID:      fmt.Sprintf("perf-pol-%03d", i),
			Enabled: i%10 != 0, // 90% enabled, 10% disabled
			Statement: []models.PolicyStatement{
				{
					Sid:    fmt.Sprintf("PerfStatement-%d", i),
					Effect: "Allow",
					Action: models.JSONActionResource{
						Single:  fmt.Sprintf("service-%d:action:*", i%5),
						IsArray: false,
					},
					Resource: models.JSONActionResource{
						Single:  fmt.Sprintf("api:resource-%d:*", i%10),
						IsArray: false,
					},
					Condition: map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"user.department": fmt.Sprintf("Department-%d", i%3),
						},
						"NumericGreaterThan": map[string]interface{}{
							"user.level": i % 10,
						},
					},
				},
			},
		}
		policies = append(policies, policy)
	}

	mockStorage.SetPolicies(policies)

	// Create test subjects for performance test
	perfSubjects := []*models.Subject{
		{
			ID:          "user-perf-001",
			SubjectType: "user",
			Attributes: map[string]interface{}{
				"department": "Department-1",
				"level":      5,
			},
		},
		{
			ID:          "user-perf-002",
			SubjectType: "user",
			Attributes: map[string]interface{}{
				"department": "Department-2",
				"level":      3,
			},
		},
		{
			ID:          "user-perf-003",
			SubjectType: "user",
			Attributes: map[string]interface{}{
				"department": "Department-3",
				"level":      7,
			},
		},
	}

	for _, subject := range perfSubjects {
		err := mockStorage.CreateSubject(subject)
		if err != nil {
			t.Fatalf("Failed to create performance test subject %s: %v", subject.ID, err)
		}
	}

	// Create test resources for performance test
	perfResources := []*models.Resource{
		{
			ID:         "api:resource-1:item-123",
			ResourceID: "api:resource-1:item-123",
			Attributes: map[string]interface{}{
				"type": "item",
			},
		},
		{
			ID:         "api:resource-2:item-456",
			ResourceID: "api:resource-2:item-456",
			Attributes: map[string]interface{}{
				"type": "item",
			},
		},
		{
			ID:         "api:resource-3:item-789",
			ResourceID: "api:resource-3:item-789",
			Attributes: map[string]interface{}{
				"type": "item",
			},
		},
		{
			ID:         "api:resource-5:item-456",
			ResourceID: "api:resource-5:item-456",
			Attributes: map[string]interface{}{
				"type": "item",
			},
		},
		{
			ID:         "api:resource-7:item-789",
			ResourceID: "api:resource-7:item-789",
			Attributes: map[string]interface{}{
				"type": "item",
			},
		},
	}

	for _, resource := range perfResources {
		err := mockStorage.CreateResource(resource)
		if err != nil {
			t.Fatalf("Failed to create performance test resource %s: %v", resource.ID, err)
		}
	}

	// Create test actions for performance test
	perfActions := []*models.Action{
		{
			ID:             "service-1:action:read",
			ActionName:     "service-1:action:read",
			ActionCategory: "read",
		},
		{
			ID:             "service-2:action:write",
			ActionName:     "service-2:action:write",
			ActionCategory: "write",
		},
		{
			ID:             "service-3:action:delete",
			ActionName:     "service-3:action:delete",
			ActionCategory: "delete",
		},
	}

	for _, action := range perfActions {
		err := mockStorage.CreateAction(action)
		if err != nil {
			t.Fatalf("Failed to create performance test action %s: %v", action.ID, err)
		}
	}

	pdp := NewPolicyDecisionPoint(mockStorage)

	// Test requests
	requests := []*models.EvaluationRequest{
		{
			RequestID:  "perf-001",
			SubjectID:  "user-perf-001",
			ResourceID: "api:resource-1:item-123",
			Action:     "service-1:action:read",
			Timestamp:  timePtr(time.Now()),
			Environment: &models.EnvironmentInfo{
				ClientIP:  "192.168.1.100",
				TimeOfDay: "14:30",
				DayOfWeek: "Thursday",
			},
			Context: map[string]interface{}{
				"department": "Department-1",
				"level":      5,
			},
		},
		{
			RequestID:  "perf-002",
			SubjectID:  "user-perf-002",
			ResourceID: "api:resource-5:item-456",
			Action:     "service-2:action:write",
			Timestamp:  timePtr(time.Now()),
			Environment: &models.EnvironmentInfo{
				ClientIP:  "10.0.1.50",
				TimeOfDay: "15:45",
				DayOfWeek: "Thursday",
			},
			Context: map[string]interface{}{
				"department": "Department-2",
				"level":      7,
			},
		},
	}

	// Measure performance
	var totalEvaluationTime int
	for i, request := range requests {
		start := time.Now()
		decision, err := pdp.Evaluate(request)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}

		totalEvaluationTime += decision.EvaluationTimeMs

		t.Logf("Request %d: %s (internal: %dms, total: %v)",
			i+1, decision.Result, decision.EvaluationTimeMs, elapsed)
	}

	averageTime := totalEvaluationTime / len(requests)
	t.Logf("Average evaluation time: %dms", averageTime)

	// Performance should be reasonable even with many policies
	if averageTime > 100 { // 100ms threshold
		t.Errorf("Average evaluation time too high: %dms", averageTime)
	}
}

// TestImprovedPDP_ComplexConditionScenarios tests complex condition combinations
func TestImprovedPDP_ComplexConditionScenarios(t *testing.T) {
	mockStorage := storage.NewMockStorage()

	// Policy with complex nested conditions
	policy := &models.Policy{
		ID:      "complex-condition-policy",
		Enabled: true,
		Statement: []models.PolicyStatement{
			{
				Sid:      "ComplexConditionAccess",
				Effect:   "Allow",
				Action:   models.JSONActionResource{Single: "document:read", IsArray: false},
				Resource: models.JSONActionResource{Single: "api:documents:*", IsArray: false},
				Condition: map[string]interface{}{
					// Simplified conditions that should work
					"StringEquals": map[string]interface{}{
						"user.department": "Engineering",
					},
					"NumericGreaterThanEquals": map[string]interface{}{
						"user.level": 5,
					},
				},
			},
		},
	}

	mockStorage.SetPolicies([]*models.Policy{policy})

	// Create test subjects for complex condition scenarios
	complexSubjects := []*models.Subject{
		{
			ID:          "dev-001",
			SubjectType: "developer",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      7,
				"roles":      []string{"developer", "reviewer"},
				"email":      "dev001@company.com",
			},
		},
		{
			ID:          "dev-002",
			SubjectType: "developer",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      4,
				"roles":      []string{"developer"},
				"email":      "dev002@company.com",
			},
		},
		{
			ID:          "dev-003",
			SubjectType: "developer",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"level":      8,
				"roles":      []string{"developer", "lead"},
				"email":      "dev003@company.com",
			},
		},
	}

	for _, subject := range complexSubjects {
		err := mockStorage.CreateSubject(subject)
		if err != nil {
			t.Fatalf("Failed to create complex test subject %s: %v", subject.ID, err)
		}
	}

	// Create test resources for complex condition scenarios
	complexResources := []*models.Resource{
		{
			ID:         "api:documents:project-file.pdf",
			ResourceID: "api:documents:project-file.pdf",
			Attributes: map[string]interface{}{
				"type":           "document",
				"classification": "internal",
				"project":        "alpha",
			},
		},
	}

	for _, resource := range complexResources {
		err := mockStorage.CreateResource(resource)
		if err != nil {
			t.Fatalf("Failed to create complex test resource %s: %v", resource.ID, err)
		}
	}

	// Create test actions for complex condition scenarios
	complexActions := []*models.Action{
		{
			ID:             "document:read",
			ActionName:     "document:read",
			ActionCategory: "read",
		},
	}

	for _, action := range complexActions {
		err := mockStorage.CreateAction(action)
		if err != nil {
			t.Fatalf("Failed to create complex test action %s: %v", action.ID, err)
		}
	}

	pdp := NewPolicyDecisionPoint(mockStorage)

	tests := []struct {
		name           string
		request        *models.EvaluationRequest
		expectedResult string
	}{
		{
			name: "All conditions match - should allow",
			request: &models.EvaluationRequest{
				RequestID:  "complex-001",
				SubjectID:  "dev-001",
				ResourceID: "api:documents:project-file.pdf",
				Action:     "document:read",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 14, 30, 0, 0, time.UTC)),
				Environment: &models.EnvironmentInfo{
					ClientIP:  "192.168.1.100",
					TimeOfDay: "14:30",
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"department": "Engineering Team",
					"level":      7,
					"roles":      []interface{}{"developer", "code_reviewer"},
					"email":      "john.doe@company.com",
				},
			},
			expectedResult: "permit",
		},
		{
			name: "Time condition fails - should deny",
			request: &models.EvaluationRequest{
				RequestID:  "complex-002",
				SubjectID:  "dev-002",
				ResourceID: "api:documents:project-file.pdf",
				Action:     "document:read",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 20, 0, 0, 0, time.UTC)), // After hours
				Environment: &models.EnvironmentInfo{
					ClientIP:  "192.168.1.101",
					TimeOfDay: "20:00", // After 17:00
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"department": "Engineering Team",
					"level":      7,
					"roles":      []interface{}{"developer"},
					"email":      "jane.doe@company.com",
				},
			},
			expectedResult: "deny",
		},
		{
			name: "All basic conditions match - should allow",
			request: &models.EvaluationRequest{
				RequestID:  "complex-003",
				SubjectID:  "dev-003",
				ResourceID: "api:documents:project-file.pdf",
				Action:     "document:read",
				Timestamp:  timePtr(time.Date(2024, 10, 24, 14, 30, 0, 0, time.UTC)),
				Environment: &models.EnvironmentInfo{
					ClientIP:  "203.0.113.100", // External IP
					TimeOfDay: "14:30",
					DayOfWeek: "Thursday",
				},
				Context: map[string]interface{}{
					"department": "Engineering",
					"level":      8,
					"roles":      []interface{}{"developer", "lead"},
					"email":      "dev003@company.com",
				},
			},
			expectedResult: "permit",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision, err := pdp.Evaluate(test.request)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			if decision.Result != test.expectedResult {
				t.Errorf("Expected %s, got %s. Reason: %s",
					test.expectedResult, decision.Result, decision.Reason)
			}

			t.Logf("Complex condition result: %s (in %dms)",
				decision.Result, decision.EvaluationTimeMs)
		})
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
