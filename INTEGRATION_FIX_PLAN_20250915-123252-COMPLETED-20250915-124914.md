# Fix Plan: Remove Stub Implementations from gitea-client-split-002

## R320 VIOLATION - CRITICAL BLOCKER

### Issue Analysis

The file `pkg/registry/stubs.go` contains mock implementations that should not be in production code. While these are testing mocks rather than true stubs, they violate the principle of keeping test code separate from production code.

### Stubs/Mocks Identified

1. **MockRegistry** (lines 12-156)
   - Full mock implementation of Registry interface
   - Contains test-specific methods like InjectError, InjectDelay
   - Should be in a test file or separate test package

2. **MockRepository** (lines 18-23)
   - Test data structure for mock registry
   - Not for production use

3. **MockImage** (lines 25-30)
   - Test data structure for mock images
   - Not for production use

4. **TestHelper** (lines 157-175)
   - Pure test helper functionality
   - Should be in test files only

### Root Cause Analysis

The mock implementations were placed in `pkg/registry/stubs.go` instead of:
1. A `*_test.go` file (which would exclude them from production builds)
2. A separate test package like `pkg/registry/testing/`
3. A dedicated test utilities directory

This violates Go best practices where test helpers should be in test files or test packages.

### Required Fixes

#### 1. Move Mock Implementations to Test Files

**Action**: Relocate all mock implementations to proper test locations

**Step 1**: Create `pkg/registry/mocks_test.go`
```go
package registry

// Move all mock types here from stubs.go
// This ensures they're only compiled for tests
```

**Step 2**: Delete `pkg/registry/stubs.go` entirely
```bash
rm pkg/registry/stubs.go
```

#### 2. Alternative: Create Test Package

If mocks need to be shared across packages:

**Step 1**: Create test package structure
```bash
mkdir -p pkg/registry/testing
```

**Step 2**: Create `pkg/registry/testing/mocks.go`
```go
package testing

import (
    "context"
    "io"
    // ... other imports
    "github.com/idpbuilder/gitea-client/pkg/registry"
)

// Move all mock implementations here
// Export them for use in tests across packages
```

#### 3. Update Import References

Find and update all test files that import the mocks:

```bash
# Find all test files using the mocks
grep -r "MockRegistry\|TestHelper" --include="*_test.go"
```

Update imports from:
```go
// Old
registry.NewMockRegistry()
```

To:
```go
// New (if in same package test file)
NewMockRegistry()

// Or (if in testing package)
import "github.com/idpbuilder/gitea-client/pkg/registry/testing"
testing.NewMockRegistry()
```

### Implementation Steps

1. **Analyze Current Usage**
   ```bash
   cd efforts/phase2/wave1/gitea-client-split-002
   grep -r "MockRegistry\|TestHelper\|MockRepository\|MockImage" --include="*.go"
   ```

2. **Create New Test File**
   ```bash
   cd pkg/registry
   # Move content from stubs.go to mocks_test.go
   mv stubs.go mocks_test.go
   # Edit file to ensure package is still "registry"
   ```

3. **Update Test Imports**
   - Find all test files using these mocks
   - Ensure they can still access the mock implementations

4. **Verify Build**
   ```bash
   # Ensure production build excludes test code
   go build -v ./...

   # Verify tests still work
   go test ./...
   ```

5. **Verify No Stubs in Production**
   ```bash
   # Check compiled binary doesn't include test code
   go build -o app ./...
   go tool nm app | grep -i mock
   # Should return nothing
   ```

### Verification Steps

1. **Build Verification**
   ```bash
   cd efforts/phase2/wave1/gitea-client-split-002
   go build ./...  # Must succeed
   ```

2. **Test Verification**
   ```bash
   go test ./...  # All tests must pass
   ```

3. **Production Build Check**
   ```bash
   # Ensure mocks aren't in production binary
   go build -o gitea-client ./...
   strings gitea-client | grep -i "MockRegistry\|InjectError\|TestHelper"
   # Should return nothing
   ```

4. **R320 Compliance Check**
   ```bash
   # Verify no stub patterns
   grep -r "not.*implemented\|TODO\|unimplemented" --include="*.go" --exclude="*_test.go"
   # Should return nothing concerning
   ```

### Additional Checks

#### Check for Other Issues

1. **TODO Comments**: Found legitimate TODO comments that should be addressed:
   - `pkg/cmd/get/clusters.go`: Uses `context.TODO()` - should use proper context
   - `pkg/cmd/get/packages.go`: TODO comment about LocalBuild assumption
   - `pkg/util/idp.go`: TODO comment about LocalBuild assumption
   - `pkg/controllers/gitrepository/controller.go`: TODO about using notifyChan

2. **Context.TODO() Usage**: Replace with proper context propagation:
   ```go
   // Instead of:
   err = cli.List(context.TODO(), &nodeList)

   // Use:
   err = cli.List(ctx, &nodeList)  // Pass context from function parameter
   ```

### Estimated Time

- Moving mocks to test files: 15 minutes
- Updating imports: 10 minutes
- Testing and verification: 15 minutes
- Fixing context.TODO() issues: 20 minutes
**Total: 60 minutes**

### Priority

**CRITICAL** - This is an R320 violation that blocks integration

### Success Criteria

✅ No `stubs.go` file in production code
✅ All mocks in `*_test.go` files or test packages
✅ Production build contains no test/mock code
✅ All tests still pass
✅ No "not implemented" patterns in non-test code
✅ Context.TODO() replaced with proper contexts