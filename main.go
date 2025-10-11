package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"abac_go_example/evaluator"
	"abac_go_example/models"
	"abac_go_example/storage"
)

// ABACService - HTTP service với ABAC authorization
type ABACService struct {
	pdp     *evaluator.PolicyDecisionPoint
	storage storage.Storage
}

// ABACMiddleware - Middleware để check ABAC permissions
func (service *ABACService) ABACMiddleware(requiredAction string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Lấy subject từ header (trong thực tế sẽ từ JWT token)
			subjectID := r.Header.Get("X-Subject-ID")
			if subjectID == "" {
				http.Error(w, "Missing X-Subject-ID header", http.StatusUnauthorized)
				return
			}

			// Tạo evaluation request
			request := &models.EvaluationRequest{
				RequestID:  fmt.Sprintf("req_%d", time.Now().UnixNano()),
				SubjectID:  subjectID,
				ResourceID: r.URL.Path,
				Action:     requiredAction,
				Context: map[string]interface{}{
					"method":    r.Method,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"user_ip":   r.RemoteAddr,
				},
			}

			// Kiểm tra quyền với PDP
			decision, err := service.pdp.Evaluate(request)
			if err != nil {
				log.Printf("ABAC evaluation error: %v", err)
				http.Error(w, "Authorization error", http.StatusInternalServerError)
				return
			}

			// Log decision
			log.Printf("ABAC Decision: %s - Subject: %s, Resource: %s, Action: %s, Reason: %s",
				decision.Result, subjectID, r.URL.Path, requiredAction, decision.Reason)

			// Kiểm tra kết quả
			if decision.Result != "permit" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":    "Access denied",
					"reason":   decision.Reason,
					"subject":  subjectID,
					"resource": r.URL.Path,
					"action":   requiredAction,
				})
				return
			}

			// Cho phép request tiếp tục
			next.ServeHTTP(w, r)
		})
	}
}

// API Handlers
func (service *ABACService) handleUsers(w http.ResponseWriter, r *http.Request) {
	users := []map[string]interface{}{
		{"id": "1", "name": "John Doe", "department": "Engineering"},
		{"id": "2", "name": "Alice Smith", "department": "Finance"},
		{"id": "3", "name": "Bob Wilson", "department": "Engineering"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":   users,
		"message": "Users retrieved successfully",
	})
}

func (service *ABACService) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"user_id": "new_user_123",
	})
}

func (service *ABACService) handleFinancialData(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"revenue":  "$1,000,000",
		"expenses": "$800,000",
		"profit":   "$200,000",
		"quarter":  "Q1 2024",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"financial_data": data,
		"message":        "Financial data retrieved successfully",
	})
}

func (service *ABACService) handleAdminPanel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Admin panel accessed",
		"admin_functions": []string{"user_management", "system_config", "audit_logs"},
	})
}

// Health check endpoint (không cần ABAC)
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "ABAC Authorization Service",
	})
}

func main() {
	fmt.Println("🚀 Starting ABAC HTTP Service...")

	// Khởi tạo storage
	storage, err := storage.NewMockStorage(".")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Khởi tạo PDP
	pdp := evaluator.NewPolicyDecisionPoint(storage)

	// Khởi tạo service
	service := &ABACService{
		pdp:     pdp,
		storage: storage,
	}

	// Setup routes
	mux := http.NewServeMux()

	// Health check (không cần authorization)
	mux.HandleFunc("/health", handleHealth)

	// Protected endpoints với ABAC middleware
	mux.Handle("/api/v1/users", service.ABACMiddleware("read")(http.HandlerFunc(service.handleUsers)))
	mux.Handle("/api/v1/users/create", service.ABACMiddleware("write")(http.HandlerFunc(service.handleCreateUser)))
	mux.Handle("/api/v1/financial", service.ABACMiddleware("read")(http.HandlerFunc(service.handleFinancialData)))
	mux.Handle("/api/v1/admin", service.ABACMiddleware("admin")(http.HandlerFunc(service.handleAdminPanel)))

	// Debug: List all routes
	mux.HandleFunc("/debug/routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		routes := []string{"/health", "/api/v1/users", "/api/v1/users/create", "/api/v1/financial", "/api/v1/admin"}
		json.NewEncoder(w).Encode(map[string]interface{}{"routes": routes})
	})

	// CORS middleware (đơn giản)
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Subject-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	// HTTP server
	server := &http.Server{
		Addr:    ":8081",
		Handler: corsHandler(mux),
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n🛑 Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	fmt.Println("✅ ABAC HTTP Service started on :8081")
	fmt.Println("\n📋 Available endpoints:")
	fmt.Println("  GET  /health                    - Health check (no auth)")
	fmt.Println("  GET  /api/v1/users              - List users (read permission)")
	fmt.Println("  POST /api/v1/users/create       - Create user (write permission)")
	fmt.Println("  GET  /api/v1/financial          - Financial data (read permission)")
	fmt.Println("  GET  /api/v1/admin              - Admin panel (admin permission)")
	fmt.Println("\n💡 Usage examples:")
	fmt.Println("  curl http://localhost:8081/health")
	fmt.Println("  curl -H 'X-Subject-ID: sub-001' http://localhost:8081/api/v1/users")
	fmt.Println("  curl -H 'X-Subject-ID: sub-002' http://localhost:8081/api/v1/financial")
	fmt.Println("  curl -H 'X-Subject-ID: sub-004' http://localhost:8081/api/v1/users  # Should be denied")
	fmt.Println("\n🔑 Subject IDs in test data:")
	fmt.Println("  sub-001: John Doe (Engineering) - Can read APIs")
	fmt.Println("  sub-002: Alice Smith (Finance) - Can read financial data")
	fmt.Println("  sub-003: Payment Service - Service account")
	fmt.Println("  sub-004: Bob Wilson (On probation) - Limited access")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	fmt.Println("👋 Server stopped")
}
