package util

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kevinburke/ssh_config"
)

type RepoMap struct {
	repos sync.Map
}

func (r *RepoMap) LoadOrStore(repoName, dir string) *RepoState {
	v, _ := r.repos.LoadOrStore(repoName, &RepoState{Dir: dir})
	return v.(*RepoState)
}

type RepoState struct {
	MU  sync.Mutex
	Dir string
}

func NewRepoLock() *RepoMap {
	return &RepoMap{
		repos: sync.Map{},
	}
}

func RepoUrlHash(repoUrl string) string {
	sha := sha256.New()
	sha.Write([]byte(repoUrl))
	return hex.EncodeToString(sha.Sum(nil))
}

func RepoDir(repoUrl, parent string) string {
	return filepath.Join(parent, RepoUrlHash(repoUrl))
}

func FirstRemoteURL(repo *git.Repository) (string, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return "", err
	}
	if len(remotes) <= 0 {
		return "", fmt.Errorf("no remotes found")
	}
	r := remotes[0].Config().URLs
	if len(r) <= 0 {
		return "", fmt.Errorf("no remote url found")
	}
	return r[0], nil
}

// returns all files with yaml or yml suffix from a worktree
func GetWorktreeYamlFiles(parent string, wt billy.Filesystem, recurse bool) ([]string, error) {
	if strings.HasSuffix(parent, "/") {
		parent = strings.TrimSuffix(parent, "/")
	}
	paths := make([]string, 0, 10)
	ents, err := wt.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	for i := range ents {
		ent := ents[i]
		if ent.IsDir() && recurse {
			dir := fmt.Sprintf("%s/%s", parent, ent.Name())
			rPaths, dErr := GetWorktreeYamlFiles(dir, wt, recurse)
			if dErr != nil {
				return nil, fmt.Errorf("reading %s : %w", ent.Name(), dErr)
			}
			paths = append(paths, rPaths...)
		}
		if ent.Mode().IsRegular() && IsYamlFile(ent.Name()) {
			paths = append(paths, fmt.Sprintf("%s/%s", parent, ent.Name()))
		}
	}
	return paths, nil
}

func ReadWorktreeFile(wt billy.Filesystem, path string) ([]byte, error) {
	f, fErr := wt.Open(path)
	if fErr != nil {
		return nil, fmt.Errorf("opening %s", path)
	}
	defer f.Close()

	b := new(bytes.Buffer)
	_, fErr = b.ReadFrom(f)
	if fErr != nil {
		return nil, fmt.Errorf("reading %s", path)
	}

	return b.Bytes(), nil
}

func CloneRemoteRepoToMemory(ctx context.Context, remote v1alpha1.RemoteRepositorySpec, depth int, insecureSkipTLS bool) (billy.Filesystem, *git.Repository, error) {
	cloneOptions := &git.CloneOptions{
		URL:               remote.Url,
		Depth:             depth,
		ShallowSubmodules: true,
		SingleBranch:      true,
		Tags:              git.AllTags,
		InsecureSkipTLS:   insecureSkipTLS,
	}
	if remote.CloneSubmodules {
		cloneOptions.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	if remote.Ref != "" {
		cloneOptions.ReferenceName = plumbing.NewTagReferenceName(remote.Ref)
	}

	wt := memfs.New()
	var cloned *git.Repository
	cloned, err := git.CloneContext(ctx, memory.NewStorage(), wt, cloneOptions)
	if err != nil {
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(remote.Ref)
		cloned, err = git.CloneContext(ctx, memory.NewStorage(), wt, cloneOptions)
		if err != nil {
			return nil, nil, err
		}
	}
	return wt, cloned, nil
}

func CloneRemoteRepoToDir(ctx context.Context, remote v1alpha1.RemoteRepositorySpec, depth int, insecureSkipTLS bool, dir, fallbackUrl string) (billy.Filesystem, *git.Repository, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			ep, eErr := transport.NewEndpoint(remote.Url)
			if eErr != nil {
				return nil, nil, fmt.Errorf("reading endpoint %s: %w", remote.Url, eErr)
			}

			var auth transport.AuthMethod
			if ep.Protocol == "ssh" {
				a, aErr := ssh.DefaultAuthBuilder(ep.User)
				if aErr != nil {
					// go-git default auth relies on ssh agent. if not available, get from ~/.ssh/config.
					if strings.Contains(aErr.Error(), "SSH agent requested but SSH_AUTH_SOCK not-specified") {
						sshConfigPath, sErr := getSSHConfigAbsPath()
						if sErr != nil {
							return nil, nil, fmt.Errorf("getting ssh config file: %w", sErr)
						}

						au, sErr := getSSHKeyAuth(sshConfigPath, ep.Host, ep.User)
						if sErr != nil {
							return nil, nil, fmt.Errorf("ssh key auth: %w", sErr)
						}

						auth = au
					} else {
						return nil, nil, aErr
					}
				} else {
					auth = a
				}
			}

			cloneOptions := &git.CloneOptions{
				URL:               remote.Url,
				Depth:             depth,
				ShallowSubmodules: true,
				Tags:              git.AllTags,
				InsecureSkipTLS:   insecureSkipTLS,
				Auth:              auth,
			}
			if remote.CloneSubmodules {
				cloneOptions.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
			}
			repo, eErr = git.PlainCloneContext(ctx, dir, false, cloneOptions)
			if eErr != nil {
				if fallbackUrl != "" {
					cloneOptions.URL = fallbackUrl
					repo, eErr = git.PlainCloneContext(ctx, dir, false, cloneOptions)
					if eErr != nil {
						return nil, nil, fmt.Errorf("cloning repo with fall back url: %w", eErr)
					}
				}
				return nil, nil, fmt.Errorf("cloning repo: %w", eErr)
			}
		} else {
			return nil, nil, fmt.Errorf("opening repo at %s %w", dir, err)
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("getting repo worktree: %w", err)
	}
	if remote.Ref != "" {
		cErr := checkoutCommitOrRef(ctx, wt, remote.Ref)
		if cErr != nil {
			return nil, nil, fmt.Errorf("checkout %s: %w", remote.Ref, cErr)
		}
	}

	return wt.Filesystem, repo, nil
}

