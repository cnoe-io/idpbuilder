package kind

import (
	"github.com/cnoe-io/idpbuilder/pkg/k8s/provider"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
)

type Cluster struct {
}

func (c *Cluster) Provision(config provider.Config) error {
	return nil
}

func (c *Cluster) ListClusters() ([]string, error) {
	return []string{}, nil
}

func (c *Cluster) ListNodes(cluster string) ([]nodes.Node, error) {
	return []nodes.Node{}, nil
}

func (c *Cluster) Delete(name string) error {
	return nil
}

func (c *Cluster) GetAPIServerEndpoint(cluster string) (string, error) {
	return "", nil
}

func (c *Cluster) GetAPIServerInternalEndpoint(cluster string) (string, error) {
	return "", nil
}
