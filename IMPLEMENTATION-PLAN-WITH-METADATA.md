<!-- ⚠️ EFFORT INFRASTRUCTURE METADATA (R213 - ORCHESTRATOR DEFINED) ⚠️ -->
**METADATA_SOURCE**: ORCHESTRATOR (Single Source of Truth)
**METADATA_VERSION**: 1.0
**GENERATED_AT**: $(date -Iseconds)
**GENERATED_BY**: orchestrator

## 🔧 EFFORT INFRASTRUCTURE METADATA
**WORKING_DIRECTORY**: efforts/phase2/wave1/gitea-client
**BRANCH**: idpbuilder-oci-build-push/phase2/wave1/gitea-client
**EFFORT_NAME**: E2.1.2-gitea-client
**EFFORT_NUMBER**: E2.1.2
**PHASE**: 2
**WAVE**: 1
<!-- END EFFORT METADATA -->

# Gitea Registry Client Implementation Plan

**Created**: 2025-09-08T00:18:59Z  
**Location**: .software-factory/phase2/wave1/gitea-client/  
**Phase**: 2 - Build & Push Implementation  
**Wave**: 1 - Core Build & Push  
**Effort**: E2.1.2 - gitea-registry-client  
**Planner**: Code Reviewer Agent  

## 🚨 CRITICAL EFFORT METADATA (FROM WAVE PLAN)

**Branch**: `idpbuilder-oci-build-push/phase2/wave1/gitea-client`  
**Base Branch**: `idpbuilder-oci-build-push/phase1/integration`  
**Can Parallelize**: Yes  
**Parallel With**: [E2.1.1 - image-builder]  
**Size Estimate**: 600 lines (well within 800 limit)  
**Dependencies**: Phase 1 Certificate Infrastructure (certs, certvalidation, fallback)  
**Directory**: `efforts/phase2/wave1/gitea-client/pkg/`  

## 🔴🔴🔴 EXPLICIT SCOPE CONTROL (R311 MANDATORY) 🔴🔴🔴

### IMPLEMENT EXACTLY

**Core Functions (5 total):**
1. `NewGiteaRegistry(config RegistryConfig) Registry` (~40 lines)
2. `Push(ctx context.Context, image v1.Image, reference string) error` (~150 lines)
3. `Authenticate(ctx context.Context) error` (~60 lines)
4. `ListRepositories(ctx context.Context) ([]string, error)` (~50 lines)
5. `GetRemoteOptions() []remote.Option` (~40 lines)

**Core Types (4 total):**
1. `Registry interface` (~15 lines)
2. `giteaRegistryImpl struct` (~20 lines)
3. `RegistryConfig struct` (~15 lines)
4. `authenticator struct` (~10 lines)

**Test Functions (8 total):**
1. `TestNewGiteaRegistry` (~40 lines)
2. `TestPushWithValidCerts` (~50 lines)
3. `TestPushWithInsecureMode` (~50 lines)
4. `TestAuthentication` (~40 lines)
5. `TestListRepositories` (~30 lines)
6. `TestRetryLogic` (~40 lines)
7. `TestProgressReporting` (~30 lines)
8. `TestPhase1Integration` (~50 lines)

**TOTAL ESTIMATED**: ~630 lines (130 lines buffer to 800 limit)

### DO NOT IMPLEMENT

- ❌ Pull operations (future effort)
- ❌ Delete/Remove operations (future effort)
- ❌ Tag management operations (future effort)
- ❌ Manifest inspection (future effort)
- ❌ Registry catalog operations (beyond basic list)
- ❌ Token caching mechanism (future optimization)
- ❌ Connection pooling (future optimization)
- ❌ Comprehensive logging framework
- ❌ Metrics collection
- ❌ Circuit breaker pattern (keep retry simple)
- ❌ Multiple registry support (Gitea only for MVP)
- ❌ OAuth/OIDC authentication (basic auth only)

## 🔄 ATOMIC PR DESIGN (R220 COMPLIANCE)

### PR Summary
"Single PR implementing Gitea registry push operations with Phase 1 certificate integration"

### Can Merge to Main Alone
**YES** - This PR will compile and pass all tests when merged independently

### Feature Flags Needed
```yaml
- flag: "GITEA_REGISTRY_ENABLED"
  purpose: "Enable Gitea registry operations"
  default: false
  location: "pkg/config/features.go"
  activation: "When image-builder effort is also merged"
```

### Stubs Required
```yaml
- stub: "MockImageLoader"
  replaces: "Image loading from builder (E2.1.1)"
  interface: "ImageLoader"
  behavior: "Returns test image for push operations"
  location: "pkg/registry/stubs.go"
```

### Interfaces to Implement
```yaml
- interface: "Registry"
  methods: ["Push", "Authenticate", "ListRepositories"]
  implementation: "Complete in this PR"
  location: "pkg/registry/interface.go"
```

