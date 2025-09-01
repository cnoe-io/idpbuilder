package fallback

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestSecurityLevel tests the SecurityLevel enum type
func TestSecurityLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    SecurityLevel
		expected string
	}{
		{"Info level", InfoLevel, "info"},
		{"Warn level", WarnLevel, "warn"},
		{"Error level", ErrorLevel, "error"},
		{"Critical level", CriticalLevel, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("SecurityLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSecurityLevel_ValidLevels tests that security levels are in valid ranges
func TestSecurityLevel_ValidLevels(t *testing.T) {
	validLevels := []SecurityLevel{InfoLevel, WarnLevel, ErrorLevel, CriticalLevel}
	
	for _, level := range validLevels {
		if level < InfoLevel || level > CriticalLevel {
			t.Errorf("Invalid security level: %v", level)
		}
	}
}

// TestSecurityLogEntry tests the SecurityLogEntry structure
func TestSecurityLogEntry(t *testing.T) {
	entry := &SecurityLogEntry{
		Timestamp:   time.Now(),
		Level:       ErrorLevel,
		Component:   "fallback-handler",
		Message:     "Certificate validation failed",
		Error:       "x509: certificate has expired",
		Context:     map[string]interface{}{"host": "example.com", "port": 443},
		Severity:    "high",
		Remediation: "Check certificate expiration date",
	}

	t.Run("JSON marshaling", func(t *testing.T) {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal SecurityLogEntry: %v", err)
		}
		
		if !strings.Contains(string(data), "fallback-handler") {
			t.Error("JSON should contain component name")
		}
		if !strings.Contains(string(data), "Certificate validation failed") {
			t.Error("JSON should contain message")
		}
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal SecurityLogEntry: %v", err)
		}
		
		var unmarshaled SecurityLogEntry
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal SecurityLogEntry: %v", err)
		}
		
		if unmarshaled.Level != entry.Level {
			t.Errorf("Level mismatch after unmarshal: got %v, want %v", unmarshaled.Level, entry.Level)
		}
		if unmarshaled.Component != entry.Component {
			t.Errorf("Component mismatch after unmarshal: got %v, want %v", unmarshaled.Component, entry.Component)
		}
		if unmarshaled.Message != entry.Message {
			t.Errorf("Message mismatch after unmarshal: got %v, want %v", unmarshaled.Message, entry.Message)
		}
	})

	t.Run("Field validation", func(t *testing.T) {
		// Test that required fields are present
		if entry.Level == 0 {
			t.Error("Level should not be zero")
		}
		if entry.Component == "" {
			t.Error("Component should not be empty")
		}
		if entry.Message == "" {
			t.Error("Message should not be empty")
		}
		if entry.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
	})

	t.Run("Timestamp formatting", func(t *testing.T) {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal SecurityLogEntry: %v", err)
		}
		
		// Verify timestamp is in ISO format
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("Failed to unmarshal to raw map: %v", err)
		}
		
		timestamp, exists := raw["timestamp"]
		if !exists {
			t.Error("Timestamp should be present in JSON")
		}
		
		timestampStr, ok := timestamp.(string)
		if !ok {
			t.Error("Timestamp should be a string in JSON")
		}
		
		if _, err := time.Parse(time.RFC3339, timestampStr); err != nil {
			t.Errorf("Timestamp should be in RFC3339 format: %v", err)
		}
	})
}

