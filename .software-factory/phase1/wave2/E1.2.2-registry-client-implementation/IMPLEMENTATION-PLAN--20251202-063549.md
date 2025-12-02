# E1.2.2 Implementation Plan - Registry Client Implementation

**Phase**: Phase 1 - Core OCI Push Implementation
**Wave**: Wave 2 - Core Implementation
**Effort**: E1.2.2 - Registry Client Implementation
**Created**: 2025-12-02T06:35:49Z
**Author**: Code Reviewer Agent
**Fidelity Level**: **EXACT** (detailed effort specification with R213 metadata)
**R383 Compliance**: Timestamp included in filename

---

## EFFORT INFRASTRUCTURE METADATA (ORCHESTRATOR DEFINED)

**EFFORT_NAME**: E1.2.2-registry-client-implementation
**EFFORT_ID**: E1.2.2
**PHASE_NUMBER**: 1
**WAVE_NUMBER**: 2
**BRANCH**: `idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.2-registry-client-implementation`
**BASE_BRANCH**: `idpbuilder-oci-push/phase-1-wave-2-integration`
**WORKING_DIRECTORY**: `/home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.2-registry-client-implementation`
**CAN_PARALLELIZE**: true
**PARALLEL_WITH**: E1.2.3

---

## R213 Effort Metadata

```yaml
effort_id: "E1.2.2"
effort_name: "registry-client-implementation"
estimated_lines: 380
dependencies: ["Wave 1 Integration (RegistryClient interface, error types, ProgressReporter)"]
files_touched:
  - "pkg/registry/registry.go"
  - "pkg/registry/registry_test.go"
  - "pkg/registry/progress.go" (new file for StderrProgressReporter implementation)
  - "go.mod" (modification - add go-containerregistry)
branch_name: "idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.2-registry-client-implementation"
can_parallelize: true
parallel_with: ["E1.2.3"]
base_branch: "idpbuilder-oci-push/phase-1-wave-2-integration"
```

---

## Pre-Planning Research Results (R374 MANDATORY)

### Existing Interfaces Found (MUST IMPLEMENT)

| Interface | Location | Signature | Must Implement |
|-----------|----------|-----------|----------------|
| `RegistryClient` | `pkg/registry/client.go:36` | `Push(ctx context.Context, imageRef, destRef string, progress ProgressReporter) (*PushResult, error)` | YES - Main implementation target |
| `ProgressReporter` | `pkg/registry/client.go:99` | `Start(), LayerProgress(), LayerComplete(), Complete(), Error()` | YES - StderrProgressReporter |
| `RegistryClientFactory` | `pkg/registry/client.go:49` | `NewClient(config RegistryConfig) (RegistryClient, error)` | NO - Factory pattern not needed for this effort |

### Existing Types to REUSE (R373)

| Type | Location | Purpose | How to Use |
|------|----------|---------|------------|
| `RegistryConfig` | `pkg/registry/client.go:21-32` | Configuration struct | Accept in NewDefaultClient() |
| `PushResult` | `pkg/registry/client.go:10-18` | Return type for Push | Return from Push() method |
| `RegistryError` | `pkg/registry/client.go:56-78` | Error with classification | Create in classifyRemoteError() |
| `AuthError` | `pkg/registry/client.go:80-97` | Auth failure error | Create for 401/403 errors |
| `NoOpProgressReporter` | `pkg/registry/client.go:114-122` | No-op progress | Use for nil progress handling |
| `StderrProgressReporter` | `pkg/registry/client.go:124-148` | Progress to stderr | IMPLEMENT (currently stubbed) |

### FORBIDDEN DUPLICATIONS (R373)

- DO NOT create new RegistryClient interface (exists at pkg/registry/client.go:36)
- DO NOT create new RegistryConfig struct (exists at pkg/registry/client.go:21)
- DO NOT create new PushResult struct (exists at pkg/registry/client.go:10)
- DO NOT create new error types (RegistryError, AuthError exist)
- DO NOT create new ProgressReporter interface (exists at pkg/registry/client.go:99)

### REQUIRED IMPLEMENTATIONS (R373)

