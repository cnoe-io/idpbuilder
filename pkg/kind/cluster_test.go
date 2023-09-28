package kind

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
	  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
		endpoint = ["http://idpbuilder-registry:5000"]
	nodes:
	- role: control-plane
	  image: "kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1"
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
	  image: "kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1"
	  kubeadmConfigPatches:
	  - |
		kind: JoinConfiguration
		nodeRegistration:
		  kubeletExtraArgs:
			system-reserved: memory=4Gi`

	t.Errorf("Got config: %s", string(cfg))
	if diff := cmp.Diff(expectConfig, string(cfg)); diff != "" {
		t.Errorf("Expected config mismatch (-want +got):\n%s", diff)
	}
}
