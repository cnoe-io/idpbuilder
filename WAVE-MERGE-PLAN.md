# WAVE MERGE PLAN - Phase 1 Wave 1 Integration

## Integration Context
- **Phase**: 1
- **Wave**: 1  
- **Base Branch**: main
- **Integration Branch**: phase1/wave1/integration
- **Created**: 2025-09-11 (Re-created after R327 mandatory re-integration)
- **Reason**: Re-integration after fixes applied to source branches

## Efforts to Merge

### Effort Branches (in dependency order):
1. **E1.1.1**: idpbuilder-oci-build-push/phase1/wave1/kind-cert-extraction
   - Status: PASSED
   - Lines: 650
   - No dependencies

2. **E1.1.2**: idpbuilder-oci-build-push/phase1/wave1/registry-tls-trust
   - Status: PASSED (with recent fixes applied)
   - Lines: 700
   - Fixed: Duplicate removals applied during ERROR_RECOVERY

3. **E1.1.3-SPLIT-001**: idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-001
   - Status: PASSED (split from original effort)
   - Lines: 800
   - First part of registry-auth-types implementation

4. **E1.1.3-SPLIT-002**: idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-002
   - Status: PASSED (with recent fixes applied)
   - Lines: 800
   - Fixed: TLSConfig consolidation applied during ERROR_RECOVERY
   - Second part of registry-auth-types implementation

## Merge Instructions

### Step 1: Verify Current Branch
```bash
# Ensure we're on the integration branch
git branch --show-current
# Expected: phase1/wave1/integration
```

### Step 2: Fetch Latest Updates
```bash
# Get latest from all branches
git fetch origin
```

### Step 3: Merge E1.1.1 - kind-cert-extraction
```bash
echo "=== Merging E1.1.1: kind-cert-extraction ==="
git merge origin/idpbuilder-oci-build-push/phase1/wave1/kind-cert-extraction --no-edit
```

### Step 4: Merge E1.1.2 - registry-tls-trust (with fixes)
```bash
echo "=== Merging E1.1.2: registry-tls-trust (includes fixes) ==="
git merge origin/idpbuilder-oci-build-push/phase1/wave1/registry-tls-trust --no-edit
```

### Step 5: Merge E1.1.3-SPLIT-001 - registry-auth-types (part 1)
```bash
echo "=== Merging E1.1.3-SPLIT-001: registry-auth-types part 1 ==="
git merge origin/idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-001 --no-edit
```

### Step 6: Merge E1.1.3-SPLIT-002 - registry-auth-types (part 2 with fixes)
```bash
echo "=== Merging E1.1.3-SPLIT-002: registry-auth-types part 2 (includes fixes) ==="
git merge origin/idpbuilder-oci-build-push/phase1/wave1/registry-auth-types-split-002 --no-edit
```

### Step 7: Run Tests
```bash
echo "=== Running tests after all merges ==="
make test
```

### Step 8: Verify Build
```bash
echo "=== Verifying build ==="
make build
```

## Conflict Resolution Strategy

If conflicts occur:
1. Prioritize code from the later split (SPLIT-002) as it has the latest fixes
2. Ensure TLSConfig is properly consolidated (fixed in SPLIT-002)
3. Remove any duplicates that might have been reintroduced
4. Document resolution in work-log.md

## Success Criteria
- [ ] All 4 branches merged successfully
- [ ] No merge conflicts OR conflicts resolved properly
- [ ] Tests pass after all merges
- [ ] Build succeeds
- [ ] No duplicate definitions
- [ ] TLSConfig properly consolidated

## Important Notes
- This is a RE-INTEGRATION per R327 after fixes were applied
- The previous integration had build failures that were fixed in source branches
- Pay special attention to the fixes:
  - registry-auth-types-split-002: TLSConfig consolidation
  - registry-tls-trust: Duplicate removals