- MUST implement `RegistryClient` interface from `pkg/registry/client.go` with EXACT signature
- MUST implement `StderrProgressReporter` methods (currently stubbed with "Wave 3" comments)
- MUST use `RegistryConfig` for configuration
- MUST return `*PushResult` from Push method
- MUST return `*RegistryError` or `*AuthError` for error cases

---

## 1. Effort Overview

### 1.1 Purpose

This effort implements the `DefaultClient` struct that satisfies the `RegistryClient` interface defined in Wave 1. It uses `go-containerregistry` library for:
- Reading images from the local Docker daemon
- Pushing images to OCI-compliant registries
- Handling authentication (basic auth, token auth, anonymous)
- Classifying errors (transient vs permanent, auth errors)

### 1.2 Dependencies

| Dependency | Source | Used For |
|------------|--------|----------|
| `RegistryClient` interface | Wave 1 pkg/registry/client.go | Interface to implement |
| `RegistryConfig` | Wave 1 pkg/registry/client.go | Configuration struct |
| `PushResult` | Wave 1 pkg/registry/client.go | Return type |
| `RegistryError`, `AuthError` | Wave 1 pkg/registry/client.go | Error types |
| `ProgressReporter` interface | Wave 1 pkg/registry/client.go | Progress callbacks |
| `go-containerregistry` | External (NEW) | OCI operations library |

### 1.3 Execution Order

This effort **CAN run in parallel** with E1.2.3 because:
- Implements independent interface
- No code dependencies on E1.2.3
- Only compile-time dependency on Wave 1 interfaces

---

## EXPLICIT SCOPE (R311 MANDATORY)

### IMPLEMENT EXACTLY:

**File: pkg/registry/registry.go (~200 lines)**
- Type: `DefaultClient` struct with 3 fields (~15 lines)
- Function: `NewDefaultClient(config RegistryConfig) (*DefaultClient, error)` (~30 lines)
- Method: `(c *DefaultClient) Push(ctx, imageRef, destRef, progress) (*PushResult, error)` (~80 lines)
- Function: `classifyRemoteError(err error) error` (~35 lines)
- Function: `containsInsensitive(s, substr string) bool` (~10 lines)
- Helper constants and imports (~30 lines)

**File: pkg/registry/progress.go (~30 lines)**
- Implement `StderrProgressReporter.Start()` (~8 lines)
- Implement `StderrProgressReporter.LayerProgress()` (~8 lines)
- Implement `StderrProgressReporter.LayerComplete()` (~5 lines)
- Implement `StderrProgressReporter.Complete()` (~5 lines)
- Implement `StderrProgressReporter.Error()` (~4 lines)

**File: pkg/registry/registry_test.go (~150 lines)**
- Test: `TestDefaultClient_Push_BasicAuth` (~25 lines)
- Test: `TestDefaultClient_Push_TokenAuth` (~20 lines)
- Test: `TestDefaultClient_Push_Anonymous` (~15 lines)
- Test: `TestDefaultClient_Push_ProgressCallbacks` (~25 lines)
- Test: `TestDefaultClient_Push_AuthError_Classification` (~30 lines)
- Test: `TestDefaultClient_Push_TransientError_Classification` (~20 lines)
- Test: `TestDefaultClient_Push_InsecureMode` (~15 lines)

**TOTAL IMPLEMENTATION: ~380 lines**

### DO NOT IMPLEMENT (Deferred/Out of Scope):

- DO NOT create `RegistryClientFactory` implementation (not needed per architecture)
- DO NOT implement retry logic (future wave feature)
- DO NOT implement layer-by-layer progress (basic progress only in Wave 2)
- DO NOT implement parallel layer uploads (single-threaded push)
- DO NOT implement image manifest listing (read-only operations future effort)
- DO NOT implement image deletion (future effort)
- DO NOT add caching layer (future wave)
- DO NOT implement fancy progress bars with spinners (Wave 3 E1.3.2)

---

## 2. Files to Create/Modify

### 2.1 pkg/registry/registry.go (NEW - ~200 lines)

**Purpose**: DefaultClient implementation using go-containerregistry