### PR Verification
- ✅ Tests pass alone: Unit tests with mocks
- ✅ Build remains working: No breaking changes
- ✅ Feature flag tested both ways: Tests with flag on/off
- ✅ No external dependencies: Uses stubs for missing components
- ✅ Backward compatible: New functionality only

## 📁 File Structure

```
efforts/phase2/wave1/gitea-client/
├── pkg/
│   └── registry/
│       ├── interface.go        # Registry interface definition (~15 lines)
│       ├── gitea.go           # Main GiteaRegistry implementation (~200 lines)
│       ├── auth.go            # Authentication handling (~60 lines)
│       ├── push.go            # Push operation with cert integration (~150 lines)
│       ├── list.go            # Repository listing operations (~50 lines)
│       ├── retry.go           # Simple retry logic (~60 lines)
│       ├── remote_options.go  # Remote options configuration (~40 lines)
│       ├── stubs.go           # Temporary stubs for missing dependencies (~30 lines)
│       ├── gitea_test.go      # Unit tests for main implementation (~100 lines)
│       ├── push_test.go       # Push operation tests (~100 lines)
│       ├── auth_test.go       # Authentication tests (~40 lines)
│       ├── integration_test.go # Integration tests with Phase 1 (~50 lines)
│       └── test_helpers.go    # Test utilities and mocks (~40 lines)
├── pkg/config/
│   └── features.go            # Feature flags (~20 lines)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🔨 Implementation Sequence

### Step 1: Core Interface Definition (50 lines)
```go
// pkg/registry/interface.go
package registry

