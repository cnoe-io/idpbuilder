//go:build e2e

package shared

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/tests/e2e"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
	"time"
)

type ContainerEngine string

func cleanUp(t *testing.T) {
	t.Log("cleaning up")
}

// test idpbuilder create
func TestCreateCluster(t *testing.T, engine string, cmd *exec.Cmd) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer cleanUp(t)

	t.Log("running idpbuilder create")
	cmd.Args = append(cmd.Args, "create")
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, b)

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)

	e2e.TestGiteaRegistry(ctx, t, engine, fmt.Sprintf("gitea.%s", e2e.DefaultBaseDomain), e2e.DefaultPort)
}
