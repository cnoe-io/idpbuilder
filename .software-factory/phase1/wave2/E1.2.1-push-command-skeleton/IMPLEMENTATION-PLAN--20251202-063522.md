# E1.2.1 Implementation Plan - Push Command Skeleton + Credential Resolution

**Phase**: Phase 1 - Core OCI Push Implementation
**Wave**: Wave 2 - Core Implementation
**Effort**: E1.2.1 - Push Command Skeleton + Credential Resolution
**Created**: 2025-12-02T06:35:22Z
**Author**: Code Reviewer Agent
**State**: EFFORT_PLAN_CREATION
**Fidelity Level**: **EXACT** (detailed effort specification with R213 metadata)

---

## EFFORT INFRASTRUCTURE METADATA (From Orchestrator)

**EFFORT_NAME**: E1.2.1-push-command-skeleton
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.1-push-command-skeleton
**BRANCH**: idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.1-push-command-skeleton
**REMOTE**: origin (https://github.com/jessesanford/idpbuilder.git)
**BASE_BRANCH**: idpbuilder-oci-push/phase-1-wave-2-integration
**PHASE_NUMBER**: 1
**WAVE_NUMBER**: 2

---

## R213 Effort Metadata

```yaml
effort_id: "E1.2.1"
effort_name: "push-command-skeleton"
estimated_lines: 350
dependencies: ["Wave 1 Integration (interfaces and credential resolver)"]
files_touched:
  - "pkg/cmd/push/push.go" (NEW)
  - "pkg/cmd/push/push_test.go" (NEW)
  - "pkg/cmd/push/register.go" (NEW)
  - "pkg/cmd/root.go" (MODIFICATION)
branch_name: "idpbuilder-oci-push/phase-1-wave-2-effort-E1.2.1-push-command-skeleton"
can_parallelize: false
parallel_with: []
base_branch: "idpbuilder-oci-push/phase-1-wave-2-integration"
```

---

## Pre-Planning Research Results (R374 MANDATORY)

### Existing Interfaces Found

| Interface | Location | Signature | Must Implement |
|-----------|----------|-----------|----------------|
| `CredentialResolver` | `pkg/cmd/push/credentials.go` | `Resolve(flags CredentialFlags, env Environment) (*Credentials, error)` | Use existing `DefaultCredentialResolver` |
| `RegistryClient` | `pkg/registry/client.go` | `Push(ctx, imageRef, destRef, progress) (*PushResult, error)` | Import for wiring (E1.2.2 implements) |
| `DaemonClient` | `pkg/daemon/client.go` | `GetImage/ImageExists/Ping` | Import for wiring (E1.2.3 implements) |
| `ProgressReporter` | `pkg/registry/client.go` | `Start/LayerProgress/LayerComplete/Complete/Error` | Import for wiring (E1.2.2 implements) |
| `Environment` | `pkg/cmd/push/credentials.go` | `Get(key string) string` | Use existing `DefaultEnvironment` |

### Existing Implementations to Reuse

| Component | Location | Purpose | How to Use |
|-----------|----------|---------|------------|
| `DefaultCredentialResolver` | `pkg/cmd/push/credentials.go` | Credential resolution with priority chain | Instantiate and call `Resolve()` |
| `DefaultEnvironment` | `pkg/cmd/push/credentials.go` | Reads environment variables | Pass to `DefaultCredentialResolver.Resolve()` |
| `CredentialFlags` | `pkg/cmd/push/credentials.go` | Struct for CLI flags | Populate from cobra flags |
| `Credentials` | `pkg/cmd/push/credentials.go` | Result struct with Username/Password/Token | Receive from `Resolve()` |
| Error Types | `pkg/registry/client.go`, `pkg/daemon/client.go` | `AuthError`, `RegistryError`, `DaemonError`, `ImageNotFoundError` | Use for error classification |

### APIs Already Defined

| API | Method | Signature | Notes |
|-----|--------|-----------|-------|
| CredentialResolver | Resolve | `(flags, env) -> (*Credentials, error)` | Wave 1 implementation exists |
| RegistryClient | Push | `(ctx, imageRef, destRef, progress) -> (*PushResult, error)` | E1.2.2 implements |
| DaemonClient | ImageExists | `(ctx, reference) -> (bool, error)` | E1.2.3 implements |
| DaemonClient | Ping | `(ctx) -> error` | E1.2.3 implements |

### FORBIDDEN DUPLICATIONS (R373)

- DO NOT create another CredentialResolver interface (already exists in pkg/cmd/push/credentials.go)
- DO NOT reimplement credential resolution logic (use `DefaultCredentialResolver`)
- DO NOT create alternative Environment interface (use `DefaultEnvironment`)
- DO NOT define new error types (use existing from pkg/registry and pkg/daemon)

### REQUIRED INTEGRATIONS (R373)

- MUST use `DefaultCredentialResolver` from `pkg/cmd/push/credentials.go`
- MUST use `DefaultEnvironment` for environment variable access
- MUST use `RegistryClient` interface for registry operations
- MUST use `DaemonClient` interface for daemon operations
- MUST use existing error types for error classification

---

## 1. Effort Overview

### 1.1 Purpose

This effort implements the Cobra push command skeleton that orchestrates the entire push workflow:
- Parses CLI flags for registry, authentication, and insecure mode
- Integrates with `DefaultCredentialResolver` from Wave 1
- Orchestrates daemon client (image existence check) and registry client (push)
- Handles context cancellation for Ctrl+C (REQ-013)
- Returns proper exit codes per POSIX conventions

### 1.2 Dependencies

| Dependency | Source | Used For |
|------------|--------|----------|
| `DefaultCredentialResolver` | Wave 1 E1.1.1 | Credential resolution |
| `CredentialFlags` | Wave 1 E1.1.1 | Flag struct |
| `DefaultEnvironment` | Wave 1 E1.1.1 | Environment variable access |
| `RegistryClient` interface | Wave 1 E1.1.2 | Registry operations |
| `RegistryConfig` | Wave 1 E1.1.2 | Registry configuration |
| `ProgressReporter` interface | Wave 1 E1.1.2 | Progress callbacks |
| `DaemonClient` interface | Wave 1 E1.1.3 | Daemon operations |
| `DaemonError`, `ImageNotFoundError` | Wave 1 E1.1.3 | Error types |

### 1.3 Execution Order

This effort MUST be completed first in Wave 2 because:
- Establishes the command structure used by tests
- Wires together the registry and daemon clients
- E1.2.2 and E1.2.3 implement interfaces used here

---

## 2. EXPLICIT SCOPE (R311 MANDATORY)

### IMPLEMENT EXACTLY:

**File: pkg/cmd/push/push.go** (~200 lines)
- Constant: `DefaultRegistry` (~5 lines)
- Variable: `PushCmd *cobra.Command` with Use, Short, Long, Args, RunE (~20 lines)
- Variables: Flag variables (flagRegistry, flagUsername, flagPassword, flagToken, flagInsecure) (~10 lines)
- Function: `init()` - registers flags (~15 lines)
- Function: `runPush(cmd, args) error` - main command logic (~80 lines)
- Function: `buildDestinationRef(registryURL, imageRef) string` (~15 lines)
- Function: `extractHost(registryURL) string` (~15 lines)
- Function: `parseImageRef(ref) (repo, tag)` (~20 lines)
- Function: `exitWithError(err) int` - error classification (~20 lines)

**File: pkg/cmd/push/push_test.go** (~120 lines)
- Test: `TestPushCmd_Success_OutputsReference` (~25 lines)
- Test: `TestPushCmd_CredentialIntegration` (~30 lines)
- Test: `TestPushCmd_ImageNotFound_ExitCode2` (~15 lines)
- Test: `TestPushCmd_DaemonNotRunning_ExitCode2` (~15 lines)
- Test: `TestPushCmd_AuthFailure_ExitCode1` (~15 lines)
- Test: `TestPushCmd_FlagParsing` (~15 lines)
- Helper: `createPushCmdWithDependencies(daemon, registry, progress) *cobra.Command` (~20 lines)
- Helper: `executePushWithExitCode(args, daemon, registry) int` (~15 lines)

**File: pkg/cmd/push/register.go** (~15 lines)
- Function: `AddToRoot(rootCmd *cobra.Command)` (~10 lines)

**File: pkg/cmd/root.go** (~5 lines added)
- Import: `"github.com/cnoe-io/idpbuilder/pkg/cmd/push"` (~1 line)
- Call: `push.AddToRoot(RootCmd)` in init() (~1 line)

### TOTAL: ~340 lines (well under 800 limit)

### DO NOT IMPLEMENT:

- DO NOT create mock implementations (tests use Wave 1 mocks via dependency injection)
- DO NOT implement RegistryClient (E1.2.2's responsibility)
- DO NOT implement DaemonClient (E1.2.3's responsibility)
- DO NOT implement ProgressReporter (E1.2.2's responsibility)
- DO NOT add logging configuration (use existing idpbuilder helpers)
- DO NOT add retry logic (future enhancement)
- DO NOT add verbose output mode (future enhancement)
- DO NOT add timeout configuration (use defaults)
- DO NOT add progress bar rendering (E1.2.2 implements basic progress)
- DO NOT create new error types (use existing from Wave 1)

---

## 3. Files to Create/Modify

### 3.1 pkg/cmd/push/push.go (NEW - ~200 lines)

**Purpose**: Main push command implementation with Cobra integration

**Function Signatures**:

```go
package push

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
    "github.com/cnoe-io/idpbuilder/pkg/daemon"
    "github.com/cnoe-io/idpbuilder/pkg/registry"
    "github.com/spf13/cobra"
)

const (
    // DefaultRegistry is the standard idpbuilder Gitea registry URL
    DefaultRegistry = "https://gitea.cnoe.localtest.me:8443"
)

// PushCmd represents the push command
var PushCmd = &cobra.Command{
    Use:   "push IMAGE",
    Short: "Push a local Docker image to an OCI registry",
    Long:  `Push a local Docker image to an OCI-compliant registry...`,
    Args:  cobra.ExactArgs(1),
    RunE:  runPush,
}

// Command flags (private package-level)
var (
    flagRegistry string
    flagUsername string
    flagPassword string
    flagToken    string
    flagInsecure bool
)

// init registers flags
func init() {...}

// runPush is the main command execution function
func runPush(cmd *cobra.Command, args []string) error {...}

// buildDestinationRef constructs the full registry reference
func buildDestinationRef(registryURL, imageRef string) string {...}

// extractHost extracts the host:port from a registry URL
func extractHost(registryURL string) string {...}

// parseImageRef extracts repository and tag from image reference
func parseImageRef(ref string) (repo, tag string) {...}

// exitWithError handles error classification and exit codes
func exitWithError(err error) int {...}
```

**Implementation Details**:

1. **Signal Handling (REQ-013)**: Setup `os/signal.Notify` for SIGINT/SIGTERM
2. **Credential Resolution**: Use `DefaultCredentialResolver.Resolve()`
3. **Daemon Check**: Call `daemonClient.ImageExists()` before push
4. **Registry Push**: Call `registryClient.Push()` with progress reporter
5. **Output**: Print only pushed reference to stdout (REQ-001)

**Exit Codes**:
| Code | Meaning | Trigger |
|------|---------|---------|
| 0 | Success | Push completed |
| 1 | General error | Auth failure, registry error |
| 2 | Resource not found | Image not found, daemon not running |
| 130 | Interrupted | User pressed Ctrl+C |

### 3.2 pkg/cmd/push/push_test.go (NEW - ~120 lines)

**Purpose**: Unit tests for push command using dependency injection for mocks

**Test Functions to Implement**:

```go
// TestPushCmd_Success_OutputsReference - W2-PC-001
func TestPushCmd_Success_OutputsReference(t *testing.T) {...}

// TestPushCmd_CredentialIntegration - W2-PC-002
func TestPushCmd_CredentialIntegration(t *testing.T) {...}

// TestPushCmd_ImageNotFound_ExitCode2 - W2-PC-003
func TestPushCmd_ImageNotFound_ExitCode2(t *testing.T) {...}

// TestPushCmd_DaemonNotRunning_ExitCode2 - W2-PC-004
func TestPushCmd_DaemonNotRunning_ExitCode2(t *testing.T) {...}

// TestPushCmd_AuthFailure_ExitCode1 - W2-PC-005
func TestPushCmd_AuthFailure_ExitCode1(t *testing.T) {...}

// TestPushCmd_FlagParsing - W2-PC-008
func TestPushCmd_FlagParsing(t *testing.T) {...}

// Helper: createPushCmdWithDependencies creates cmd with injectable deps
func createPushCmdWithDependencies(
    daemonClient daemon.DaemonClient,
    registryClient registry.RegistryClient,
    progress registry.ProgressReporter,
) *cobra.Command {...}

// Helper: executePushWithExitCode runs command and captures exit code
func executePushWithExitCode(args []string, daemon daemon.DaemonClient, registry registry.RegistryClient) int {...}
```

### 3.3 pkg/cmd/push/register.go (NEW - ~15 lines)

**Purpose**: Helper for registering push command with root

```go
package push

import "github.com/spf13/cobra"

// AddToRoot adds the push command to the root command.
// Called from pkg/cmd/root.go during initialization.
func AddToRoot(rootCmd *cobra.Command) {
    rootCmd.AddCommand(PushCmd)
}
```

### 3.4 pkg/cmd/root.go (MODIFICATION - ~5 lines added)

**Purpose**: Register push command at startup

**Changes**:
```go
// Add to imports
import (
    // ... existing imports ...
    "github.com/cnoe-io/idpbuilder/pkg/cmd/push"
)

// In init() or Execute(), add:
func init() {
    // ... existing command registrations ...
    push.AddToRoot(RootCmd)
}
```

---

## 4. Test Cases to Satisfy (from Wave 2 Test Plan)

### 4.1 Push Command Tests (W2-PC-*)

| Test ID | Description | Status |
|---------|-------------|--------|
| W2-PC-001 | Successful push outputs reference to stdout | Must Pass |
| W2-PC-002 | Credential integration (flags > env) | Must Pass |
| W2-PC-003 | Exit code 2 for missing local image | Must Pass |
| W2-PC-004 | Exit code 2 for daemon unavailable | Must Pass |
| W2-PC-005 | Exit code 1 for auth failure | Must Pass |
| W2-PC-006 | Exit code 1 for registry errors | Must Pass |
| W2-PC-007 | Context cancellation (Ctrl+C) | Should Pass |
| W2-PC-008 | All flags parsed correctly | Must Pass |
| W2-PC-009 | Default registry used | Must Pass |
| W2-PC-010 | Credentials never in logs | Must Pass |

---

## 5. Implementation Steps

### Step 1: Create push.go skeleton (~50 min)
1. Create `pkg/cmd/push/push.go`
2. Define `PushCmd` with proper metadata (Use, Short, Long, Args)
3. Register flags in `init()`
4. Implement `runPush()` main logic

### Step 2: Implement helper functions (~20 min)
1. `buildDestinationRef()` - construct registry reference
2. `extractHost()` - extract host from URL
3. `parseImageRef()` - extract repo:tag from reference

### Step 3: Implement error handling (~20 min)
1. Add exit code logic based on error types
2. Ensure credentials never logged
3. Add signal handler for Ctrl+C

### Step 4: Create register.go (~5 min)
1. Create `pkg/cmd/push/register.go`
2. Implement `AddToRoot()`

### Step 5: Modify root.go (~10 min)
1. Add import for push package
2. Register push command in init()

### Step 6: Write push_test.go (~60 min)
1. Create test file with test helper imports
2. Implement all W2-PC-* test functions
3. Create helper functions for test setup

### Step 7: Run tests and verify (~15 min)
1. `go test ./pkg/cmd/push/... -v`
2. Verify all tests pass
3. Check coverage >= 80%

---

## 6. R355 PRODUCTION READINESS - ZERO TOLERANCE

This implementation MUST be production-ready from the first commit:

- NO STUBS or placeholder implementations
- NO MOCKS except in test directories
- NO hardcoded credentials or secrets
- NO static configuration values (except DefaultRegistry which is a well-known constant)
- NO TODO/FIXME markers in code
- NO returning nil or empty for "later implementation"
- NO panic("not implemented") patterns
- NO fake or dummy data

### Configuration Requirements (R355 Mandatory)

**WRONG - Will fail review:**
```go
// VIOLATION - Hardcoded credential
password := "admin123"

// VIOLATION - Stub implementation
func ProcessPush() error {
    // TODO: implement later
    return nil
}
```

**CORRECT - Production ready:**
```go
// From flags or environment variable
creds, err := resolver.Resolve(credFlags, env)
if err != nil {
    return fmt.Errorf("credential resolution failed: %w", err)
}

// Full implementation with proper error handling
func runPush(cmd *cobra.Command, args []string) error {
    // Complete implementation as specified
}
```

---

## 7. Line Count Estimate

| File | Lines | Notes |
|------|-------|-------|
| push.go | ~200 | Main command implementation |
| push_test.go | ~120 | 6 test functions + 2 helpers |
| register.go | ~15 | Simple registration helper |
| root.go changes | ~5 | Import + AddCommand |
| **Total** | **~340** | Within 800 line limit |

**R220 Compliance**: 340 lines is well under the 800-line hard limit.

---

## 8. Size Limit Clarification (R359)

- The 800-line limit applies to NEW CODE YOU ADD
- Repository will grow by ~340 lines (EXPECTED)
- NEVER delete existing code to meet size limits
- This effort ADDS functionality on top of Wave 1 foundation

---

## 9. Demo Requirements (R330 MANDATORY)

### Demo Objectives (3-5 specific, verifiable objectives)

- [ ] Demonstrate `idpbuilder push --help` displays command help with all flags
- [ ] Show successful push outputs reference to stdout (not stderr)
- [ ] Verify exit code 2 when image not found in daemon
- [ ] Verify exit code 1 when authentication fails
- [ ] Display proper progress reporting to stderr during push

**Success Criteria**: All objectives checked = demo passes

### Demo Scenarios (IMPLEMENT EXACTLY THESE - 3 scenarios)

#### Scenario 1: Help Command Display
- **Setup**: idpbuilder binary built
- **Input**: None
- **Action**: `idpbuilder push --help`
- **Expected Output**:
  ```
  Push a local Docker image to an OCI registry.

  Usage:
    idpbuilder push IMAGE [flags]

  Flags:
    -r, --registry string   Registry URL (default "https://gitea.cnoe.localtest.me:8443")
    -u, --username string   Registry username
    -p, --password string   Registry password
    -t, --token string      Registry token
        --insecure          Skip TLS verification
  ```
- **Verification**: Output contains Usage and all 5 flags
- **Script Lines**: ~10 lines

#### Scenario 2: Image Not Found Error
- **Setup**: Ensure `nonexistent:test` does not exist locally
- **Input**: `nonexistent:test`
- **Action**: `idpbuilder push nonexistent:test; echo "Exit code: $?"`
- **Expected Output**:
  ```
  Error: image not found: nonexistent:test
  Exit code: 2
  ```
- **Verification**: Exit code equals 2
- **Script Lines**: ~10 lines

#### Scenario 3: Successful Push (with mock)
- **Setup**: Local test image exists, mock registry available
- **Input**: `test:latest`
- **Action**: `idpbuilder push test:latest --registry localhost:5000 --insecure`
- **Expected Output**:
  ```
  Pushing test:latest (3 layers)...
  localhost:5000/test:latest@sha256:abc123...
  Push complete: sha256:abc123...
  ```
- **Verification**: stdout contains pushed reference with digest
- **Script Lines**: ~15 lines

**TOTAL SCENARIO LINES**: ~35 lines

### Demo Size Planning

#### Demo Artifacts (Excluded from line count per R007)
```
demo-features.sh:     35 lines  # Executable script
DEMO.md:              50 lines  # Documentation
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

## 10. Effort Size Summary

```
Implementation:     340 lines  # ONLY this counts toward 800
Tests:             120 lines  # Excluded per R007
Demos:              85 lines  # Excluded per R007

Implementation:    340/800 = 42.5% (well within limit)
```

---

## 11. Dependencies on Other Wave 2 Efforts

| Dependency | Effort | Type |
|------------|--------|------|
| `NewDefaultClient(config)` | E1.2.2 | Factory for registry client |
| `daemon.NewDefaultClient()` | E1.2.3 | Factory for daemon client |
| `registry.StderrProgressReporter` | E1.2.2 | Progress implementation |

**Note**: Tests use dependency injection to mock these until E1.2.2 and E1.2.3 provide real implementations. The command structure is designed to accept injected clients for testing.

---

## 12. Atomic PR Design (R220 Compliance)

```yaml
effort_atomic_pr_design:
  pr_summary: "Single PR implementing push command skeleton with credential wiring"
  can_merge_to_main_alone: true

  r355_production_ready_checklist:
    no_hardcoded_values: true
    all_config_from_env: true
    no_stub_implementations: true
    no_todo_markers: true
    all_functions_complete: true

  feature_flags_needed:
    - flag: none
      reason: "Push command is fully functional, only depends on client implementations"

  interface_implementations:
    - interface: "Uses RegistryClient"
      implementation: "Will be injected (E1.2.2 provides DefaultClient)"
      production_ready: true
    - interface: "Uses DaemonClient"
      implementation: "Will be injected (E1.2.3 provides DefaultDaemonClient)"
      production_ready: true

  pr_verification:
    tests_pass_alone: true
    build_remains_working: true
    no_external_dependencies: true
    backward_compatible: true
```

---

## 13. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Exit code handling complexity | Medium | Use type assertions on error types |
| Signal handling edge cases | Low | Existing idpbuilder patterns available |
| Integration with root.go | Low | Simple AddCommand pattern |
| Credential logging prevention | High | Explicit checks, no String() on Credentials |
| Tests require mocks from E1.2.2/E1.2.3 | Medium | Use dependency injection pattern |

---

## 14. Success Criteria

- [ ] `idpbuilder push --help` displays command help
- [ ] All W2-PC-* tests pass
- [ ] Exit codes match specification (0, 1, 2, 130)
- [ ] Credentials never appear in any log output
- [ ] Default registry is `https://gitea.cnoe.localtest.me:8443`
- [ ] Context cancellation handled gracefully
- [ ] Code compiles with `go build ./...`
- [ ] Coverage >= 80% on push.go

---

## 15. Traceability

| Requirement | Implementation | Test |
|-------------|----------------|------|
| REQ-001 | `fmt.Println(result.Reference)` | W2-PC-001 |
| REQ-013 | Signal handling in runPush | W2-PC-007 |
| REQ-014 | CredentialResolver integration | W2-PC-002 |
| REQ-020 | No credential logging | W2-PC-010 |
| Exit codes | exitWithError() | W2-PC-003/4/5/6 |

---

## 16. Approvals

| Stakeholder | Role | Status | Date |
|-------------|------|--------|------|
| Code Reviewer Agent | Planning Authority | Approved | 2025-12-02 |
| Human Reviewer | Project Owner | Pending | - |
