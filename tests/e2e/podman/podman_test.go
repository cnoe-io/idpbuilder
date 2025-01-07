//go:build e2e

package podman

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/tests/e2e"
	"github.com/cnoe-io/idpbuilder/tests/e2e/container"
	"github.com/cnoe-io/idpbuilder/tests/e2e/shared"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type PodmanEngine struct {
	container.Engine
	Client string
}

func (p *PodmanEngine) GetClient() string {
	return p.Client
}

func (p *PodmanEngine) IdpCmd() *exec.Cmd {
	cmd := exec.Command(e2e.IdpbuilderBinaryLocation)
	cmd.Env = append(os.Environ(), "KIND_EXPERIMENTAL_PROVIDER=podman")
	return cmd
}

func (p *PodmanEngine) RunCommand(ctx context.Context, command string, timeout time.Duration) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmds := strings.Split(command, " ")
	if len(cmds) == 0 {
		return nil, fmt.Errorf("supply at least one command")
	}

	binary := cmds[0]
	args := make([]string, 0, len(cmds)-1)
	if len(cmds) > 1 {
		args = append(args, cmds[1:]...)
	}

	// Append some args to the podman command only
	if !strings.Contains(binary, "idpbuilder") {
		args = append(args, "--tls-verify=false")
	}

	c := exec.CommandContext(cmdCtx, binary, args...)

	// DOCKER_HOST = unix:///var/run/docker.sock is needed for podman running in rootless mode
	c.Env = append(os.Environ(), "DOCKER_HOST="+os.Getenv("DOCKER_HOST"))

	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while running %s: %s, %s", command, err, b)
	}

	return b, nil
}

func (p *PodmanEngine) RunIdpCommand(ctx context.Context, command string, timeout time.Duration) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmds := strings.Split(command, " ")
	if len(cmds) == 0 {
		return nil, fmt.Errorf("supply at least one command")
	}

	binary := cmds[0]
	args := make([]string, 0, len(cmds)-1)
	if len(cmds) > 1 {
		args = append(args, cmds[1:]...)
	}

	c := exec.CommandContext(cmdCtx, binary, args...)

	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while running %s: %s, %s", command, err, b)
	}

	return b, nil
}

func Test_CreateCluster(t *testing.T) {
	containerEngine := &PodmanEngine{Client: "podman"}
	shared.TestCreateCluster(t, containerEngine)
}
