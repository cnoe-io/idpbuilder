package provider

import "sigs.k8s.io/kind/pkg/cluster/nodes"

// Configuration for providers
type Config struct {
	Name string
}

// This was mostly copied from sigs.k8s.io/kind/pkg/cluster/internal/providers
// We cannot import it directly as its an internal module

// Provider represents a provider of cluster / node infrastructure
type Provider interface {
	// Provision should create and start the nodes, just short of
	// actually starting up Kubernetes, based on the given cluster config
	Provision(config Config) error
	// ListClusters discovers the clusters that currently have resources
	// under this providers
	ListClusters() ([]string, error)
	// ListNodes returns the nodes under this provider for the given
	// cluster name, they may or may not be running correctly
	Delete(name string) error
	ListNodes(cluster string) ([]nodes.Node, error)
	// GetAPIServerEndpoint returns the host endpoint for the cluster's API server
	GetAPIServerEndpoint(cluster string) (string, error)
	// GetAPIServerInternalEndpoint returns the internal network endpoint for the cluster's API server
	GetAPIServerInternalEndpoint(cluster string) (string, error)
}

type ProviderType string

const (
	KindProvider ProviderType = "Kind"
)
