package kind

import (
	"context"
	"io"
	"testing"

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
        system-reserved: memory=4Gi
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  `
	assert.YAMLEq(t, expectConfig, string(cfg))
}

func TestExtraPortMappings(t *testing.T) {

	cluster, err := NewCluster("testcase", "v1.26.3", "", "", "22:32222", util.CorePackageTemplateConfig{
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
        system-reserved: memory=4Gi
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  - containerPort: 32222
    hostPort: 22
    protocol: TCP`

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

	// Test when everything works fine
	container1 := types.Container{
		Names: []string{"/test-cluster"},
		Ports: []types.Port{
			{
				PublicPort: uint16(8080),
			},
		},
	}
	// Mock Docker client
	mockDockerClient1 := &DockerClientMock{}
	mockDockerClient1.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{container1}, nil)

	result, err := cluster.RunsOnRightPort(mockDockerClient1, context.Background())
	assert.NoError(t, err)
	assert.True(t, result)

	// Test when there's an error from the provider
	container2 := types.Container{
		Names: []string{"/test-cluster"},
		Ports: []types.Port{
			{
				PublicPort: uint16(9090),
			},
		},
	}
	// Mock Docker client
	mockDockerClient2 := &DockerClientMock{}
	mockDockerClient2.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{container2}, nil)
	result, err = cluster.RunsOnRightPort(mockDockerClient2, context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}
