package kind

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/template"

	"sigs.k8s.io/kind/pkg/cluster"
)

type Cluster struct {
	provider          *cluster.Provider
	name              string
	kubeVersion       string
	kubeConfigPath    string
	kindConfigPath    string
	extraPortsMapping string
}

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

//go:embed resources/kind.yaml
var configFS embed.FS

func SplitFunc(input, sep string) []string {
	return strings.Split(input, sep)
}
func (c *Cluster) getConfig() ([]byte, error) {

	if c.kindConfigPath != "" {
		f, err := os.ReadFile(c.kindConfigPath)
		if err != nil {
			return []byte{}, err
		} else {
			return f, nil
		}
	}

	rawConfigTempl, err := fs.ReadFile(configFS, "resources/kind.yaml")
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

	template, err := template.New("kind.yaml").Parse(string(rawConfigTempl))
	if err != nil {
		return []byte{}, err
	}

	retBuff := bytes.Buffer{}
	if err = template.Execute(&retBuff, struct {
		KubernetesVersion string
		ExtraPortsMapping []PortMapping
	}{
		KubernetesVersion: c.kubeVersion,
		ExtraPortsMapping: portMappingPairs,
	}); err != nil {
		return []byte{}, err
	}
	return retBuff.Bytes(), nil
}

func NewCluster(name, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping string) (*Cluster, error) {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())

	return &Cluster{
		provider:          provider,
		name:              name,
		kindConfigPath:    kindConfigPath,
		kubeVersion:       kubeVersion,
		kubeConfigPath:    kubeConfigPath,
		extraPortsMapping: extraPortsMapping,
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
			c.provider.Delete(c.name, "")
		} else {
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
