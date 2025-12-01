# E1.1.1 - Credential Resolution System Implementation Plan

**Effort ID**: E1.1.1
**Effort Name**: Credential Resolution System
**Phase**: Phase 1 - Core OCI Push Implementation
**Wave**: Wave 1 - Foundation (TDD - Tests First)
**Created**: 2025-12-01T11:05:42Z
**Author**: Code Reviewer Agent
**Status**: Active

---

## EFFORT INFRASTRUCTURE METADATA

**EFFORT_ID**: E1.1.1
**EFFORT_NAME**: credential-resolution
**BRANCH**: idpbuilder-oci-push/phase-1-wave-1-effort-E1.1.1-credential-resolution
**BASE_BRANCH**: idpbuilder-oci-push/phase-1-wave-1-integration
**WORKSPACE_PATH**: /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave1/E1.1.1-credential-resolution
**TARGET_REPO**: https://github.com/jessesanford/idpbuilder.git

---

## R213 Metadata Block

```yaml
effort_id: "E1.1.1"
effort_name: "credential-resolution"
estimated_lines: 300
dependencies: []
files_touched:
  - "pkg/cmd/push/credentials.go"
  - "pkg/cmd/push/credentials_test.go"
branch_name: "idpbuilder-oci-push/phase-1-wave-1-effort-E1.1.1-credential-resolution"
base_branch: "idpbuilder-oci-push/phase-1-wave-1-integration"
can_parallelize: false
parallel_with: []
```

---

## PRIOR WORK ANALYSIS (R420 MANDATORY)

### Discovery Phase Results

- **Previous Efforts Reviewed**: None (first effort in Wave 1)
- **Previous Plans Reviewed**: None (first effort in Wave 1)
- **Research Timestamp**: 2025-12-01T11:05:42Z
- **Research Status**: COMPLETE

### File Structure Findings

| File Path | Source Effort | Status | Action Required |
|-----------|---------------|--------|-----------------|
| pkg/cmd/push/credentials.go | E1.1.1 (this effort) | NEW | MUST create |
| pkg/cmd/push/credentials_test.go | E1.1.1 (this effort) | NEW | MUST create |

### Interface/API Findings

| Interface/API | Source | Signature | Action Required |
|---------------|--------|-----------|-----------------|
| CredentialResolver | E1.1.1 | `Resolve(flags CredentialFlags, env EnvironmentLookup) (*Credentials, error)` | MUST implement |
| EnvironmentLookup | E1.1.1 | `Get(key string) string` | MUST implement |

### Type/Struct Findings

| Type | Source | Exported | Action Required |
|------|--------|----------|-----------------|
| Credentials | E1.1.1 | YES | MUST create |
| CredentialFlags | E1.1.1 | YES | MUST create |
| DefaultEnvironment | E1.1.1 | YES | MUST create |
| DefaultCredentialResolver | E1.1.1 | YES | MUST create |

### Method Visibility Findings

| Method | Type | Visibility | Can Access? | Action Required |
|--------|------|------------|-------------|-----------------|
| Resolve | DefaultCredentialResolver | EXPORTED | YES | Primary method to implement |
| Get | DefaultEnvironment | EXPORTED | YES | Environment lookup method |

### Conflicts Detected

- NO duplicate file paths detected (first effort)
- NO API mismatches detected (first effort)
- NO method visibility violations detected (first effort)

### Required Integrations

1. MUST use `github.com/stretchr/testify` for testing (already in idpbuilder go.mod)
2. MUST follow existing idpbuilder code patterns

### Forbidden Actions

- DO NOT create files outside `pkg/cmd/push/` directory
- DO NOT modify existing idpbuilder files
- DO NOT add new dependencies to go.mod (testify already exists)

---

## 1. Scope and Objectives

### 1.1 Primary Objective

Implement a complete credential resolution subsystem that resolves authentication credentials for OCI registry operations. This is a foundational component that other Wave 1 and Wave 2 efforts will depend on for authentication.

### 1.2 Scope Boundaries

