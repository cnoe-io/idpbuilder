package build

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTarFromContext(t *testing.T) {
	contextDir := createTestContext(t)
	tarReader, err := createTarFromContext(contextDir, nil)
	require.NoError(t, err)
	defer tarReader.Close()

	// Read tar contents and verify expected files
	tr := tar.NewReader(tarReader)
	files := make(map[string]bool)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[header.Name] = true
	}

	expectedFiles := []string{"file1.txt", "file2.txt", "subdir", "subdir/subfile.txt"}
	for _, expected := range expectedFiles {
		assert.True(t, files[expected])
	}
}

func TestCreateTarFromContextNonexistent(t *testing.T) {
	_, err := createTarFromContext("/nonexistent/path", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context directory does not exist")
}

func TestCreateTarFromContextWithExclusions(t *testing.T) {
	contextDir := t.TempDir()
	os.WriteFile(filepath.Join(contextDir, "file.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(contextDir, "exclude.tmp"), []byte("excluded"), 0644)
	
	tempDir := filepath.Join(contextDir, "temp")
	os.MkdirAll(tempDir, 0755)
	os.WriteFile(filepath.Join(tempDir, "tempfile.txt"), []byte("temp"), 0644)

	tarReader, err := createTarFromContext(contextDir, []string{"*.tmp", "temp/*"})
	require.NoError(t, err)
	defer tarReader.Close()

	tr := tar.NewReader(tarReader)
	files := make(map[string]bool)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[header.Name] = true
	}

	assert.True(t, files["file.txt"])
	assert.False(t, files["exclude.tmp"])
	assert.False(t, files["temp/tempfile.txt"])
}

func TestShouldExclude(t *testing.T) {
	assert.False(t, shouldExclude("file.txt", nil))
	assert.True(t, shouldExclude("file.txt", []string{"file.txt"}))
	assert.True(t, shouldExclude("file.tmp", []string{"*.tmp"}))
	assert.False(t, shouldExclude("file.txt", []string{"*.tmp"}))
	assert.True(t, shouldExclude("temp/file.txt", []string{"temp"}))
	assert.True(t, shouldExclude("src/node_modules/pkg.json", []string{"node_modules"}))
}