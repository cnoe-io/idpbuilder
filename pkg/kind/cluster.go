package kind

import (
	"context"
	"errors"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/cnoe-io/idpbuilder/pkg/util/files"
	"net/http"
	"strconv"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	kindexec "sigs.k8s.io/kind/pkg/exec"
	"sigs.k8s.io/yaml"
)

const (
	ingressNginxNodeLabelKey   = "ingress-ready"
	ingressNginxNodeLabelValue = "true"
)

var (
	setupLog = log.Log.WithName("setup")
)

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Cluster struct {
	provider          IProvider
	httpClient        HttpClient
	name              string
	kubeVersion       string
	kubeConfigPath    string
	kindConfigPath    string
	extraPortsMapping string
	registryConfig    []string
	cfg               v1alpha1.BuildCustomizationSpec
}

type IProvider interface {
	List() ([]string, error)
	ListNodes(string) ([]nodes.Node, error)
	CollectLogs(string, string) error
	Delete(string, string) error
	Create(string, ...cluster.CreateOption) error
	ExportKubeConfig(string, string, bool) error
}

func (c *Cluster) getConfig() ([]byte, error) {
	rawConfigTempl, err := loadConfig(c.kindConfigPath, c.httpClient)

	portMappingPairs := parsePortMappings(c.extraPortsMapping)

	registryConfig := findRegistryConfig(c.registryConfig)

	if len(c.registryConfig) > 0 && registryConfig == "" {
		return nil, errors.New("--registry-config flag used but no registry config was found")
	}

	var retBuff []byte
	if retBuff, err = files.ApplyTemplate(rawConfigTempl, TemplateConfig{
		BuildCustomizationSpec: c.cfg,
		KubernetesVersion:      c.kubeVersion,
		ExtraPortsMapping:      portMappingPairs,
		RegistryConfig:         registryConfig,
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

func NewCluster(name, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping string, registryConfig []string, cfg v1alpha1.BuildCustomizationSpec, cliLogger logr.Logger) (*Cluster, error) {
	detectOpt, err := util.DetectKindNodeProvider()
	if err != nil {
		return nil, err
	}

	provider := cluster.NewProvider(cluster.ProviderWithLogger(KindLoggerFromLogr(&cliLogger)), detectOpt)

	return &Cluster{
		provider:          provider,
		httpClient:        util.GetHttpClient(),
		name:              name,
		kindConfigPath:    kindConfigPath,
		kubeVersion:       kubeVersion,
		kubeConfigPath:    kubeConfigPath,
		extraPortsMapping: extraPortsMapping,
		registryConfig:    registryConfig,
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

func (c *Cluster) Reconcile(ctx context.Context, recreate bool) error {
	clusterExitsts, err := c.Exists()
	if err != nil {
		return err
	}

	if clusterExitsts {
		if recreate {
			setupLog.Info("Existing cluster found. Deleting.", "cluster", c.name)
			err := c.provider.Delete(c.name, "")
			if err != nil {
				return fmt.Errorf("deleting cluster %w", err)
			}
		} else {
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
		t := &kindexec.RunError{}
		if errors.As(err, &t) {
			return fmt.Errorf("%w: %s", err, t.Output)
		}
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
	// the port and ingress-nginx label must be on the same node to ensure nginx runs on the node with the right port.
	appendNecessaryPort := true
	appendIngressNodeLabel := true
	// pick the first node for the ingress-nginx if we need to configure node port.
	nodePosition := 0

	if parsedCluster.Nodes == nil || len(parsedCluster.Nodes) == 0 {
		return kindv1alpha4.Cluster{}, fmt.Errorf("provided kind config does not have the node field defined")
	}

nodes:
	for i := range parsedCluster.Nodes {
		node := parsedCluster.Nodes[i]
		for _, pm := range node.ExtraPortMappings {
			if strconv.Itoa(int(pm.HostPort)) == c.cfg.Port {
				appendNecessaryPort = false
				nodePosition = i
				if node.Labels != nil {
					v, ok := node.Labels[ingressNginxNodeLabelKey]
					if ok && v == ingressNginxNodeLabelValue {
						appendIngressNodeLabel = false
					}
				}
				break nodes
			}
		}
		if node.Labels != nil {
			v, ok := node.Labels[ingressNginxNodeLabelKey]
			if ok && v == ingressNginxNodeLabelValue {
				appendIngressNodeLabel = false
				nodePosition = i
				break nodes
			}
		}
	}

	if appendNecessaryPort {
		hp, err := strconv.Atoi(c.cfg.Port)
		if err != nil {
			return kindv1alpha4.Cluster{}, fmt.Errorf("converting port, %s, to int: %w", c.cfg.Port, err)
		}
		// either "80" or "443". No need to check for err
		cp, _ := strconv.Atoi(containerPort)

		if parsedCluster.Nodes[nodePosition].ExtraPortMappings == nil {
			parsedCluster.Nodes[nodePosition].ExtraPortMappings = make([]kindv1alpha4.PortMapping, 0, 1)
		}
		parsedCluster.Nodes[nodePosition].ExtraPortMappings =
			append(parsedCluster.Nodes[nodePosition].ExtraPortMappings, kindv1alpha4.PortMapping{ContainerPort: int32(cp), HostPort: int32(hp), Protocol: "TCP"})
	}
	if appendIngressNodeLabel {
		if parsedCluster.Nodes[nodePosition].Labels == nil {
			parsedCluster.Nodes[nodePosition].Labels = make(map[string]string)
		}
		parsedCluster.Nodes[nodePosition].Labels[ingressNginxNodeLabelKey] = ingressNginxNodeLabelValue
	}

	return parsedCluster, nil
}
