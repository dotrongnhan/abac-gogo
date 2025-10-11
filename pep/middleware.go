package pep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"abac_go_example/models"
)

// PEPInterface defines the interface for policy enforcement points
type PEPInterface interface {
	EnforceRequest(ctx context.Context, request *models.EvaluationRequest) (*EnforcementResult, error)
}

// HTTPMiddleware provides HTTP middleware for web applications
type HTTPMiddleware struct {
	pep               PEPInterface
	config            *MiddlewareConfig
	subjectExtractor  SubjectExtractor
	resourceExtractor ResourceExtractor
	contextExtractor  ContextExtractor
}

// MiddlewareConfig holds configuration for HTTP middleware
type MiddlewareConfig struct {
	// Response configuration
	UnauthorizedStatusCode  int  `json:"unauthorized_status_code"`
	ForbiddenStatusCode     int  `json:"forbidden_status_code"`
	ErrorStatusCode         int  `json:"error_status_code"`
	IncludeReasonInResponse bool `json:"include_reason_in_response"`

	// Request handling
	SkipPaths             []string `json:"skip_paths"`
	RequireAuthentication bool     `json:"require_authentication"`
	DefaultAction         string   `json:"default_action"`

	// Headers
	SubjectHeader       string `json:"subject_header"`
	AuthorizationHeader string `json:"authorization_header"`
	RequestIDHeader     string `json:"request_id_header"`

	// Logging
	LogRequests       bool `json:"log_requests"`
	LogDeniedRequests bool `json:"log_denied_requests"`
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		UnauthorizedStatusCode:  http.StatusUnauthorized,
		ForbiddenStatusCode:     http.StatusForbidden,
		ErrorStatusCode:         http.StatusInternalServerError,
		IncludeReasonInResponse: true,
		RequireAuthentication:   true,
		DefaultAction:           "read",
		SubjectHeader:           "X-Subject-ID",
		AuthorizationHeader:     "Authorization",
		RequestIDHeader:         "X-Request-ID",
		LogRequests:             true,
		LogDeniedRequests:       true,
		SkipPaths:               []string{"/health", "/metrics", "/favicon.ico"},
	}
}

// SubjectExtractor extracts subject information from HTTP request
type SubjectExtractor func(*http.Request) (string, error)

// ResourceExtractor extracts resource information from HTTP request
type ResourceExtractor func(*http.Request) (string, error)

// ContextExtractor extracts additional context from HTTP request
type ContextExtractor func(*http.Request) (map[string]interface{}, error)

// NewHTTPMiddleware creates a new HTTP middleware
func NewHTTPMiddleware(pep PEPInterface, config *MiddlewareConfig) *HTTPMiddleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return &HTTPMiddleware{
		pep:               pep,
		config:            config,
		subjectExtractor:  DefaultSubjectExtractor(config.SubjectHeader, config.AuthorizationHeader),
		resourceExtractor: DefaultResourceExtractor(),
		contextExtractor:  DefaultContextExtractor(),
	}
}

// SetSubjectExtractor sets a custom subject extractor
func (m *HTTPMiddleware) SetSubjectExtractor(extractor SubjectExtractor) {
	m.subjectExtractor = extractor
}

// SetResourceExtractor sets a custom resource extractor
func (m *HTTPMiddleware) SetResourceExtractor(extractor ResourceExtractor) {
	m.resourceExtractor = extractor
}

// SetContextExtractor sets a custom context extractor
func (m *HTTPMiddleware) SetContextExtractor(extractor ContextExtractor) {
	m.contextExtractor = extractor
}

