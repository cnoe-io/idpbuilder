# PROJECT FIX PLAN

## Bug Summary (R266 Documentation Reference)

This fix plan addresses the bug documented in the INTEGRATION-REPORT.md during project integration, as required by R266.

### Bug #1: Incorrect Import Paths in Registry Package
- **Source**: INTEGRATION-REPORT.md lines 84-92
- **R266 Compliance**: Bug was properly documented during integration, not fixed
- **Impact**: HIGH - Registry package cannot compile, blocking all registry functionality
- **Root Cause**: Import statements use incorrect repository owner (jessesanford instead of cnoe-io)

## Fix Strategy

### Bug #1: Incorrect Import Paths
- **Source Branch**: `origin/idpbuilder-oci-build-push/phase2/wave1/gitea-client`
  - This is the original branch where the registry package was implemented (E2.1.2)
  - Per R321, fixes must be applied to source branches, NOT the integration branch
- **Files to Fix**: pkg/registry/gitea.go (lines 14-16)
- **Fix Type**: Simple import path correction

#### Changes Required:
```go
// Current (INCORRECT):
import (
    "github.com/jessesanford/idpbuilder/pkg/certs"
    "github.com/jessesanford/idpbuilder/pkg/certvalidation"
    "github.com/jessesanford/idpbuilder/pkg/fallback"
)

// Fixed (CORRECT):
import (
    "github.com/cnoe-io/idpbuilder/pkg/certs"
    "github.com/cnoe-io/idpbuilder/pkg/certvalidation"
    "github.com/cnoe-io/idpbuilder/pkg/fallback"
)
```

## SW Engineer Spawn Instructions

### Single Fix Required
- **Engineer 1**: Fix import paths in gitea-client branch
  - **Working Directory**: Fresh clone/checkout of the repository
  - **Target Branch**: `origin/idpbuilder-oci-build-push/phase2/wave1/gitea-client`
  - **Task**: Update three import statements in pkg/registry/gitea.go
  - **Validation**: Ensure `go build ./pkg/registry` succeeds after fix

### Detailed Instructions for SW Engineer:
1. Clone/fetch the repository to get latest remote branches
2. Checkout `origin/idpbuilder-oci-build-push/phase2/wave1/gitea-client`
3. Navigate to pkg/registry/gitea.go
4. Update lines 14-16 to use correct import paths (cnoe-io instead of jessesanford)
5. Run `go build ./pkg/registry` to verify compilation
6. Run tests: `go test ./pkg/registry/...`
7. Commit with message: "fix: correct import paths in registry package (R266 bug fix)"
8. Push the fix back to the source branch

## Parallelization Strategy
- **Single bug = Single SW Engineer spawn**
- **No parallelization needed** - Only one bug to fix
- **Sequential execution not required** - Single atomic fix

## Post-Fix Integration

After fix is complete and verified:
1. **Verify Fix**: SW Engineer confirms build succeeds on source branch
2. **Merge to Source**: Push fix to `phase2/wave1/gitea-client` branch
3. **Re-run Integration**: Orchestrator will:
   - Re-create phase2/wave1/integration from phase1/integration
   - Re-merge fixed gitea-client branch
   - Re-merge to project-integration
4. **Verify Build**: Confirm full project builds successfully
5. **Run Full Test Suite**: Ensure all tests pass including registry tests
6. **Update Integration Report**: Document successful fix and build

## Success Criteria
- ✅ pkg/registry/gitea.go compiles successfully
- ✅ All registry package tests pass
- ✅ Full project build succeeds without errors
- ✅ Integration tests can run against registry

## Risk Assessment
- **Risk Level**: LOW - Simple string replacement
- **Testing Required**: Minimal - compilation and existing tests
- **Rollback Strategy**: Revert single commit if issues arise

---
**Plan Created**: 2025-09-09
**Code Reviewer**: CREATE_PROJECT_FIX_PLAN state
**R321 Compliance**: Fix targets source branch, not integration branch