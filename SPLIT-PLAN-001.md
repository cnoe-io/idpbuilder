# SPLIT-PLAN-001.md
## Split 001 of 2: Authentication Types and Documentation
**Planner**: Code Reviewer code-reviewer-1756082516 (same for ALL splits)
**Parent Effort**: registry-auth-types

<!-- ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Boundaries (CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: None (first split of THIS effort)
  - Path: N/A (this is Split 001)
  - Branch: N/A
- **This Split**: Split 001 of phase1/wave1/registry-auth-types
  - Path: efforts/phase1/wave1/registry-auth-types/split-001/
  - Branch: phase1/wave1/registry-auth-types-split-001
- **Next Split**: Split 002 of phase1/wave1/registry-auth-types
  - Path: efforts/phase1/wave1/registry-auth-types/split-002/
  - Branch: phase1/wave1/registry-auth-types-split-002
- **File Boundaries**:
  - This Split Start: pkg/auth/types.go (first file)
  - This Split End: pkg/doc.go (last file)
  - Next Split Start: pkg/certs/types.go (first cert file)

## Files in This Split (EXCLUSIVE - no overlap with other splits)
- `pkg/auth/types.go` (224 lines) - Core authentication types and interfaces
- `pkg/auth/credentials.go` (232 lines) - Credential structures and management
- `pkg/auth/constants.go` (104 lines) - Auth-related constants and error messages
- `pkg/doc.go` (89 lines) - Package documentation for the entire registry-auth-types package

**Total Lines**: 649 lines (COMPLIANT - under 800 line limit)

## Functionality
### Authentication Types (`pkg/auth/types.go`)
- `RegistryAuth` interface with GetCredentials(), Validate(), Type() methods
- `AuthConfig` struct for registry authentication configuration
- `AuthType` enum (Basic, Bearer, OAuth2)
- `DockerConfig` struct for docker config.json compatibility
- `AuthStore` interface for credential storage
- `RegistryAuthOptions` for configuration

### Credential Management (`pkg/auth/credentials.go`)
- `Credentials` struct with Username, Password, Token fields
- `CredentialHelper` interface for external helpers
- `CredentialStore` type with Get/Set/Delete methods
- `TokenResponse` for OAuth token flows
- Validation and expiration handling methods

### Constants (`pkg/auth/constants.go`)
- Authentication type constants
- HTTP header names (Authorization, WWW-Authenticate)
- Default token expiry times
- Registry URL patterns
- Error messages for auth failures

### Documentation (`pkg/doc.go`)
- Package overview and purpose
- Usage examples for authentication flows
- Security best practices
- Integration guidelines

## Dependencies
```go
// Standard library only for Phase 1
import (
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"
)
```

## Implementation Instructions
1. **Create sparse checkout** with ONLY these files:
   ```bash
   git sparse-checkout set pkg/auth pkg/doc.go
   ```

2. **Verify isolation**:
   - Ensure no references to pkg/certs files
   - All auth functionality must be self-contained
   - Doc.go should have general overview but focus on auth

3. **Implementation order**:
   - Start with constants.go (defines foundation)
   - Implement types.go (interfaces and structs)
   - Implement credentials.go (uses types)
   - Update doc.go with auth-specific examples

4. **Quality checks**:
   - All types must compile independently
   - No circular dependencies
   - Clear godoc comments on all exported types
   - Validate with: `go build ./pkg/auth`

5. **Size verification**:
   ```bash
   ${PROJECT_ROOT}/tools/line-counter.sh
   # Must show <800 lines for this split
   ```

## Test Requirements
- **Unit Tests**: Create corresponding test files
  - `pkg/auth/types_test.go` - Interface compliance tests
  - `pkg/auth/credentials_test.go` - Credential operations
  - `pkg/auth/constants_test.go` - Constant usage validation
- **Coverage Target**: 80% minimum
- **Test Scenarios**:
  - Valid/invalid credentials
  - Token expiration
  - Auth type detection
  - Credential store operations

## Split Branch Strategy
- **Branch Name**: `phase1/wave1/registry-auth-types-split-001`
- **Base Branch**: `phase1/wave1/registry-auth-types`
- **Merge Target**: Back to `phase1/wave1/registry-auth-types` after review
- **Commit Message**: "split-001: Implement authentication types and credentials"

## Success Criteria
- All auth types compile without errors
- No dependencies on certificate types (split 002)
- Clear separation of concerns
- Secure credential handling patterns
- Complete godoc documentation
- Total implementation <800 lines
- Tests provide 80% coverage

## Review Checklist
- [ ] Files match exactly those listed (no extras)
- [ ] No references to pkg/certs
- [ ] All interfaces properly defined
- [ ] Security patterns followed (no credential logging)
- [ ] Line count verified with designated tool
- [ ] Tests cover main functionality
- [ ] Documentation complete
## 🚨 SPLIT INFRASTRUCTURE METADATA (Added by Orchestrator)
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-oci-mgmt/efforts/phase1/wave1/registry-auth-types--split-001
**BRANCH**: phase1/wave1/registry-auth-types--split-001
**REMOTE**: origin/phase1/wave1/registry-auth-types--split-001
**BASE_BRANCH**: main
**SPLIT_NUMBER**: 001
**TOTAL_SPLITS**: 2

### SW Engineer Instructions (R205)
1. READ this metadata FIRST
2. cd to WORKING_DIRECTORY above
3. Verify branch matches BRANCH above
4. ONLY THEN proceed with preflight checks
