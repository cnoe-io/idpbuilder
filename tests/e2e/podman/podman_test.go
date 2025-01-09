//go:build e2e

package podman

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/tests/e2e"
	"github.com/go-logr/logr"
	"log/slog"
	"os"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"testing"
	"time"
)

type PodmanEngine struct {
	Client string
}

// Implementation of the method Getclient of the interface: container.Engine
func (p *PodmanEngine) GetClient() string {
	return p.Client
}

// Implementation of the method RunCommand of the interface: container.Engine
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

	if cmds[1] == "login" || cmds[1] == "push" || cmds[1] == "pull" {
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

// Implementation of the method RunIdpCommand of the interface: container.Engine
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

	var c *exec.Cmd
	if timeout > 0 {
		c = exec.CommandContext(cmdCtx, binary, args...)
	} else {
		c = exec.Command(binary, args...)
	}
	c.Env = append(os.Environ(), "KIND_EXPERIMENTAL_PROVIDER=podman")

	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while running %s: %s, %s", command, err, b)
	}

	return b, nil
}

func Test_CreateCluster(t *testing.T) {
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctrl.SetLogger(logr.FromSlogHandler(slogger.Handler()))

	containerEngine := &PodmanEngine{Client: "podman"}
	e2e.TestCreateCluster(t, containerEngine)
	e2e.TestCreatePath(t, containerEngine)
	e2e.TestCreatePort(t, containerEngine)
	e2e.TestCustomPkg(t, containerEngine)
}
