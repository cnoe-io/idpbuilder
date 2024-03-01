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
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
)

type Cluster struct {
	provider          IProvider
	name              string
	kubeVersion       string
	kubeConfigPath    string
	kindConfigPath    string
	extraPortsMapping string
	cfg               util.TemplateConfig
}

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

type IProvider interface {
	List() ([]string, error)
	ListNodes(string) ([]nodes.Node, error)
	CollectLogs(string, string) error
	Delete(string, string) error
	Create(string, ...cluster.CreateOption) error
	ExportKubeConfig(string, string, bool) error
}

//go:embed resources/kind.yaml
var configFS embed.FS

func SplitFunc(input, sep string) []string {
	return strings.Split(input, sep)
}
func (c *Cluster) getConfig() ([]byte, error) {

	var rawConfigTempl []byte
	var err error

	if c.kindConfigPath != "" {
		rawConfigTempl, err = os.ReadFile(c.kindConfigPath)
	} else {
		rawConfigTempl, err = fs.ReadFile(configFS, "resources/kind.yaml")
	}

	if err != nil {
		return []byte{}, err
	}

	var portMappingPairs []PortMapping
	if len(c.extraPortsMapping) > 0 {
		// Split pairs of ports "11=1111","22=2222",etc
		pairs := strings.Split(c.extraPortsMapping, ",")
		// Create a slice to store PortMapping pairs.
		portMappingPairs = make([]PortMapping, len(pairs))
		// Parse each pair into PortPair objects.
		for i, pair := range pairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				portMappingPairs[i] = PortMapping{parts[0], parts[1]}
			}
		}
	} else {
		portMappingPairs = nil
	}

	var retBuff []byte
	if retBuff, err = util.ApplyTemplate(rawConfigTempl, struct {
		KubernetesVersion string
		ExtraPortsMapping []PortMapping
		Port              string
	}{
		KubernetesVersion: c.kubeVersion,
		ExtraPortsMapping: portMappingPairs,
		Port:              c.cfg.Port,
	}); err != nil {
		return []byte{}, err
	}

	return retBuff, nil
}

func NewCluster(name, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping string, cfg util.TemplateConfig) (*Cluster, error) {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())

	return &Cluster{
		provider:          provider,
		name:              name,
		kindConfigPath:    kindConfigPath,
		kubeVersion:       kubeVersion,
		kubeConfigPath:    kubeConfigPath,
		extraPortsMapping: extraPortsMapping,
		cfg:               cfg,
	}, nil
}

func (c *Cluster) Exists() (bool, error) {
	providerClusters, err := c.provider.List()
	if err != nil {
		return false, err
	}

	for _, pc := range providerClusters {
		if pc == c.name {
			return true, nil
		}
	}
	return false, nil
}

func (c *Cluster) RunsOnRightPort(cli client.APIClient, ctx context.Context) (bool, error) {

	allNodes, err := c.provider.ListNodes(c.name)
	if err != nil {
		return false, err
	}

	cpNodes, err := nodeutils.ControlPlaneNodes(allNodes)
	if err != nil {
		return false, err
	}

	var cpNodeName string
	for _, cpNode := range cpNodes {
		if strings.Contains(cpNode.String(), c.name) {
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

	userPort, err := toUint16(c.cfg.Port)
	if err != nil {
		return false, err
	}

	return docker.IsUsingPort(container, userPort), nil
}

func (c *Cluster) Reconcile(ctx context.Context, recreate bool) error {
	clusterExitsts, err := c.Exists()
	if err != nil {
		return err
	}

	if clusterExitsts {
		if recreate {
			fmt.Printf("Existing cluster %s found. Deleting.\n", c.name)
			c.provider.Delete(c.name, "")
		} else {
			// check if user is requesting a different port
			// for the idpBuilder
			cli, err := docker.GetDockerClient()
			if err != nil {
				return err
			}

			rightPort, err := c.RunsOnRightPort(cli, ctx)
			if err != nil {
				return err
			}

			if !rightPort {
				return fmt.Errorf("cant serve port %s. cluster %s is already running on a different port", c.cfg.Port, c.name)
			}

			// reuse if there is no port conflict
			fmt.Printf("Cluster %s already exists\n", c.name)
			return nil
		}
	}

	rawConfig, err := c.getConfig()
	if err != nil {
		return err
	}

	fmt.Print("########################### Our kind config ############################\n")
	fmt.Printf("%s", rawConfig)
	fmt.Print("\n#########################   config end    ############################\n")

	fmt.Printf("Creating kind cluster %s\n", c.name)
	if err = c.provider.Create(
		c.name,
		cluster.CreateWithRawConfig(rawConfig),
	); err != nil {
		return err
	}

	fmt.Printf("Done creating cluster %s\n", c.name)

	return nil
}

func (c *Cluster) ExportKubeConfig(name string, internal bool) error {
	return c.provider.ExportKubeConfig(name, c.kubeConfigPath, internal)
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
