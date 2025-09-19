package fallback

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockTrustStore for testing
type MockTrustStore struct {
	setInsecureCalls    []SetInsecureCall
	useSystemCertsCalls []UseSystemCertsCall
	addCertificateCalls []AddCertificateCall
	shouldFail          bool
	failMessage         string
}

type SetInsecureCall struct {
	Registry string
	Insecure bool
}

type UseSystemCertsCall struct {
	Registry string
	Use      bool
}

type AddCertificateCall struct {
	Registry string
	CertData []byte
}

func NewMockTrustStore() *MockTrustStore {
	return &MockTrustStore{
		setInsecureCalls:    make([]SetInsecureCall, 0),
		useSystemCertsCalls: make([]UseSystemCertsCall, 0),
		addCertificateCalls: make([]AddCertificateCall, 0),
	}
}

func (m *MockTrustStore) SetInsecure(registry string, insecure bool) error {
	m.setInsecureCalls = append(m.setInsecureCalls, SetInsecureCall{
		Registry: registry,
		Insecure: insecure,
	})

	if m.shouldFail {
		return errors.New(m.failMessage)
	}
	return nil
}

func (m *MockTrustStore) SetUseSystemCerts(registry string, use bool) error {
	m.useSystemCertsCalls = append(m.useSystemCertsCalls, UseSystemCertsCall{
		Registry: registry,
		Use:      use,
	})

	if m.shouldFail {
		return errors.New(m.failMessage)
	}
	return nil
}

func (m *MockTrustStore) AddCertificate(registry string, certData []byte) error {
	m.addCertificateCalls = append(m.addCertificateCalls, AddCertificateCall{
		Registry: registry,
		CertData: certData,
	})

	if m.shouldFail {
		return errors.New(m.failMessage)
	}
	return nil
}

func (m *MockTrustStore) SetShouldFail(shouldFail bool, message string) {
	m.shouldFail = shouldFail
	m.failMessage = message
}

// MockStrategy for testing
type MockStrategy struct {
	name        string
	priority    int
	shouldFail  bool
	failCount   int
	attempts    int
	shouldRetry bool
	executed    bool
}

func NewMockStrategy(name string, priority int) *MockStrategy {
	return &MockStrategy{
		name:        name,
		priority:    priority,
		shouldRetry: true,
	}
}

func (m *MockStrategy) Name() string {
	return m.name
}

func (m *MockStrategy) Priority() int {
	return m.priority
}

func (m *MockStrategy) Execute(ctx context.Context, registry string) error {
	m.attempts++
	m.executed = true

	if m.shouldFail && (m.failCount == 0 || m.attempts <= m.failCount) {
		return errors.New("mock strategy failed")
	}

	return nil
}

func (m *MockStrategy) ShouldRetry(err error) bool {
	return m.shouldRetry
}

func (m *MockStrategy) SetShouldFail(fail bool, failCount int, retry bool) {
	m.shouldFail = fail
	m.failCount = failCount
	m.shouldRetry = retry
}

// Test functions

func TestNewFallbackManager(t *testing.T) {
	mockStore := NewMockTrustStore()

	fm := NewFallbackManager(mockStore)

	if fm == nil {
		t.Fatal("Expected fallback manager to be created")
	}

	if fm.trustStore != mockStore {
		t.Error("Expected trust store to be set")
	}

	if fm.maxRetries != 3 {
		t.Errorf("Expected default max retries to be 3, got %d", fm.maxRetries)
	}

	if fm.retryDelay != time.Second {
		t.Errorf("Expected default retry delay to be 1s, got %v", fm.retryDelay)
	}

	if len(fm.strategies) != 3 {
		t.Errorf("Expected 3 default strategies, got %d", len(fm.strategies))
	}
}

func TestFallbackManager_InsecureMode(t *testing.T) {
	tests := []struct {
		name         string
		insecureMode bool
		expectBypass bool
	}{
		{
			name:         "insecure mode enabled",
			insecureMode: true,
			expectBypass: true,
		},
		{
			name:         "insecure mode disabled",
			insecureMode: false,
			expectBypass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockTrustStore()
			fm := NewFallbackManager(mockStore, WithInsecureMode(tt.insecureMode))

			if !tt.insecureMode {
				// Replace strategies with failing ones to test the "all strategies fail" path
				failingStrategy1 := NewMockStrategy("fail1", 1)
				failingStrategy1.SetShouldFail(true, 0, false)
				failingStrategy2 := NewMockStrategy("fail2", 2)
				failingStrategy2.SetShouldFail(true, 0, false)
				fm.strategies = []FallbackStrategy{failingStrategy1, failingStrategy2}
			}

			ctx := context.Background()
			err := fm.HandleValidationFailure(ctx, "test.registry", errors.New("cert error"))

			if tt.expectBypass {
				if err != nil {
					t.Errorf("Expected no error in insecure mode, got: %v", err)
				}
				if len(mockStore.setInsecureCalls) != 1 {
					t.Errorf("Expected 1 SetInsecure call, got %d", len(mockStore.setInsecureCalls))
				}
				if mockStore.setInsecureCalls[0].Registry != "test.registry" {
					t.Errorf("Expected registry 'test.registry', got '%s'", mockStore.setInsecureCalls[0].Registry)
				}
				if !mockStore.setInsecureCalls[0].Insecure {
					t.Error("Expected insecure to be true")
				}
			} else {
				if err == nil {
					t.Error("Expected error when not in insecure mode and all strategies fail")
				}
				// Verify SetInsecure was NOT called for bypassing (only for strategies)
				bypassCalls := 0
				for _, call := range mockStore.setInsecureCalls {
					if call.Registry == "test.registry" && call.Insecure {
						bypassCalls++
					}
				}
				if bypassCalls > 0 {
					t.Error("Expected no bypass calls to SetInsecure in non-insecure mode")
				}
			}
		})
	}
}

