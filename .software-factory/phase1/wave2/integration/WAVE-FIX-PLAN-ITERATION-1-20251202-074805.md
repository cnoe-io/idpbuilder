# Wave 2 Fix Plan - Iteration 1
## E1.2.1 Push Command Skeleton Bug Fixes

**Created**: 2025-12-02T07:48:05Z
**Effort**: E1.2.1-push-command-skeleton
**Phase**: 1, Wave: 2
**Bugs Addressed**: 3 (1 MEDIUM, 2 LOW)
**Estimated Complexity**: MEDIUM

---

## Bug Summary

| Bug ID | Severity | Summary | File | Lines |
|--------|----------|---------|------|-------|
| BUG-001-MOCK_INJECTION | MEDIUM | Test mock injection not functional | push_test.go | 387-395 |
| BUG-002-PARSE_IMAGEREF | LOW | parseImageRef edge case with semver tags | push.go | 174-192 |
| BUG-003-NIL_CLIENT | LOW | runPush nil client check blocking tests | push.go | 89-91 |

---

## Section 1: Architectural Root Cause Analysis

### BUG-001-MOCK_INJECTION (MEDIUM)

**Root Cause**: The `createPushCmdWithDependencies` function signature accepts mock clients as parameters but does NOT wire them into the command execution. The function simply wraps `PushCmd` (the global command) without any mechanism to inject the dependencies.

**Why This Occurred**:
- The design pattern intended for dependency injection is incomplete
- `runPush` uses package-level nil variables (`daemonClient`, `registryClient`) that are never set
- The test helper function was created as scaffolding but never connected to actual injection
- There's no mechanism to pass the mock clients from `createPushCmdWithDependencies` to `runPush`

**Impact Assessment**:
- **Test Coverage**: All tests using mocks are not actually testing the push command logic - they hit the nil check immediately
- **Development Blocking**: Tests cannot validate the main workflow without this fix
- **Severity**: MEDIUM - Blocks proper test execution and validation

### BUG-002-PARSE_IMAGEREF (LOW)

**Root Cause**: The `parseImageRef` function at lines 174-192 attempts to distinguish between ports/domains and tags by checking if the string after the last colon contains `.`, `/`, or `:`. However, semver tags like `v1.0` contain a dot, causing them to be incorrectly identified as a port/domain part.

**Code Analysis** (lines 181-185):
```go
potentialTag := ref[lastColon+1:]
if strings.ContainsAny(potentialTag, "./:") {
    // Looks like a port number or part of domain, not a tag
    return ref, ""
}
```

**Why This Occurred**:
- The heuristic assumes dots only appear in hostnames/ports, not tags
- Semver versioning (`v1.0`, `v2.3.1`) was not considered in the original design
- The function prioritizes avoiding false positives for ports over handling valid tag formats

**Impact Assessment**:
- Images with semver tags will not be parsed correctly
- Tags like `v1.0`, `v1.2.3`, `alpine3.18` will be incorrectly treated as the repository
- **Severity**: LOW - Affects a subset of tag formats, workaround possible (use full format)

### BUG-003-NIL_CLIENT (LOW)

**Root Cause**: The `runPush` function at lines 89-91 contains an early exit check:
```go
if daemonClient == nil || registryClient == nil {
    return fmt.Errorf("daemon or registry client not initialized")
}
```

This check is intentional scaffolding (see comment at lines 83-88), but it blocks all test execution when mock clients cannot be injected.

**Why This Occurred**:
- Placeholder for future client initialization (E1.2.2 and E1.2.3)
- The check is correct for production, but without BUG-001 fixed, tests always hit this
- This is actually a symptom of BUG-001, not an independent bug

**Relationship to BUG-001**:
- BUG-003 is **directly caused by** BUG-001
- Once mock injection works (BUG-001 fixed), this check will correctly allow mocks through
- No changes needed to this code IF BUG-001 is properly fixed

**Impact Assessment**:
- **Severity**: LOW - This is a symptom, not a root cause
- The nil check itself is correct production behavior
- Fix is dependent on BUG-001 resolution

---

## Section 2: Dependency Graph

```
BUG-001-MOCK_INJECTION (MEDIUM)
        |
        | (enables)
        v
BUG-003-NIL_CLIENT (LOW) -- RESOLVED BY BUG-001

BUG-002-PARSE_IMAGEREF (LOW) -- INDEPENDENT
```

### Fix Order (MANDATORY)

1. **First**: BUG-001-MOCK_INJECTION
   - Must be fixed first as it's the root cause blocking tests
   - BUG-003 is automatically resolved when this is fixed

