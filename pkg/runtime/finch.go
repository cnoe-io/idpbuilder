package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type Conainer struct {
	NetworkSettings NetworkSettings `json:"NetworkSettings"`
	State           State           `json:"State"`
}

type NetworkSettings struct {
	Ports map[string][]PortBinding `json:"Ports"`
}

type State struct {
	Status string `json:"Status"`
}

type PortBinding struct {
	HostIp   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type FinchRuntime struct {
	cmd *exec.Cmd
}

func NewFinchRuntime() (IRuntime, error) {
	return &FinchRuntime{
		cmd: exec.Command("finch"),
	}, nil
}

func (f *FinchRuntime) ContainerWithPort(ctx context.Context, name string, port string) (bool, error) {
	println("checking the container for port", name, port)
	// add arguments to inspect the container
	f.cmd.Args = append([]string{"finch"}, "container", "inspect", name)

	// Execute the command
	output, err := f.cmd.Output()
	if err != nil {
		return false, err
	}

	var containers []Conainer
	err = json.Unmarshal(output, &containers)
	if err != nil {
		return false, err
	}

	if len(containers) > 1 {
		return false, errors.New("more than one container for the kind control plane")
	}

	if containers[0].State.Status != Running {
		return false, fmt.Errorf("control plane container %s exists but is not in a running state", name)
	}

	for _, container := range containers {
		for _, bindings := range container.NetworkSettings.Ports {
			for _, binding := range bindings {
				if binding.HostPort == port {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
