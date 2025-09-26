# effort-1.2.2-command-testing-framework Implementation Plan

## CRITICAL EFFORT METADATA (FROM WAVE PLAN)
**Branch**: `igp/phase1/wave2/effort-1.2.2-command-testing-framework`
**Can Parallelize**: No (Sequential - depends on effort-1.2.1)
**Parallel With**: None
**Size Estimate**: ~300 lines
**Dependencies**: effort-1.2.1-test-fixtures-setup
**Base Branch**: phase1-wave1-integration

## EFFORT INFRASTRUCTURE METADATA
**EFFORT_NAME**: effort-1.2.2-command-testing-framework
**EFFORT_DIR**: /home/vscode/workspaces/idpbuilder-gitea-push/efforts/phase1/wave2/effort-1.2.2-command-testing-framework
**BRANCH**: igp/phase1/wave2/effort-1.2.2-command-testing-framework
**BASE_BRANCH**: phase1-wave1-integration
**PHASE**: 1
**WAVE**: 2

## Overview
- **Effort**: Implement comprehensive command testing framework for push command
- **Phase**: 1, Wave: 2
- **Theme**: Test framework and command testing
- **Estimated Size**: 300 lines
- **Implementation Time**: 3-4 hours

## Dependencies Analysis (R219)
This effort depends on:

### 1. Wave 1 Functionality (Complete)
- **effort-1.1.1-push-command-skeleton**: Basic push command structure with Cobra
- **effort-1.1.2-auth-flags**: Authentication flags (--username, --password)
- **effort-1.1.3-tls-config**: TLS configuration (--insecure-tls flag)

### 2. effort-1.2.1-test-fixtures-setup (Must be completed first)
- Test fixtures and mock registries
- Helper functions for test setup
- Common test utilities

## File Structure

### New Files to Create
```
cmd/
├── push_test.go                    # Unit tests for push command (~200 lines)
    ├── Test_PushCommand_Basic      # Basic command creation
    ├── Test_PushCommand_Flags      # Flag parsing and validation
    ├── Test_PushCommand_Execute    # Command execution with mocks
    ├── Test_PushCommand_Auth       # Authentication flag handling
    ├── Test_PushCommand_TLS        # TLS configuration tests
    └── Test_PushCommand_Errors     # Error handling scenarios

test/
├── integration/                    # Integration test structure (~100 lines)
    ├── push_integration_test.go    # End-to-end push command tests
    │   ├── TestPushIntegration_BasicFlow
    │   ├── TestPushIntegration_WithAuth
    │   └── TestPushIntegration_WithTLS
    └── suite_test.go               # Test suite setup and teardown
```

### Files to Modify
```
None - This effort creates new test files only
```

## Implementation Steps

### Step 1: Create cmd/push_test.go (200 lines)
```go
package cmd

import (
    "bytes"
    "testing"

    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    // Import test fixtures from effort-1.2.1
    "github.com/cnoe-io/idpbuilder/test/fixtures"
    "github.com/cnoe-io/idpbuilder/test/helpers"
)

// Test basic command creation
func Test_PushCommand_Basic(t *testing.T) {
    // Verify command exists
    // Check command metadata (Use, Short, Long)
    // Verify subcommand registration
}

// Test flag parsing
func Test_PushCommand_Flags(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantUser string
        wantPass string
        wantTLS  bool
    }{
        // Test cases for various flag combinations
    }
    // Run table-driven tests
}

// Test command execution with mocks
func Test_PushCommand_Execute(t *testing.T) {
    // Use fixtures from effort-1.2.1
    // Mock registry interactions
    // Verify execution flow
}

// Test authentication handling
func Test_PushCommand_Auth(t *testing.T) {
    // Test with valid credentials
    // Test with invalid credentials
    // Test without credentials
}

// Test TLS configuration
func Test_PushCommand_TLS(t *testing.T) {
    // Test with insecure-tls flag
    // Test with secure connection
    // Test certificate validation
}

// Test error scenarios
func Test_PushCommand_Errors(t *testing.T) {
    // Test missing required args
    // Test invalid image format
    // Test network errors
    // Test registry errors
}
```