**Required Imports**:
```go
package registry

import (
    "context"
    "crypto/tls"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/google/go-containerregistry/pkg/authn"
    "github.com/google/go-containerregistry/pkg/name"
    "github.com/google/go-containerregistry/pkg/v1/daemon"
    "github.com/google/go-containerregistry/pkg/v1/remote"
)
```

**Struct Definition**:
```go
// DefaultClient implements RegistryClient using go-containerregistry
type DefaultClient struct {
    config     RegistryConfig
    auth       authn.Authenticator
    httpClient *http.Client
}
```

**Constructor Function**:
```go
// NewDefaultClient creates a new registry client with the given configuration
func NewDefaultClient(config RegistryConfig) (*DefaultClient, error) {
    // 1. Create client instance
    // 2. Configure authentication based on config:
    //    - If config.Token != "" -> use authn.Bearer
    //    - Else if config.Username != "" && config.Password != "" -> use authn.Basic
    //    - Else -> use authn.Anonymous
    // 3. Configure http.Client with optional InsecureSkipVerify
    // 4. Return configured client
}
```

**Push Method** (implements RegistryClient interface):
```go
// Push implements RegistryClient.Push using go-containerregistry
func (c *DefaultClient) Push(ctx context.Context, imageRef, destRef string, progress ProgressReporter) (*PushResult, error) {
    // 1. Parse source reference for daemon access
    // 2. Parse destination reference for registry
    // 3. Get image from local daemon using daemon.Image()
    // 4. Get layer count for progress reporting
    // 5. Call progress.Start() if progress != nil
    // 6. Build remote.Options (WithAuth, WithContext, WithTransport if insecure)
    // 7. Call remote.Write() to push image
    // 8. Handle errors with classifyRemoteError()
    // 9. Get digest from pushed image
    // 10. Calculate total size from layers
    // 11. Build and return PushResult
    // 12. Call progress.Complete() if progress != nil
}
```

**Error Classification Function**:
```go
// classifyRemoteError converts go-containerregistry errors to our error types
func classifyRemoteError(err error) error {
    // 1. Get error string (lowercase for comparison)
    // 2. Check for auth errors (401, 403, unauthorized, forbidden) -> AuthError
    // 3. Check for transient errors (5xx, timeout, connection refused) -> RegistryError{IsTransient: true}
    // 4. Default to permanent error -> RegistryError{IsTransient: false}
}
```

### 2.2 pkg/registry/progress.go (NEW - ~30 lines)

**Purpose**: Implement StderrProgressReporter (basic version for Wave 2)

**Note**: The interface and struct already exist in client.go. This file implements the methods.

```go
package registry

import (
    "fmt"
    "io"
)

// StderrProgressReporter methods (basic Wave 2 implementation)
// These replace the stubbed methods in client.go

func (s *StderrProgressReporter) Start(imageRef string, totalLayers int) {
    if s.Out != nil {
        fmt.Fprintf(s.Out, "Pushing %s (%d layers)...\n", imageRef, totalLayers)
    }
}

func (s *StderrProgressReporter) LayerProgress(layerDigest string, current, total int64) {
    // Basic implementation - log at 25% milestones only
    if s.Out != nil && total > 0 {
        percent := (current * 100) / total
        if percent%25 == 0 && percent > 0 {
            shortDigest := layerDigest
            if len(shortDigest) > 12 {
                shortDigest = shortDigest[:12]
            }
            fmt.Fprintf(s.Out, "  %s: %d%%\n", shortDigest, percent)
        }
    }
}

func (s *StderrProgressReporter) LayerComplete(layerDigest string) {
    if s.Out != nil {
        shortDigest := layerDigest
        if len(shortDigest) > 12 {
            shortDigest = shortDigest[:12]
        }
        fmt.Fprintf(s.Out, "  %s: done\n", shortDigest)
    }
}

func (s *StderrProgressReporter) Complete(result *PushResult) {
    if s.Out != nil {
        fmt.Fprintf(s.Out, "Push complete: %s\n", result.Digest)
    }
}

func (s *StderrProgressReporter) Error(err error) {
    if s.Out != nil {
        fmt.Fprintf(s.Out, "Push failed: %v\n", err)
    }
}
```