// Handler returns the HTTP middleware handler
func (m *HTTPMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be skipped
		if m.shouldSkipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract request information
		subjectID, err := m.subjectExtractor(r)
		if err != nil {
			if m.config.RequireAuthentication {
				m.sendErrorResponse(w, m.config.UnauthorizedStatusCode, "Authentication required", err.Error())
				return
			}
			subjectID = "anonymous"
		}

		resourceID, err := m.resourceExtractor(r)
		if err != nil {
			m.sendErrorResponse(w, m.config.ErrorStatusCode, "Failed to extract resource", err.Error())
			return
		}

		requestContext, err := m.contextExtractor(r)
		if err != nil {
			m.sendErrorResponse(w, m.config.ErrorStatusCode, "Failed to extract context", err.Error())
			return
		}

		// Determine action from HTTP method
		action := m.httpMethodToAction(r.Method)
		if action == "" {
			action = m.config.DefaultAction
		}

		// Create evaluation request
		requestID := m.getRequestID(r)
		evalRequest := &models.EvaluationRequest{
			RequestID:  requestID,
			SubjectID:  subjectID,
			ResourceID: resourceID,
			Action:     action,
			Context:    requestContext,
		}

		// Enforce policy
		result, err := m.pep.EnforceRequest(r.Context(), evalRequest)
		if err != nil {
			m.sendErrorResponse(w, m.config.ErrorStatusCode, "Policy evaluation failed", err.Error())
			return
		}

		// Log request if configured
		if m.config.LogRequests || (m.config.LogDeniedRequests && !result.Allowed) {
			m.logRequest(r, evalRequest, result)
		}

		// Check if access is allowed
		if !result.Allowed {
			reason := "Access denied"
			if m.config.IncludeReasonInResponse && result.Reason != "" {
				reason = result.Reason
			}
			m.sendErrorResponse(w, m.config.ForbiddenStatusCode, reason, "")
			return
		}

		// Add enforcement result to request context
		reqCtx := context.WithValue(r.Context(), "pep_result", result)
		r = r.WithContext(reqCtx)

		// Add headers with enforcement information
		w.Header().Set("X-PEP-Decision", result.Decision.Result)
		w.Header().Set("X-PEP-Evaluation-Time", fmt.Sprintf("%dms", result.EvaluationTimeMs))
		if result.CacheHit {
			w.Header().Set("X-PEP-Cache-Hit", "true")
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// shouldSkipPath checks if the path should be skipped
func (m *HTTPMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// httpMethodToAction maps HTTP methods to ABAC actions
func (m *HTTPMiddleware) httpMethodToAction(method string) string {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

// getRequestID extracts or generates a request ID
func (m *HTTPMiddleware) getRequestID(r *http.Request) string {
	if requestID := r.Header.Get(m.config.RequestIDHeader); requestID != "" {
		return requestID
	}

	// Generate a simple request ID if not provided
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// sendErrorResponse sends an error response
func (m *HTTPMiddleware) sendErrorResponse(w http.ResponseWriter, statusCode int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":     message,
		"status":    statusCode,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if detail != "" && m.config.IncludeReasonInResponse {
		response["detail"] = detail
	}

	json.NewEncoder(w).Encode(response)
}

// logRequest logs the request and enforcement result
func (m *HTTPMiddleware) logRequest(r *http.Request, evalRequest *models.EvaluationRequest, result *EnforcementResult) {
	// This would typically use a structured logger
	fmt.Printf("[PEP] %s %s - Subject: %s, Resource: %s, Action: %s, Decision: %s, Time: %dms\n",
		r.Method, r.URL.Path, evalRequest.SubjectID, evalRequest.ResourceID,
		evalRequest.Action, result.Decision.Result, result.EvaluationTimeMs)
}

// DefaultSubjectExtractor creates a default subject extractor
func DefaultSubjectExtractor(subjectHeader, authHeader string) SubjectExtractor {
	return func(r *http.Request) (string, error) {
		// Try subject header first
		if subjectID := r.Header.Get(subjectHeader); subjectID != "" {
			return subjectID, nil
		}

		// Try to extract from Authorization header (JWT, etc.)
		if authValue := r.Header.Get(authHeader); authValue != "" {
			// Simple Bearer token extraction (in real implementation, you'd parse JWT)
			if strings.HasPrefix(authValue, "Bearer ") {
				token := strings.TrimPrefix(authValue, "Bearer ")
				// For demo purposes, map specific tokens to existing subjects
				if token == "test-token" {
					return "sub-001", nil // Map to existing engineering user
				}
				// Here you would typically decode JWT and extract subject
				return fmt.Sprintf("user_%s", token[:min(8, len(token))]), nil
			}
		}

		// Try basic auth
		if username, _, ok := r.BasicAuth(); ok {
			// For demo purposes, map specific usernames to existing subjects
			if username == "user" {
				return "sub-001", nil // Map to existing engineering user
			}
			return username, nil
		}

		return "", fmt.Errorf("no subject found in request")
	}
}

// DefaultResourceExtractor creates a default resource extractor
func DefaultResourceExtractor() ResourceExtractor {
	return func(r *http.Request) (string, error) {
		// Use the request path as resource ID
		resource := r.URL.Path

		// Clean up the path
		if resource == "" || resource == "/" {
			resource = "/root"
		}

		return resource, nil
	}
}

// DefaultContextExtractor creates a default context extractor
func DefaultContextExtractor() ContextExtractor {
	return func(r *http.Request) (map[string]interface{}, error) {
		context := map[string]interface{}{
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"method":       r.Method,
			"user_agent":   r.UserAgent(),
			"content_type": r.Header.Get("Content-Type"),
		}

		// Add IP address
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			context["source_ip"] = strings.Split(ip, ",")[0]
		} else if ip := r.Header.Get("X-Real-IP"); ip != "" {
			context["source_ip"] = ip
		} else {
			context["source_ip"] = r.RemoteAddr
		}

		// Add query parameters if present
		if len(r.URL.RawQuery) > 0 {
			context["query_params"] = r.URL.RawQuery
		}

		return context, nil
	}
}

// GetEnforcementResult extracts the enforcement result from request context
func GetEnforcementResult(r *http.Request) (*EnforcementResult, bool) {
	if result, ok := r.Context().Value("pep_result").(*EnforcementResult); ok {
		return result, true
	}
	return nil, false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RESTfulMiddleware provides RESTful API specific middleware
type RESTfulMiddleware struct {
	*HTTPMiddleware
}

// NewRESTfulMiddleware creates middleware optimized for RESTful APIs
func NewRESTfulMiddleware(pep PEPInterface) *RESTfulMiddleware {
	config := DefaultMiddlewareConfig()
	config.DefaultAction = "read"
	config.SkipPaths = []string{"/api/health", "/api/metrics", "/api/docs"}

	middleware := NewHTTPMiddleware(pep, config)

	// Set RESTful resource extractor
	middleware.SetResourceExtractor(func(r *http.Request) (string, error) {
		// Extract resource from RESTful path patterns
		path := r.URL.Path

		// Remove API prefix if present
		if strings.HasPrefix(path, "/api/") {
			path = strings.TrimPrefix(path, "/api")
		}
		if strings.HasPrefix(path, "/v1/") {
			path = strings.TrimPrefix(path, "/v1")
		}

		// Handle empty path
		if path == "" || path == "/" {
			path = "/root"
		}

		return path, nil
	})

	return &RESTfulMiddleware{HTTPMiddleware: middleware}
}
