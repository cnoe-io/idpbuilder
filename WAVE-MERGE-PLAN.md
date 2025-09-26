# WAVE MERGE PLAN - Phase 1 Wave 2

## Overview
**Integration Branch**: `phase1-wave2-integration`
**Base Branch**: `phase1-wave1-integration`
**Integration Directory**: `/home/vscode/workspaces/idpbuilder-gitea-push/efforts/phase1/wave2/integration-workspace`
**Total Efforts**: 2
**Total Lines**: 115 (9 + 106)
**Created**: 2025-09-26T08:05:00Z
**Planner**: Code Reviewer Agent

## Integration Strategy
**Merge Order**: Sequential (dependency-based)
- effort-1.2.1 provides test fixtures/helpers foundation
- effort-1.2.2 builds on test fixtures with command testing framework

## Pre-Merge Verification Commands
```bash
# Verify we're in the integration workspace
cd /home/vscode/workspaces/idpbuilder-gitea-push/efforts/phase1/wave2/integration-workspace
pwd  # Should show: .../efforts/phase1/wave2/integration-workspace

# Verify we're on the correct integration branch
git branch --show-current  # Should show: phase1-wave2-integration

# Verify base is Wave 1 integration
git log --oneline -1  # Should show Wave 1 integration as base

# Fetch latest from all remotes
git fetch --all

# Verify effort branches exist
git branch -r | grep "effort-1.2"
```

## Branches to Merge (IN ORDER)

### 1. Branch: `igp/phase1/wave2/effort-1.2.1-test-fixtures-setup`
**Effort**: Test Fixtures Setup
**Size**: 9 lines
**Source Directory**: `efforts/phase1/wave2/effort-1.2.1-test-fixtures-setup`
**Status**: ACCEPTED (reviewed 2025-09-26T06:33:28Z)

#### Files Added/Modified:
- `test/fixtures/auth/credentials.yaml` - Authentication test fixtures
- `test/helpers.go` - Test helper functions
- `test/helpers_test.go` - Helper function tests

#### Merge Command:
```bash
git merge origin/igp/phase1/wave2/effort-1.2.1-test-fixtures-setup --no-ff \
  -m "feat(integration): Merge effort-1.2.1-test-fixtures-setup into phase1-wave2-integration

- Added test fixtures for authentication scenarios
- Implemented test helper functions for Wave 2 testing
- Foundation for command testing framework

Size: 9 lines"
```

#### Expected Conflicts:
- **None expected** - This is foundational test infrastructure not present in Wave 1

#### Post-Merge Validation:
```bash
# Verify test fixtures exist
ls -la test/fixtures/auth/
test -f test/helpers.go && echo "✅ Helpers present" || echo "❌ Missing helpers"

# Run basic tests to verify no breakage
go test ./test/... -v
```

---

### 2. Branch: `igp/phase1/wave2/effort-1.2.2-command-testing-framework`
**Effort**: Command Testing Framework
**Size**: 106 lines
**Source Directory**: `efforts/phase1/wave2/effort-1.2.2-command-testing-framework`
**Status**: ACCEPTED (reviewed 2025-09-26T07:02:30Z)
**Dependencies**: Requires effort-1.2.1 test fixtures

#### Files Added/Modified:
- `pkg/cmd/push/root.go` - Push command implementation (from Wave 1)
- `pkg/cmd/push/push_test.go` - Unit tests for push command
- `test/integration/push_integration_test.go` - Integration tests
- `test/integration/suite_test.go` - Test suite setup

#### Merge Command:
```bash
git merge origin/igp/phase1/wave2/effort-1.2.2-command-testing-framework --no-ff \
  -m "feat(integration): Merge effort-1.2.2-command-testing-framework into phase1-wave2-integration

- Added comprehensive unit tests for push command
- Implemented integration test suite
- Builds on test fixtures from effort-1.2.1
- Complete test coverage for Wave 1 functionality

Size: 106 lines"
```

#### Expected Conflicts:
- **Possible conflict** in `pkg/cmd/push/root.go` if both efforts modified the push command
- **Possible overlap** in test organization if both added similar test structures

#### Conflict Resolution Strategy:
```bash
# If conflicts in pkg/cmd/push/root.go:
# - Keep the more comprehensive implementation
# - Ensure all flags from Wave 1 are preserved
# - Verify authentication and TLS settings are intact

# If test conflicts:
# - Merge test functions, avoiding duplicates
# - Ensure test fixtures from 1.2.1 are used by 1.2.2 tests
# - Maintain proper test suite organization
```

#### Post-Merge Validation:
```bash
# Verify all test files present
ls -la pkg/cmd/push/push_test.go
ls -la test/integration/

# Run unit tests
go test ./pkg/cmd/push/... -v

# Run integration tests
go test ./test/integration/... -v

# Verify no compilation errors
go build ./...
```

## Final Integration Validation

After completing ALL merges, run these validation steps:

```bash
# 1. Verify branch state
git status  # Should be clean after all merges
git log --oneline -5  # Should show both merge commits

# 2. Full test suite
go test ./... -v

# 3. Build verification
go build -o idpbuilder ./main.go
./idpbuilder push --help  # Verify command works

# 4. Line count verification
# Use project line counter to verify total
$PROJECT_ROOT/tools/line-counter.sh

# 5. Check for any missed files
git diff --name-only phase1-wave1-integration..HEAD

# 6. Final commit (if needed)
git add -A
git commit -m "chore: finalize Wave 2 integration - 2 efforts, 115 lines total"
git push origin phase1-wave2-integration
```

## Post-Integration Checklist

- [ ] Both efforts merged successfully
- [ ] No unresolved conflicts
- [ ] All tests passing
- [ ] Build successful
- [ ] Total line count verified (115 lines)
- [ ] Integration branch pushed to remote
- [ ] Ready for architect review

## Notes for Orchestrator

1. **Dependencies Respected**: effort-1.2.1 MUST be merged before effort-1.2.2
2. **Incremental Build**: Wave 2 builds on Wave 1 integration successfully
3. **Test Coverage**: Wave 2 adds comprehensive testing for Wave 1 functionality
4. **No Breaking Changes**: All Wave 1 functionality preserved
5. **Clean Integration**: Expected minimal to no conflicts

## Risk Assessment

**Low Risk Integration**:
- Small, focused efforts (9 + 106 = 115 lines total)
- Clear separation of concerns (test fixtures vs test implementation)
- Dependencies well-defined
- Building on stable Wave 1 base

**Potential Issues**:
- If Wave 1 integration modified test structure, may need adjustment
- Ensure test helper naming doesn't conflict with existing helpers

## Success Criteria

✅ Both efforts merged without data loss
✅ All Wave 1 functionality still works
✅ New test suite executes successfully
✅ Integration branch ready for Wave 3 development
✅ Total implementation under 800 lines (115 lines ✓)

---

**Document Status**: COMPLETE
**Ready for Execution**: YES
**Executor**: Orchestrator Agent (NOT Code Reviewer)