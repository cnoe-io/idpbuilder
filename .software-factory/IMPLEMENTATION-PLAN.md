# P1W1-E3: Registry Configuration Schema - Implementation Plan

## EFFORT INFRASTRUCTURE METADATA
**EFFORT_NAME**: P1W1-E3-registry-config
**PHASE**: 1
**WAVE**: 1
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-gitea-push/efforts/phase1/wave1/P1W1-E3-registry-config
**BRANCH**: phase1/wave1/P1W1-E3-registry-config
**REMOTE**: https://github.com/cnoe-io/idpbuilder
**BASE_BRANCH**: main

## Overview
- **Effort**: Registry Configuration Schema - Define configuration schema for registry connections
- **Phase**: 1, Wave: 1
- **Branch**: `phase1/wave1/P1W1-E3-registry-config`
- **Estimated Size**: 180 lines
- **Can Parallelize**: Yes
- **Parallel With**: [P1W1-E1, P1W1-E2, P1W1-E4]
- **Dependencies**: None - foundational effort
- **Implementation Time**: 2-3 hours

## Objectives
1. Define comprehensive registry configuration structures
2. Create authentication configuration for multiple auth types
3. Implement robust configuration validation
4. Support multiple registry types (Docker Hub, Harbor, Gitea, Generic OCI)
5. Provide secure credential management structures

## File Structure
```
pkg/
└── config/
    ├── registry.go       # Registry configuration types and structures (~70 lines)
    ├── auth.go          # Authentication configuration types (~60 lines)
    └── validation.go    # Configuration validation logic (~50 lines)
```

## Implementation Steps

### Step 1: Create Registry Configuration (pkg/config/registry.go)
**File**: `pkg/config/registry.go`
**Lines**: ~70

1. Define package and imports
2. Create `RegistryConfig` struct:
   - URL/Host configuration
   - Registry type enumeration (DockerHub, Harbor, Gitea, Generic)
   - TLS configuration options
   - Connection timeout settings
   - Retry policy configuration
3. Define `RegistryType` enumeration with constants
4. Create `TLSConfig` struct:
   - Insecure skip verify option
   - CA certificate path
   - Client certificate configuration
5. Define `RetryPolicy` struct:
   - Max attempts
   - Backoff configuration
   - Timeout settings
6. Add helper methods:
   - `GetRegistryEndpoint()` - Returns formatted registry URL
   - `IsSecure()` - Determines if TLS is enabled
   - `GetTimeout()` - Returns configured timeout with defaults

### Step 2: Create Authentication Configuration (pkg/config/auth.go)
**File**: `pkg/config/auth.go`
**Lines**: ~60

1. Define package and imports
2. Create `AuthConfig` struct:
   - Auth type enumeration (Basic, Token, OAuth2, Anonymous)
   - Credentials storage
   - Token configuration
   - OAuth2 configuration
3. Define `AuthType` enumeration:
   - `AuthTypeBasic` - Username/password
   - `AuthTypeToken` - Bearer token
   - `AuthTypeOAuth2` - OAuth2 flow
   - `AuthTypeAnonymous` - No authentication
4. Create `BasicAuth` struct:
   - Username field
   - Password field (with secure storage consideration)
5. Create `TokenAuth` struct:
   - Token value
   - Token type (Bearer, etc.)
   - Refresh token (if applicable)
6. Create `OAuth2Config` struct:
   - Client ID
   - Client secret
   - Token URL
   - Scopes
7. Add helper methods:
   - `GetAuthHeader()` - Returns appropriate auth header
   - `IsAnonymous()` - Checks if auth is required
   - `Validate()` - Validates auth configuration

### Step 3: Implement Configuration Validation (pkg/config/validation.go)
**File**: `pkg/config/validation.go`
**Lines**: ~50

1. Define package and imports
2. Create validation functions:
   - `ValidateRegistryConfig(config *RegistryConfig) error`
   - `ValidateAuthConfig(config *AuthConfig) error`
   - `ValidateTLSConfig(config *TLSConfig) error`
