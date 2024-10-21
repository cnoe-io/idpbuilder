package logger

import (
	"strings"
	"testing"
)

func TestFormatMessage(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Message without controller info",
			input:    "2024-10-20 17:06:42,779 INFO Simple log message without controller",
			expected: "2024-10-20 17:06:42,779 INFO Simple log message without controller",
		},
		{
			name:     "Message with insufficient parts",
			input:    "2024-10-20 17:06:42,779 INFO",
			expected: "2024-10-20 17:06:42,779 INFO",
		},
		{
			name:     "Long controller info",
			input:    "2024-10-20 17:06:42,779 INFO Message controller=localbuild controllerGroup=idpbuilder.cnoe.io controllerKind=Localbuild name=localdev name=localdev reconcileID=34cd11fb-3f43-4e1c-8582-ac37add91248 error=failed installing gitea: Internal error occurred: failed calling webhook validate.nginx.ingress.kubernetes.io: failed to call webhook: Post https://ingress-nginx-controller-admission.ingress-nginx.svc:443/networking/v1/ingresses?timeout=10s: dial tcp 10.96.14.62:443: connect: connection refused",
			expected: "2024-10-20 17:06:42,779 INFO Message \r\n                                 controller=localbuild controllerGroup=idpbuilder.cnoe.io\r\n                                 controllerKind=Localbuild name=localdev name=localdev reconcileID=34cd11fb-3f43-4e1c-8582-ac37add91248\r\n                                 error=failed installing gitea: Internal error occurred: failed calling webhook\r\n                                 validate.nginx.ingress.kubernetes.io: failed to call webhook: Post\r\n                                 https://ingress-nginx-controller-admission.ingress-nginx.svc:443/networking/v1/ingresses?timeout=10s: dial\r\n                                 tcp 10.96.14.62:443: connect: connection refused",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.Trim(formatMessage(tt.input), "\r\n")
			if result != tt.expected {
				t.Errorf("formatMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatMessageIndentation(t *testing.T) {
	input := "2024-10-20 17:06:42,779 INFO Message controller=test"
	result := strings.Trim(formatMessage(input), "\r\n")
	lines := strings.Split(result, "\r\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	firstLineLength := len("2024-10-20 17:06:42,779") + levelWidth
	expectedIndentation := strings.Repeat(" ", firstLineLength)

	for i, line := range lines[1:] {
		if !strings.HasPrefix(line, expectedIndentation) {
			t.Errorf("Line %d does not have correct indentation. Expected prefix: '%s', got: '%s'", i+2, expectedIndentation, line)
		}
	}
}
