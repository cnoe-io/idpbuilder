package kind

import (
	"os"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {

	type tc struct {
		host           string
		port           string
		registryConfig []string
		usePathRouting bool
		expectConfig   string
	}

	tcs := []tc{
		{
			host:           "cnoe.localtest.me",
			port:           "8443",
			registryConfig: []string{},
			usePathRouting: false,
			expectConfig: `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: "kindest/node:v1.26.3"
  labels:
    ingress-ready: "true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  - containerPort: 32222
    hostPort: 32222
    protocol: TCP
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gitea.cnoe.localtest.me:8443"]
    endpoint = ["https://gitea.cnoe.localtest.me"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."gitea.cnoe.localtest.me".tls]
    insecure_skip_verify = true`,
		},
		{
			host:           "cnoe.localtest.me",
			port:           "8443",
			registryConfig: []string{"testdata/doesnt-exist.json", "testdata/empty.json"},
			usePathRouting: true,
			expectConfig: `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: "kindest/node:v1.26.3"
  labels:
    ingress-ready: "true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  - containerPort: 32222
    hostPort: 32222
    protocol: TCP
  extraMounts:
  - containerPath: /var/lib/kubelet/config.json
    hostPath: testdata/empty.json
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."cnoe.localtest.me:8443"]
    endpoint = ["https://cnoe.localtest.me"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."cnoe.localtest.me".tls]
    insecure_skip_verify = true`,
		},
	}

	for i := range tcs {
		c := tcs[i]
		cluster, err := NewCluster("testcase", "v1.26.3", "", "", "", c.registryConfig, v1alpha1.BuildCustomizationSpec{
			Host:           c.host,
			Port:           c.port,
			UsePathRouting: c.usePathRouting,
		}, logr.Discard())
		assert.NoError(t, err)

		cfg, err := cluster.getConfig()
		assert.NoError(t, err)
		assert.YAMLEq(t, c.expectConfig, string(cfg))
	}
}

func TestExtraPortMappings(t *testing.T) {

	cluster, err := NewCluster("testcase", "v1.26.3", "", "", "22:32222", nil, v1alpha1.BuildCustomizationSpec{
		Host: "cnoe.localtest.me",
		Port: "8443",
	}, logr.Discard())
	if err != nil {
		t.Fatalf("Initializing cluster resource: %v", err)
	}

	cfg, err := cluster.getConfig()
	if err != nil {
		t.Errorf("Error getting kind config: %v", err)
	}

	expectConfig := `# Kind kubernetes release images https://github.com/kubernetes-sigs/kind/releases
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: "kindest/node:v1.26.3"
  labels:
    ingress-ready: "true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  - containerPort: 32222
    hostPort: 32222
    protocol: TCP
  - containerPort: 32222
    hostPort: 22
    protocol: TCP
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gitea.cnoe.localtest.me:8443"]
    endpoint = ["https://gitea.cnoe.localtest.me"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."gitea.cnoe.localtest.me".tls]
    insecure_skip_verify = true`

	assert.YAMLEq(t, expectConfig, string(cfg))
}

func TestGetConfigCustom(t *testing.T) {

	type testCase struct {
		inputPath  string
		outputPath string
		hostPort   string
		protocol   string
		error      bool
	}

	cases := []testCase{
		{
			inputPath:  "testdata/no-port.yaml",
			outputPath: "testdata/expected/no-port.yaml",
			hostPort:   "8443",
			protocol:   "https",
		},
		{
			inputPath:  "testdata/port-only.yaml",
			outputPath: "testdata/expected/port-only.yaml",
			hostPort:   "80",
			protocol:   "http",
		},
		{
			inputPath:  "testdata/no-port-multi.yaml",
			outputPath: "testdata/expected/no-port-multi.yaml",
			hostPort:   "8443",
			protocol:   "https",
		},
		{
			inputPath:  "testdata/label-only.yaml",
			outputPath: "testdata/expected/label-only.yaml",
			hostPort:   "8443",
			protocol:   "https",
		},
		{
			inputPath: "testdata/no-node",
			error:     true,
		},
	}

	for _, v := range cases {
		c, _ := NewCluster("testcase", "v1.26.3", "", v.inputPath, "", nil, v1alpha1.BuildCustomizationSpec{
			Host:     "cnoe.localtest.me",
			Port:     v.hostPort,
			Protocol: v.protocol,
		}, logr.Discard())

		b, err := c.getConfig()
		if v.error {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		expected, _ := os.ReadFile(v.outputPath)
		assert.YAMLEq(t, string(expected), string(b))
	}
}
