package runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
)

type DockerRuntime struct {
	client *dockerClient.Client
	name   string
}

func NewDockerRuntime(name string) (IRuntime, error) {
	client, err := dockerClient.NewClientWithOpts(
		dockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return &DockerRuntime{
		client: client,
		name:   name,
	}, nil
}

func (p *DockerRuntime) Name() string {
	return p.name
}

func (p *DockerRuntime) GetContainerByName(ctx context.Context, name string) (*types.Container, error) {
	gotContainers, err := p.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	if len(gotContainers) == 0 {
		return nil, fmt.Errorf("control plane container %s exists but is not in a running state", name)
	}

	var gotContainer *types.Container
	for _, container := range gotContainers {
		for _, containerName := range container.Names {
			// internal docker container name includes a fwd slash in the name
			if containerName == fmt.Sprintf("/%s", name) {
				gotContainer = &container
				break
			}
		}
		if gotContainer != nil {
			break
		}
	}

	return gotContainer, nil
}

func (p *DockerRuntime) IsUsingPort(container *types.Container, port uint16) bool {
	for _, p := range container.Ports {
		if p.PublicPort == port {
			return true
		}
	}
	return false
}

func (p *DockerRuntime) ContainerWithPort(ctx context.Context, name, port string) (bool, error) {
	container, err := p.GetContainerByName(ctx, name)
	if err != nil {
		return false, err
	}

	if container == nil {
		return false, nil
	}

	userPort, err := toUint16(port)
	if err != nil {
		return false, err
	}

	return p.IsUsingPort(container, userPort), nil
}
