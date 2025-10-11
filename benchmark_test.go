package main

import (
	"testing"

	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func BenchmarkSingleEvaluation(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	request := &models.EvaluationRequest{
		RequestID:  "bench-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
			"source_ip":   "10.0.1.50",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pdp.Evaluate(request)
		if err != nil {
			b.Fatalf("Evaluation failed: %v", err)
		}
	}
}

func BenchmarkBatchEvaluation(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	// Create batch of requests
	requests := make([]*models.EvaluationRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = &models.EvaluationRequest{
			RequestID:  "batch-bench-" + string(rune(i)),
			SubjectID:  "sub-001",
			ResourceID: "res-001",
			Action:     "read",
			Context: map[string]interface{}{
				"timestamp":   "2024-01-15T14:00:00Z",
				"time_of_day": "14:00",
				"source_ip":   "10.0.1.50",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pdp.BatchEvaluate(requests)
		if err != nil {
			b.Fatalf("Batch evaluation failed: %v", err)
		}
	}
}

func BenchmarkDenyEvaluation(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	// Request that should be denied (probation user writing)
	request := &models.EvaluationRequest{
		RequestID:  "deny-bench-001",
		SubjectID:  "sub-004",
		ResourceID: "res-002",
		Action:     "write",
		Context: map[string]interface{}{
			"timestamp": "2024-01-15T14:00:00Z",
			"source_ip": "10.0.1.50",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decision, err := pdp.Evaluate(request)
		if err != nil {
			b.Fatalf("Evaluation failed: %v", err)
		}
		if decision.Result != "deny" {
			b.Fatalf("Expected deny, got %s", decision.Result)
		}
	}
}

func BenchmarkComplexEvaluation(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	// Request that matches multiple policies
	request := &models.EvaluationRequest{
		RequestID:  "complex-bench-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "write",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
			"source_ip":   "10.0.1.50",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pdp.Evaluate(request)
		if err != nil {
			b.Fatalf("Evaluation failed: %v", err)
		}
	}
}

func BenchmarkExplainDecision(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	request := &models.EvaluationRequest{
		RequestID:  "explain-bench-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
			"source_ip":   "10.0.1.50",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pdp.ExplainDecision(request)
		if err != nil {
			b.Fatalf("Explain decision failed: %v", err)
		}
	}
}

func BenchmarkStorageOperations(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	b.Run("GetSubject", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := mockStorage.GetSubject("sub-001")
			if err != nil {
				b.Fatalf("GetSubject failed: %v", err)
			}
		}
	})

	b.Run("GetResource", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := mockStorage.GetResource("res-001")
			if err != nil {
				b.Fatalf("GetResource failed: %v", err)
			}
		}
	})

	b.Run("GetAction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := mockStorage.GetAction("read")
			if err != nil {
				b.Fatalf("GetAction failed: %v", err)
			}
		}
	})

	b.Run("GetPolicies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := mockStorage.GetPolicies()
			if err != nil {
				b.Fatalf("GetPolicies failed: %v", err)
			}
		}
	})
}

func BenchmarkOperators(b *testing.B) {
	// This would benchmark individual operators
	// For now, we'll benchmark through the evaluation system
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	testCases := []struct {
		name    string
		request *models.EvaluationRequest
	}{
		{
			name: "EqualOperator",
			request: &models.EvaluationRequest{
				RequestID:  "op-eq-001",
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "read",
				Context: map[string]interface{}{
					"timestamp": "2024-01-15T14:00:00Z",
					"source_ip": "10.0.1.50",
				},
			},
		},
		{
			name: "ContainsOperator",
			request: &models.EvaluationRequest{
				RequestID:  "op-contains-001",
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "write",
				Context: map[string]interface{}{
					"timestamp":   "2024-01-15T14:00:00Z",
					"time_of_day": "14:00",
					"source_ip":   "10.0.1.50",
				},
			},
		},
		{
			name: "BetweenOperator",
			request: &models.EvaluationRequest{
				RequestID:  "op-between-001",
				SubjectID:  "sub-001",
				ResourceID: "res-001",
				Action:     "write",
				Context: map[string]interface{}{
					"timestamp":   "2024-01-15T14:00:00Z",
					"time_of_day": "14:00",
					"source_ip":   "10.0.1.50",
				},
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := pdp.Evaluate(tc.request)
				if err != nil {
					b.Fatalf("Evaluation failed: %v", err)
				}
			}
		})
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocation(b *testing.B) {
	mockStorage, err := storage.NewMockStorage(".")
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	pdp := evaluator.NewPolicyDecisionPoint(mockStorage)

	request := &models.EvaluationRequest{
		RequestID:  "mem-bench-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
		Context: map[string]interface{}{
			"timestamp":   "2024-01-15T14:00:00Z",
			"time_of_day": "14:00",
			"source_ip":   "10.0.1.50",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pdp.Evaluate(request)
		if err != nil {
			b.Fatalf("Evaluation failed: %v", err)
		}
	}
}
