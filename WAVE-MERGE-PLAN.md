# Wave 2 Merge Plan - CASCADE REBASE Operation #2

**Generated:** 2025-09-19T13:12:51Z
**Code Reviewer:** code-reviewer
**State:** WAVE_MERGE_PLANNING
**CASCADE Context:** This is a recreation of the P1W2 integration branch as part of the full project rebase cascade

## Target Integration Branch
- **Branch Name:** idpbuilder-oci-build-push/phase1-wave2-integration
- **Base:** Phase 1 Wave 1 Integration (newly recreated)
- **Location:** /efforts/phase1/wave2/integration-workspace/repo
- **Purpose:** Clean integration of P1W2 efforts onto new P1W1 foundation

## 🎬 Demo Execution Plan (R330/R291 Compliance)

### Demo Requirements Overview
Per R330 and R291, ALL integrations MUST demonstrate working functionality.

### Demo Execution Sequence
1. **After Each Effort Merge:**
   - Run effort-specific demo script if exists
   - Capture output in `demo-results/effort-X-demo.log`
   - Continue even if individual demo fails (document for review)

2. **After All Merges Complete:**
   - Run integrated wave demo: `./wave-demo.sh`
   - Verify all effort features work together
   - Capture evidence in `demo-results/wave-integration-demo.log`

3. **Demo Validation Gates (R291):**
   - ✅ BUILD GATE: Code must compile
   - ✅ TEST GATE: All tests must pass
   - ✅ DEMO GATE: Demo scripts must execute
   - ✅ ARTIFACT GATE: Build outputs must exist

### Demo Files Expected
Based on effort plans, these demos should exist:
- [ ] cert-validation/demo-features.sh (certificate validation demo)
- [ ] fallback-strategies/demo-features.sh (fallback mechanism demo)
- [ ] WAVE-DEMO.md (integration demo documentation)
- [ ] wave-demo.sh (integrated demo script)

### Demo Failure Protocol
If ANY demo fails during integration:
1. Document failure in INTEGRATION_REPORT.md
2. Mark Demo Status: FAILED
3. This will trigger ERROR_RECOVERY per R291
4. Fixes must be made in effort branches (R292)

## Branches to Merge (IN ORDER)

### ⚠️ CRITICAL: Sequential Split Dependencies
The cert-validation effort was split into 3 parts that MUST be merged sequentially.
Each split depends on the previous one being fully integrated.
The fallback-strategies effort depends on ALL cert-validation splits being complete.

### 1. idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001
- **Type:** Split branch (1 of 3)
- **Base:** Built on P1W1 integration foundation
- **Purpose:** Core certificate validation types and interfaces
- **Dependencies:** None (first split)
- **Recent Work:** R321 backport fixes completed
- **Conflicts Expected:** None (first merge)
- **Merge Command:**
  ```bash
  git fetch origin idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001
  git merge origin/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001 --no-ff \
    -m "Integrate cert-validation-split-001 into wave 2 (CASCADE REBASE)"
  ```
- **Post-Merge Validation:**
  ```bash
  # Test compilation
  go build ./...

  # Run unit tests
  go test ./pkg/certvalidation/...

  # Run effort demo if exists
  if [ -f "./cert-validation-split-001/demo-features.sh" ]; then
      echo "🎬 Running cert-validation-split-001 demo..."
      bash "./cert-validation-split-001/demo-features.sh" > "demo-results/cert-validation-split-001-demo.log" 2>&1
      echo "Demo exit code: $?"
  fi
  ```

### 2. idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002
- **Type:** Split branch (2 of 3)
- **Base:** Built on cert-validation-split-001
- **Purpose:** Certificate chain validation and intermediate CA handling
- **Dependencies:** MUST merge after split-001
- **Recent Work:** R321 test fixture additions, test setup fixes
- **Conflicts Expected:** None (sequential split, clean merge expected)
- **Merge Command:**
  ```bash
  git fetch origin idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002
  git merge origin/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002 --no-ff \
    -m "Integrate cert-validation-split-002 into wave 2 (CASCADE REBASE)"
  ```
