package kind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
)

func TestRegistryMirrorHostsTomlContent(t *testing.T) {
	// Test that the generated hosts.toml for mirrors has the correct content
	cfg := v1alpha1.BuildCustomizationSpec{
		Host:                    "cnoe.localtest.me",
		Port:                    "8443",
		InsecureRegistryMirrors: true,
		RegistryMirrors: []v1alpha1.RegistryMirror{
			{
				TargetRegistry:  "docker.io",
				RegistryAddress: "http://kind-registry:5000",
			},
		},
	}

	dir, err := renderRegistryCertsDir(cfg)
	if err != nil {
		t.Fatalf("failed to render registry certs dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Check docker.io hosts.toml content
	dockerHostsFile := filepath.Join(dir, "docker.io", "hosts.toml")
	content, err := os.ReadFile(dockerHostsFile)
	if err != nil {
		t.Fatalf("failed to read hosts.toml: %v", err)
	}

	contentStr := string(content)

	// Verify key content exists
	expectedContent := []string{
		`server = "http://kind-registry:5000"`,
		`skip_verify = true`,
		`[host."http://kind-registry:5000"]`,
		`capabilities = ["pull", "resolve"]`,
	}

	for _, expected := range expectedContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("hosts.toml missing expected content: %s\nActual content:\n%s", expected, contentStr)
		}
	}

	// Verify content that should NOT exist
	unexpectedContent := []string{
		`proxy`,
		`https://docker.io`, // Should not reference target registry directly
	}

	for _, unexpected := range unexpectedContent {
		if strings.Contains(contentStr, unexpected) {
			t.Errorf("hosts.toml should not contain: %s\nActual content:\n%s", unexpected, contentStr)
		}
	}
}

