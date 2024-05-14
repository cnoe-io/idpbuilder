package gitrepository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const addFileContent = "added\n"

type mockGitea struct {
	GiteaClient
	getRepo    func() (*gitea.Repository, *gitea.Response, error)
	createRepo func() (*gitea.Repository, *gitea.Response, error)
}

func (g mockGitea) SetBasicAuth(user, pass string) {}

func (g mockGitea) SetContext(ctx context.Context) {}

func (g mockGitea) CreateOrgRepo(org string, opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	if g.createRepo != nil {
		return g.createRepo()
	}
	return &gitea.Repository{}, &gitea.Response{}, nil
}

func (g mockGitea) GetRepo(owner, reponame string) (*gitea.Repository, *gitea.Response, error) {
	if g.getRepo != nil {
		return g.getRepo()
	}
	return &gitea.Repository{}, &gitea.Response{}, nil
}

type expect struct {
	resource v1alpha1.GitRepositoryStatus
	err      error
}

type testCase struct {
	giteaClient GiteaClient
	input       v1alpha1.GitRepository
	expect      expect
}

func (t testCase) giteaProvider(ctx context.Context, repo *v1alpha1.GitRepository, kubeClient client.Client, scheme *runtime.Scheme, tmplConfig util.CorePackageTemplateConfig) (gitProvider, error) {
	return &giteaProvider{
		Client:      kubeClient,
		Scheme:      scheme,
		giteaClient: t.giteaClient,
		config:      tmplConfig,
	}, nil
}

type fakeClient struct {
	client.Client
	patchObj client.Object
}

func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	s := obj.(*v1.Secret)
	s.Data = map[string][]byte{
		giteaAdminUsernameKey: []byte("abc"),
		giteaAdminPasswordKey: []byte("abc"),
	}
	return nil
}

func (f *fakeClient) Status() client.StatusWriter {
	return fakeStatusWriter{}
}

func (f *fakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	f.patchObj = obj
	return nil
}

type fakeStatusWriter struct {
	client.StatusWriter
}

func (f fakeStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return nil
}

func setUpLocalRepo() (string, string, error) {
	repoDir, err := os.MkdirTemp("", fmt.Sprintf("test"))
	if err != nil {
		return "", "", fmt.Errorf("creating temporary directory: %w", err)
	}
	// create a repo for pushing. MUST BE BARE
	repo, err := git.PlainInit(repoDir, true)
	if err != nil {
		return "", "", fmt.Errorf("repo init: %w", err)
	}

	// init it with a static file (in-memory), set default branch name, then get the hash
	defaultBranchName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", DefaultBranchName))

	repoConfig, _ := repo.Config()
	repoConfig.Init.DefaultBranch = DefaultBranchName
	repo.SetConfig(repoConfig)

	h := plumbing.NewSymbolicReference(plumbing.HEAD, defaultBranchName)
	repo.Storer.SetReference(h)

	fileObject := plumbing.MemoryObject{}
	fileObject.SetType(plumbing.BlobObject)
	w, _ := fileObject.Writer()

	file, err := os.ReadFile("test/resources/file1")
	if err != nil {
		return "", "", fmt.Errorf("reading file from resources dir: %w", err)
	}
	w.Write(file)
	w.Close()

	fileHash, _ := repo.Storer.SetEncodedObject(&fileObject)

	treeEntry := object.TreeEntry{
		Name: "file1",
		Mode: filemode.Regular,
		Hash: fileHash,
	}

	tree := object.Tree{
		Entries: []object.TreeEntry{treeEntry},
	}

	treeObject := plumbing.MemoryObject{}
	tree.Encode(&treeObject)

	initHash, _ := repo.Storer.SetEncodedObject(&treeObject)

	commit := object.Commit{
		Author: object.Signature{
			Name:  gitCommitAuthorName,
			Email: gitCommitAuthorEmail,
			When:  time.Now(),
		},
		Message:  "init",
		TreeHash: initHash,
	}

	commitObject := plumbing.MemoryObject{}
	commit.Encode(&commitObject)

	commitHash, _ := repo.Storer.SetEncodedObject(&commitObject)

	repo.Storer.SetReference(plumbing.NewHashReference(defaultBranchName, commitHash))

	return repoDir, commitHash.String(), nil
}

