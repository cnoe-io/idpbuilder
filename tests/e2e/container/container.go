package container

import (
	"context"
	"os"
	"os/exec"
	"time"
)

const (
	IdpbuilderBinaryLocation = "../../../idpbuilder"
)

type Engine interface {
	IdpCmd() *exec.Cmd
	RunIdpCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	RunCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	GetClient() string
}

func ContainerClient() string {
	if os.Getenv("CONTAINER_ENGINE") == "podman" {
		return "podman"
	} else {
		return "docker"
	}
}
