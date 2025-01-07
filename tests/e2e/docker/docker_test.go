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

func (p *DockerEngine) RunCommand(ctx context.Context, command string, timeout time.Duration) ([]byte, error) {
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

func (p *DockerEngine) RunIdpCommand(ctx context.Context, command string, timeout time.Duration) ([]byte, error) {
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

func Test_CreateDocker(t *testing.T) {
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctrl.SetLogger(logr.FromSlogHandler(slogger.Handler()))

	containerEngine := &DockerEngine{Client: "docker"}
	shared.TestCreateCluster(t, containerEngine)
	shared.TestCreatePath(t, containerEngine)
	shared.TestCreatePort(t, containerEngine)
	shared.TestCustomPkg(t, containerEngine)
}
