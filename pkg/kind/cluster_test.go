package kind

import (
	"context"
	"io"
	"testing"

	runtime "github.com/cnoe-io/idpbuilder/pkg/runtime"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/kind/pkg/cluster/constants"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/exec"
)

func TestGetConfig(t *testing.T) {
	cluster, err := NewCluster("testcase", "v1.26.3", "", "", "", util.CorePackageTemplateConfig{
		Host: "cnoe.localtest.me",
		Port: "8443",
	})
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
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gitea.cnoe.localtest.me:8443"]
    endpoint = ["https://gitea.cnoe.localtest.me"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."gitea.cnoe.localtest.me".tls]
    insecure_skip_verify = true`
	assert.YAMLEq(t, expectConfig, string(cfg))
}

func TestExtraPortMappings(t *testing.T) {

	cluster, err := NewCluster("testcase", "v1.26.3", "", "", "22:32222", util.CorePackageTemplateConfig{
		Host: "cnoe.localtest.me",
		Port: "8443",
	})
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
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
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

// Mock provider for testing
type mockProvider struct {
	mock.Mock
	IProvider
}

func (m *mockProvider) ListNodes(name string) ([]nodes.Node, error) {
	args := m.Called(name)
	return args.Get(0).([]nodes.Node), args.Error(1)
}

type mockRuntime struct {
	mock.Mock
	runtime.IRuntime
}

func (m *mockRuntime) ContainerWithPort(ctx context.Context, name string, port string) (bool, error) {
	args := m.Called(ctx, name, port)
	return args.Get(0).(bool), args.Error(1)
}

// Mock Docker client for testing
type DockerClientMock struct {
	client.APIClient
	mock.Mock
}

func (m *DockerClientMock) ContainerList(ctx context.Context, listOptions types.ContainerListOptions) ([]types.Container, error) {
	mockArgs := m.Called(ctx, listOptions)
	return mockArgs.Get(0).([]types.Container), mockArgs.Error(1)
}

type NodeMock struct {
	mock.Mock
}

func (n *NodeMock) Command(command string, args ...string) exec.Cmd {
	argsMock := append([]string{command}, args...)
	mockArgs := n.Called(argsMock)
	return mockArgs.Get(0).(exec.Cmd)
}

func (n *NodeMock) String() string {
	args := n.Called()
	return args.String(0)
}

func (n *NodeMock) Role() (string, error) {
	args := n.Called()
	return args.String(0), args.Error(1)
}

func (n *NodeMock) IP() (ipv4 string, ipv6 string, err error) {
	args := n.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (n *NodeMock) SerialLogs(writer io.Writer) error {
	args := n.Called(writer)
	return args.Error(0)
}

func (n *NodeMock) CommandContext(ctx context.Context, cmd string, args ...string) exec.Cmd {
	mockArgs := n.Called(nil)
	return mockArgs.Get(0).(exec.Cmd)
}

func TestRunsOnWrongPort(t *testing.T) {
	// Mock node
	mockNode := &NodeMock{}
	mockNode.On("Role").Return(constants.ControlPlaneNodeRoleValue, nil)
	mockNode.On("String").Return("test-cluster")

	mockNodes := []nodes.Node{
		mockNode,
	}

	// Mock provider
	mockProvider := &mockProvider{}
	mockProvider.On("ListNodes", "test-cluster").Return(mockNodes, nil)

	cluster := &Cluster{
		name:     "test-cluster",
		provider: mockProvider,
		cfg: util.CorePackageTemplateConfig{
			Port: "8080",
		},
	}

	// Mock runtime
	mockRuntime1 := &mockRuntime{}
	mockRuntime1.On("ContainerWithPort", context.Background(), "test-cluster", "8080").Return(true, nil)
	cluster.runtime = mockRuntime1

	result, err := cluster.RunsOnRightPort(context.Background())
	assert.NoError(t, err)
	assert.True(t, result)

	// Mock Docker client
	mockRuntime2 := &mockRuntime{}
	mockRuntime2.On("ContainerWithPort", context.Background(), "test-cluster", "8080").Return(false, nil)
	cluster.runtime = mockRuntime2
	result, err = cluster.RunsOnRightPort(context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}
