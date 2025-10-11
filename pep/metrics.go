package pep

import (
	"sync/atomic"
	"time"
)

// PEPMetrics holds performance and operational metrics for the PEP
type PEPMetrics struct {
	// Request metrics
	TotalRequests          int64 `json:"total_requests"`
	PermitDecisions        int64 `json:"permit_decisions"`
	DenyDecisions          int64 `json:"deny_decisions"`
	NotApplicableDecisions int64 `json:"not_applicable_decisions"`

	// Error metrics
	ValidationErrors int64 `json:"validation_errors"`
	EvaluationErrors int64 `json:"evaluation_errors"`
	TimeoutErrors    int64 `json:"timeout_errors"`

	// Performance metrics
	CacheHits               int64 `json:"cache_hits"`
	CacheMisses             int64 `json:"cache_misses"`
	AverageEvaluationTimeMs int64 `json:"average_evaluation_time_ms"`

	// Rate limiting metrics
	RateLimitExceeded int64 `json:"rate_limit_exceeded"`

	// Circuit breaker metrics
	CircuitBreakerOpen int64 `json:"circuit_breaker_open"`

	// Timing metrics
	totalEvaluationTime int64 // Internal counter for average calculation
	startTime           time.Time
}

// NewPEPMetrics creates a new metrics instance
func NewPEPMetrics() *PEPMetrics {
	return &PEPMetrics{
		startTime: time.Now(),
	}
}

// IncrementTotalRequests increments the total request counter
func (m *PEPMetrics) IncrementTotalRequests() {
	atomic.AddInt64(&m.TotalRequests, 1)
}

// IncrementPermitDecisions increments permit decision counter
func (m *PEPMetrics) IncrementPermitDecisions() {
	atomic.AddInt64(&m.PermitDecisions, 1)
}

// IncrementDenyDecisions increments deny decision counter
func (m *PEPMetrics) IncrementDenyDecisions() {
	atomic.AddInt64(&m.DenyDecisions, 1)
}

// IncrementNotApplicableDecisions increments not applicable decision counter
func (m *PEPMetrics) IncrementNotApplicableDecisions() {
	atomic.AddInt64(&m.NotApplicableDecisions, 1)
}

// IncrementValidationErrors increments validation error counter
func (m *PEPMetrics) IncrementValidationErrors() {
	atomic.AddInt64(&m.ValidationErrors, 1)
}

// IncrementEvaluationErrors increments evaluation error counter
func (m *PEPMetrics) IncrementEvaluationErrors() {
	atomic.AddInt64(&m.EvaluationErrors, 1)
}

// IncrementTimeoutErrors increments timeout error counter
func (m *PEPMetrics) IncrementTimeoutErrors() {
	atomic.AddInt64(&m.TimeoutErrors, 1)
}

// IncrementCacheHits increments cache hit counter
func (m *PEPMetrics) IncrementCacheHits() {
	atomic.AddInt64(&m.CacheHits, 1)
}

// IncrementCacheMisses increments cache miss counter
func (m *PEPMetrics) IncrementCacheMisses() {
	atomic.AddInt64(&m.CacheMisses, 1)
}

// IncrementRateLimitExceeded increments rate limit exceeded counter
func (m *PEPMetrics) IncrementRateLimitExceeded() {
	atomic.AddInt64(&m.RateLimitExceeded, 1)
}

// IncrementCircuitBreakerOpen increments circuit breaker open counter
func (m *PEPMetrics) IncrementCircuitBreakerOpen() {
	atomic.AddInt64(&m.CircuitBreakerOpen, 1)
}

