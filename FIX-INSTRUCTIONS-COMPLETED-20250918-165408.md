# Fix Instructions for registry-auth (E1.1.2B)

## 🔴 CRITICAL FIXES REQUIRED

### Priority 1: Add Unit Tests (BLOCKING)

**Issue**: Complete absence of unit tests for security-critical auth package
**Severity**: CRITICAL - Cannot approve without tests
**Target Coverage**: Minimum 80%

#### Required Test Files

1. **Create `pkg/registry/auth/authenticator_test.go`**
```go
// Test the factory function
func TestNewAuthenticator(t *testing.T) {
    // Test basic auth creation
    // Test token auth creation
    // Test no-op auth creation
    // Test unsupported auth type error
}

// Test NoOpAuthenticator
func TestNoOpAuthenticator(t *testing.T) {
    // Test Authenticate (should do nothing)
    // Test Refresh (should do nothing)
    // Test IsValid (should return true)
}
```

2. **Create `pkg/registry/auth/basic_test.go`**
```go
func TestBasicAuthenticator(t *testing.T) {
    // Test NewBasicAuthenticator with valid credentials
    // Test NewBasicAuthenticator with missing credentials (error case)
    // Test Authenticate adds correct Authorization header
    // Test encoded credential format
    // Test IsValid returns true when encoded exists
}
```

3. **Create `pkg/registry/auth/token_test.go`**
```go
func TestTokenAuthenticator(t *testing.T) {
    // Test NewTokenAuthenticator with direct token
    // Test NewTokenAuthenticator with token client
    // Test Authenticate with valid token
    // Test Authenticate triggers refresh when no token
    // Test Refresh with mock TokenClient
    // Test IsValid with expired token
    // Test IsValid with valid token
    // Test concurrent access (race conditions)
}
```

4. **Create `pkg/registry/auth/middleware_test.go`**
```go
func TestTransport(t *testing.T) {
    // Test RoundTrip with valid auth
    // Test RoundTrip with no auth
    // Test 401 retry logic
    // Test request cloning (original not modified)
    // Test auth refresh on 401
    // Test error propagation
}
```

5. **Create `pkg/registry/auth/manager_test.go`**
```go
func TestManager(t *testing.T) {
    // Test GetAuthenticator creates new auth
    // Test GetAuthenticator returns cached auth
    // Test GetAuthenticator refreshes invalid auth
    // Test Clear removes specific auth
    // Test ClearAll removes all auths
    // Test concurrent access to cache
}
```

### Priority 2: Verify/Implement Dependencies

**Issue**: Missing TokenClient implementation
**Location**: token.go references TokenClient interface

#### Actions Required

1. **Check if TokenClient exists in registry-types effort**:
```bash
cd ../registry-types
grep -r "type TokenClient" .
```

2. **If not found, add mock implementation for testing**:
```go
// In token_test.go
type mockTokenClient struct {
    token     string
    expiresIn int
    err       error
}

func (m *mockTokenClient) RequestToken(ctx context.Context, config *types.AuthConfig) (*types.TokenResponse, error) {
    if m.err != nil {
        return nil, m.err
    }
    return &types.TokenResponse{
        Token:     m.token,
        ExpiresIn: m.expiresIn,
    }, nil
}
```

3. **Verify types.CredentialStore interface exists**:
```bash
cd ../registry-types
grep -r "type CredentialStore" .
```

### Priority 3: Add Documentation

**Issue**: Missing package-level documentation

1. **Add to each file header**:
```go
// Package auth provides authentication implementations for OCI registry operations.
// It supports multiple authentication methods including basic auth, bearer tokens,
// and anonymous access.
package auth
```

2. **Create `pkg/registry/auth/README.md`**:
```markdown
# Registry Authentication

This package provides authentication for OCI registry operations.

## Supported Authentication Methods
- Basic Authentication (username/password)
- Bearer Token Authentication
- Anonymous (no authentication)

## Usage
[Add usage examples]
```

## 📋 Implementation Checklist

### Must Complete
- [ ] Add authenticator_test.go with factory tests
- [ ] Add basic_test.go with auth header tests
- [ ] Add token_test.go with refresh logic tests
- [ ] Add middleware_test.go with 401 retry tests
- [ ] Add manager_test.go with caching tests
- [ ] Achieve minimum 80% test coverage
- [ ] Verify all tests pass
- [ ] Add package documentation

### Should Verify
- [ ] Confirm TokenClient interface location
- [ ] Confirm CredentialStore interface exists
- [ ] Add integration test if time permits

## 🧪 Testing Commands

Run these after adding tests:
```bash
# Run all tests
cd efforts/phase1/wave1/registry-auth
go test ./pkg/registry/auth/... -v

# Check coverage
go test ./pkg/registry/auth/... -cover

# Generate coverage report
go test ./pkg/registry/auth/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## ⏰ Estimated Time

- Adding all test files: 2-3 hours
- Verifying dependencies: 30 minutes
- Adding documentation: 30 minutes
- **Total**: 3-4 hours

## 🎯 Success Criteria

Your fixes will be approved when:
1. ✅ All test files created and passing
2. ✅ Test coverage ≥ 80%
3. ✅ Dependencies verified or mocked
4. ✅ Package documentation added
5. ✅ All existing code still works

## 📝 Notes

- Focus on test coverage first - this is the blocking issue
- The implementation itself is good, just needs tests
- Use table-driven tests where appropriate
- Don't modify the existing implementation unless fixing bugs

---

*Fix instructions created by Code Reviewer Agent*
*Priority: Add tests immediately to unblock approval*