# E1.1.3 - Docker Daemon Client Interface + Mocks

## Implementation Plan

**Effort ID**: E1.1.3
**Effort Name**: Docker Daemon Client Interface + Mocks
**Phase**: 1 - Core OCI Push Implementation
**Wave**: 1 - Foundation (TDD)
**Created**: 2025-12-01T10:48:57Z
**Author**: Code Reviewer Agent
**Status**: Active

---

## EFFORT INFRASTRUCTURE METADATA

**EFFORT_NAME**: E1.1.3-daemon-client-interface
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave1/E1.1.3-daemon-client-interface
**BRANCH**: idpbuilder-oci-push/phase-1-wave-1-effort-E1.1.3-daemon-client-interface
**REMOTE**: origin (https://github.com/jessesanford/idpbuilder.git)
**BASE_BRANCH**: idpbuilder-oci-push/phase-1-wave-1-integration

---

## R213 Metadata Block

```yaml
effort_id: "E1.1.3"
effort_name: "daemon-client-interface"
estimated_lines: 300
dependencies: []
files_touched:
  - "pkg/daemon/client.go"
  - "pkg/daemon/client_test.go"
branch_name: "idpbuilder-oci-push/phase-1-wave-1-effort-E1.1.3-daemon-client-interface"
base_branch: "idpbuilder-oci-push/phase-1-wave-1-integration"
can_parallelize: true
parallel_with: ["E1.1.2"]
```

---

## PRIOR WORK ANALYSIS (R420 MANDATORY)

### Discovery Phase Results

- **Previous Efforts Reviewed**: E1.1.1 (credential-resolution), E1.1.2 (registry-client-interface)
- **Previous Plans Reviewed**: WAVE-1-IMPLEMENTATION-PLAN.md, WAVE-1-ARCHITECTURE-PLAN.md, WAVE-1-TEST-PLAN.md
- **Research Timestamp**: 2025-12-01T10:48:43Z
- **Research Status**: COMPLETE

### File Structure Findings

| File Path | Source Effort | Status | Action Required |
|-----------|---------------|--------|-----------------|
| pkg/daemon/client.go | E1.1.3 (this effort) | NEW | MUST create |
| pkg/daemon/client_test.go | E1.1.3 (this effort) | NEW | MUST create |
| pkg/registry/client.go | E1.1.2 | PARALLEL | DO NOT create (different package) |
| pkg/cmd/push/credentials.go | E1.1.1 | PARALLEL | DO NOT create (different package) |

### Interface/API Findings

| Interface/API | Source | Signature | Action Required |
|---------------|--------|-----------|-----------------|
| DaemonClient | Wave Architecture | `GetImage(ctx, ref) (*ImageInfo, ImageReader, error)`, `ImageExists(ctx, ref) (bool, error)`, `Ping(ctx) error` | MUST define exactly |
| ImageReader | Wave Architecture | `io.ReadCloser` | MUST define |
| DaemonError | Wave Architecture | `Error() string`, `Unwrap() error` | MUST implement |
| ImageNotFoundError | Wave Architecture | `Error() string` | MUST implement |

### Type/Struct Findings

| Type | Source | Exported | Action Required |
|------|--------|----------|-----------------|
| ImageInfo | Wave Architecture | YES | MUST define with ID, RepoTags, Size, LayerCount |
| DaemonClient | Wave Architecture | YES | MUST define as interface |
| ImageReader | Wave Architecture | YES | MUST define as interface |
| DaemonError | Wave Architecture | YES | MUST define with Message, IsNotRunning, Cause |
| ImageNotFoundError | Wave Architecture | YES | MUST define with Reference |

### Method Visibility Findings

| Method | Type | Visibility | Can Access? | Action Required |
|--------|------|------------|-------------|-----------------|
| GetImage | DaemonClient | EXPORTED | YES | Define in interface |
| ImageExists | DaemonClient | EXPORTED | YES | Define in interface |
| Ping | DaemonClient | EXPORTED | YES | Define in interface |
| Error | DaemonError | EXPORTED | YES | Implement |
| Unwrap | DaemonError | EXPORTED | YES | Implement |
| Error | ImageNotFoundError | EXPORTED | YES | Implement |

### Conflicts Detected

- NO duplicate file paths detected
- NO API mismatches detected
- NO method visibility violations detected

### Required Integrations

1. MUST use `github.com/stretchr/testify/mock` for mock implementations (already in go.mod v1.9.0)
2. MUST use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require` for test assertions
3. MUST follow existing idpbuilder test patterns (table-driven tests)

### Forbidden Actions

- DO NOT create any files in pkg/registry/ (E1.1.2 responsibility)
- DO NOT create any files in pkg/cmd/push/ (E1.1.1 responsibility)
- DO NOT modify go.mod (testify already present)
- DO NOT implement real Docker SDK calls (mock-only in Wave 1)

---

## Scope and Objectives

### Objective

Define the Docker daemon client interfaces, error types, and create mock implementations for testing. This effort provides the contract for interacting with the local Docker daemon that will be implemented in Wave 2.

### Scope

**In Scope:**
- `DaemonClient` interface definition with GetImage, ImageExists, Ping methods
- `ImageInfo` struct for image metadata
- `ImageReader` interface (wraps io.ReadCloser)
- `DaemonError` struct with error chaining (Unwrap support)
- `ImageNotFoundError` struct for missing images
- `MockDaemonClient` test implementation using testify/mock
- `MockImageReader` test implementation
- Comprehensive unit tests for all mock behaviors

**Out of Scope:**
- Real Docker SDK implementation (Wave 2)
- Integration with Docker daemon (Wave 2)
- CLI integration (Wave 3)

---

## Technical Approach

### Package Structure

```
pkg/
  daemon/
    client.go           # Interface definitions + error types (~100 lines)
    client_test.go      # Mock implementations + tests (~200 lines)
```

### Interface Design

The interfaces follow Go conventions:
- Small, focused interfaces (DaemonClient has 3 methods)
- Error types implement `error` interface and support `errors.Unwrap()`
- ImageReader wraps `io.ReadCloser` for clean resource management

---

## Implementation Steps

### Step 1: Create pkg/daemon Directory

```bash
mkdir -p pkg/daemon
```

### Step 2: Implement client.go (~100 lines)

Create `pkg/daemon/client.go` with the following structure:

```go
// pkg/daemon/client.go
package daemon

import (
    "context"
    "io"
)

// ImageInfo contains metadata about a local Docker image.
type ImageInfo struct {
    // ID is the image's unique identifier (digest)
    ID string
    // RepoTags are the image's repository tags (e.g., ["myapp:latest", "myapp:v1.0"])
    RepoTags []string
    // Size is the total image size in bytes
    Size int64
    // LayerCount is the number of layers in the image
    LayerCount int
}

// DaemonClient defines operations for interacting with the local Docker daemon.
// All operations work with the daemon's image store.
type DaemonClient interface {
    // GetImage retrieves an image from the local Docker daemon.
    // Parameters:
    //   - ctx: Context for cancellation and timeout
    //   - reference: Image reference (e.g., "myapp:latest", "sha256:abc...")
    // Returns:
    //   - ImageInfo with metadata about the image
    //   - ImageReader for accessing image content (caller must close)
    //   - Error if image not found or daemon unavailable
    GetImage(ctx context.Context, reference string) (*ImageInfo, ImageReader, error)

    // ImageExists checks if an image exists in the local Docker daemon.
    // This is a lighter-weight check than GetImage when you only need presence.
    // Parameters:
    //   - ctx: Context for cancellation and timeout
    //   - reference: Image reference to check
    // Returns:
    //   - true if image exists, false otherwise
    //   - Error only if daemon communication fails (not for missing images)
    ImageExists(ctx context.Context, reference string) (bool, error)

    // Ping checks connectivity to the Docker daemon.
    // Returns error if daemon is not available.
    Ping(ctx context.Context) error
}

// ImageReader provides access to image content for push operations.
// Callers must call Close() when done to release resources.
type ImageReader interface {
    io.ReadCloser
}

// DaemonError represents an error from the Docker daemon.
type DaemonError struct {
    // Message is a human-readable error description
    Message string
    // IsNotRunning indicates the daemon is not available
    IsNotRunning bool
    // Cause is the underlying error
    Cause error
}

// Error implements the error interface.
func (e *DaemonError) Error() string {
    if e.Cause != nil {
        return e.Message + ": " + e.Cause.Error()
    }
    return e.Message
}

// Unwrap implements errors.Unwrap for error chaining.
func (e *DaemonError) Unwrap() error {
    return e.Cause
}

// ImageNotFoundError is returned when a requested image doesn't exist locally.
type ImageNotFoundError struct {
    // Reference is the image reference that was not found
    Reference string
}

// Error implements the error interface.
func (e *ImageNotFoundError) Error() string {
    return "image not found: " + e.Reference
}
```

**Line Count Breakdown:**
- Package declaration + imports: 10 lines
- ImageInfo struct: 15 lines
- DaemonClient interface: 25 lines
- ImageReader interface: 5 lines
- DaemonError struct + methods: 25 lines
- ImageNotFoundError struct + method: 10 lines
- Comments and spacing: 10 lines
- **Total: ~100 lines**

### Step 3: Implement client_test.go (~200 lines)

Create `pkg/daemon/client_test.go` with mock implementations and tests:

```go
// pkg/daemon/client_test.go
package daemon

import (
    "bytes"
    "context"
    "errors"
    "io"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// MockDaemonClient implements DaemonClient for testing.
type MockDaemonClient struct {
    mock.Mock
}

// GetImage implements DaemonClient.GetImage for mocking.
func (m *MockDaemonClient) GetImage(ctx context.Context, reference string) (*ImageInfo, ImageReader, error) {
    args := m.Called(ctx, reference)
    var info *ImageInfo
    var reader ImageReader
    if args.Get(0) != nil {
        info = args.Get(0).(*ImageInfo)
    }
    if args.Get(1) != nil {
        reader = args.Get(1).(ImageReader)
    }
    return info, reader, args.Error(2)
}

// ImageExists implements DaemonClient.ImageExists for mocking.
func (m *MockDaemonClient) ImageExists(ctx context.Context, reference string) (bool, error) {
    args := m.Called(ctx, reference)
    return args.Bool(0), args.Error(1)
}

// Ping implements DaemonClient.Ping for mocking.
func (m *MockDaemonClient) Ping(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

// MockImageReader is a mock implementation of ImageReader for testing.
type MockImageReader struct {
    *bytes.Reader
}

func (m *MockImageReader) Close() error {
    return nil
}

// NewMockImageReader creates a new MockImageReader with the given data.
func NewMockImageReader(data []byte) *MockImageReader {
    return &MockImageReader{Reader: bytes.NewReader(data)}
}

// Test cases follow...
```

**Test Cases to Implement:**

1. **TestDaemonClient_GetImage_Success** - Successful image retrieval
2. **TestDaemonClient_GetImage_NotFound** - Image not found error
3. **TestDaemonClient_GetImage_DaemonNotRunning** - Daemon unavailable error
4. **TestDaemonClient_ImageExists_True** - Image exists
5. **TestDaemonClient_ImageExists_False** - Image does not exist
6. **TestDaemonClient_Ping_Success** - Ping succeeds
7. **TestDaemonClient_Ping_Failure** - Ping fails
8. **TestDaemonError_ErrorChaining** - Error wrapping works
9. **TestImageNotFoundError** - Error message format

**Line Count Breakdown:**
- Package declaration + imports: 15 lines
- MockDaemonClient implementation: 40 lines
- MockImageReader implementation: 15 lines
- TestDaemonClient_GetImage_Success: 25 lines
- TestDaemonClient_GetImage_NotFound: 20 lines
- TestDaemonClient_GetImage_DaemonNotRunning: 20 lines
- TestDaemonClient_ImageExists_True: 12 lines
- TestDaemonClient_ImageExists_False: 12 lines
- TestDaemonClient_Ping_Success: 10 lines
- TestDaemonClient_Ping_Failure: 15 lines
- TestDaemonError_ErrorChaining: 12 lines
- TestImageNotFoundError: 8 lines
- Comments and spacing: 16 lines
- **Total: ~200 lines**

### Step 4: Run Tests

```bash
# Compile packages
go build ./pkg/daemon/...

# Run tests
go test ./pkg/daemon/... -v

# Run with race detection
go test -race ./pkg/daemon/...

# Check coverage
go test ./pkg/daemon/... -cover
```

---

## TDD Test Plan (R400/R401)

### Test-First Approach

All tests are written FIRST. The implementation follows the tests.

### Test Cases (from Wave 1 Test Plan)

| Test ID | Test Function | Description | Priority |
|---------|---------------|-------------|----------|
| W1-DC-001 | TestDaemonClient_GetImage_Success | Successfully retrieve local image | Critical |
| W1-DC-002 | TestDaemonClient_GetImage_NotFound | ImageNotFoundError for missing image | Critical |
| W1-DC-003 | TestDaemonClient_GetImage_DaemonNotRunning | DaemonError when daemon unavailable | Critical |
| W1-DC-004 | TestDaemonClient_ImageExists_True | Returns true for existing image | High |
| W1-DC-005 | TestDaemonClient_ImageExists_False | Returns false for non-existent image | High |
| W1-DC-006 | TestDaemonClient_Ping_Success | Ping succeeds when daemon running | High |
| W1-DC-007 | TestDaemonClient_Ping_Failure | Ping fails when daemon unavailable | High |
| W1-DC-008 | TestDaemonError_ErrorChaining | DaemonError wrapping with Unwrap() | Medium |
| W1-DC-009 | TestImageNotFoundError | ImageNotFoundError message format | Medium |

### Test Pseudo-Code

```go
// W1-DC-001: TestDaemonClient_GetImage_Success
func TestDaemonClient_GetImage_Success(t *testing.T) {
    // GIVEN a mock daemon client with a configured successful response
    ctx := context.Background()
    mockClient := new(MockDaemonClient)
    expectedInfo := &ImageInfo{
        ID:         "sha256:abc123def456",
        RepoTags:   []string{"myapp:latest", "myapp:v1.0"},
        Size:       1024 * 1024 * 50,
        LayerCount: 5,
    }
    mockReader := NewMockImageReader([]byte("mock image data"))
    mockClient.On("GetImage", ctx, "myapp:latest").Return(expectedInfo, mockReader, nil)

    // WHEN GetImage is called
    info, reader, err := mockClient.GetImage(ctx, "myapp:latest")

    // THEN no error is returned
    require.NoError(t, err)
    // AND image info matches expected
    assert.Equal(t, "sha256:abc123def456", info.ID)
    assert.Equal(t, []string{"myapp:latest", "myapp:v1.0"}, info.RepoTags)
    // AND reader contains expected data
    data, _ := io.ReadAll(reader)
    assert.Equal(t, []byte("mock image data"), data)
    reader.Close()
}

// W1-DC-002: TestDaemonClient_GetImage_NotFound
func TestDaemonClient_GetImage_NotFound(t *testing.T) {
    // GIVEN a mock client configured to return ImageNotFoundError
    ctx := context.Background()
    mockClient := new(MockDaemonClient)
    notFoundErr := &ImageNotFoundError{Reference: "nonexistent:latest"}
    mockClient.On("GetImage", ctx, "nonexistent:latest").Return(nil, nil, notFoundErr)

    // WHEN GetImage is called with non-existent image
    info, reader, err := mockClient.GetImage(ctx, "nonexistent:latest")

    // THEN error is returned
    require.Error(t, err)
    // AND error is ImageNotFoundError
    var imgNotFound *ImageNotFoundError
    assert.True(t, errors.As(err, &imgNotFound))
    assert.Contains(t, imgNotFound.Error(), "image not found")
    // AND info and reader are nil
    assert.Nil(t, info)
    assert.Nil(t, reader)
}

// W1-DC-003: TestDaemonClient_GetImage_DaemonNotRunning
func TestDaemonClient_GetImage_DaemonNotRunning(t *testing.T) {
    // GIVEN a mock client configured to return DaemonError
    ctx := context.Background()
    mockClient := new(MockDaemonClient)
    daemonErr := &DaemonError{
        Message:      "Cannot connect to Docker daemon",
        IsNotRunning: true,
        Cause:        errors.New("connection refused"),
    }
    mockClient.On("GetImage", ctx, "myapp:latest").Return(nil, nil, daemonErr)

    // WHEN GetImage is called
    info, reader, err := mockClient.GetImage(ctx, "myapp:latest")

    // THEN DaemonError is returned
    require.Error(t, err)
    var de *DaemonError
    assert.True(t, errors.As(err, &de))
    assert.True(t, de.IsNotRunning)
}
```

---

## Demo Requirements (R330)

### Demo Scenario: Interface Definition Validation

**Objective**: Verify that all daemon client interfaces are properly defined and compile correctly.

**Demo Script**:

```bash
#!/bin/bash
# demo-e113.sh - E1.1.3 Interface Demo

echo "=========================================="
echo "E1.1.3: Docker Daemon Client Interface Demo"
echo "=========================================="

# Step 1: Verify interface definition
echo ""
echo "[1/4] Verifying interface definitions..."
grep -q "type DaemonClient interface" pkg/daemon/client.go && \
  echo "  DaemonClient interface: OK" || echo "  DaemonClient interface: MISSING"
grep -q "type ImageReader interface" pkg/daemon/client.go && \
  echo "  ImageReader interface: OK" || echo "  ImageReader interface: MISSING"
grep -q "type ImageInfo struct" pkg/daemon/client.go && \
  echo "  ImageInfo struct: OK" || echo "  ImageInfo struct: MISSING"
grep -q "type DaemonError struct" pkg/daemon/client.go && \
  echo "  DaemonError struct: OK" || echo "  DaemonError struct: MISSING"
grep -q "type ImageNotFoundError struct" pkg/daemon/client.go && \
  echo "  ImageNotFoundError struct: OK" || echo "  ImageNotFoundError struct: MISSING"

# Step 2: Compile package
echo ""
echo "[2/4] Compiling package..."
go build ./pkg/daemon/... && echo "  pkg/daemon: COMPILE OK"

# Step 3: Run tests
echo ""
echo "[3/4] Running tests..."
go test ./pkg/daemon/... -v 2>&1 | grep -E "^(---|\s+(PASS|FAIL|ok|FAIL))"

# Step 4: Coverage report
echo ""
echo "[4/4] Coverage summary..."
go test ./pkg/daemon/... -cover 2>&1 | grep -E "coverage:|ok|FAIL"

echo ""
echo "=========================================="
echo "E1.1.3 Demo Complete"
echo "=========================================="
```

**Success Criteria**:
1. All 5 interface/struct definitions present
2. Package compiles without errors
3. All 9 tests pass
4. No race conditions detected

---

## Acceptance Criteria

### Functional Criteria

- [ ] `ImageInfo` struct defined with ID, RepoTags, Size, LayerCount fields
- [ ] `DaemonClient` interface with GetImage, ImageExists, Ping methods
- [ ] `ImageReader` interface defined (io.ReadCloser)
- [ ] `DaemonError` implements `error` and `Unwrap()` interfaces
- [ ] `ImageNotFoundError` implements `error` interface
- [ ] `MockDaemonClient` works for all interface methods
- [ ] `MockImageReader` works with reader content verification

### Test Criteria

- [ ] All 9 daemon client tests pass
- [ ] `go test ./pkg/daemon/...` passes with 0 failures
- [ ] No race conditions (`go test -race`)
- [ ] Test coverage > 85%

### Quality Criteria

- [ ] Code follows Go conventions
- [ ] All exported types have documentation comments
- [ ] Error types support error chaining via Unwrap()
- [ ] Mock implementations use testify/mock correctly

---

## Size Estimates

**Estimated Lines**: ~300 lines

| File | Lines | Description |
|------|-------|-------------|
| pkg/daemon/client.go | ~100 | Interface definitions + error types |
| pkg/daemon/client_test.go | ~200 | Mock implementations + 9 test cases |

**Size Compliance**:
- **Soft Limit (R007)**: 700 lines - COMPLIANT (300 << 700)
- **Hard Limit (R007)**: 800 lines - COMPLIANT (300 << 800)
- **Code Reviewer Enforcement (R535)**: 900 lines - COMPLIANT (300 << 900)
- **Split Required**: NO

---

## Potential Issues and Mitigations

### Issue 1: Package Directory Creation

**Risk**: pkg/daemon directory may not exist
**Mitigation**: SW Engineer should create directory before creating files

### Issue 2: Import Path Consistency

**Risk**: Package may be imported with wrong path
**Mitigation**: Use relative imports from module root: `github.com/cnoe-io/idpbuilder/pkg/daemon`

### Issue 3: Mock Return Values

**Risk**: testify/mock may have nil pointer issues
**Mitigation**: Test cases explicitly check for nil before dereferencing

---

## References

- **Wave Implementation Plan**: `/planning/phase1/wave1/WAVE-1-IMPLEMENTATION-PLAN.md`
- **Wave Architecture Plan**: `/planning/phase1/wave1/WAVE-1-ARCHITECTURE-PLAN.md`
- **Wave Test Plan**: `/planning/phase1/wave1/WAVE-1-TEST-PLAN.md`
- **testify/mock Documentation**: https://github.com/stretchr/testify

---

## Approvals

| Stakeholder | Role | Status | Date |
|-------------|------|--------|------|
| Code Reviewer Agent | Implementation Planning Authority | Approved | 2025-12-01 |
| Human Reviewer | Project Owner | Pending | - |
