# SPLIT-PLAN-001: Core Interfaces and Authentication
## Split 001 of 2: Foundation Components
**Planner**: Code Reviewer
**Parent Effort**: gitea-client (E2.1.2)
**Created**: 2025-01-09 10:45:00

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Split Metadata
- **Split Number**: 001 of 2
- **Branch**: phase2/wave1/gitea-client-split-001
- **Size Estimate**: 635 lines
- **Focus**: Core interfaces, authentication, and main registry implementation

## Boundaries (⚠️ CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: None (first split of THIS effort)
  - Path: N/A (this is Split 001)
  - Branch: N/A
- **This Split**: Split 001 of phase2/wave1/gitea-client
  - Path: efforts/phase2/wave1/gitea-client/split-001/
  - Branch: phase2/wave1/gitea-client-split-001
- **Next Split**: Split 002 of phase2/wave1/gitea-client
  - Path: efforts/phase2/wave1/gitea-client/split-002/
  - Branch: phase2/wave1/gitea-client-split-002
  - Summary: Push/list operations, retry logic, and test stubs

## Files in This Split (EXCLUSIVE - no overlap with other splits)
```
pkg/registry/
├── interface.go       (24 lines)  - Core Registry interface definition
├── auth.go           (138 lines) - Authentication with token management
├── gitea.go          (204 lines) - Main Gitea registry implementation
└── remote_options.go (269 lines) - Remote registry configuration
```

## Functionality Implemented
### 1. Core Registry Interface (`interface.go`)
- Define Registry interface with required methods
- Establish contract for implementations
- Export types for use by other packages

### 2. Authentication System (`auth.go`)
- Token-based authentication for Gitea
- Credential management
- Authorization header generation
- Token refresh logic

### 3. Main Gitea Client (`gitea.go`)
- Implement Registry interface
- Connection management
- Base registry operations
- Error handling and logging

### 4. Remote Configuration (`remote_options.go`)
- Configure remote registry settings
- TLS/SSL configuration
- Proxy settings
- Timeout and retry configurations
- Registry URL management

## Dependencies
- **External**: 
  - github.com/google/go-containerregistry
  - github.com/sirupsen/logrus (logging)
  - Standard library (net/http, crypto/tls)
- **Internal**: None (foundational split)

## Implementation Instructions
1. **Setup Split Directory**
   ```bash
   cd efforts/phase2/wave1/gitea-client
   mkdir -p split-001/pkg/registry
   ```

2. **Create Branch**
   ```bash
   git checkout -b phase2/wave1/gitea-client-split-001
   ```

3. **Implement Files in Order**
   - Start with `interface.go` (defines contracts)
   - Then `auth.go` (authentication layer)
   - Then `gitea.go` (main implementation)
   - Finally `remote_options.go` (configuration)

4. **Ensure Compilation**
   ```bash
   cd split-001
   go mod init github.com/cnoe-io/idpbuilder-gitea-client
   go mod tidy
   go build ./...
   ```

5. **Write Unit Tests**
   - Test authentication flow
   - Test configuration options
   - Mock registry responses

6. **Measure Size**
   ```bash
   $PROJECT_ROOT/tools/line-counter.sh
   ```

## Testing Requirements
- Unit tests for authentication logic
- Tests for configuration validation
- Interface compliance tests
- Mock server tests for basic operations

## Success Criteria
- [x] All files compile without errors
- [x] Interfaces are well-defined and documented
- [x] Authentication works with Gitea tokens
- [x] Configuration options are validated
- [x] Size remains under 700 lines
- [x] Code follows Go best practices
- [x] No dependency on Split 002 files

## Notes
- This split provides the foundation for all registry operations
- Focus on clean interface design and robust authentication
- Ensure all exported types are properly documented
- Configuration should be flexible but secure by default