//go:build e2e

package docker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/tests/e2e"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func CleanUpDocker(t *testing.T) {
	t.Log("cleaning up docker env")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b, err := e2e.RunCommand(ctx, `docker ps -aqf name=localdev-control-plane`, 10*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while listing docker containers: %s, %s", err, b))

	conts := strings.TrimSpace(string(b))
	if len(conts) == 0 {
		return
	}
	b, err = e2e.RunCommand(ctx, fmt.Sprintf("docker container rm -f %s", conts), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while removing docker containers: %s, %s", err, b))

	b, err = e2e.RunCommand(ctx, "docker system prune -f", 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while pruning system: %s, %s", err, b))

	b, err = e2e.RunCommand(ctx, "docker volume prune -f", 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while pruning volumes: %s, %s", err, b))
	t.Log("finished cleaning up docker env")
}

func Test_CreateDocker(t *testing.T) {
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctrl.SetLogger(logr.FromSlogHandler(slogger.Handler()))

	testCreate(t)
	testCreatePath(t)
	testCreatePort(t)
	testCustomPkg(t)
	testPackagePriority(t)
}

// test idpbuilder create
func testCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	t.Log("running idpbuilder create")
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create")
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, b)

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)

	e2e.TestGiteaRegistry(ctx, t, "docker", fmt.Sprintf("gitea.%s", e2e.DefaultBaseDomain), e2e.DefaultPort)
}

// test idpbuilder create --use-path-routing
func testCreatePath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	t.Log("running idpbuilder create --use-path-routing")
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create", "--use-path-routing")
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://%s:%s/argocd", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://%s:%s/gitea", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)

	e2e.TestGiteaRegistry(ctx, t, "docker", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestGiteaRegistryInCluster(ctx, t, "docker", e2e.DefaultBaseDomain, e2e.DefaultPort, kubeClient)
}

func testCreatePort(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	port := "2443"
	t.Logf("running idpbuilder create --port %s", port)
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create", "--port", port)
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, port)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, port)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)
}

func testCustomPkg(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	cmdString := "create --package ../../../pkg/controllers/custompackage/test/resources/customPackages/testDir"

	t.Log(fmt.Sprintf("running %s", cmdString))
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, strings.Split(cmdString, " ")...)
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()

	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))
	if err != nil {
		assert.FailNow(t, "failed creating cluster")
	}

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)

	expectedApps := map[string]string{
		"my-app":  "argocd",
		"my-app2": "argocd",
	}
	e2e.TestArgoCDApps(ctx, t, kubeClient, expectedApps)
	repos, err := e2e.GetGiteaRepos(ctx, giteaBaseUrl)
	assert.NoError(t, err)
	expectedRepoNames := map[string]struct{}{
		"idpbuilder-localdev-my-app-app1":  {},
		"idpbuilder-localdev-my-app2-app2": {},
	}

	for i := range repos {
		repo := repos[i]
		_, ok := expectedRepoNames[repo.Name]
		if ok {
			delete(expectedRepoNames, repo.Name)
		}
	}
	assert.Empty(t, expectedRepoNames)
}

// testPackagePriority tests the priority-based package reconciliation feature
// where multiple packages for the same app can be specified, and only the highest priority wins
func testPackagePriority(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	// Create with multiple package directories
	// The packages will be assigned priorities based on their order (0, 1, 2, ...)
	cmdString := "create --package ../../../pkg/controllers/custompackage/test/resources/customPackages/testDir --package ../../../pkg/controllers/custompackage/test/resources/customPackages/testDir2"

	t.Log(fmt.Sprintf("running %s to test package priority", cmdString))
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, strings.Split(cmdString, " ")...)
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))
	if err != nil {
		assert.FailNow(t, "failed creating cluster")
	}

	// Wait for core packages to be ready
	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	// Verify CustomPackages have priority annotations
	t.Log("Verifying CustomPackages have correct priority annotations")

	customPkgList := &e2e.CustomPackageList{}
	err = kubeClient.List(ctx, customPkgList, &e2e.ListOptions{Namespace: "idpbuilder-localdev"})
	assert.NoError(t, err, "failed to list custom packages")

	// Verify that packages have priority annotations
	foundPriorities := make(map[string]string)
	for _, pkg := range customPkgList.Items {
		if pkg.ObjectMeta.Annotations != nil {
			if priority, exists := pkg.ObjectMeta.Annotations["cnoe.io/package-priority"]; exists {
				t.Logf("Package %s has priority: %s", pkg.Name, priority)
				foundPriorities[pkg.Name] = priority
			}
			if sourcePath, exists := pkg.ObjectMeta.Annotations["cnoe.io/package-source-path"]; exists {
				t.Logf("Package %s has source path: %s", pkg.Name, sourcePath)
			}
		}
	}

	// At least some packages should have priority annotations
	assert.NotEmpty(t, foundPriorities, "expected custom packages to have priority annotations")

	// Wait for custom packages to reconcile
	time.Sleep(10 * time.Second)

	// Verify apps are deployed
	expectedApps := map[string]string{
		"my-app":    "argocd",
		"guestbook": "argocd",
	}
	e2e.TestArgoCDApps(ctx, t, kubeClient, expectedApps)

	t.Log("Package priority test completed successfully")
}
