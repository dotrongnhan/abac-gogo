package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"abac_go_example/evaluator/core"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func runHierarchicalExtendedDemo() {
	fmt.Println("🏢 Hierarchical + Extended Format Demo")
	fmt.Println("=====================================")

	// Initialize storage and PDP
	mockStorage := storage.NewMockStorage()
	pdp := core.NewPolicyDecisionPoint(mockStorage)

	// Demo different enterprise use cases
	fmt.Println("\n=== Use Case 1: Multi-tenant SaaS Platform ===")
	demoMultiTenantSaaS(pdp, mockStorage)

	fmt.Println("\n=== Use Case 2: Healthcare System ===")
	demoHealthcareSystem(pdp, mockStorage)

	fmt.Println("\n=== Use Case 3: Financial Services ===")
	demoFinancialServices(pdp, mockStorage)

	fmt.Println("\n=== Use Case 4: E-commerce Marketplace ===")
	demoEcommerceMarketplace(pdp, mockStorage)
}

func demoMultiTenantSaaS(pdp core.PolicyDecisionPointInterface, mockStorage storage.Storage) {
	fmt.Println("Scenario: Document Management System")
	fmt.Println("Structure: Organization → Department → Project → Document")

	// Create policy for multi-tenant SaaS
	policy := &models.Policy{
		ID:          "saas-hierarchical-001",
		PolicyName:  "Multi-tenant Document Access",
		Description: "Hierarchical access control for SaaS platform",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "OrgAdminFullAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "document-service:*:*",
				},
				Resource: models.JSONActionResource{
					Single: "api:documents:org:${user:Organization}/*/*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "org-admin",
					},
				},
			},
			{
				Sid:    "DeptManagerAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "document-service:file:*",
				},
				Resource: models.JSONActionResource{
					Single: "api:documents:org:${user:Organization}/dept:${user:Department}/*/*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "dept-manager",
					},
				},
			},
			{
				Sid:    "ProjectMemberRead",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "document-service:file:read",
				},
				Resource: models.JSONActionResource{
					Single: "api:documents:org:${user:Organization}/dept:*/project:${user:CurrentProject}/*",
				},
			},
		},
	}

	// Store policy
	err := mockStorage.CreatePolicy(policy)
	if err != nil {
		log.Printf("Error storing policy: %v", err)
		return
	}

	// Test cases
	testCases := []struct {
		name     string
		user     map[string]interface{}
		resource string
		action   string
		expected string
	}{
		{
			name: "Org Admin - Full Access",
			user: map[string]interface{}{
				"user:Organization": "acme-corp",
				"user:Role":         "org-admin",
			},
			resource: "api:documents:org:acme-corp/dept:engineering/project:alpha/file:design.pdf",
			action:   "document-service:file:write",
			expected: "PERMIT",
		},
		{
			name: "Dept Manager - Department Access",
			user: map[string]interface{}{
				"user:Organization": "acme-corp",
				"user:Department":   "engineering",
				"user:Role":         "dept-manager",
			},
			resource: "api:documents:org:acme-corp/dept:engineering/project:beta/file:spec.pdf",
			action:   "document-service:file:read",
			expected: "PERMIT",
		},
		{
			name: "Project Member - Project Access",
			user: map[string]interface{}{
				"user:Organization":   "acme-corp",
				"user:CurrentProject": "alpha",
				"user:Role":           "developer",
			},
			resource: "api:documents:org:acme-corp/dept:engineering/project:alpha/file:code.pdf",
			action:   "document-service:file:read",
			expected: "PERMIT",
		},
		{
			name: "Cross-org Access Denied",
			user: map[string]interface{}{
				"user:Organization": "other-corp",
				"user:Role":         "org-admin",
			},
			resource: "api:documents:org:acme-corp/dept:engineering/project:alpha/file:secret.pdf",
			action:   "document-service:file:read",
			expected: "DENY",
		},
	}

	runTestCases("Multi-tenant SaaS", testCases, pdp)
}

