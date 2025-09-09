# SPLIT-PLAN-002: Operations and Utilities
## Split 002 of 2: Push/List Operations and Support
**Planner**: Code Reviewer
**Parent Effort**: gitea-client (E2.1.2)
**Created**: 2025-01-09 10:45:00

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Split Metadata
- **Split Number**: 002 of 2
- **Branch**: phase2/wave1/gitea-client-split-002
- **Size Estimate**: 633 lines
- **Focus**: Push/list operations, retry logic, and testing utilities

## Boundaries (⚠️ CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: Split 001 of phase2/wave1/gitea-client
  - Path: efforts/phase2/wave1/gitea-client/split-001/
  - Branch: phase2/wave1/gitea-client-split-001
  - Summary: Core interfaces, authentication, and main registry implementation
- **This Split**: Split 002 of phase2/wave1/gitea-client
  - Path: efforts/phase2/wave1/gitea-client/split-002/
  - Branch: phase2/wave1/gitea-client-split-002
- **Next Split**: None (final split of THIS effort)

## Files in This Split (EXCLUSIVE - no overlap with other splits)
```
pkg/registry/
├── push.go   (302 lines) - Image push operations
├── list.go   (90 lines)  - Image listing functionality
├── retry.go  (52 lines)  - Retry logic with exponential backoff
└── stubs.go  (189 lines) - Test stubs and mocks
```

## Functionality Implemented
### 1. Push Operations (`push.go`)
- Complete image push workflow
- Multi-layer push support
- Manifest handling
- Progress reporting
- Error recovery
- Chunked uploads
- Digest verification

### 2. List Operations (`list.go`)
- List images in registry
- Filter by repository
- Tag enumeration
- Pagination support
- Metadata retrieval

### 3. Retry Logic (`retry.go`)
- Exponential backoff implementation
- Configurable retry policies
- Network error handling
- Transient failure recovery
- Rate limiting support

### 4. Test Stubs (`stubs.go`)
- Mock registry implementation
- Test helpers
- Stub responses
- Error injection for testing
- Performance testing utilities

## Dependencies
- **From Split 001**:
  - Registry interface
  - Authentication types
  - Configuration structures
- **External**:
  - github.com/google/go-containerregistry
  - github.com/sirupsen/logrus
  - Standard library (time, errors, io)

## Implementation Instructions
1. **Setup Split Directory**
   ```bash
   cd efforts/phase2/wave1/gitea-client
   mkdir -p split-002/pkg/registry
   ```

2. **Create Branch from Split 001**
   ```bash
   # Branch from split-001 to get interfaces
   git checkout phase2/wave1/gitea-client-split-001
   git checkout -b phase2/wave1/gitea-client-split-002
   ```

3. **Copy Interfaces from Split 001**
   ```bash
   # Import needed types from split-001
   cp split-001/pkg/registry/interface.go split-002/pkg/registry/
   ```

4. **Implement Files in Order**
   - Start with `retry.go` (utility used by others)
   - Then `push.go` (main operation)
   - Then `list.go` (secondary operation)
   - Finally `stubs.go` (testing support)

5. **Ensure Compilation**
   ```bash
   cd split-002
   go mod init github.com/cnoe-io/idpbuilder-gitea-client
   go mod tidy
   go build ./...
   ```

6. **Write Comprehensive Tests**
   - Test push with various image sizes
   - Test list with pagination
   - Test retry logic scenarios
   - Use stubs for isolated testing

7. **Measure Size**
   ```bash
   $PROJECT_ROOT/tools/line-counter.sh
   ```

## Testing Requirements
- **Push Operations**:
  - Single layer push
  - Multi-layer push
  - Large image handling
  - Failure recovery
  - Progress reporting accuracy
  
- **List Operations**:
  - Empty registry
  - Multiple repositories
  - Tag filtering
  - Pagination
  
- **Retry Logic**:
  - Network failures
  - Timeout scenarios
  - Rate limiting
  - Max retry limits
  
- **Test Coverage**: Minimum 80% for all files

## Integration Points
- Push operations use authentication from Split 001
- List operations use registry client from Split 001
- Retry logic wraps registry operations
- Stubs implement interfaces from Split 001

## Success Criteria
- [x] All operations work with real Gitea registry
- [x] Retry logic handles common failure scenarios
- [x] Test stubs enable comprehensive testing
- [x] Size remains under 700 lines
- [x] Integration with Split 001 is seamless
- [x] Performance meets requirements
- [x] Error messages are informative

## Notes
- Push operation is the most complex - ensure proper error handling
- Retry logic should be configurable per operation
- Test stubs should simulate real registry behavior accurately
- Consider memory efficiency for large image pushes
- Ensure proper cleanup on failures