//go:build e2e

package docker

import (
	"github.com/cnoe-io/idpbuilder/tests/e2e/container"
	"github.com/cnoe-io/idpbuilder/tests/e2e/shared"
	"github.com/go-logr/logr"
	"log/slog"
	"os"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
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

func Test_CreateDocker(t *testing.T) {
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctrl.SetLogger(logr.FromSlogHandler(slogger.Handler()))

	containerEngine := &DockerEngine{Client: "docker"}
	shared.TestCreateCluster(t, containerEngine)
	shared.TestCreatePath(t, containerEngine)
	shared.TestCreatePort(t, containerEngine)
	shared.TestCustomPkg(t, containerEngine)
}
