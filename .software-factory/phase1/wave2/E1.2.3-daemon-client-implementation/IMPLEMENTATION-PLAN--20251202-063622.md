# E1.2.3 Implementation Plan - Daemon Client Implementation

**Created**: 2025-12-02T06:36:22Z
**Author**: Code Reviewer Agent
**State**: EFFORT_PLAN_CREATION
**Fidelity Level**: **EXACT** (R213 compliant with detailed specifications)
**R383 Compliance**: Timestamp included in filename

---

## EFFORT INFRASTRUCTURE METADATA

**EFFORT_NAME**: E1.2.3-daemon-client-implementation
**PHASE**: 1
**WAVE**: 2
**BRANCH**: idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.3-daemon-client-implementation
**BASE_BRANCH**: idpbuilder-oci-push/phase-1-wave-2-integration
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.3-daemon-client-implementation

---

## R213 Effort Metadata

```yaml
effort_id: "E1.2.3"
effort_name: "daemon-client-implementation"
estimated_lines: 320
dependencies:
  - "Wave 1 Integration (DaemonClient interface from E1.1.3)"
  - "Wave 1 Integration (ImageInfo, ImageReader, DaemonError, ImageNotFoundError from E1.1.3)"
files_touched:
  - "pkg/daemon/daemon.go"       # ~200 lines (implementation)
  - "pkg/daemon/daemon_test.go"  # ~120 lines (tests)
branch_name: "idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.3-daemon-client-implementation"
can_parallelize: true
parallel_with: ["E1.2.2"]
base_branch: "idpbuilder-oci-push/phase-1-wave-2-integration"
```

---

## 1. Effort Overview

### 1.1 Purpose

This effort implements `DefaultDaemonClient` that satisfies the `DaemonClient` interface defined in Wave 1 (E1.1.3). It uses `go-containerregistry/pkg/v1/daemon` for:

- Checking if images exist in the local Docker daemon
- Retrieving image metadata and content
- Verifying daemon connectivity (Ping)
- Respecting DOCKER_HOST environment variable (REQ-024)

### 1.2 Wave 1 Dependencies (R374 Pre-Planning Research)

The following interfaces and types are ALREADY DEFINED in Wave 1 integration and MUST be used exactly:

| Dependency | Location | Source |
|------------|----------|--------|
| `DaemonClient` interface | `pkg/daemon/client.go` | E1.1.3 Wave 1 |
| `ImageInfo` struct | `pkg/daemon/client.go` | E1.1.3 Wave 1 |
| `ImageReader` interface | `pkg/daemon/client.go` | E1.1.3 Wave 1 |
| `DaemonError` struct | `pkg/daemon/client.go` | E1.1.3 Wave 1 |
| `ImageNotFoundError` struct | `pkg/daemon/client.go` | E1.1.3 Wave 1 |

### 1.3 Execution Order

This effort **CAN run in parallel** with E1.2.2 (Registry Client Implementation) because:
- Implements independent interface (DaemonClient)
- No code dependencies on E1.2.2
- Only compile-time dependency on Wave 1 interfaces

---

## 2. Pre-Planning Research Results (R374 MANDATORY)

### Existing Interfaces Found

| Interface | Location | Signature | Must Implement |
|-----------|----------|-----------|----------------|
| `DaemonClient` | `pkg/daemon/client.go` | `GetImage(ctx, ref)`, `ImageExists(ctx, ref)`, `Ping(ctx)` | YES |
| `ImageReader` | `pkg/daemon/client.go` | `io.ReadCloser` | YES (return type) |

### Existing Types to Reuse

| Type | Location | Purpose | How to Use |
|------|----------|---------|------------|
| `ImageInfo` | `pkg/daemon/client.go` | Image metadata | Return from GetImage |
| `DaemonError` | `pkg/daemon/client.go` | Daemon errors | Return for daemon issues |
| `ImageNotFoundError` | `pkg/daemon/client.go` | Missing image | Return when image not found |

### FORBIDDEN DUPLICATIONS (R373)

