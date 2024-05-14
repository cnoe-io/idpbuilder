package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Container struct {
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

func (p *FinchRuntime) Name() string {
	return "finch"
}

func (f *FinchRuntime) ContainerWithPort(ctx context.Context, name string, port string) (bool, error) {
	logger := log.FromContext(ctx)
	// add arguments to inspect the container
	f.cmd.Args = append([]string{"finch"}, "container", "inspect", name)

	var stdout, stderr bytes.Buffer
	f.cmd.Stdout = &stdout
	f.cmd.Stderr = &stderr

	// Execute the command
	logger.V(1).Info("inspect existing cluster for the configuration", "container", name, "port", port)
	err := f.cmd.Run()
	if err != nil {
		return false, err
	}

	var containers []Container
	err = json.Unmarshal(stdout.Bytes(), &containers)
	if err != nil {
		return false, fmt.Errorf("%v: %s", err, stderr.String())
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
					logger.V(1).Info("existing cluster matches the configuration", "container", name, "port", port)
					return true, nil
				}
			}
		}
	}

	logger.V(1).Info("existing cluster does not match the configuration", "container", name, "port", port)
	return false, nil
}
