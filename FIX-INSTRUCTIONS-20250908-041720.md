# Fix Instructions for E2.1.1 Image Builder

## 🔴 CRITICAL: R320 Violation Must Be Fixed Immediately

### Priority 1: Remove ALL Stub Implementations (BLOCKING)

**File**: `pkg/build/image_builder.go`
**Lines**: 160-176

**Required Action**: DELETE the following methods entirely:
```go
// DELETE THESE METHODS - DO NOT LEAVE AS STUBS:

// ListImages returns all built image tags (stub for future implementation)
func (b *Builder) ListImages() []string {
	// Stub implementation - not implemented in this effort per R311
	return nil
}

// RemoveImage removes a built image (stub for future implementation)
func (b *Builder) RemoveImage(tag string) error {
	// Stub implementation - not implemented in this effort per R311
	return fmt.Errorf("not implemented: RemoveImage will be implemented in future effort")
}

// TagImage creates a new tag for an existing image (stub for future implementation)
func (b *Builder) TagImage(sourceTag, targetTag string) error {
	// Stub implementation - not implemented in this effort per R311
	return fmt.Errorf("not implemented: TagImage will be implemented in future effort")
}
```

**Why**: R320 has ZERO TOLERANCE for stub implementations. Any "not implemented" error or placeholder return violates this supreme law and results in automatic review failure.

**Test Update Required**: After removing these methods, also remove the test for them:
- File: `pkg/build/image_builder_test.go`
- Lines: 111-113 (remove the stub method tests)

### Priority 2: Increase Test Coverage to 80% (MAJOR)

Current coverage is 47.9%, which is well below the 80% requirement.

**Add the following tests to `pkg/build/context_test.go`**:

1. **Test for exclusion edge cases**:
```go
func TestCreateTarFromContextEdgeCases(t *testing.T) {
    // Test with empty directory
    emptyDir := t.TempDir()
    reader, err := createTarFromContext(emptyDir, nil)
    require.NoError(t, err)
    reader.Close()
    
    // Test with symbolic links
    // Test with special characters in filenames
    // Test with deeply nested directories
}
```

2. **Test for concurrent tar creation**:
```go
func TestConcurrentTarCreation(t *testing.T) {
    contextDir := createTestContext(t)
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            reader, err := createTarFromContext(contextDir, nil)
            assert.NoError(t, err)
            reader.Close()
        }()
    }
    wg.Wait()
}
```

**Add the following tests to `pkg/build/storage_test.go`** (create new file):

```go
package build

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/google/go-containerregistry/pkg/v1/empty"
)

func TestSaveImageLocallyErrors(t *testing.T) {
    // Test nil image
    _, err := saveImageLocally(nil, "test:latest", t.TempDir())
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "image cannot be nil")
    
    // Test empty tag
    _, err = saveImageLocally(empty.Image, "", t.TempDir())
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "tag cannot be empty")
    
    // Test invalid storage directory
    _, err = saveImageLocally(empty.Image, "test:latest", "/root/nopermission")
    assert.Error(t, err)
}

func TestSanitizeTagForFilename(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"image:latest", "image_latest"},
        {"registry/repo/image:v1.0", "registry_repo_image_v1.0"},
        {"image<>|?*:tag", "image______tag"},
        {"", "unnamed"},
        {strings.Repeat("a", 200), strings.Repeat("a", 100)},
    }
    
    for _, tt := range tests {
        result := sanitizeTagForFilename(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}
```

**Add more edge case tests to `pkg/build/image_builder_test.go`**:

```go
func TestBuildImageWithExclusions(t *testing.T) {
    os.Setenv(EnableImageBuilderFlag, "true")
    defer os.Unsetenv(EnableImageBuilderFlag)
    
    tempDir := t.TempDir()
    builder, _ := NewBuilder(tempDir)
    contextDir := createTestContext(t)
    
    // Create files to exclude
    os.WriteFile(filepath.Join(contextDir, ".git"), []byte("git"), 0644)
    os.WriteFile(filepath.Join(contextDir, "temp.swp"), []byte("swap"), 0644)
    
    result, err := builder.BuildImage(context.Background(), BuildOptions{
        ContextPath: contextDir,
        Tag:         "test:excluded",
        Exclusions:  []string{".git", "*.swp"},
    })
    
    require.NoError(t, err)
    assert.NotNil(t, result)
    // Verify excluded files are not in the image
}
```

### Priority 3: Security Enhancements (MINOR)

While not blocking, consider adding these improvements:

1. **Add tar size limits in `context.go`**:
```go
const MaxTarSize = 1 << 30 // 1GB limit

// In createTarFromContext, track total size:
var totalSize int64
// ... in the walk function:
totalSize += info.Size()
if totalSize > MaxTarSize {
    return fmt.Errorf("context size exceeds maximum allowed (%d bytes)", MaxTarSize)
}
```

## Verification Steps

After making these fixes:

1. **Verify no stubs remain**:
```bash
grep -r "not implemented\|TODO\|unimplemented" --include="*.go" ./pkg/build/
# Should return nothing
```

2. **Verify test coverage**:
```bash
go test ./pkg/build -cover
# Should show >= 80% coverage
```

3. **Verify tests pass**:
```bash
go test ./pkg/build -v
# All tests should pass
```

4. **Verify build**:
```bash
go build ./pkg/build
# Should build without errors
```

## Summary of Required Changes

1. **DELETE** all stub methods from `image_builder.go` (lines 160-176)
2. **REMOVE** stub method tests from `image_builder_test.go` (lines 111-113)
3. **ADD** comprehensive tests to reach 80% coverage
4. **CONSIDER** adding security limits for tar operations

Once these changes are complete, the implementation will comply with all requirements and be ready for integration.

---
**Instructions Created**: 2025-09-08 04:17:20 UTC
**For**: SW Engineer implementing fixes
**Priority**: CRITICAL - R320 violation blocks all progress