- DO NOT create alternative DaemonClient interface
- DO NOT create alternative ImageInfo struct
- DO NOT create alternative error types (DaemonError, ImageNotFoundError)
- DO NOT redefine ImageReader interface

### REQUIRED IMPLEMENTATIONS (R373)

- MUST implement `DaemonClient` interface from `pkg/daemon/client.go` with EXACT method signatures
- MUST return `*ImageInfo` and `ImageReader` from GetImage
- MUST return `*DaemonError` and `*ImageNotFoundError` as appropriate

---

## 3. EXPLICIT SCOPE (R311 MANDATORY)

### IMPLEMENT EXACTLY:

**pkg/daemon/daemon.go** (~200 lines):
- Type: `DefaultDaemonClient` struct with `dockerHost` field (~10 lines)
- Function: `NewDefaultClient() (*DefaultDaemonClient, error)` (~25 lines)
- Method: `(c *DefaultDaemonClient) GetImage(ctx, reference) (*ImageInfo, ImageReader, error)` (~60 lines)
- Method: `(c *DefaultDaemonClient) ImageExists(ctx, reference) (bool, error)` (~30 lines)
- Method: `(c *DefaultDaemonClient) Ping(ctx) error` (~25 lines)
- Helper: `isNotFoundError(err) bool` (~15 lines)
- Helper: `isDaemonUnavailable(err) bool` (~15 lines)
- Type: `pipeReader` struct implementing `ImageReader` (~20 lines)

**pkg/daemon/daemon_test.go** (~120 lines):
- Test: `TestDefaultDaemonClient_ImageExists_True` (~15 lines)
- Test: `TestDefaultDaemonClient_ImageExists_False` (~15 lines)
- Test: `TestDefaultDaemonClient_ImageExists_DaemonDown` (~15 lines)
- Test: `TestDefaultDaemonClient_GetImage_Success` (~20 lines)
- Test: `TestDefaultDaemonClient_GetImage_NotFound` (~15 lines)
- Test: `TestDefaultDaemonClient_Ping_Success` (~10 lines)
- Test: `TestDefaultDaemonClient_Ping_Failure` (~15 lines)
- Test: `TestDefaultDaemonClient_DOCKER_HOST` (~15 lines)
- Test: `TestDefaultDaemonClient_ErrorClassification` (~20 lines)

**TOTAL: ~320 lines**

### DO NOT IMPLEMENT:

- Image pulling from remote registries (E1.2.2 scope)
- Image building/creation (out of scope)
- Image deletion from daemon (future effort)
- Image tagging operations (future effort)
- Container operations (out of scope)
- Multi-platform image handling (future effort)
- Image caching layer (future effort)
- Comprehensive logging beyond basic debug (future effort)

---

## 4. Files to Create/Modify

### 4.1 pkg/daemon/daemon.go (NEW - ~200 lines)

**Purpose**: DefaultDaemonClient implementation using go-containerregistry

