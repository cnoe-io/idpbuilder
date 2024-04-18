package provider

import (
	"context"

	"sigs.k8s.io/kind/pkg/cluster/nodes"
)

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

type KindConfig struct {
	KindConfigPath string
}

// Configuration for providers
type Config struct {
	KubernetesVersion string
	ExtraPortsMapping []PortMapping
	Port              string
	IngressProtocol   string
	Host              string

	Kind KindConfig
}

// This was mostly derived from sigs.k8s.io/kind/pkg/cluster/internal/providers
// We cannot import it directly as its an internal module

// Provider represents a provider of cluster / node infrastructure
type Provider interface {
	// Provision should create the Kubernetes cluster
	Provision(ctx context.Context, clusterName string, config *Config) error
	// ListClusters discovers the clusters that currently have resources
	// under this providers
	ListClusters() ([]string, error)
	// ListNodes returns the nodes under this provider for the given
	// cluster name, they may or may not be running correctly
	Delete(clusterName string) error
	ListNodes(clusterName string) ([]nodes.Node, error)
	// GetAPIServerEndpoint returns the host endpoint for the cluster's API server
	GetAPIServerEndpoint(clusterName string) (string, error)
	// GetAPIServerInternalEndpoint returns the internal network endpoint for the cluster's API server
	GetAPIServerInternalEndpoint(clusterName string) (string, error)
	// ExportKubeConfig writes out kube config under name to path
	ExportKubeConfig(clusterName string, path string, internal bool) error
}

type ProviderType string

const (
	KindProvider ProviderType = "Kind"
)
