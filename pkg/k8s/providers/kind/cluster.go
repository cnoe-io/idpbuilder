package kind

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/cnoe-io/idpbuilder/pkg/k8s/provider"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
)

type Provider struct {
	provider cluster.Provider
}

func NewProvider() (*Provider, error) {
	detectOpt, err := cluster.DetectNodeProvider()
	if err != nil {
		return nil, err
	}
	return &Provider{
		provider: *cluster.NewProvider(detectOpt),
	}, nil
}

func (p *Provider) runsOnRightPort(ctx context.Context, clusterName string, config *provider.Config, cli client.APIClient) (bool, error) {
	allNodes, err := p.provider.ListNodes(clusterName)
	if err != nil {
		return false, err
	}

	cpNodes, err := nodeutils.ControlPlaneNodes(allNodes)
	if err != nil {
		return false, err
	}

	var cpNodeName string
	for _, cpNode := range cpNodes {
		if strings.Contains(cpNode.String(), clusterName) {
			cpNodeName = cpNode.String()
		}
	}
	if cpNodeName == "" {
		return false, nil
	}

	container, err := docker.GetContainerByName(ctx, cpNodeName, cli, types.ContainerListOptions{})
	if err != nil {
		return false, err
	}

	if container == nil {
		return false, nil
	}

	userPort, err := toUint16(config.Port)
	if err != nil {
		return false, err
	}

	return docker.IsUsingPort(container, userPort), nil
}

func (p *Provider) Provision(ctx context.Context, clusterName string, config *provider.Config) error {
	// check if user is requesting a different port
	// for the idpBuilder
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}

	rightPort, err := p.runsOnRightPort(ctx, clusterName, config, cli)
	if err != nil {
		return err
	}

	if !rightPort {
		return fmt.Errorf("cant serve port %s. cluster %s is already running on a different port", config.Port, clusterName)
	}

	rawConfig, err := p.getConfig(config)
	if err != nil {
		return err
	}

	fmt.Print("########################### Our kind config ############################\n")
	fmt.Printf("%s", rawConfig)
	fmt.Print("\n#########################   config end    ############################\n")

	fmt.Printf("Creating kind cluster %s\n", clusterName)
	if err = p.provider.Create(
		clusterName,
		cluster.CreateWithRawConfig(rawConfig),
	); err != nil {
		return err
	}

	fmt.Printf("Done creating cluster %s\n", clusterName)
	return nil
}

func (p *Provider) ListClusters() ([]string, error) {
	return p.provider.List()
}

func (p *Provider) ListNodes(clusterName string) ([]nodes.Node, error) {
	return p.provider.ListNodes(clusterName)
}

func (p *Provider) Delete(name string) error {
	return nil
}

func (p *Provider) GetAPIServerEndpoint(cluster string) (string, error) {
	return "", nil
}

func (p *Provider) GetAPIServerInternalEndpoint(cluster string) (string, error) {
	return "", nil
}

func (p *Provider) ExportKubeConfig(name string, path string, internal bool) error {
	return nil
}

type templateConfig struct {
	KubernetesVersion string
	ExtraPortsMapping []provider.PortMapping
	IngressProtocol   string
	Port              string
}

//go:embed resources/*
var configFS embed.FS

func (p *Provider) getConfig(config *provider.Config) ([]byte, error) {
	var rawConfigTempl []byte
	var err error

	if config.Kind.KindConfigPath != "" {
		rawConfigTempl, err = os.ReadFile(config.Kind.KindConfigPath)
	} else {
		rawConfigTempl, err = fs.ReadFile(configFS, "resources/kind.yaml.tmpl")
	}

	if err != nil {
		return []byte{}, err
	}

	var retBuff []byte
	if retBuff, err = util.ApplyTemplate(rawConfigTempl, templateConfig{
		KubernetesVersion: config.KubernetesVersion,
		ExtraPortsMapping: config.ExtraPortsMapping,
		IngressProtocol:   config.IngressProtocol,
		Port:              config.Port,
	}); err != nil {
		return []byte{}, err
	}

	return retBuff, nil
}

func toUint16(portString string) (uint16, error) {
	// Convert port string to uint16
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return 0, err
	}

	// Port validation
	if port > 65535 {
		return 0, errors.New("Invalid port number")
	}

	return uint16(port), nil
}
