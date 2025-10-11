package pep

import (
	"fmt"
	"log"
	"os"
)

// SimpleAuditLogger is a basic implementation of AuditLogger interface
type SimpleAuditLogger struct {
	logger *log.Logger
}

// NewSimpleAuditLogger creates a new simple audit logger
func NewSimpleAuditLogger(filename string) (*SimpleAuditLogger, error) {
	if filename == "" {
		// Use stdout if no filename provided
		return &SimpleAuditLogger{
			logger: log.New(os.Stdout, "[AUDIT] ", log.LstdFlags),
		}, nil
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &SimpleAuditLogger{
		logger: log.New(file, "[AUDIT] ", log.LstdFlags),
	}, nil
}

// LogDecision logs a policy decision
func (sal *SimpleAuditLogger) LogDecision(data map[string]interface{}) {
	sal.logger.Printf("Decision: %+v", data)
}

// NoOpAuditLogger is a no-operation audit logger for testing
type NoOpAuditLogger struct{}

// NewNoOpAuditLogger creates a new no-op audit logger
func NewNoOpAuditLogger() *NoOpAuditLogger {
	return &NoOpAuditLogger{}
}

// LogDecision does nothing (no-op)
func (nol *NoOpAuditLogger) LogDecision(data map[string]interface{}) {
	// No-op
}