- **Post-Merge Validation:**
  ```bash
  # Test compilation
  go build ./...

  # Run unit tests (should now include chain validation tests)
  go test ./pkg/certvalidation/...

  # Run effort demo if exists
  if [ -f "./cert-validation-split-002/demo-features.sh" ]; then
      echo "🎬 Running cert-validation-split-002 demo..."
      bash "./cert-validation-split-002/demo-features.sh" > "demo-results/cert-validation-split-002-demo.log" 2>&1
      echo "Demo exit code: $?"
  fi
  ```

### 3. idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003
- **Type:** Split branch (3 of 3)
- **Base:** Built on cert-validation-split-002
- **Purpose:** Advanced validation features, custom validators, integration points
- **Dependencies:** MUST merge after split-002
- **Recent Work:** R321 backport fixes, chain validator syntax fixes, Bug #4 resolution
- **Conflicts Expected:** None (sequential split, clean merge expected)
- **Merge Command:**
  ```bash
  git fetch origin idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003
  git merge origin/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003 --no-ff \
    -m "Integrate cert-validation-split-003 into wave 2 (CASCADE REBASE)"
  ```
- **Post-Merge Validation:**
  ```bash
  # Test compilation
  go build ./...

  # Run full cert validation test suite
  go test ./pkg/certvalidation/...

  # Run effort demo if exists
  if [ -f "./cert-validation-split-003/demo-features.sh" ]; then
      echo "🎬 Running cert-validation-split-003 demo..."
      bash "./cert-validation-split-003/demo-features.sh" > "demo-results/cert-validation-split-003-demo.log" 2>&1
      echo "Demo exit code: $?"
  fi
  ```

### 4. idpbuilder-oci-build-push/phase1/wave2/fallback-strategies
- **Type:** Original effort branch (within size limit)
- **Base:** Built on cert-validation-split-003
- **Purpose:** Fallback mechanisms for certificate validation failures
- **Dependencies:** MUST merge after ALL cert-validation splits (depends on complete validation framework)
- **Recent Work:** R321 fallback strategy backport analysis complete
- **Conflicts Expected:** Minimal (depends on cert-validation but in separate package)
- **Merge Command:**
  ```bash
  git fetch origin idpbuilder-oci-build-push/phase1/wave2/fallback-strategies
  git merge origin/idpbuilder-oci-build-push/phase1/wave2/fallback-strategies --no-ff \
    -m "Integrate fallback-strategies into wave 2 (CASCADE REBASE)"
  ```
- **Post-Merge Validation:**
  ```bash
  # Test compilation
  go build ./...

  # Run fallback strategy tests
  go test ./pkg/fallback/...

  # Run integrated cert validation with fallback tests
  go test ./pkg/certvalidation/... ./pkg/fallback/...

  # Run effort demo if exists
  if [ -f "./fallback-strategies/demo-features.sh" ]; then
      echo "🎬 Running fallback-strategies demo..."
      bash "./fallback-strategies/demo-features.sh" > "demo-results/fallback-strategies-demo.log" 2>&1
      echo "Demo exit code: $?"
  fi
  ```

## Excluded Branches (DO NOT MERGE)
These branches should NOT be merged:
- ❌ idpbuilder-oci-build-push/phase1/wave2-integration (old integration branch - NEVER merge from integration branches per R270)
- ❌ Any "cert-validation" branch without "-split" suffix (original too-large branch if it existed)

## Merge Strategy
1. **Merge Type:** --no-ff (preserve branch history for CASCADE tracking)
2. **Conflict Resolution:** Unlikely due to sequential nature, but if conflicts occur:
   - Preserve the newer implementation (from the branch being merged)
   - Ensure test fixtures are not duplicated
   - Validate no functionality is lost
3. **Testing:** Run tests after EACH merge to catch issues early
4. **CASCADE Tracking:** Each merge commit message includes "(CASCADE REBASE)" marker

## Expected Conflicts
Based on the CASCADE rebase analysis and sequential split structure:
- **Low probability** of conflicts due to:
  - Sequential split development (each builds cleanly on previous)
  - Clear separation of concerns between splits
  - Fallback strategies in separate package
