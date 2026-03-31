package kind

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
)

type MockHttpClient struct{}

func (o *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	if url == "https://doesnotexist" || url == "http://doesnotexist" {
		return nil, errors.New("connection error")
	} else if url == "https://404" {
		body := io.NopCloser(strings.NewReader(""))
		r := http.Response{
			Status:     "404 NotFound",
			StatusCode: 404,
			Body:       body,
		}
		return &r, nil
	}

	body := io.NopCloser(strings.NewReader("foo: bar"))
	r := http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       body,
	}

	return &r, nil
}

func TestLoadConfig(t *testing.T) {
	httpClient := MockHttpClient{}
	defaultTemplate, err := fs.ReadFile(configFS, "resources/kind.yaml.tmpl")
	if err != nil {
		t.Fatalf("failed to load default kind template: %v", err)
	}

	customTemplate, err := fs.ReadFile(configFS, "testdata/custom-kind.yaml.tmpl")
	if err != nil {
		t.Fatalf("failed to load custom kind template: %v", err)
	}

	httpsTemplate := []byte("foo: bar")

	connectionErr := "fetching remote kind config: connection error"
	notFoundErr := "got 404 status code when fetching kind config"

	type test struct {
		path     string
		expected []byte
		err      *string
	}
	tests := []test{
		{
			path:     "",
			expected: defaultTemplate,
			err:      nil,
		},
		{
			path:     "testdata/custom-kind.yaml.tmpl",
			expected: customTemplate,
			err:      nil,
		},
		{
			path:     "https://doesnotexist",
			expected: defaultTemplate,
			err:      &connectionErr,
		},
		{
			path:     "http://doesnotexist",
			expected: customTemplate,
			err:      &connectionErr,
		},
		{
			path:     "https://404",
			expected: defaultTemplate,
			err:      &notFoundErr,
		},
		{
			path:     "https://anyurlworks",
			expected: httpsTemplate,
			err:      nil,
		},
	}

	for _, tc := range tests {
		out, err := loadConfig(tc.path, &httpClient)
		if tc.err != nil {
			if err != nil {
				if err.Error() != *tc.err {
					t.Errorf("expected error: %v\nfound error: %v", *tc.err, err.Error())
				}
			} else {
				t.Errorf("expected error: %v\ndidnt find an error", *tc.err)
			}
		} else {
			if err != nil {
				t.Errorf("failed to load kind config: %v", err)
			}
			if !reflect.DeepEqual(tc.expected, out) {
				t.Errorf("expected:\n%v\ngot:\n%v", string(tc.expected), string(out))
			}
		}
	}
}

func TestExtraPortMappingsUtilFunc(t *testing.T) {
	type test struct {
		extraPortMappings string
		expected          []PortMapping
	}
	tests := []test{
		{
			extraPortMappings: "",
			expected:          []PortMapping(nil),
		},
		{
			extraPortMappings: "22:32222",
			expected: []PortMapping{
				{
					HostPort:      "22",
					ContainerPort: "32222",
				},
			},
		},
		{
			extraPortMappings: "11:1111,33:3333,4444:4444",
			expected: []PortMapping{
				{
					HostPort:      "11",
					ContainerPort: "1111",
				},
				{
					HostPort:      "33",
					ContainerPort: "3333",
				},
				{
					HostPort:      "4444",
					ContainerPort: "4444",
				},
			},
		},
	}

	for _, tc := range tests {
		pmOutput := parsePortMappings(tc.extraPortMappings)
		if !reflect.DeepEqual(tc.expected, pmOutput) {
			t.Errorf("expected: %v, got: %v", tc.expected, pmOutput)
		}
	}
}