// RecordEvaluationTime records evaluation time for average calculation
func (m *PEPMetrics) RecordEvaluationTime(timeMs int64) {
	atomic.AddInt64(&m.totalEvaluationTime, timeMs)

	// Calculate running average
	totalRequests := atomic.LoadInt64(&m.TotalRequests)
	if totalRequests > 0 {
		avgTime := atomic.LoadInt64(&m.totalEvaluationTime) / totalRequests
		atomic.StoreInt64(&m.AverageEvaluationTimeMs, avgTime)
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *PEPMetrics) GetSnapshot() *PEPMetricsSnapshot {
	return &PEPMetricsSnapshot{
		TotalRequests:           atomic.LoadInt64(&m.TotalRequests),
		PermitDecisions:         atomic.LoadInt64(&m.PermitDecisions),
		DenyDecisions:           atomic.LoadInt64(&m.DenyDecisions),
		NotApplicableDecisions:  atomic.LoadInt64(&m.NotApplicableDecisions),
		ValidationErrors:        atomic.LoadInt64(&m.ValidationErrors),
		EvaluationErrors:        atomic.LoadInt64(&m.EvaluationErrors),
		TimeoutErrors:           atomic.LoadInt64(&m.TimeoutErrors),
		CacheHits:               atomic.LoadInt64(&m.CacheHits),
		CacheMisses:             atomic.LoadInt64(&m.CacheMisses),
		AverageEvaluationTimeMs: atomic.LoadInt64(&m.AverageEvaluationTimeMs),
		RateLimitExceeded:       atomic.LoadInt64(&m.RateLimitExceeded),
		CircuitBreakerOpen:      atomic.LoadInt64(&m.CircuitBreakerOpen),
		UptimeSeconds:           int64(time.Since(m.startTime).Seconds()),
	}
}

// Reset resets all metrics to zero
func (m *PEPMetrics) Reset() {
	atomic.StoreInt64(&m.TotalRequests, 0)
	atomic.StoreInt64(&m.PermitDecisions, 0)
	atomic.StoreInt64(&m.DenyDecisions, 0)
	atomic.StoreInt64(&m.NotApplicableDecisions, 0)
	atomic.StoreInt64(&m.ValidationErrors, 0)
	atomic.StoreInt64(&m.EvaluationErrors, 0)
	atomic.StoreInt64(&m.TimeoutErrors, 0)
	atomic.StoreInt64(&m.CacheHits, 0)
	atomic.StoreInt64(&m.CacheMisses, 0)
	atomic.StoreInt64(&m.AverageEvaluationTimeMs, 0)
	atomic.StoreInt64(&m.RateLimitExceeded, 0)
	atomic.StoreInt64(&m.CircuitBreakerOpen, 0)
	atomic.StoreInt64(&m.totalEvaluationTime, 0)
	m.startTime = time.Now()
}

// PEPMetricsSnapshot represents a point-in-time snapshot of metrics
type PEPMetricsSnapshot struct {
	TotalRequests           int64 `json:"total_requests"`
	PermitDecisions         int64 `json:"permit_decisions"`
	DenyDecisions           int64 `json:"deny_decisions"`
	NotApplicableDecisions  int64 `json:"not_applicable_decisions"`
	ValidationErrors        int64 `json:"validation_errors"`
	EvaluationErrors        int64 `json:"evaluation_errors"`
	TimeoutErrors           int64 `json:"timeout_errors"`
	CacheHits               int64 `json:"cache_hits"`
	CacheMisses             int64 `json:"cache_misses"`
	AverageEvaluationTimeMs int64 `json:"average_evaluation_time_ms"`
	RateLimitExceeded       int64 `json:"rate_limit_exceeded"`
	CircuitBreakerOpen      int64 `json:"circuit_breaker_open"`
	UptimeSeconds           int64 `json:"uptime_seconds"`

	// Calculated metrics
	PermitRate        float64 `json:"permit_rate"`
	DenyRate          float64 `json:"deny_rate"`
	ErrorRate         float64 `json:"error_rate"`
	CacheHitRatio     float64 `json:"cache_hit_ratio"`
	RequestsPerSecond float64 `json:"requests_per_second"`
}

// CalculateDerivedMetrics calculates derived metrics from raw counters
func (s *PEPMetricsSnapshot) CalculateDerivedMetrics() {
	if s.TotalRequests > 0 {
		s.PermitRate = float64(s.PermitDecisions) / float64(s.TotalRequests)
		s.DenyRate = float64(s.DenyDecisions) / float64(s.TotalRequests)

		totalErrors := s.ValidationErrors + s.EvaluationErrors + s.TimeoutErrors
		s.ErrorRate = float64(totalErrors) / float64(s.TotalRequests)
	}

	totalCacheRequests := s.CacheHits + s.CacheMisses
	if totalCacheRequests > 0 {
		s.CacheHitRatio = float64(s.CacheHits) / float64(totalCacheRequests)
	}

	if s.UptimeSeconds > 0 {
		s.RequestsPerSecond = float64(s.TotalRequests) / float64(s.UptimeSeconds)
	}
}
