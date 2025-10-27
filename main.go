package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"abac_go_example/evaluator/core"
	"abac_go_example/models"
	"abac_go_example/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🚀 Starting ABAC HTTP Service with Gin...")

	// Khởi tạo PostgreSQL storage
	dbConfig := storage.DefaultDatabaseConfig()
	storage, err := storage.NewPostgreSQLStorage(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL storage: %v", err)
	}
	defer storage.Close()

	// Khởi tạo PDP
	pdp := core.NewPolicyDecisionPoint(storage)

	// Khởi tạo service
	service := &ABACService{
		pdp:     pdp,
		storage: storage,
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health check (không cần authorization)
	router.GET("/health", handleHealth)

	// Protected endpoints với ABAC middleware
	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/users", service.ABACMiddleware("read"), service.handleUsers)
		apiV1.POST("/users/create", service.ABACMiddleware("write"), service.handleCreateUser)
		apiV1.GET("/financial", service.ABACMiddleware("read"), service.handleFinancialData)
		apiV1.GET("/admin", service.ABACMiddleware("admin"), service.handleAdminPanel)
	}

	// Debug: List all routes (Gin does this automatically in debug mode)
	// You can add a custom one if needed
	router.GET("/debug/routes", func(c *gin.Context) {
		routes := []gin.H{}
		for _, r := range router.Routes() {
			routes = append(routes, gin.H{"method": r.Method, "path": r.Path})
		}
		c.JSON(http.StatusOK, gin.H{"routes": routes})
	})

	// HTTP server
	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
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

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed to start: %v", err)
	}

	fmt.Println("👋 Server stopped")
}

// ABACService - HTTP service với ABAC authorization
type ABACService struct {
	pdp     core.PolicyDecisionPointInterface
	storage storage.Storage
}

// ABACMiddleware - Middleware để check ABAC permissions
func (service *ABACService) ABACMiddleware(requiredAction string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy subject từ header (trong thực tế sẽ từ JWT token)
		subjectID := c.GetHeader("X-Subject-ID")
		if subjectID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing X-Subject-ID header"})
			c.Abort()
			return
		}

		// Tạo evaluation request
		request := &models.EvaluationRequest{
			RequestID:  fmt.Sprintf("req_%d", time.Now().UnixNano()),
			SubjectID:  subjectID,
			ResourceID: c.Request.URL.Path,
			Action:     requiredAction,
			Context: map[string]interface{}{
				"method":    c.Request.Method,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"user_ip":   c.ClientIP(),
			},
		}

		// Kiểm tra quyền với PDP
		decision, err := service.pdp.Evaluate(request)
		if err != nil {
			log.Printf("ABAC evaluation error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization error"})
			c.Abort()
			return
		}

		// Log decision
		log.Printf("ABAC Decision: %s - Subject: %s, Resource: %s, Action: %s, Reason: %s",
			decision.Result, subjectID, c.Request.URL.Path, requiredAction, decision.Reason)

		// Kiểm tra kết quả
		if decision.Result != "permit" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Access denied",
				"reason":   decision.Reason,
				"subject":  subjectID,
				"resource": c.Request.URL.Path,
				"action":   requiredAction,
			})
			c.Abort()
			return
		}

		// Cho phép request tiếp tục
		c.Next()
	}
}

// API Handlers
func (service *ABACService) handleUsers(c *gin.Context) {
	users := []map[string]interface{}{
		{"id": "1", "name": "John Doe", "department": "Engineering"},
		{"id": "2", "name": "Alice Smith", "department": "Finance"},
		{"id": "3", "name": "Bob Wilson", "department": "Engineering"},
	}
	c.JSON(http.StatusOK, gin.H{
		"users":   users,
		"message": "Users retrieved successfully",
	})
}

func (service *ABACService) handleCreateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user_id": "new_user_123",
	})
}

func (service *ABACService) handleFinancialData(c *gin.Context) {
	data := map[string]interface{}{
		"revenue":  "$1,000,000",
		"expenses": "$800,000",
		"profit":   "$200,000",
		"quarter":  "Q1 2024",
	}
	c.JSON(http.StatusOK, gin.H{
		"financial_data": data,
		"message":        "Financial data retrieved successfully",
	})
}

func (service *ABACService) handleAdminPanel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":         "Admin panel accessed",
		"admin_functions": []string{"user_management", "system_config", "audit_logs"},
	})
}

// Health check endpoint (không cần ABAC)
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "ABAC Authorization Service",
	})
}

// CORS middleware (đơn giản)
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Subject-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