**IN SCOPE:**
- `Credentials` struct with Username, Password, Token, IsAnonymous fields
- `CredentialFlags` struct for CLI flag values
- `EnvironmentLookup` interface for testable environment access
- `DefaultEnvironment` implementation using `os.Getenv`
- `CredentialResolver` interface for credential resolution
- `DefaultCredentialResolver` implementation with flag > env precedence
- Environment variable constants for registry credentials
- Comprehensive table-driven tests following idpbuilder patterns
- MockEnvironment for test isolation

**OUT OF SCOPE:**
- CLI command implementation (Wave 3)
- Registry client usage (E1.1.2, E1.2.1)
- Docker daemon interaction (E1.1.3, E1.2.2)
- Progress reporting (E1.3.2)

### 1.3 Success Criteria

1. All interfaces defined and implemented
2. All 7 test cases pass (table-driven tests)
3. Go race detector finds no issues
4. Code compiles with `go build ./pkg/cmd/push/...`
5. Total implementation lines <= 300 (well under 800 limit)
6. Code follows existing idpbuilder patterns (testify/mock style)

---

## 2. Technical Approach

### 2.1 Architecture Design

The credential resolution system follows a simple, testable design:

```
                    CLI Application
                          |
                          v
                   CredentialFlags
                   (from cobra flags)
                          |
                          v
              +---------------------+
              | CredentialResolver  |<---- Interface
              +---------------------+
                          |
                          v
        +-----------------------------+
        | DefaultCredentialResolver   |<---- Implementation
        +-----------------------------+
                    |           |
                    v           v
            CredentialFlags  EnvironmentLookup
                              (interface)
                                  |
                    +-------------+-------------+
                    |                           |
          DefaultEnvironment           MockEnvironment
          (os.Getenv)                  (for testing)
                    |
                    v
              *Credentials
              (resolved auth)
```

### 2.2 Resolution Priority (REQ-014)

The credential resolver MUST follow this priority order:
1. **CLI Flags** (highest priority) - Always override environment variables
2. **Environment Variables** - Used when flags are not provided
3. **Anonymous Access** - When neither flags nor env vars are set

### 2.3 Authentication Modes

The system supports two mutually exclusive authentication modes:
1. **Basic Auth**: Username + Password combination
2. **Token Auth**: Bearer token authentication

**Conflict Handling**: If both token AND username/password are provided, return an error.

---

## 3. TDD Test Plan (R400/R401 Compliance - Tests FIRST)

### 3.1 Test Cases (Table-Driven)

The following test cases MUST be implemented in `credentials_test.go`:

```go
// TestCredentialResolver_FlagPrecedence - 7 test cases
tests := []struct {
    name           string
    flags          CredentialFlags
    envUsername    string
    envPassword    string
    envToken       string
    wantUsername   string
    wantPassword   string
    wantToken      string
    wantAnonymous  bool
    wantErr        bool
}{
    // Test Case 1: Flag overrides environment for username/password
    {
        name: "flag_overrides_env_username",
        flags: CredentialFlags{Username: "flag-user", Password: "flag-pass"},
        envUsername:  "env-user",
        envPassword:  "env-pass",
        wantUsername: "flag-user",
        wantPassword: "flag-pass",
    },

    // Test Case 2: Environment used when no flags provided
    {
        name:         "env_used_when_no_flags",
        flags:        CredentialFlags{},
        envUsername:  "env-user",
        envPassword:  "env-pass",
        wantUsername: "env-user",
        wantPassword: "env-pass",
    },

    // Test Case 3: Token flag overrides token env
    {
        name:      "token_flag_overrides_token_env",
        flags:     CredentialFlags{Token: "flag-token"},
        envToken:  "env-token",
        wantToken: "flag-token",
    },

    // Test Case 4: Token env used when no token flag
    {
        name:      "token_env_used_when_no_token_flag",
        flags:     CredentialFlags{},
        envToken:  "env-token",
        wantToken: "env-token",
    },

    // Test Case 5: Anonymous access when no credentials
    {
        name:          "anonymous_when_no_credentials",
        flags:         CredentialFlags{},
        wantAnonymous: true,
    },

    // Test Case 6: Error when both token and basic auth
    {
        name:    "error_when_both_token_and_basic_auth",
        flags:   CredentialFlags{Username: "user", Token: "token"},
        wantErr: true,
    },

    // Test Case 7: Partial flag override (flag username, env password)
    {
        name:         "partial_flag_override",
        flags:        CredentialFlags{Username: "flag-user"},
        envUsername:  "env-user",
        envPassword:  "env-pass",
        wantUsername: "flag-user",
        wantPassword: "env-pass",  // From environment
    },
}
```