func setupDir() (string, error) {
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("test"))
	if err != nil {
		return "", fmt.Errorf("creating temporary directory: %w", err)
	}

	file, err := os.ReadFile("test/resources/file1")
	if err != nil {
		return "", fmt.Errorf("reading file from resources dir: %w", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "file1"), file, 0644)
	if err != nil {
		return "", fmt.Errorf("writing file to temp dir: %w", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "add"), []byte(addFileContent), 0644)
	if err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return tempDir, nil
}

func TestGitRepositoryContentReconcile(t *testing.T) {
	ctx := context.Background()
	localRepoDir, _, err := setUpLocalRepo()
	defer os.RemoveAll(localRepoDir)
	if err != nil {
		t.Fatalf("failed setting up local git repo: %v", err)
	}

	srcDir, err := setupDir()
	defer os.RemoveAll(srcDir)
	if err != nil {
		t.Fatalf("failed to set up dirs: %v", err)
	}

	testCloneDir, _ := os.MkdirTemp("", "gitrepo-test")
	defer os.RemoveAll(testCloneDir)

	m := metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
	}
	resource := v1alpha1.GitRepository{
		ObjectMeta: m,
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Path: srcDir,
				Type: "local",
			},
		},
	}

	t.Run("files modified", func(t *testing.T) {
		p := giteaProvider{
			Client:      &fakeClient{},
			giteaClient: mockGitea{},
		}
		// add file to source directory, reconcile, clone the repo and check if the added file exists
		err = p.updateRepoContent(ctx, &resource, repoInfo{cloneUrl: localRepoDir}, gitProviderCredentials{}, testCloneDir, util.NewRepoLock())
		if err != nil {
			t.Fatalf("failed adding %v", err)
		}

		repo, _ := git.PlainClone(testCloneDir, false, &git.CloneOptions{
			URL: localRepoDir,
		})
		c, err := os.ReadFile(filepath.Join(testCloneDir, "add"))
		if err != nil {
			t.Fatalf("failed to read file at %s. %v", filepath.Join(testCloneDir, "add"), err)
		}
		if string(c) != addFileContent {
			t.Fatalf("expected %s, got %s", addFileContent, c)
		}

		// remove added file, reconcile, pull, check if the file is removed
		err = os.Remove(filepath.Join(srcDir, "add"))
		if err != nil {
			t.Fatalf("failed to remove added file %v", err)
		}
		err = p.updateRepoContent(ctx, &resource, repoInfo{cloneUrl: localRepoDir}, gitProviderCredentials{}, testCloneDir, util.NewRepoLock())
		if err != nil {
			t.Fatalf("failed removing %v", err)
		}

		w, _ := repo.Worktree()
		err = w.Pull(&git.PullOptions{})
		if err != nil {
			t.Fatalf("failed pulling changes %v", err)
		}

		_, err = os.Stat(filepath.Join(testCloneDir, "add"))
		if err == nil {
			t.Fatalf("file should not exist")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("received unexpected error %v", err)
		}
	})
}

func TestGitRepositoryContentReconcileEmbedded(t *testing.T) {
	ctx := context.Background()
	localRepoDir, _, err := setUpLocalRepo()
	defer os.RemoveAll(localRepoDir)
	if err != nil {
		t.Fatalf("failed setting up local git repo: %v", err)
	}

	tmpDir, _ := os.MkdirTemp("", "add")
	defer os.RemoveAll(tmpDir)

	m := metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
	}
	resource := v1alpha1.GitRepository{
		ObjectMeta: m,
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				EmbeddedAppName: "nginx",
				Type:            "embedded",
			},
			Provider: v1alpha1.Provider{
				InternalGitURL: "http://cnoe.io",
			},
		},
	}

	t.Run("should update content", func(t *testing.T) {
		p := giteaProvider{
			Client:      &fakeClient{},
			giteaClient: mockGitea{},
		}
		err = p.updateRepoContent(ctx, &resource, repoInfo{cloneUrl: localRepoDir}, gitProviderCredentials{}, tmpDir, util.NewRepoLock())
		if err != nil {
			t.Fatalf("failed adding %v", err)
		}
	})
}

