package pep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"abac_go_example/evaluator/core"
	"abac_go_example/storage"
)

func TestHTTPMiddleware_Handler(t *testing.T) {
	// Setup test environment
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)

	pdp := core.NewPolicyDecisionPoint(testStorage)
	auditLogger := NewNoOpAuditLogger()
	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

	config := &MiddlewareConfig{
		UnauthorizedStatusCode:  http.StatusUnauthorized,
		ForbiddenStatusCode:     http.StatusForbidden,
		ErrorStatusCode:         http.StatusInternalServerError,
		IncludeReasonInResponse: true,
		RequireAuthentication:   true,
		DefaultAction:           "read",
		SubjectHeader:           "X-Subject-ID",
		AuthorizationHeader:     "Authorization",
		SkipPaths:               []string{"/health"},
	}

	middleware := NewHTTPMiddleware(pep, config)

	// Set custom resource extractor that maps paths to resource IDs
	middleware.SetResourceExtractor(func(r *http.Request) (string, error) {
		path := r.URL.Path
		// Map common paths to resource IDs for testing
		switch path {
		case "/api/v1/users":
			return "res-001", nil
		case "/api/v1/test":
			return "res-001", nil
		default:
			return path, nil
		}
	})

	// Test handler that should be called if access is granted
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := middleware.Handler(testHandler)

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
		expectBody     string
	}{
		{
			name:   "Successful request with valid subject",
			method: "GET",
			path:   "/api/v1/users", // Valid path
			headers: map[string]string{
				"X-Subject-ID": "sub-001", // Engineering user
			},
			expectedStatus: http.StatusOK,
			expectBody:     "success",
		},
		{
			name:   "Denied request - probation user write",
			method: "POST",
			path:   "/api/v1/users",
			headers: map[string]string{
				"X-Subject-ID": "sub-004", // User on probation
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing authentication",
			method:         "GET",
			path:           "/api/v1/users",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Skip path - health check",
			method:         "GET",
			path:           "/health",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			expectBody:     "success",
		},
		{
			name:   "Bearer token authentication",
			method: "GET",
			path:   "/api/v1/test",
			headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
			expectedStatus: http.StatusOK,
			expectBody:     "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectBody != "" {
				body := strings.TrimSpace(rr.Body.String())
				if body != tt.expectBody {
					t.Errorf("Expected body '%s', got '%s'", tt.expectBody, body)
				}
			}
		})
	}
}

func TestHTTPMiddleware_SubjectExtractor(t *testing.T) {
	config := DefaultMiddlewareConfig()
	extractor := DefaultSubjectExtractor(config.SubjectHeader, config.AuthorizationHeader)

	tests := []struct {
		name        string
		headers     map[string]string
		expected    string
		expectError bool
	}{
		{
			name: "Subject header",
			headers: map[string]string{
				"X-Subject-ID": "user-123",
			},
			expected:    "user-123",
			expectError: false,
		},
		{
			name: "Bearer token",
			headers: map[string]string{
				"Authorization": "Bearer abc123def",
			},
			expected:    "user_abc123de", // Truncated token
			expectError: false,
		},
		{
			name: "Basic auth",
			headers: map[string]string{
				"Authorization": "Basic dXNlcjpwYXNz", // user:pass
			},
			expected:    "sub-001", // Maps to existing subject
			expectError: false,
		},
		{
			name:        "No authentication",
			headers:     map[string]string{},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result, err := extractor(req)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestHTTPMiddleware_ResourceExtractor(t *testing.T) {
	extractor := DefaultResourceExtractor()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "API path",
			path:     "/api/v1/users",
			expected: "/api/v1/users",
		},
		{
			name:     "Root path",
			path:     "/",
			expected: "/root",
		},
		{
			name:     "Empty path",
			path:     "/",
			expected: "/root",
		},
		{
			name:     "Complex path",
			path:     "/api/v1/users/123/profile",
			expected: "/api/v1/users/123/profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			result, err := extractor(req)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestHTTPMiddleware_ContextExtractor(t *testing.T) {
	extractor := DefaultContextExtractor()

	req := httptest.NewRequest("POST", "/api/v1/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")

	context, err := extractor(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check required fields
	if context["method"] != "POST" {
		t.Errorf("Expected method 'POST', got %v", context["method"])
	}

	if context["user_agent"] != "test-agent" {
		t.Errorf("Expected user_agent 'test-agent', got %v", context["user_agent"])
	}

	if context["content_type"] != "application/json" {
		t.Errorf("Expected content_type 'application/json', got %v", context["content_type"])
	}

	if context["source_ip"] != "192.168.1.1" {
		t.Errorf("Expected source_ip '192.168.1.1', got %v", context["source_ip"])
	}

	if context["query_params"] != "param=value" {
		t.Errorf("Expected query_params 'param=value', got %v", context["query_params"])
	}

	// Check timestamp exists
	if _, ok := context["timestamp"]; !ok {
		t.Error("Expected timestamp in context")
	}
}

func TestRESTfulMiddleware(t *testing.T) {
	testStorage := storage.NewTestStorage(t)
	defer storage.CleanupTestStorage(t, testStorage)
	storage.SeedTestData(t, testStorage)
	pdp := core.NewPolicyDecisionPoint(testStorage)
	auditLogger := NewNoOpAuditLogger()
	pep := NewSimplePolicyEnforcementPoint(pdp, auditLogger, nil)

	middleware := NewRESTfulMiddleware(pep)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := middleware.Handler(testHandler)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "API health check - should skip",
			path:           "/api/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API metrics - should skip",
			path:           "/api/metrics",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Regular API endpoint - should require auth",
			path:           "/api/v1/users",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func BenchmarkHTTPMiddleware_Handler(b *testing.B) {
	b.Skip("Skipping benchmark - requires database setup")
}

func TestGetEnforcementResult(t *testing.T) {
	// Test with context that has enforcement result
	req := httptest.NewRequest("GET", "/test", nil)

	result := &EnforcementResult{
		Allowed: true,
		Reason:  "test result",
	}

	ctx := context.WithValue(req.Context(), "pep_result", result)
	req = req.WithContext(ctx)

	retrievedResult, ok := GetEnforcementResult(req)
	if !ok {
		t.Error("Expected to find enforcement result in context")
	}

	if retrievedResult.Allowed != true {
		t.Errorf("Expected allowed=true, got %v", retrievedResult.Allowed)
	}

	if retrievedResult.Reason != "test result" {
		t.Errorf("Expected reason='test result', got %v", retrievedResult.Reason)
	}

	// Test with context that doesn't have enforcement result
	req2 := httptest.NewRequest("GET", "/test", nil)
	_, ok2 := GetEnforcementResult(req2)
	if ok2 {
		t.Error("Expected not to find enforcement result in empty context")
	}
}
