# SPLIT-PLAN-002.md
## Split 002 of 2: Certificate Types and TLS Configuration
**Planner**: Code Reviewer code-reviewer-1756082516 (same for ALL splits)
**Parent Effort**: registry-auth-types

<!-- ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

## Boundaries (CRITICAL: All splits MUST reference SAME effort!)
- **Previous Split**: Split 001 of phase1/wave1/registry-auth-types
  - Path: efforts/phase1/wave1/registry-auth-types/split-001/
  - Branch: phase1/wave1/registry-auth-types-split-001
  - Summary: Implemented all authentication types, credentials, and package documentation
- **This Split**: Split 002 of phase1/wave1/registry-auth-types
  - Path: efforts/phase1/wave1/registry-auth-types/split-002/
  - Branch: phase1/wave1/registry-auth-types-split-002
- **Next Split**: None (final split of this effort)
  - Path: N/A
  - Branch: N/A
- **File Boundaries**:
  - Previous Split End: pkg/doc.go
  - This Split Start: pkg/certs/types.go
  - This Split End: pkg/certs/constants.go (last file)

## Files in This Split (EXCLUSIVE - no overlap with other splits)
- `pkg/certs/types.go` (175 lines) - Certificate types and TLS configuration
- `pkg/certs/constants.go` (135 lines) - Certificate-related constants

**Total Lines**: 310 lines (COMPLIANT - well under 800 line limit)

## Functionality
### Certificate Types (`pkg/certs/types.go`)
- `CertificateBundle` struct:
  - CACert, ClientCert, ClientKey fields
  - Validity period tracking
  - Certificate chain validation
- `TLSConfig` struct:
  - InsecureSkipVerify option
  - RootCAs and ClientCAs configuration
  - Certificate array management
- `Certificate` type wrapper for x509.Certificate
- `CertificateValidator` interface for validation
- Methods for certificate validation and verification

### Certificate Constants (`pkg/certs/constants.go`)
- Certificate type identifiers
- Default certificate paths
- Validation error messages
- TLS version constants (TLS 1.2, 1.3)
- Certificate file extensions (.crt, .pem, .key)
- Common certificate field names

## Dependencies
```go
// Standard library only for Phase 1
import (
    "crypto/tls"
    "crypto/x509"
    "encoding/pem"
    "errors"
    "fmt"
    "io"
    "time"
)
```

## Implementation Instructions
1. **Create sparse checkout** with ONLY these files:
   ```bash
   git sparse-checkout set pkg/certs
   ```

2. **Verify isolation**:
   - No dependencies on pkg/auth files (split 001)
   - Certificate functionality must be self-contained
   - Can reference standard crypto libraries only

3. **Implementation order**:
   - Start with constants.go (defines foundation)
   - Implement types.go (uses constants)
   - Ensure all certificate operations are secure

4. **Quality checks**:
   - All types must compile independently
   - Proper error handling for invalid certificates
   - Clear godoc comments on all exported types
   - Validate with: `go build ./pkg/certs`

5. **Size verification**:
   ```bash
   ${PROJECT_ROOT}/tools/line-counter.sh
   # Must show <800 lines for this split (expect ~310)
   ```

## Test Requirements
- **Unit Tests**: Create corresponding test files
  - `pkg/certs/types_test.go` - Certificate validation tests
  - `pkg/certs/constants_test.go` - Constant usage validation
- **Coverage Target**: 80% minimum
- **Test Scenarios**:
  - Valid/invalid certificates
  - Certificate expiration checks
  - TLS configuration validation
  - Certificate chain verification
  - PEM encoding/decoding

## Integration Considerations
- This split provides certificate types used by authentication
- After both splits merge, the full registry-auth-types package is complete
- The types defined here will be consumed by registry client implementations
- Must maintain backward compatibility with Docker registry auth

## Split Branch Strategy
- **Branch Name**: `phase1/wave1/registry-auth-types-split-002`
- **Base Branch**: `phase1/wave1/registry-auth-types`
- **Merge Target**: Back to `phase1/wave1/registry-auth-types` after review
- **Commit Message**: "split-002: Implement certificate types and TLS configuration"

## Success Criteria
- All certificate types compile without errors
- No dependencies on authentication types (split 001)
- Secure certificate handling patterns
- Support for common certificate formats
- Complete godoc documentation
- Total implementation <800 lines (actual: ~310)
- Tests provide 80% coverage

## Review Checklist
- [ ] Files match exactly those listed (no extras)
- [ ] No references to pkg/auth
- [ ] All certificate types properly defined
- [ ] Security patterns followed (secure TLS defaults)
- [ ] Line count verified with designated tool
- [ ] Tests cover certificate validation
- [ ] Documentation complete

## Notes for SW Engineer
- This is a smaller split (310 lines) but critical for TLS security
- Focus on making the certificate validation robust
- Consider future extensibility for custom certificate validators
- Ensure constants cover common registry certificate scenarios
- Default to secure TLS configurations (minimum TLS 1.2)
## 🚨 SPLIT INFRASTRUCTURE METADATA (Added by Orchestrator)
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-oci-mgmt/efforts/phase1/wave1/registry-auth-types--split-002
**BRANCH**: phase1/wave1/registry-auth-types--split-002
**REMOTE**: origin/phase1/wave1/registry-auth-types--split-002
**BASE_BRANCH**: phase1/wave1/registry-auth-types--split-001
**SPLIT_NUMBER**: 002
**TOTAL_SPLITS**: 2

### SW Engineer Instructions (R205)
1. READ this metadata FIRST
2. cd to WORKING_DIRECTORY above
3. Verify branch matches BRANCH above
4. ONLY THEN proceed with preflight checks
