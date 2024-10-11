package util

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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

func TestGetKeyfileAbsPath(t *testing.T) {
	homeDir, _ := getHomeDir()
	cwd, _ := os.Getwd()
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"Relative path", "testkey", filepath.Join(cwd, "testkey"), false},
		{"Home directory", "~/testkey", filepath.Join(homeDir, "testkey"), false},
		{"Absolute path", "/tmp/testkey", "/tmp/testkey", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getKeyfileAbsPath(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetSSHKeyAuth(t *testing.T) {
	// Create a temporary SSH config file
	sshConfFile, err := os.CreateTemp("", "sshconfig")
	assert.NoError(t, err)
	defer os.Remove(sshConfFile.Name())

	keyPath, err := createTestPrivateKey()
	assert.NoError(t, err)
	defer os.Remove(keyPath)

	_, _ = sshConfFile.Write([]byte(fmt.Sprintf("Host testhost\nIdentityFile %s", keyPath)))
	sshConfFile.Close()

	auth, err := getSSHKeyAuth(sshConfFile.Name(), "testhost", "git")
	assert.NoError(t, err)
	assert.IsType(t, &ssh.PublicKeys{}, auth)

	_, err = getSSHKeyAuth("/nonexistent/path", "testhost", "git")
	assert.Error(t, err)

	_, err = getSSHKeyAuth(sshConfFile.Name(), "not-in-config", "git")
	assert.Error(t, err)
}

func TestGetSSHConfigAbsPath(t *testing.T) {
	expected, err := filepath.Abs(filepath.Join(os.Getenv("HOME"), ".ssh/config"))
	assert.NoError(t, err)

	result, err := getSSHConfigAbsPath()
	assert.NoError(t, err)
	assert.True(t, filepath.IsAbs(result))
	assert.Equal(t, expected, result)
}

func createTestPrivateKey() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	privKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	keyfile, err := os.CreateTemp("", "key")
	if err != nil {
		return "", err
	}
	defer keyfile.Close()

	pem.Encode(keyfile, privKeyPEM)
	return keyfile.Name(), nil
}