### 2.3 pkg/registry/registry_test.go (NEW - ~150 lines)

**Purpose**: Implementation tests for DefaultClient

**Test Functions** (matching Wave 2 Test Plan):
- `TestDefaultClient_Push_BasicAuth` (W2-RC-001)
- `TestDefaultClient_Push_TokenAuth` (W2-RC-002)
- `TestDefaultClient_Push_Anonymous` (W2-RC-003)
- `TestDefaultClient_Push_ProgressCallbacks` (W2-RC-004)
- `TestDefaultClient_Push_AuthError_Classification` (W2-RC-005)
- `TestDefaultClient_Push_TransientError_Classification` (W2-RC-006)
- `TestDefaultClient_Push_InsecureMode` (W2-RC-007)

### 2.4 go.mod (MODIFICATION)

**Purpose**: Add go-containerregistry dependency

**Command**:
```bash
go get github.com/google/go-containerregistry@latest
go mod tidy
```

This adds:
```
require github.com/google/go-containerregistry v0.x.x
```

---

## R355 PRODUCTION READINESS - ZERO TOLERANCE

This implementation MUST be production-ready from the first commit:
- NO STUBS or placeholder implementations
- NO MOCKS except in test files
- NO hardcoded credentials or secrets
- NO static configuration values
- NO TODO/FIXME markers in code
- NO returning nil or empty for "later implementation"
- NO panic("not implemented") patterns
- NO fake or dummy data

VIOLATION = -100% AUTOMATIC FAILURE

### Configuration Requirements (R355 Mandatory)

**WRONG - Will fail review:**
```go
// VIOLATION - Hardcoded timeout
timeout := 30 * time.Second

// VIOLATION - Static URL
registryURL := "https://localhost:5000"
```

**CORRECT - Production ready:**
```go
// From configuration struct
timeout := 60 * time.Second  // Reasonable default

// From config parameter
registryURL := config.URL
if registryURL == "" {
    return nil, errors.New("registry URL is required")
}
```

---

## 3. Test Cases to Satisfy (from Wave 2 Test Plan)

### 3.1 Registry Client Tests (W2-RC-*)

| Test ID | Description | Priority | Status |
|---------|-------------|----------|--------|
| W2-RC-001 | Push with username/password | Critical | Must Pass |
| W2-RC-002 | Push with bearer token | Critical | Must Pass |
| W2-RC-003 | Push without credentials (anonymous) | High | Must Pass |
| W2-RC-004 | Progress reporter invoked correctly | Critical | Must Pass |
| W2-RC-005 | 401/403 mapped to AuthError | Critical | Must Pass |
| W2-RC-006 | 5xx/network mapped to IsTransient | Critical | Must Pass |
| W2-RC-007 | TLS skip with --insecure | Medium | Must Pass |

---

## 4. Implementation Steps

### Step 1: Add go-containerregistry dependency (~5 min)
```bash
cd /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.2-registry-client-implementation
go get github.com/google/go-containerregistry@latest
go mod tidy
```

### Step 2: Create registry.go (~60 min)
1. Create `pkg/registry/registry.go`
2. Define `DefaultClient` struct
3. Implement `NewDefaultClient()` with auth config
4. Implement `Push()` using `remote.Write()`
5. Implement `classifyRemoteError()` for error classification

### Step 3: Create progress.go (~15 min)
1. Create `pkg/registry/progress.go`
2. Implement `StderrProgressReporter` methods
3. Add basic percentage logging at milestones

### Step 4: Create registry_test.go (~90 min)
1. Create `pkg/registry/registry_test.go`
2. Implement all W2-RC-* test functions
3. Use `httptest` for mock registry server
4. Test error classification logic

### Step 5: Run tests and verify (~20 min)
1. `go test ./pkg/registry/... -v`
2. Verify all tests pass
3. Check coverage >= 80%

---

## 5. Line Count Estimate

| File | Lines | Notes |
|------|-------|-------|
| registry.go | ~200 | DefaultClient implementation |
| progress.go | ~30 | StderrProgressReporter basic impl |
| registry_test.go | ~150 | 7 test functions + helpers |
| **Total** | **~380** | Within 800 line limit |

