package kind

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/runtime"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
)

var (
	setupLog = log.Log.WithName("setup")
)

type Cluster struct {
	provider          IProvider
	runtime           runtime.IRuntime
	name              string
	kubeVersion       string
	kubeConfigPath    string
	kindConfigPath    string
	extraPortsMapping string
	cfg               util.CorePackageTemplateConfig
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

type TemplateConfig struct {
	KubernetesVersion string
	ExtraPortsMapping []PortMapping
	IngressProtocol   string
	Port              string
}

//go:embed resources/*
var configFS embed.FS

func (c *Cluster) getConfig() ([]byte, error) {

	var rawConfigTempl []byte
	var err error

	if c.kindConfigPath != "" {
		rawConfigTempl, err = os.ReadFile(c.kindConfigPath)
	} else {
		rawConfigTempl, err = fs.ReadFile(configFS, "resources/kind.yaml.tmpl")
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
	if retBuff, err = util.ApplyTemplate(rawConfigTempl, TemplateConfig{
		KubernetesVersion: c.kubeVersion,
		ExtraPortsMapping: portMappingPairs,
		IngressProtocol:   c.cfg.Protocol,
		Port:              c.cfg.Port,
	}); err != nil {
		return []byte{}, err
	}

	return retBuff, nil
}

func NewCluster(name, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping string, cfg util.CorePackageTemplateConfig) (*Cluster, error) {
	detectOpt, err := cluster.DetectNodeProvider()

	if err != nil {
		return nil, err
	}
	provider := cluster.NewProvider(detectOpt)

	rt, err := runtime.DetectRuntime()
	if err != nil {
		return nil, err
	}
	setupLog.Info("Runtime detected", "provider", rt.Name())

	return &Cluster{
		provider:          provider,
		runtime:           rt,
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

func (c *Cluster) RunsOnRightPort(ctx context.Context) (bool, error) {
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

	return c.runtime.ContainerWithPort(ctx, cpNodeName, c.cfg.Port)

}

func (c *Cluster) Reconcile(ctx context.Context, recreate bool) error {
	clusterExitsts, err := c.Exists()
	if err != nil {
		return err
	}

	if clusterExitsts {
		if recreate {
			setupLog.Info("Existing cluster found. Deleting.", "cluster", c.name)
			c.provider.Delete(c.name, "")
		} else {
			rightPort, err := c.RunsOnRightPort(ctx)
			if err != nil {
				return err
			}

			if !rightPort {
				return fmt.Errorf("can't serve port %s. cluster %s is already running on a different port", c.cfg.Port, c.name)
			}

			// reuse if there is no port conflict
			setupLog.Info("Cluster already exists", "cluster", c.name)
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

	setupLog.Info("Creating kind cluster", "cluster", c.name)
	if err = c.provider.Create(
		c.name,
		cluster.CreateWithRawConfig(rawConfig),
	); err != nil {
		return err
	}

	setupLog.Info("Done creating cluster", "cluster", c.name)

	return nil
}

func (c *Cluster) ExportKubeConfig(name string, internal bool) error {
	return c.provider.ExportKubeConfig(name, c.kubeConfigPath, internal)
}