2. **Second**: BUG-002-PARSE_IMAGEREF
   - Independent fix, can be done in parallel with BUG-001
   - But should be verified after BUG-001 is complete

3. **Third**: BUG-003-NIL_CLIENT
   - **No code changes required** - verify it's working after BUG-001 fix
   - Only add documentation comment if desired

### Affected Files

| File | Bug(s) | Type of Change |
|------|--------|----------------|
| pkg/cmd/push/push.go | BUG-001, BUG-002 | Structural refactor + logic fix |
| pkg/cmd/push/push_test.go | BUG-001 | Test helper update |

### Cascade Effects

- Changes to `push.go` will require re-running all push command tests
- Refactoring `runPush` signature will affect how tests call the command
- No impact on other efforts (E1.2.2, E1.2.3) as they consume interfaces only

---

## Section 3: File Ownership Determination

| File | Owner Effort | Rationale |
|------|--------------|-----------|
| pkg/cmd/push/push.go | E1.2.1 | Push command skeleton - this effort |
| pkg/cmd/push/push_test.go | E1.2.1 | Push command tests - this effort |
| pkg/daemon/interfaces.go | E1.2.2 | Daemon integration (not modified) |
| pkg/registry/interfaces.go | E1.2.3 | Registry integration (not modified) |

**Ownership Confirmation**: All bug fixes are within E1.2.1 scope. No cross-effort modifications needed.

---

## Section 4: Fix Implementation Steps

### BUG-001-MOCK_INJECTION - Detailed Fix

**Problem**: `createPushCmdWithDependencies` accepts mocks but doesn't inject them into command execution.

**Solution**: Refactor to support proper dependency injection.

**Option A (Recommended): Function parameter injection**

Modify `runPush` to accept clients as parameters and create a factory function:

```go
// BEFORE (lines 66-67):
func runPush(cmd *cobra.Command, args []string) error {

// AFTER:
// runPushWithClients is the internal implementation with injectable dependencies
func runPushWithClients(cmd *cobra.Command, args []string,
    daemonClient daemon.DaemonClient,
    registryClient registry.RegistryClient) error {

    imageRef := args[0]
    // ... rest of implementation using injected clients ...
    // Remove nil check - clients are now guaranteed by caller
}

// runPush is the production entry point
func runPush(cmd *cobra.Command, args []string) error {
    // Production client initialization (will be implemented in E1.2.2/E1.2.3)
    var daemonClient daemon.DaemonClient   // TODO: Initialize from E1.2.2
    var registryClient registry.RegistryClient // TODO: Initialize from E1.2.3

    if daemonClient == nil || registryClient == nil {
        return fmt.Errorf("daemon or registry client not initialized")
    }

    return runPushWithClients(cmd, args, daemonClient, registryClient)
}
```

**Update test helper (push_test.go lines 387-395)**:

```go
// BEFORE:
func createPushCmdWithDependencies(
    daemonClient daemon.DaemonClient,
    registryClient registry.RegistryClient,
) *PushCommandWrapper {
    return &PushCommandWrapper{
        baseCmd: PushCmd,
    }
}

// AFTER:
func createPushCmdWithDependencies(
    daemonClient daemon.DaemonClient,
    registryClient registry.RegistryClient,
) *PushCommandWrapper {
    // Create a NEW command with the injected dependencies
    testCmd := &cobra.Command{
        Use:   "push IMAGE",
        Short: "Push a local Docker image to an OCI registry",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runPushWithClients(cmd, args, daemonClient, registryClient)
        },
        SilenceErrors: true,
        SilenceUsage:  true,
    }

    // Copy flag definitions
    testCmd.Flags().StringVarP(&flagRegistry, "registry", "r", DefaultRegistry, "Registry URL")
    testCmd.Flags().StringVarP(&flagUsername, "username", "u", "", "Registry username")
    testCmd.Flags().StringVarP(&flagPassword, "password", "p", "", "Registry password")
    testCmd.Flags().StringVarP(&flagToken, "token", "t", "", "Registry token")
    testCmd.Flags().BoolVar(&flagInsecure, "insecure", false, "Skip TLS verification")

    return &PushCommandWrapper{
        baseCmd: testCmd,
    }
}
```

**Testing Requirements**:
- All existing tests should pass after this change
- Tests should now exercise the actual push logic (not just hit nil check)
- Verify mock clients are being called

### BUG-002-PARSE_IMAGEREF - Detailed Fix

**Problem**: Semver tags with dots (like `v1.0`) are incorrectly parsed.

**Solution**: Improve the heuristic to distinguish ports from tags.