3. Implement `ValidateRegistryConfig`:
   - Check required URL/host
   - Validate URL format
   - Ensure registry type is valid
   - Validate timeout ranges
   - Check retry policy sanity
4. Implement `ValidateAuthConfig`:
   - Ensure auth type is valid
   - Validate credentials based on auth type
   - Check OAuth2 configuration completeness
   - Validate token format if provided
5. Implement `ValidateTLSConfig`:
   - Validate certificate paths exist (if provided)
   - Check CA certificate validity
   - Ensure client cert/key pair if provided
6. Add helper validation functions:
   - `isValidURL(url string) bool`
   - `fileExists(path string) bool`
   - `isValidTimeout(duration time.Duration) bool`

## Dependencies and Imports

### External Dependencies
```go
import (
    "crypto/tls"
    "errors"
    "fmt"
    "net/url"
    "os"
    "time"
)
```

### Internal Dependencies
None - This is a foundational effort with no internal dependencies

## Testing Requirements

### Unit Tests Required
1. Registry configuration validation tests
2. Authentication configuration tests
3. TLS configuration validation tests
4. URL parsing and validation tests
5. Edge cases for all validation functions

### Test Coverage Target
- **Unit Tests**: 90% coverage minimum
- **Focus Areas**: Validation logic, configuration parsing
- **Test Files**: Will be created in Phase 1 Wave 3 or Phase 2

## Integration Points

### Used By (Future Efforts)
- P1W2-E1: Base OCI Registry Client (will use these configs)
- P1W2-E2: Authentication Handler (will consume auth configs)
- P1W3-E1 through P1W3-E4: Provider implementations

### Interfaces Defined
- Configuration structures that all registry providers will use
- Standard validation patterns for registry connections
- Authentication configuration interface

## Security Considerations

1. **Credential Storage**:
   - Never log passwords or tokens
   - Consider using credential helpers in future
   - Support environment variable substitution

2. **TLS Configuration**:
   - Default to secure connections
   - Clear warnings for insecure configurations
   - Validate certificate chains

3. **Validation**:
   - Strict URL validation to prevent injection
   - Path traversal prevention for certificate paths
   - Timeout limits to prevent DoS

## Error Handling Strategy

1. Use descriptive error messages
2. Wrap errors with context
3. Define custom error types for validation failures
4. Provide actionable error messages for configuration issues

## Code Style Guidelines

1. Follow Go best practices and idioms
2. Use meaningful variable and function names
3. Add comprehensive godoc comments
4. Keep functions focused and small
5. Use table-driven tests where appropriate

## Success Criteria

1. All three files implemented within 180 lines total
2. Clean compilation with no warnings
3. All configuration types properly defined
4. Validation functions comprehensive and working
5. Code is well-documented and maintainable
6. Ready for use by Wave 2 efforts

## Risk Mitigation

1. **Size Risk**: Low - Schema definitions are straightforward
2. **Complexity Risk**: Low - No external dependencies or complex logic
3. **Integration Risk**: Low - Foundational effort with clear interfaces

## Notes for SW Engineer

- Start with registry.go to establish base types
- Ensure all structs have JSON tags for serialization
- Consider YAML tags as well for configuration files
- Use pointer fields where optional values make sense
- Keep validation separate from type definitions for clarity
- This effort has NO dependencies - can be implemented immediately in parallel with other P1W1 efforts

## Checklist for Implementation

- [ ] Create pkg/config directory
- [ ] Implement registry.go with all types
- [ ] Implement auth.go with auth configurations
- [ ] Implement validation.go with validation logic
- [ ] Ensure total lines < 180
- [ ] Add comprehensive godoc comments
- [ ] Verify clean compilation
- [ ] Prepare for code review

---
*Generated by Code Reviewer Agent - Phase 1, Wave 1, Effort 3*