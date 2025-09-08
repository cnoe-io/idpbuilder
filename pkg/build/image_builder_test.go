package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	tempDir := t.TempDir()
	builder, err := NewBuilder(tempDir)
	require.NoError(t, err)
	assert.Equal(t, tempDir, builder.storageDir)
	assert.NotNil(t, builder.images)

	// Test empty directory error
	_, err = NewBuilder("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage directory cannot be empty")
}

func TestBuildImageFeatureDisabled(t *testing.T) {
	os.Unsetenv(EnableImageBuilderFlag)
	tempDir := t.TempDir()
	builder, _ := NewBuilder(tempDir)
	contextDir := createTestContext(t)
	
	result, err := builder.BuildImage(context.Background(), BuildOptions{
		ContextPath: contextDir,
		Tag:         "test:latest",
	})
	
	assert.Error(t, err)
	assert.Equal(t, ErrFeatureDisabled, err)
	assert.Nil(t, result)
}

func TestBuildImageSuccess(t *testing.T) {
	os.Setenv(EnableImageBuilderFlag, "true")
	defer os.Unsetenv(EnableImageBuilderFlag)

	tempDir := t.TempDir()
	builder, _ := NewBuilder(tempDir)
	contextDir := createTestContext(t)

	result, err := builder.BuildImage(context.Background(), BuildOptions{
		ContextPath: contextDir,
		Tag:         "test:latest",
		Labels:      map[string]string{"test.label": "test-value"},
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ImageID)
	assert.NotEqual(t, v1.Hash{}, result.Digest)
	assert.Greater(t, result.Size, int64(0))
	assert.NotEmpty(t, result.StoragePath)

	// Verify storage file exists and tag is registered
	_, err = os.Stat(result.StoragePath)
	assert.NoError(t, err)
	
	storagePath, exists := builder.GetStoragePath("test:latest")
	assert.True(t, exists)
	assert.Equal(t, result.StoragePath, storagePath)
}

func TestBuildImageInvalidOptions(t *testing.T) {
	os.Setenv(EnableImageBuilderFlag, "true")
	defer os.Unsetenv(EnableImageBuilderFlag)

	tempDir := t.TempDir()
	builder, _ := NewBuilder(tempDir)

	// Empty context path
	_, err := builder.BuildImage(context.Background(), BuildOptions{Tag: "test:latest"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context path cannot be empty")

	// Empty tag
	_, err = builder.BuildImage(context.Background(), BuildOptions{ContextPath: tempDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tag cannot be empty")

	// Nonexistent context
	_, err = builder.BuildImage(context.Background(), BuildOptions{
		ContextPath: "/nonexistent", Tag: "test:latest",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context directory does not exist")
}

func TestGetStoragePathAndStubs(t *testing.T) {
	builder := &Builder{images: map[string]string{"test:latest": "/path/to/test.tar"}}

	// Test existing tag
	path, exists := builder.GetStoragePath("test:latest")
	assert.True(t, exists)
	assert.Equal(t, "/path/to/test.tar", path)

	// Test non-existing tag
	_, exists = builder.GetStoragePath("nonexistent")
	assert.False(t, exists)

	// Test stub methods
	assert.Nil(t, builder.ListImages())
	assert.Error(t, builder.RemoveImage("test:tag"))
	assert.Error(t, builder.TagImage("source", "target"))
}

// createTestContext creates a temporary directory with test files
func createTestContext(t *testing.T) string {
	contextDir := t.TempDir()
	os.WriteFile(filepath.Join(contextDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(contextDir, "file2.txt"), []byte("content2"), 0644)
	
	subDir := filepath.Join(contextDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "subfile.txt"), []byte("subcontent"), 0644)
	
	return contextDir
}