import (
    "context"
    v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Registry defines operations for OCI registry interaction
type Registry interface {
    // Push uploads an image to the registry
    Push(ctx context.Context, image v1.Image, reference string) error
    
    // Authenticate performs registry authentication
    Authenticate(ctx context.Context) error
    
    // ListRepositories returns available repositories
    ListRepositories(ctx context.Context) ([]string, error)
}

// RegistryConfig holds registry configuration
type RegistryConfig struct {
    URL      string
    Username string
    Password string
    Insecure bool
}
```

### Step 2: Phase 1 Integration Setup (100 lines)
```go
// pkg/registry/gitea.go
package registry

import (
    "github.com/jessesanford/idpbuilder/pkg/certs"
    "github.com/jessesanford/idpbuilder/pkg/certvalidation"
    "github.com/jessesanford/idpbuilder/pkg/fallback"
)

type giteaRegistryImpl struct {
    config      RegistryConfig
    trustStore  certs.TrustStoreManager
    validator   certvalidation.CertValidator
    fallback    fallback.FallbackHandler
    authn       *authenticator
}

func NewGiteaRegistry(config RegistryConfig) (Registry, error) {
    // Initialize Phase 1 components
    trustStore := certs.NewTrustStoreManager()
    validator := certvalidation.NewCertValidator()
    fallback := fallback.NewFallbackHandler()
    
    return &giteaRegistryImpl{
        config:     config,
        trustStore: trustStore,
        validator:  validator,
        fallback:   fallback,
    }, nil
}
```

### Step 3: Authentication Implementation (60 lines)
```go
// pkg/registry/auth.go
package registry

type authenticator struct {
    username string
    password string
    token    string
}

func (r *giteaRegistryImpl) Authenticate(ctx context.Context) error {
    // Basic authentication with username/password
    // Store token for subsequent operations
    // Return clear error if auth fails
}
```

### Step 4: Push Operation with Certificates (150 lines)
```go
// pkg/registry/push.go
package registry

func (r *giteaRegistryImpl) Push(ctx context.Context, image v1.Image, reference string) error {
    // Parse reference
    // Get remote options with certificate configuration
    // Perform push with progress reporting
    // Handle retries for transient failures
    // Return comprehensive error information
}
```

### Step 5: Remote Options Configuration (40 lines)
```go
// pkg/registry/remote_options.go
package registry

func (r *giteaRegistryImpl) GetRemoteOptions() []remote.Option {
    // Configure TLS using Phase 1 trust store
    // Handle --insecure flag with fallback handler
    // Add authentication
    // Return configured options
}
```

### Step 6: List Operations (50 lines)
```go
// pkg/registry/list.go
package registry

func (r *giteaRegistryImpl) ListRepositories(ctx context.Context) ([]string, error) {
    // Query registry catalog
    // Parse response
    // Return repository list
}
```

### Step 7: Retry Logic (60 lines)
```go
// pkg/registry/retry.go
package registry

func retryWithExponentialBackoff(operation func() error) error {
    // Simple exponential backoff
    // Max 3 retries
    // Return last error if all fail
}
```

### Step 8: Test Implementation (330 lines)
- Unit tests for all public functions
- Mock registry responses
- Test certificate integration
- Test --insecure mode
- Validate error handling
- Progress reporting tests

## 📏 Size Management Strategy

### Monitoring Points
1. **After Step 2**: ~150 lines (Check point)
2. **After Step 4**: ~400 lines (Mid-point check)
3. **After Step 6**: ~500 lines (Warning threshold)
4. **After Step 8**: ~630 lines (Final check)

### Size Measurement Command
```bash
# From effort directory
PROJECT_ROOT=$(pwd)
while [ "$PROJECT_ROOT" != "/" ]; do 
    [ -f "$PROJECT_ROOT/orchestrator-state.yaml" ] && break
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
done
$PROJECT_ROOT/tools/line-counter.sh
```

### Split Trigger
If size approaches 700 lines at any checkpoint:
1. STOP implementation immediately
2. Document completed vs remaining work
3. Request split planning
4. DO NOT continue past 750 lines

## 🧪 Test Requirements

### Unit Test Coverage (80% minimum)
- All public functions must have tests
- Error paths must be tested
- Mock external dependencies
- Test both secure and insecure modes

### Integration Tests
- Test with real Phase 1 certificate components
- Validate TLS configuration
- Test fallback handler for --insecure
- End-to-end push simulation

### Test Scenarios
1. **Happy Path**: Successful push with valid certificates
2. **Insecure Mode**: Push with --insecure flag
3. **Auth Failure**: Invalid credentials handling
4. **Network Issues**: Retry logic validation
5. **Certificate Issues**: Proper error messages
6. **Large Images**: Progress reporting accuracy

## 🔗 Phase 1 Dependencies

### Required Imports
```go
import (
    "github.com/jessesanford/idpbuilder/pkg/certs"
    "github.com/jessesanford/idpbuilder/pkg/certvalidation"
    "github.com/jessesanford/idpbuilder/pkg/fallback"
)
```

### Integration Points
1. **TrustStoreManager**: Configure TLS for registry connection
2. **CertValidator**: Validate registry certificates
3. **FallbackHandler**: Handle --insecure mode safely
4. **Error Types**: Use Phase 1 error definitions

## 🚀 Implementation Checklist

### Pre-Implementation
- [ ] Verify working directory is correct
- [ ] Confirm on correct git branch
- [ ] Review Phase 1 interfaces
- [ ] Understand go-containerregistry remote API

### During Implementation
- [ ] Follow exact function signatures
- [ ] Add godoc comments to all exports
- [ ] Write tests alongside code
- [ ] Monitor size at checkpoints
- [ ] Handle errors comprehensively

### Post-Implementation
- [ ] Run all tests
- [ ] Check test coverage (>80%)
- [ ] Verify no TODO comments
- [ ] Measure final size
- [ ] Commit and push

## 📊 Success Metrics

### Code Quality
- Test coverage ≥ 80%
- No TODO comments
- All linting rules pass
- Clear error messages
- Comprehensive godoc

### Performance
- Push throughput >10MB/s
- Retry adds <5s overhead
- Memory usage <100MB
- No goroutine leaks

### Integration
- Phase 1 components work seamlessly
- --insecure flag functions correctly
- Feature flag controls activation
- No breaking changes

## 🔒 Security Considerations

### Must Have
- No credentials in logs
- Secure token storage
- TLS verification by default
- Explicit --insecure requirement
- No credential leakage in errors

### Must NOT Have
- Hardcoded credentials
- Silent certificate bypass
- Token persistence to disk
- Unencrypted credential transmission

## 📝 Notes for SW Engineer

### Key Points
1. This effort can run in parallel with E2.1.1 (image-builder)
2. Use stubs for image loading until E2.1.1 completes
3. Focus on push operation - no pull/delete/tag management
4. Keep retry logic simple - max 3 attempts
5. Progress reporting is important for UX

### Common Pitfalls to Avoid
- Don't implement features not in scope
- Don't add complex caching mechanisms
- Don't include comprehensive logging framework
- Don't over-engineer retry logic
- Don't forget to test --insecure mode

### Integration with Wave 1
- Your Registry interface will be used by Wave 2 CLI
- Keep interface clean and minimal
- Document all error conditions
- Provide clear progress callbacks
- Test with various image sizes

## 🏁 Deliverable Summary

### Must Deliver
1. Working Registry interface implementation
2. Successful push to Gitea registry
3. Phase 1 certificate integration
4. --insecure mode support
5. 80% test coverage
6. All files under pkg/registry/

### Success Criteria
- ✅ Pushes image to Gitea without cert errors
- ✅ --insecure flag works correctly
- ✅ Authentication succeeds with credentials
- ✅ Lists repositories successfully
- ✅ Retries transient failures
- ✅ All tests pass independently
- ✅ Under 800 lines total

---

**Remember**: 
- Monitor size continuously (stop at 700)
- Test coverage is mandatory (80% minimum)
- This is ONE atomic PR
- Can merge independently to main
- Use Phase 1 certificate infrastructure