# Effort 1.1.2: Add Authentication Flags Implementation Plan

## 🚨 CRITICAL EFFORT METADATA (FROM WAVE PLAN)
**Branch**: `phase1-wave1-effort-1.1.2-auth-flags`
**Can Parallelize**: Yes
**Parallel With**: Efforts 1.1.1, 1.1.3
**Size Estimate**: ~250 lines
**Dependencies**: None (foundational effort)

## Overview
- **Effort**: Add authentication command-line flags for push command
- **Phase**: 1, Wave: 1
- **Estimated Size**: ~250 lines
- **Implementation Time**: 3-4 hours

## Scope
This effort implements the --username and --password flags for the push command, including validation logic and credential handling structures. The implementation focuses on command-line flag parsing, validation, and secure credential storage in memory.

## File Structure
```
cmd/
├── push.go                 # Add authentication flags (~80 lines)
pkg/
├── auth/
│   ├── flags.go           # Authentication flag definitions (~70 lines)
│   ├── validator.go       # Credential validation logic (~50 lines)
│   └── types.go           # Authentication types and structs (~30 lines)
tests/
├── cmd/
│   └── push_flags_test.go # Command flag tests (~20 lines)
```

## Implementation Steps

### Step 1: Define Authentication Types (30 lines)
**File**: `pkg/auth/types.go`
1. Create `Credentials` struct with Username and Password fields
2. Define `AuthConfig` struct for holding authentication configuration
3. Add constants for flag names ("username", "password")
4. Create error types for authentication failures
5. Add interface for credential validation

### Step 2: Create Flag Definitions (70 lines)
**File**: `pkg/auth/flags.go`
1. Import cobra and pflags packages
2. Create `AddAuthenticationFlags()` function
3. Define --username flag with description "Registry username"
4. Define --password flag with description "Registry password"
5. Add flag validation callbacks
6. Create helper function to extract credentials from flags
7. Add function to bind flags to viper config

### Step 3: Implement Credential Validation (50 lines)
**File**: `pkg/auth/validator.go`
1. Create `ValidateCredentials()` function
2. Check for empty username when password is provided
3. Check for empty password when username is provided
4. Validate username format (no special characters that could break URLs)
5. Add length validation (reasonable limits)
6. Return structured errors with helpful messages
7. Add helper to check if auth is required

### Step 4: Integrate Flags with Push Command (80 lines)
**File**: `cmd/push.go` (modifications)
1. Import auth package
2. Add authentication flags in init() or NewPushCommand()
3. Create credential extraction logic in RunE function
4. Add validation before push execution
5. Store credentials in command context
6. Add helper text for authentication flags
7. Update command long description with auth examples

### Step 5: Write Unit Tests (20 lines)
**File**: `tests/cmd/push_flags_test.go`
1. Test flag parsing with valid credentials
2. Test validation with missing username
3. Test validation with missing password
4. Test invalid username format rejection
5. Test flag help text generation

## Size Management
- **Estimated Lines**: ~250 lines total
- **Measurement Tool**: Will use `${PROJECT_ROOT}/tools/line-counter.sh`
- **Check Frequency**: After each major step
- **Split Threshold**: 700 lines (warning), 800 lines (stop)

## Test Requirements

### Unit Tests (Required)
- **Coverage Target**: 85%
- **Test Files**:
  - `pkg/auth/flags_test.go`: Flag creation and parsing
  - `pkg/auth/validator_test.go`: Validation logic
  - `pkg/auth/types_test.go`: Type construction

### Test Scenarios
1. **Valid Credentials**:
   - Both username and password provided
   - Credentials extracted correctly
   - No validation errors

2. **Missing Credentials**:
   - Missing username with password (error)
   - Missing password with username (error)
   - Both missing (valid - no auth)

3. **Invalid Formats**:
   - Username with invalid characters
   - Excessive length inputs
   - Empty strings vs nil

4. **Flag Integration**:
   - Flags appear in help text
   - Flags parse from command line
   - Flag values accessible in command

### Integration Test Stubs
- Mock credential validation against registry
- Simulate authentication flow (stub only)
- Test credential passing to registry client

## Pattern Compliance
- **Cobra Command Patterns**: Follow existing IDPBuilder command structure
- **Error Handling**: Use structured errors with context
- **Security**: No credential logging, clear sensitive data after use
- **Testing**: Test-first development approach

## Dependencies and Imports
```go
// Standard library
import (
    "errors"
    "fmt"
    "strings"
)

// Third-party
import (
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
)

// Internal (will be created)
import (
    "github.com/cnoe-io/idpbuilder/pkg/auth"
)
```

## Success Criteria
✅ Authentication flags appear in `idpbuilder push --help`
✅ Credentials are validated before use
✅ Clear error messages for invalid inputs
✅ Unit test coverage ≥85%
✅ No hardcoded credentials in code
✅ Implementation under 300 lines (target: ~250)

## Risk Mitigation
1. **Security Risk**: Credentials in memory
   - Mitigation: Clear after use, no logging
2. **Validation Complexity**: Over-engineering validation
   - Mitigation: Keep validation simple and focused
3. **Integration Risk**: Unknown registry auth requirements
   - Mitigation: Design for flexibility, interface-based

## Notes
- This is a foundational effort that other efforts will depend on
- Keep implementation simple - just flags and validation
- Actual authentication logic will be in Phase 2
- Focus on clean interfaces for future extension
- Ensure thread-safe credential handling for potential concurrent use

## Next Steps
After this effort is complete:
1. Effort 1.1.3 will add TLS configuration (parallel)
2. Wave 1.2 will add test infrastructure
3. Phase 2 will implement actual authentication with registry