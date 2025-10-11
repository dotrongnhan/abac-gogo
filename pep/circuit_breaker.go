package pep

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	// StateClosed - normal operation
	StateClosed CircuitBreakerState = iota
	// StateOpen - circuit breaker is open, requests are rejected
	StateOpen
	// StateHalfOpen - testing if service has recovered
	StateHalfOpen
)

// CircuitBreaker implements circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	failureThreshold      int
	recoveryTimeout       time.Duration
	maxConcurrentRequests int

	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	currentRequests int

	mu sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, recoveryTimeout time.Duration, maxConcurrentRequests int) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold:      failureThreshold,
		recoveryTimeout:       recoveryTimeout,
		maxConcurrentRequests: maxConcurrentRequests,
		state:                 StateClosed,
	}
}

// Allow checks if a request should be allowed through the circuit breaker
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Normal operation - check concurrent requests limit
		if cb.currentRequests >= cb.maxConcurrentRequests {
			return false
		}
		cb.currentRequests++
		return true

	case StateOpen:
		// Check if recovery timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.recoveryTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.currentRequests = 1
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests to test recovery
		if cb.currentRequests >= cb.maxConcurrentRequests/2 {
			return false
		}
		cb.currentRequests++
		return true

	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.currentRequests--

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0

	case StateHalfOpen:
		cb.successCount++
		// If we have enough successful requests, close the circuit
		if cb.successCount >= cb.failureThreshold/2 {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.currentRequests--
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Open circuit if failure threshold is reached
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}

	case StateHalfOpen:
		// Go back to open state on any failure
		cb.state = StateOpen
		cb.successCount = 0
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"current_requests":  cb.currentRequests,
		"last_failure_time": cb.lastFailureTime,
	}
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateOpen
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
}

// String returns string representation of circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
