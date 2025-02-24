package kind

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
)

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

type TemplateConfig struct {
	v1alpha1.BuildCustomizationSpec
	KubernetesVersion string
	ExtraPortsMapping []PortMapping
	RegistryConfig    string
}

//go:embed resources/* testdata/custom-kind.yaml.tmpl
var configFS embed.FS

func loadConfig(path string, httpClient HttpClient) ([]byte, error) {
	var rawConfigTempl []byte
	var err error
	if path != "" {
		if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
			resp, err := httpClient.Get(path)
			if err != nil {
				return nil, fmt.Errorf("fetching remote kind config: %w", err)
			}
			defer resp.Body.Close()
			if !(resp.StatusCode < 300 && resp.StatusCode >= 200) {
				return nil, fmt.Errorf("got %d status code when fetching kind config", resp.StatusCode)
			}
			rawConfigTempl, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("reading remote kind config body: %w", err)
			}
		} else {
			rawConfigTempl, err = os.ReadFile(path)
		}
	} else {
		rawConfigTempl, err = fs.ReadFile(configFS, "resources/kind.yaml.tmpl")
	}

	if err != nil {
		return nil, fmt.Errorf("reading kind config: %w", err)
	}
	return rawConfigTempl, nil
}

func parsePortMappings(extraPortsMapping string) []PortMapping {
	var portMappingPairs []PortMapping
	if len(extraPortsMapping) > 0 {
		// Split pairs of ports "11=1111","22=2222",etc
		pairs := strings.Split(extraPortsMapping, ",")
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
	return portMappingPairs
}

func findRegistryConfig(registryConfigPaths []string) string {
	for _, s := range registryConfigPaths {
		path := os.ExpandEnv(s)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
