# Registry Tests Implementation Plan

## EFFORT INFRASTRUCTURE METADATA
**EFFORT_NAME**: registry-tests
**EFFORT_ID**: E1.1.2D
**PHASE**: 1
**WAVE**: 1
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/registry-tests
**BRANCH**: idpbuilder-oci-build-push/phase1/wave1/registry-tests
**BASE_BRANCH**: idpbuilder-oci-build-push/phase1/wave1/registry-helpers
**REMOTE**: origin
**PLAN_CREATED**: 2025-09-18T15:43:43Z
**PLAN_CREATOR**: code-reviewer

## CRITICAL EFFORT METADATA (FROM WAVE PLAN)
**Branch**: `idpbuilder-oci-build-push/phase1/wave1/registry-tests`
**Can Parallelize**: No
**Parallel With**: None
**Size Estimate**: 115 lines
**Dependencies**: [registry-types (E1.1.2A), registry-auth (E1.1.2B), registry-helpers (E1.1.2C)]
**Base Branch**: `idpbuilder-oci-build-push/phase1/wave1/registry-helpers`

## Overview
- **Effort**: Registry Tests - Comprehensive test coverage for registry types package
- **Phase**: 1, Wave: 1
- **Estimated Size**: 115 lines (CRITICAL CONSTRAINT - MUST NOT EXCEED)
- **Implementation Time**: 1-2 hours

## Critical Context

### Existing Test Coverage
Based on current branch analysis:
- **`pkg/registry/auth/*_test.go`**: 2207 lines (COMPLETE - DO NOT MODIFY)
  - authenticator_test.go: Full coverage
  - basic_test.go: Full coverage
  - manager_test.go: Full coverage
  - middleware_test.go: Full coverage
  - token_test.go: Full coverage
- **`pkg/registry/types/*_test.go`**: 0 lines (MISSING - THIS IS OUR TARGET)

### Implementation Files to Test
From registry-types, registry-auth, and registry-helpers efforts:
- `pkg/registry/types/registry.go` - Registry configuration types
- `pkg/registry/types/credentials.go` - Authentication configuration
- `pkg/registry/types/errors.go` - Error types and helpers
- `pkg/registry/types/options.go` - Registry options

## File Structure

```
pkg/
└── registry/
    └── types/
        ├── registry_test.go      (30 lines) - Core registry config tests
        ├── credentials_test.go   (30 lines) - Authentication config tests
        ├── errors_test.go        (25 lines) - Error handling tests
        └── options_test.go       (30 lines) - Options validation tests

Total: 115 lines (EXACTLY at budget)
```

## Implementation Steps

### Step 1: Registry Configuration Tests (`pkg/registry/types/registry_test.go`)
**Target: 30 lines**

```go
package types

import (
    "testing"
    "time"
)

func TestRegistryConfig(t *testing.T) {
    tests := []struct {
        name string
        cfg  *RegistryConfig
        want bool
    }{
        {
            name: "valid config",
            cfg: &RegistryConfig{
                URL:       "registry.example.com",
                Namespace: "myorg",
                Timeout:   30 * time.Second,
            },
            want: true,
        },
        {
            name: "insecure config",
            cfg: &RegistryConfig{
                URL:      "localhost:5000",
                Insecure: true,
            },
            want: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test basic config validation
            if (tt.cfg.URL != "") != tt.want {
                t.Errorf("config validation failed")
            }
        })
    }
}
```

### Step 2: Credentials Tests (`pkg/registry/types/credentials_test.go`)
**Target: 30 lines**

```go
package types

import "testing"

func TestAuthConfig(t *testing.T) {
    tests := []struct {
        name    string
        auth    *AuthConfig
        wantErr bool
    }{
        {
            name: "basic auth",
            auth: &AuthConfig{
                AuthType: AuthTypeBasic,
                Username: "user",
                Password: "pass",
            },
            wantErr: false,
        },
        {
            name: "token auth",
            auth: &AuthConfig{
                AuthType: AuthTypeToken,
                Token:    "bearer-token",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.auth == nil && !tt.wantErr {
                t.Error("expected non-nil auth")
            }
        })
    }
}
```

### Step 3: Error Tests (`pkg/registry/types/errors_test.go`)
**Target: 25 lines**

```go
package types

import (
    "testing"
)

func TestRegistryError(t *testing.T) {
    err := &RegistryError{
        Code:    "AUTH_FAILED",
        Message: "authentication failed",
    }

    if err.Error() != "AUTH_FAILED: authentication failed" {
        t.Errorf("Error() = %v", err.Error())
    }

    // Test error type checking
    if err.Code != "AUTH_FAILED" {
        t.Errorf("expected AUTH_FAILED")
    }
}

func TestErrorHelpers(t *testing.T) {
    // Test IsAuthError, IsNetworkError helpers
    if IsAuthError(nil) {
        t.Error("nil should not be auth error")
    }
}
```

