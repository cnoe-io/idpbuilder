# SPLIT-PLAN-002: Support Systems and Complete Testing

## Split 002 of 2: Logging, Recommendations, and Full Test Coverage
**Planner**: Code Reviewer $$
**Parent Effort**: fallback-strategies
**Created**: 2025-08-31 20:38:30

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Boundaries (⚠️⚠️⚠️ CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: Split 001 of phase1/wave2/fallback-strategies
  - Path: efforts/phase1/wave2/fallback-strategies/split-001/
  - Branch: idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-001
  - Summary: Implemented core detection and handler interfaces
- **This Split**: Split 002 of phase1/wave2/fallback-strategies
  - Path: efforts/phase1/wave2/fallback-strategies/split-002/
  - Branch: idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-002
- **Next Split**: None (final split of phase1/wave2/fallback-strategies)
  - Path: N/A
  - Branch: N/A

⚠️ NEVER reference splits from different efforts!
✅ RIGHT: "Previous Split: Split 001 of phase1/wave2/fallback-strategies"
❌ WRONG: "Previous Split: Split 001 (certificate-validation-pipeline)"

## Files in This Split (EXCLUSIVE - no overlap except test completion)
| File | Lines | Purpose |
|------|-------|---------|
| pkg/certs/fallback/logger.go | 232 | Security decision logging implementation |
| pkg/certs/fallback/recommendations.go | 344 | User-friendly recommendation generation |
| pkg/certs/fallback/handler_test.go | 395 | Complete handler test suite |
| pkg/certs/fallback/detector_test.go | +285 | Complete detector tests (adds to Split 001's basic tests) |
| **Total Estimated** | ~756 | Under 800 limit |

## Functionality
This split completes the fallback strategies implementation with support systems:

1. **Security Logging** (logger.go):
   - SecurityLogEntry structure
   - Structured audit logging
   - Security severity levels
   - Persistent log management
   - Compliance-ready audit trails

2. **Recommendation Engine** (recommendations.go):
   - User-friendly error messages
   - Actionable remediation steps
   - Registry-specific recommendations
   - Certificate diagnostic helpers
   - Command suggestions for fixes

3. **Complete Test Coverage**:
   - Full handler test suite (handler_test.go)
   - Complete detector tests (detector_test.go - remaining tests)
   - Integration test scenarios
   - Edge case coverage
   - Error simulation tests

## Dependencies
- **From Split 001** (MUST be completed first):
  - CertErrorDetector interface
  - FallbackHandler interface
  - CertErrorType enumeration
  - ErrorDetails structure
  - FallbackStrategy structure

- **External Dependencies**:
  - encoding/json (for structured logging)
  - Standard testing libraries
  - Test fixture certificates

- **Provides to Phase 2**:
  - Complete fallback strategy system
  - Production-ready error handling
  - Audit-compliant logging

## Implementation Instructions

### Step 1: Verify Split 001 Completion
```bash
# Ensure Split 001 is merged
git checkout idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies
git pull

# Create split-002 branch
git checkout -b idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-002

# Create split directory
mkdir -p efforts/phase1/wave2/fallback-strategies/split-002/pkg/certs/fallback
```

### Step 2: Import Split 001 Interfaces
```go
// Ensure you can import from Split 001
import (
    "your-module/pkg/certs/fallback" // From Split 001
)
```

### Step 3: Implement Security Logger (logger.go)
1. Copy logger.go from main implementation
2. Implement full logging functionality:
   - Structured log entries
   - JSON formatting
   - File persistence
   - Rotation policies
   - Security audit compliance

### Step 4: Implement Recommendations (recommendations.go)
1. Copy recommendations.go from main implementation
2. Include:
   - Error-to-recommendation mapping
   - Registry-specific suggestions
   - Command generation for fixes
   - Diagnostic information formatting

### Step 5: Complete Test Coverage
1. **handler_test.go** (395 lines):
   - Full handler functionality tests
   - Insecure mode tests
   - Strategy selection tests
   - Integration scenarios

2. **detector_test.go** (add ~285 lines):
   - Complete the basic tests from Split 001
   - Add comprehensive error scenarios
   - Test all error types
   - Edge cases and malformed certificates

### Step 6: Validate Split
```bash
# Measure size
$PROJECT_ROOT/tools/line-counter.sh

# Verify under 800 lines
# Run all tests
go test -v ./pkg/certs/fallback/...

# Check coverage
go test -cover ./pkg/certs/fallback/...

# Ensure >80% coverage
```

## Split Branch Strategy
- Branch: `idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies-split-002`
- Base: `idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies` (after Split 001 merged)
- Must merge to: `idpbuidler-oci-go-cr/phase1/wave2/fallback-strategies` after review
- Final merge: Both splits combine to complete the effort

## Quality Requirements
- [ ] Size measured and confirmed <800 lines
- [ ] All tests passing (including Split 001 tests)
- [ ] Test coverage >80% for complete package
- [ ] No compilation errors
- [ ] Security logging validated
- [ ] Recommendations are actionable
- [ ] No sensitive data in logs

## Testing Requirements
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Coverage >80% overall
- [ ] No race conditions
- [ ] Logger handles concurrent writes
- [ ] Recommendations are accurate
- [ ] Error scenarios fully covered

## Integration Checklist
- [ ] Split 001 interfaces imported correctly
- [ ] No duplicate code from Split 001
- [ ] Combined splits equal original implementation
- [ ] All original functionality preserved
- [ ] Tests cover both splits' code
- [ ] Documentation complete

## Final Validation
After this split is complete:
1. Both splits should merge cleanly
2. Combined size should match original (1136 lines)
3. All tests should pass
4. Coverage should exceed 80%
5. Ready for Phase 2 integration

## Notes
- This is the FINAL split for fallback-strategies effort
- Completes all functionality from original implementation
- Must maintain compatibility with Split 001
- No breaking changes to interfaces
- Security logging is critical - ensure compliance

---

**Status**: Ready for Implementation (after Split 001)
**Dependencies**: Split 001 MUST be completed first
**Size Validated**: ~756 lines (under limit)
**Next Step**: Complete implementation, then merge both splits