### Step 2: Create test/integration Directory Structure (50 lines)
```go
// test/integration/suite_test.go
package integration

import (
    "testing"

    "github.com/stretchr/testify/suite"
)

type PushIntegrationSuite struct {
    suite.Suite
    // Test registry setup from effort-1.2.1
    registry *fixtures.MockRegistry
}

func (suite *PushIntegrationSuite) SetupSuite() {
    // Initialize test registry
    // Setup test environment
}

func (suite *PushIntegrationSuite) TearDownSuite() {
    // Cleanup test resources
}

func TestPushIntegrationSuite(t *testing.T) {
    suite.Run(t, new(PushIntegrationSuite))
}
```

### Step 3: Create Integration Tests (50 lines)
```go
// test/integration/push_integration_test.go
package integration

import (
    "github.com/cnoe-io/idpbuilder/cmd"
    "github.com/cnoe-io/idpbuilder/test/fixtures"
)

func (suite *PushIntegrationSuite) TestPushIntegration_BasicFlow() {
    // End-to-end test of push command
    // Use real command execution
    // Verify against mock registry
}

func (suite *PushIntegrationSuite) TestPushIntegration_WithAuth() {
    // Test with authentication
    // Verify credentials are used correctly
}

func (suite *PushIntegrationSuite) TestPushIntegration_WithTLS() {
    // Test TLS configuration
    // Verify secure connection handling
}
```

## Test Requirements

### Unit Tests
- **Coverage Target**: 85%
- **Test Types**:
  - Command creation and configuration
  - Flag parsing and validation
  - Mock-based execution tests
  - Error handling scenarios

### Integration Tests
- **Coverage Target**: 70%
- **Test Types**:
  - End-to-end command execution
  - Real registry interaction (mocked)
  - Authentication flow
  - TLS configuration

### Test Files Expected
```
cmd/push_test.go
test/integration/push_integration_test.go
test/integration/suite_test.go
```

## Implementation Order

1. **First**: Wait for effort-1.2.1 completion
   - Need test fixtures and helpers
   - Need mock registry implementation

2. **Create Unit Tests** (cmd/push_test.go)
   - Start with basic command tests
   - Add flag parsing tests
   - Implement mock-based execution tests
   - Add error handling tests

3. **Create Integration Framework** (test/integration/)
   - Set up test suite structure
   - Initialize test environment

4. **Implement Integration Tests**
   - Basic flow tests
   - Authentication tests
   - TLS configuration tests

## Size Management
- **Estimated Lines**: 300
- **Current Breakdown**:
  - cmd/push_test.go: ~200 lines
  - test/integration/: ~100 lines
- **Measurement Tool**: `${PROJECT_ROOT}/tools/line-counter.sh`
- **Check Frequency**: After each test file completion
- **Split Threshold**: 700 lines (warning), 800 lines (stop)

## Pattern Compliance

### Go Testing Patterns
- Use testify for assertions
- Table-driven tests where appropriate
- Test fixtures from effort-1.2.1
- Clear test naming (Test_Component_Scenario)

### Cobra Command Testing
- Test command structure separately from execution
- Mock command context for unit tests
- Use real command for integration tests

### Project-Specific Patterns
- Follow idpbuilder test conventions
- Use existing test helper functions
- Maintain consistent error messaging

## Security Considerations
- Test credential handling without exposing secrets
- Verify TLS configuration is properly tested
- Ensure no hardcoded test credentials

## Dependencies on Previous Work

### From Wave 1:
- Push command implementation (effort-1.1.1)
- Authentication flags (effort-1.1.2)
- TLS configuration (effort-1.1.3)

### From effort-1.2.1:
- Test fixtures and mock registries
- Helper functions for test setup
- Common test utilities

## Success Criteria
- All tests pass
- Unit test coverage >= 85%
- Integration test coverage >= 70%
- No flaky tests
- Clear test documentation
- Tests verify all Wave 1 functionality

## Notes for SW Engineer
- This is a SEQUENTIAL effort - wait for effort-1.2.1 completion
- Import test helpers from effort-1.2.1 package
- Use mock registry from fixtures for all tests
- Follow existing idpbuilder test patterns
- Ensure tests are isolated and repeatable
- Document any complex test scenarios

## Review Checklist
- [ ] All Wave 1 functionality has tests
- [ ] Unit test coverage meets target
- [ ] Integration tests are comprehensive
- [ ] No hardcoded values or secrets
- [ ] Tests use fixtures from effort-1.2.1
- [ ] All tests pass consistently
- [ ] Size within 300 line limit