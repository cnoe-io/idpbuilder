package kind

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"text/template"

	"sigs.k8s.io/kind/pkg/cluster"
)

type Cluster struct {
	provider       *cluster.Provider
	name           string
	kubeConfigPath string
}

//go:embed resources/kind.yaml
var configFS embed.FS

func (c *Cluster) getConfig() ([]byte, error) {
	rawConfigTempl, err := fs.ReadFile(configFS, "resources/kind.yaml")
	if err != nil {
		return []byte{}, err
	}

	template, err := template.New("kind.yaml").Parse(string(rawConfigTempl))
	if err != nil {
		return []byte{}, err
	}

	retBuff := bytes.Buffer{}
	if err = template.Execute(&retBuff, struct {
		RegistryHostname     string
		ExposedRegistryPort  uint16
		InternalRegistryPort uint16
	}{
		RegistryHostname:     c.getRegistryContainerName(),
		ExposedRegistryPort:  ExposedRegistryPort,
		InternalRegistryPort: InternalRegistryPort,
	}); err != nil {
		return []byte{}, err
	}
	return retBuff.Bytes(), nil
}

func NewCluster(name string, kubeConfigPath string) (*Cluster, error) {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())

	return &Cluster{
		provider:       provider,
		name:           name,
		kubeConfigPath: kubeConfigPath,
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
			fmt.Printf("Existing cluster %s found. Deleting.\n", c.name)
			c.provider.Delete(c.name, "")
		} else {
			fmt.Printf("Cluster %s already exists\n", c.name)
			return c.ReconcileRegistry(ctx)
		}
	}

	rawConfig, err := c.getConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Creating kind cluster %s\n", c.name)
	if err = c.provider.Create(
		c.name,
		cluster.CreateWithRawConfig(rawConfig),
	); err != nil {
		return err
	}

	fmt.Printf("Done creating cluster %s\n", c.name)

	return c.ReconcileRegistry(ctx)
}

func (c *Cluster) ExportKubeConfig(name string, internal bool) error {
	return c.provider.ExportKubeConfig(name, c.kubeConfigPath, internal)
}
