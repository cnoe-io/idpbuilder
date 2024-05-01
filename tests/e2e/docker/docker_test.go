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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b, err := e2e.RunCommand(ctx, `docker ps -aqf name=localdev-control-plane`, 10*time.Second)
	assert.Nil(t, err, fmt.Sprintf("error while listing docker containers: %s, %s", err, b))

	conts := strings.TrimSpace(string(b))
	if len(conts) == 0 {
		return
	}
	b, err = e2e.RunCommand(ctx, fmt.Sprintf("docker container rm -f %s", conts), 60*time.Second)
	assert.Nil(t, err, fmt.Sprintf("error while removing docker containers: %s, %s", err, b))

	b, err = e2e.RunCommand(ctx, "docker system prune -f", 60*time.Second)
	assert.Nil(t, err, fmt.Sprintf("error while pruning system: %s, %s", err, b))

	b, err = e2e.RunCommand(ctx, "docker volume prune -f", 60*time.Second)
	assert.Nil(t, err, fmt.Sprintf("error while pruning volumes: %s, %s", err, b))
}

func Test_CreateDocker(t *testing.T) {
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctrl.SetLogger(logr.FromSlogHandler(slogger.Handler()))

	testCreate(t)
	testCreatePath(t)
	testCreatePort(t)
}

// test idpbuilder create
func testCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	t.Log("running idpbuilder create")
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create")
	b, err := cmd.CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.Nil(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)
}

// test idpbuilder create --use-path-routing
func testCreatePath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	t.Log("running idpbuilder create --use-path-routing")
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create", "--use-path-routing")
	b, err := cmd.CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.Nil(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://%s:%s/argocd", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://%s:%s/gitea", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)
}

func testCreatePort(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer CleanUpDocker(t)

	port := "2443"
	t.Logf("running idpbuilder create --port %s", port)
	cmd := exec.CommandContext(ctx, e2e.IdpbuilderBinaryLocation, "create", "--port", port)
	b, err := cmd.CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := e2e.GetKubeClient()
	assert.Nil(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, port)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, port)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)
}