**R220 Compliance**: 380 lines is well under the 800-line hard limit.

### Size Limit Clarification (R359):
- The 800-line limit applies to NEW CODE YOU ADD
- Repository will grow by ~380 lines (EXPECTED)
- NEVER delete existing code to meet size limits
- Current codebase: pkg/registry has ~149 lines (client.go)
- Expected total after: ~529 lines in pkg/registry/

---

## 6. External Library Usage (R381 - Version Consistency)

### go-containerregistry Usage

**Library**: `github.com/google/go-containerregistry`
**Version**: Use latest stable (will be pinned after `go get`)

```go
// Import paths
import (
    "github.com/google/go-containerregistry/pkg/authn"       // Authentication
    "github.com/google/go-containerregistry/pkg/name"        // Reference parsing
    "github.com/google/go-containerregistry/pkg/v1/daemon"   // Local daemon access
    "github.com/google/go-containerregistry/pkg/v1/remote"   // Remote registry ops
)

// Authentication patterns
authn.Anonymous                           // No auth
&authn.Basic{Username: "u", Password: "p"} // Basic auth
&authn.Bearer{Token: "t"}                   // Token auth

// Reference parsing
ref, err := name.ParseReference("image:tag", name.WeakValidation)

// Read from daemon
img, err := daemon.Image(ref)

// Push to registry
err := remote.Write(destRef, img, remote.WithAuth(auth), remote.WithContext(ctx))
```

---

## 7. Demo Requirements (R330 MANDATORY)

### Demo Objectives (3-5 specific, verifiable objectives)
- [ ] Demonstrate DefaultClient can push to a local registry with basic auth
- [ ] Show proper error handling for 401 authentication failures
- [ ] Verify progress reporter receives Start/Complete callbacks
- [ ] Prove push with insecure mode works against HTTP registry
- [ ] Display proper error classification for 5xx errors

**Success Criteria**: All objectives checked = demo passes

### Demo Scenarios (IMPLEMENT EXACTLY THESE)

#### Scenario 1: Successful Push with Basic Auth
- **Setup**: Local registry running (docker run -p 5000:5000 registry:2)
- **Input**: Image `alpine:latest`, registry `localhost:5000`, basic auth credentials
- **Action**:
  ```go
  config := RegistryConfig{URL: "localhost:5000", Username: "user", Password: "pass", Insecure: true}
  client, _ := NewDefaultClient(config)
  result, err := client.Push(ctx, "alpine:latest", "localhost:5000/alpine:demo", progress)
  ```
- **Expected Output**:
  ```
  Pushing alpine:latest (1 layers)...
    sha256:abc123...: done
  Push complete: sha256:...
  ```
- **Verification**: `result.Digest` is non-empty, `result.Reference` contains digest
- **Script Lines**: ~20 lines

#### Scenario 2: Auth Failure Handling
- **Setup**: Local registry with authentication required
- **Input**: Wrong credentials
- **Action**: Attempt push with invalid username/password
- **Expected Output**:
  ```
  Push failed: authentication failed: UNAUTHORIZED
  ```
- **Verification**: Error is `*AuthError` type
- **Script Lines**: ~15 lines

#### Scenario 3: Error Classification (5xx)
- **Setup**: Mock registry returning 503 Service Unavailable
- **Input**: Any valid image reference
- **Action**: Attempt push to mock server returning 503
- **Expected Output**: Error is `*RegistryError` with `IsTransient: true`
- **Verification**: `errors.As(err, &registryErr) && registryErr.IsTransient`
- **Script Lines**: ~20 lines

**TOTAL SCENARIO LINES**: ~55 lines

### Demo Size Planning

#### Demo Artifacts (Excluded from line count per R007)
```
demo-registry-client.sh:     30 lines  # Executable script
DEMO.md:                     25 lines  # Documentation
integration-hook.sh:         10 lines  # For wave integration
-----------------------------------
TOTAL DEMO FILES:            65 lines (NOT counted toward 800)
```