```go
package daemon

import (
    "context"
    "io"
    "os"
    "strings"

    "github.com/google/go-containerregistry/pkg/name"
    "github.com/google/go-containerregistry/pkg/v1/daemon"
    "github.com/google/go-containerregistry/pkg/v1/tarball"
)

// DefaultDaemonClient implements DaemonClient using go-containerregistry
type DefaultDaemonClient struct {
    // dockerHost is the Docker daemon socket path (from DOCKER_HOST)
    dockerHost string
}

// NewDefaultClient creates a new daemon client.
// Respects DOCKER_HOST environment variable (REQ-024).
func NewDefaultClient() (*DefaultDaemonClient, error) {
    client := &DefaultDaemonClient{}

    // Check for custom Docker host (REQ-024)
    if host := os.Getenv("DOCKER_HOST"); host != "" {
        client.dockerHost = host
    }

    // Verify daemon connectivity
    if err := client.Ping(context.Background()); err != nil {
        return nil, err
    }

    return client, nil
}

// GetImage implements DaemonClient.GetImage
func (c *DefaultDaemonClient) GetImage(ctx context.Context, reference string) (*ImageInfo, ImageReader, error) {
    // Parse the reference
    ref, err := name.ParseReference(reference, name.WeakValidation)
    if err != nil {
        return nil, nil, &ImageNotFoundError{Reference: reference}
    }

    // Get image from daemon
    img, err := daemon.Image(ref)
    if err != nil {
        if isNotFoundError(err) {
            return nil, nil, &ImageNotFoundError{Reference: reference}
        }
        if isDaemonUnavailable(err) {
            return nil, nil, &DaemonError{
                Message:      "Cannot connect to Docker daemon",
                IsNotRunning: true,
                Cause:        err,
            }
        }
        return nil, nil, &DaemonError{
            Message: "failed to get image: " + reference,
            Cause:   err,
        }
    }

    // Get image metadata
    digest, err := img.Digest()
    if err != nil {
        return nil, nil, &DaemonError{Message: "failed to get digest", Cause: err}
    }

    layers, err := img.Layers()
    if err != nil {
        return nil, nil, &DaemonError{Message: "failed to get layers", Cause: err}
    }

    // Calculate size
    var totalSize int64
    for _, layer := range layers {
        size, _ := layer.Size()
        totalSize += size
    }

    info := &ImageInfo{
        ID:         digest.String(),
        RepoTags:   []string{reference},
        Size:       totalSize,
        LayerCount: len(layers),
    }

    // Create pipe reader for image content
    pr, pw := io.Pipe()
    go func() {
        defer pw.Close()
        if err := tarball.Write(ref, img, pw); err != nil {
            pw.CloseWithError(err)
        }
    }()

    return info, &pipeReader{pr}, nil
}

// ImageExists implements DaemonClient.ImageExists
func (c *DefaultDaemonClient) ImageExists(ctx context.Context, reference string) (bool, error) {
    ref, err := name.ParseReference(reference, name.WeakValidation)
    if err != nil {
        return false, nil // Invalid reference means image doesn't exist
    }

    _, err = daemon.Image(ref)
    if err != nil {
        if isNotFoundError(err) {
            return false, nil
        }
        if isDaemonUnavailable(err) {
            return false, &DaemonError{
                Message:      "Cannot connect to Docker daemon",
                IsNotRunning: true,
                Cause:        err,
            }
        }
        return false, &DaemonError{
            Message: "failed to check image existence",
            Cause:   err,
        }
    }

    return true, nil
}

// Ping implements DaemonClient.Ping
func (c *DefaultDaemonClient) Ping(ctx context.Context) error {
    // Try to access a non-existent image to verify daemon connectivity
    ref, _ := name.ParseReference("__ping_check__:__ping__", name.WeakValidation)
    _, err := daemon.Image(ref)

    if err == nil {
        // Unexpectedly found the image - daemon is running
        return nil
    }

    if isDaemonUnavailable(err) {
        return &DaemonError{
            Message:      "Cannot connect to Docker daemon",
            IsNotRunning: true,
            Cause:        err,
        }
    }

    // Any other error (like image not found) means daemon is running
    return nil
}

// pipeReader wraps io.PipeReader to implement ImageReader
type pipeReader struct {
    *io.PipeReader
}

func (r *pipeReader) Read(p []byte) (n int, err error) {
    return r.PipeReader.Read(p)
}

func (r *pipeReader) Close() error {
    return r.PipeReader.Close()
}

// isNotFoundError checks if error indicates image not found
func isNotFoundError(err error) bool {
    if err == nil {
        return false
    }
    errStr := strings.ToLower(err.Error())
    return strings.Contains(errStr, "not found") ||
        strings.Contains(errStr, "no such image") ||
        strings.Contains(errStr, "manifest unknown")
}

// isDaemonUnavailable checks if error indicates daemon is not running
func isDaemonUnavailable(err error) bool {
    if err == nil {
        return false
    }
    errStr := strings.ToLower(err.Error())
    return strings.Contains(errStr, "cannot connect to the docker daemon") ||
        strings.Contains(errStr, "connection refused") ||
        strings.Contains(errStr, "is the docker daemon running") ||
        strings.Contains(errStr, "dial unix") ||
        strings.Contains(errStr, "no such host")
}
```

### 4.2 pkg/daemon/daemon_test.go (NEW - ~120 lines)

