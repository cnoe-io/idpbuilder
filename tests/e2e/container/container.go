package container

import (
	"os"
	"os/exec"
)

type Engine interface {
	IdpCmd() *exec.Cmd
}

func ContainerClient() string {
	if os.Getenv("CONTAINER_ENGINE") == "podman" {
		return "podman"
	} else {
		return "docker"
	}
}
