# Phase 1 Wave 1 Integration Merge Plan

## Critical Context (R327 Re-Integration)
- **Integration Branch**: `idpbuilder-oci-build-push/phase1/wave1/integration-20250912-032401`
- **Integration Directory**: `/home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/integration-workspace`
- **Base**: Fresh from main branch (post-R321 fixes)
- **Created**: 2025-09-12 03:24:01
- **Purpose**: Re-integrate Wave 1 with proper R321 fixes applied

## Critical Requirements (R269, R270)
- ✅ Use ONLY original effort branches (no integration branches)
- ✅ Exclude parent 'too-large' branches for split efforts
- ✅ Include only split branches for efforts that were split
- ✅ Determine merge order based on dependencies
- ✅ Document conflict resolution strategies

## Effort Summary

### Wave 1 Efforts
1. **E1.1.1-kind-cert-extraction** [650 lines]
   - Branch: `idpbuilder-oci-build-push/phase1/wave1/kind-cert-extraction`
   - Status: Within limit, use main branch
   - Location: `/efforts/phase1/wave1/kind-cert-extraction`

2. **E1.1.2-registry-tls-trust** [700 lines]
   - Branch: `idpbuilder-oci-build-push/phase1/wave1/registry-tls-trust`
   - Status: Within limit, use main branch
   - Location: `/efforts/phase1/wave1/registry-tls-trust`

3. **E1.1.3-registry-auth-types** [SPLIT]
   - DO NOT USE: Parent branch (too large)
   - Split-001: `idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-001`
   - Split-002: `idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-002`
   - Location: `/efforts/phase1/wave1/registry-auth-types-split-00X`

### Wave 2 Efforts (Part of Wave 1 Completion)
4. **E1.2.1-cert-validation** [SPLIT]
   - DO NOT USE: Parent branch (too large)
   - Split-001: `idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001`
   - Split-002: `idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002`
   - Split-003: `idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003`
   - Location: `/efforts/phase1/wave2/cert-validation-split-00X`

5. **E1.2.2-fallback-strategies** [560 lines]
   - Branch: `idpbuilder-oci-build-push/phase1/wave2/fallback-strategies`
   - Status: Within limit, use main branch
   - Location: `/efforts/phase1/wave2/fallback-strategies`

## Optimal Merge Order

Based on dependency analysis and file modifications:

### Merge Sequence
1. **kind-cert-extraction** (Foundation - Kind cluster certificate extraction)
2. **registry-tls-trust** (Builds on cert extraction - TLS trust management)
3. **registry-auth-types-split-001** (Types and constants)
4. **registry-auth-types-split-002** (Implementation using types)
5. **cert-validation-split-001** (Validation foundations)
6. **cert-validation-split-002** (Validation implementation)
7. **cert-validation-split-003** (Validation completion)
8. **fallback-strategies** (Uses all previous functionality)

## Potential Conflicts Analysis

### Expected Conflicts