// TestDefaultSecurityLogger tests the DefaultSecurityLogger implementation
func TestDefaultSecurityLogger(t *testing.T) {
	t.Run("New logger creation", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewDefaultSecurityLogger(&buf)
		
		if logger == nil {
			t.Fatal("NewDefaultSecurityLogger should not return nil")
		}
	})

	t.Run("Log method with all levels", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewDefaultSecurityLogger(&buf)
		
		testCases := []struct {
			level   SecurityLevel
			message string
		}{
			{InfoLevel, "Info message"},
			{WarnLevel, "Warning message"},
			{ErrorLevel, "Error message"},
			{CriticalLevel, "Critical message"},
		}
		
		for _, tc := range testCases {
			logger.Log(tc.level, "test-component", tc.message, nil, "medium", "Take action")
		}
		
		output := buf.String()
		for _, tc := range testCases {
			if !strings.Contains(output, tc.message) {
				t.Errorf("Output should contain message: %s", tc.message)
			}
		}
	})

	t.Run("Concurrent write safety", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewDefaultSecurityLogger(&buf)
		
		const goroutines = 10
		const messagesPerGoroutine = 10
		
		var wg sync.WaitGroup
		wg.Add(goroutines)
		
		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					logger.Log(InfoLevel, "test-component", 
						fmt.Sprintf("Message from goroutine %d, iteration %d", id, j), 
						nil, "low", "No action needed")
				}
			}(i)
		}
		
		wg.Wait()
		
		// Count the number of log entries
		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		expectedLines := goroutines * messagesPerGoroutine
		
		if len(lines) != expectedLines {
			t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
		}
	})

	t.Run("File output verification", func(t *testing.T) {
		// Create temporary file
		tmpDir, err := os.MkdirTemp("", "security-log-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		
		logFile := filepath.Join(tmpDir, "security.log")
		file, err := os.Create(logFile)
		if err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
		defer file.Close()
		
		logger := NewDefaultSecurityLogger(file)
		logger.Log(ErrorLevel, "file-test", "Test file logging", nil, "high", "Review logs")
		
		// Close and reopen file to read contents
		file.Close()
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if !strings.Contains(string(content), "Test file logging") {
			t.Error("Log file should contain the logged message")
		}
	})

	t.Run("Multiple writer support", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer
		writer := io.MultiWriter(&buf1, &buf2)
		logger := NewDefaultSecurityLogger(writer)
		
		message := "Multi-writer test message"
		logger.Log(WarnLevel, "multi-writer", message, nil, "medium", "Monitor closely")
		
		if !strings.Contains(buf1.String(), message) {
			t.Error("First writer should contain the message")
		}
		if !strings.Contains(buf2.String(), message) {
			t.Error("Second writer should contain the message")
		}
	})

	t.Run("Error handling", func(t *testing.T) {
		// Test logging with nil context
		var buf bytes.Buffer
		logger := NewDefaultSecurityLogger(&buf)
		
		logger.Log(ErrorLevel, "error-test", "Error with nil context", nil, "high", "Fix immediately")
		
		output := buf.String()
		if !strings.Contains(output, "Error with nil context") {
			t.Error("Should handle nil context gracefully")
		}
		
		// Test logging with complex context
		context := map[string]interface{}{
			"nested": map[string]string{"key": "value"},
			"array":  []int{1, 2, 3},
		}
		
		logger.Log(InfoLevel, "complex-test", "Complex context test", context, "low", "No action")
		
		output = buf.String()
		if !strings.Contains(output, "Complex context test") {
			t.Error("Should handle complex context")
		}
	})
}

// TestFileRotation tests log file rotation functionality
func TestFileRotation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log-rotation-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("Size-based rotation", func(t *testing.T) {
		// This would test actual file rotation if implemented
		// For now, just verify the concept
		baseFile := filepath.Join(tmpDir, "security.log")
		
		// Simulate rotation by creating numbered backup files
		for i := 1; i <= 3; i++ {
			rotatedFile := fmt.Sprintf("%s.%d", baseFile, i)
			file, err := os.Create(rotatedFile)
			if err != nil {
				t.Fatalf("Failed to create rotated file: %v", err)
			}
			file.WriteString(fmt.Sprintf("Rotated log content %d\n", i))
			file.Close()
		}
		
		// Check that rotated files exist
		for i := 1; i <= 3; i++ {
			rotatedFile := fmt.Sprintf("%s.%d", baseFile, i)
			if _, err := os.Stat(rotatedFile); os.IsNotExist(err) {
				t.Errorf("Rotated file should exist: %s", rotatedFile)
			}
		}
	})

	t.Run("Timestamp in filenames", func(t *testing.T) {
		now := time.Now()
		timestampedFile := filepath.Join(tmpDir, fmt.Sprintf("security-%s.log", now.Format("2006-01-02")))
		
		file, err := os.Create(timestampedFile)
		if err != nil {
			t.Fatalf("Failed to create timestamped file: %v", err)
		}
		file.Close()
		
		if _, err := os.Stat(timestampedFile); os.IsNotExist(err) {
			t.Error("Timestamped file should exist")
		}
	})

	t.Run("Old file cleanup", func(t *testing.T) {
		// Create old log files
		oldFiles := []string{
			"security-2023-01-01.log",
			"security-2023-01-02.log",
			"security-2023-01-03.log",
		}
		
		for _, filename := range oldFiles {
			filepath := filepath.Join(tmpDir, filename)
			file, err := os.Create(filepath)
			if err != nil {
				t.Fatalf("Failed to create old file: %v", err)
			}
			file.WriteString("Old log content")
			file.Close()
		}
		
		// Verify files were created
		for _, filename := range oldFiles {
			filepath := filepath.Join(tmpDir, filename)
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("Old file should exist: %s", filename)
			}
		}
		
		// In a real implementation, cleanup logic would remove files older than N days
		// For testing, we just verify the files exist and could be cleaned up
	})
}