### 3.2 Security Test (P1.3 Property Verification)

```go
// TestCredentialResolver_NoCredentialLogging
// Verifies that Credentials struct does NOT have a String() method
// that could accidentally expose secrets in logs
func TestCredentialResolver_NoCredentialLogging(t *testing.T) {
    creds := &Credentials{
        Username: "secret-user",
        Password: "secret-pass",
        Token:    "secret-token",
    }

    // Verify struct exists with expected fields
    assert.NotEmpty(t, creds.Username)
    assert.NotEmpty(t, creds.Password)
    assert.NotEmpty(t, creds.Token)

    // Note: Credentials struct intentionally has NO String() method
    // to prevent accidental credential logging
}
```

### 3.3 Mock Environment Implementation

```go
// MockEnvironment implements EnvironmentLookup for testing
type MockEnvironment struct {
    mock.Mock
}

func (m *MockEnvironment) Get(key string) string {
    args := m.Called(key)
    return args.String(0)
}
```

---

## 4. Implementation Steps

### Phase 1: Create Directory Structure (5 min)

```bash
# Step 1.1: Create pkg/cmd/push directory
mkdir -p pkg/cmd/push

# Step 1.2: Verify directory exists
ls -la pkg/cmd/push/
```

### Phase 2: Create credentials.go (~135 lines)

#### Step 2.1: Package Declaration and Imports (lines 1-10)

```go
// pkg/cmd/push/credentials.go
package push

import (
    "fmt"
    "os"
)
```

#### Step 2.2: Credentials Struct (lines 12-25)

```go
// Credentials holds resolved authentication credentials for registry operations.
// Either Username/Password pair OR Token is used, never both.
type Credentials struct {
    // Username for basic authentication
    Username string
    // Password for basic authentication
    Password string
    // Token for bearer token authentication (takes precedence over basic auth)
    Token string
    // IsAnonymous indicates no credentials were provided
    IsAnonymous bool
}
```

#### Step 2.3: CredentialFlags Struct (lines 27-38)

```go
// CredentialFlags contains CLI flag values for credential resolution.
// These values take precedence over environment variables per REQ-014.
type CredentialFlags struct {
    Username string
    Password string
    Token    string
}
```

#### Step 2.4: EnvironmentLookup Interface (lines 40-50)

```go
// EnvironmentLookup abstracts environment variable access for testing.
// This allows tests to inject mock environment values without modifying os.Environ.
type EnvironmentLookup interface {
    // Get retrieves the value of an environment variable.
    // Returns empty string if not set.
    Get(key string) string
}
```

#### Step 2.5: CredentialResolver Interface (lines 52-62)

```go
// CredentialResolver resolves authentication credentials from multiple sources.
// Resolution priority: CLI flags > environment variables > anonymous access.
type CredentialResolver interface {
    // Resolve determines credentials based on flags and environment.
    // Returns Credentials with IsAnonymous=true if no credentials found.
    Resolve(flags CredentialFlags, env EnvironmentLookup) (*Credentials, error)
}
```

#### Step 2.6: Environment Variable Constants (lines 64-70)

```go
// Environment variable names for credential resolution
const (
    EnvRegistryUsername = "IDPBUILDER_REGISTRY_USERNAME"
    EnvRegistryPassword = "IDPBUILDER_REGISTRY_PASSWORD"
    EnvRegistryToken    = "IDPBUILDER_REGISTRY_TOKEN"
)
```

#### Step 2.7: DefaultEnvironment Implementation (lines 72-82)

```go
// DefaultEnvironment implements EnvironmentLookup using os.Getenv.
type DefaultEnvironment struct{}

// Get implements EnvironmentLookup.Get using os.Getenv.
func (e *DefaultEnvironment) Get(key string) string {
    return os.Getenv(key)
}
```

#### Step 2.8: DefaultCredentialResolver Implementation (lines 84-135)