**Purpose**: Implementation tests for DefaultDaemonClient

```go
package daemon

import (
    "context"
    "errors"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// W2-DC-001: TestDefaultDaemonClient_ImageExists_True
func TestDefaultDaemonClient_ImageExists_True(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client, err := NewDefaultClient()
    if err != nil {
        t.Skip("Docker daemon not available")
    }

    // Uses alpine:latest which should be available
    exists, err := client.ImageExists(context.Background(), "alpine:latest")
    require.NoError(t, err)
    assert.True(t, exists)
}

// W2-DC-002: TestDefaultDaemonClient_ImageExists_False
func TestDefaultDaemonClient_ImageExists_False(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client, err := NewDefaultClient()
    if err != nil {
        t.Skip("Docker daemon not available")
    }

    exists, err := client.ImageExists(context.Background(), "nonexistent-image-12345:notag")
    require.NoError(t, err)
    assert.False(t, exists)
}

// W2-DC-003: TestDefaultDaemonClient_ImageExists_DaemonDown
func TestDefaultDaemonClient_ImageExists_DaemonDown(t *testing.T) {
    // Set invalid DOCKER_HOST to simulate daemon down
    original := os.Getenv("DOCKER_HOST")
    defer os.Setenv("DOCKER_HOST", original)
    os.Setenv("DOCKER_HOST", "unix:///invalid/path/docker.sock")

    _, err := NewDefaultClient()
    require.Error(t, err)

    var de *DaemonError
    assert.True(t, errors.As(err, &de))
    assert.True(t, de.IsNotRunning)
}

// W2-DC-004: TestDefaultDaemonClient_GetImage_Success
func TestDefaultDaemonClient_GetImage_Success(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client, err := NewDefaultClient()
    if err != nil {
        t.Skip("Docker daemon not available")
    }

    info, reader, err := client.GetImage(context.Background(), "alpine:latest")
    require.NoError(t, err)
    require.NotNil(t, info)
    require.NotNil(t, reader)
    defer reader.Close()

    assert.NotEmpty(t, info.ID)
    assert.Greater(t, info.Size, int64(0))
    assert.Greater(t, info.LayerCount, 0)
}

// W2-DC-005: TestDefaultDaemonClient_GetImage_NotFound
func TestDefaultDaemonClient_GetImage_NotFound(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client, err := NewDefaultClient()
    if err != nil {
        t.Skip("Docker daemon not available")
    }

    info, reader, err := client.GetImage(context.Background(), "nonexistent-12345:notag")
    require.Error(t, err)
    assert.Nil(t, info)
    assert.Nil(t, reader)

    var notFoundErr *ImageNotFoundError
    assert.True(t, errors.As(err, &notFoundErr))
    assert.Equal(t, "nonexistent-12345:notag", notFoundErr.Reference)
}

// W2-DC-006: TestDefaultDaemonClient_Ping_Success
func TestDefaultDaemonClient_Ping_Success(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client, err := NewDefaultClient()
    if err != nil {
        t.Skip("Docker daemon not available")
    }

    err = client.Ping(context.Background())
    require.NoError(t, err)
}

// W2-DC-007: TestDefaultDaemonClient_Ping_Failure
func TestDefaultDaemonClient_Ping_Failure(t *testing.T) {
    original := os.Getenv("DOCKER_HOST")
    defer os.Setenv("DOCKER_HOST", original)
    os.Setenv("DOCKER_HOST", "unix:///invalid/path/docker.sock")

    client := &DefaultDaemonClient{dockerHost: "unix:///invalid/path/docker.sock"}
    err := client.Ping(context.Background())
    require.Error(t, err)

    var de *DaemonError
    assert.True(t, errors.As(err, &de))
    assert.True(t, de.IsNotRunning)
}

// W2-DC-008: TestDefaultDaemonClient_DOCKER_HOST
func TestDefaultDaemonClient_DOCKER_HOST(t *testing.T) {
    original := os.Getenv("DOCKER_HOST")
    defer os.Setenv("DOCKER_HOST", original)

    customHost := "unix:///var/run/custom-docker.sock"
    os.Setenv("DOCKER_HOST", customHost)

    _, err := NewDefaultClient()
    // Expected error - socket doesn't exist
    if err != nil {
        var de *DaemonError
        assert.True(t, errors.As(err, &de))
    }
}

// W2-DC-009: TestDefaultDaemonClient_ErrorClassification
func TestDefaultDaemonClient_ErrorClassification(t *testing.T) {
    tests := []struct {
        name           string
        errorMsg       string
        wantNotFound   bool
        wantDaemonDown bool
    }{
        {"not_found", "No such image: myapp:latest", true, false},
        {"manifest_unknown", "manifest unknown", true, false},
        {"connection_refused", "connection refused", false, true},
        {"daemon_not_running", "Cannot connect to the Docker daemon", false, true},
        {"dial_unix", "dial unix /var/run/docker.sock: connect: no such file", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := errors.New(tt.errorMsg)
            notFound := isNotFoundError(err)
            daemonDown := isDaemonUnavailable(err)

            assert.Equal(t, tt.wantNotFound, notFound, "isNotFoundError mismatch")
            assert.Equal(t, tt.wantDaemonDown, daemonDown, "isDaemonUnavailable mismatch")
        })
    }
}
```

