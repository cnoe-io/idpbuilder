//go:build e2e

package podman

import (
	"github.com/cnoe-io/idpbuilder/tests/e2e"
	"github.com/cnoe-io/idpbuilder/tests/e2e/container"
	"github.com/cnoe-io/idpbuilder/tests/e2e/shared"
	"os"
	"os/exec"
	"testing"
)

type PodmanEngine struct {
	container.Engine
}

func (p *PodmanEngine) IdpCmd() *exec.Cmd {
	cmd := exec.Command(e2e.IdpbuilderBinaryLocation)
	cmd.Env = append(os.Environ(), "KIND_EXPERIMENTAL_PROVIDER=podman")
	return cmd
}

func Test_CreateCluster(t *testing.T) {
	p := &PodmanEngine{}
	shared.TestCreateCluster(t, container.ContainerClient(), p.IdpCmd())
}