```go
// BEFORE (lines 174-192):
func parseImageRef(ref string) (repo, tag string) {
    lastColon := strings.LastIndex(ref, ":")
    if lastColon > 0 {
        potentialTag := ref[lastColon+1:]
        if strings.ContainsAny(potentialTag, "./:") {
            return ref, ""
        }
        return ref[:lastColon], potentialTag
    }
    return ref, ""
}

// AFTER:
func parseImageRef(ref string) (repo, tag string) {
    lastColon := strings.LastIndex(ref, ":")
    if lastColon > 0 {
        potentialTag := ref[lastColon+1:]

        // Port numbers are purely numeric
        // Tags can contain alphanumeric, dots, dashes, underscores
        // Key insight: registry:port/image pattern has "/" after the port

        // If there's a "/" after the colon, it's definitely a port (registry:port/image)
        if strings.Contains(potentialTag, "/") {
            return ref, ""
        }

        // If it's purely numeric, likely a port number at start of ref
        // But we need to check if there's a "/" before the colon
        beforeColon := ref[:lastColon]
        if !strings.Contains(beforeColon, "/") {
            // No "/" before colon: could be "localhost:5000" (port) or "image:tag"
            // Check if numeric - if so, likely a port
            isAllDigits := true
            for _, c := range potentialTag {
                if c < '0' || c > '9' {
                    isAllDigits = false
                    break
                }
            }
            if isAllDigits && len(potentialTag) <= 5 {
                // Looks like a port number (1-65535 range, max 5 digits)
                return ref, ""
            }
        }

        // It's a tag (including semver like v1.0, alpine3.18)
        return ref[:lastColon], potentialTag
    }
    return ref, ""
}
```

**Add Test Cases**:
```go
// Add to TestParseImageRef test cases:
{
    name:       "Semver tag v1.0",
    ref:        "myimage:v1.0",
    expectRepo: "myimage",
    expectTag:  "v1.0",
},
{
    name:       "Semver tag v1.2.3",
    ref:        "myimage:v1.2.3",
    expectRepo: "myimage",
    expectTag:  "v1.2.3",
},
{
    name:       "Alpine style tag",
    ref:        "alpine:3.18",
    expectRepo: "alpine",
    expectTag:  "3.18",
},
{
    name:       "Registry with port and semver tag",
    ref:        "localhost:5000/myimage:v1.0",
    expectRepo: "localhost:5000/myimage",
    expectTag:  "v1.0",
},
```

### BUG-003-NIL_CLIENT - No Code Changes Required

**Verification Only**: After BUG-001 is fixed, verify that:
1. Tests with mock clients no longer hit the nil check
2. The nil check still works correctly for uninitialized production use

**Optional Enhancement** (documentation only):
```go
// Lines 89-91 - Add clarifying comment:
// NOTE: This check ensures production code has properly initialized clients.
// During testing, use runPushWithClients which receives mock clients directly.
if daemonClient == nil || registryClient == nil {
    return fmt.Errorf("daemon or registry client not initialized")
}
```

---

## Section 5: Integration Simulation Instructions (CRITICAL)

### Pre-Fix Verification
Before making any changes, verify the current broken state:

```bash
cd /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.1-push-command-skeleton

# Run tests to see current failures
go test -v ./pkg/cmd/push/... 2>&1 | head -50

# Expect: Tests fail with "daemon or registry client not initialized"
```

### Post-Fix Verification - BUG-001

After implementing BUG-001 fix:

```bash
cd /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.1-push-command-skeleton

# Build first to verify no compilation errors
go build ./...

# Run all push command tests
go test -v ./pkg/cmd/push/...

# Expected output:
# - TestPushCmd_Success_OutputsReference: PASS (mock registry called)
# - TestPushCmd_CredentialIntegration: PASS
# - TestPushCmd_ImageNotFound_ExitCode2: PASS (mock returns imageExists=false)
# - TestPushCmd_DaemonNotRunning_ExitCode2: PASS (mock returns pingErr)
# - TestPushCmd_AuthFailure_ExitCode1: PASS (mock returns pushErr)
# - TestPushCmd_FlagParsing: PASS
# - TestParseImageRef: PASS
# - TestBuildDestinationRef: PASS
# - TestExtractHost: PASS

# Verify mock is being called (not nil check)
# Look for actual test assertions being hit, not early exit
```

### Post-Fix Verification - BUG-002

After implementing BUG-002 fix:

