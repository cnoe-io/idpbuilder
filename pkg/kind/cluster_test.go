package kind

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfig(t *testing.T) {
	cluster, err := NewCluster("testcase", "")
	if err != nil {
		t.Fatalf("Initializing cluster resource: %v", err)
	}

	cfg, err := cluster.getConfig()
	if err != nil {
		t.Errorf("Error getting kind config: %v", err)
	}

	expectConfig := `# two node (one workers) cluster config
# Kind kubernetes release images https://github.com/kubernetes-sigs/kind/releases
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]
    endpoint = ["http://idpbuilder-testcase-registry:5000"]
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
  - containerPort: 80
    hostPort: 8880
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  -
- role: worker
  image: "kindest/node:v1.26.3"
  kubeadmConfigPatches:
  - |
    kind: JoinConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        system-reserved: memory=4Gi`

	assert.Equal(t, expectConfig, string(cfg))
}