func CopyTreeToTree(srcWT, dstWT billy.Filesystem, srcPath, dstPath string) error {
	files, err := srcWT.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for i := range files {
		srcFile := files[i]
		fullSrcPath := filepath.Join(srcPath, srcFile.Name())
		fullDstPath := filepath.Join(dstPath, srcFile.Name())
		if srcFile.Mode().IsRegular() {
			cErr := CopyWTFile(srcWT, dstWT, fullSrcPath, fullDstPath)
			if cErr != nil {
				return cErr
			}
			continue
		}

		if srcFile.IsDir() {
			dErr := CopyTreeToTree(srcWT, dstWT, fullSrcPath, fullDstPath)
			if dErr != nil {
				return dErr
			}
		}
	}
	return nil
}

func CopyWTFile(srcWT, dstWT billy.Filesystem, srcFile, dstFile string) error {
	newFile, err := dstWT.Create(dstFile)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", dstFile, err)
	}
	defer newFile.Close()

	srcF, err := srcWT.Open(srcFile)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", srcFile, err)
	}
	defer srcF.Close()

	_, err = io.Copy(newFile, srcF)
	if err != nil {
		return fmt.Errorf("copying file %s: %w", srcFile, err)
	}
	return nil
}

// ref could be anything. Check if hash, tag, or branch in that order
func checkoutCommitOrRef(ctx context.Context, wt *git.Worktree, ref string) error {
	var refName plumbing.ReferenceName
	opts := &git.CheckoutOptions{
		Hash: plumbing.NewHash(ref),
	}

	err := wt.Checkout(opts)
	if err != nil {
		refName = plumbing.NewTagReferenceName(ref)
		opts = &git.CheckoutOptions{
			Branch: refName,
		}
		err := wt.Checkout(opts)
		if err != nil {
			refName = plumbing.NewBranchReferenceName(ref)
			opts = &git.CheckoutOptions{
				Branch: refName,
			}
			err := wt.Checkout(opts)
			if err != nil {
				return err
			}
		}
	}
	pullOpts := &git.PullOptions{
		RemoteName: "origin",
	}

	if opts.Hash.IsZero() {
		pullOpts.ReferenceName = refName
		err = wt.PullContext(ctx, pullOpts)
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("pulling latest %s: %w", ref, err)
		}
	}

	return nil
}

func getKeyfileAbsPath(relativePath string) (string, error) {
	var absPath string
	if strings.HasPrefix(relativePath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		keyFileAbs, err := filepath.Abs(filepath.Join(usr.HomeDir, relativePath[2:]))
		if err != nil {
			return "", err
		}
		absPath = keyFileAbs
	} else {
		keyFileAbs, err := filepath.Abs(relativePath)
		if err != nil {
			return "", err
		}
		absPath = keyFileAbs
	}
	return absPath, nil
}

func getSSHKeyAuth(configPath, host, user string) (transport.AuthMethod, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	conf, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	keyFileRelativePath, err := conf.Get(host, "IdentityFile")
	if err != nil {
		return nil, err
	}

	// no key specified in config, find the default key
	if keyFileRelativePath == "" {
		homeDir, hErr := getHomeDir()
		if hErr != nil {
			return nil, hErr
		}
		// from `man ssh` on Mac OpenSSH_9.7p1, LibreSSL 3.3.6
		keyFiles := []string{
			"id_rsa",
			"id_ecdsa",
			"id_ecdsa_sk",
			"id_ed25519",
			"id_ed25519_sk",
			"id_dsa",
		}
		for _, file := range keyFiles {
			path := filepath.Join(homeDir, ".ssh", file)
			if _, sErr := os.Stat(path); sErr == nil {
				keyFileRelativePath = path
				break
			}
		}
		if keyFileRelativePath == "" {
			return nil, fmt.Errorf("private key not speficied for %s. could not find default key", host)
		}
	}

	absPath, err := getKeyfileAbsPath(keyFileRelativePath)
	if err != nil {
		return nil, err
	}

	auth, err := ssh.NewPublicKeysFromFile(user, absPath, "")
	if err != nil {
		return nil, err
	}
	return auth, nil
}

func getSSHConfigAbsPath() (string, error) {
	homeDir, err := getHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(homeDir, ".ssh/config"))
}

func getHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if homeDir == "" {
		return "", fmt.Errorf("user does not have the home direcotry")
	}
	return homeDir, nil
}
