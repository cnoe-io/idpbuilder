# Phase 1 Wave 1 Integration Merge Plan

## Metadata
- **Created By**: Code Reviewer Agent (WAVE_MERGE_PLANNING state)
- **Creation Time**: 2025-09-18T22:50:00Z
- **Phase**: 1
- **Wave**: 1
- **Integration Branch**: `idpbuilder-oci-build-push/phase1-wave1-integration`
- **Base Branch**: `main`
- **Total Efforts**: 5

## Critical Requirements (R269, R270 Compliance)
✅ Using ONLY original effort branches as sources
✅ No integration branches used as merge sources
✅ Dependency-aware merge ordering
✅ Conflict resolution strategies documented
✅ This is a PLAN ONLY - no actual merges executed

## Efforts to Integrate

| Order | Effort ID | Effort Name | Branch | Lines | Dependencies | Status |
|-------|-----------|-------------|--------|-------|--------------|--------|
| 1 | E1.1.1 | kind-cert-extraction | `idpbuilder-oci-build-push/phase1/wave1/kind-cert-extraction` | 450 | None | COMPLETE |
| 2 | E1.1.2A | registry-types | `idpbuilder-oci-build-push/phase1/wave1/registry-types` | 205 | E1.1.1 | COMPLETE |
| 3 | E1.1.2B | registry-auth | `idpbuilder-oci-build-push/phase1/wave1/registry-auth` | 363 | E1.1.2A | COMPLETE |
| 4 | E1.1.2C | registry-helpers | `idpbuilder-oci-build-push/phase1/wave1/registry-helpers` | 684 | E1.1.2B | COMPLETE |
| 5 | E1.1.2D | registry-tests | `idpbuilder-oci-build-push/efforts/phase1/wave1/registry-tests` | 115 | E1.1.2C | COMPLETE (test-only) |

**Total Implementation Lines**: 1,702 (excluding test-only effort per R007)
**Total Including Tests**: 1,817

## Merge Order Analysis

### Dependency Chain
```
main
  └─> E1.1.1 (kind-cert-extraction) - Foundational, no dependencies
      └─> E1.1.2A (registry-types) - Core types, depends on E1.1.1
          └─> E1.1.2B (registry-auth) - Auth logic, depends on E1.1.2A
              └─> E1.1.2C (registry-helpers) - Utilities, depends on E1.1.2B
                  └─> E1.1.2D (registry-tests) - Tests, depends on E1.1.2C
```

### Rationale for Order
1. **kind-cert-extraction** must merge first as it's the foundational effort with no dependencies
2. **registry-types** depends on kind-cert-extraction and provides core types for subsequent efforts
3. **registry-auth** builds on registry-types, implementing authentication logic
4. **registry-helpers** extends registry-auth with utility functions
5. **registry-tests** is test-only code that validates all previous implementations

## Pre-Merge Verification Steps

### Step 0: Environment Setup
```bash
# Ensure clean integration workspace
cd /home/vscode/workspaces/this-is-not-the-target-repo-this-is-for-orchestrator-planning-only/efforts/phase1/wave1/integration-workspace/repo
git status
# Should show: On branch idpbuilder-oci-build-push/phase1-wave1-integration

# Fetch all remotes to ensure we have latest
git fetch --all

# Verify integration branch is at main
git diff origin/main --stat
# Should show: no differences (clean slate)
```

### Step 1: Verify All Effort Branches
```bash
# For each effort, verify it exists and is complete
for effort_dir in kind-cert-extraction registry-types registry-auth registry-helpers registry-tests; do
    echo "Checking $effort_dir..."
    if [ -d "../../$effort_dir/.git" ]; then
        cd "../../$effort_dir"
        git log --oneline -1
        cd -
    fi
done
```

## Detailed Merge Instructions

### Merge 1: kind-cert-extraction
```bash
# Navigate to integration directory
cd /home/vscode/workspaces/this-is-not-the-target-repo-this-is-for-orchestrator-planning-only/efforts/phase1/wave1/integration-workspace/repo

# Add effort repository as remote if not exists
git remote add kind-cert-extraction ../../kind-cert-extraction 2>/dev/null || true

# Fetch the effort branch
git fetch kind-cert-extraction

# Merge with descriptive message
git merge kind-cert-extraction/idpbuilder-oci-build-push/phase1/wave1/kind-cert-extraction \
    --no-ff \
    -m "feat(integration): merge E1.1.1 kind-cert-extraction into phase1-wave1

Integrates foundational Kind certificate extraction functionality
- Adds certificate extraction from Kind clusters
- Implements secure storage mechanisms
- Provides base infrastructure for registry auth
- Implementation: 450 lines"

# Verify merge success
git log --oneline -1
git diff --stat HEAD~1
```

### Merge 2: registry-types
```bash
# Add effort repository as remote if not exists
git remote add registry-types ../../registry-types 2>/dev/null || true

# Fetch the effort branch
git fetch registry-types

# Merge with descriptive message
git merge registry-types/idpbuilder-oci-build-push/phase1/wave1/registry-types \
    --no-ff \
    -m "feat(integration): merge E1.1.2A registry-types into phase1-wave1

Integrates core registry type definitions
- Defines registry configuration structures
- Implements type validation
- Provides interfaces for auth components
- Implementation: 205 lines
- Dependencies: E1.1.1 (kind-cert-extraction)"

# Verify merge success
git log --oneline -1
git diff --stat HEAD~1
```

