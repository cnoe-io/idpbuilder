package k8s

import (
	"github.com/cnoe-io/idpbuilder/pkg/k8s/provider"
	"github.com/cnoe-io/idpbuilder/pkg/k8s/providers/kind"
)

func CreateProvider(providerType provider.ProviderType, config provider.Config) (provider.Provider, error) {
	var prvdr provider.Provider
	switch providerType {
	case provider.KindProvider:
		prvdr = &kind.Cluster{}
	}

	if err := prvdr.Provision(config); err != nil {
		return nil, err
	}

	return prvdr, nil
}
