package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetOneContainer(ctx context.Context, dockerClient *client.Client, listOptions types.ContainerListOptions) (*types.Container, error) {
	gotContainers, err := dockerClient.ContainerList(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	if len(gotContainers) == 0 {
		return nil, nil
	} else if len(gotContainers) > 1 {
		return nil, fmt.Errorf("expected 1 container, got %d", len(gotContainers))
	}
	return &gotContainers[0], nil
}

func Exec(ctx context.Context, dockerClient *client.Client, container string, config types.ExecConfig) error {
	log := log.FromContext(ctx)
	resp, err := dockerClient.ContainerExecCreate(ctx, container, config)
	if err != nil {
		return err
	}

	err = dockerClient.ContainerExecStart(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	for {
		status, err := dockerClient.ContainerExecInspect(ctx, resp.ID)
		if err != nil {
			return err
		}
		if status.Running {
			log.Info("Waiting for hostname remapping to install")
			time.Sleep(time.Millisecond * 500)
			continue
		}

		if status.ExitCode != 0 {
			return fmt.Errorf("failed to install registry hostname remapping, exit code: %d", status.ExitCode)
		}
		break
	}
	return nil
}