func TestGitRepositoryReconcile(t *testing.T) {
	localReoDir, hash, err := setUpLocalRepo()
	defer os.RemoveAll(localReoDir)
	if err != nil {
		t.Fatalf("failed setting up local git repo: %v", err)
	}
	resourcePath, err := filepath.Abs("./test/resources")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	updateDir, _, _ := setUpLocalRepo()
	defer os.RemoveAll(updateDir)

	addDir, err := setupDir()
	fmt.Println(addDir)
	defer os.RemoveAll(addDir)
	if err != nil {
		t.Fatalf("failed to set up dirs: %v", err)
	}

	tmpDir, _ := os.MkdirTemp("", "gitrepo-test")
	defer os.RemoveAll(tmpDir)

	m := metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
	}

	cases := map[string]testCase{
		"no op": {
			giteaClient: mockGitea{
				getRepo: func() (*gitea.Repository, *gitea.Response, error) {
					return &gitea.Repository{CloneURL: localReoDir}, nil, nil
				},
			},
			input: v1alpha1.GitRepository{
				ObjectMeta: m,
				Spec: v1alpha1.GitRepositorySpec{
					Source: v1alpha1.GitRepositorySource{
						Path: resourcePath,
						Type: "local",
					},
					Provider: v1alpha1.Provider{
						Name:           v1alpha1.GitProviderGitea,
						InternalGitURL: "http://cnoe.io",
					},
				},
			},
			expect: expect{
				resource: v1alpha1.GitRepositoryStatus{
					ExternalGitRepositoryUrl: localReoDir,
					LatestCommit:             v1alpha1.Commit{Hash: hash},
					Synced:                   true,
					InternalGitRepositoryUrl: "http://cnoe.io/giteaAdmin/test-test.git",
				},
			},
		},
		"update": {
			giteaClient: mockGitea{
				getRepo: func() (*gitea.Repository, *gitea.Response, error) {
					return &gitea.Repository{CloneURL: updateDir}, nil, nil
				},
			},
			input: v1alpha1.GitRepository{
				ObjectMeta: m,
				Spec: v1alpha1.GitRepositorySpec{
					Source: v1alpha1.GitRepositorySource{
						Path: addDir,
						Type: "local",
					},
					Provider: v1alpha1.Provider{
						Name:           v1alpha1.GitProviderGitea,
						InternalGitURL: "http://cnoe.io",
					},
				},
			},
			expect: expect{
				resource: v1alpha1.GitRepositoryStatus{
					ExternalGitRepositoryUrl: updateDir,
					Synced:                   true,
					InternalGitRepositoryUrl: "http://cnoe.io/giteaAdmin/test-test.git",
				},
			},
		},
	}

	ctx := context.Background()

	t.Run("repo updates", func(t *testing.T) {
		for k := range cases {
			v := cases[k]
			r := RepositoryReconciler{
				Client:          &fakeClient{},
				GitProviderFunc: v.giteaProvider,
				TempDir:         tmpDir,
				RepoMap:         util.NewRepoLock(),
			}
			_, err := r.reconcileGitRepo(ctx, &v.input)
			if v.expect.err == nil && err != nil {
				t.Fatalf("failed %s: %v", k, err)
			}

			if v.expect.resource.LatestCommit.Hash == "" {
				v.expect.resource.LatestCommit.Hash = v.input.Status.LatestCommit.Hash
			}
			assert.Equal(t, v.input.Status, v.expect.resource)
		}
	})
}

func TestGitRepositoryPostReconcile(t *testing.T) {
	c := fakeClient{}
	tmpDir, _ := os.MkdirTemp("", "repo-updates-test")
	defer os.RemoveAll(tmpDir)
	reconciler := RepositoryReconciler{
		Client:  &c,
		TempDir: tmpDir,
		RepoMap: util.NewRepoLock(),
	}
	testTime := time.Now().Format(time.RFC3339Nano)
	repo := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Annotations: map[string]string{
				v1alpha1.CliStartTimeAnnotation: testTime,
			},
		},
	}

	reconciler.postProcessReconcile(context.Background(), ctrl.Request{}, &repo)
	annotations := c.patchObj.GetAnnotations()
	v, ok := annotations[v1alpha1.LastObservedCLIStartTimeAnnotation]
	if !ok {
		t.Fatalf("expected annotation not found: %s", v1alpha1.LastObservedCLIStartTimeAnnotation)
	}
	if v != testTime {
		t.Fatalf("annotation values does not match")
	}

	repo.Annotations[v1alpha1.LastObservedCLIStartTimeAnnotation] = "abc"
	reconciler.postProcessReconcile(context.Background(), ctrl.Request{}, &repo)
	v = annotations[v1alpha1.LastObservedCLIStartTimeAnnotation]
	if v != testTime {
		t.Fatalf("annotation values does not match")
	}
}