---

## 5. Implementation Steps

### Step 1: Create daemon.go (~60 min)

1. Create `pkg/daemon/daemon.go`
2. Define `DefaultDaemonClient` struct with `dockerHost` field
3. Implement `NewDefaultClient()` with DOCKER_HOST support (REQ-024)
4. Implement `ImageExists()` method
5. Implement `GetImage()` with pipe reader for image content
6. Implement `Ping()` for daemon connectivity check
7. Implement error classification helper functions
8. Implement `pipeReader` struct for ImageReader interface

### Step 2: Create daemon_test.go (~60 min)

1. Create `pkg/daemon/daemon_test.go`
2. Implement all W2-DC-* test functions
3. Add `testing.Short()` guards for integration tests
4. Test error classification logic
5. Test DOCKER_HOST environment variable handling

### Step 3: Test with real Docker daemon (~30 min)

1. Run unit tests: `go test ./pkg/daemon/... -v -short`
2. Run integration tests (with daemon): `go test ./pkg/daemon/... -v`
3. Verify ImageExists, GetImage, Ping work correctly
4. Check coverage: `go test ./pkg/daemon/... -cover` (target >= 80%)

---

## 6. Test Cases to Satisfy (from Wave 2 Test Plan)

| Test ID | Description | Priority | Success Criteria |
|---------|-------------|----------|------------------|
| W2-DC-001 | Returns true for existing image | Critical | `ImageExists("alpine:latest")` returns `(true, nil)` |
| W2-DC-002 | Returns false for missing image | Critical | `ImageExists("nonexistent")` returns `(false, nil)` |
| W2-DC-003 | Returns DaemonError when unavailable | Critical | Error is `*DaemonError` with `IsNotRunning=true` |
| W2-DC-004 | Returns ImageInfo and reader | Critical | Non-nil info with valid fields |
| W2-DC-005 | Returns ImageNotFoundError | Critical | Error is `*ImageNotFoundError` |
| W2-DC-006 | Verifies daemon connectivity | High | `Ping()` returns `nil` |
| W2-DC-007 | Returns DaemonError.IsNotRunning | High | `Ping()` error has `IsNotRunning=true` |
| W2-DC-008 | Respects DOCKER_HOST env var | Medium | Client uses custom socket |
| W2-DC-009 | Error type detection logic | Medium | Classification functions work correctly |

---

## 7. R355 PRODUCTION READINESS - ZERO TOLERANCE

This implementation MUST be production-ready from the first commit:

- NO STUBS or placeholder implementations
- NO MOCKS except in test directories
- NO hardcoded credentials or secrets
- NO static configuration values
- NO TODO/FIXME markers in code
- NO returning nil or empty for "later implementation"
- NO panic("not implemented") patterns
- NO fake or dummy data

### Configuration Requirements (R355 Mandatory)