#### Effort Size Summary
```
Implementation:     380 lines  # <- ONLY this counts toward 800
-----------------------------------
Tests:             150 lines  # Excluded per R007
Demos:              65 lines  # Excluded per R007
-----------------------------------
Implementation:    380/800 COMPLIANT (within limit)
```

### Demo Deliverables

Required Files:
- [ ] `demo-registry-client.sh` - Main demo script (executable)
- [ ] `DEMO.md` - Demo documentation per template
- [ ] `.demo-config` - Demo environment settings (registry URL, test image)

Integration Hooks:
- [ ] Export DEMO_READY=true when complete
- [ ] Provide integration point for wave demo
- [ ] Include cleanup function

---

## 8. Atomic PR Design (R220 MANDATORY)

```yaml
effort_atomic_pr_design:
  pr_summary: "Single PR implementing DefaultClient for RegistryClient interface"
  can_merge_to_main_alone: true

  r355_production_ready_checklist:
    no_hardcoded_values: true
    all_config_from_env: true  # Config from RegistryConfig struct
    no_stub_implementations: true
    no_todo_markers: true
    all_functions_complete: true

  configuration_approach:
    - name: "Registry URL"
      wrong: 'url := "https://localhost:5000"'
      correct: 'url := config.URL  // From RegistryConfig'
    - name: "Authentication"
      wrong: 'auth := &authn.Basic{Username: "admin", Password: "admin"}'
      correct: 'auth := &authn.Basic{Username: config.Username, Password: config.Password}'

  feature_flags_needed: []  # No feature flags required - complete implementation

  interfaces_to_implement:
    - interface: "RegistryClient"
      methods: ["Push"]
      implementation: "Complete in this PR"

  pr_verification:
    tests_pass_alone: true
    build_remains_working: true
    flags_tested_both_ways: false  # No flags
    no_external_dependencies: true  # go-containerregistry is added to go.mod
    backward_compatible: true
```

---

## 9. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| go-containerregistry API changes | Medium | Pin to specific version in go.mod |
| Error message format variations | Medium | Case-insensitive matching in classifyRemoteError |
| Progress callback timing | Low | Progress is optional (nil-safe) |
| TLS certificate issues | Medium | InsecureSkipVerify option available |
| Docker daemon unavailable in tests | Medium | Use `testing.Short()` guard |

---

## 10. Success Criteria

- [ ] `NewDefaultClient()` creates properly configured client
- [ ] All W2-RC-* tests pass
- [ ] Error classification works (AuthError vs RegistryError.IsTransient)
- [ ] Progress callbacks invoked at correct points
- [ ] TLS/insecure mode works correctly
- [ ] Context cancellation respected
- [ ] go-containerregistry dependency added to go.mod
- [ ] Coverage >= 80% on registry.go

---

## 11. Traceability

| Requirement | Implementation | Test |
|-------------|----------------|------|
| REQ-015 | `&authn.Basic{}` | W2-RC-001 |
| REQ-017 | `&authn.Bearer{}` | W2-RC-002 |
| REQ-019 | `authn.Anonymous` | W2-RC-003 |
| REQ-005 | Progress callbacks | W2-RC-004 |
| Error classification | `classifyRemoteError()` | W2-RC-005/006 |

---

## 12. Approvals

| Stakeholder | Role | Status | Date |
|-------------|------|--------|------|
| Code Reviewer Agent | Planning Authority | Approved | 2025-12-02 |
| Human Reviewer | Project Owner | Pending | - |

---

## 13. Reference Files

- Wave 2 Implementation Plan: `/home/vscode/workspaces/idpbuilder-planning/planning/phase1/wave2/WAVE-2-IMPLEMENTATION-PLAN.md`
- Wave 2 Architecture Plan: `/home/vscode/workspaces/idpbuilder-planning/planning/phase1/wave2/WAVE-2-ARCHITECTURE-PLAN.md`
- Wave 2 Test Plan: `/home/vscode/workspaces/idpbuilder-planning/planning/phase1/wave2/WAVE-2-TEST-PLAN.md`
- Existing interfaces: `pkg/registry/client.go`
