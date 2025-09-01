# SPLIT-PLAN-001A.md
## Split 001A of 2: Certificate Error Detection and Handler Interfaces

**Planner**: Code Reviewer Instance (code-reviewer)
**Parent Effort**: E1.2.2 fallback-strategies Split 001
**Created**: 2025-09-01 06:30:00 UTC

<!-- ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Split Boundaries
- **Previous Split**: None (this is first sub-split of Split 001)
  - Path: N/A (this is Sub-Split 001A)
  - Branch: N/A
- **This Split**: Sub-Split 001A of phase1/wave2/fallback-strategies-split-001
  - Path: efforts/phase1/wave2/fallback-strategies-split-001/sub-split-001A/
  - Branch: phase1/wave2/fallback-strategies-split-001-sub-001A
- **Next Split**: Sub-Split 001B of phase1/wave2/fallback-strategies-split-001
  - Path: efforts/phase1/wave2/fallback-strategies-split-001/sub-split-001B/
  - Branch: phase1/wave2/fallback-strategies-split-001-sub-001B
- **File Boundaries**:
  - This Split Start: Line 1 / File: pkg/certs/fallback/detector.go
  - This Split End: Line 200 / File: pkg/certs/fallback/handler.go (interfaces section)
  - Next Split Start: Line 201 / File: pkg/certs/fallback/handler.go (implementation section)

## Files in This Sub-Split (EXCLUSIVE - no overlap with 001B)
1. **pkg/certs/fallback/detector.go** (COMPLETE FILE - 522 lines)
   - Lines 1-124: Core error types and interfaces
     - CertErrorType enum (lines 14-47)
     - ErrorDetails struct (lines 75-103)
     - CertErrorDetector interface (lines 105-124)
   - Lines 125-523: DefaultCertErrorDetector implementation
     - Constructor and core methods (lines 141-179)
     - Error classification logic (lines 181-241)
     - Chain validation (lines 243-321)
     - CA trust checking (lines 323-349)
     - Certificate extraction (lines 351-374)
     - Recoverability determination (lines 376-400)
     - Error detail population (lines 402-499)
     - Configuration methods (lines 501-523)

2. **pkg/certs/fallback/handler_types.go** (NEW FILE - ~200 lines)
   - Extract from handler.go lines 1-200:
     - Package declaration and imports
     - FallbackAction enum (lines 14-50)
     - FallbackDecision struct (lines 52-77)
     - FallbackStrategy struct (lines 79-113)
     - FallbackMode enum (lines 115-151)
     - FallbackHandler interface (lines 153-178)
     - UserPrompter interface (lines 180-187)
     - SecurityLogger interface (lines 189-200)

## Functionality in This Sub-Split
- Complete certificate error detection and classification
- Error type definitions and enumerations
- Certificate chain validation logic
- Trust authority verification
- Error recoverability assessment
- All handler type definitions and interfaces
- No implementation of handler logic (that's in 001B)

## Dependencies
- Standard library: crypto/tls, crypto/x509, fmt, net, strings, time, context, sync
- No external dependencies
- No dependencies on Split 001B (interfaces only)

## Implementation Instructions for SW Engineer
1. Create workspace for sub-split-001A:
   ```bash
   cd efforts/phase1/wave2/fallback-strategies-split-001
   mkdir sub-split-001A
   cd sub-split-001A
   git checkout -b phase1/wave2/fallback-strategies-split-001-sub-001A
   ```

2. Copy the complete detector.go file:
   ```bash
   mkdir -p pkg/certs/fallback
   cp ../pkg/certs/fallback/detector.go pkg/certs/fallback/
   ```

3. Create handler_types.go with ONLY types and interfaces:
   ```bash
   # Extract lines 1-200 from handler.go
   # Create new file handler_types.go with these contents
   # Ensure package declaration and imports are included
   ```

4. Verify the split compiles:
   ```bash
   go build ./pkg/certs/fallback/...
   ```

5. Measure size to confirm under 800 lines:
   ```bash
   $PROJECT_ROOT/tools/line-counter.sh
   ```

## Size Estimates
- detector.go: 522 lines (complete file)
- handler_types.go: ~200 lines (extracted interfaces)
- **Total Estimated**: ~722 lines (WELL UNDER 800 LIMIT)

## Split Integration Strategy
- Sub-Split 001A provides all types and interfaces
- Sub-Split 001B will import from 001A for type definitions
- Clean interface boundary at handler implementation
- No circular dependencies possible

## Verification Checklist
- [ ] detector.go copied completely (522 lines)
- [ ] handler_types.go contains ONLY types/interfaces (no implementation)
- [ ] Package compiles successfully
- [ ] Total size under 800 lines (target: ~722)
- [ ] No overlap with Split 001B content
- [ ] All imports properly handled

## Notes for Orchestrator
- This is a sub-split of the oversized Split 001
- Must be implemented sequentially BEFORE 001B
- 001B depends on type definitions from 001A
- Both sub-splits must be merged back to parent branch after completion