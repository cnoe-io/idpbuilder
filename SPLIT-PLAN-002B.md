# SPLIT-PLAN-002B.md - Recommendations System

## Split 002B of 3: User Recommendations Implementation
**Planner**: Code Reviewer (R199 - sole split planner)
**Parent Effort**: E1.2.2 fallback-strategies-split-002
**Created**: 2025-09-01 10:38:00 UTC

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Boundaries (⚠️⚠️⚠️ CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: Split 002A of phase1/wave2/fallback-strategies-split-002
  - Path: efforts/phase1/wave2/fallback-strategies-split-002/split-002a/
  - Branch: phase1/wave2/fallback-strategies-split-002a
  - Summary: Security logger implementation (370 lines)
- **This Split**: Split 002B of phase1/wave2/fallback-strategies-split-002
  - Path: efforts/phase1/wave2/fallback-strategies-split-002/split-002b/
  - Branch: phase1/wave2/fallback-strategies-split-002b
- **Next Split**: Split 002C of phase1/wave2/fallback-strategies-split-002
  - Path: efforts/phase1/wave2/fallback-strategies-split-002/split-002c/
  - Branch: phase1/wave2/fallback-strategies-split-002c

## Files in This Split (EXCLUSIVE - no overlap with other splits)
- `pkg/certs/fallback/recommendations.go` - 691 lines (ALREADY IMPLEMENTED)
  - CertErrorType enumeration
  - Recommendation structure
  - RecommendationEngine interface
  - DefaultRecommendationEngine implementation
  - Error-specific recommendation generation
  - Context-aware suggestions
  - Recovery strategies

## Current Status
**IMPORTANT**: This code has ALREADY been implemented by the SW Engineer in the parent split-002 directory.
This sub-split plan is for organizational purposes to manage the size violation.

## Size Analysis
- **Current Lines**: 691 lines (verified)
- **Limit**: 800 lines
- **Status**: COMPLIANT (but near limit)
- **Margin**: 109 lines available
- **Risk**: HIGH - very close to limit

## Functionality Implemented
1. **Error Type Classification**: 8 different certificate error types
2. **Recommendation Generation**: Context-aware user guidance
3. **Risk Assessment**: Security level determination
4. **Action Suggestions**: Step-by-step remediation
5. **Context Integration**: Registry-specific recommendations
6. **Formatting**: User-friendly message construction

## Key Components
- **CertErrorType**: Enumeration of certificate error categories
- **Recommendation**: Structure containing severity, message, actions, and context
- **RecommendationEngine**: Interface for generating recommendations
- **DefaultRecommendationEngine**: Main implementation with error-specific logic

## Dependencies
- Standard library only (fmt, strings, time)
- Imports from Split 002A: Uses SecurityLevel from logger.go

## Integration Points
- Depends on: Split 002A (logger.go for SecurityLevel)
- Used by: Split 002C (tests will verify recommendations)
- Used by: Split 001 (detector will call recommendation engine)

## Testing Requirements (To be done in Split 002C)
- Unit tests for each error type
- Recommendation accuracy verification
- Context handling tests
- Edge case coverage
- Integration with logger

## Implementation Instructions for Orchestrator
Since this code is ALREADY implemented:
1. **Critical**: This file is at 691 lines - VERY close to 700-line soft limit
2. **Monitor**: Any additions could push it over limit
3. **Option A**: Keep as-is if no changes needed
4. **Option B**: If ANY additions needed, must refactor first
5. **Recommendation**: Accept as-is, put ALL tests in Split 002C

## Risk Mitigation
- **DO NOT** add any more code to recommendations.go
- **ALL** additional functionality goes to new files
- **ALL** tests go to Split 002C
- Consider refactoring if any changes needed

## Validation Checklist
- [x] File compiles with dependencies
- [x] Imports from Split 002A work
- [x] Size under limit (691 < 700 soft, < 800 hard)
- [x] All error types handled
- [ ] Tests in Split 002C
- [ ] Integration verified

## Notes
- This file grew larger than originally estimated (691 vs 344 planned)
- Very close to soft limit - handle with care
- Comprehensive implementation covers all certificate error scenarios
- Ready for testing in Split 002C