<<<<<<< HEAD
# Effort 1.1.1: Push Command Skeleton Implementation Plan

## 🚨 CRITICAL EFFORT METADATA (FROM WAVE PLAN)
**Branch**: `phase1/wave1/effort-1.1.1-push-command-skeleton`
**Can Parallelize**: Yes
**Parallel With**: [1.1.2, 1.1.3]
**Size Estimate**: ~350 lines (well under 800 limit)
**Dependencies**: None (foundational effort)
**Base Branch**: main

## Overview
- **Effort**: Create basic push command structure with cobra
- **Phase**: 1, Wave: 1
- **Estimated Size**: ~350 lines
- **Implementation Time**: 3-4 hours

## Scope Statement (R371 Compliance)
This effort implements ONLY the basic push command structure and registration. It does NOT include:
- Authentication logic (effort 1.1.2)
- TLS configuration (effort 1.1.3)
- Registry client implementation (Phase 2)
- Image handling logic (Phase 3)
- Actual push functionality (Phase 4)

## Theme Coherence (R372 Compliance)
**Single Theme**: Basic CLI command structure and registration
- All changes support creating the push command skeleton
- No mixed concerns with auth, TLS, or registry logic
- Pure CLI framework setup

## Architectural Compliance (R362)
**Framework Used**: spf13/cobra (existing in project)
- Following existing IDPBuilder command patterns
- No custom CLI framework implementation
- Using standard cobra command structure

## Library Version Requirements (R381)
**Locked Dependencies (DO NOT UPDATE)**:
- github.com/spf13/cobra: Use existing version in go.mod
- github.com/spf13/viper: Use existing version if needed

**New Dependencies**: None for this effort

## File Structure
```
idpbuilder-gitea-push/
├── cmd/
│   ├── root.go         # [MODIFY] Register push command (~10 lines)
│   └── push.go         # [CREATE] Push command implementation (~200 lines)
├── cmd/push_test.go    # [CREATE] Command tests (~100 lines)
└── docs/
    └── push-help.txt   # [CREATE] Help text template (~40 lines)
=======
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
>>>>>>> effort2/phase1-wave1-effort-1.1.2-auth-flags
```

## Implementation Steps

<<<<<<< HEAD
### Step 1: Create Push Command Structure
**File**: `cmd/push.go`
**Lines**: ~200

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
    Use:   "push IMAGE_NAME",
    Short: "Push an OCI image to the integrated Gitea registry",
    Long: `Push an OCI image to the integrated Gitea registry.

The push command uploads container images to the Gitea registry
at https://gitea.cnoe.localtest.me:8443/.

Examples:
  # Push an image with authentication
  idpbuilder push myapp:latest --username admin --password secret

  # Push with insecure TLS (self-signed certificates)
  idpbuilder push myapp:latest --username admin --password secret --insecure`,
    Args: cobra.ExactArgs(1),
    RunE: runPush,
}

// Command configuration
type pushConfig struct {
    imageName string
    // Placeholder for future flags (auth, TLS)
}

func init() {
    // Command will be registered in root.go
    // Future flags will be added in subsequent efforts
}

// runPush executes the push command
func runPush(cmd *cobra.Command, args []string) error {
    config := &pushConfig{
        imageName: args[0],
    }

    // Validate image name format
    if err := validateImageName(config.imageName); err != nil {
        return fmt.Errorf("invalid image name: %w", err)
    }

    // Log command execution (temporary until implementation)
    fmt.Printf("Pushing image: %s\n", config.imageName)
    fmt.Println("Note: Push functionality will be implemented in Phase 4")

    return nil
}

// validateImageName performs basic validation on the image name
func validateImageName(name string) error {
    if name == "" {
        return fmt.Errorf("image name cannot be empty")
    }

    // Basic validation - will be enhanced in Phase 3
    // Check for basic format: [registry/]namespace/name[:tag]
    // For now, just ensure non-empty

    return nil
}
```

### Step 2: Register Command in Root
**File**: `cmd/root.go`
**Lines**: ~10 (modification)

```go
// Add to init() function or command registration area:
func init() {
    // ... existing init code ...

    // Register push command
    rootCmd.AddCommand(pushCmd)
}
```

### Step 3: Create Command Tests
**File**: `cmd/push_test.go`
**Lines**: ~100

```go
package cmd

