# SPLIT-PLAN-001B.md
## Split 001B of 2: Fallback Handler Implementation

**Planner**: Code Reviewer Instance (code-reviewer)
**Parent Effort**: E1.2.2 fallback-strategies Split 001
**Created**: 2025-09-01 06:32:00 UTC

<!-- ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Split Boundaries
- **Previous Split**: Sub-Split 001A of phase1/wave2/fallback-strategies-split-001
  - Path: efforts/phase1/wave2/fallback-strategies-split-001/sub-split-001A/
  - Branch: phase1/wave2/fallback-strategies-split-001-sub-001A
  - Summary: Implemented detector.go and handler types/interfaces
- **This Split**: Sub-Split 001B of phase1/wave2/fallback-strategies-split-001
  - Path: efforts/phase1/wave2/fallback-strategies-split-001/sub-split-001B/
  - Branch: phase1/wave2/fallback-strategies-split-001-sub-001B
- **Next Split**: None (final sub-split for Split 001)
  - Note: Original Split 002 (logging/recommendations) still pending as separate effort
- **File Boundaries**:
  - Previous Split End: Line 200 / File: pkg/certs/fallback/handler.go (interfaces)
  - This Split Start: Line 201 / File: pkg/certs/fallback/handler.go (implementation)
  - This Split End: Line 608 / File: pkg/certs/fallback/handler.go (end of file)

## Files in This Sub-Split (EXCLUSIVE - no overlap with 001A)
1. **pkg/certs/fallback/handler_impl.go** (NEW FILE - ~408 lines)
   - Extract from handler.go lines 201-608:
     - DefaultFallbackHandler struct (lines 202-227)
     - Constructor functions (lines 229-327)
       - NewDefaultFallbackHandler (lines 229-237)
       - NewSecureStrategy (lines 239-261)
       - NewDevelopmentStrategy (lines 263-294)
       - NewInteractiveStrategy (lines 296-327)
     - Core handler methods (lines 329-479)
       - HandleError (lines 329-384)
       - GetStrategy/UpdateStrategy (lines 386-405)
       - Host trust management (lines 407-436)
       - CreateTLSConfig (lines 438-470)
       - LogSecurityDecision stub (lines 472-479)
     - Helper methods (lines 481-608)
       - SetUserPrompter/SetSecurityLogger (lines 481-489)
       - determineAction (lines 491-522)
       - createDecision (lines 524-547)
       - createDecisionKey (lines 549-553)
       - rememberDecision (lines 555-564)
       - getReasonForAction (lines 566-582)
       - assessSecurityRisk (lines 584-608)

## Functionality in This Sub-Split
- Complete implementation of DefaultFallbackHandler
- Strategy creation functions (Secure, Development, Interactive)
- Error handling decision logic
- TLS configuration generation
- Host trust management
- Decision caching and memory
- Security risk assessment
- Integration points for logging (stub only - full impl in Split 002)

## Dependencies
- Imports types from Sub-Split 001A:
  - CertErrorDetector interface
  - CertErrorType enum
  - ErrorDetails struct
  - All types defined in handler_types.go
- Standard library: context, crypto/tls, fmt, sync, time
- No external dependencies

## Implementation Instructions for SW Engineer
1. Create workspace for sub-split-001B:
   ```bash
   cd efforts/phase1/wave2/fallback-strategies-split-001
   mkdir sub-split-001B
   cd sub-split-001B
   git checkout -b phase1/wave2/fallback-strategies-split-001-sub-001B
   ```

2. Copy type definitions from 001A (for compilation):
   ```bash
   mkdir -p pkg/certs/fallback
   # Copy detector.go and handler_types.go from 001A for compilation
   cp ../sub-split-001A/pkg/certs/fallback/detector.go pkg/certs/fallback/
   cp ../sub-split-001A/pkg/certs/fallback/handler_types.go pkg/certs/fallback/
   ```

3. Create handler_impl.go with implementation code:
   ```bash
   # Extract lines 201-608 from original handler.go
   # Create new file handler_impl.go
   # Add appropriate package declaration and imports
   # Reference types from handler_types.go
   ```

4. Verify the split compiles:
   ```bash
   go build ./pkg/certs/fallback/...
   ```

5. Run tests if available:
   ```bash
   go test ./pkg/certs/fallback/...
   ```

6. Measure size to confirm under 800 lines:
   ```bash
   $PROJECT_ROOT/tools/line-counter.sh
   # Should show ~408 lines for handler_impl.go (NEW code)
   # detector.go and handler_types.go are from 001A (not counted)
   ```

## Size Estimates
- handler_impl.go: ~408 lines (lines 201-608 from original handler.go)
- **Total NEW Code**: ~408 lines (WELL UNDER 800 LIMIT)
- Note: Files from 001A are included for compilation but not counted

## Split Integration Strategy
- Sub-Split 001B depends on types from 001A
- Must be implemented AFTER 001A completes
- Both sub-splits combine to provide complete Split 001 functionality
- Clean separation at interface/implementation boundary
- No circular dependencies

## Verification Checklist
- [ ] handler_impl.go contains ONLY implementation (lines 201-608)
- [ ] All type definitions imported from handler_types.go
- [ ] Package compiles successfully with 001A dependencies
- [ ] Total NEW code under 800 lines (target: ~408)
- [ ] No duplication of 001A content
- [ ] All methods properly implemented
- [ ] Stub for logging (full impl deferred to Split 002)

## Notes for Orchestrator
- This is the second sub-split of oversized Split 001
- Must be implemented AFTER 001A completes
- Depends on type definitions from 001A
- Both sub-splits (001A + 001B) must be merged to parent branch
- After both complete, original Split 002 (logging) can proceed
- Total combined size: 001A (~722) + 001B (~408) = ~1130 lines (original problem)