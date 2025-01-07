//go:build e2e

package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/entity"
	"github.com/cnoe-io/idpbuilder/tests/e2e"
	"github.com/cnoe-io/idpbuilder/tests/e2e/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	IdpbuilderBinaryLocation = "../../../idpbuilder"
)

type ContainerEngine string

func cleanUp(t *testing.T) {
	t.Log("cleaning up")
}

// test idpbuilder create
func TestCreateCluster(t *testing.T, containerEngine container.Engine) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer cleanUp(t)

	t.Log("running idpbuilder create")
	cmd := containerEngine.IdpCmd()
	cmd.Args = append(cmd.Args, "create")
	b, err := cmd.CombinedOutput()
	assert.NoError(t, err, b)

	kubeClient, err := e2e.GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	e2e.TestArgoCDApps(ctx, t, kubeClient, e2e.CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", e2e.DefaultBaseDomain, e2e.DefaultPort)
	e2e.TestCoreEndpoints(ctx, t, argoBaseUrl, giteaBaseUrl)

	TestGiteaRegistry(ctx, t, containerEngine, fmt.Sprintf("gitea.%s", e2e.DefaultBaseDomain), e2e.DefaultPort)
}

// login, build a test image, push, then pull.
func TestGiteaRegistry(ctx context.Context, t *testing.T, containerEngine container.Engine, giteaHost, giteaPort string) {
	t.Log("testing gitea container registry")
	b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s get secrets -o json -p gitea", IdpbuilderBinaryLocation), 10*time.Second)
	assert.NoError(t, err)

	secs := make([]entity.Secret, 1)
	err = json.Unmarshal(b, &secs)
	assert.NoError(t, err)

	sec := secs[0]
	user := sec.Username
	pass := sec.Password

	login, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s login %s:%s -u %s -p %s", containerEngine.GetClient(), giteaHost, giteaPort, user, pass), 10*time.Second)
	require.NoErrorf(t, err, "%s login err: %s", containerEngine.GetClient(), login)

	tag := fmt.Sprintf("%s:%s/giteaadmin/test:latest", giteaHost, giteaPort)

	build, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s build -f test-dockerfile -t %s .", containerEngine.GetClient(), tag), 10*time.Second)
	require.NoErrorf(t, err, "%s build err: %s", containerEngine.GetClient(), build)

	push, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s push %s", containerEngine.GetClient(), tag), 10*time.Second)
	require.NoErrorf(t, err, "%s push err: %s", containerEngine.GetClient(), push)

	pull, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s pull %s", containerEngine.GetClient(), tag), 10*time.Second)
	require.NoErrorf(t, err, "%s pull err: %s", containerEngine.GetClient(), pull)
}