func demoHealthcareSystem(pdp core.PolicyDecisionPointInterface, mockStorage storage.Storage) {
	fmt.Println("Scenario: Hospital Management System")
	fmt.Println("Structure: Hospital → Department → Ward → Patient → Record")

	policy := &models.Policy{
		ID:          "healthcare-hierarchical-001",
		PolicyName:  "Healthcare Records Access",
		Description: "HIPAA-compliant hierarchical access control",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "DoctorPatientAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "medical:record:*",
				},
				Resource: models.JSONActionResource{
					Single: "api:medical:hospital:${user:Hospital}/dept:${user:Department}/ward:*/patient:${user:AssignedPatients}/record:*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "doctor",
					},
				},
			},
			{
				Sid:    "NurseWardAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "medical:record:read",
				},
				Resource: models.JSONActionResource{
					Single: "api:medical:hospital:${user:Hospital}/dept:*/ward:${user:AssignedWard}/patient:*/record:vital-signs",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "nurse",
					},
				},
			},
			{
				Sid:    "EmergencyAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "medical:record:read",
				},
				Resource: models.JSONActionResource{
					Single: "api:medical:hospital:${user:Hospital}/*/*/patient:*/record:emergency-info",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "emergency-doctor",
					},
				},
			},
		},
	}

	err := mockStorage.CreatePolicy(policy)
	if err != nil {
		log.Printf("Error storing policy: %v", err)
		return
	}

	testCases := []struct {
		name     string
		user     map[string]interface{}
		resource string
		action   string
		expected string
	}{
		{
			name: "Doctor - Assigned Patient",
			user: map[string]interface{}{
				"user:Hospital":         "general-hospital",
				"user:Department":       "cardiology",
				"user:AssignedPatients": "p-12345",
				"user:Role":             "doctor",
			},
			resource: "api:medical:hospital:general-hospital/dept:cardiology/ward:icu/patient:p-12345/record:lab-results",
			action:   "medical:record:read",
			expected: "PERMIT",
		},
		{
			name: "Nurse - Ward Vital Signs",
			user: map[string]interface{}{
				"user:Hospital":     "general-hospital",
				"user:AssignedWard": "icu",
				"user:Role":         "nurse",
			},
			resource: "api:medical:hospital:general-hospital/dept:cardiology/ward:icu/patient:p-67890/record:vital-signs",
			action:   "medical:record:read",
			expected: "PERMIT",
		},
		{
			name: "Emergency Doctor - Emergency Info",
			user: map[string]interface{}{
				"user:Hospital": "general-hospital",
				"user:Role":     "emergency-doctor",
			},
			resource: "api:medical:hospital:general-hospital/dept:cardiology/ward:icu/patient:p-99999/record:emergency-info",
			action:   "medical:record:read",
			expected: "PERMIT",
		},
	}

	runTestCases("Healthcare System", testCases, pdp)
}

func demoFinancialServices(pdp core.PolicyDecisionPointInterface, mockStorage storage.Storage) {
	fmt.Println("Scenario: Banking System")
	fmt.Println("Structure: Bank → Branch → Customer → Account → Transaction")

	policy := &models.Policy{
		ID:          "banking-hierarchical-001",
		PolicyName:  "Banking Access Control",
		Description: "SOX-compliant banking access control",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "CustomerOwnAccounts",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "banking:account:*",
				},
				Resource: models.JSONActionResource{
					Single: "api:banking:bank:*/branch:*/customer:${user:CustomerId}/account:*/transaction:*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "customer",
					},
				},
			},
			{
				Sid:    "BranchManagerAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "banking:account:view",
				},
				Resource: models.JSONActionResource{
					Single: "api:banking:bank:${user:Bank}/branch:${user:Branch}/customer:*/account:*/transaction:*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "branch-manager",
					},
				},
			},
		},
	}

	err := mockStorage.CreatePolicy(policy)
	if err != nil {
		log.Printf("Error storing policy: %v", err)
		return
	}

	testCases := []struct {
		name     string
		user     map[string]interface{}
		resource string
		action   string
		expected string
	}{
		{
			name: "Customer - Own Account",
			user: map[string]interface{}{
				"user:CustomerId": "c-67890",
				"user:Role":       "customer",
			},
			resource: "api:banking:bank:chase/branch:manhattan/customer:c-67890/account:checking-001/transaction:txn-98765",
			action:   "banking:account:view",
			expected: "PERMIT",
		},
		{
			name: "Branch Manager - Branch Accounts",
			user: map[string]interface{}{
				"user:Bank":   "chase",
				"user:Branch": "manhattan",
				"user:Role":   "branch-manager",
			},
			resource: "api:banking:bank:chase/branch:manhattan/customer:c-11111/account:savings-002/transaction:txn-55555",
			action:   "banking:account:view",
			expected: "PERMIT",
		},
	}

	runTestCases("Financial Services", testCases, pdp)
}