import (
    "bytes"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPushCommand(t *testing.T) {
    tests := []struct {
        name      string
        args      []string
        wantErr   bool
        errMsg    string
    }{
        {
            name:    "valid image name",
            args:    []string{"push", "myapp:latest"},
            wantErr: false,
        },
        {
            name:    "missing image name",
            args:    []string{"push"},
            wantErr: true,
            errMsg:  "requires exactly 1 arg(s)",
        },
        {
            name:    "too many arguments",
            args:    []string{"push", "image1", "image2"},
            wantErr: true,
            errMsg:  "requires exactly 1 arg(s)",
        },
        {
            name:    "empty image name",
            args:    []string{"push", ""},
            wantErr: true,
            errMsg:  "image name cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := rootCmd
            buf := new(bytes.Buffer)
            cmd.SetOut(buf)
            cmd.SetErr(buf)
            cmd.SetArgs(tt.args)

            err := cmd.Execute()

            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestValidateImageName(t *testing.T) {
    tests := []struct {
        name      string
        imageName string
        wantErr   bool
    }{
        {
            name:      "valid simple name",
            imageName: "myapp",
            wantErr:   false,
        },
        {
            name:      "valid with tag",
            imageName: "myapp:latest",
            wantErr:   false,
        },
        {
            name:      "valid with namespace",
            imageName: "namespace/myapp:v1.0",
            wantErr:   false,
        },
        {
            name:      "empty name",
            imageName: "",
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateImageName(tt.imageName)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Step 4: Create Help Documentation
**File**: `docs/push-help.txt`
**Lines**: ~40

```
IDPBuilder Push Command

Usage:
  idpbuilder push IMAGE_NAME [flags]

Description:
  Push an OCI image to the integrated Gitea registry at
  https://gitea.cnoe.localtest.me:8443/

Arguments:
  IMAGE_NAME    The name of the image to push
                Format: [namespace/]name[:tag]
                Examples: myapp:latest, team/myapp:v1.0

Flags:
  -h, --help    help for push

Global Flags:
  [inherited from root command]

Examples:
  # Basic push (will add auth flags in effort 1.1.2)
  idpbuilder push myapp:latest

  # Push with specific tag
  idpbuilder push myapp:v1.0.0

  # Push with namespace
  idpbuilder push myteam/myapp:latest

Exit Codes:
  0    Success
  1    General error
  2    Invalid arguments

Notes:
  - Authentication flags will be added in effort 1.1.2
  - TLS configuration will be added in effort 1.1.3
  - Full push functionality will be implemented in Phase 4
```

## Size Management
- **Estimated Lines**: ~350 lines total
  - cmd/push.go: ~200 lines
  - cmd/push_test.go: ~100 lines
  - docs/push-help.txt: ~40 lines
  - cmd/root.go modification: ~10 lines
- **Measurement Tool**: ${PROJECT_ROOT}/tools/line-counter.sh
- **Check Frequency**: After each file completion
- **Split Threshold**: 700 lines (warning), 800 lines (stop)

## Test Requirements
- **Unit Tests**: 90% coverage target
  - Command registration test
  - Argument validation tests
  - Image name validation tests
  - Help text display test
- **Integration Tests**: Not required for skeleton
- **E2E Tests**: Not required for skeleton
- **Test Files**:
  - cmd/push_test.go (primary test file)

## Pattern Compliance
- **Cobra Patterns**: Follow existing IDPBuilder command structure
- **Error Handling**: Use fmt.Errorf with error wrapping
- **Logging**: Use existing logging framework when available
- **Naming**: Follow Go naming conventions

## Dependencies and Integration Points
- **Dependencies**: None (using existing cobra)
- **Integration with 1.1.2**: Auth flags will extend pushCmd
- **Integration with 1.1.3**: TLS flags will extend pushCmd
- **Integration with Phase 4**: runPush will call push orchestrator

## Success Criteria
- ✅ Push command registered and appears in help
- ✅ Command accepts exactly one argument (image name)
- ✅ Basic validation of image name
- ✅ All tests passing with >90% coverage
- ✅ Help text is clear and comprehensive
- ✅ Code follows project patterns

## Risk Mitigation
1. **Risk**: Command conflicts with existing commands
   - **Mitigation**: Review existing commands first
   - **Validation**: Test command registration

2. **Risk**: Image name validation too restrictive
   - **Mitigation**: Keep validation minimal in skeleton
   - **Enhancement**: Full validation in Phase 3

## Implementation Order
1. Create cmd/push.go with basic structure
2. Add command registration to cmd/root.go
3. Write tests in cmd/push_test.go
4. Create help documentation
5. Run tests and verify coverage
6. Measure with line-counter.sh

## Notes for SW Engineer
- Start with test-first approach: write tests before implementation
- Keep implementation simple - this is just the skeleton
- Don't implement actual push logic - that's Phase 4
- Ensure command follows existing IDPBuilder patterns
- Validate that command appears in `idpbuilder --help`
=======
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
>>>>>>> effort2/phase1-wave1-effort-1.1.2-auth-flags