func TestFallbackManager_RetryLogic(t *testing.T) {
	mockStore := NewMockTrustStore()
	fm := NewFallbackManager(mockStore,
		WithMaxRetries(3),
		WithRetryDelay(10*time.Millisecond))

	// Create a mock strategy that fails twice then succeeds
	mockStrategy := NewMockStrategy("test-strategy", 1)
	mockStrategy.SetShouldFail(true, 2, true) // Fail first 2 attempts, then succeed
	fm.strategies = []FallbackStrategy{mockStrategy}

	ctx := context.Background()
	err := fm.HandleValidationFailure(ctx, "test.registry", errors.New("initial error"))

	if err != nil {
		t.Errorf("Expected no error after retry success, got: %v", err)
	}

	if mockStrategy.attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", mockStrategy.attempts)
	}
}

func TestFallbackManager_ContextCancellation(t *testing.T) {
	mockStore := NewMockTrustStore()
	fm := NewFallbackManager(mockStore)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := fm.HandleValidationFailure(ctx, "test.registry", errors.New("cert error"))

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestFallbackManager_AllStrategiesFail(t *testing.T) {
	mockStore := NewMockTrustStore()
	fm := NewFallbackManager(mockStore)

	// Replace strategies with failing mock strategies
	mockStrategy1 := NewMockStrategy("strategy1", 1)
	mockStrategy1.SetShouldFail(true, 0, false) // Always fail, no retry

	mockStrategy2 := NewMockStrategy("strategy2", 2)
	mockStrategy2.SetShouldFail(true, 0, false) // Always fail, no retry

	fm.strategies = []FallbackStrategy{mockStrategy1, mockStrategy2}

	ctx := context.Background()
	err := fm.HandleValidationFailure(ctx, "test.registry", errors.New("initial error"))

	if err == nil {
		t.Error("Expected error when all strategies fail")
	}

	if !mockStrategy1.executed {
		t.Error("Expected strategy1 to be executed")
	}

	if !mockStrategy2.executed {
		t.Error("Expected strategy2 to be executed")
	}
}

func TestFallbackManager_Options(t *testing.T) {
	mockStore := NewMockTrustStore()

	customWarningCalled := false
	customWarning := func(msg string) {
		customWarningCalled = true
	}

	fm := NewFallbackManager(mockStore,
		WithInsecureMode(true),
		WithMaxRetries(5),
		WithRetryDelay(2*time.Second),
		WithWarningCallback(customWarning))

	if !fm.IsInsecureMode() {
		t.Error("Expected insecure mode to be enabled")
	}

	if fm.maxRetries != 5 {
		t.Errorf("Expected max retries to be 5, got %d", fm.maxRetries)
	}

	if fm.retryDelay != 2*time.Second {
		t.Errorf("Expected retry delay to be 2s, got %v", fm.retryDelay)
	}

	// Test custom warning callback
	ctx := context.Background()
	fm.HandleValidationFailure(ctx, "test.registry", errors.New("cert error"))

	if !customWarningCalled {
		t.Error("Expected custom warning callback to be called")
	}
}

func TestFallbackManager_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		setupFM     func() *FallbackManager
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			setupFM: func() *FallbackManager {
				return NewFallbackManager(NewMockTrustStore())
			},
			expectError: false,
		},
		{
			name: "nil trust store",
			setupFM: func() *FallbackManager {
				return NewFallbackManager(nil)
			},
			expectError: true,
			errorMsg:    "trust store manager is required",
		},
		{
			name: "zero max retries",
			setupFM: func() *FallbackManager {
				return NewFallbackManager(NewMockTrustStore(), WithMaxRetries(0))
			},
			expectError: true,
			errorMsg:    "max retries must be at least 1",
		},
		{
			name: "negative retry delay",
			setupFM: func() *FallbackManager {
				return NewFallbackManager(NewMockTrustStore(), WithRetryDelay(-1*time.Second))
			},
			expectError: true,
			errorMsg:    "retry delay must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			err := fm.ValidateConfiguration()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}