```go
// DefaultCredentialResolver implements CredentialResolver.
// Priority order: flags > environment > anonymous
type DefaultCredentialResolver struct{}

// Resolve implements CredentialResolver.Resolve.
// Validates that either basic auth (username+password) or token is provided, not both.
func (r *DefaultCredentialResolver) Resolve(flags CredentialFlags, env EnvironmentLookup) (*Credentials, error) {
    creds := &Credentials{}

    // Token resolution: flag takes precedence over environment (REQ-014)
    creds.Token = flags.Token
    if creds.Token == "" {
        creds.Token = env.Get(EnvRegistryToken)
    }

    // Username resolution: flag takes precedence over environment (REQ-014)
    creds.Username = flags.Username
    if creds.Username == "" {
        creds.Username = env.Get(EnvRegistryUsername)
    }

    // Password resolution: flag takes precedence over environment (REQ-014)
    creds.Password = flags.Password
    if creds.Password == "" {
        creds.Password = env.Get(EnvRegistryPassword)
    }

    // Determine auth mode
    hasToken := creds.Token != ""
    hasBasic := creds.Username != "" || creds.Password != ""

    // Validate: cannot have both token and basic auth
    if hasToken && hasBasic {
        return nil, fmt.Errorf("cannot specify both token and username/password credentials")
    }

    // If token is provided, clear basic auth fields for consistency
    if hasToken {
        creds.Username = ""
        creds.Password = ""
    }

    // If no credentials at all, mark as anonymous
    creds.IsAnonymous = !hasToken && !hasBasic

    return creds, nil
}
```

### Phase 3: Create credentials_test.go (~165 lines)

#### Step 3.1: Package and Imports (lines 1-15)

```go
// pkg/cmd/push/credentials_test.go
package push

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)
```

#### Step 3.2: MockEnvironment (lines 17-28)

```go
// MockEnvironment implements EnvironmentLookup for testing.
type MockEnvironment struct {
    mock.Mock
}

// Get implements EnvironmentLookup.Get for mocking.
func (m *MockEnvironment) Get(key string) string {
    args := m.Called(key)
    return args.String(0)
}
```

#### Step 3.3: TestCredentialResolver_FlagPrecedence (lines 30-140)

Complete table-driven test implementation as specified in Section 3.1.

#### Step 3.4: TestCredentialResolver_NoCredentialLogging (lines 142-165)

Complete security test as specified in Section 3.2.

### Phase 4: Verification and Testing

```bash
# Step 4.1: Compile the package
go build ./pkg/cmd/push/...

# Step 4.2: Run tests
go test ./pkg/cmd/push/... -v

# Step 4.3: Check for race conditions
go test -race ./pkg/cmd/push/...

# Step 4.4: Verify line count
wc -l pkg/cmd/push/*.go
```

---

## 5. Demo Requirements (R330/R291)

### 5.1 Demo Script

```bash
#!/bin/bash
# demo-e1.1.1.sh - E1.1.1 Credential Resolution Demo

echo "=========================================="
echo "E1.1.1: Credential Resolution Demo"
echo "=========================================="

# Step 1: Verify files exist
echo ""
echo "[1/4] Verifying file structure..."
[ -f "pkg/cmd/push/credentials.go" ] && echo "  credentials.go: OK" || echo "  credentials.go: MISSING"
[ -f "pkg/cmd/push/credentials_test.go" ] && echo "  credentials_test.go: OK" || echo "  credentials_test.go: MISSING"

# Step 2: Verify interfaces exist
echo ""
echo "[2/4] Verifying interface definitions..."
grep -q "type CredentialResolver interface" pkg/cmd/push/credentials.go && \
  echo "  CredentialResolver interface: OK" || echo "  CredentialResolver interface: MISSING"
grep -q "type EnvironmentLookup interface" pkg/cmd/push/credentials.go && \
  echo "  EnvironmentLookup interface: OK" || echo "  EnvironmentLookup interface: MISSING"

# Step 3: Compile package
echo ""
echo "[3/4] Compiling package..."
go build ./pkg/cmd/push/... && echo "  Build: OK" || echo "  Build: FAILED"

# Step 4: Run tests
echo ""
echo "[4/4] Running tests..."
go test ./pkg/cmd/push/... -v 2>&1 | grep -E "^(---|\s+(PASS|FAIL)|ok|FAIL)"

echo ""
echo "=========================================="
echo "E1.1.1 Demo Complete"
echo "=========================================="
```