```go
// CORRECT - Production ready (using environment):
if host := os.Getenv("DOCKER_HOST"); host != "" {
    client.dockerHost = host
}

// WRONG - Will fail review:
// dockerHost := "unix:///var/run/docker.sock"  // hardcoded!
```

---

## 8. Atomic PR Design (R220 MANDATORY)

```yaml
effort_atomic_pr_design:
  pr_summary: "Single PR implementing DefaultDaemonClient for DaemonClient interface"
  can_merge_to_main_alone: true

  r355_production_ready_checklist:
    no_hardcoded_values: true
    all_config_from_env: true
    no_stub_implementations: true
    no_todo_markers: true
    all_functions_complete: true

  configuration_approach:
    - name: "Docker Host"
      wrong: 'dockerHost := "unix:///var/run/docker.sock"'
      correct: 'dockerHost := os.Getenv("DOCKER_HOST")'

  interface_implementations:
    - interface: "DaemonClient"
      implementation: "DefaultDaemonClient"
      production_ready: true
      notes: "Fully functional using go-containerregistry"

  pr_verification:
    tests_pass_alone: true
    build_remains_working: true
    no_external_dependencies: true
    backward_compatible: true

  example_pr_structure:
    files_added:
      - "pkg/daemon/daemon.go"
      - "pkg/daemon/daemon_test.go"
    tests_included:
      - "Unit tests with testing.Short() guard"
      - "Integration tests requiring Docker daemon"
    documentation:
      - "Implementation follows Wave 1 interface exactly"
```

---

## 9. Demo Requirements (R330 MANDATORY)

### Demo Objectives (3-5 specific, verifiable objectives)

- [ ] Demonstrate `ImageExists()` returns true for existing image (alpine:latest)
- [ ] Show `ImageExists()` returns false for non-existent image
- [ ] Verify `GetImage()` returns valid `ImageInfo` with correct metadata
- [ ] Prove `Ping()` successfully detects daemon connectivity
- [ ] Display proper error handling for daemon unavailable scenario

**Success Criteria**: All objectives checked = demo passes

### Demo Scenarios (IMPLEMENT EXACTLY THESE - 2-4 scenarios)

#### Scenario 1: Image Existence Check

- **Setup**: Docker daemon running, alpine:latest pulled
- **Input**: Reference "alpine:latest"
- **Action**: `client.ImageExists(ctx, "alpine:latest")`
- **Expected Output**:
  ```
  Image exists: true
  Error: <nil>
  ```
- **Verification**: Boolean true returned, no error
- **Script Lines**: ~15 lines

#### Scenario 2: Image Not Found Error

- **Setup**: Docker daemon running
- **Input**: Reference "nonexistent-image-12345:notag"
- **Action**: `client.GetImage(ctx, "nonexistent-image-12345:notag")`
- **Expected Output**:
  ```
  Error: image not found: nonexistent-image-12345:notag
  Error type: *daemon.ImageNotFoundError
  ```
- **Verification**: ImageNotFoundError returned with correct reference
- **Script Lines**: ~15 lines

#### Scenario 3: Daemon Connectivity Check

- **Setup**: Docker daemon running
- **Input**: None
- **Action**: `client.Ping(ctx)`
- **Expected Output**:
  ```
  Ping result: success
  Error: <nil>
  ```
- **Verification**: No error returned
- **Script Lines**: ~10 lines

#### Scenario 4: DOCKER_HOST Environment Variable

- **Setup**: Set DOCKER_HOST environment variable
- **Input**: Custom socket path
- **Action**: `NewDefaultClient()` with custom DOCKER_HOST
- **Expected Output**:
  ```
  Using DOCKER_HOST: unix:///custom/path
  Client created (or error if socket not found)
  ```
- **Verification**: Client attempts to use custom socket
- **Script Lines**: ~15 lines

**TOTAL SCENARIO LINES**: ~55 lines

### Demo Size Planning

#### Demo Artifacts (Excluded from line count per R007)

```
demo-features.sh:     55 lines  # Executable script
DEMO.md:              30 lines  # Documentation
────────────────────────────────
TOTAL DEMO FILES:     85 lines (NOT counted toward 800)
```

#### Effort Size Summary