func demoEcommerceMarketplace(pdp core.PolicyDecisionPointInterface, mockStorage storage.Storage) {
	fmt.Println("Scenario: Multi-vendor Marketplace")
	fmt.Println("Structure: Platform → Vendor → Category → Product → Variant")

	policy := &models.Policy{
		ID:          "marketplace-hierarchical-001",
		PolicyName:  "Marketplace Access Control",
		Description: "Multi-vendor marketplace access control",
		Version:     "1.0",
		Enabled:     true,
		Statement: []models.PolicyStatement{
			{
				Sid:    "VendorProductManagement",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "marketplace:product:*",
				},
				Resource: models.JSONActionResource{
					Single: "api:marketplace:platform:*/vendor:${user:VendorId}/category:*/product:*/variant:*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "vendor",
					},
				},
			},
			{
				Sid:    "CategoryManagerAccess",
				Effect: "Allow",
				Action: models.JSONActionResource{
					Single: "marketplace:product:moderate",
				},
				Resource: models.JSONActionResource{
					Single: "api:marketplace:platform:${user:Platform}/vendor:*/category:${user:ManagedCategories}/product:*/variant:*",
				},
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"user:Role": "category-manager",
					},
				},
			},
		},
	}

	err := mockStorage.CreatePolicy(policy)
	if err != nil {
		log.Printf("Error storing policy: %v", err)
		return
	}

	testCases := []struct {
		name     string
		user     map[string]interface{}
		resource string
		action   string
		expected string
	}{
		{
			name: "Vendor - Own Products",
			user: map[string]interface{}{
				"user:VendorId": "apple",
				"user:Role":     "vendor",
			},
			resource: "api:marketplace:platform:amazon/vendor:apple/category:electronics/product:iphone-15/variant:pro-max-256gb",
			action:   "marketplace:product:update",
			expected: "PERMIT",
		},
		{
			name: "Category Manager - Category Products",
			user: map[string]interface{}{
				"user:Platform":          "amazon",
				"user:ManagedCategories": "electronics",
				"user:Role":              "category-manager",
			},
			resource: "api:marketplace:platform:amazon/vendor:samsung/category:electronics/product:galaxy-s24/variant:ultra-512gb",
			action:   "marketplace:product:moderate",
			expected: "PERMIT",
		},
	}

	runTestCases("E-commerce Marketplace", testCases, pdp)
}

func runTestCases(scenario string, testCases []struct {
	name     string
	user     map[string]interface{}
	resource string
	action   string
	expected string
}, pdp core.PolicyDecisionPointInterface) {

	fmt.Printf("\nTesting %s:\n", scenario)
	fmt.Println(strings.Repeat("-", 50))

	for i, tc := range testCases {
		fmt.Printf("%d. %s\n", i+1, tc.name)
		fmt.Printf("   Resource: %s\n", tc.resource)
		fmt.Printf("   Action:   %s\n", tc.action)

		// Create evaluation request
		request := &models.EvaluationRequest{
			RequestID:  fmt.Sprintf("test-%d", i+1),
			Subject:    models.NewMockUserSubject("test-user", "test-user"),
			ResourceID: tc.resource,
			Action:     tc.action,
			Context:    tc.user,
		}

		// Add resource ID to context
		request.Context["request:ResourceId"] = tc.resource

		// Evaluate
		decision, err := pdp.Evaluate(request)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}

		// Check result
		if decision.Result == tc.expected {
			fmt.Printf("   ✅ %s (Expected: %s)\n", decision.Result, tc.expected)
		} else {
			fmt.Printf("   ❌ %s (Expected: %s)\n", decision.Result, tc.expected)
		}

		if decision.Reason != "" {
			fmt.Printf("   📝 Reason: %s\n", decision.Reason)
		}
		fmt.Println()
	}
}

// Helper function to pretty print JSON
func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(string(b))
}
