//go:build e2e

package docker

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/tests/e2e/container"
	"github.com/cnoe-io/idpbuilder/tests/e2e/shared"
	"github.com/go-logr/logr"
	"log/slog"
	"os"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"testing"
	"time"
)

type DockerEngine struct {
	container.Engine
	Client string
}

func (p *DockerEngine) GetClient() string {
	return p.Client
}

func (p *DockerEngine) IdpCmd() *exec.Cmd {
	return exec.Command(container.IdpbuilderBinaryLocation)
}

func CleanUpDocker(t *testing.T) {
	t.Log("cleaning up docker env")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b, err := e2e.RunCommand(ctx, fmt.Sprintf("%s ps -aqf name=localdev-control-plane", containerClientEngine), 10*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while listing docker containers: %s, %s", err, b))

	conts := strings.TrimSpace(string(b))
	if len(conts) == 0 {
		return
	}
	b, err = e2e.RunCommand(ctx, fmt.Sprintf("%s container rm -f %s", containerClientEngine, conts), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while removing docker containers: %s, %s", err, b))

	b, err = e2e.RunCommand(ctx, fmt.Sprintf("%s system prune -f", containerClientEngine), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while pruning system: %s, %s", err, b))

	b, err = e2e.RunCommand(ctx, fmt.Sprintf("%s volume prune -f", containerClientEngine), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while pruning volumes: %s, %s", err, b))
	t.Log("finished cleaning up docker env")
}

func Test_CreateDocker(t *testing.T) {
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctrl.SetLogger(logr.FromSlogHandler(slogger.Handler()))

	containerEngine := &DockerEngine{Client: "docker"}
	shared.testCreate(t, containerEngine)
	shared.testCreatePath(t, containerEngine)
	shared.testCreatePort(t, containerEngine)
	shared.testCustomPkg(t, containerEngine)
}

// test idpbuilder create
func testCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	t.Log("running idpbuilder create")
	cmd = exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create")
	cmd = enableKindExperimentalProvider(*cmd)
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, b)

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)

	e2e.TestGiteaRegistry(ctx, t, containerClientEngine, fmt.Sprintf("gitea.%s", e2e.DefaultBaseDomain), e2e.DefaultPort)
}

// test idpbuilder create --use-path-routing
func testCreatePath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	t.Log("running idpbuilder create --use-path-routing")
	cmd = exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create", "--use-path-routing")
	cmd = enableKindExperimentalProvider(*cmd)
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://%s:%s/argocd", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://%s:%s/gitea", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)

	e2e.TestGiteaRegistry(ctx, t, containerClientEngine, e2e.DefaultBaseDomain, e2e.DefaultPort)
}

func testCreatePort(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	port := "2443"
	t.Logf("running idpbuilder create --port %s", port)
	cmd = exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create", "--port", port)
	cmd = enableKindExperimentalProvider(*cmd)
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
	cmd = exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, strings.Split(cmdString, " ")...)
	cmd = enableKindExperimentalProvider(*cmd)
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
