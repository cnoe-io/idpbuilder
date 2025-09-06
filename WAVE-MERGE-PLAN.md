# Wave 1 Integration Merge Plan

## 🔧 Post-Error Recovery Merge Plan

**Date Created**: 2025-09-06 22:35:00 UTC  
**Plan Type**: Fresh Integration After Duplicate Resolution  
**Target Branch**: `idpbuilder-oci-build-push/phase1/wave1/integration`  
**Integration Directory**: `/home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/integration-workspace`  
**Planner**: Code Reviewer Agent  
**Executor**: Integration Agent (will execute this plan)  

## 📋 Pre-Merge Verification

### Context Summary
- **Previous Issue**: Duplicate declarations between E1.1.1 and E1.1.2 caused build failures
- **Resolution Applied**: 
  - E1.1.1: Renamed `CertValidator` → `KindCertValidator`, `isFeatureEnabled` → `isKindFeatureEnabled`
  - E1.1.2: Renamed `CertValidator` → `RegistryCertValidator`, `isFeatureEnabled` → `isRegistryFeatureEnabled`
- **Current State**: Both efforts have fixes committed and pushed to remote

### Effort Branches to Integrate
1. **E1.1.1 - Kind Certificate Extraction**
   - Branch: `phase1/wave1/effort-kind-cert-extraction`
   - Location: `efforts/phase1/wave1/kind-cert-extraction`
   - Latest Commit: `13f8a4f` - "fix: resolve duplicate declarations and interface issues"
   - Dependencies: None (foundational)
   - Can Parallelize: Yes

2. **E1.1.2 - Registry TLS Trust Integration**  
   - Branch: `phase1/wave1/effort-registry-tls-trust`
   - Location: `efforts/phase1/wave1/registry-tls-trust`
   - Latest Commit: `4f0e259` - "chore: mark duplicate declaration fixes complete"
   - Dependencies: None (can use mock certificates)
   - Can Parallelize: Yes

### Integration Characteristics
- **Merge Order**: Independent (both efforts have no dependencies on each other)
- **Base Commit**: Both efforts branch from `e210954` (same base)
- **Conflict Resolution Applied**: Naming conflicts already resolved in effort branches
- **Feature Flags**: Each effort has independent feature flags

## 🔄 Merge Execution Plan

### Step 1: Prepare Integration Workspace
```bash
# Navigate to integration workspace
cd /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave1/integration-workspace

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