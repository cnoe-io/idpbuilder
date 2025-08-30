package kind

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util/files"
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
	RegistryCertsDir  string
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

func renderRegistryCertsDir(cfg v1alpha1.BuildCustomizationSpec) (string, error) {
	// Render out the template
	rawConfigTempl, err := fs.ReadFile(configFS, "resources/hosts.toml.tmpl")
	if err != nil {
		return "", fmt.Errorf("reading insecure registry config %w", err)
	}

	var retBuff []byte
	if retBuff, err = files.ApplyTemplate(rawConfigTempl, cfg); err != nil {
		return "", fmt.Errorf("templating insecure registry config %w", err)
	}

	// Generate the directory structure and write the file to hosts.toml
	dir, err := os.MkdirTemp("", "idpbuilder-registry-certs.d-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir %w", err)
	}

	var hostAndPort string
	if cfg.UsePathRouting {
		hostAndPort = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	} else {
		hostAndPort = fmt.Sprintf("gitea.%s:%s", cfg.Host, cfg.Port)
	}
	hostCertsDir := filepath.Join(dir, hostAndPort)
	err = os.Mkdir(hostCertsDir, 0700)
	if err != nil {
		return "", fmt.Errorf("creating temp dir for host %w", err)
	}
	hostsFile := filepath.Join(hostCertsDir, "hosts.toml")

	err = os.WriteFile(hostsFile, retBuff, 0700)
	if err != nil {
		return "", fmt.Errorf("writing insecure registry config %w", err)
	}

	return dir, nil
}
