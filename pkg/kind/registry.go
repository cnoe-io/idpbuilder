package kind

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const registryImage string = "registry:2"
const ExposedRegistryPort uint16 = 5001
const InternalRegistryPort uint16 = 5000

func (c *Cluster) getRegistryContainerName() string {
	return fmt.Sprintf("%s-%s-registry", globals.ProjectName, c.name)
}

func (c *Cluster) getRegistryContainer(ctx context.Context, dockerClient *client.Client) (*types.Container, error) {
	return docker.GetOneContainer(ctx, dockerClient, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: c.getRegistryContainerName(),
		}),
	})
}

func (c *Cluster) ReconcileRegistry(ctx context.Context) error {
	log := log.FromContext(ctx)

	log.Info("Reconciling registry container")

	dockerCli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	defer dockerCli.Close()

	log.Info("Pulling registry image")
	reader, err := dockerCli.ImagePull(ctx, registryImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
	log.Info("Done pulling registry image")

	regContainer, err := c.getRegistryContainer(ctx, dockerCli)
	if err != nil {
		return err
	}
	if regContainer == nil {
		log.Info("Creating registry container")
		resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
			Image: registryImage,
			Tty:   false,
			ExposedPorts: nat.PortSet{
				nat.Port(fmt.Sprintf("%d/tcp", InternalRegistryPort)): struct{}{},
			},
		}, &container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%d/tcp", InternalRegistryPort)): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: fmt.Sprintf("%d", ExposedRegistryPort),
					},
				},
			},
		}, nil, nil, c.getRegistryContainerName())
		if err != nil {
			log.Error(err, "Error creating registry container")
			return err
		}

		if err := dockerCli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
		regContainer, err = c.getRegistryContainer(ctx, dockerCli)
		if err != nil {
			return err
		}
	} else {
		log.Info("Registry container found, skipping create")
	}

	onKindNetwork := false
	for network := range regContainer.NetworkSettings.Networks {
		if network == "kind" {
			onKindNetwork = true
		}
	}

	if !onKindNetwork {
		log.Info("Putting the registry container on the kind network")
		err = dockerCli.NetworkConnect(ctx, "kind", regContainer.ID, &network.EndpointSettings{})
		if err != nil {
			return err
		}
	} else {
		log.Info("Registry container on kind network, skipping connecting network")
	}

	log.Info("Done reconciling registry container")
	return nil
}
