package k8s

import (
	"bytes"
	"embed"
	"os"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

//go:embed test-resources/*
var testDataFS embed.FS

func TestBuildCustomizedManifests(t *testing.T) {
	cases := map[string]struct {
		fsPath           string
		filePath         string
		expectedFilepath string
	}{
		"argocd": {
			fsPath:           "test-resources/input/argocd",
			filePath:         "test-resources/input/argocd-cm.yaml",
			expectedFilepath: "test-resources/output/argocd/install.yaml",
		},
		"nginx": {
			fsPath:           "test-resources/input/nginx",
			filePath:         "test-resources/input/extra.yaml",
			expectedFilepath: "test-resources/output/nginx/install.yaml",
		},
		"nginx-template": {
			fsPath:           "test-resources/input/nginx",
			filePath:         "test-resources/input/extra.yaml.tmpl",
			expectedFilepath: "test-resources/output/nginx/install-tmpl.yaml",
		},
	}

	for key := range cases {
		c := cases[key]
		b, err := BuildCustomizedManifests(c.filePath, c.fsPath, testDataFS, GetScheme(), v1alpha1.BuildCustomizationSpec{
			Protocol:       "http",
			Host:           "cnoe.localtest.me",
			IngressHost:    "localhost",
			Port:           "8443",
			UsePathRouting: false,
		})
		if err != nil {
			t.Fatalf("failed %s: %v", key, err)
		}

		expected, _ := os.ReadFile(c.expectedFilepath)
		expectedYamls := bytes.Split(expected, []byte{'-', '-', '-'})
		testYamls := make([][]byte, 0, 10)

		for f := range b {
			y := bytes.Split(b[f], []byte{'-', '-', '-'})
			testYamls = append(testYamls, y...)
		}

		if len(expectedYamls) != len(testYamls) {
			t.Fatalf("failed %s: number of yaml objects do not match", key)
		}

		for i := 0; i < len(expectedYamls); i++ {
			ok := assert.YAMLEq(t, string(expectedYamls[i]), string(testYamls[i]))
			if !ok {
				t.Fatalf("failed %s", key)
			}
		}
	}
}
