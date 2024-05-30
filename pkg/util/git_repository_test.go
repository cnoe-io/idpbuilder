package util

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestCloneRemoteRepoToDir(t *testing.T) {
	spec := v1alpha1.RemoteRepositorySpec{
		CloneSubmodules: false,
		Path:            "examples/basic",
		Url:             "https://github.com/cnoe-io/idpbuilder",
		Ref:             "v0.3.0",
	}
	dir, _ := os.MkdirTemp("", "TestCopyToDir")
	defer os.RemoveAll(dir)
	// new clone
	_, _, err := CloneRemoteRepoToDir(context.Background(), spec, 0, false, dir, "")
	assert.Nil(t, err)
	testDir, _ := os.MkdirTemp("", "TestCopyToDir")
	defer os.RemoveAll(testDir)

	repo, err := git.PlainClone(testDir, false, &git.CloneOptions{URL: dir})
	assert.Nil(t, err)
	ref, err := repo.Head()
	assert.Nil(t, err)
	assert.Equal(t, "52783df3a8942cc882ebeb6168f80e1876a2f129", ref.Hash().String())

	// existing
	spec.Ref = "v0.4.0"
	testDir2, _ := os.MkdirTemp("", "TestCopyToDir")
	defer os.RemoveAll(testDir2)

	_, _, err = CloneRemoteRepoToDir(context.Background(), spec, 0, false, dir, "")
	repo, err = git.PlainClone(testDir2, false, &git.CloneOptions{URL: dir})
	assert.Nil(t, err)
	ref, err = repo.Head()
	assert.Nil(t, err)
	assert.Equal(t, "11eccd57fde9f4ef6de8bfa1fc11d168a4d30fe1", ref.Hash().String())

	assert.Nil(t, err)
}

func TestCopyTreeToTree(t *testing.T) {
	spec := v1alpha1.RemoteRepositorySpec{
		CloneSubmodules: false,
		Path:            "examples/basic",
		Url:             "https://github.com/cnoe-io/idpbuilder",
		Ref:             "",
	}

	dst := memfs.New()
	src, _, err := CloneRemoteRepoToMemory(context.Background(), spec, 1, false)
	assert.Nil(t, err)

	err = CopyTreeToTree(src, dst, spec.Path, ".")
	assert.Nil(t, err)
	testCopiedFiles(t, src, dst, spec.Path, ".")
}

func testCopiedFiles(t *testing.T, src, dst billy.Filesystem, srcStartPath, dstStartPath string) {
	files, err := src.ReadDir(srcStartPath)
	assert.Nil(t, err)

	for i := range files {
		file := files[i]
		if file.Mode().IsRegular() {
			srcB, err := ReadWorktreeFile(src, filepath.Join(srcStartPath, file.Name()))
			assert.Nil(t, err)

			dstB, err := ReadWorktreeFile(dst, filepath.Join(dstStartPath, file.Name()))
			assert.Nil(t, err)
			assert.Equal(t, srcB, dstB)
		}
		if file.IsDir() {
			testCopiedFiles(t, src, dst, filepath.Join(srcStartPath, file.Name()), filepath.Join(dstStartPath, file.Name()))
		}
	}
}

func TestGetWorktreeYamlFiles(t *testing.T) {
	filepath.Join()
	cloneOptions := &git.CloneOptions{
		URL:               "https://github.com/cnoe-io/idpbuilder",
		Depth:             1,
		ShallowSubmodules: true,
	}

	wt := memfs.New()
	_, err := git.CloneContext(context.Background(), memory.NewStorage(), wt, cloneOptions)
	if err != nil {
		t.Fatalf(err.Error())
	}

	paths, err := GetWorktreeYamlFiles("./pkg", wt, true)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, len(paths))
	for _, s := range paths {
		assert.Equal(t, true, strings.HasSuffix(s, "yaml") || strings.HasSuffix(s, "yml"))
	}

	paths, err = GetWorktreeYamlFiles("./pkg", wt, false)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(paths))
}
