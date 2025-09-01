# SPLIT-PLAN-001: Core Detection and Handling

## Split 001 of 2: Certificate Error Detection and Core Handler
**Planner**: Code Reviewer $$
**Parent Effort**: fallback-strategies
**Created**: 2025-08-31 20:38:00

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Boundaries (⚠️⚠️⚠️ CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: None (first split of THIS effort)
  - Path: N/A (this is Split 001)
  - Branch: N/A
- **This Split**: Split 001 of phase1/wave2/fallback-strategies
  - Path: efforts/phase1/wave2/fallback-strategies/split-001/
  - Branch: idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-001
- **Next Split**: Split 002 of phase1/wave2/fallback-strategies
  - Path: efforts/phase1/wave2/fallback-strategies/split-002/
  - Branch: idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-002
- **File Boundaries**:
  - This Split Start: New implementation (no previous lines)
  - This Split End: Complete detector.go and handler.go with basic tests
  - Next Split Start: logger.go (new file)

## Files in This Split (EXCLUSIVE - no overlap with other splits)
| File | Lines | Purpose |
|------|-------|---------|
| pkg/certs/fallback/detector.go | 298 | Certificate error detection and classification |
| pkg/certs/fallback/handler.go | 220 | Main fallback handler implementation |
| pkg/certs/fallback/detector_test.go | ~200 | Basic unit tests for detector (partial) |
| **Total Estimated** | ~718 | Well under 800 limit |

## Functionality
This split implements the core certificate error detection and handling infrastructure:

1. **Error Detection** (detector.go):
   - CertErrorType enumeration
   - CertErrorDetector interface
   - Error classification logic
   - Certificate chain validation
   - Error detail extraction

2. **Fallback Handler** (handler.go):
   - FallbackHandler interface
   - FallbackStrategy structure
   - Action determination logic
   - Basic insecure mode configuration
   - Interface for security logging (implementation in Split 002)

3. **Basic Testing** (detector_test.go - partial):
   - Unit tests for error type detection
   - Tests for error classification
   - Basic validation tests

## Dependencies
- **External Dependencies**:
  - crypto/tls
  - crypto/x509
  - Standard Go libraries

- **From Previous Efforts**:
  - E1.1.1 (kind-certificate-extraction): Import certificate extraction interfaces
  - E1.1.2 (registry-tls-trust-integration): Import TrustStoreManager interface

- **Creates for Next Split**:
  - CertErrorDetector interface
  - FallbackHandler interface
  - Core types and constants

## Implementation Instructions

### Step 1: Create Split Infrastructure
```bash
# Create split directory structure
mkdir -p efforts/phase1/wave2/fallback-strategies/split-001/pkg/certs/fallback

# Create split branch
git checkout -b idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-001
```

### Step 2: Implement Core Detection (detector.go)
1. Copy detector.go from main implementation
2. Ensure all imports are satisfied
3. Verify compilation with `go build`
4. File should include:
   - CertErrorType constants
   - ErrorDetails structure
   - CertErrorDetector interface and implementation
   - Error classification methods

### Step 3: Implement Fallback Handler (handler.go)
1. Copy handler.go from main implementation
2. Ensure interfaces are complete
3. File should include:
   - FallbackHandler interface
   - FallbackStrategy structure
   - FallbackAction enumeration
   - Core handler implementation
   - NOTE: LogSecurityDecision will call stub (actual logger in Split 002)

### Step 4: Create Basic Tests (detector_test.go - partial)
1. Implement essential unit tests only (~200 lines)
2. Focus on:
   - Error type detection tests
   - Classification accuracy tests
   - Edge case handling
3. Leave comprehensive testing for Split 002

### Step 5: Validate Split
```bash
# Measure size with line counter
$PROJECT_ROOT/tools/line-counter.sh

# Verify under 800 lines
# Run tests
go test ./pkg/certs/fallback/...

# Ensure compilation
go build ./...
```

## Split Branch Strategy
- Branch: `idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-001`
- Base: `main`
- Must merge to: `idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies` after review
- Sequential dependency: Split 002 depends on this split

## Quality Requirements
- [ ] Size measured and confirmed <800 lines
- [ ] All interfaces properly defined
- [ ] Basic tests passing
- [ ] No compilation errors
- [ ] Comments and documentation included
- [ ] No hardcoded values
- [ ] Thread-safe implementations

## Testing Requirements
- [ ] Error detection tests pass
- [ ] Handler interface tests pass
- [ ] No nil pointer exceptions
- [ ] Proper error handling
- [ ] At least 60% code coverage for this split

## Known Limitations
This split provides core functionality but:
- Logging implementation is stubbed (completed in Split 002)
- Recommendation generation not included (Split 002)
- Full test coverage deferred to Split 002
- Integration tests in Split 002

## Integration Notes
- This split MUST be completed and merged before Split 002
- Split 002 will import these interfaces
- No breaking changes allowed after merge
- Maintain backward compatibility

---

**Status**: Ready for Implementation
**Dependencies Verified**: Yes
**Size Validated**: ~718 lines (under limit)
**Next Step**: Implement Split 001, then proceed to Split 002