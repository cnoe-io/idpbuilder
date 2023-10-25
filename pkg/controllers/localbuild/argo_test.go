package localbuild

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/k8s"
)

func TestGetRawInstallResources(t *testing.T) {
	resources, err := GetRawInstallResources()
	if err != nil {
		t.Fatalf("GetRawInstallResources() error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("GetRawInstallResources() resources len != 1, got %d", len(resources))
	}

	resourcePrefix := "# UCP ARGO INSTALL RESOURCES\n"
	checkPrefix := resources[0][0:len(resourcePrefix)]
	if resourcePrefix != string(checkPrefix) {
		t.Fatalf("GetRawInstallResources() exptected 1 resource with prefix %q, got %q", resourcePrefix, checkPrefix)
	}
}

func TestGetK8sInstallResources(t *testing.T) {
	objs, err := GetK8sInstallResources(k8s.GetScheme())
	if err != nil {
		t.Fatalf("GetK8sInstallResources() error: %v", err)
	}

	if len(objs) != 55 {
		t.Fatalf("Expected 57 Argo Install Resources, got: %d", len(objs))
	}
}
