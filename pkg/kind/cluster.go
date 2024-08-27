package kind

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/runtime"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/log"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
	"sigs.k8s.io/yaml"
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
	util.CorePackageTemplateConfig
	KubernetesVersion string
	ExtraPortsMapping []PortMapping
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
		return nil, fmt.Errorf("reading kind config: %w", err)
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
	}

	var retBuff []byte
	if retBuff, err = util.ApplyTemplate(rawConfigTempl, TemplateConfig{
		CorePackageTemplateConfig: c.cfg,
		KubernetesVersion:         c.kubeVersion,
		ExtraPortsMapping:         portMappingPairs,
	}); err != nil {
		return nil, err
	}

	if c.kindConfigPath != "" {
		parsedCluster, err := c.ensureCorrectConfig(retBuff)
		if err != nil {
			return nil, fmt.Errorf("ensuring custom kind config is correct: %w", err)
		}

		out, err := yaml.Marshal(parsedCluster)
		if err != nil {
			return nil, fmt.Errorf("marshaling custom kind cluster config: %w", err)
		}
		return out, nil
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

func (c *Cluster) ensureCorrectConfig(in []byte) (kindv1alpha4.Cluster, error) {
	// see pkg/kind/resources/kind.yaml.tmpl and pkg/controllers/localbuild/resources/nginx/k8s/ingress-nginx.yaml
	// defines which container port we should be looking for.
	containerPort := "443"
	if c.cfg.Protocol == "http" {
		containerPort = "80"
	}
	parsedCluster := kindv1alpha4.Cluster{}
	err := yaml.Unmarshal(in, &parsedCluster)
	if err != nil {
		return kindv1alpha4.Cluster{}, fmt.Errorf("parsing kind config: %w", err)
	}

	appendNecessaryPort := true
nodes:
	for i := range parsedCluster.Nodes {
		node := parsedCluster.Nodes[i]
		for _, pm := range node.ExtraPortMappings {
			if strconv.Itoa(int(pm.HostPort)) == c.cfg.Port {
				appendNecessaryPort = false
				break nodes
			}
		}
	}

	if appendNecessaryPort && len(parsedCluster.Nodes) != 0 {
		hp, err := strconv.Atoi(c.cfg.Port)
		if err != nil {
			return kindv1alpha4.Cluster{}, fmt.Errorf("converting port, %s, to int: %w", c.cfg.Port, err)
		}
		// either "80" or "443". No need to check for err
		cp, _ := strconv.Atoi(containerPort)

		parsedCluster.Nodes[0].ExtraPortMappings = append(parsedCluster.Nodes[0].ExtraPortMappings, kindv1alpha4.PortMapping{ContainerPort: int32(cp), HostPort: int32(hp)})
	}

	return parsedCluster, nil
}