func TestFindRegistryConfig(t *testing.T) {
	type test struct {
		paths    []string
		expected string
	}
	tests := []test{
		{
			paths:    []string{"testdata/empty.json"},
			expected: "testdata/empty.json",
		},
		{
			paths:    []string{"doesntexist"},
			expected: "",
		},
		{
			paths:    []string{"doesntexist", "testdata/empty.json"},
			expected: "testdata/empty.json",
		},
	}

	for _, tc := range tests {
		out := findRegistryConfig(tc.paths)
		if !reflect.DeepEqual(tc.expected, out) {
			t.Errorf("expected:\n%v\ngot:\n%v", tc.expected, out)
		}
	}
}

func TestRenderRegistryCertsDirWithMirrors(t *testing.T) {
	type test struct {
		name              string
		cfg               v1alpha1.BuildCustomizationSpec
		expectedDirs      []string
		expectedFileCount int
		expectSkipVerify  bool
	}

	tests := []test{
		{
			name: "no mirrors",
			cfg: v1alpha1.BuildCustomizationSpec{
				Host: "cnoe.localtest.me",
				Port: "8443",
			},
			expectedDirs:      []string{"gitea.cnoe.localtest.me:8443"},
			expectedFileCount: 1,
		},
		{
			name: "with single mirror",
			cfg: v1alpha1.BuildCustomizationSpec{
				Host: "cnoe.localtest.me",
				Port: "8443",
				RegistryMirrors: []v1alpha1.RegistryMirror{
					{
						TargetRegistry:  "docker.io",
						RegistryAddress: "http://kind-registry:5000",
					},
				},
			},
			expectedDirs:      []string{"gitea.cnoe.localtest.me:8443", "docker.io"},
			expectedFileCount: 2,
		},
		{
			name: "with mirrors and insecure skip verify",
			cfg: v1alpha1.BuildCustomizationSpec{
				Host:                    "cnoe.localtest.me",
				Port:                    "8443",
				InsecureRegistryMirrors: true,
				RegistryMirrors: []v1alpha1.RegistryMirror{
					{
						TargetRegistry:  "docker.io",
						RegistryAddress: "http://kind-registry:5000",
					},
				},
			},
			expectedDirs:      []string{"gitea.cnoe.localtest.me:8443", "docker.io"},
			expectedFileCount: 2,
			expectSkipVerify:  true,
		},
		{
			name: "with multiple mirrors",
			cfg: v1alpha1.BuildCustomizationSpec{
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
				},
			},
			expectedDirs:      []string{"gitea.cnoe.localtest.me:8443", "docker.io", "ghcr.io"},
			expectedFileCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := renderRegistryCertsDir(tc.cfg)
			if err != nil {
				t.Fatalf("failed to render registry certs dir: %v", err)
			}
			defer os.RemoveAll(dir)

			// Check all expected directories exist
			for _, expectedDir := range tc.expectedDirs {
				fullPath := filepath.Join(dir, expectedDir)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("expected directory %s does not exist", fullPath)
				}

				// Check hosts.toml exists in each directory
				hostsFile := filepath.Join(fullPath, "hosts.toml")
				if _, err := os.Stat(hostsFile); os.IsNotExist(err) {
					t.Errorf("expected hosts.toml file %s does not exist", hostsFile)
				}

				// For mirrors, check the content
				if expectedDir != "gitea.cnoe.localtest.me:8443" {
					content, err := os.ReadFile(hostsFile)
					if err != nil {
						t.Fatalf("failed to read hosts.toml: %v", err)
					}
					contentStr := string(content)
					if tc.expectSkipVerify {
						if !strings.Contains(contentStr, "skip_verify = true") {
							t.Errorf("hosts.toml for mirror %s should contain skip_verify = true", expectedDir)
						}
					} else if strings.Contains(contentStr, "skip_verify = true") {
						t.Errorf("hosts.toml for mirror %s should not contain skip_verify = true", expectedDir)
					}
					if !strings.Contains(contentStr, "[host.") {
						t.Errorf("hosts.toml for mirror %s should contain host configuration", expectedDir)
					}
				}
			}
		})
	}
}
