package evaluator

// This package provides a unified interface to all evaluator components.
// Import the specific subpackages directly for better type safety and clarity:
//
// - abac_go_example/evaluator/core        - Core PDP and policy validation
// - abac_go_example/evaluator/conditions  - Condition evaluation
// - abac_go_example/evaluator/matchers    - Action and resource matching
// - abac_go_example/evaluator/path        - Path resolution utilities
//
// Example usage:
//   import "abac_go_example/evaluator/core"
//   pdp := core.NewPolicyDecisionPoint(storage)