### Merge 3: registry-auth
```bash
# Add effort repository as remote if not exists
git remote add registry-auth ../../registry-auth 2>/dev/null || true

# Fetch the effort branch
git fetch registry-auth

# Merge with descriptive message
git merge registry-auth/idpbuilder-oci-build-push/phase1/wave1/registry-auth \
    --no-ff \
    -m "feat(integration): merge E1.1.2B registry-auth into phase1-wave1

Integrates registry authentication logic
- Implements authentication handlers
- Adds credential management
- Provides auth middleware
- Implementation: 363 lines
- Dependencies: E1.1.2A (registry-types)"

# Verify merge success
git log --oneline -1
git diff --stat HEAD~1
```

### Merge 4: registry-helpers
```bash
# Add effort repository as remote if not exists
git remote add registry-helpers ../../registry-helpers 2>/dev/null || true

# Fetch the effort branch
git fetch registry-helpers

# Merge with descriptive message
git merge registry-helpers/idpbuilder-oci-build-push/phase1/wave1/registry-helpers \
    --no-ff \
    -m "feat(integration): merge E1.1.2C registry-helpers into phase1-wave1

Integrates registry helper utilities
- Adds utility functions for registry operations
- Implements convenience wrappers
- Provides shared helper methods
- Implementation: 684 lines
- Dependencies: E1.1.2B (registry-auth)"

# Verify merge success
git log --oneline -1
git diff --stat HEAD~1
```

### Merge 5: registry-tests
```bash
# Add effort repository as remote if not exists
git remote add registry-tests ../../registry-tests 2>/dev/null || true

# Fetch the effort branch
git fetch registry-tests

# Note: Branch name might be slightly different based on git output
# Use the correct branch name: idpbuilder-oci-build-push/efforts/phase1/wave1/registry-tests
git merge registry-tests/idpbuilder-oci-build-push/efforts/phase1/wave1/registry-tests \
    --no-ff \
    -m "feat(integration): merge E1.1.2D registry-tests into phase1-wave1

Integrates comprehensive test suite
- Adds unit tests for all registry components
- Includes integration test scenarios
- Provides test utilities and mocks
- Test code: 115 lines (not counted toward implementation limit per R007)
- Dependencies: E1.1.2C (registry-helpers)
- Test coverage: 91.0% overall"

# Verify merge success
git log --oneline -1
git diff --stat HEAD~1
```

## Expected Conflicts

Based on the dependency chain and incremental development approach, minimal conflicts are expected:

### Potential Conflict Areas
1. **go.mod/go.sum**: Each effort may have added dependencies
   - **Resolution**: Accept all additions (merge both sides)

2. **Import paths**: If efforts reference each other
   - **Resolution**: Ensure all imports use correct package paths

3. **Test files**: Overlapping test utilities
   - **Resolution**: Keep all tests, rename if necessary

### Conflict Resolution Strategy
```bash
# If conflicts occur during any merge:
# 1. Review conflicts
git status
git diff --name-only --diff-filter=U

# 2. For go.mod/go.sum conflicts:
git checkout --theirs go.mod go.sum
go mod tidy

# 3. For source conflicts:
# Open conflicted files and merge manually
# Ensure all functionality from both sides is preserved

# 4. Complete the merge
git add .
git commit --no-edit

# 5. Verify build still works
go build ./...
go test ./...
```

## Post-Merge Validation

### After All Merges Complete
```bash
# 1. Verify all efforts are integrated
git log --oneline --graph -10

# 2. Check total line count
PROJECT_ROOT="/home/vscode/workspaces/this-is-not-the-target-repo-this-is-for-orchestrator-planning-only"
$PROJECT_ROOT/tools/line-counter.sh

# 3. Run build verification
go build ./...

# 4. Run test suite
go test ./... -v

# 5. Verify no files were lost
git diff origin/main --name-status

# 6. Create integration marker
git commit --allow-empty -m "marker: phase1-wave1 integration complete

All 5 efforts successfully integrated:
- E1.1.1: kind-cert-extraction (450 lines)
- E1.1.2A: registry-types (205 lines)
- E1.1.2B: registry-auth (363 lines)
- E1.1.2C: registry-helpers (684 lines)
- E1.1.2D: registry-tests (115 lines, test-only)

Total implementation: 1,702 lines
Test coverage: 91.0%"
```

## Integration Success Criteria

✅ **All merges complete without unresolved conflicts**
✅ **Build succeeds**: `go build ./...` passes
✅ **Tests pass**: `go test ./...` shows all green
✅ **Line count verified**: Total ~1,702 implementation lines
✅ **No functionality lost**: All features from efforts present
✅ **Clean history**: Merge commits clearly document each integration

## Rollback Plan

If integration fails at any point:

```bash
# Reset to last known good state
git reset --hard HEAD

# Or completely restart from main
git reset --hard origin/main

# Clean any artifacts
git clean -fd

# Retry merge sequence from failed point
```

## Notes for Integration Agent

1. **Execute merges in exact order specified** - Dependencies are critical
2. **Use provided commit messages** - They document the integration history
3. **Stop immediately if merge fails** - Do not attempt to continue
4. **Run validation after each merge** - Ensure incremental success
5. **Document any deviations** - If branches don't exist as expected

## Alternative Approach (if effort repos not accessible)

If the separate effort repositories cannot be accessed as remotes, the Integration Agent should:

1. Clone each effort repository separately
2. Create patches from each:
   ```bash
   cd /path/to/effort
   git format-patch origin/main..HEAD --stdout > effort.patch
   ```
3. Apply patches in order to integration branch:
   ```bash
   git am < effort.patch
   ```

---

**Plan Status**: COMPLETE
**Ready for**: Integration Agent execution
**Created**: 2025-09-18T22:50:00Z
**Review Required**: No (plan only, no code changes)