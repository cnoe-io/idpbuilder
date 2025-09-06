# Wave 2 Rebase Plan - Image Builder Branch

## Critical Context (Phase 2 Wave 1 Rebase)
- **Source Branch**: `idpbuilder-oci-build-push/phase2/wave1/image-builder`
- **Target Base**: `origin/idpbuilder-oci-build-push/phase1/integration` (commit: 2c39501)
- **Purpose**: Update image-builder to latest Phase 1 integration foundation
- **Date**: 2025-09-14

## Rebase Purpose
This rebase updates the image-builder branch from an old phase1/integration commit to the latest complete Phase 1 integration. The latest base includes:
- All Wave 1 work (kind-cert-extraction, registry-tls-trust, registry-auth-types)
- All Wave 2 work (cert-validation, fallback-strategies)
- Complete Phase 1 foundation for Phase 2 efforts

## Phase 2 Wave 1 Context
The image-builder effort is part of Phase 2 Wave 1 and should be based on the complete Phase 1 work.
```bash
# Navigate to integration workspace
cd /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/integration-workspace

<<<<<<< HEAD
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
=======
# Verify we're on the integration branch
git branch --show-current
# Expected: idpbuilder-oci-build-push/phase1/wave1/integration

# Ensure clean working directory
git status
# Expected: nothing to commit, working tree clean

# Fetch latest changes
git fetch origin
```

### Step 2: Add Effort Remotes (if not already added)
```bash
# Add remotes for the effort branches
git remote add kind-cert ../kind-cert-extraction || true
git remote add registry-tls ../registry-tls-trust || true

# Fetch from the effort remotes
git fetch kind-cert
git fetch registry-tls
```

### Step 3: Merge E1.1.1 - Kind Certificate Extraction
```bash
# Merge E1.1.1 into integration
git merge origin/phase1/wave1/effort-kind-cert-extraction --no-ff \
  -m "feat: integrate E1.1.1 - Kind Certificate Extraction

- Adds certificate extraction from Kind clusters
- Implements KindCertValidator interface
- Provides kubectl-based client for pod access
- Feature flag: KIND_CERT_EXTRACTION_ENABLED"

# Verify no conflicts
if [ $? -ne 0 ]; then
    echo "❌ Merge conflict detected - manual resolution required"
    exit 1
fi

# Quick build test
go build ./... || echo "⚠️ Build issue detected - review required"
```

### Step 4: Validate E1.1.1 Integration
```bash
# Run tests for E1.1.1 functionality
go test ./pkg/certs/... -v

# Verify renamed functions are present
grep -r "KindCertValidator" pkg/
grep -r "isKindFeatureEnabled" pkg/

# Check for any remaining duplicates
grep -r "^type CertValidator interface" pkg/ | wc -l
# Expected: 1 (only KindCertValidator should exist at this point)
```

### Step 5: Merge E1.1.2 - Registry TLS Trust Integration
```bash
# Merge E1.1.2 into integration
git merge origin/phase1/wave1/effort-registry-tls-trust --no-ff \
  -m "feat: integrate E1.1.2 - Registry TLS Trust Integration

- Adds TLS trust store management for registries
- Implements RegistryCertValidator interface  
- Provides go-containerregistry transport configuration
- Feature flag: REGISTRY_TLS_TRUST_ENABLED"

# Verify no conflicts
if [ $? -ne 0 ]; then
    echo "❌ Merge conflict detected - manual resolution required"
    exit 1
fi
```

### Step 6: Final Integration Validation
```bash
# Full build test
echo "🔨 Running full build..."
go build ./...

# Run all tests
echo "🧪 Running all tests..."
go test ./... -v

# Verify both renamed interfaces exist
echo "✅ Verifying interface naming..."
grep -r "KindCertValidator" pkg/ | head -2
grep -r "RegistryCertValidator" pkg/ | head -2

# Verify both renamed functions exist
echo "✅ Verifying function naming..."
grep -r "isKindFeatureEnabled" pkg/ | head -2
grep -r "isRegistryFeatureEnabled" pkg/ | head -2

# Check for any duplicate declarations
echo "🔍 Checking for duplicates..."
DUPLICATES=$(grep -r "^type CertValidator interface\|^func isFeatureEnabled" pkg/ | wc -l)
if [ $DUPLICATES -gt 0 ]; then
    echo "❌ Found duplicate declarations!"
    grep -r "^type CertValidator interface\|^func isFeatureEnabled" pkg/
    exit 1
else
    echo "✅ No duplicate declarations found"
fi

# Final line count
echo "📏 Total integration size:"
git diff --stat origin/main
```

### Step 7: Push Integration Branch
```bash
# Commit any integration-specific files if needed
git add -A
git commit -m "chore: wave 1 integration complete" || true

# Push the integration branch
git push origin idpbuilder-oci-build-push/phase1/wave1/integration
```

## ✅ Success Criteria

The integration is successful when:
1. ✅ Both effort branches merged without conflicts
2. ✅ No duplicate type/function declarations
3. ✅ All tests pass
4. ✅ Build succeeds without errors
5. ✅ Both feature flags can be toggled independently
6. ✅ Integration branch pushed to remote

## ⚠️ Rollback Plan

If integration fails:
```bash
# Reset to pre-merge state
git reset --hard origin/idpbuilder-oci-build-push/phase1/wave1/integration

# Investigate specific failure
# - Check for missed duplicate declarations
# - Review test failures
# - Examine build errors

# Report issues back to orchestrator for ERROR_RECOVERY
```

## 📊 Expected Outcomes

### File Structure After Integration
```
pkg/certs/
├── extractor.go         # E1.1.1 - KindCertValidator
├── extractor_test.go    # E1.1.1 tests
├── helpers.go           # E1.1.1 - isKindFeatureEnabled
├── helpers_test.go      # E1.1.1 tests
├── kubectl_client.go    # E1.1.1
├── storage.go           # E1.1.1
├── trust.go            # E1.1.2 - isRegistryFeatureEnabled
├── utilities.go        # E1.1.2 - RegistryCertValidator
├── transport.go        # E1.1.2
└── transport_test.go   # E1.1.2 tests
```

### Interface Summary
- `KindCertValidator` - Interface for Kind certificate validation (E1.1.1)
- `RegistryCertValidator` - Interface for registry certificate validation (E1.1.2)
- No generic `CertValidator` interface should exist

### Function Summary
- `isKindFeatureEnabled()` - Check Kind feature flag (E1.1.1)
- `isRegistryFeatureEnabled()` - Check registry feature flag (E1.1.2)
- No generic `isFeatureEnabled()` function should exist

## 🔍 Integration Verification Checklist

- [ ] Integration workspace is clean
- [ ] Both effort branches fetched
- [ ] E1.1.1 merged successfully
- [ ] E1.1.1 tests pass
- [ ] E1.1.2 merged successfully
- [ ] E1.1.2 tests pass
- [ ] No duplicate declarations
- [ ] Full build succeeds
- [ ] All tests pass
- [ ] Integration branch pushed

## 📝 Notes for Integration Agent

1. **Execute steps sequentially** - Do not parallelize merge operations
2. **Stop on first failure** - Do not continue if any step fails
3. **Document any issues** - Create ERROR-REPORT.md if problems occur
4. **Verify each step** - Run validation commands after each merge
5. **Use --no-ff** - Preserve merge commit history for tracking

---

**Plan Status**: READY FOR EXECUTION  
**Created By**: Code Reviewer Agent  
**For Execution By**: Integration Agent  
>>>>>>> dccee8f (docs: add integration plan and work log)
