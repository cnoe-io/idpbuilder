//go:build e2e

package docker

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

type DockerEngine struct {
	Client string
}

func (p *DockerEngine) GetClient() string {
	return p.Client
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

	var c *exec.Cmd
	if timeout > 0 {
		c = exec.CommandContext(cmdCtx, binary, args...)
	} else {
		c = exec.Command(binary, args...)
	}

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
	e2e.TestCreateCluster(t, containerEngine)
	e2e.TestCreatePath(t, containerEngine)
	e2e.TestCreatePort(t, containerEngine)
	e2e.TestCustomPkg(t, containerEngine)
}
