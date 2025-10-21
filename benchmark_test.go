package main

import (
	"testing"

	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

func BenchmarkSingleEvaluation(b *testing.B) {
	// Skip benchmark if no database available
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	testStorage := storage.NewTestStorage(&testing.T{})
	defer storage.CleanupTestStorage(&testing.T{}, testStorage)
	storage.SeedTestData(&testing.T{}, testStorage)

	pdp := evaluator.NewPolicyDecisionPoint(testStorage)

	request := &models.EvaluationRequest{
		RequestID:  "bench-001",
		SubjectID:  "sub-001",
		ResourceID: "res-001",
		Action:     "read",
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
	b.Skip("Skipping benchmark - requires database setup")
}

func BenchmarkDenyEvaluation(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}

func BenchmarkComplexEvaluation(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}

func BenchmarkExplainDecision(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}

func BenchmarkStorageOperations(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}

func BenchmarkOperators(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}

func BenchmarkMemoryAllocation(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}
