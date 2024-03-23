package k8s

import (
	"fmt"

	"github.com/cnoe-io/idpbuilder/pkg/k8s/provider"
	"github.com/cnoe-io/idpbuilder/pkg/k8s/providers/kind"
)

func GetProvider(providerType provider.ProviderType, config *provider.Config) (provider.Provider, error) {
	switch providerType {
	case provider.KindProvider:
		return kind.NewProvider()
	}
	return nil, fmt.Errorf("invalid provider type")
}