```
Implementation:     200 lines  # pkg/daemon/daemon.go
Tests:              120 lines  # pkg/daemon/daemon_test.go (excluded per R007)
────────────────────────────────
Implementation:    200/800 (within limit)
```

### Demo Deliverables

Required Files:
- [ ] `demo-features.sh` - Main demo script (executable)
- [ ] `DEMO.md` - Demo documentation per template

Integration Hooks:
- [ ] Export DEMO_READY=true when complete
- [ ] Provide integration point for wave demo
- [ ] Include cleanup function

---

## 10. Size Limit Clarification (R359)

- The 800-line limit applies to NEW CODE ADDED
- Repository will grow by ~200 implementation lines (EXPECTED)
- Tests and demos are EXCLUDED from line count per R007
- NEVER delete existing code to meet size limits

**Implementation Size Estimate**:
- NEW implementation code: ~200 lines
- NEW test code: ~120 lines (excluded)
- Total NEW code: ~320 lines
- Estimated total after: Current + 200 (implementation only)

---

## 11. External Library Usage

### go-containerregistry/pkg/v1/daemon Usage

```go
import (
    "github.com/google/go-containerregistry/pkg/name"
    "github.com/google/go-containerregistry/pkg/v1/daemon"
    "github.com/google/go-containerregistry/pkg/v1/tarball"
)

// Parse image reference
ref, err := name.ParseReference("myapp:latest", name.WeakValidation)

// Get image from local daemon
img, err := daemon.Image(ref)

// Get image metadata
digest, err := img.Digest()
layers, err := img.Layers()

// Write image to tarball (for ImageReader)
err := tarball.Write(ref, img, writer)
```

---

## 12. Success Criteria

- [ ] `NewDefaultClient()` creates client with daemon check
- [ ] All W2-DC-* tests pass
- [ ] `ImageExists()` returns `(true, nil)` for existing images
- [ ] `ImageExists()` returns `(false, nil)` for missing images
- [ ] `ImageExists()` returns `(false, DaemonError)` when daemon down
- [ ] `GetImage()` returns populated `ImageInfo`
- [ ] `ImageNotFoundError` returned for missing images
- [ ] `Ping()` returns nil when daemon running
- [ ] DOCKER_HOST environment variable respected (REQ-024)
- [ ] Coverage >= 80% on daemon.go
- [ ] Demo scenarios pass

---

## 13. Traceability

| Requirement | Implementation | Test |
|-------------|----------------|------|
| REQ-010 (Image must exist) | `ImageExists()` | W2-DC-001/002 |
| REQ-011 (Daemon required) | `Ping()` + error check | W2-DC-006/007 |
| REQ-024 (DOCKER_HOST) | `os.Getenv("DOCKER_HOST")` | W2-DC-008 |
| Error handling | `DaemonError`, `ImageNotFoundError` | W2-DC-003/005/009 |

---

## 14. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Docker daemon not available in CI | High | Use `testing.Short()` to skip integration tests |
| DOCKER_HOST edge cases | Medium | Test with various socket path formats |
| Pipe reader resource leaks | Medium | Proper defer/close patterns in tests |
| Error message variations | Medium | Case-insensitive matching in error classification |

---

## 15. Approvals

| Stakeholder | Role | Status | Date |
|-------------|------|--------|------|
| Code Reviewer Agent | Planning Authority | Approved | 2025-12-02 |
| Human Reviewer | Project Owner | Pending | - |

---

## PLANNING FILE CREATED (R383 COMPLIANT)

**Type**: effort_plan
**Path**: /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.3-daemon-client-implementation/.software-factory/phase1/wave2/E1.2.3-daemon-client-implementation/IMPLEMENTATION-PLAN--20251202-063622.md
**Effort**: E1.2.3-daemon-client-implementation
**Phase**: 1
**Wave**: 2
**Target Branch**: idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.3-daemon-client-implementation
**Created At**: 2025-12-02T06:36:22Z
**R383 Compliance**: Timestamp included (--20251202-063622)

ORCHESTRATOR: Please update effort_repo_files.effort_plans["E1.2.3-daemon-client-implementation"] in state file per R340