```bash
cd /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/E1.2.1-push-command-skeleton

# Run parseImageRef specific tests
go test -v -run TestParseImageRef ./pkg/cmd/push/...

# Expected: All test cases pass, including new semver cases
# - "Semver tag v1.0" should return repo="myimage", tag="v1.0"
# - "Semver tag v1.2.3" should return repo="myimage", tag="v1.2.3"
```

### Integration Workspace Verification

**CRITICAL**: Test in integration context to catch any cross-effort issues:

```bash
cd /home/vscode/workspaces/idpbuilder-planning/efforts/phase1/wave2/integration

# Verify the fix works when merged into integration branch
# (This step depends on integration branch being set up)

# Run full test suite
go test -v ./...

# Build the full binary
go build -o idpbuilder ./cmd/idpbuilder

# Verify command is registered
./idpbuilder push --help
# Expected: Shows push command help with all flags
```

### Full Regression Test

```bash
# Run all project tests to ensure no regression
go test ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./pkg/cmd/push/...
# Expected: >80% coverage on push package
```

---

## Section 6: Success Criteria

### BUG-001-MOCK_INJECTION

| Criterion | Verification | Status |
|-----------|--------------|--------|
| Tests no longer fail with "client not initialized" | Run `go test -v ./pkg/cmd/push/...` | [ ] |
| Mock daemon client Ping() is called | Add debug log or breakpoint | [ ] |
| Mock registry client Push() is called | Test output contains mock reference | [ ] |
| All existing tests pass | Full test run | [ ] |
| New runPushWithClients function exists | Code review | [ ] |
| createPushCmdWithDependencies creates test command | Code review | [ ] |

### BUG-002-PARSE_IMAGEREF

| Criterion | Verification | Status |
|-----------|--------------|--------|
| `myimage:v1.0` parses correctly | TestParseImageRef passes | [ ] |
| `myimage:v1.2.3` parses correctly | TestParseImageRef passes | [ ] |
| `alpine:3.18` parses correctly | TestParseImageRef passes | [ ] |
| `localhost:5000/myimage` still works | TestParseImageRef passes | [ ] |
| `localhost:5000/myimage:v1.0` works | TestParseImageRef passes | [ ] |
| No regression on existing cases | All TestParseImageRef cases pass | [ ] |

### BUG-003-NIL_CLIENT

| Criterion | Verification | Status |
|-----------|--------------|--------|
| Production nil check still exists | Code review | [ ] |
| Tests bypass nil check via mock injection | Tests pass | [ ] |
| No code changes needed (resolved by BUG-001) | Verification only | [ ] |

### Overall Success Criteria

- [ ] All 3 bugs addressed
- [ ] `go build ./...` succeeds with no errors
- [ ] `go test ./...` passes with no failures
- [ ] Code coverage on push package >= 80%
- [ ] No new linting errors introduced
- [ ] Changes committed and pushed

### QA Acceptance Criteria

1. **Functional Tests Pass**: All unit tests in `push_test.go` pass
2. **Mock Injection Works**: Tests actually exercise push logic, not just fail at nil check
3. **Semver Tags Work**: Image references like `myimage:v1.0` parse correctly
4. **No Regression**: All previously passing tests still pass
5. **Build Succeeds**: `go build ./...` completes without errors
6. **Integration Ready**: Changes don't break integration with E1.2.2/E1.2.3

---

## Fix Plan Metadata

**Plan Created By**: Code Reviewer Agent
**Plan Creation Time**: 2025-12-02T07:48:05Z
**Plan Version**: 1.0
**Target Effort**: E1.2.1-push-command-skeleton
**Estimated Fix Time**: 1-2 hours
**Complexity Rating**: MEDIUM

### Bug Cross-References

- BUG-001-MOCK_INJECTION: Related to architectural scaffolding in effort plan
- BUG-002-PARSE_IMAGEREF: Edge case not covered in original specification
- BUG-003-NIL_CLIENT: Symptom of BUG-001, not independent issue

### Change Impact Summary

| Change Type | Files | Lines Modified (est) |
|-------------|-------|---------------------|
| Refactor | push.go | +30 / -10 |
| Refactor | push_test.go | +35 / -5 |
| New Tests | push_test.go | +20 |
| **Total** | 2 files | ~90 lines |

---

## R727 Compliance Checklist

- [x] Section 1: Architectural Root Cause Analysis (complete)
- [x] Section 2: Dependency Graph (complete)
- [x] Section 3: File Ownership Determination (complete)
- [x] Section 4: Fix Implementation Steps (complete)
- [x] Section 5: Integration Simulation Instructions (complete)
- [x] Section 6: Success Criteria (complete)
- [x] Minimum 1000 bytes (this document exceeds requirement)
