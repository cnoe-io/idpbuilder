# SPLIT-PLAN-002C.md - Comprehensive Test Suite

## Split 002C of 3: Test Implementation
**Planner**: Code Reviewer (R199 - sole split planner)
**Parent Effort**: E1.2.2 fallback-strategies-split-002
**Created**: 2025-09-01 10:38:00 UTC

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Boundaries (⚠️⚠️⚠️ CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: Split 002B of phase1/wave2/fallback-strategies-split-002
  - Path: efforts/phase1/wave2/fallback-strategies-split-002/split-002b/
  - Branch: phase1/wave2/fallback-strategies-split-002b
  - Summary: Recommendations system implementation (691 lines)
- **This Split**: Split 002C of phase1/wave2/fallback-strategies-split-002
  - Path: efforts/phase1/wave2/fallback-strategies-split-002/split-002c/
  - Branch: phase1/wave2/fallback-strategies-split-002c
- **Next Split**: None (final split of effort)

## Files in This Split (NEW IMPLEMENTATION REQUIRED)
- `pkg/certs/fallback/logger_test.go` - ~250 lines (to be created)
- `pkg/certs/fallback/recommendations_test.go` - ~350 lines (to be created)
- `pkg/certs/fallback/integration_test.go` - ~80 lines (to be created)

## Current Status
**IMPORTANT**: These test files have NOT been implemented yet.
This split represents NEW WORK that needs to be done.

## Size Analysis
- **Estimated Lines**: ~680 lines total
  - logger_test.go: ~250 lines
  - recommendations_test.go: ~350 lines
  - integration_test.go: ~80 lines
- **Limit**: 800 lines
- **Status**: COMPLIANT (with buffer)
- **Margin**: 120 lines available

## Test Coverage Requirements

### Logger Tests (logger_test.go) - ~250 lines
1. **SecurityLevel Tests**
   - String representation
   - Valid level ranges
   
2. **SecurityLogEntry Tests**
   - JSON marshaling/unmarshaling
   - Field validation
   - Timestamp formatting
   
3. **DefaultSecurityLogger Tests**
   - Log method with all levels
   - Concurrent write safety
   - File output verification
   - Multiple writer support
   - Error handling

4. **File Rotation Tests**
   - Size-based rotation
   - Timestamp in filenames
   - Old file cleanup

### Recommendations Tests (recommendations_test.go) - ~350 lines
1. **CertErrorType Tests**
   - String representation
   - All error types covered

2. **Recommendation Structure Tests**
   - Field validation
   - JSON serialization

3. **DefaultRecommendationEngine Tests**
   - Test each error type:
     * CertExpired
     * CertNotYetValid
     * CertHostnameMismatch
     * CertUntrustedRoot
     * CertSelfSigned
     * CertRevoked
     * CertInvalidSignature
     * CertUnknownError
   - Context-aware recommendations
   - Risk level assignment
   - Action list generation

4. **Edge Cases**
   - Nil inputs
   - Empty contexts
   - Invalid error types

### Integration Tests (integration_test.go) - ~80 lines
1. **Logger-Recommendation Integration**
   - Log recommendations
   - Security level consistency
   
2. **End-to-End Scenarios**
   - Certificate error → Recommendation → Log
   - Multiple error handling
   
3. **Performance Tests**
   - Recommendation generation speed
   - Logger throughput

## Implementation Instructions
1. **Create test structure**:
   ```go
   package fallback
   
   import (
       "testing"
       "encoding/json"
       "bytes"
       // other imports
   )
   ```

2. **Follow Go testing conventions**:
   - Test functions: `Test<FunctionName>`
   - Table-driven tests where appropriate
   - Subtests for organization
   - Proper cleanup with t.Cleanup()

3. **Use testify assertions** (if available):
   ```go
   assert.Equal(t, expected, actual)
   assert.NoError(t, err)
   ```

4. **Mock external dependencies**:
   - File system operations
   - Time-based functions

## Dependencies for Tests
- Standard library testing package
- Optional: testify/assert for better assertions
- Optional: testify/mock for mocking

## Test Execution Requirements
- All tests must pass: `go test ./pkg/certs/fallback/...`
- Coverage target: >80%
- No race conditions: `go test -race`
- Benchmarks for critical paths

## Validation Checklist
- [ ] All public functions have tests
- [ ] >80% code coverage achieved
- [ ] Race condition free
- [ ] Tests are deterministic
- [ ] Error cases covered
- [ ] Integration scenarios work
- [ ] Performance acceptable

## Risk Areas to Focus On
1. **Concurrent Access**: Logger mutex protection
2. **File Operations**: Proper cleanup in tests
3. **Error Handling**: All error paths tested
4. **Context Handling**: Various registry contexts

## Notes
- This is the final split for the fallback-strategies effort
- Tests are critical for validating Splits 002A and 002B
- Keep tests focused and avoid redundancy
- Use table-driven tests to stay within size limit