### 5.2 Success Criteria

1. `credentials.go` exists and contains CredentialResolver interface
2. `credentials_test.go` exists with all 7 test cases + security test
3. `go build ./pkg/cmd/push/...` succeeds
4. `go test ./pkg/cmd/push/...` shows all tests passing
5. `go test -race ./pkg/cmd/push/...` shows no race conditions

---

## 6. Acceptance Criteria Checklist

### 6.1 Functional Requirements

- [ ] `Credentials` struct defined with Username, Password, Token, IsAnonymous fields
- [ ] `CredentialFlags` struct defined for CLI flag values
- [ ] `EnvironmentLookup` interface with `Get(key string) string` method
- [ ] `DefaultEnvironment` implements `EnvironmentLookup` using `os.Getenv`
- [ ] `CredentialResolver` interface with `Resolve` method signature
- [ ] `DefaultCredentialResolver` implements resolution with flag > env precedence
- [ ] Environment variable constants defined (IDPBUILDER_REGISTRY_*)

### 6.2 Test Requirements

- [ ] All 7 test cases in `TestCredentialResolver_FlagPrecedence` pass
- [ ] Token and basic auth conflict returns error
- [ ] Anonymous credential returned when no credentials provided
- [ ] `go test ./pkg/cmd/push/...` passes with 0 failures
- [ ] No race conditions (`go test -race`)

### 6.3 Quality Requirements

- [ ] Code compiles without warnings
- [ ] Follows existing idpbuilder patterns (testify/mock style)
- [ ] Line count verified under 300 lines (well under 800 limit)
- [ ] All code committed and pushed

---

## 7. Size Estimate

| File | Estimated Lines |
|------|-----------------|
| pkg/cmd/push/credentials.go | ~135 lines |
| pkg/cmd/push/credentials_test.go | ~165 lines |
| **Total** | **~300 lines** |

**Size Compliance**: PASS (300 lines << 800 line limit)

**Measurement Tool**: Use `$CLAUDE_PROJECT_DIR/tools/line-counter.sh` after implementation

---

## 8. Risk Assessment

### 8.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| testify not in go.mod | LOW | Medium | Verify testify already exists before implementation |
| Package conflicts | LOW | Low | Using new pkg/cmd/push directory |
| Pattern mismatch | LOW | Medium | Follow existing idpbuilder test patterns |

### 8.2 Size Risks

| Metric | Value | Limit | Risk Level |
|--------|-------|-------|------------|
| Estimated Lines | 300 | 800 | LOW |
| Implementation Files | 2 | - | LOW |

---

## 9. Dependencies

### 9.1 Internal Dependencies

- None (first effort in Wave 1)

### 9.2 External Dependencies

| Dependency | Version | Status |
|------------|---------|--------|
| github.com/stretchr/testify | (existing) | Already in idpbuilder go.mod |

---

## 10. Property Coverage Matrix

| Property ID | Source | Covered By | Test File | Status |
|-------------|--------|------------|-----------|--------|
| P1.1 | Phase Architecture | TestCredentialResolver_FlagPrecedence | credentials_test.go | Planned |
| P1.3 | Phase Architecture | TestCredentialResolver_NoCredentialLogging | credentials_test.go | Planned |

---

## 11. References

- **Wave Implementation Plan**: planning/phase1/wave1/WAVE-1-IMPLEMENTATION-PLAN.md
- **Wave Architecture Plan**: planning/phase1/wave1/WAVE-1-ARCHITECTURE-PLAN.md
- **Wave Test Plan**: planning/phase1/wave1/WAVE-1-TEST-PLAN.md
- **Phase Architecture Plan**: planning/phase1/PHASE-1-ARCHITECTURE-PLAN.md
- **Target Repository**: https://github.com/jessesanford/idpbuilder.git

---

## Approvals

| Stakeholder | Role | Status | Date |
|-------------|------|--------|------|
| Code Reviewer Agent | Effort Planning Authority | Approved | 2025-12-01 |
| SW Engineer | Implementation | Pending | - |

---

**CHECKLIST[1]: R420 research complete - 0 previous efforts analyzed (first effort in wave) [2025-12-01T11:05:42Z]**