1. **pkg/testutil/** (Multiple efforts add test utilities)
   - Affected: kind-cert-extraction, registry-tls-trust, cert-validation splits
   - Resolution: Accept all additions, they should be complementary

2. **pkg/certs/** (Modified by multiple efforts)
   - registry-tls-trust: Adds trust.go, utilities.go
   - cert-validation-split-001: Adds validation_errors.go, storage.go, extractor.go
   - Resolution: These are additive changes, should merge cleanly

3. **go.mod/go.sum** (Dependency additions)
   - Multiple efforts may add dependencies
   - Resolution: Accept all additions, run `go mod tidy` after each merge

4. **pkg/util/** (Shared utilities)
   - kind-cert-extraction modifies pkg/util/env
   - Resolution: Accept all changes, they should be independent

### Low Conflict Risk Areas
- **pkg/kind/**: Only modified by kind-cert-extraction
- **pkg/oci/**: Only modified by registry-auth-types splits
- **pkg/controllers/**: Minimal modifications, should merge cleanly

## Exact Git Commands for Integration Agent

### Prerequisites
```bash
# Navigate to integration workspace
cd /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/integration-workspace

# Verify clean state
git status --short

# Ensure on integration branch
git checkout idpbuilder-oci-build-push/phase1/wave1/integration-20250912-032401

# Fetch all remotes to ensure latest
git fetch --all
```

### Merge Commands (Execute in Order)

#### 1. Merge kind-cert-extraction
```bash
# Add remote for effort (if not exists)
git remote add kind-cert-extraction ../kind-cert-extraction || true

# Fetch the effort
git fetch kind-cert-extraction

# Merge the effort branch
git merge kind-cert-extraction/idpbuilder-oci-build-push/phase1/wave1/kind-cert-extraction \
  --no-ff \
  -m "merge: integrate E1.1.1-kind-cert-extraction (650 lines) into Wave 1 integration"

# If conflicts, resolve and continue
# Expected: No conflicts (first merge)
```

#### 2. Merge registry-tls-trust
```bash
# Add remote for effort
git remote add registry-tls-trust ../registry-tls-trust || true

# Fetch the effort
git fetch registry-tls-trust

# Merge the effort branch
git merge registry-tls-trust/idpbuilder-oci-build-push/phase1/wave1/registry-tls-trust \
  --no-ff \
  -m "merge: integrate E1.1.2-registry-tls-trust (700 lines) into Wave 1 integration"

# If conflicts in pkg/testutil:
# git add pkg/testutil/
# git commit --no-edit
```

#### 3. Merge registry-auth-types-split-001
```bash
# Add remote for split
git remote add registry-auth-types-split-001 ../registry-auth-types-split-001 || true

# Fetch the split
git fetch registry-auth-types-split-001

# Merge the split branch
git merge registry-auth-types-split-001/idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-001 \
  --no-ff \
  -m "merge: integrate E1.1.3-registry-auth-types-split-001 (types/constants) into Wave 1 integration"

# Expected: Clean merge (new pkg/oci directory)
```

#### 4. Merge registry-auth-types-split-002
```bash
# Add remote for split
git remote add registry-auth-types-split-002 ../registry-auth-types-split-002 || true

# Fetch the split
git fetch registry-auth-types-split-002

# Merge the split branch
git merge registry-auth-types-split-002/idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-002 \
  --no-ff \
  -m "merge: integrate E1.1.3-registry-auth-types-split-002 (implementation) into Wave 1 integration"

# Expected: Clean merge (extends pkg/oci)
```

#### 5. Merge cert-validation-split-001
```bash
# Add remote for split
git remote add cert-validation-split-001 ../../wave2/cert-validation-split-001 || true

# Fetch the split
git fetch cert-validation-split-001

# Merge the split branch
git merge cert-validation-split-001/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001 \
  --no-ff \
  -m "merge: integrate E1.2.1-cert-validation-split-001 (validation foundations) into Wave 1 integration"

# If conflicts in pkg/certs:
# Review both versions, likely additive
# git add pkg/certs/
# git commit --no-edit
```

#### 6. Merge cert-validation-split-002
```bash
# Add remote for split
git remote add cert-validation-split-002 ../../wave2/cert-validation-split-002 || true

# Fetch the split
git fetch cert-validation-split-002

# Merge the split branch
git merge cert-validation-split-002/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002 \
  --no-ff \
  -m "merge: integrate E1.2.1-cert-validation-split-002 (validation implementation) into Wave 1 integration"

# Expected: Clean merge or minor conflicts in pkg/certs
```

#### 7. Merge cert-validation-split-003
```bash
# Add remote for split
git remote add cert-validation-split-003 ../../wave2/cert-validation-split-003 || true

# Fetch the split
git fetch cert-validation-split-003

# Merge the split branch
git merge cert-validation-split-003/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003 \
  --no-ff \
  -m "merge: integrate E1.2.1-cert-validation-split-003 (validation completion) into Wave 1 integration"

# Expected: Clean merge
```

#### 8. Merge fallback-strategies
```bash
# Add remote for effort
git remote add fallback-strategies ../../wave2/fallback-strategies || true

# Fetch the effort
git fetch fallback-strategies

# Merge the effort branch
git merge fallback-strategies/idpbuilder-oci-build-push/phase1/wave2/fallback-strategies \
  --no-ff \
  -m "merge: integrate E1.2.2-fallback-strategies (560 lines) into Wave 1 integration"

# Expected: Clean merge (uses previous functionality)
```

### Post-Merge Validation
```bash
# After all merges, validate the integration
go mod tidy
go build ./...
go test ./...

# Verify all efforts integrated
git log --oneline --graph -20

# Check final size
find pkg -name "*.go" | xargs wc -l | tail -1
```

## Conflict Resolution Strategies

### General Resolution Approach
1. **Additive Changes**: Accept both sides when adding new files/functions
2. **Import Conflicts**: Merge all imports, remove duplicates
3. **Test Utilities**: Keep all test helpers from different efforts
4. **go.mod Conflicts**: Accept all dependency additions, run `go mod tidy`

### Specific File Resolution

#### pkg/testutil/* conflicts:
```go
// Accept all test utility additions
// These are helper functions that shouldn't conflict functionally
```

#### pkg/certs/* conflicts:
```go
// registry-tls-trust adds: trust.go, utilities.go
// cert-validation adds: validation_errors.go, storage.go, extractor.go
// These should be additive - accept all
```

#### go.mod conflicts:
```bash
# Accept all require statements
# Then run:
go mod tidy
```

## Validation Checklist

After each merge:
- [ ] No uncommitted changes (`git status`)
- [ ] Build succeeds (`go build ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] No duplicate declarations
- [ ] Dependencies resolved (`go mod tidy`)

After all merges:
- [ ] All 8 efforts integrated
- [ ] Total size within expectations (~4,000 lines)
- [ ] Integration tests pass
- [ ] No missing functionality
- [ ] Clean commit history

## Notes for Integration Agent

1. **DO NOT** merge the parent 'registry-auth-types' branch - only splits
2. **DO NOT** merge the parent 'cert-validation' branch - only splits
3. **ALWAYS** use `--no-ff` to preserve merge history
4. **RESOLVE** conflicts conservatively - when in doubt, accept both
5. **TEST** after each merge to catch issues early
6. **DOCUMENT** any unexpected conflicts in the integration report

## Expected Final State

After successful integration:
- Branch: `idpbuilder-oci-build-push/phase1/wave1/integration-20250912-032401`
- Contains: All Wave 1 and Wave 2 efforts (8 total branches merged)
- Size: Approximately 4,000 lines of new code
- Status: Ready for architect review and main branch merge

---

**Created**: 2025-09-12 03:30:00 UTC
**Created By**: Code Reviewer Agent (WAVE_MERGE_PLANNING state)
**Integration Target**: idpbuilder-oci-build-push/phase1/wave1/integration-20250912-032401