- **Potential minor conflicts:**
  - Test helper functions (if any were backported differently)
  - Import statements (ensure correct package references)

## Integration Agent Instructions
1. **CRITICAL**: CD to integration directory before starting:
   ```bash
   cd /efforts/phase1/wave2/integration-workspace/repo
   pwd  # Verify: Should show integration workspace path
   ```
2. **Execute merges in the EXACT order specified** (splits are dependent!)
3. **Run tests after EACH merge** (don't wait until the end)
4. **Document any unexpected conflicts** in work-log.md
5. **Create demo-results directory** for demo outputs
6. **Generate INTEGRATION-REPORT.md** when complete with:
   - List of all merged branches
   - Test results after each merge
   - Demo execution results
   - Final validation status
   - CASCADE tracking information

## Validation Steps

### After Each Individual Merge:
```bash
# 1. Verify compilation
go build ./...
if [ $? -ne 0 ]; then
    echo "❌ BUILD FAILED after merging $BRANCH_NAME"
    # Document in work-log and continue to identify all issues
fi

# 2. Run unit tests for affected packages
go test ./... -short
if [ $? -ne 0 ]; then
    echo "⚠️ TESTS FAILED after merging $BRANCH_NAME"
    # Document failures but continue merging
fi

# 3. Check for demo scripts (R330 compliance)
# (See individual merge sections for demo commands)
```

### After All Merges Complete:
```bash
# 1. Final compilation check
echo "🔨 Final build validation..."
go build ./...

# 2. Run full test suite
echo "🧪 Running complete test suite..."
go test ./... -v

# 3. Check size compliance
echo "📏 Verifying size compliance..."
PROJECT_ROOT=$(pwd)
while [ "$PROJECT_ROOT" != "/" ]; do
    [ -f "$PROJECT_ROOT/orchestrator-state.json" ] && break
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
done
$PROJECT_ROOT/tools/line-counter.sh

# 4. Run integrated wave demo (R291 requirement)
echo "🎬 Running integrated wave demo..."
if [ -f "./wave-demo.sh" ]; then
    bash ./wave-demo.sh > demo-results/wave-integration-demo.log 2>&1
    DEMO_STATUS=$?
    echo "Wave demo status: $DEMO_STATUS"
else
    echo "⚠️ WARNING: No wave-demo.sh found"
    echo "Integration Agent should create basic demo showing:"
    echo "  - Certificate validation working"
    echo "  - Fallback strategies triggering correctly"
fi

# 5. Verify all efforts are integrated
echo "✅ Verifying effort integration..."
for effort in cert-validation-split-001 cert-validation-split-002 cert-validation-split-003 fallback-strategies; do
    if git log --oneline | grep -q "$effort"; then
        echo "✅ $effort integrated"
    else
        echo "❌ MISSING: $effort not found in integration!"
    fi
done
```

## Risk Assessment
- **Low Risk:** Sequential splits developed to work together
- **Low Risk:** Clear dependency chain minimizes conflicts
- **Medium Risk:** R321 backport fixes may have subtle interactions
- **Mitigation:** Comprehensive testing after each merge
- **CASCADE Benefit:** Clean rebase from new P1W1 foundation reduces technical debt

## Success Criteria
1. ✅ All 4 effort branches merged in correct order
2. ✅ No integration branches used as sources (R270 compliance)
3. ✅ Code compiles after each merge
4. ✅ Tests pass (or failures documented)
5. ✅ Demos executed and results captured (R330/R291)
6. ✅ CASCADE markers in all merge commits
7. ✅ Integration report generated
8. ✅ Total size within limits

## CASCADE-Specific Notes
This merge plan is part of CASCADE REBASE Operation #2, recreating the P1W2 integration branch on top of the newly recreated P1W1 foundation. The sequential nature of the cert-validation splits and the dependency of fallback-strategies on the complete cert-validation framework makes the merge order critical for success.

---

**Integration Agent:** Follow this plan exactly. The success of the CASCADE rebase depends on proper integration of these P1W2 efforts.