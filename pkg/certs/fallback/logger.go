// Package fallback provides certificate fallback strategies with comprehensive security logging
package fallback

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SecurityLevel represents the severity of a security event
type SecurityLevel int

const (
	// SecurityInfo indicates informational security events
	SecurityInfo SecurityLevel = iota
	// SecurityWarning indicates potentially concerning security events
	SecurityWarning
	// SecurityCritical indicates high-risk security events requiring immediate attention
	SecurityCritical
	// SecurityBlocked indicates security events that resulted in access denial
	SecurityBlocked
)

// String returns the string representation of SecurityLevel
func (s SecurityLevel) String() string {
	switch s {
	case SecurityInfo:
		return "INFO"
	case SecurityWarning:
		return "WARNING"
	case SecurityCritical:
		return "CRITICAL"
	case SecurityBlocked:
		return "BLOCKED"
	default:
		return "UNKNOWN"
	}
}

// SecurityLogEntry represents a structured security log entry for certificate fallback decisions
type SecurityLogEntry struct {
	// Timestamp when the security event occurred
	Timestamp time.Time `json:"timestamp"`
	// Level indicates the severity of the security event
	Level SecurityLevel `json:"level"`
	// Action describes what security action was taken
	Action string `json:"action"`
	// Registry identifies the OCI registry involved
	Registry string `json:"registry"`
	// ErrorType categorizes the certificate error encountered
	ErrorType string `json:"error_type,omitempty"`
	// ErrorDetails provides specific information about the certificate error
	ErrorDetails string `json:"error_details,omitempty"`
	// FallbackStrategy describes the fallback approach applied
	FallbackStrategy string `json:"fallback_strategy,omitempty"`
	// UserAgent identifies the client making the request
	UserAgent string `json:"user_agent,omitempty"`
	// RemoteAddr contains the client's IP address
	RemoteAddr string `json:"remote_addr,omitempty"`
	// RequestID provides correlation with other logs
	RequestID string `json:"request_id,omitempty"`
	// Allowed indicates whether access was ultimately granted
	Allowed bool `json:"allowed"`
	// RiskScore provides a calculated risk assessment (0-100)
	RiskScore int `json:"risk_score,omitempty"`
	// Metadata contains additional context-specific information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityLogger provides structured security audit logging for certificate fallback decisions
type SecurityLogger interface {
	// LogFallbackDecision records a certificate fallback decision
	LogFallbackDecision(entry SecurityLogEntry) error
	// LogSecurityEvent records a general security event
	LogSecurityEvent(level SecurityLevel, action, registry, details string) error
	// LogAccessDenied records an access denial event
	LogAccessDenied(registry, reason, userAgent, remoteAddr string) error
	// LogAccessGranted records an access grant event with fallback details
	LogAccessGranted(registry, strategy, userAgent, remoteAddr string, riskScore int) error
	// Rotate initiates log rotation
	Rotate() error
	// Close closes the logger and flushes any buffered data
	Close() error
}

// FileSecurityLogger implements SecurityLogger with file-based logging
type FileSecurityLogger struct {
	// logDir is the directory where log files are stored
	logDir string
	// currentFile is the current log file
	currentFile *os.File
	// encoder handles JSON encoding
	encoder *json.Encoder
	// mutex protects concurrent access
	mutex sync.RWMutex
	// maxSize is the maximum log file size before rotation
	maxSize int64
	// maxAge is the maximum age before log rotation
	maxAge time.Duration
	// retention is how long to keep old log files
	retention time.Duration
	// currentSize tracks the current log file size
	currentSize int64
	// lastRotation tracks when the last rotation occurred
	lastRotation time.Time
	// enabledLevels controls which security levels are logged
	enabledLevels map[SecurityLevel]bool
}

// NewFileSecurityLogger creates a new file-based security logger
func NewFileSecurityLogger(logDir string, maxSizeMB int, maxAgeHours int, retentionDays int) (*FileSecurityLogger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Initialize logger
	logger := &FileSecurityLogger{
		logDir:        logDir,
		maxSize:       int64(maxSizeMB) * 1024 * 1024, // Convert MB to bytes
		maxAge:        time.Duration(maxAgeHours) * time.Hour,
		retention:     time.Duration(retentionDays) * 24 * time.Hour,
		lastRotation:  time.Now(),
		enabledLevels: make(map[SecurityLevel]bool),
	}

	// Enable all security levels by default
	logger.enabledLevels[SecurityInfo] = true
	logger.enabledLevels[SecurityWarning] = true
	logger.enabledLevels[SecurityCritical] = true
	logger.enabledLevels[SecurityBlocked] = true

	// Create initial log file
	if err := logger.createLogFile(); err != nil {
		return nil, fmt.Errorf("failed to create initial log file: %w", err)
	}

	return logger, nil
}

// createLogFile creates a new log file and sets up the encoder
func (f *FileSecurityLogger) createLogFile() error {
	filename := fmt.Sprintf("security-fallback-%s.jsonl", time.Now().Format("2006-01-02-15"))
	filepath := filepath.Join(f.logDir, filename)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filepath, err)
	}

	// Close existing file if any
	if f.currentFile != nil {
		f.currentFile.Close()
	}

	f.currentFile = file
	f.encoder = json.NewEncoder(file)
	f.currentSize = 0
	f.lastRotation = time.Now()

	// Get current file size
	if stat, err := file.Stat(); err == nil {
		f.currentSize = stat.Size()
	}

	return nil
}