### Step 4: Options Tests (`pkg/registry/types/options_test.go`)
**Target: 30 lines**

```go
package types

import "testing"

func TestRegistryOptions(t *testing.T) {
    tests := []struct {
        name string
        opts *RegistryOptions
        want int
    }{
        {
            name: "default options",
            opts: &RegistryOptions{
                MaxRetries:    3,
                RetryInterval: 1,
            },
            want: 3,
        },
        {
            name: "custom options",
            opts: &RegistryOptions{
                MaxRetries:    5,
                RetryInterval: 2,
            },
            want: 5,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.opts.MaxRetries != tt.want {
                t.Errorf("MaxRetries = %d, want %d", tt.opts.MaxRetries, tt.want)
            }
        })
    }
}
```

## Size Management Strategy

### Measurement Protocol
```bash
# After each file implementation
cd /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/registry-tests
PROJECT_ROOT=$(pwd)
while [ "$PROJECT_ROOT" != "/" ]; do
    [ -f "$PROJECT_ROOT/orchestrator-state.json" ] && break
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
done
$PROJECT_ROOT/tools/line-counter.sh
```

### Size Control Points
1. **After Step 1**: Should be ~30 lines
2. **After Step 2**: Should be ~60 lines
3. **After Step 3**: Should be ~85 lines
4. **After Step 4**: Should be ~115 lines (STOP HERE)

### Critical Limits
- **Warning Threshold**: 100 lines
- **Hard Stop**: 115 lines
- **Buffer**: 0 lines (we're using exact budget)

## Test Requirements

### Coverage Goals
- **Target Coverage**: 80% for types package
- **Focus Areas**:
  1. Type construction and initialization
  2. Configuration validation
  3. Error handling and type checking
  4. Options application

### Test Quality Criteria
- ✅ Table-driven tests (Go best practice)
- ✅ Independent test cases
- ✅ Clear test names
- ✅ No test stubs or TODO markers
- ✅ All tests must pass first time

## Pattern Compliance

### Go Testing Standards
- Test function names: `TestXxx` format
- Table-driven test structure
- Subtests using `t.Run()`
- Error messages include context

### Project Patterns
- Follow existing test patterns from auth package
- Use consistent assertion style
- Include both positive and negative test cases

## Implementation Priority

Given exactly 115-line budget:
1. **MUST HAVE** (85 lines):
   - registry_test.go (30 lines) - Core functionality
   - credentials_test.go (30 lines) - Auth config
   - errors_test.go (25 lines) - Error handling

2. **SHOULD HAVE** (30 lines):
   - options_test.go (30 lines) - Configuration options

## Critical Constraints

### MUST Requirements
- ✅ Total implementation EXACTLY 115 lines (no buffer)
- ✅ Focus ONLY on types package
- ✅ All tests must compile and pass
- ✅ No stub implementations or placeholders
- ✅ Follow Go testing best practices

### MUST NOT Requirements
- ❌ DO NOT modify auth package tests (already complete)
- ❌ DO NOT exceed 115 lines under any circumstances
- ❌ DO NOT create mock servers or complex fixtures
- ❌ DO NOT create integration tests
- ❌ DO NOT create test helpers or utilities

## Success Criteria

1. **Size Compliance**: Exactly 115 lines (measured by line-counter.sh)
2. **Test Coverage**: Types package has >80% coverage
3. **Quality**: All tests pass on first run
4. **Completeness**: All 4 test files created
5. **Standards**: Follows Go and project conventions

## Dependency Context

### From registry-types (E1.1.2A)
- Type definitions for RegistryConfig, RegistryInfo
- Base error types
- Constants and enums

### From registry-auth (E1.1.2B)
- AuthConfig structure
- Authentication type constants
- Token handling interfaces

### From registry-helpers (E1.1.2C)
- Helper validation functions
- Options processing
- Error helper functions

## Implementation Notes

### Critical Reminders
1. The auth package already has 2207 lines of tests - DO NOT TOUCH
2. We have EXACTLY 115 lines - no flexibility
3. Focus on high-value test cases only
4. Each test file should be self-contained
5. Run line-counter.sh after EVERY file

### Risk Mitigation
- If approaching limit early, prioritize core tests
- Skip edge cases if necessary to stay within budget
- Focus on testing public APIs only
- Avoid test duplication

## Final Checklist

Before marking complete:
- [ ] All 4 test files created
- [ ] Line count exactly 115 (per line-counter.sh)
- [ ] All tests pass (`go test ./pkg/registry/types`)
- [ ] No compilation errors
- [ ] No TODO or stub markers
- [ ] Coverage >80% for types package