func TestMultipleMirrors(t *testing.T) {
	// Test with multiple mirrors pointing to different registries
	cfg := v1alpha1.BuildCustomizationSpec{
		Host: "cnoe.localtest.me",
		Port: "8443",
		RegistryMirrors: []v1alpha1.RegistryMirror{
			{
				TargetRegistry:  "docker.io",
				RegistryAddress: "http://kind-registry:5000",
			},
			{
				TargetRegistry:  "ghcr.io",
				RegistryAddress: "http://kind-registry:5000",
			},
			{
				TargetRegistry:  "quay.io",
				RegistryAddress: "https://my-registry:5000",
			},
		},
	}

	dir, err := renderRegistryCertsDir(cfg)
	if err != nil {
		t.Fatalf("failed to render registry certs dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Verify all mirror directories exist
	registries := []string{"docker.io", "ghcr.io", "quay.io"}
	for _, registry := range registries {
		mirrorDir := filepath.Join(dir, registry)
		if _, err := os.Stat(mirrorDir); os.IsNotExist(err) {
			t.Errorf("expected mirror directory %s does not exist", mirrorDir)
		}

		hostsFile := filepath.Join(mirrorDir, "hosts.toml")
		if _, err := os.Stat(hostsFile); os.IsNotExist(err) {
			t.Errorf("expected hosts.toml file %s does not exist", hostsFile)
		}
	}

	// Verify docker.io uses http
	dockerFile := filepath.Join(dir, "docker.io", "hosts.toml")
	content, _ := os.ReadFile(dockerFile)
	if !strings.Contains(string(content), `server = "http://kind-registry:5000"`) {
		t.Errorf("docker.io should have http server URL")
	}

	// Verify quay.io uses https (since RegistryAddress starts with https)
	quayFile := filepath.Join(dir, "quay.io", "hosts.toml")
	content, _ = os.ReadFile(quayFile)
	if !strings.Contains(string(content), `server = "https://my-registry:5000"`) {
		t.Errorf("quay.io should have https server URL")
	}
}

func TestMirrorWithExistingGiteaConfig(t *testing.T) {
	// Test that mirrors work alongside the existing gitea registry config
	cfg := v1alpha1.BuildCustomizationSpec{
		Host: "cnoe.localtest.me",
		Port: "8443",
		RegistryMirrors: []v1alpha1.RegistryMirror{
			{
				TargetRegistry:  "docker.io",
				RegistryAddress: "http://kind-registry:5000",
			},
		},
	}

	dir, err := renderRegistryCertsDir(cfg)
	if err != nil {
		t.Fatalf("failed to render registry certs dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Verify gitea config exists
	giteaDir := filepath.Join(dir, "gitea.cnoe.localtest.me:8443")
	if _, err := os.Stat(giteaDir); os.IsNotExist(err) {
		t.Errorf("expected gitea directory %s does not exist", giteaDir)
	}

	giteaHostsFile := filepath.Join(giteaDir, "hosts.toml")
	if _, err := os.Stat(giteaHostsFile); os.IsNotExist(err) {
		t.Errorf("expected gitea hosts.toml file %s does not exist", giteaHostsFile)
	}

	// Verify docker.io mirror exists
	dockerDir := filepath.Join(dir, "docker.io")
	if _, err := os.Stat(dockerDir); os.IsNotExist(err) {
		t.Errorf("expected docker.io directory %s does not exist", dockerDir)
	}

	dockerHostsFile := filepath.Join(dockerDir, "hosts.toml")
	if _, err := os.Stat(dockerHostsFile); os.IsNotExist(err) {
		t.Errorf("expected docker.io hosts.toml file %s does not exist", dockerHostsFile)
	}
}

func TestMirrorWithHTTPS(t *testing.T) {
	// Test that mirrors work with HTTPS addresses
	cfg := v1alpha1.BuildCustomizationSpec{
		Host: "cnoe.localtest.me",
		Port: "8443",
		RegistryMirrors: []v1alpha1.RegistryMirror{
			{
				TargetRegistry:  "docker.io",
				RegistryAddress: "https://secure-registry:5000",
			},
		},
	}

	dir, err := renderRegistryCertsDir(cfg)
	if err != nil {
		t.Fatalf("failed to render registry certs dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Check docker.io hosts.toml content
	dockerHostsFile := filepath.Join(dir, "docker.io", "hosts.toml")
	content, err := os.ReadFile(dockerHostsFile)
	if err != nil {
		t.Fatalf("failed to read hosts.toml: %v", err)
	}

	contentStr := string(content)

	// Verify https is used
	if !strings.Contains(contentStr, `server = "https://secure-registry:5000"`) {
		t.Errorf("hosts.toml should contain https server URL\nActual content:\n%s", contentStr)
	}

	if !strings.Contains(contentStr, `[host."https://secure-registry:5000"]`) {
		t.Errorf("hosts.toml should contain https host configuration\nActual content:\n%s", contentStr)
	}
}

func TestMirrorWithHTTP(t *testing.T) {
	// Test that mirrors work with HTTP addresses
	cfg := v1alpha1.BuildCustomizationSpec{
		Host: "cnoe.localtest.me",
		Port: "8443",
		RegistryMirrors: []v1alpha1.RegistryMirror{
			{
				TargetRegistry:  "docker.io",
				RegistryAddress: "http://insecure-registry:5000",
			},
		},
	}

	dir, err := renderRegistryCertsDir(cfg)
	if err != nil {
		t.Fatalf("failed to render registry certs dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Check docker.io hosts.toml content
	dockerHostsFile := filepath.Join(dir, "docker.io", "hosts.toml")
	content, err := os.ReadFile(dockerHostsFile)
	if err != nil {
		t.Fatalf("failed to read hosts.toml: %v", err)
	}

	contentStr := string(content)

	// Verify http is used
	if !strings.Contains(contentStr, `server = "http://insecure-registry:5000"`) {
		t.Errorf("hosts.toml should contain http server URL\nActual content:\n%s", contentStr)
	}

	if !strings.Contains(contentStr, `[host."http://insecure-registry:5000"]`) {
		t.Errorf("hosts.toml should contain http host configuration\nActual content:\n%s", contentStr)
	}

	if strings.Contains(contentStr, `skip_verify = true`) {
		t.Errorf("hosts.toml should not contain skip_verify unless insecure-registry-mirrors is set\nActual content:\n%s", contentStr)
	}
}

func TestMirrorWithHTTPInsecure(t *testing.T) {
	cfg := v1alpha1.BuildCustomizationSpec{
		Host:                    "cnoe.localtest.me",
		Port:                    "8443",
		InsecureRegistryMirrors: true,
		RegistryMirrors: []v1alpha1.RegistryMirror{
			{
				TargetRegistry:  "docker.io",
				RegistryAddress: "http://insecure-registry:5000",
			},
		},
	}

	dir, err := renderRegistryCertsDir(cfg)
	if err != nil {
		t.Fatalf("failed to render registry certs dir: %v", err)
	}
	defer os.RemoveAll(dir)

	dockerHostsFile := filepath.Join(dir, "docker.io", "hosts.toml")
	content, err := os.ReadFile(dockerHostsFile)
	if err != nil {
		t.Fatalf("failed to read hosts.toml: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `skip_verify = true`) {
		t.Errorf("hosts.toml should contain skip_verify when insecure-registry-mirrors is set\nActual content:\n%s", contentStr)
	}
}