// SetEnabledLevels configures which security levels should be logged
func (f *FileSecurityLogger) SetEnabledLevels(levels []SecurityLevel) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Reset all levels to disabled
	for level := range f.enabledLevels {
		f.enabledLevels[level] = false
	}

	// Enable specified levels
	for _, level := range levels {
		f.enabledLevels[level] = true
	}
}

// LogFallbackDecision records a certificate fallback decision
func (f *FileSecurityLogger) LogFallbackDecision(entry SecurityLogEntry) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check if this level is enabled
	if !f.enabledLevels[entry.Level] {
		return nil
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Check if rotation is needed
	if err := f.checkRotation(); err != nil {
		return fmt.Errorf("failed to check log rotation: %w", err)
	}

	// Write log entry
	if err := f.encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode log entry: %w", err)
	}

	// Update current size (approximate)
	entrySize, _ := json.Marshal(entry)
	f.currentSize += int64(len(entrySize)) + 1 // +1 for newline

	return nil
}

// LogSecurityEvent records a general security event
func (f *FileSecurityLogger) LogSecurityEvent(level SecurityLevel, action, registry, details string) error {
	entry := SecurityLogEntry{
		Timestamp:    time.Now(),
		Level:        level,
		Action:       action,
		Registry:     registry,
		ErrorDetails: details,
		Allowed:      level != SecurityBlocked,
	}

	return f.LogFallbackDecision(entry)
}

// LogAccessDenied records an access denial event
func (f *FileSecurityLogger) LogAccessDenied(registry, reason, userAgent, remoteAddr string) error {
	entry := SecurityLogEntry{
		Timestamp:    time.Now(),
		Level:        SecurityBlocked,
		Action:       "access_denied",
		Registry:     registry,
		ErrorDetails: reason,
		UserAgent:    userAgent,
		RemoteAddr:   remoteAddr,
		Allowed:      false,
		RiskScore:    100, // Maximum risk for denied access
	}

	return f.LogFallbackDecision(entry)
}

// LogAccessGranted records an access grant event with fallback details
func (f *FileSecurityLogger) LogAccessGranted(registry, strategy, userAgent, remoteAddr string, riskScore int) error {
	level := SecurityInfo
	if riskScore > 70 {
		level = SecurityCritical
	} else if riskScore > 40 {
		level = SecurityWarning
	}

	entry := SecurityLogEntry{
		Timestamp:        time.Now(),
		Level:            level,
		Action:           "access_granted",
		Registry:         registry,
		FallbackStrategy: strategy,
		UserAgent:        userAgent,
		RemoteAddr:       remoteAddr,
		Allowed:          true,
		RiskScore:        riskScore,
	}

	return f.LogFallbackDecision(entry)
}

// checkRotation checks if log rotation is needed and performs it if necessary
func (f *FileSecurityLogger) checkRotation() error {
	needsRotation := false

	// Check size-based rotation
	if f.maxSize > 0 && f.currentSize >= f.maxSize {
		needsRotation = true
	}

	// Check time-based rotation
	if f.maxAge > 0 && time.Since(f.lastRotation) >= f.maxAge {
		needsRotation = true
	}

	if needsRotation {
		return f.rotate()
	}

	return nil
}

// rotate performs log rotation
func (f *FileSecurityLogger) rotate() error {
	if err := f.createLogFile(); err != nil {
		return err
	}

	// Clean up old log files based on retention policy
	return f.cleanupOldLogs()
}

// Rotate initiates log rotation
func (f *FileSecurityLogger) Rotate() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.rotate()
}

// cleanupOldLogs removes log files older than the retention period
func (f *FileSecurityLogger) cleanupOldLogs() error {
	if f.retention <= 0 {
		return nil // No cleanup if retention is not set
	}

	cutoff := time.Now().Add(-f.retention)

	return filepath.Walk(f.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Only process our log files
		if !info.IsDir() && filepath.Ext(path) == ".jsonl" && 
		   len(info.Name()) > 15 && info.Name()[:15] == "security-fallback" {
			if info.ModTime().Before(cutoff) {
				if removeErr := os.Remove(path); removeErr != nil {
					// Log error but don't fail the cleanup
					fmt.Fprintf(os.Stderr, "Failed to remove old log file %s: %v\n", path, removeErr)
				}
			}
		}

		return nil
	})
}

// Close closes the logger and flushes any buffered data
func (f *FileSecurityLogger) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.currentFile != nil {
		if err := f.currentFile.Sync(); err != nil {
			f.currentFile.Close()
			return fmt.Errorf("failed to sync log file: %w", err)
		}
		return f.currentFile.Close()
	}

	return nil
}

// NewNullSecurityLogger creates a logger that discards all log entries (for testing)
func NewNullSecurityLogger() SecurityLogger {
	return &nullSecurityLogger{}
}

// nullSecurityLogger discards all log entries
type nullSecurityLogger struct{}

func (n *nullSecurityLogger) LogFallbackDecision(entry SecurityLogEntry) error { return nil }
func (n *nullSecurityLogger) LogSecurityEvent(level SecurityLevel, action, registry, details string) error { return nil }
func (n *nullSecurityLogger) LogAccessDenied(registry, reason, userAgent, remoteAddr string) error { return nil }
func (n *nullSecurityLogger) LogAccessGranted(registry, strategy, userAgent, remoteAddr string, riskScore int) error { return nil }
func (n *nullSecurityLogger) Rotate() error { return nil }
func (n *nullSecurityLogger) Close() error { return nil }